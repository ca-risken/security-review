package review

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewReviewService(t *testing.T) {
	ctx := context.Background()
	type Args struct {
		opt    *ReviewOption
		logger *slog.Logger
	}
	testCases := []struct {
		name             string
		args             *Args
		wantOpt          *ReviewOption
		wantRiskenClient bool
		wantGithubClient bool
	}{
		{
			name: "OK",
			args: &Args{
				opt: &ReviewOption{
					GithubToken:       "github_token",
					RiskenApiEndpoint: "http://risken.api",
					RiskenApiToken:    "risken_api_token",
					RiskenConsoleURL:  "http://risken.console",
				},
				logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
			},
			wantOpt: &ReviewOption{
				GithubToken:       "github_token",
				RiskenApiEndpoint: "http://risken.api",
				RiskenApiToken:    "risken_api_token",
				RiskenConsoleURL:  "http://risken.console",
			},
			wantRiskenClient: true,
			wantGithubClient: true,
		},
		{
			name: "OK (No Risken console URL)",
			args: &Args{
				opt: &ReviewOption{
					GithubToken:       "github_token",
					RiskenApiEndpoint: "http://risken.api",
					RiskenApiToken:    "risken_api_token",
				},
				logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
			},
			wantOpt: &ReviewOption{
				GithubToken:       "github_token",
				RiskenApiEndpoint: "http://risken.api",
				RiskenApiToken:    "risken_api_token",
				RiskenConsoleURL:  "http://risken.api",
			},
			wantRiskenClient: true,
			wantGithubClient: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewReviewService(ctx, tc.args.opt, tc.args.logger).(*reviewService)

			if (service.githubClient != nil) != tc.wantGithubClient {
				t.Errorf("NewReviewService() GitHubClient = %v, want %v", service.githubClient != nil, tc.wantGithubClient)
			}
			if (service.riskenClient != nil) != tc.wantRiskenClient {
				t.Errorf("NewReviewService() RiskenClient = %v, want %v", service.riskenClient != nil, tc.wantRiskenClient)
			}
			if diff := cmp.Diff(service.opt, tc.wantOpt); diff != "" {
				t.Errorf("NewReviewService() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
