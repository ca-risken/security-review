package review

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ca-risken/security-review/pkg/scanner"
)

type ReviewService interface {
	Run(ctx context.Context) error
}

type ReviewOption struct {
	GithubToken       string
	GithubEventPath   string
	GithubWorkspace   string
	RiskenConsoleURL  string
	RiskenApiEndpoint string
	RiskenApiToken    string
	ErrorFlag         bool
	NoPRComment       bool
}

type reviewService struct {
	opt          *ReviewOption
	githubClient GitHubClient
	riskenClient RiskenClient
	logger       *slog.Logger
}

func NewReviewService(ctx context.Context, opt *ReviewOption, logger *slog.Logger) ReviewService {
	var riskenClient RiskenClient
	if opt.RiskenApiEndpoint != "" && opt.RiskenApiToken != "" {
		riskenClient = NewRiskenClient(opt.RiskenApiToken, opt.RiskenApiEndpoint)
		if opt.RiskenConsoleURL == "" {
			logger.Info("RISKEN Console URL is not set. Use RISKEN API endpoint instead.")
			opt.RiskenConsoleURL = opt.RiskenApiEndpoint
		}
	}
	return &reviewService{
		opt:          opt,
		githubClient: NewGitHubClient(ctx, opt.GithubToken),
		riskenClient: riskenClient,
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
	changeFiles, err := r.ListPRFiles(ctx, pr)
	if err != nil {
		return err
	}

	// スキャン
	semgrep := scanner.NewSemgrepScanner(r.logger)
	semgrepResults, err := semgrep.Scan(ctx, pr.Repository, pr.PullRequest, r.opt.GithubWorkspace, changeFiles)
	if err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success semgrep scan", slog.Int("results", len(semgrepResults)))

	gitleaks := scanner.NewGitleaksScanner(r.logger)
	gitleaksResults, err := gitleaks.Scan(ctx, pr.Repository, pr.PullRequest, r.opt.GithubWorkspace, changeFiles)
	if err != nil {
		return err
	}
	r.logger.InfoContext(ctx, "Success gitleaks scan", slog.Int("results", len(gitleaksResults)))
	scanResult := append(semgrepResults, gitleaksResults...)

	// RISKNEN APIを叩く(optional)
	if r.riskenClient != nil {
		projectID, err := r.getProjectID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get project ID: %w", err)
		}
		for i, result := range scanResult {
			resp, err := r.putFinding(ctx, *projectID, result)
			if err != nil {
				return fmt.Errorf("failed to put finding: %w", err)
			}
			scanResult[i].RiskenURL = scanner.GenerateRiskenURL(r.opt.RiskenConsoleURL, resp.Finding.ProjectId, resp.Finding.FindingId)
		}
	} else {
		r.logger.InfoContext(ctx, "Skip RISKEN integration")
	}

	// PRコメント
	if r.opt.NoPRComment {
		r.logger.InfoContext(ctx, "Skip PR comment")
	} else {
		if err := r.PullRequestComment(ctx, pr, scanResult); err != nil {
			return err
		}
		r.logger.InfoContext(ctx, "Success PR comment")
	}

	if r.opt.ErrorFlag && len(scanResult) > 0 {
		return fmt.Errorf("there are findings(%d)", len(scanResult))
	}
	return nil
}
