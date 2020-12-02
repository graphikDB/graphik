package database

import (
	apipb "github.com/autom8ter/graphik/gen/go"
	"github.com/autom8ter/graphik/helpers"
	"sort"
)

func (g *Graph) isGraphikAdmin(identity *apipb.Doc) bool {
	if identity.GetAttributes().GetFields() == nil {
		return false
	}
	return helpers.ContainsString(identity.GetAttributes().GetFields()["email"].GetStringValue(), g.rootUsers)
}

func removeConnection(path *apipb.Path, paths []*apipb.Path) []*apipb.Path {
	var newPaths []*apipb.Path
	for _, p := range paths {
		if path.Gid == p.Gid && path.Gtype == p.Gtype {
			newPaths = append(newPaths, p)
		}
	}
	sortPaths(newPaths)
	return newPaths
}

func sortPaths(paths []*apipb.Path) {
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].String() < paths[j].String()
	})
}
