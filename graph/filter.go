package graph

type Filter struct {
	Type        string
	Expressions []string
	Limit       int
}

type PathFilter struct {
	Path   string
	Filter *Filter
}
