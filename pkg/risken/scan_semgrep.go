package risken

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/google/go-github/v57/github"
)

type SemgrepScanner struct {
	logger *slog.Logger
}

func NewSemgrepScanner(logger *slog.Logger) Scanner {
	return &SemgrepScanner{
		logger: logger,
	}
}

func (s *SemgrepScanner) Scan(ctx context.Context, repositoryURL, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error) {
	var semgrepFindings []*semgrepFinding
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
		findings, err := parseSemgrepResult(sourceCodePath, stdout.String(), changeFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to parse semgrep: targetPath=%s, err=%w", targetPath, err)
		}
		semgrepFindings = append(semgrepFindings, findings...)
	}
	return generateScanResultFromSemgrepResults(repositoryURL, semgrepFindings), nil
}

func parseSemgrepResult(sourceCodePath, scanResult string, changeFiles []*github.CommitFile) ([]*semgrepFinding, error) {
	var results semgrepResults
	err := json.Unmarshal([]byte(scanResult), &results)
	if err != nil {
		return nil, err
	}
	findings := make([]*semgrepFinding, 0, len(results.Results))
	for _, r := range results.Results {
		if !isChangeLine(changeFiles, r.Extra.Lines) {
			continue
		}
		r.Path = strings.ReplaceAll(r.Path, sourceCodePath+"/", "") // remove dir prefix
		findings = append(findings, r)
	}
	return findings, nil
}

type semgrepResults struct {
	Results []*semgrepFinding `json:"results,omitempty"`
}

type semgrepFinding struct {
	CheckID string        `json:"check_id,omitempty"`
	Path    string        `json:"path,omitempty"`
	Start   *semgrepLine  `json:"start,omitempty"`
	End     *semgrepLine  `json:"end,omitempty"`
	Extra   *semgrepExtra `json:"extra,omitempty"`
}

type semgrepLine struct {
	Line   int `json:"line,omitempty"`
	Column int `json:"col,omitempty"`
	Offset int `json:"offset,omitempty"`
}
type semgrepExtra struct {
	Lines    string `json:"lines,omitempty"`
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
	Metadata any    `json:"metadata,omitempty"`
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

func generateScanResultFromSemgrepResults(repositoryURL string, results []*semgrepFinding) []*ScanResult {
	var scanResults []*ScanResult
	for _, r := range results {
		tech := getSemgrepTechnology(r.Extra.Metadata)
		if !isSupportedResult(tech) {
			continue
		}
		scanResults = append(scanResults, &ScanResult{
			File:          r.Path,
			Line:          r.End.Line,
			ReviewComment: generateSemgrepReviewComment(r, tech[0]),
			GitHubURL:     generateGitHubURLForSemgrep(repositoryURL, r),
			ScanResult:    r,
		})
	}
	return scanResults
}

func isSupportedResult(tech []string) bool {
	for _, t := range tech {
		if t == "secret" || t == "secrets" {
			return false // シークレットスキャンはGitleaksで行う（Semgrepは過検知も多いし、カバレッジも低い）
		}
	}
	return true
}

func generateGitHubURLForSemgrep(repositoryURL string, f *semgrepFinding) string {
	return fmt.Sprintf("%s/blob/master/%s#L%d-L%d", repositoryURL, f.Path, f.Start.Line, f.End.Line)
}

func getSemgrepTechnology(metadata interface{}) []string {
	var semgrepMetadata semgrepMetadata
	b, err := json.Marshal(metadata)
	if err != nil {
		return []string{}
	}
	err = json.Unmarshal(b, &semgrepMetadata)
	if err != nil {
		return []string{}
	}
	return semgrepMetadata.Technology
}

const (
	SEMGREP_REVIEW_COMMENT_TEMPLATE = `
問題のコードを発見しました。修正が必要か確認してください🙏

#### コードスキャン結果

- Semgrep CheckID: %s
- 説明: %s
- 問題の行:

` + "```" + `%s
%s
` + "```" + `
`
)

func generateSemgrepReviewComment(f *semgrepFinding, tech string) string {
	return fmt.Sprintf(SEMGREP_REVIEW_COMMENT_TEMPLATE, f.CheckID, f.Extra.Message, tech, f.Extra.Lines)
}
