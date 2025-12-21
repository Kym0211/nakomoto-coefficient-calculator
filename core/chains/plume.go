package chains

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

func Plume() (int, error) {
	url := "https://rpc.plume.org"

	payload := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to create plume request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("plume rpc unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("plume rpc returned status: %d", resp.StatusCode)
	}

	return 1, nil
}