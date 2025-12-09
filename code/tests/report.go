package tests

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
)

func NewTreeReport() *reportNode {
	return &reportNode{}
}

type reportNode struct {
	value         string
	childrenLeafs []string
	childrenNodes []*reportNode
}

func (t *reportNode) AddSpecReport(rep ginkgo.SpecReport) {
	if rep.LeafNodeText == "" {
		return
	}
	val := rep.LeafNodeText
	if rep.LeafNodeType == types.NodeTypeIt {
		val = "[It] " + val
	}
	r := 80 - len(val)
	if r < 0 {
		r = 0
	}
	val = val + strings.Repeat(" ", r) + " | " + rep.State.String()
	t.addSpecReport(
		rep.ContainerHierarchyTexts,
		val,
	)
}

func (t *reportNode) addSpecReport(path []string, specVal string) {
	if len(path) == 0 {
		t.childrenLeafs = append(t.childrenLeafs, specVal)
		return
	}
	p := path[0]
	path = path[1:]
	var next *reportNode
	for _, n := range t.childrenNodes {
		if n.value == p {
			next = n
			break
		}
	}
	if next == nil {
		next = &reportNode{
			value: p,
		}
		t.childrenNodes = append(t.childrenNodes, next)
	}
	next.addSpecReport(path, specVal)
}

func (t *reportNode) Print(prefix, indent string) {
	t.print(prefix, indent)
}

func (t *reportNode) print(prefix, indent string) {
	newPrefix := prefix
	if t.value != "" {
		println(prefix + t.value)
		newPrefix = prefix + indent
	}
	for _, l := range t.childrenLeafs {
		println(newPrefix + l)
	}
	for _, n := range t.childrenNodes {
		n.print(newPrefix, indent)
	}
}
