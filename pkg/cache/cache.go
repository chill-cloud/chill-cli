package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	copy2 "github.com/otiai10/copy"
	"golang.org/x/tools/go/vcs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type LocalCacheContext interface {
	GetCacheRoot() string
	Mark(mark string)
	CheckMarked(mark string) bool
}

type SimpleContext struct {
	Path  string
	Marks map[string]bool
}

func DefaultCacheContext() (LocalCacheContext, error) {
	var cacheContext SimpleContext
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cacheContext.Path = filepath.Join(dirname, ".chill", "cache")
	cacheContext.Marks = map[string]bool{}
	return &cacheContext, nil
}

func (c *SimpleContext) GetCacheRoot() string {
	return c.Path
}

func (c *SimpleContext) Mark(mark string) {
	c.Marks[mark] = true
}

func (c *SimpleContext) CheckMarked(mark string) bool {
	return c.Marks[mark]
}

type CachedSource interface {
	Update(c LocalCacheContext) error
	GetVersions(c LocalCacheContext) ([]version.Version, error)
	GetPath(c LocalCacheContext) string
	SwitchToVersion(c LocalCacheContext, v version.Version) error
}

type GitSource struct {
	Remote string
}

func (s *GitSource) GetVersions(c LocalCacheContext) ([]version.Version, error) {
	panic("implement me")
}

type GitLocalSource struct {
	LocalPath string
}

func (s *GitSource) getCacheFolderName() string {
	return hex.EncodeToString([]byte(s.Remote))
}

func (s *GitSource) GetPath(c LocalCacheContext) string {
	return path.Join(c.GetCacheRoot(), s.getCacheFolderName())
}

func (s *GitSource) Update(c LocalCacheContext) error {
	target := s.GetPath(c)
	if c.CheckMarked(target) {
		return nil
	}
	c.Mark(target)
	q, err := vcs.RepoRootForImportPath(s.Remote, false)
	if err != nil {
		return err
	}
	if stat, err := os.Stat(target); err == nil && stat.IsDir() {
		err := q.VCS.Download(target)
		if err != nil {
			return err
		}
	} else {
		err := q.VCS.Create(target, q.Repo)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *GitSource) SwitchToVersion(c LocalCacheContext, v version.Version) error {
	target := s.GetPath(c)
	q, err := vcs.RepoRootForImportPath(s.Remote, false)
	if err != nil {
		return err
	}
	err = q.VCS.TagSync(target, fmt.Sprintf("chill-%s", v.String()))
	if err != nil {
		return err
	}
	return nil
}

func (s *GitLocalSource) getCacheFolderName() string {
	x := sha256.Sum256([]byte(s.LocalPath))
	return hex.EncodeToString(x[:])
}

func (s *GitLocalSource) GetPath(c LocalCacheContext) string {
	return path.Join(c.GetCacheRoot(), s.getCacheFolderName())
}

func (s *GitLocalSource) Update(c LocalCacheContext) error {
	target := s.GetPath(c)
	if c.CheckMarked(target) {
		return nil
	}
	c.Mark(target)
	err := os.RemoveAll(target)
	if err != nil {
		return err
	}
	err = os.MkdirAll(target, os.ModePerm)
	if err != nil {
		return err
	}
	err = copy2.Copy(s.LocalPath, target)
	if err != nil {
		return err
	}
	return nil
}

func (s *GitLocalSource) GetVersions(c LocalCacheContext) ([]version.Version, error) {
	target := s.GetPath(c)
	r, err := git.PlainOpen(target)
	if err != nil {
		return nil, err
	}
	iter, err := r.TagObjects()
	if err != nil {
		return nil, err
	}

	var res []version.Version

	err = iter.ForEach(func(tag *object.Tag) error {
		if strings.HasPrefix(tag.Name, "chill-") {
			v, err := version.ParseFromString(strings.TrimPrefix(tag.Name, "chill-"))
			if err != nil {
				return err
			}
			res = append(res, *v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *GitLocalSource) SwitchToVersion(c LocalCacheContext, v version.Version) error {
	target := s.GetPath(c)
	r, err := git.PlainOpen(target)
	if err != nil {
		return err
	}
	ref, err := r.Tag(fmt.Sprintf("chill-%s", v.String()))
	if err != nil {
		return err
	}
	wt, err := r.Worktree()
	if err != nil {
		return err
	}
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: ref.Name(),
	})
	if err != nil {
		return err
	}
	return nil
}
