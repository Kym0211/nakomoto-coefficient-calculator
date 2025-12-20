package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type NamadaValidator struct {
	VotingPower string `json:"voting_power"`
}

type NamadaValidatorsResponse struct {
	Result struct {
		Validators []NamadaValidator `json:"validators"`
		Total      string            `json:"total"` 
		Count      string            `json:"count"`
	} `json:"result"`
}

func Namada() (int, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()

	var allValidators []NamadaValidator
	page := 1
	totalValidators := 0
	
	for {
		validatorsURL := fmt.Sprintf("https://rpc.namada.validatus.com/validators?page=%d&per_page=100", page)
		
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, validatorsURL, nil)
		if err != nil {
			return 0, fmt.Errorf("create request error: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("rpc fetch error: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		var valResp NamadaValidatorsResponse
		err = json.Unmarshal(body, &valResp)
		if err != nil {
			return 0, fmt.Errorf("json parse error: %v", err)
		}

		allValidators = append(allValidators, valResp.Result.Validators...)

		totalFromServer, _ := strconv.Atoi(valResp.Result.Total)
		if totalValidators == 0 {
			totalValidators = totalFromServer
		}

		if len(allValidators) >= totalValidators || len(valResp.Result.Validators) == 0 {
			break
		}

		page++
		time.Sleep(200 * time.Millisecond)
	}

	var votingPowers []*big.Int
	totalVotingPower := big.NewInt(0)

	for _, v := range allValidators {
		vp := new(big.Int)
		_, ok := vp.SetString(v.VotingPower, 10)
		if !ok {
			log.Println("Error parsing validator voting power:", v.VotingPower)
			continue
		}
		votingPowers = append(votingPowers, vp)
		totalVotingPower.Add(totalVotingPower, vp)
	}

	if len(votingPowers) == 0 {
		return 0, fmt.Errorf("no validators found")
	}

	sort.Slice(votingPowers, func(i, j int) bool {
		return votingPowers[i].Cmp(votingPowers[j]) > 0
	})
	fmt.Println("Total voting power :", totalVotingPower)

	nakamotoCoefficient := utils.CalcNakamotoCoefficientBigInt(totalVotingPower, votingPowers)

	return nakamotoCoefficient, nil
}