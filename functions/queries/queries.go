package queries

import (
	"fmt"
	"github.com/autom8ter/graphik"
	"github.com/autom8ter/graphik/functions/handlers"
	"io"
)

func DotFileQuery(w io.Writer, mods ...graphik.EdgeQueryModFunc) graphik.EdgeQuery {
	q := graphik.NewEdgeQuery()
	for _, mod := range mods {
		q = q.Mod(mod)
	}

	w.Write([]byte(fmt.Sprintf("digraph G {\n")))
	q = q.Mod(graphik.EdgeModHandler(handlers.WriteDotFile(w)))
	q = q.Mod(graphik.EdgeModCloser(func() error {
		w.Write([]byte(fmt.Sprintf("}\n")))
		return nil
	}))

	return q
}
