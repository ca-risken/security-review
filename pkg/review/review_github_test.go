package review

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ca-risken/security-review/pkg/mocks"
	"github.com/ca-risken/security-review/pkg/scanner"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v44/github"
	"github.com/stretchr/testify/mock"
)

func TestGetGithubPREvent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *GithubPREvent
		wantErr bool
	}{
		{
			name:    "OK",
			content: `{"action": "opened", "number": 1, "pull_request": {}, "repository": {"full_name": "owner/repo"}, "owner": "owner", "repo_name": "repo"}`,
			want:    &GithubPREvent{Action: "opened", Number: 1, PullRequest: &github.PullRequest{}, Repository: &github.Repository{FullName: github.String("owner/repo")}, Owner: "owner", RepoName: "repo"},
			wantErr: false,
		},
		{
			name:    "No repository",
			content: `{"action": "opened", "number": 1, "pull_request": {}, "owner": "owner", "repo_name": "repo"}`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid repository name",
			content: `{"action": "opened", "number": 1, "pull_request": {}, "repository": {"full_name": "owner#repo"}, "owner": "owner", "repo_name": "repo"}`,
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp file and write the content to it.
			tmpFile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %s", err)
			}
			defer os.Remove(tmpFile.Name())
			_, err = tmpFile.Write([]byte(tt.content))
			if err != nil {
				t.Fatalf("Failed to write to temp file: %s", err)
			}
			tmpFile.Close()

			r := reviewService{
				opt: &ReviewOption{GithubEventPath: tmpFile.Name()},
			}

			got, err := r.GetGithubPREvent()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGithubPREvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("GetGithubPREvent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestListPRFiles(t *testing.T) {
	ctx := context.Background()
	type MockResp struct {
		files []*github.CommitFile
		resp  *github.Response
		err   error
	}

	testCases := []struct {
		name     string
		args     *GithubPREvent
		mockResp *MockResp
		want     []*github.CommitFile
		wantErr  bool
	}{
		{
			name: "OK",
			args: &GithubPREvent{
				Owner:    "owner",
				RepoName: "repo",
				Number:   1,
			},
			mockResp: &MockResp{
				files: []*github.CommitFile{
					{Filename: github.String("file1.txt")},
					{Filename: github.String("file2.txt")},
				},
				resp: &github.Response{
					NextPage: 0,
				},
				err: nil,
			},
			want: []*github.CommitFile{
				{Filename: github.String("file1.txt")},
				{Filename: github.String("file2.txt")},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: &GithubPREvent{
				Owner:    "owner",
				RepoName: "repo",
				Number:   1,
			},
			mockResp: &MockResp{
				files: []*github.CommitFile{},
				resp:  nil,
				err:   errors.New("error"),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mocks.NewGitHubClient(t)
			mockClient.
				On("ListFiles", ctx, tc.args.Owner, tc.args.RepoName, tc.args.Number, mock.Anything).
				Return(tc.mockResp.files, tc.mockResp.resp, tc.mockResp.err).
				Once()

			service := &reviewService{
				githubClient: mockClient,
			}
			files, err := service.ListPRFiles(ctx, tc.args)
			if (err != nil) != tc.wantErr {
				t.Errorf("Client.Signin() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(files, tc.want); diff != "" {
				t.Errorf("ListPRFiles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPullRequestComment(t *testing.T) {
	ctx := context.Background()
	type MockRespIssueComments struct {
		comments []*github.IssueComment
		err      error
	}
	type MockRespPRComments struct {
		comments []*github.PullRequestComment
		err      error
	}
	type Args struct {
		pr          *GithubPREvent
		scanResults []*scanner.ScanResult
	}

	testCases := []struct {
		name                  string
		args                  *Args
		mockRespIssueComments *MockRespIssueComments
		mockRespPRComments    *MockRespPRComments
		wantErr               bool
	}{
		{
			name: "OK",
			args: &Args{
				pr: &GithubPREvent{
					Owner:    "owner",
					RepoName: "repo",
					Number:   1,
					PullRequest: &github.PullRequest{
						Head: &github.PullRequestBranch{
							SHA: github.String("sha"),
						},
					},
				},
				scanResults: []*scanner.ScanResult{
					{
						ScanID:        "scan_id",
						File:          "file1.txt",
						Line:          1,
						DiffHunk:      "diff_hunk",
						ReviewComment: "review_comment",
						GitHubURL:     "github_url",
						ScanResult:    "scan_result",
						RiskenURL:     "risken_url",
					},
				},
			},
			mockRespIssueComments: &MockRespIssueComments{comments: []*github.IssueComment{}},
			mockRespPRComments:    &MockRespPRComments{comments: []*github.PullRequestComment{}},
			wantErr:               false,
		},
		{
			name: "OK(no review comment))",
			args: &Args{
				pr: &GithubPREvent{
					Owner:    "owner",
					RepoName: "repo",
					Number:   1,
					PullRequest: &github.PullRequest{
						Head: &github.PullRequestBranch{
							SHA: github.String("sha"),
						},
					},
				},
				scanResults: []*scanner.ScanResult{},
			},
			mockRespIssueComments: &MockRespIssueComments{comments: []*github.IssueComment{}},
			mockRespPRComments:    &MockRespPRComments{comments: []*github.PullRequestComment{}},
			wantErr:               false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mocks.NewGitHubClient(t)
			if len(tc.args.scanResults) == 0 {
				// No review comment
				mockClient.
					On("GetAllIssueComments", ctx, tc.args.pr.Owner, tc.args.pr.RepoName, tc.args.pr.Number).
					Return(tc.mockRespIssueComments.comments, tc.mockRespIssueComments.err).Once()
				mockClient.
					On("CreateIssueComment", ctx, tc.args.pr.Owner, tc.args.pr.RepoName, tc.args.pr.Number, mock.Anything).
					Return(nil).Once()
			} else {
				// Review comment
				mockClient.
					On("GetAllPRComments", ctx, tc.args.pr.Owner, tc.args.pr.RepoName, tc.args.pr.Number).
					Return(tc.mockRespPRComments.comments, tc.mockRespPRComments.err).Once()
				mockClient.
					On("CreatePRComment", ctx, tc.args.pr.Owner, tc.args.pr.RepoName, tc.args.pr.Number, mock.Anything).
					Return(nil).Once()
			}

			service := &reviewService{
				githubClient: mockClient,
			}
			err := service.PullRequestComment(ctx, tc.args.pr, tc.args.scanResults)
			if (err != nil) != tc.wantErr {
				t.Errorf("PullRequestComment() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestExistsSimilarPRComment(t *testing.T) {
	testCases := []struct {
		name       string
		scanResult *scanner.ScanResult
		comments   []*github.PullRequestComment
		want       bool
	}{
		{
			name: "Similar Comment Exists",
			scanResult: &scanner.ScanResult{
				ScanID: "ID123",
				File:   "file.go",
				Line:   10,
			},
			comments: []*github.PullRequestComment{
				{
					Body: github.String("Issue found ID123"),
					Path: github.String("file.go"),
					Line: github.Int(10),
				},
			},
			want: true,
		},
		{
			name: "No Similar Comment",
			scanResult: &scanner.ScanResult{
				ScanID: "ID123",
				File:   "file.go",
				Line:   10,
			},
			comments: []*github.PullRequestComment{
				{
					Body: github.String("Different issue ID456"),
					Path: github.String("file.go"),
					Line: github.Int(10),
				},
			},
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := existsSimilarPRComment(tc.scanResult, tc.comments)
			if got != tc.want {
				t.Errorf("existsSimilarPRComment() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExistsSimilarIssueComment(t *testing.T) {
	testCases := []struct {
		name     string
		comments []*github.IssueComment
		key      string
		want     bool
	}{
		{
			name: "Comment Contains Key",
			comments: []*github.IssueComment{
				{Body: github.String("This is a comment containing the key: key123")},
				{Body: github.String("Another comment")},
			},
			key:  "key123",
			want: true,
		},
		{
			name: "No Comment Contains Key",
			comments: []*github.IssueComment{
				{Body: github.String("This is a comment")},
				{Body: github.String("Another comment")},
			},
			key:  "key123",
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := existsSimilarIssueComment(tc.comments, tc.key)
			if got != tc.want {
				t.Errorf("existsSimilarIssueComment() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGeneratePRReviewComment(t *testing.T) {
	testCases := []struct {
		name        string
		scanResult  *scanner.ScanResult
		wantComment string
	}{
		{
			name: "With RiskenURL",
			scanResult: &scanner.ScanResult{
				ReviewComment: "Initial review comment.",
				RiskenURL:     "http://example.com/details",
			},
			wantComment: `Initial review comment.

#### RISKENで確認

より詳細な情報や生成AIによる解説はRISKENコンソール上で確認できます。

- http://example.com/details

_By RISKEN review_`,
		},
		{
			name: "Without RiskenURL",
			scanResult: &scanner.ScanResult{
				ReviewComment: "Initial review comment.",
				RiskenURL:     "",
			},
			wantComment: `Initial review comment.

_By RISKEN review_`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := generatePRReviewComment(tc.scanResult)
			if got != tc.wantComment {
				t.Errorf("generatePRReviewComment() = %v, want %v", got, tc.wantComment)
			}
		})
	}
}
