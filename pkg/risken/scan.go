package risken

import "context"

type ScanResult struct {
	File          string
	Line          int
	ReviewComment string
	GitHubURL     string
	ScanResult    any
}

type Scanner interface {
	Scan(ctx context.Context, repositoryURL, sourceCodePath string, changeFiles []*string) ([]*ScanResult, error)
}
