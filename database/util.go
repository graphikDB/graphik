package database

import (
	apipb "github.com/autom8ter/graphik/gen/go"
	"sort"
)

func isGraphikAdmin(identity *apipb.Doc) bool {
	if identity == nil {
		return false
	}
	var isGraphikAdmin = false
	if attributes := identity.GetAttributes(); attributes != nil {
		if roles := attributes.Fields["roles"].GetListValue(); len(roles.GetValues()) > 0 {
			for _, role := range roles.GetValues() {
				if role.GetStringValue() == graphikAdminRole {
					isGraphikAdmin = true
				}
			}
		}
	}
	return isGraphikAdmin
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
