package review

import (
	"context"
	"time"

	"github.com/google/go-github/v44/github"
)

func (r *reviewService) Diff(ctx context.Context, pr *GithubPREvent) ([]*github.CommitFile, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}
	changeFiles := []*github.CommitFile{}
	for {
		files, resp, err := r.githubClient.PullRequests.ListFiles(ctx, pr.Owner, pr.RepoName, pr.Number, opts)
		if err != nil {
			return nil, err
		}
		changeFiles = append(changeFiles, files...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		time.Sleep(1 * time.Second)
	}
	return changeFiles, nil
}
