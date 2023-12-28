package risken

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v57/github"
)

func (r *riskenService) Diff(ctx context.Context, sourceCodePath string, pr github.PullRequest) ([]*string, error) {
	repo, err := git.PlainOpen(sourceCodePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: path=%s, err=%w", sourceCodePath, err)
	}

	if pr.Base == nil || pr.Base.SHA == nil {
		return nil, fmt.Errorf("base is nil")
	}
	baseRef, err := getCommitObject(repo, *pr.Base.SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get base commit: path=%s, sha=%s, err=%w", sourceCodePath, *pr.Base.SHA, err)
	}

	if pr.Head == nil || pr.Head.SHA == nil {
		return nil, fmt.Errorf("head is nil")
	}
	headRef, err := getCommitObject(repo, *pr.Head.SHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get head commit: path=%s, sha=%s, err=%w", sourceCodePath, *pr.Head.SHA, err)
	}

	patch, err := baseRef.Patch(headRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get patch: err=%w", err)
	}

	// 変更されたファイルを取得
	changeFiles := []*string{}
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		var filePath string
		if to != nil {
			filePath = to.Path()
		} else if from != nil {
			filePath = from.Path()
		}
		changeFiles = append(changeFiles, &filePath)
	}
	return changeFiles, nil
}

func getCommitObject(repo *git.Repository, sha string) (*object.Commit, error) {
	commit, err := repo.CommitObject(plumbing.NewHash(sha))
	if err != nil {
		return nil, err
	}
	return commit, nil
}
