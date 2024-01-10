package review

import (
	"context"
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
}

type Scanner interface {
	Scan(ctx context.Context, repo *github.Repository, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error)
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
	return strings.ReplaceAll(path, dir+"/", "")
}
