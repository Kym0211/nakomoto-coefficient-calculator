package chains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sort"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type HyperliquidValidator struct {
	Validator string  `json:"validator"` 
	Name      string  `json:"name"`
	Stake     float64 `json:"stake"`  
	IsActive  bool    `json:"isActive"` 
}

type HyperliquidResponse []HyperliquidValidator

func Hyperliquid() (int, error) {
	url := "https://api.hyperliquid.xyz/info"
	
	payload := []byte(`{"type": "validatorSummaries"}`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var validators HyperliquidResponse
	err = json.Unmarshal(body, &validators)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hyperliquid response: %v", err)
	}

	var votingPowers []*big.Int
	totalVotingPower := big.NewInt(0)
	activeCount := 0

	for _, v := range validators {
		if !v.IsActive {
			continue
		}

		vp := big.NewInt(int64(v.Stake))
		
		votingPowers = append(votingPowers, vp)
		totalVotingPower.Add(totalVotingPower, vp)
		activeCount++
	}

	if len(votingPowers) == 0 {
		return 0, fmt.Errorf("no active validators found for Hyperliquid")
	}

	sort.Slice(votingPowers, func(i, j int) bool {
		return votingPowers[i].Cmp(votingPowers[j]) > 0
	})

	log.Printf("Hyperliquid: Fetched %d active validators. Total Stake: %s", activeCount, totalVotingPower.String())

	nakamotoCoefficient := utils.CalcNakamotoCoefficientBigInt(totalVotingPower, votingPowers)
	log.Printf("The Nakamoto coefficient for Hyperliquid is %d", nakamotoCoefficient)

	return nakamotoCoefficient, nil
}