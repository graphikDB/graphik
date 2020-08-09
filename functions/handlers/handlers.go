//go:generate godocdown -o README.md

package handlers

import (
	"fmt"
	"github.com/autom8ter/graphik"
	"io"
	"strings"
)

func WriteDotFile(w io.Writer) graphik.EdgeHandlerFunc {
	return func(g graphik.Graph, e graphik.Edge) error {
		_, err := w.Write([]byte(fmt.Sprintf("%s -> %s;\n", cleanse(e.From().PathString()), cleanse(e.To().PathString()))))
		if err != nil {
			return err
		}
		return nil
	}
}

func cleanse(str string) string {
	return strings.NewReplacer("/", "_", ".", "_", "-", "_").Replace(str)
}
