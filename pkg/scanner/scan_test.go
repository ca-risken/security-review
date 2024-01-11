package scanner

import (
	"testing"

	"github.com/google/go-github/v44/github"
)

func TestIsChangeLine(t *testing.T) {
	testCases := []struct {
		name     string
		files    []*github.CommitFile
		fileName string
		line     string
		patch    string
		want     bool
	}{
		{
			name: "OK",
			files: []*github.CommitFile{
				{
					Filename: github.String("file1.go"),
					Patch: github.String(`diff --git file1.go file1.go
+line 1
+target line
-line 3
line 4`),
				},
			},
			fileName: "file1.go",
			line:     "target line",
			want:     true,
		},
		{
			name: "File not match",
			files: []*github.CommitFile{
				{
					Filename: github.String("file1.go"),
					Patch: github.String(`diff --git file1.go file1.go
+line 1
+target line
-line 3
line 4`),
				},
			},
			fileName: "unknown.go",
			line:     "target line",
			want:     false,
		},
		{
			name: "File not match",
			files: []*github.CommitFile{
				{
					Filename: github.String("file1.go"),
					Patch: github.String(`diff --git file1.go file1.go
+line 1
+target line
-line 3
line 4`),
				},
			},
			fileName: "file1.go",
			line:     "unknown line",
			want:     false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isChangeLine(tc.files, tc.fileName, tc.line)
			if got != tc.want {
				t.Errorf("isChangeLine() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRemoveDirPrefix(t *testing.T) {
	type Args struct {
		dir  string
		path string
	}
	testCases := []struct {
		name string
		args *Args
		want string
	}{
		{
			name: "Dir at Start",
			args: &Args{
				dir:  "dir",
				path: "dir/file.txt",
			},
			want: "file.txt",
		},
		{
			name: "Dir Not at Start",
			args: &Args{
				dir:  "dir",
				path: "some/dir/file.txt",
			},
			want: "some/dir/file.txt",
		},
		{
			name: "Dir Repeated in Path",
			args: &Args{
				dir:  "dir",
				path: "dir/dir/file.txt",
			},
			want: "dir/file.txt",
		},
		{
			name: "Empty Dir",
			args: &Args{
				dir:  "",
				path: "file.txt",
			},
			want: "file.txt",
		},
		{
			name: "Dir Equal to Path",
			args: &Args{
				dir:  "dir",
				path: "dir",
			},
			want: "dir",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := removeDirPrefix(tc.args.dir, tc.args.path)
			if got != tc.want {
				t.Errorf("removeDirPrefix() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerateRiskenURL(t *testing.T) {
	type Args struct {
		riskenConsoleURL string
		projectID        uint32
		findingID        uint64
	}
	tests := []struct {
		name        string
		args        *Args
		expectedURL string
	}{
		{
			name: "OK",
			args: &Args{
				riskenConsoleURL: "https://example.com",
				projectID:        123,
				findingID:        456,
			},
			expectedURL: "https://example.com/finding/finding/?from_score=0&status=0&project_id=123&finding_id=456",
		},
		{
			name: "Empty URL",
			args: &Args{
				riskenConsoleURL: "",
				projectID:        123,
				findingID:        456,
			},
			expectedURL: "/finding/finding/?from_score=0&status=0&project_id=123&finding_id=456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRiskenURL(tt.args.riskenConsoleURL, tt.args.projectID, tt.args.findingID)
			if result != tt.expectedURL {
				t.Errorf("GenerateRiskenURL() = %q, want %q", result, tt.expectedURL)
			}
		})
	}
}
