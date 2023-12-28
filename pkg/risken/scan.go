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

func isChangeLine(files []*github.CommitFile, line string) bool {
	for _, f := range files {
		if strings.Contains(*f.Patch, line) {
			return true
		}
	}
	return false
}
