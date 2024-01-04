package risken

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const (
	RISKEN_API_GITHUB_REVIEW_PATH = "/api/v1/code/list-github-setting/?project_id=1247" // TODO: fix
)

func (s *riskenService) InvokePRReview(ctx context.Context) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", s.opt.RiskenEndpoint+RISKEN_API_GITHUB_REVIEW_PATH, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.opt.RiskenApiToken))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "Success RISKEN API request.", slog.String("status", resp.Status), slog.String("response", string(body)))
	return nil
}
