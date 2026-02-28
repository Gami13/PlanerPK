package types

import "sync"

// TimetableCacheState is a concurrency-safe cache for per-group timetables.
type TimetableCacheState struct {
	mu      sync.RWMutex
	data    map[int64]Timetable
	isStale bool
}

func CloneTimetable(timetable Timetable) Timetable {
	cloned := make(Timetable, len(timetable))
	for weekDay, classes := range timetable {
		clonedClasses := make([]Class, len(classes))
		copy(clonedClasses, classes)
		cloned[weekDay] = clonedClasses
	}
	return cloned
}

func (c *TimetableCacheState) GetIfFresh(groupColumn int64) (Timetable, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isStale {
		return nil, false
	}

	timetable, ok := c.data[groupColumn]
	if !ok {
		return nil, false
	}

	return CloneTimetable(timetable), true
}

func (c *TimetableCacheState) SetFresh(groupColumn int64, timetable Timetable) {
	c.mu.Lock()
	if c.data == nil {
		c.data = make(map[int64]Timetable)
	}
	c.data[groupColumn] = CloneTimetable(timetable)
	c.isStale = false
	c.mu.Unlock()
}

func (c *TimetableCacheState) MarkStale() {
	c.mu.Lock()
	c.isStale = true
	c.mu.Unlock()
}

func NewTimetableCache() *TimetableCacheState {
	return &TimetableCacheState{
		data:    make(map[int64]Timetable),
		isStale: true,
	}
}

// AvailableGroupsCacheState is a concurrency-safe cache for column groups.
type AvailableGroupsCacheState struct {
	mu      sync.RWMutex
	groups  []ColumnGroup
	isStale bool
}

func CopyAvailableGroups(groups []ColumnGroup) []ColumnGroup {
	out := make([]ColumnGroup, len(groups))
	copy(out, groups)
	return out
}

func (c *AvailableGroupsCacheState) GetIfFresh() ([]ColumnGroup, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isStale {
		return nil, false
	}

	return CopyAvailableGroups(c.groups), true
}

func (c *AvailableGroupsCacheState) SetFresh(groups []ColumnGroup) {
	c.mu.Lock()
	c.groups = CopyAvailableGroups(groups)
	c.isStale = false
	c.mu.Unlock()
}

func (c *AvailableGroupsCacheState) MarkStale() {
	c.mu.Lock()
	c.isStale = true
	c.mu.Unlock()
}

func NewAvailableGroupsCache() *AvailableGroupsCacheState {
	return &AvailableGroupsCacheState{isStale: true}
}

// LegendCacheState is a concurrency-safe cache for the parsed legend.
type LegendCacheState struct {
	mu      sync.RWMutex
	data    *Legend
	isStale bool
}

func (c *LegendCacheState) GetIfFresh() (*Legend, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isStale || c.data == nil {
		return nil, false
	}
	return c.data, true
}

func (c *LegendCacheState) SetFresh(legend *Legend) {
	c.mu.Lock()
	c.data = legend
	c.isStale = false
	c.mu.Unlock()
}

func (c *LegendCacheState) MarkStale() {
	c.mu.Lock()
	c.isStale = true
	c.mu.Unlock()
}

func NewLegendCache() *LegendCacheState {
	return &LegendCacheState{isStale: true}
}
