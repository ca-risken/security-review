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
	Use:   "risken-review --github-event-path <path> --github-token <token> --github-workspace <path> [--error --risken-endpoint <endpoint>] [--risken-api-token <token>]",
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
	rootCmd.PersistentFlags().BoolVar(&opt.ErrorFlag, "error", false, "Exit 1 if there are findings")
	rootCmd.PersistentFlags().StringVar(&opt.RiskenEndpoint, "risken-endpoint", "", "RISKEN API endpoint")
	rootCmd.PersistentFlags().StringVar(&opt.RiskenApiToken, "risken-api-token", "", "RISKEN API token for authentication")

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
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}
