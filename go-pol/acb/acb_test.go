package acb

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"os"
	"strconv"
	"testing"
)

type valACBSummary struct {
	cntValidators          uint64
	cntValsNoIncentives    uint64
	cntVals50PNoIncentives uint64

	totalRewardRate    float64
	noIncentiveRate    float64
	lastDayRewardTotal float64

	totalWeightIncentives   float64
	totalWeightNoIncentives float64

	totalScaledWeightIncentives   float64
	totalScaledWeightNoIncentives float64
}

func processValidator(valResult gjson.Result, summary *valACBSummary) {

	if !valResult.Get("lastBlockUptime.isActive").Bool() {
		return
	}

	rewardAlloWeights := valResult.Get("rewardAllocationWeights")

	rewardRateStr := valResult.Get("dynamicData.rewardRate")
	rewardRate, _ := strconv.ParseFloat(rewardRateStr.Str, 64)

	lastDayRewardStr := valResult.Get("dynamicData.lastDayDistributedBGTAmount")
	lastDayReward, _ := strconv.ParseFloat(lastDayRewardStr.Str, 64)

	summary.lastDayRewardTotal += lastDayReward
	summary.totalRewardRate += rewardRate
	summary.cntValidators += 1

	totalValIncentives := 0.0

	alloNoIncentive := 0.0

	rewardAlloWeights.ForEach(func(key, alloResult gjson.Result) bool {
		alloPercent := alloResult.Get("percentageNumerator").Num
		alloValueUsdStr := alloResult.Get("receivingVault.dynamicData.activeIncentivesValueUsd").String()

		alloValueUsd, _ := strconv.ParseFloat(alloValueUsdStr, 64)

		totalValIncentives += alloValueUsd

		if alloValueUsd > 0.0 {
			summary.totalWeightIncentives += alloPercent
			summary.totalScaledWeightIncentives += alloPercent * lastDayReward
		} else {
			alloNoIncentive += alloPercent
			summary.totalWeightNoIncentives += alloPercent
			summary.totalScaledWeightNoIncentives += alloPercent * lastDayReward
		}
		return true
	})

	if alloNoIncentive >= 50.0 {
		summary.cntVals50PNoIncentives += 1
	}

	if totalValIncentives == 0.0 {
		summary.cntValsNoIncentives += 1
		summary.noIncentiveRate += rewardRate
	}
}

func TestBasicACBAnalysis(t *testing.T) {

	fileName := "./validators-query-result2.json"
	jsonBytes, err := os.ReadFile(fileName)
	require.NoError(t, err)

	var jsonObj map[string]any
	err = json.Unmarshal(jsonBytes, &jsonObj)
	require.NoError(t, err)

	value := gjson.Get(string(jsonBytes), "data.validators.validators")

	summary := &valACBSummary{}

	cntVals := 0
	valProcFunc := func(key, valResult gjson.Result) bool {
		processValidator(valResult, summary)
		cntVals++
		return true
	}
	value.ForEach(valProcFunc)

	println(fmt.Sprintf("VALS with 50P no INCENTIVES: %v", summary.cntVals50PNoIncentives))
	println(fmt.Sprintf("VALS with no INCENTIVES: %v", summary.cntValsNoIncentives))
	println(fmt.Sprintf("no INCENTIVES BGT rate: %v", summary.noIncentiveRate))
	println(fmt.Sprintf("TOTAL INCENTIVES 24hr BGT: %v", summary.lastDayRewardTotal))

	println()

	println(fmt.Sprintf("WITH incentives: %v", summary.totalWeightIncentives/float64(summary.cntValidators)/100.0))
	println(fmt.Sprintf("NO incentives: %v", summary.totalWeightNoIncentives/float64(summary.cntValidators)/100.0))

	println()

	println(fmt.Sprintf("SCALED WITH incentives: %v", summary.totalScaledWeightIncentives/100.0/summary.lastDayRewardTotal))
	println(fmt.Sprintf("SCALED NO incentives: %v", summary.totalScaledWeightNoIncentives/100.0/summary.lastDayRewardTotal))
}
