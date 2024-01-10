package review

import (
	"context"
	"fmt"

	"github.com/ca-risken/code/pkg/codescan"
	"github.com/ca-risken/code/pkg/gitleaks"
	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/datasource-api/pkg/message"
)

func (r *reviewService) getProjectID(ctx context.Context) (*uint32, error) {
	signinResp, err := r.riskenClient.Signin(ctx)
	if err != nil {
		return nil, err
	}
	return &signinResp.ProjectID, nil
}

func (r *reviewService) getPutFindingRequest(ctx context.Context, projectID uint32, s *ScanResult) (*finding.PutFindingRequest, error) {
	var putReq *finding.PutFindingRequest
	switch scanResult := s.ScanResult.(type) {
	case *codescan.SemgrepFinding:
		req, err := codescan.GeneratePutFindingRequest(projectID, scanResult)
		if err != nil {
			return nil, err
		}
		putReq = req
	case *gitleaks.GitleaksFinding:
		req, err := gitleaks.GeneratePutFindingRequest(projectID, scanResult)
		if err != nil {
			return nil, err
		}
		putReq = req
	default:
		return nil, fmt.Errorf("unknown scan result type: %T", scanResult)
	}
	return putReq, nil
}

func (r *reviewService) getPutRecommendRequest(ctx context.Context, projectID uint32, findingID uint64, s *ScanResult) (*finding.PutRecommendRequest, error) {
	var recReq *finding.PutRecommendRequest
	switch scanResult := s.ScanResult.(type) {
	case *codescan.SemgrepFinding:
		recommendContent := codescan.GetSemgrepRecommend(
			scanResult.Repository,
			scanResult.Path,
			scanResult.CheckID,
			scanResult.Extra.Message,
			scanResult.GitHubURL,
			scanResult.Extra.Lines,
		)
		recReq = &finding.PutRecommendRequest{
			ProjectId:      projectID,
			FindingId:      findingID,
			DataSource:     message.CodeScanDataSource,
			Type:           codescan.GenerateDataSourceIDForSemgrep(scanResult),
			Risk:           recommendContent.Risk,
			Recommendation: recommendContent.Recommendation,
		}

	case *gitleaks.GitleaksFinding:
		recommendContent := gitleaks.GetRecommend(
			scanResult.Result.RuleDescription,
			scanResult.Result.Repo,
			scanResult.Result.File,
			*scanResult.RepositoryMetadata.Visibility,
			scanResult.Result.URL,
			scanResult.Result.Author,
			scanResult.Result.Email,
		)
		recReq = &finding.PutRecommendRequest{
			ProjectId:      projectID,
			FindingId:      findingID,
			DataSource:     message.GitleaksDataSource,
			Type:           scanResult.Result.RuleDescription,
			Risk:           recommendContent.Risk,
			Recommendation: recommendContent.Recommendation,
		}

	default:
		return nil, fmt.Errorf("unknown scan result type: %T", scanResult)
	}
	return recReq, nil
}

func (r *reviewService) putFinding(ctx context.Context, projectID uint32, s *ScanResult) (*finding.PutFindingResponse, error) {
	// Finding
	putReq, err := r.getPutFindingRequest(ctx, projectID, s)
	if err != nil {
		return nil, err
	}
	putResp, err := r.riskenClient.PutFinding(ctx, putReq)
	if err != nil {
		return nil, err
	}

	// Recommendation
	findingID := putResp.Finding.FindingId
	recReq, err := r.getPutRecommendRequest(ctx, projectID, findingID, s)
	if err != nil {
		return nil, err
	}
	if _, err = r.riskenClient.PutRecommend(ctx, recReq); err != nil {
		return nil, err
	}

	return putResp, nil
}
