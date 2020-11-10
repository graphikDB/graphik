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

func FilterFrom(m map[string]interface{}) *Filter {
	filter := &Filter{}
	if m["type"] != nil {
		filter.Type = m["type"].(string)
	}
	if m["limit"] != nil {
		filter.Limit = m["limit"].(int)
	}
	if m["expressions"] != nil {
		filter.Expressions = m["expressions"].([]string)
	}
	return filter
}
