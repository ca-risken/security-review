package review

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/ca-risken/code/pkg/codescan"
	"github.com/google/go-github/v44/github"
)

type SemgrepScanner struct {
	logger *slog.Logger
}

func NewSemgrepScanner(logger *slog.Logger) Scanner {
	return &SemgrepScanner{
		logger: logger,
	}
}

func (s *SemgrepScanner) Scan(ctx context.Context, repo *github.Repository, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error) {
	var semgrepFindings []*codescan.SemgrepFinding
	for _, file := range changeFiles {
		targetPath := fmt.Sprintf("%s/%s", sourceCodePath, *file.Filename)
		cmd := exec.CommandContext(ctx,
			"semgrep",
			"scan",
			"--metrics=off",
			"--timeout=60",
			"--config=p/default",
			"--json",
			targetPath,
		)
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		s.logger.InfoContext(ctx, "Start semgrep scan", slog.String("file", targetPath))

		err := cmd.Run()
		if err != nil {
			return nil, fmt.Errorf("failed to execute semgrep: targetPath=%s, err=%w, stderr=%+v", targetPath, err, stderr.String())
		}
		findings, err := parseSemgrepResult(sourceCodePath, stdout.String(), repo, changeFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to parse semgrep: targetPath=%s, err=%w", targetPath, err)
		}
		semgrepFindings = append(semgrepFindings, findings...)
	}
	return generateScanResultFromSemgrepResults(repo, semgrepFindings), nil
}

func parseSemgrepResult(sourceCodePath, scanResult string, repo *github.Repository, changeFiles []*github.CommitFile) ([]*codescan.SemgrepFinding, error) {
	results, err := codescan.ParseSemgrepResult(sourceCodePath, scanResult, *repo.FullName, *repo.HTMLURL)
	if err != nil {
		return nil, err
	}

	findings := []*codescan.SemgrepFinding{}
	for _, r := range results {
		fileName := strings.ReplaceAll(r.Path, sourceCodePath+"/", "") // remove dir prefix
		if !isChangeLine(changeFiles, fileName, r.Extra.Lines) {
			continue
		}
		tech := getSemgrepTechnology(r.Extra.Metadata)
		log.Println(tech)
		if !isSupportedResult(tech) {
			continue
		}
		r.Path = fileName
		findings = append(findings, r)
	}
	return findings, nil
}

type semgrepMetadata struct {
	Category    string   `json:"category,omitempty"`
	Refences    []string `json:"references,omitempty"`
	Technology  []string `json:"technology,omitempty"`
	Confidence  string   `json:"confidence,omitempty"`
	Likelihood  string   `json:"likelihood,omitempty"`
	Impact      string   `json:"impact,omitempty"`
	Subcategory []string `json:"subcategory,omitempty"`
	CWE         []string `json:"cwe,omitempty"`
}

func parseSemgrepMetadata(metadata interface{}) (*semgrepMetadata, error) {
	var semgrepMetadata semgrepMetadata
	b, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &semgrepMetadata)
	if err != nil {
		return nil, err
	}
	return &semgrepMetadata, nil
}

func getSemgrepTechnology(metadata interface{}) []string {
	parseMeta, err := parseSemgrepMetadata(metadata)
	if err != nil {
		return []string{}
	}
	return parseMeta.Technology
}

func isSupportedResult(tech []string) bool {
	for _, t := range tech {
		if t == "secret" || t == "secrets" {
			return false // シークレットスキャンはGitleaksで行う（Semgrepは過検知も多いし、カバレッジも低い）
		}
	}
	return true
}

func generateScanResultFromSemgrepResults(repo *github.Repository, results []*codescan.SemgrepFinding) []*ScanResult {
	var scanResults []*ScanResult
	for _, r := range results {
		scanResults = append(scanResults, &ScanResult{
			ScanID:        r.CheckID,
			File:          r.Path,
			Line:          r.End.Line,
			DiffHunk:      r.Extra.Lines,
			ReviewComment: generateSemgrepReviewComment(r),
			GitHubURL:     generateGitHubURLForSemgrep(*repo.HTMLURL, r),
			ScanResult:    r,
		})
	}
	return scanResults
}

func generateGitHubURLForSemgrep(repositoryURL string, f *codescan.SemgrepFinding) string {
	return fmt.Sprintf("%s/blob/master/%s#L%d-L%d", repositoryURL, f.Path, f.Start.Line, f.End.Line)
}

const (
	SEMGREP_REVIEW_COMMENT_TEMPLATE = `
問題のコードを発見しました。修正が必要か確認してください🙏

#### コードスキャン結果

- Semgrep CheckID:
	%s
- 説明:
	%s
`
)

func generateSemgrepReviewComment(f *codescan.SemgrepFinding) string {
	return fmt.Sprintf(SEMGREP_REVIEW_COMMENT_TEMPLATE, f.CheckID, f.Extra.Message)
}
