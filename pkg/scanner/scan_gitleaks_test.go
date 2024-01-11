package scanner

import (
	"testing"

	"github.com/ca-risken/code/pkg/gitleaks"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v44/github"
	"github.com/zricethezav/gitleaks/v8/report"
)

func TestGenerateScanResultFromGitleaksResults(t *testing.T) {
	type Args struct {
		repo           *github.Repository
		sourceCodePath string
		results        []report.Finding
	}
	tests := []struct {
		name string
		args *Args
		want []*ScanResult
	}{
		{
			name: "OK",
			args: &Args{
				repo: &github.Repository{
					HTMLURL:  github.String("https://github.com/owner/repo"),
					FullName: github.String("owner/repo"),
				},
				sourceCodePath: "/path/to/source",
				results: []report.Finding{
					{
						Description: "rule1",
						File:        "/path/to/source/file1.go",
						StartLine:   1,
						EndLine:     1,
						StartColumn: 1,
						EndColumn:   1,
						Match:       "REDACTED",
						Commit:      "commit",
						Author:      "author",
						Email:       "email",
						Message:     "message",
					},
				},
			},
			want: []*ScanResult{
				{
					ScanID:   "rule1",
					File:     "file1.go",
					Line:     1,
					DiffHunk: "",
					ReviewComment: `
シークレット情報が含まれている可能性があります👀

#### シークレットスキャン結果

- シークレットタイプ: rule1
- 説明:
  対象データがテスト用のダミーデータや公開可能な情報であるか確認してください。
	もし、有効なシークレット情報の場合はキーのローテーションや権限の削除（キーの無効化）を行ってください。
	どうしてもシークレット情報をコミットする必要がある場合は、プライベートリポジトリになっていることを確認しアクセスできる人を限定してください。
`,
					GitHubURL: "https://github.com/owner/repo/blob/commit//path/to/source/file1.go#L1-L1",
					ScanResult: &gitleaks.GitleaksFinding{
						RepositoryMetadata: &gitleaks.RepositoryMetadata{FullName: github.String("owner/repo")},
						Result: &gitleaks.LeakFinding{
							DataSourceID:    "cfb3a824c546bd5fd24a2e7450a4a08e8fb5c7bc682262ab4e36cfdc6b9d570f",
							File:            "/path/to/source/file1.go",
							StartLine:       1,
							EndLine:         1,
							StartColumn:     1,
							Secret:          "",
							Commit:          "commit",
							Repo:            "owner/repo",
							RuleDescription: "rule1",
							Author:          "author",
							Email:           "email",
							Message:         "message",
							URL:             "https://github.com/owner/repo/blob/commit//path/to/source/file1.go#L1-L1",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateScanResultFromGitleaksResults(tt.args.repo, tt.args.sourceCodePath, tt.args.results)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("generateScanResultFromGitleaksResults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
