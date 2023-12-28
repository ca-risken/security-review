package risken

import (
	"context"
	"strings"

	"github.com/google/go-github/v57/github"
)

type ScanResult struct {
	File          string
	Line          int
	ReviewComment string
	GitHubURL     string
	ScanResult    any
}

type Scanner interface {
	Scan(ctx context.Context, repositoryURL, sourceCodePath string, changeFiles []*github.CommitFile) ([]*ScanResult, error)
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
