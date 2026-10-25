package cache

import (
	"sync/atomic"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/snapshot"
	"github.com/woodpkrs/PlanerPK/types"
)

type Cache struct {
	availableGroups atomic.Pointer[[]types.ColumnGroup]
	timetable       atomic.Pointer[types.Timetable]
	timetableHash   atomic.Pointer[string]
	legend          atomic.Pointer[legend.Legend]
	buildings       atomic.Pointer[buildings.Buildings]
	snapshot        atomic.Pointer[snapshot.State]
}

func (c *Cache) AvailableGroups() *[]types.ColumnGroup {
	return c.availableGroups.Load()
}

func (c *Cache) SetAvailableGroups(groups *[]types.ColumnGroup) {
	c.availableGroups.Store(groups)
}

func (c *Cache) Timetable() *types.Timetable {
	return c.timetable.Load()
}

func (c *Cache) SetTimetable(timetable *types.Timetable) {
	c.timetable.Store(timetable)
}

func (c *Cache) TimetableHash() string {
	if hash := c.timetableHash.Load(); hash != nil {
		return *hash
	}
	return ""
}

func (c *Cache) SetTimetableHash(hash string) {
	c.timetableHash.Store(&hash)
}

func (c *Cache) Snapshot() *snapshot.State {
	return c.snapshot.Load()
}

func (c *Cache) SetSnapshot(state *snapshot.State) {
	c.snapshot.Store(state)
}

func (c *Cache) Legend() *legend.Legend {
	return c.legend.Load()
}

func (c *Cache) SetLegend(legend *legend.Legend) {
	c.legend.Store(legend)
}

func (c *Cache) Buildings() *buildings.Buildings {
	return c.buildings.Load()
}

func (c *Cache) SetBuildings(buildings *buildings.Buildings) {
	c.buildings.Store(buildings)
}
