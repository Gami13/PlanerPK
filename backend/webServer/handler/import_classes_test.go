package handler

import (
	"connectrpc.com/connect"
	"context"
	"database/sql"
	"github.com/woodpkrs/PlanerPK/cache"
	"github.com/woodpkrs/PlanerPK/gen/db"
	pb "github.com/woodpkrs/PlanerPK/gen/pkplaner/v1"
	"github.com/woodpkrs/PlanerPK/types"
	_ "modernc.org/sqlite"
	"testing"
	"time"
)

func TestImportClassLifecycle(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	_, err = database.Exec(`CREATE TABLE overrides (id INTEGER PRIMARY KEY, user_id INTEGER, class_hash TEXT, start_hour TEXT NOT NULL DEFAULT '', is_hidden BOOLEAN NOT NULL DEFAULT 0, is_transparent BOOLEAN NOT NULL DEFAULT 0, source_column INTEGER NOT NULL DEFAULT 0, UNIQUE(user_id, class_hash))`)
	if err != nil {
		t.Fatal(err)
	}
	base := types.Class{Name: "Base", Day: time.Monday}
	other := types.Class{Name: "Imported", Day: time.Tuesday}
	all := types.Timetable{1: {time.Monday: {base}}, 2: {time.Monday: {base}, time.Tuesday: {other}}, 3: {time.Tuesday: {other}}}
	groups := []types.ColumnGroup{{Column: 1, Category: "I st. sem. 1"}, {Column: 2, Category: "I st. sem. 1"}, {Column: 3, Category: "I st. sem. 3"}}
	c := &cache.Cache{}
	c.SetTimetable(&all)
	c.SetAvailableGroups(&groups)
	h := NewPKPlanerHandler(db.New(database), c)
	column := int64(1)
	ctx := context.WithValue(context.Background(), AuthUserKey, db.User{ID: 1, ChosenColumn: &column})
	eligible, err := h.GetImportGroups(ctx, connect.NewRequest(&pb.GetColumnsRequest{}))
	if err != nil || len(eligible.Msg.Columns) != 1 || eligible.Msg.Columns[0].Column != 2 {
		t.Fatalf("incorrect eligible groups: %v %v", eligible, err)
	}
	set := func(override *pb.ClassOverride) error {
		_, err := h.SetOverride(ctx, connect.NewRequest(&pb.SetOverrideRequest{Override: override}))
		return err
	}
	if err := set(&pb.ClassOverride{ClassHash: other.Hash(), SourceColumn: 3}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("cross-year accepted: %v", err)
	}
	if err := set(&pb.ClassOverride{ClassHash: "missing", SourceColumn: 2}); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("missing class accepted: %v", err)
	}
	if err := set(&pb.ClassOverride{ClassHash: base.Hash(), SourceColumn: 2}); connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("base duplicate accepted: %v", err)
	}
	if err := set(&pb.ClassOverride{ClassHash: other.Hash(), SourceColumn: 2}); err != nil {
		t.Fatal(err)
	}
	// Adding twice is idempotent, and user-specific classes never leak into public schedules.
	if err := set(&pb.ClassOverride{ClassHash: other.Hash(), SourceColumn: 2}); err != nil {
		t.Fatal(err)
	}
	result, err := h.GetClasses(ctx, connect.NewRequest(&pb.GetClassesRequest{}))
	if err != nil || len(result.Msg.Classes) != 2 {
		t.Fatalf("missing addition: %v %v", result, err)
	}
	public, err := h.GetTimetable(context.Background(), connect.NewRequest(&pb.GetTimetableRequest{Column: 1}))
	if err != nil || len(public.Msg.Classes) != 1 {
		t.Fatalf("public leak: %v %v", public, err)
	}
	if err := set(&pb.ClassOverride{ClassHash: other.Hash(), SourceColumn: 2, StartHour: "10:00", IsTransparent: true}); err != nil {
		t.Fatal(err)
	}
	rows, err := h.GetOverrides(ctx, connect.NewRequest(&pb.GetOverridesRequest{}))
	if err != nil || len(rows.Msg.Overrides) != 1 || rows.Msg.Overrides[0].SourceColumn != 2 {
		t.Fatalf("source lost: %v %v", rows, err)
	}
	if err := set(&pb.ClassOverride{ClassHash: other.Hash()}); err != nil {
		t.Fatal(err)
	}
	result, err = h.GetClasses(ctx, connect.NewRequest(&pb.GetClassesRequest{}))
	if err != nil || len(result.Msg.Classes) != 1 {
		t.Fatalf("removal failed: %v %v", result, err)
	}
}
