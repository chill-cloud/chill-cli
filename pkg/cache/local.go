package cache

import (
	"errors"
	"fmt"
	"github.com/chill-cloud/chill-cli/pkg/version"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"strings"
)

type NotCommittedError interface {
	error
	NotCommittedFiles() []string
}

type notCommitedError []string

func (e notCommitedError) Error() string {
	return strings.Join(e, "\n")
}

func (e notCommitedError) NotCommittedFiles() []string {
	return e
}

type SourceOfTruth interface {
	CheckVersion(v version.Version) (bool, error)
	GetVersions() ([]version.Version, error)
	FreezeVersion(v version.Version) error
	IsFrozen() (bool, error)
	IsClean() ([]string, error)
}

type localSourceOfTruth struct {
	Path       string
	Repository *git.Repository
	Worktree   *git.Worktree
}

func NewLocalSourceOfTruth(path string) (SourceOfTruth, error) {
	s := localSourceOfTruth{Path: path}
	r, err := git.PlainOpen(s.Path)
	s.Repository = r
	if err != nil {
		return nil, err
	}
	wt, err := r.Worktree()
	s.Worktree = wt
	if err != nil {
		return nil, err
	}
	return &s, err
}

func (s *localSourceOfTruth) CheckVersion(v version.Version) (bool, error) {
	list, err := s.GetVersions()
	if err != nil {
		return false, err
	}
	has := false
	for _, ver := range list {
		if v == ver {
			has = true
			break
		}
	}
	if !has {
		return false, nil
	}
	ref, err := s.Repository.Tag(fmt.Sprintf("chill-%s", v.String()))
	if err != nil {
		return false, err
	}
	_, err = s.Repository.TagObject(ref.Hash())
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, plumbing.ErrObjectNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (s *localSourceOfTruth) GetVersions() ([]version.Version, error) {
	iter, err := s.Repository.TagObjects()
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

func (s *localSourceOfTruth) IsFrozen() (bool, error) {
	r, err := s.Repository.Head()
	if err != nil {
		return false, err
	}
	commit, err := s.Repository.CommitObject(r.Hash())
	if err != nil {
		return false, err
	}
	iter, err := s.Repository.TagObjects()
	if err != nil {
		return false, err
	}
	res := false
	err = iter.ForEach(func(tag *object.Tag) error {
		if strings.HasPrefix(tag.Name, "chill-") {
			_, err := version.ParseFromString(strings.TrimPrefix(tag.Name, "chill-"))
			if err != nil {
				return err
			}
			tagCommit, err := tag.Commit()
			if err != nil {
				return err
			}
			if tagCommit.Hash == commit.Hash {
				res = true
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return res, nil
}

func (s *localSourceOfTruth) IsClean() ([]string, error) {
	workTreeStatus, err := s.Worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("unable to get work tree status: %w", err)
	}

	if !workTreeStatus.IsClean() {
		var badFiles []string
		for name, status := range workTreeStatus {
			if status.Worktree != git.Unmodified || status.Staging != git.Unmodified {
				badFiles = append(badFiles, name)
			}
		}
		return badFiles, nil
	}
	return nil, nil
}

func (s *localSourceOfTruth) FreezeVersion(v version.Version) error {
	clean, err := s.IsClean()

	if err != nil {
		return err
	}

	if clean != nil {
		return notCommitedError(clean)
	}

	head, err := s.Repository.Head()
	if err != nil {
		return fmt.Errorf("unable to get HEAD of the Git repository: %w", err)
	}

	tagString := fmt.Sprintf("chill-%s", v.String())

	tag := object.Tag{
		Name:       tagString,
		Message:    "Auto-generated with Chill",
		Target:     head.Hash(),
		TargetType: plumbing.CommitObject,
	}

	e := s.Repository.Storer.NewEncodedObject()
	err = tag.Encode(e)
	if err != nil {
		return fmt.Errorf("unable to assign tag: %w", err)
	}
	hash, err := s.Repository.Storer.SetEncodedObject(e)
	if err != nil {
		return fmt.Errorf("unable to store tag: %w", err)
	}

	err = s.Repository.Storer.SetReference(plumbing.NewReferenceFromStrings("refs/tags/"+tagString, hash.String()))
	if err != nil {
		return fmt.Errorf("unable to set reference for tag: %w", err)
	}

	return nil
}
