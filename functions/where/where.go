//go:generate godocdown -o README.md

package where

import (
	"github.com/autom8ter/graphik"
	"reflect"
)

// Equals returns true if the key-val is exactly the same within the attributes
func Equals(key string, val interface{}) graphik.WhereFunc {
	return func(g graphik.Graph, a graphik.Attributer) bool {
		return a.GetAttribute(key) == val
	}
}

// Exists returns true if the key exists within the attributes
func Exists(key string) graphik.WhereFunc {
	return func(g graphik.Graph, a graphik.Attributer) bool {
		return a.GetAttribute(key) != nil
	}
}

// DeepEqual returns true if the key-val is exactly the same within the attributes using reflection
func DeepEqual(key string, val interface{}) graphik.WhereFunc {
	return func(g graphik.Graph, a graphik.Attributer) bool {
		return reflect.DeepEqual(a.GetAttribute(key), val)
	}
}

// All returns true if the attributes passes all of the provided Where functions
func All(wheres ...graphik.WhereFunc) graphik.WhereFunc {
	return func(g graphik.Graph, a graphik.Attributer) bool {
		for _, where := range wheres {
			if !where(g, a) {
				return false
			}
		}
		return true
	}
}

// Any returns true if the attributes passes any of the provided Where functions
func Any(wheres ...graphik.WhereFunc) graphik.WhereFunc {
	return func(g graphik.Graph, a graphik.Attributer) bool {
		for _, where := range wheres {
			if where(g, a) {
				return true
			}
		}
		return false
	}
}
