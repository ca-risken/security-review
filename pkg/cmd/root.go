package cmd

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/ca-risken/security-review/pkg/review"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "risken-review",
	Short: "risken-review command is a GitHub Custom Action to review pull request with Risken",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
		defer cancel()
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		riskenService := review.NewReviewService(ctx, &opt, logger)
		return riskenService.Run(ctx)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

var opt review.ReviewOption

func init() {
	rootCmd.PersistentFlags().StringVar(&opt.GithubToken, "github-token", "", "GitHub token")
	rootCmd.PersistentFlags().StringVar(&opt.GithubEventPath, "github-event-path", "", "GitHub event path")
	rootCmd.PersistentFlags().StringVar(&opt.GithubWorkspace, "github-workspace", "", "GitHub workspace path")
	rootCmd.PersistentFlags().StringVar(&opt.RiskenConsoleURL, "risken-console-url", "", "RISKEN Console URL (optional)")
	rootCmd.PersistentFlags().StringVar(&opt.RiskenApiEndpoint, "risken-api-endpoint", "", "RISKEN API endpoint (optional)")
	rootCmd.PersistentFlags().StringVar(&opt.RiskenApiToken, "risken-api-token", "", "RISKEN API token for authentication (optional)")
	rootCmd.PersistentFlags().BoolVar(&opt.ErrorFlag, "error", false, "Exit 1 if there are findings (optional)")
	rootCmd.PersistentFlags().BoolVar(&opt.NoPRComment, "no-pr-comment", false, "If true, do not post PR comments (optional)")

	cobra.OnInitialize(initoptig)
}

func initoptig() {
	// Try to get values from environment variables
	// https://docs.github.com/ja/actions/learn-github-actions/variables#default-environment-variables
	if opt.GithubToken == "" {
		opt.GithubToken = getEnv("GITHUB_TOKEN")
	}
	if opt.GithubEventPath == "" {
		opt.GithubEventPath = getEnv("GITHUB_EVENT_PATH")
	}
	if opt.GithubWorkspace == "" {
		opt.GithubWorkspace = getEnv("GITHUB_WORKSPACE")
	}
	if opt.GithubToken == "" || opt.GithubEventPath == "" || opt.GithubWorkspace == "" {
		log.Fatal("Missing required parameters")
	}
	if opt.RiskenConsoleURL == "" {
		opt.RiskenConsoleURL = getEnv("RISKEN_CONSOLE_URL")
	}
	if opt.RiskenApiEndpoint == "" {
		opt.RiskenApiEndpoint = getEnv("RISKEN_API_ENDPOINT")
	}
	if opt.RiskenApiToken == "" {
		opt.RiskenApiToken = getEnv("RISKEN_API_TOKEN")
	}
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}
