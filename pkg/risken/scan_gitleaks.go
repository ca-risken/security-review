package risken

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v57/github"
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

func (s *GitleaksScanner) Scan(ctx context.Context, repositoryURL, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error) {
	d, err := detect.NewDetectorDefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize detector: %w", err)
	}

	gitleaksFindings := []report.Finding{}
	for _, file := range changeFiles {
		targetPath := fmt.Sprintf("%s/%s", sourceCodePath, *file.Filename)
		findings, err := d.DetectFiles(targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to detect %s: %w", targetPath, err)
		}
		for _, f := range findings {

			if strings.Contains(*file.Patch, f.Line) {
				gitleaksFindings = append(gitleaksFindings, findings...)
			}
		}
	}
	return generateScanResultFromGitleaksResults(repositoryURL, sourceCodePath, gitleaksFindings), nil
}

func generateScanResultFromGitleaksResults(repositoryURL, sourceCodePath string, results []report.Finding) []*ScanResult {
	var scanResults []*ScanResult
	for _, r := range results {
		scanResults = append(scanResults, &ScanResult{
			File:          strings.ReplaceAll(r.File, sourceCodePath+"/", ""), // remove dir prefix
			Line:          r.EndLine,
			ReviewComment: generateGitleaksReviewComment(&r),
			GitHubURL:     generateGitHubURLForGitleaks(repositoryURL, &r),
			ScanResult:    r,
		})
	}
	return scanResults
}

func generateGitHubURLForGitleaks(repositoryURL string, f *report.Finding) string {
	return fmt.Sprintf("%s/blob/master/%s#L%d-L%d", repositoryURL, f.File, f.StartLine, f.EndLine)
}

const (
	GITLEAKS_REVIEW_COMMENT_TEMPLATE = `
シークレット情報が含まれている可能性があります👀

#### シークレットスキャン結果

- Gitleaks RuleID: %s
- シークレットタイプ: %s
`
)

func generateGitleaksReviewComment(f *report.Finding) string {
	return fmt.Sprintf(GITLEAKS_REVIEW_COMMENT_TEMPLATE, f.RuleID, f.Description)
}
