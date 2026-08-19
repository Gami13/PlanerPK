package types

import (
	"sync/atomic"

	"github.com/woodpkrs/PlanerPK/buildings"
	"github.com/woodpkrs/PlanerPK/legend"
)

type Cache struct {
	availableGroups atomic.Pointer[[]ColumnGroup]
	timetable       atomic.Pointer[Timetable]
	legend          atomic.Pointer[legend.Legend]
	buildings       atomic.Pointer[buildings.Buildings]
}

func (c *Cache) AvailableGroups() *[]ColumnGroup {
	return c.availableGroups.Load()
}

func (c *Cache) SetAvailableGroups(groups *[]ColumnGroup) {
	c.availableGroups.Store(groups)
}

func (c *Cache) Timetable() *Timetable {
	return c.timetable.Load()
}

func (c *Cache) SetTimetable(timetable *Timetable) {
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
