package risken

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type RiskenService interface {
	Run(ctx context.Context) error
}

type RiskenConfig struct {
	GithubToken     string
	GithubEventPath string
	GithubWorkspace string
	RiskenEndpoint  string
	RiskenApiToken  string
}

type riskenService struct {
	conf         *RiskenConfig
	githubClient *github.Client
	logger       *slog.Logger
}

func NewRiskenService(ctx context.Context, conf *RiskenConfig, logger *slog.Logger) RiskenService {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: conf.GithubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &riskenService{
		conf:         conf,
		githubClient: client,
		logger:       logger,
	}
}

func (r *riskenService) Run(ctx context.Context) error {
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
	semgrepResults, err := semgrep.Scan(ctx, *pr.Repository.HTMLURL, r.conf.GithubWorkspace, changeFiles)
	if err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success semgrep scan", slog.Int("results", len(semgrepResults)))

	gitleaks := NewGitleaksScanner(r.logger)
	gitleaksResults, err := gitleaks.Scan(ctx, *pr.Repository.HTMLURL, r.conf.GithubWorkspace, changeFiles)
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
	return nil
}
