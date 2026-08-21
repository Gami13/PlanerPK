package types

import (
	"sync/atomic"
)

type Cache struct {
	availableGroups atomic.Pointer[[]ColumnGroup]
	timetable       atomic.Pointer[Timetable]
	legend          atomic.Pointer[Legend]
	buildings       atomic.Pointer[Buildings]
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

func (c *Cache) Legend() *Legend {
	return c.legend.Load()
}

func (c *Cache) SetLegend(legend *Legend) {
	c.legend.Store(legend)
}

func (c *Cache) Buildings() *Buildings {
	return c.buildings.Load()
}

func (c *Cache) SetBuildings(buildings *Buildings) {
	c.buildings.Store(buildings)
}
