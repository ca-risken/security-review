package risken

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
)

// GithubPREvent is a struct for GitHub Pull Request Event.
// ref: https://docs.github.com/ja/webhooks/webhook-events-and-payloads#pull_request
// example: https://github.com/pingdotgg/sample_hooks/blob/main/github_pr_opened.json
type GithubPREvent struct {
	Action      string              `json:"action"`
	Number      int                 `json:"number"`
	PullRequest *github.PullRequest `json:"pull_request"`
	Repository  *github.Repository  `json:"repository"`
	Owner       string              `json:"owner"`
	RepoName    string              `json:"repo_name"`
}

func (r *riskenService) GetGithubPREvent() (*GithubPREvent, error) {
	file, err := os.Open(r.conf.GithubEventPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: path=%s, err=%w", r.conf.GithubEventPath, err)
	}
	defer file.Close()

	var event GithubPREvent
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&event)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json: err=%w", err)
	}

	if event.Repository == nil || event.Repository.FullName == nil {
		return nil, fmt.Errorf("invalid repository: %v", event.Repository)
	}
	fullName := *event.Repository.FullName
	if !strings.Contains(fullName, "/") {
		return nil, fmt.Errorf("invalid repository name: %s", fullName)
	}
	event.Owner = strings.Split(fullName, "/")[0]
	event.RepoName = strings.Split(fullName, "/")[1]
	return &event, nil
}

func (r *riskenService) PullRequestComment(ctx context.Context, pr *GithubPREvent, scanResults []*ScanResult) error {
	if len(scanResults) == 0 {
		comment := &github.IssueComment{
			Body: github.String("セキュリティレビューを実施しました。\n特に問題は見つかりませんでした👏\n\n_By RISKEN review_"),
		}
		_, _, err := r.githubClient.Issues.CreateComment(ctx, pr.Owner, pr.RepoName, pr.Number, comment)
		if err != nil {
			return fmt.Errorf("failed to create comment: err=%w", err)
		}
		return nil
	}

	// Review Comment
	comments, err := r.getAllComments(ctx, pr.Owner, pr.RepoName, pr.Number)
	if err != nil {
		return fmt.Errorf("failed to get all comments: err=%w", err)
	}
	for _, result := range scanResults {
		if existsSimilarComment(comments, result.ScanID) {
			r.logger.WarnContext(ctx, "already exists similar comment", slog.String("file", result.File), slog.Int("line", result.Line), slog.String("ID", result.ScanID))
			continue
		}
		comment := &github.PullRequestComment{
			Body:     github.String(result.ReviewComment + "\n\n_By RISKEN review_"),
			CommitID: github.String(*pr.PullRequest.Head.SHA),
			Path:     github.String(result.File),
			Line:     github.Int(result.Line),
		}
		_, _, err := r.githubClient.PullRequests.CreateComment(ctx, pr.Owner, pr.RepoName, pr.Number, comment)
		if err != nil {
			r.logger.WarnContext(ctx, "failed to create comment", slog.String("file", result.File), slog.Int("line", result.Line), slog.String("err", err.Error()))
			continue
		}
	}
	return nil
}

func (r *riskenService) getAllComments(ctx context.Context, owner, repo string, prNumber int) ([]*github.PullRequestComment, error) {
	var allComments []*github.PullRequestComment
	opts := &github.PullRequestListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		comments, resp, err := r.githubClient.PullRequests.ListComments(ctx, owner, repo, prNumber, opts)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		time.Sleep(500 * time.Millisecond)
	}
	return allComments, nil
}

func existsSimilarComment(comments []*github.PullRequestComment, key string) bool {
	for _, c := range comments {
		if strings.Contains(*c.Body, key) {
			return true
		}
	}
	return false
}
