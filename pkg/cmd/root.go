package cmd

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/ca-risken/security-review/pkg/risken"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "risken-review --github-event-path <path> --github-token <token> --github-workspace <path> [--risken-endpoint <endpoint>] [--risken-api-token <token>]",
	Short: "risken-review command is a GitHub Custom Action to review pull request with Risken",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
		defer cancel()
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		riskenService := risken.NewRiskenService(ctx, &conf, logger)
		return riskenService.Run(ctx)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

var conf risken.RiskenConfig

func init() {
	rootCmd.PersistentFlags().StringVar(&conf.GithubToken, "github-token", "", "GitHub token")
	rootCmd.PersistentFlags().StringVar(&conf.GithubEventPath, "github-event-path", "", "GitHub event path")
	rootCmd.PersistentFlags().StringVar(&conf.GithubWorkspace, "github-workspace", "", "GitHub workspace path")
	rootCmd.PersistentFlags().StringVar(&conf.RiskenEndpoint, "risken-endpoint", "", "RISKEN API endpoint")
	rootCmd.PersistentFlags().StringVar(&conf.RiskenApiToken, "risken-api-token", "", "RISKEN API token for authentication")

	// rootCmd.MarkPersistentFlagRequired("github-event-path")
	// rootCmd.MarkPersistentFlagRequired("github-token")
	// rootCmd.MarkPersistentFlagRequired("github-workspace")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Try to get values from environment variables
	// https://docs.github.com/ja/actions/learn-github-actions/variables#default-environment-variables
	if conf.GithubToken == "" {
		conf.GithubToken = getEnv("GITHUB_TOKEN")
	}
	if conf.GithubEventPath == "" {
		conf.GithubEventPath = getEnv("GITHUB_EVENT_PATH")
	}
	if conf.GithubWorkspace == "" {
		conf.GithubWorkspace = getEnv("GITHUB_WORKSPACE")
	}
	if conf.GithubToken == "" || conf.GithubEventPath == "" || conf.GithubWorkspace == "" {
		log.Fatal("Missing required parameters")
	}
}

func getEnv(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return ""
}
