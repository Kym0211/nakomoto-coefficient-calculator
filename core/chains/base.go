package chains

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

func Base() (int, error) {
	url := "https://mainnet.base.org"

	payload := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to create base request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("base rpc unreachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("base rpc returned status: %d", resp.StatusCode)
	}

	return 1, nil
}