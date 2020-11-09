package sortable

import "sort"

type Sortable struct {
	LenFunc  func() int
	LessFunc func(i, j int) bool
	SwapFunc func(i, j int)
}

func (s Sortable) Len() int {
	if s.LenFunc == nil {
		return 0
	}
	return s.LenFunc()
}

func (s Sortable) Less(i, j int) bool {
	if s.LessFunc == nil {
		return false
	}
	return s.LessFunc(i, j)
}

func (s Sortable) Swap(i, j int) {
	if s.SwapFunc == nil {
		return
	}
	s.SwapFunc(i, j)
}

func (s Sortable) Sort() {
	sort.Sort(s)
}
