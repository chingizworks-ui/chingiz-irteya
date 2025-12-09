package ports

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniquePorts(t *testing.T) {
	uniquePorts := make(map[int]string)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "./", nil, parser.AllErrors)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	pkg := pkgs["ports"]

	for _, file := range pkg.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				return true
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, name := range valueSpec.Names {
					basicLit, ok := valueSpec.Values[i].(*ast.BasicLit)
					if ok && basicLit.Kind == token.INT {
						port, err := strconv.Atoi(basicLit.Value)
						require.NoError(t, err)
						existedConstName, exist := uniquePorts[port]
						require.False(t, exist, fmt.Sprintf("порт %d должен быть уникальным, встречается в следующих константах: %s и %s", port, name.Name, existedConstName))
						uniquePorts[port] = name.Name
					}
				}
			}
			return true
		})
	}
}
