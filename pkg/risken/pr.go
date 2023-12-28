package risken

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// GithubPREvent is a struct for GitHub Pull Request Event.
// ref: https://docs.github.com/ja/webhooks/webhook-events-and-payloads#pull_request
// example: https://github.com/pingdotgg/sample_hooks/blob/main/github_pr_opened.json
type GithubPREvent struct {
	Action      string              `json:"action"`
	Number      int                 `json:"number"`
	PullRequest *github.PullRequest `json:"pull_request"`
	Repository  *github.Repository  `json:"repository"`
}

func (r *riskenService) GetGithubPREvent(ctx context.Context, githubEventPath string) (*GithubPREvent, error) {
	file, err := os.Open(githubEventPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: path=%s, err=%w", githubEventPath, err)
	}
	defer file.Close()

	var event GithubPREvent
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&event)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json: err=%w", err)
	}
	return &event, nil
}

func (r *riskenService) PullRequestComment(ctx context.Context, repo *github.Repository, pr *github.PullRequest, scanResults []*ScanResult) error {
	fullName := *repo.FullName
	if !strings.Contains(fullName, "/") {
		return fmt.Errorf("invalid repository name: %s", fullName)
	}
	owner := strings.Split(fullName, "/")[0]
	repoName := strings.Split(fullName, "/")[1]
	if pr.Number == nil {
		return fmt.Errorf("invalid pull request number: %v", pr.Number)
	}
	prNumber := *pr.Number

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.conf.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for _, r := range scanResults {
		comment := &github.PullRequestComment{
			Body:     github.String(r.ReviewComment + "\n\n_By RISKEN review_"),
			CommitID: github.String(*pr.Head.SHA),
			Path:     github.String(r.File),
			Position: github.Int(r.Line),
		}

		_, _, err := client.PullRequests.CreateComment(ctx, owner, repoName, prNumber, comment)
		if err != nil {
			return fmt.Errorf("failed to create comment: err=%w", err)
		}
	}
	return nil
}
