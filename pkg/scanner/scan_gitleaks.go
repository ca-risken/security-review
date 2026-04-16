package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ca-risken/code/pkg/gitleaks"
	"github.com/google/go-github/v44/github"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
)

type GitleaksScanner struct {
	logger *slog.Logger
}

func NewGitleaksScanner(logger *slog.Logger) Scanner {
	return &GitleaksScanner{
		logger: logger,
	}
}

func (s *GitleaksScanner) Scan(ctx context.Context, repo *github.Repository, pr *github.PullRequest, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error) {
	d, err := detect.NewDetectorDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detector: %w", err)
	}

	gitleaksFindings := []report.Finding{}
	for _, file := range changeFiles {
		targetPath := fmt.Sprintf("%s/%s", sourceCodePath, *file.Filename)
		info, err := os.Stat(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to stat %s: %w", targetPath, err)
		}
		if info.Size() == 0 {
			s.logger.InfoContext(ctx, "Skip gitleaks scan for empty file", slog.String("file", targetPath))
			continue
		}
		findings, err := d.DetectFiles(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to detect %s: %w", targetPath, err)
		}
		for _, f := range findings {
			if isLineInDiff(file, f.Match) {
				gitleaksFindings = append(gitleaksFindings, f)
			}
		}
	}
	return generateScanResultFromGitleaksResults(repo, sourceCodePath, gitleaksFindings), nil
}

func generateScanResultFromGitleaksResults(repo *github.Repository, sourceCodePath string, results []report.Finding) []*ScanResult {
	gitleaksFinding := gitleaks.GenrateGitleaksFinding(repo, results)

	var scanResults []*ScanResult
	for _, g := range gitleaksFinding {
		scanResults = append(scanResults, &ScanResult{
			ScanID:        g.Result.RuleDescription,
			File:          removeDirPrefix(sourceCodePath, g.Result.File),
			Line:          g.Result.EndLine,
			DiffHunk:      g.Result.Secret,
			ReviewComment: generateGitleaksReviewComment(sourceCodePath, g.Result),
			GitHubURL:     g.Result.GenerateGitHubURL(*repo.HTMLURL),
			ScanResult:    g,
		})
	}
	return scanResults
}

const (
	GITLEAKS_REVIEW_COMMENT_TEMPLATE = `
シークレット情報が含まれている可能性があります👀

#### シークレットスキャン結果

- シークレットタイプ: %s
- 説明:
  対象データがテスト用のダミーデータや公開可能な情報であるか確認してください。
	もし、有効なシークレット情報の場合はキーのローテーションや権限の削除（キーの無効化）を行ってください。
	どうしてもシークレット情報をコミットする必要がある場合は、プライベートリポジトリになっていることを確認しアクセスできる人を限定してください。
`
)

func generateGitleaksReviewComment(sourceCodePath string, f *gitleaks.LeakFinding) string {
	return fmt.Sprintf(GITLEAKS_REVIEW_COMMENT_TEMPLATE, f.RuleDescription)
}
