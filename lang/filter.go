package lang

type Filter struct {
	Type        string
	Expressions []string
	Limit       int32
}

type PathFilter struct {
	Path   string
	Filter *Filter
}
