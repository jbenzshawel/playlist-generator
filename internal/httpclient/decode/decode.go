package decode

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func JSON[T any](resp *http.Response) (T, error) {
	var zero T

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var result T
		err := json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			return zero, err
		}
		return result, nil
	}

	if resp.StatusCode >= 400 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return zero, fmt.Errorf("error reading response body: %w", err)
		}

		slog.Warn("http request failed with error", "error", string(bodyBytes), "status", resp.StatusCode)
	}

	return zero, fmt.Errorf("http request failed with %d status", resp.StatusCode)
}

func Bytes(resp *http.Response) ([]byte, error) {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return bodyBytes, nil
	}

	if resp.StatusCode >= 400 {
		slog.Warn("http request failed with error", "error", string(bodyBytes), "status", resp.StatusCode)
	}

	return nil, fmt.Errorf("http request failed with %d status", resp.StatusCode)
}
