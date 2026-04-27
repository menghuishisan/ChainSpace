package service

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// AgentBattleScoreService 负责对抗赛轮次结算与最终排行统计。
type AgentBattleScoreService struct {
	contractRepo *repository.AgentContractRepository
	eventRepo    *repository.AgentBattleEventRepository
	scoreRepo    *repository.AgentBattleScoreRepository
	teamRepo     *repository.TeamRepository
}

func NewAgentBattleScoreService(
	contractRepo *repository.AgentContractRepository,
	eventRepo *repository.AgentBattleEventRepository,
	scoreRepo *repository.AgentBattleScoreRepository,
	teamRepo *repository.TeamRepository,
) *AgentBattleScoreService {
	return &AgentBattleScoreService{
		contractRepo: contractRepo,
		eventRepo:    eventRepo,
		scoreRepo:    scoreRepo,
		teamRepo:     teamRepo,
	}
}

func (s *AgentBattleScoreService) CalculateFinalScores(ctx context.Context, round *model.AgentBattleRound) error {
	client, err := ethclient.Dial(round.ChainRPCURL)
	if err != nil {
		return fmt.Errorf("connect to chain: %w", err)
	}
	defer client.Close()

	contracts, err := s.contractRepo.ListByRound(ctx, round.ID)
	if err != nil {
		return err
	}

	events, _, _ := s.eventRepo.List(ctx, round.ID, "", 1, 10000)
	attackCounts := make(map[uint]int)
	defenseCounts := make(map[uint]int)
	penaltyCounts := make(map[uint]int)
	attackGain := make(map[uint]int64)
	defenseGain := make(map[uint]int64)
	resourceGain := make(map[uint]int64)
	for _, event := range events {
		if event.TeamID == nil {
			continue
		}
		teamID := *event.TeamID
		scoreDelta := readEventInt64(event.EventData, "score_delta", "points", "point_delta")
		resourceDelta := readEventInt64(event.EventData, "resource_delta", "resource_change", "resource")

		switch event.EventType {
		case "attack":
			attackCounts[teamID]++
			attackGain[teamID] += maxInt64(scoreDelta, resourceDelta)
		case "defense":
			defenseCounts[teamID]++
			defenseGain[teamID] += maxInt64(scoreDelta, resourceDelta)
		case model.EventTypePenalty, model.EventTypeError:
			penaltyCounts[teamID]++
		}

		if resourceDelta > 0 {
			resourceGain[teamID] += resourceDelta
		}
	}

	scoreWeights := resolveBattleScoreWeights(round)

	type scoreData struct {
		teamID       uint
		finalBalance *big.Int
		attackCount  int
		defenseCount int
		penaltyCount int
		score        int
		resourceHeld int
	}

	scores := make([]scoreData, 0, len(contracts))
	for _, contract := range contracts {
		if contract.TeamID == 0 || contract.ContractAddress == "" {
			continue
		}

		address := common.HexToAddress(contract.ContractAddress)
		balance, balanceErr := client.BalanceAt(ctx, address, nil)
		if balanceErr != nil {
			logger.Error("Failed to get contract balance", zap.String("address", contract.ContractAddress), zap.Error(balanceErr))
			balance = big.NewInt(0)
		}

		scores = append(scores, scoreData{
			teamID:       contract.TeamID,
			finalBalance: balance,
			attackCount:  attackCounts[contract.TeamID],
			defenseCount: defenseCounts[contract.TeamID],
			penaltyCount: penaltyCounts[contract.TeamID],
			score: calculateBattleCompositeScore(
				scoreWeights,
				normalizeTokenBalance(balance),
				attackCounts[contract.TeamID],
				defenseCounts[contract.TeamID],
				penaltyCounts[contract.TeamID],
				attackGain[contract.TeamID],
				defenseGain[contract.TeamID],
				resourceGain[contract.TeamID],
				contract.Status,
			),
			resourceHeld: calculateBattleResourceHeld(balance, resourceGain[contract.TeamID]),
		})
	}

	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score || (scores[j].score == scores[i].score && scores[j].finalBalance.Cmp(scores[i].finalBalance) > 0) {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	for rank, sd := range scores {
		score := &model.AgentBattleScore{
			ContestID:    round.ContestID,
			RoundID:      round.ID,
			TeamID:       sd.teamID,
			Score:        sd.score,
			TokenBalance: sd.finalBalance.String(),
			SuccessCount: sd.attackCount,
			FailCount:    sd.penaltyCount,
			Rank:         rank + 1,
			Details: model.JSONMap{
				"attack_count":    sd.attackCount,
				"defense_count":   sd.defenseCount,
				"penalty_count":   sd.penaltyCount,
				"resource_held":   sd.resourceHeld,
				"resource_weight": scoreWeights["resource"],
				"attack_weight":   scoreWeights["attack"],
				"defense_weight":  scoreWeights["defense"],
				"survival_weight": scoreWeights["survival"],
			},
		}
		if err := s.scoreRepo.Upsert(ctx, score); err != nil {
			logger.Error("Failed to save score", zap.Uint("teamID", sd.teamID), zap.Error(err))
		}
	}

	return nil
}

func resolveBattleScoreWeights(round *model.AgentBattleRound) map[string]int {
	if round != nil && round.Contest != nil && round.Contest.BattleOrchestration.Judge.ScoreWeights != nil {
		return round.Contest.BattleOrchestration.Judge.ScoreWeights
	}

	return map[string]int{
		"resource": 35,
		"attack":   30,
		"defense":  20,
		"survival": 15,
	}
}

func normalizeTokenBalance(balance *big.Int) int64 {
	if balance == nil || balance.Sign() <= 0 {
		return 0
	}

	weiPerPoint := big.NewInt(1_000_000_000_000_000)
	scaled := new(big.Int).Div(new(big.Int).Set(balance), weiPerPoint)
	if scaled.Sign() == 0 {
		return 1
	}
	if !scaled.IsInt64() {
		return int64(^uint64(0) >> 1)
	}
	return scaled.Int64()
}

func calculateBattleResourceHeld(balance *big.Int, resourceGain int64) int {
	held := normalizeTokenBalance(balance)
	if resourceGain > held {
		held = resourceGain
	}
	if held < 0 {
		return 0
	}
	if held > int64(^uint(0)>>1) {
		return int(^uint(0) >> 1)
	}
	return int(held)
}

func calculateBattleCompositeScore(
	weights map[string]int,
	resourceMetric int64,
	attackCount int,
	defenseCount int,
	penaltyCount int,
	attackGain int64,
	defenseGain int64,
	resourceGain int64,
	contractStatus string,
) int {
	resourceWeight := int64(weights["resource"])
	attackWeight := int64(weights["attack"])
	defenseWeight := int64(weights["defense"])
	survivalWeight := int64(weights["survival"])

	survivalMetric := int64(0)
	if strings.EqualFold(contractStatus, model.ContractStatusDeployed) {
		survivalMetric = 1
	}

	resourceScore := resourceMetric*resourceWeight + resourceGain*(resourceWeight/2)
	attackScore := (int64(attackCount) + attackGain) * maxInt64(attackWeight, 1)
	defenseScore := (int64(defenseCount) + defenseGain) * maxInt64(defenseWeight, 1)
	survivalScore := survivalMetric * maxInt64(survivalWeight*10, 10)
	penaltyScore := int64(penaltyCount) * maxInt64(resourceWeight+attackWeight+defenseWeight, 10)

	total := resourceScore + attackScore + defenseScore + survivalScore - penaltyScore
	if total < 0 {
		return 0
	}
	if total > int64(^uint(0)>>1) {
		return int(^uint(0) >> 1)
	}
	return int(total)
}

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
