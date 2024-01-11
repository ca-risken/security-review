package review

import (
	"context"

	"github.com/ca-risken/core/proto/finding"
	"github.com/ca-risken/go-risken"
)

type RiskenClient interface {
	Signin(ctx context.Context) (*risken.SigninResponse, error)
	PutFinding(ctx context.Context, req *finding.PutFindingRequest) (*finding.PutFindingResponse, error)
	PutRecommend(ctx context.Context, req *finding.PutRecommendRequest) (*finding.PutRecommendResponse, error)
}

type riskenClient struct {
	*risken.Client
}

func NewRiskenClient(token, endpoint string) RiskenClient {
	return &riskenClient{
		Client: risken.NewClient(token, risken.WithAPIEndpoint(endpoint)),
	}
}
