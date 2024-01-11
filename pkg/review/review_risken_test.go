package review

import (
	"context"
	"errors"
	"testing"

	"github.com/ca-risken/code/pkg/codescan"
	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/go-risken"
	"github.com/ca-risken/security-review/pkg/mocks"
	"github.com/ca-risken/security-review/pkg/scanner"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"
)

// Helper function to get a pointer to a uint32 value
func uint32Ptr(val uint32) *uint32 {
	return &val
}

func TestGetProjectID(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name           string
		mockSigninResp *risken.SigninResponse
		mockSigninErr  error
		want           *uint32
		wantErr        bool
	}{
		{
			name: "Successful Signin",
			mockSigninResp: &risken.SigninResponse{
				ProjectID: 123,
			},
			mockSigninErr: nil,
			want:          uint32Ptr(123),
			wantErr:       false,
		},
		{
			name:           "Signin Error",
			mockSigninResp: nil,
			mockSigninErr:  errors.New("signin error"),
			want:           nil,
			wantErr:        true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRiskenClient := mocks.NewRiskenClient(t)
			mockRiskenClient.
				On("Signin", ctx).
				Return(tc.mockSigninResp, tc.mockSigninErr).
				Once()

			service := &reviewService{
				riskenClient: mockRiskenClient,
			}

			gotProjectID, err := service.getProjectID(ctx)
			if (err != nil) != tc.wantErr {
				t.Fatalf("getProjectID() error = %v, wantErr %v", err, tc.wantErr)
			}
			if diff := cmp.Diff(gotProjectID, tc.want); diff != "" {
				t.Errorf("getProjectID() %s", diff)
			}
		})
	}
}

func TestPutFinding(t *testing.T) {
	ctx := context.Background()
	type Args struct {
		projectID  uint32
		scanResult *scanner.ScanResult
	}

	type testCase struct {
		name      string
		args      *Args
		setupMock func(*mocks.RiskenClient)
		want      *finding.PutFindingResponse
		wantErr   bool
	}

	testCases := []testCase{
		{
			name: "OK",
			args: &Args{
				projectID: 123,
				scanResult: &scanner.ScanResult{
					ScanID:        "scanID",
					File:          "file",
					Line:          1,
					DiffHunk:      "diffHunk",
					ReviewComment: "reviewComment",
					GitHubURL:     "githubURL",
					RiskenURL:     "riskenURL",
					ScanResult: &codescan.SemgrepFinding{
						Repository:     "repository",
						RepoVisibility: "public",
						Path:           "path",
						CheckID:        "checkID",
						GitHubURL:      "githubURL",
						Start: &codescan.SemgrepLine{
							Line:   1,
							Column: 1,
						},
						End: &codescan.SemgrepLine{
							Line:   1,
							Column: 1,
						},
						Extra: &codescan.SemgrepExtra{
							Severity: "high",
							Message:  "message",
							Lines:    "lines",
						},
					},
				},
			},
			setupMock: func(m *mocks.RiskenClient) {
				m.On("PutFinding", ctx, mock.Anything, mock.Anything).
					Return(&finding.PutFindingResponse{
						Finding: &finding.Finding{FindingId: 1},
					}, nil).Once()
				m.On("PutRecommend", ctx, mock.Anything).
					Return(&finding.PutRecommendResponse{
						Recommend: &finding.Recommend{RecommendId: 1},
					}, nil).
					Once()
			},
			want: &finding.PutFindingResponse{
				Finding: &finding.Finding{FindingId: 1},
			},
			wantErr: false,
		},

		{
			name: "Error Unknown ScanResult Type",
			args: &Args{
				projectID: 123,
				scanResult: &scanner.ScanResult{
					ScanResult: `{"test":true}`,
				},
			},
			setupMock: func(m *mocks.RiskenClient) {},
			want:      nil,
			wantErr:   true,
		},
		{
			name: "Error PutFinding",
			args: &Args{
				projectID: 123,
				scanResult: &scanner.ScanResult{
					ScanID:        "scanID",
					File:          "file",
					Line:          1,
					DiffHunk:      "diffHunk",
					ReviewComment: "reviewComment",
					GitHubURL:     "githubURL",
					RiskenURL:     "riskenURL",
					ScanResult: &codescan.SemgrepFinding{
						Repository:     "repository",
						RepoVisibility: "public",
						Path:           "path",
						CheckID:        "checkID",
						GitHubURL:      "githubURL",
						Start: &codescan.SemgrepLine{
							Line:   1,
							Column: 1,
						},
						End: &codescan.SemgrepLine{
							Line:   1,
							Column: 1,
						},
						Extra: &codescan.SemgrepExtra{
							Severity: "high",
							Message:  "message",
							Lines:    "lines",
						},
					},
				},
			},
			setupMock: func(m *mocks.RiskenClient) {
				m.On("PutFinding", ctx, mock.Anything).
					Return(nil, errors.New("put finding error")).
					Once()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error PutRecommend",
			args: &Args{
				projectID: 123,
				scanResult: &scanner.ScanResult{
					ScanID:        "scanID",
					File:          "file",
					Line:          1,
					DiffHunk:      "diffHunk",
					ReviewComment: "reviewComment",
					GitHubURL:     "githubURL",
					RiskenURL:     "riskenURL",
					ScanResult: &codescan.SemgrepFinding{
						Repository:     "repository",
						RepoVisibility: "public",
						Path:           "path",
						CheckID:        "checkID",
						GitHubURL:      "githubURL",
						Start: &codescan.SemgrepLine{
							Line:   1,
							Column: 1,
						},
						End: &codescan.SemgrepLine{
							Line:   1,
							Column: 1,
						},
						Extra: &codescan.SemgrepExtra{
							Severity: "high",
							Message:  "message",
							Lines:    "lines",
						},
					},
				},
			},
			setupMock: func(m *mocks.RiskenClient) {
				m.On("PutFinding", ctx, mock.Anything, mock.Anything).
					Return(&finding.PutFindingResponse{
						Finding: &finding.Finding{FindingId: 1},
					}, nil).Once()
				m.On("PutRecommend", ctx, mock.Anything).
					Return(nil, errors.New("put recommend error")).
					Once()
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRiskenClient := mocks.NewRiskenClient(t)
			tc.setupMock(mockRiskenClient)

			service := &reviewService{
				riskenClient: mockRiskenClient,
			}
			gotResp, err := service.putFinding(ctx, tc.args.projectID, tc.args.scanResult)
			if (err != nil) != tc.wantErr {
				t.Errorf("putFinding() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if diff := cmp.Diff(gotResp, tc.want, cmpopts.IgnoreUnexported(
				finding.PutFindingResponse{},
				finding.Finding{},
			)); diff != "" {
				t.Errorf("putFinding() %s", diff)
			}
		})
	}
}
