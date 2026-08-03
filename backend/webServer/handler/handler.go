package handler

import (
	"github.com/woodpkrs/PlanerPK/gen/db"
	"github.com/woodpkrs/PlanerPK/gen/pkplaner/v1/pkplanerv1connect"
	"github.com/woodpkrs/PlanerPK/types"
)

// PKPlanerHandler implements pkplanerv1connect.PKPlanerServiceHandler.
type PKPlanerHandler struct {
	pkplanerv1connect.UnimplementedPKPlanerServiceHandler
	Query db.Querier
	cache *types.Cache
}

func NewPKPlanerHandler(query db.Querier, cache *types.Cache) *PKPlanerHandler {
	if query == nil {
		panic("Query is nil!")
	}
	return &PKPlanerHandler{Query: query, cache: cache}
}

const AuthUserKey ContextKey = "AuthenticatedUser"

type ContextKey string
