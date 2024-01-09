package review

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type ReviewService interface {
	Run(ctx context.Context) error
}

type ReviewOption struct {
	GithubToken     string
	GithubEventPath string
	GithubWorkspace string
	ErrorFlag       bool
	RiskenEndpoint  string
	RiskenApiToken  string
}

type reviewService struct {
	opt          *ReviewOption
	githubClient *github.Client
	logger       *slog.Logger
}

func NewReviewService(ctx context.Context, opt *ReviewOption, logger *slog.Logger) ReviewService {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opt.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &reviewService{
		opt:          opt,
		githubClient: client,
		logger:       logger,
	}
}

func (r *reviewService) Run(ctx context.Context) error {
	// PR情報を取得（なければ終了）
	pr, err := r.GetGithubPREvent()
	if err != nil {
		r.logger.WarnContext(ctx, "Failed to get PR info.", slog.String("err", err.Error()))
		return nil
	}
	if pr == nil || pr.PullRequest == nil {
		r.logger.WarnContext(ctx, "PR info is nil.")
		return nil
	}

	// ソースコードの差分を取得
	changeFiles, err := r.Diff(ctx, pr)
	if err != nil {
		return err
	}

	// スキャン
	semgrep := NewSemgrepScanner(r.logger)
	semgrepResults, err := semgrep.Scan(ctx, *pr.Repository.HTMLURL, r.opt.GithubWorkspace, changeFiles)
	if err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success semgrep scan", slog.Int("results", len(semgrepResults)))

	gitleaks := NewGitleaksScanner(r.logger)
	gitleaksResults, err := gitleaks.Scan(ctx, *pr.Repository.HTMLURL, r.opt.GithubWorkspace, changeFiles)
	if err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success gitleaks scan", slog.Int("results", len(gitleaksResults)))
	scanResult := append(semgrepResults, gitleaksResults...)

	// RISKNEN APIを叩く(optional)

	// PRコメント
	if err := r.PullRequestComment(ctx, pr, scanResult); err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success PR comment")

	if r.opt.ErrorFlag && len(scanResult) > 0 {
		return fmt.Errorf("there are findings(%d)", len(scanResult))
	}
	return nil
}
