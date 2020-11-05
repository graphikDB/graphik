package generic

import "sort"

type Interface struct {
	LenFunc  func() int
	LessFunc func(i, j int) bool
	SwapFunc func(i, j int)
}

func (s Interface) Len() int {
	if s.LenFunc == nil {
		return 0
	}
	return s.LenFunc()
}

func (s Interface) Less(i, j int) bool {
	if s.LessFunc == nil {
		return false
	}
	return s.LessFunc(i, j)
}

func (s Interface) Swap(i, j int) {
	if s.SwapFunc == nil {
		return
	}
	s.SwapFunc(i, j)
}

func (s Interface) Sort() {
	sort.Sort(s)
}
