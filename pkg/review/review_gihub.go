package review

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/ca-risken/security-review/pkg/scanner"
	"github.com/google/go-github/v44/github"
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

func (r *reviewService) GetGithubPREvent() (*GithubPREvent, error) {
	file, err := os.Open(r.opt.GithubEventPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: path=%s, err=%w", r.opt.GithubEventPath, err)
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

func (r *reviewService) ListPRFiles(ctx context.Context, pr *GithubPREvent) ([]*github.CommitFile, error) {
	opts := &github.ListOptions{
		PerPage: 100,
	}
	changeFiles := []*github.CommitFile{}
	for {
		files, resp, err := r.githubClient.ListFiles(ctx, pr.Owner, pr.RepoName, pr.Number, opts)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			// https://docs.github.com/ja/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests-files
			if f.Status != nil && *f.Status == *github.String("removed") {
				// Can not scan removed files
				continue
			}
			changeFiles = append(changeFiles, f)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
		time.Sleep(1 * time.Second)
	}
	return changeFiles, nil
}

const (
	NO_REVIEW_COMMENT = "セキュリティレビューを実施しました。\n特に問題は見つかりませんでした👏\n\n_By RISKEN review_"
)

func (r *reviewService) PullRequestComment(ctx context.Context, pr *GithubPREvent, scanResults []*scanner.ScanResult) error {
	// No Review Comment
	if len(scanResults) == 0 {
		comments, err := r.githubClient.GetAllIssueComments(ctx, pr.Owner, pr.RepoName, pr.Number)
		if err != nil {
			return fmt.Errorf("failed to get all issue comments: err=%w", err)
		}
		if existsSimilarIssueComment(comments, "特に問題は見つかりませんでした") {
			r.logger.WarnContext(ctx, "already exists similar issue comment")
			return nil
		}

		comment := &github.IssueComment{
			Body: github.String(NO_REVIEW_COMMENT),
		}
		if err = r.githubClient.CreateIssueComment(ctx, pr.Owner, pr.RepoName, pr.Number, comment); err != nil {
			return fmt.Errorf("failed to create comment: err=%w", err)
		}
		return nil
	}

	// Review Comment
	comments, err := r.githubClient.GetAllPRComments(ctx, pr.Owner, pr.RepoName, pr.Number)
	if err != nil {
		return fmt.Errorf("failed to get all comments: err=%w", err)
	}
	for _, result := range scanResults {
		if existsSimilarPRComment(result, comments) {
			r.logger.WarnContext(ctx, "already exists similar comment", slog.String("file", result.File), slog.Int("line", result.Line), slog.String("ID", result.ScanID))
			continue
		}
		comment := &github.PullRequestComment{
			Body:     github.String(generatePRReviewComment(result)),
			CommitID: github.String(*pr.PullRequest.Head.SHA),
			Path:     github.String(result.File),
			Line:     github.Int(result.Line),
		}
		if err := r.githubClient.CreatePRComment(ctx, pr.Owner, pr.RepoName, pr.Number, comment); err != nil {
			r.logger.WarnContext(ctx, "failed to create comment", slog.String("file", result.File), slog.Int("line", result.Line), slog.String("err", err.Error()))
			continue
		}
	}
	return nil
}

func existsSimilarPRComment(scanResult *scanner.ScanResult, comments []*github.PullRequestComment) bool {
	for _, c := range comments {
		if c.Path == nil || c.Line == nil {
			continue
		}
		if strings.Contains(*c.Body, scanResult.ScanID) && *c.Path == scanResult.File && *c.Line == scanResult.Line {
			return true
		}
	}
	return false
}

func existsSimilarIssueComment(comments []*github.IssueComment, key string) bool {
	for _, c := range comments {
		if strings.Contains(*c.Body, key) {
			return true
		}
	}
	return false
}

const (
	RISKEN_COMMENT_TEMPLATE = `

#### RISKENで確認

より詳細な情報や生成AIによる解説はRISKENコンソール上で確認できます。

- %s`
)

func generatePRReviewComment(result *scanner.ScanResult) string {
	reviewComment := result.ReviewComment
	if result.RiskenURL != "" {
		reviewComment += fmt.Sprintf(RISKEN_COMMENT_TEMPLATE, result.RiskenURL)
	}
	reviewComment += "\n\n_By RISKEN review_"
	return reviewComment
}
