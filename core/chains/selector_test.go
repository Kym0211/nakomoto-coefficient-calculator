package chains

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// This test tries multiple common function signatures to find the right one
func TestFindMonadSelector(t *testing.T) {
	rpcUrl := "https://rpc.monad.xyz"
	contractAddr := "0x0000000000000000000000000000000000001000"

	// Map of common validator getter functions and their Keccak-256 signatures
	candidates := map[string]string{
		"getAllValidators()":    "0x3233c09b",
		"getValidators()":       "0x9c424d9c",
		"getActiveValidators()": "0xc3c9a405",
		"getValidatorSet()":     "0x6436329d",
		"validators()":          "0x3a4b66f1",
		"currentValidators()":   "0x247e954a",
	}

	t.Logf("🔍 Hunting for correct selector on %s...", contractAddr)

	found := false

	for name, selector := range candidates {
		// Construct Request
		payload := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "eth_call",
			"params": []interface{}{
				map[string]string{
					"to":   contractAddr,
					"data": selector,
				},
				"latest",
			},
			"id": 1,
		}

		jsonPayload, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", rpcUrl, bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("❌ Network error checking %s: %v", name, err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		// Check response
		if bytes.Contains(body, []byte("execution reverted")) {
			t.Logf("❌ Failed: %s (%s) - Method not supported", name, selector)
		} else if bytes.Contains(body, []byte("result")) {
			// If we get a result (even empty 0x), the method exists!
			var r struct {
				Result string `json:"result"`
			}
			json.Unmarshal(body, &r)
			
			if len(r.Result) > 10 {
				t.Logf("✅✅✅ FOUND IT! Function: %s | Selector: %s", name, selector)
				t.Logf("Returned Data Length: %d chars", len(r.Result))
				found = true
			} else {
				t.Logf("⚠️  Method exists but returned empty data: %s (%s)", name, selector)
			}
		}
	}

	if !found {
		t.Fatal("Could not find a working selector. The contract address might be wrong or the method is private.")
	}
}