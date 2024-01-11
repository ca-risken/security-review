package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v44/github"
)

type ScanResult struct {
	ScanID        string
	File          string
	Line          int
	DiffHunk      string
	ReviewComment string
	GitHubURL     string
	ScanResult    any
	RiskenURL     string
}

type Scanner interface {
	Scan(ctx context.Context, repo *github.Repository, pr *github.PullRequest, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error)
}

func isChangeLine(files []*github.CommitFile, fileName, line string) bool {
	for _, f := range files {
		if *f.Filename != fileName {
			continue
		}
		if isLineInDiff(f, line) {
			return true
		}
	}
	return false
}

func isLineInDiff(file *github.CommitFile, line string) bool {
	patchLines := strings.Split(file.GetPatch(), "\n")
	for _, patchLine := range patchLines {
		// "+" で始まる行は追加された行を示します。
		if strings.HasPrefix(patchLine, "+") {
			if strings.Contains(patchLine, line) {
				return true
			}
		}
	}
	return false
}

func removeDirPrefix(dir, path string) string {
	if strings.HasPrefix(path, dir+"/") {
		return strings.TrimPrefix(path, dir+"/")
	}
	return path
}

func GenerateRiskenURL(riskenConsoleURL string, projectID uint32, findingID uint64) string {
	return fmt.Sprintf("%s/finding/finding/?from_score=0&status=0&project_id=%d&finding_id=%d", riskenConsoleURL, projectID, findingID)
}
