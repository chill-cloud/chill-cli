package validate

import (
	"container/list"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/cache"
	"github.com/chill-cloud/chill-cli/pkg/config"
	"github.com/chill-cloud/chill-cli/pkg/logging"
	"github.com/chill-cloud/chill-cli/pkg/service"
	"strings"
)

type SccContext struct {
	Adj     map[string][]string
	Rev     map[string][]string
	order   []string
	visited map[string]bool
	scc     []string
}

func NewSccContext() *SccContext {
	return &SccContext{
		Adj:     map[string][]string{},
		Rev:     map[string][]string{},
		visited: map[string]bool{},
	}
}

func (c *SccContext) dfsIn(v string) {
	c.visited[v] = true
	for _, u := range c.Adj[v] {
		if !c.visited[u] {
			c.dfsIn(u)
		}
	}
	c.order = append(c.order, v)
}

func (c *SccContext) dfsOut(v string) {
	c.visited[v] = true
	c.scc = append(c.scc, v)
	for _, u := range c.Rev[v] {
		if !c.visited[u] {
			c.dfsOut(u)
		}
	}
}

func (c *SccContext) FindScc() [][]string {
	c.visited = map[string]bool{}
	for v := range c.Adj {
		if !c.visited[v] {
			c.dfsIn(v)
		}
	}
	c.visited = map[string]bool{}
	var res [][]string
	for i := len(c.order) - 1; i >= 0; i-- {
		v := c.order[i]
		if !c.visited[v] {
			c.scc = nil
			c.dfsOut(v)
			res = append(res, c.scc)
		}
	}
	return res
}

func ValidateGraph(pc *service.ProjectConfig, c cache.LocalCacheContext, forceLocal bool) error {
	adj := map[string][]string{}
	rev := map[string][]string{}

	stack := list.List{}
	stack.PushBack(pc)
	visited := map[string]bool{}
	adj[pc.Name] = []string{}
	rev[pc.Name] = []string{}
	for stack.Len() > 0 {
		back := stack.Back()
		cur, ok := back.Value.(*service.ProjectConfig)
		if !ok {
			return fmt.Errorf("wrong type")
		}
		stack.Remove(back)

		visited[cur.Name] = true
		logging.Logger.Info(cur.Name)

		for dep := range cur.Dependencies {
			if !forceLocal {
				err := dep.Cache().Update(c)
				if err != nil {
					return err
				}
			}
			cfg, err := config.ParseConfig(dep.Cache().GetPath(c), config.LockConfigName, true)
			if err != nil {
				return err
			}
			if cfg.Name != dep.GetName() {
				return fmt.Errorf("dependency name must follow the name specified in its config")
			}
			if !visited[cfg.Name] {
				stack.PushBack(cfg)
			}
			adj[cur.Name] = append(adj[cur.Name], cfg.Name)
			rev[cfg.Name] = append(rev[cfg.Name], cur.Name)
		}
	}
	ctx := SccContext{Adj: adj, Rev: rev}
	res := ctx.FindScc()
	if len(res) != len(visited) {
		var goodOne []string
		for _, scc := range res {
			if len(scc) > 1 {
				goodOne = scc
				break
			}
		}
		return fmt.Errorf("cyclic dependency found; these services form a stronly connected component:\n"+
			"%s", strings.Join(append(goodOne, goodOne[0]), " -> "))
	}
	return nil
}
