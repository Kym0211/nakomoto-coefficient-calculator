package chains

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type StoryRpcResponse struct {
	Result struct {
		Validators []struct {
			Address     string `json:"address"`
			VotingPower string `json:"voting_power"`
		} `json:"validators"`
		Total string `json:"total"` 
		Count string `json:"count"`
	} `json:"result"`
}

func Story() (int, error) {
	url := "https://story-mainnet-rpc.itrocket.net/"
	nc, err := fetchStoryRpc(url)
	if err != nil {
		return 0, fmt.Errorf("all Story RPC endpoints failed: %v", err)
	}
	return nc, nil
}

func fetchStoryRpc(baseURL string) (int, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	var allValidators []struct {
		Address     string `json:"address"`
		VotingPower string `json:"voting_power"`
	}
	
	page := 1
	totalValidators := 0

	for {
		url := fmt.Sprintf("%s/validators?page=%d&per_page=100", baseURL, page)
		
		req, _ := http.NewRequest("GET", url, nil)
		resp, err := client.Do(req)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return 0, fmt.Errorf("status %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var rpcResp StoryRpcResponse
		if err := json.Unmarshal(body, &rpcResp); err != nil {
			return 0, fmt.Errorf("parse error: %v", err)
		}

		allValidators = append(allValidators, rpcResp.Result.Validators...)

		serverTotal, _ := strconv.Atoi(rpcResp.Result.Total)
		if totalValidators == 0 {
			totalValidators = serverTotal
		}

		if len(allValidators) >= totalValidators || len(rpcResp.Result.Validators) == 0 {
			break
		}
		
		page++
		time.Sleep(100 * time.Millisecond) 
	}

	if len(allValidators) == 0 {
		return 0, fmt.Errorf("no validators found")
	}

	var votingPowers []*big.Int
	totalVotingPower := big.NewInt(0)

	for _, v := range allValidators {
		vp := new(big.Int)
		vp.SetString(v.VotingPower, 10)
		
		if vp.Cmp(big.NewInt(0)) > 0 {
			votingPowers = append(votingPowers, vp)
			totalVotingPower.Add(totalVotingPower, vp)
		}
	}

	sort.Slice(votingPowers, func(i, j int) bool {
		return votingPowers[i].Cmp(votingPowers[j]) > 0
	})

	nc := utils.CalcNakamotoCoefficientBigInt(totalVotingPower, votingPowers)
	
	return nc, nil
}