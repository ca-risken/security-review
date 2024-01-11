package review

import (
	"context"
	"time"

	"github.com/google/go-github/v44/github"
	"golang.org/x/oauth2"
)

// GitHubClient github.Client をラップしたInterfaceを定義（テストのためにそのままのクライアントの使用は避ける）
type GitHubClient interface {
	ListFiles(ctx context.Context, owner string, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error)
	GetAllIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*github.IssueComment, error)
	GetAllPRComments(ctx context.Context, owner, repo string, prNumber int) ([]*github.PullRequestComment, error)
	CreateIssueComment(ctx context.Context, owner, repoName string, prNumber int, comment *github.IssueComment) error
	CreatePRComment(ctx context.Context, owner, repoName string, prNumber int, comment *github.PullRequestComment) error
}

type githubClient struct {
	*github.Client
}

func NewGitHubClient(ctx context.Context, token string) GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &githubClient{
		github.NewClient(tc),
	}
}

func (c *githubClient) ListFiles(ctx context.Context, owner string, repo string, number int, opts *github.ListOptions) ([]*github.CommitFile, *github.Response, error) {
	files, resp, err := c.PullRequests.ListFiles(ctx, owner, repo, number, opts)
	if err != nil {
		return nil, resp, err
	}
	return files, resp, nil
}

func (c *githubClient) GetAllIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]*github.IssueComment, error) {
	var allComments []*github.IssueComment
	opts := &github.IssueListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {
		comments, resp, err := c.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
		if err != nil {
			return nil, err
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return allComments, nil
}

func (c *githubClient) GetAllPRComments(ctx context.Context, owner, repo string, prNumber int) ([]*github.PullRequestComment, error) {
	var allComments []*github.PullRequestComment
	opts := &github.PullRequestListCommentsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for {
		comments, resp, err := c.PullRequests.ListComments(ctx, owner, repo, prNumber, opts)
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

func (c *githubClient) CreateIssueComment(ctx context.Context, owner, repoName string, prNumber int, comment *github.IssueComment) error {
	_, _, err := c.Issues.CreateComment(ctx, owner, repoName, prNumber, comment)
	return err
}

func (c *githubClient) CreatePRComment(ctx context.Context, owner, repoName string, prNumber int, comment *github.PullRequestComment) error {
	_, _, err := c.PullRequests.CreateComment(ctx, owner, repoName, prNumber, comment)
	return err
}
