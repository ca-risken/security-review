package review

import (
	"context"
)

func (r *reviewService) getProjectID(ctx context.Context) (*uint32, error) {
	signinResp, err := r.riskenClient.Signin(ctx)
	if err != nil {
		return nil, err
	}
	return &signinResp.ProjectID, nil
}
