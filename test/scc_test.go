package test

import (
	"github.com/chill-cloud/chill-cli/pkg/validate"
	"testing"
)

func AddEdge(g *validate.SccContext, a, b string) {
	g.Adj[a] = append(g.Adj[a], b)
	g.Rev[b] = append(g.Rev[b], a)
}

func TestScc(t *testing.T) {
	g1 := validate.NewSccContext()
	AddEdge(g1, "a", "b")
	AddEdge(g1, "b", "c")
	AddEdge(g1, "c", "a")

	if len(g1.FindScc()) != 1 {
		t.Fatal("cycle not found")
	}

	g2 := validate.NewSccContext()
	AddEdge(g2, "a", "b")
	AddEdge(g2, "b", "c")
	AddEdge(g2, "c", "d")

	if len(g2.FindScc()) != 4 {
		t.Fatal("cycle found")
	}

	g3 := validate.NewSccContext()
	AddEdge(g3, "b", "a")
	AddEdge(g3, "c", "a")
	AddEdge(g3, "d", "a")

	if len(g3.FindScc()) != 4 {
		t.Fatal("cycle found")
	}

	g4 := validate.NewSccContext()
	AddEdge(g4, "a", "b")
	AddEdge(g4, "b", "c")
	AddEdge(g4, "c", "a")
	AddEdge(g4, "b", "d")
	AddEdge(g4, "d", "e")
	AddEdge(g4, "e", "b")

	if len(g4.FindScc()) != 1 {
		t.Fatal("cycle not found")
	}
}
