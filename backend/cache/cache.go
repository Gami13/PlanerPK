package cache

import (
	"sync/atomic"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
	"github.com/woodpkrs/PlanerPK/types"
)

type Cache struct {
	availableGroups atomic.Pointer[[]types.ColumnGroup]
	timetable       atomic.Pointer[types.Timetable]
	legend          atomic.Pointer[legend.Legend]
	buildings       atomic.Pointer[buildings.Buildings]
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
