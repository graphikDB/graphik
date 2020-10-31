package model

import "fmt"

type Path struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (p *Path) String() string {
	return fmt.Sprintf("%s.%s", p.Type, p.ID)
}
