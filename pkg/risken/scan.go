package risken

import (
	"context"

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
