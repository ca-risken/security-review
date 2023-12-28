package risken

import (
	"context"
	"log/slog"
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
	conf   *RiskenConfig
	logger *slog.Logger
}

func NewRiskenService(conf *RiskenConfig, logger *slog.Logger) RiskenService {
	return &riskenService{
		conf:   conf,
		logger: logger,
	}
}

func (r *riskenService) Run(ctx context.Context) error {
	// PR情報を取得（なければ終了）
	pr, err := r.GetGithubPREvent(ctx, r.conf.GithubEventPath)
	if err != nil {
		r.logger.WarnContext(ctx, "Failed to get PR info.", slog.String("err", err.Error()))
		return nil
	}
	if pr == nil || pr.PullRequest == nil {
		r.logger.WarnContext(ctx, "PR info is nil.")
		return nil
	}

	// ソースコードの差分を取得
	changeFiles, err := r.Diff(ctx, r.conf.GithubWorkspace, *pr.PullRequest)
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
	if err := r.PullRequestComment(ctx, pr.Repository, pr.PullRequest, scanResult); err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success PR comment")
	return nil
}
