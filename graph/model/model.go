package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Path struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (p *Path) String() string {
	return fmt.Sprintf("%s/%s", p.Type, p.ID)
}

func PathFromString(path string) Path {
	parts := strings.Split(path, "/")
	if len(parts) == 2 {
		return Path{
			ID:   parts[1],
			Type: parts[0],
		}
	}
	return Path{
		Type: parts[0],
	}
}

func (p *Path) UnmarshalGQL(v interface{}) error {
	pointStr, ok := v.(string)
	if !ok {
		return fmt.Errorf("path must be string ({type}/{id})")
	}
	parts := strings.Split(pointStr, "/")
	if len(parts) == 0 {
		return fmt.Errorf("empty path ({type}/{id})")
	}
	if len(parts) > 2 {
		return fmt.Errorf("path contains multiple separators(/) %s ({type}/{id})", pointStr)
	}
	p.Type = parts[0]
	if p.Type == "" {
		return fmt.Errorf("path does not contain type ({type}/{id})")
	}
	if len(parts) == 2 {
		p.ID = parts[1]
	}
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (p Path) MarshalGQL(w io.Writer) {
	json.NewEncoder(w).Encode(p.String())
}
