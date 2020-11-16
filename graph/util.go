package graph

import (
	apipb "github.com/autom8ter/graphik/api"
	"github.com/pkg/errors"
)

func removeEdge(path *apipb.Path, paths []*apipb.Path) []*apipb.Path {
	var newPaths []*apipb.Path
	for _, p := range paths {
		if p != path {
			newPaths = append(newPaths, p)
		}
	}
	return newPaths
}

func noExist(path *apipb.Path) error {
	return errors.Errorf("%s.%s does not exist", path.Gtype, path.Gid)
}
