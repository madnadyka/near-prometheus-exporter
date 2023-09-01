package collector

import (
	"encoding/json"
	"fmt"
	nearapi "github.com/masknetgoal634/near-exporter/client"
	"github.com/prometheus/client_golang/prometheus"
)

type NodeRpcMetrics struct {
	accountId                 string
	client                    *nearapi.Client
	epochBlockProducedDesc    *prometheus.Desc
	epochBlockExpectedDesc    *prometheus.Desc
	epochChunksProducedDesc   *prometheus.Desc
	epochChunksExpectedDesc   *prometheus.Desc
	seatPriceDesc             *prometheus.Desc
	delegatorStakeDesc        *prometheus.Desc
	epochStartHeightDesc      *prometheus.Desc
	blockNumberDesc           *prometheus.Desc
	syncingDesc               *prometheus.Desc
	versionBuildDesc          *prometheus.Desc
	currentValidatorStakeDesc *prometheus.Desc
	nextValidatorStakeDesc    *prometheus.Desc
	prevEpochKickoutDesc      *prometheus.Desc
	currentProposalsDesc      *prometheus.Desc
}

type DelegatorAccount struct {
	AccountId       string `json:"account_id"`
	UnstakedBalance string `json:"unstaked_balance"`
	StakedBalance   string `json:"staked_balance"`
	CanWithdraw     bool   `json:"can_withdraw"`
}

func NewNodeRpcMetrics(client *nearapi.Client, accountId string) *NodeRpcMetrics {
	return &NodeRpcMetrics{
		accountId: accountId,
		client:    client,
		epochBlockProducedDesc: prometheus.NewDesc(
			"near_account_epoch_block_produced_number",
			"The number of block produced in epoch of a given account id",
			nil,
			nil,
		),
		epochBlockExpectedDesc: prometheus.NewDesc(
			"near_account_epoch_block_expected_number",
			"The number of block expected in epoch of a given account id",
			nil,
			nil,
		),
		epochChunksProducedDesc: prometheus.NewDesc(
			"near_account_epoch_chunks_produced_number",
			"The number of chunks produced in epoch of a given account id",
			nil,
			nil,
		),
		epochChunksExpectedDesc: prometheus.NewDesc(
			"near_account_epoch_chunks_expected_number",
			"The number of chunks expected in epoch of a given account id",
			nil,
			nil,
		),
		delegatorStakeDesc: prometheus.NewDesc(
			"near_account_delegator_stake",
			"Delegators stake of a given account id",
			[]string{"delegator_account_id"},
			nil,
		),
		currentValidatorStakeDesc: prometheus.NewDesc(
			"near_account_current_validator_stake",
			"Current amount of validator stake of a given account id",
			nil,
			nil,
		),
		nextValidatorStakeDesc: prometheus.NewDesc(
			"near_account_next_validator_stake",
			"The next validator stake of a given account id",
			nil,
			nil,
		),
		currentProposalsDesc: prometheus.NewDesc(
			"near_account_current_proposals_stake",
			"Current proposals of a given account id",
			nil,
			nil,
		),
		prevEpochKickoutDesc: prometheus.NewDesc(
			"near_account_prev_epoch_kickout",
			"Near previous epoch kicked out of a given account id",
			[]string{"reason"},
			nil,
		),
		epochStartHeightDesc: prometheus.NewDesc(
			"near_epoch_start_height",
			"Near epoch start height",
			nil,
			nil,
		),
		blockNumberDesc: prometheus.NewDesc(
			"near_block_number",
			"The number of most recent block",
			nil,
			nil,
		),
		syncingDesc: prometheus.NewDesc(
			"near_sync_state",
			"Sync state",
			nil,
			nil,
		),
		versionBuildDesc: prometheus.NewDesc(
			"near_version_build",
			"The Near node version build",
			[]string{"version", "build"},
			nil,
		),
		seatPriceDesc: prometheus.NewDesc(
			"near_seat_price",
			"Validator seat price",
			nil,
			nil,
		),
	}
}

func (collector *NodeRpcMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.epochBlockProducedDesc
	ch <- collector.epochBlockExpectedDesc
	ch <- collector.epochChunksProducedDesc
	ch <- collector.epochChunksExpectedDesc
	ch <- collector.seatPriceDesc
	ch <- collector.delegatorStakeDesc
	ch <- collector.epochStartHeightDesc
	ch <- collector.blockNumberDesc
	ch <- collector.syncingDesc
	ch <- collector.versionBuildDesc
	ch <- collector.currentValidatorStakeDesc
	ch <- collector.nextValidatorStakeDesc
	ch <- collector.currentProposalsDesc
	ch <- collector.prevEpochKickoutDesc
}

func (collector *NodeRpcMetrics) Collect(ch chan<- prometheus.Metric) {
	sr, err := collector.client.Get("status", nil)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.versionBuildDesc, err)
		return
	}
	syn := sr.Status.SyncInfo.Syncing
	var isSyncing int
	if syn {
		isSyncing = 1
	} else {
		isSyncing = 0
	}
	ch <- prometheus.MustNewConstMetric(collector.syncingDesc, prometheus.GaugeValue, float64(isSyncing))

	blockHeight := sr.Status.SyncInfo.LatestBlockHeight
	ch <- prometheus.MustNewConstMetric(collector.blockNumberDesc, prometheus.GaugeValue, float64(blockHeight))

	versionBuildInt := HashString(sr.Status.Version.Build)
	ch <- prometheus.MustNewConstMetric(collector.versionBuildDesc, prometheus.GaugeValue, float64(versionBuildInt), sr.Status.Version.Version, sr.Status.Version.Build)

	r, err := collector.client.Get("validators", "latest")
	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.epochBlockProducedDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.epochBlockExpectedDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.epochChunksProducedDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.epochChunksExpectedDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.seatPriceDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.epochStartHeightDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.blockNumberDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.syncingDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.versionBuildDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.currentValidatorStakeDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.nextValidatorStakeDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.currentProposalsDesc, err)
		ch <- prometheus.NewInvalidMetric(collector.prevEpochKickoutDesc, err)
		return
	}

	ch <- prometheus.MustNewConstMetric(collector.epochStartHeightDesc, prometheus.GaugeValue, float64(r.Validators.EpochStartHeight))

	var seatPrice float64
	for _, v := range r.Validators.CurrentValidators {
		stake := GetStakeFromString(v.Stake)
		if seatPrice == 0 {
			seatPrice = stake
		}
		if seatPrice > stake {
			seatPrice = stake
		}
		if v.AccountId == collector.accountId {
			ch <- prometheus.MustNewConstMetric(collector.currentValidatorStakeDesc, prometheus.GaugeValue, stake)
			ch <- prometheus.MustNewConstMetric(collector.epochBlockProducedDesc, prometheus.GaugeValue, float64(v.NumProducedBlocks))
			ch <- prometheus.MustNewConstMetric(collector.epochBlockExpectedDesc, prometheus.GaugeValue, float64(v.NumExpectedBlocks))
			ch <- prometheus.MustNewConstMetric(collector.epochChunksProducedDesc, prometheus.GaugeValue, float64(v.NumProducedChunks))
			ch <- prometheus.MustNewConstMetric(collector.epochChunksExpectedDesc, prometheus.GaugeValue, float64(v.NumExpectedChunks))
		}
	}
	ch <- prometheus.MustNewConstMetric(collector.seatPriceDesc, prometheus.GaugeValue, seatPrice)
	for _, v := range r.Validators.NextValidators {
		if v.AccountId == collector.accountId {
			ch <- prometheus.MustNewConstMetric(collector.nextValidatorStakeDesc, prometheus.GaugeValue, float64(GetStakeFromString(v.Stake)))
		}
	}

	for _, v := range r.Validators.CurrentProposals {
		if v.AccountId == collector.accountId {
			ch <- prometheus.MustNewConstMetric(collector.currentProposalsDesc, prometheus.GaugeValue, float64(GetStakeFromString(v.Stake)))
		}
	}

	for _, v := range r.Validators.PrevEpochKickOut {
		if v.AccountId == collector.accountId {
			ch <- prometheus.MustNewConstMetric(collector.prevEpochKickoutDesc, prometheus.GaugeValue, 0, fmt.Sprintf("%v", v.Reason))
		}
	}

	d, err := collector.client.Get("query", map[string]interface{}{"request_type": "call_function",
		"finality":    "final",
		"account_id":  collector.accountId,
		"method_name": "get_accounts",
		"args_base64": "eyJmcm9tX2luZGV4IjogMCwgImxpbWl0IjogMTAwfQ=="})

	if err != nil {
		ch <- prometheus.NewInvalidMetric(collector.delegatorStakeDesc, err)
		return
	}

	resultString := ""
	for _, n := range d.Result.Result {
		resultString += string(n)

	}
	res := []DelegatorAccount{}
	_ = json.Unmarshal([]byte(resultString), &res)

	for _, delegator := range res {
		ch <- prometheus.MustNewConstMetric(collector.delegatorStakeDesc, prometheus.GaugeValue, float64(GetStakeFromString(delegator.StakedBalance)), delegator.AccountId)
	}

}
