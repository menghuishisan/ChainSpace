package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"regexp"

	"github.com/chainspace/backend/internal/dto/request"
	"github.com/chainspace/backend/internal/dto/response"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/cache"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/pkg/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ContestScoreService 负责解题赛提交判分、动态分与排行榜。
type ContestScoreService struct {
	contestRepo           *repository.ContestRepository
	challengeRepo         *repository.ChallengeRepository
	contestChallengeRepo  *repository.ContestChallengeRepository
	teamRepo              *repository.TeamRepository
	contestSubmissionRepo *repository.ContestSubmissionRepository
	contestScoreRepo      *repository.ContestScoreRepository
	redis                 *redis.Client
}

func NewContestScoreService(
	contestRepo *repository.ContestRepository,
	challengeRepo *repository.ChallengeRepository,
	contestChallengeRepo *repository.ContestChallengeRepository,
	teamRepo *repository.TeamRepository,
	contestSubmissionRepo *repository.ContestSubmissionRepository,
	contestScoreRepo *repository.ContestScoreRepository,
	redisClient *redis.Client,
) *ContestScoreService {
	return &ContestScoreService{
		contestRepo:           contestRepo,
		challengeRepo:         challengeRepo,
		contestChallengeRepo:  contestChallengeRepo,
		teamRepo:              teamRepo,
		contestSubmissionRepo: contestSubmissionRepo,
		contestScoreRepo:      contestScoreRepo,
		redis:                 redisClient,
	}
}

func (s *ContestScoreService) SubmitFlag(ctx context.Context, userID uint, schoolID uint, role string, contestID uint, req *request.SubmitFlagRequest, ip string) (*response.FlagSubmitResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, userID, schoolID, role); err != nil {
		return nil, err
	}
	if err := ensureContestOngoing(contest); err != nil {
		return nil, err
	}

	team, err := s.teamRepo.GetUserTeam(ctx, contestID, userID)
	if err != nil || team == nil {
		return nil, errors.ErrPermissionDenied.WithMessage("请先报名参赛")
	}
	teamID := &team.ID

	solved, _ := s.contestSubmissionRepo.HasSolved(ctx, contestID, req.ChallengeID, teamID, userID)
	if solved {
		return &response.FlagSubmitResponse{Correct: false, Message: "已解决该题目"}, nil
	}

	contestChallenge, err := s.contestChallengeRepo.GetByID(ctx, contestID, req.ChallengeID)
	if err != nil {
		return nil, errors.ErrChallengeNotFound
	}

	isCorrect := s.validateFlag(ctx, contestChallenge.Challenge, req.Flag, userID)
	submission := &model.ContestSubmission{
		ContestID:   contestID,
		ChallengeID: req.ChallengeID,
		UserID:      userID,
		TeamID:      teamID,
		Flag:        req.Flag,
		IsCorrect:   isCorrect,
		IP:          ip,
	}

	if isCorrect {
		submission.Points = contestChallenge.CurrentPoints
		firstBlood, _ := s.contestSubmissionRepo.GetFirstBlood(ctx, contestID, req.ChallengeID)
		if firstBlood == nil && contest.FirstBloodBonus > 0 {
			submission.Points += contest.FirstBloodBonus
		}
		if contest.DynamicScore {
			s.updateDynamicScore(ctx, contestID, req.ChallengeID, contestChallenge)
		}
		_ = s.contestScoreRepo.AddScore(ctx, contestID, teamID, userID, submission.Points)
		_ = s.contestScoreRepo.UpdateRanks(ctx, contestID)
	}

	_ = s.contestSubmissionRepo.Create(ctx, submission)
	_ = s.challengeRepo.IncrementAttemptCount(ctx, req.ChallengeID)
	if isCorrect {
		_ = s.challengeRepo.IncrementSolveCount(ctx, req.ChallengeID)
		s.invalidateScoreboardCache(ctx, contestID)
	}

	resp := &response.FlagSubmitResponse{Correct: isCorrect, Points: submission.Points}
	if isCorrect {
		resp.Message = "答案正确"
	} else {
		resp.Message = "答案错误"
	}
	return resp, nil
}

func (s *ContestScoreService) GetScoreboard(ctx context.Context, contestID, currentUserID, schoolID uint, role string) (*response.ScoreboardResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		return nil, errors.ErrContestNotFound
	}
	if err := ensureContestAccessible(contest, currentUserID, schoolID, role); err != nil {
		return nil, err
	}

	list, ok := s.getScoreboardListFromCache(ctx, contestID)
	if !ok {
		list, err = s.loadScoreboardList(ctx, contestID)
		if err != nil {
			return nil, err
		}
		s.cacheScoreboardList(ctx, contestID, list)
	}

	var myRank int
	var myScore int
	var myTeamID *uint
	if currentUserID > 0 {
		if team, teamErr := s.teamRepo.GetUserTeam(ctx, contestID, currentUserID); teamErr == nil && team != nil {
			myTeamID = &team.ID
		}
	}
	for _, entry := range list {
		if myTeamID != nil && entry.TeamID != nil && *entry.TeamID == *myTeamID {
			myRank = entry.Rank
			myScore = entry.TotalScore
			break
		}
		if myTeamID == nil && entry.UserID == currentUserID {
			myRank = entry.Rank
			myScore = entry.TotalScore
			break
		}
	}

	return &response.ScoreboardResponse{List: list, MyRank: myRank, MyScore: myScore}, nil
}

func (s *ContestScoreService) loadScoreboardList(ctx context.Context, contestID uint) ([]response.ScoreboardEntry, error) {
	scores, _, err := s.contestScoreRepo.GetScoreboard(ctx, contestID, 1, 100)
	if err != nil {
		return nil, errors.ErrDatabaseError.WithError(err)
	}

	list := make([]response.ScoreboardEntry, len(scores))
	for i, score := range scores {
		list[i] = response.ScoreboardEntry{
			Rank:            score.Rank,
			TotalScore:      score.TotalScore,
			SolveCount:      score.SolveCount,
			LastSolveAt:     score.LastSolveAt,
			FirstBloodCount: score.FirstBloodCount,
		}
		if list[i].Rank == 0 {
			list[i].Rank = i + 1
		}
		if score.Team != nil {
			list[i].TeamID = &score.Team.ID
			list[i].TeamName = score.Team.Name
		}
		if score.User != nil {
			list[i].UserID = score.UserID
			list[i].DisplayName = score.User.DisplayName()
		}
	}
	return list, nil
}

func (s *ContestScoreService) getScoreboardListFromCache(ctx context.Context, contestID uint) ([]response.ScoreboardEntry, bool) {
	if s.redis == nil {
		return nil, false
	}

	data, err := s.redis.Get(ctx, cache.ScoreboardKey(contestID)).Bytes()
	if err != nil {
		if err != redis.Nil {
			logger.Warn("Failed to get contest scoreboard cache", zap.Uint("contest_id", contestID), zap.Error(err))
		}
		return nil, false
	}

	var list []response.ScoreboardEntry
	if err := json.Unmarshal(data, &list); err != nil {
		logger.Warn("Failed to unmarshal contest scoreboard cache", zap.Uint("contest_id", contestID), zap.Error(err))
		return nil, false
	}
	return list, true
}

func (s *ContestScoreService) cacheScoreboardList(ctx context.Context, contestID uint, list []response.ScoreboardEntry) {
	if s.redis == nil {
		return
	}

	data, err := json.Marshal(list)
	if err != nil {
		logger.Warn("Failed to marshal contest scoreboard cache", zap.Uint("contest_id", contestID), zap.Error(err))
		return
	}

	if err := s.redis.Set(ctx, cache.ScoreboardKey(contestID), data, cache.TTLShort).Err(); err != nil {
		logger.Warn("Failed to set contest scoreboard cache", zap.Uint("contest_id", contestID), zap.Error(err))
	}
}

func (s *ContestScoreService) invalidateScoreboardCache(ctx context.Context, contestID uint) {
	if s.redis == nil {
		return
	}

	if err := s.redis.Del(ctx, cache.ScoreboardKey(contestID)).Err(); err != nil {
		logger.Warn("Failed to invalidate contest scoreboard cache", zap.Uint("contest_id", contestID), zap.Error(err))
	}
}

func (s *ContestScoreService) validateFlag(ctx context.Context, challenge *model.Challenge, flag string, userID uint) bool {
	switch challenge.FlagType {
	case "static":
		return challenge.FlagTemplate == flag
	case "dynamic":
		return generateDynamicFlag(challenge, userID) == flag
	case "regex":
		if challenge.FlagRegex != "" {
			matched, _ := regexp.MatchString(challenge.FlagRegex, flag)
			return matched
		}
		return challenge.FlagTemplate == flag
	case "service":
		return s.validateFlagWithService(ctx, challenge, flag, userID)
	default:
		return challenge.FlagTemplate == flag
	}
}

func generateDynamicFlag(challenge *model.Challenge, userID uint) string {
	data := fmt.Sprintf("%s:%d:%d", challenge.FlagSecret, userID, challenge.ID)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("flag{%s}", hex.EncodeToString(hash[:16]))
}

func (s *ContestScoreService) validateFlagWithService(ctx context.Context, challenge *model.Challenge, flag string, userID uint) bool {
	if challenge.ValidationConfig == nil {
		return false
	}

	validationType, _ := challenge.ValidationConfig["type"].(string)
	switch validationType {
	case "transaction_hash":
		return s.validateTransactionFlag(ctx, challenge, flag)
	case "contract_state":
		return s.validateContractStateFlag(ctx, challenge, flag, userID)
	case "balance_change":
		return s.validateBalanceFlag(ctx, challenge, flag, userID)
	default:
		return false
	}
}

func (s *ContestScoreService) validateTransactionFlag(ctx context.Context, challenge *model.Challenge, txHash string) bool {
	rpcURL, _ := challenge.ValidationConfig["rpc_url"].(string)
	if rpcURL == "" {
		return false
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return false
	}
	defer client.Close()

	hash := common.HexToHash(txHash)
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil || receipt.Status != 1 {
		return false
	}

	targetAddr, _ := challenge.ValidationConfig["target_address"].(string)
	if targetAddr != "" && receipt.ContractAddress.Hex() != targetAddr {
		tx, _, err := client.TransactionByHash(ctx, hash)
		if err != nil || tx.To() == nil || tx.To().Hex() != targetAddr {
			return false
		}
	}
	return true
}

func (s *ContestScoreService) validateContractStateFlag(ctx context.Context, challenge *model.Challenge, flag string, _ uint) bool {
	rpcURL, _ := challenge.ValidationConfig["rpc_url"].(string)
	contractAddr, _ := challenge.ValidationConfig["contract_address"].(string)
	if rpcURL == "" || contractAddr == "" {
		return false
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return false
	}
	defer client.Close()

	slotIndex, _ := challenge.ValidationConfig["storage_slot"].(float64)
	addr := common.HexToAddress(contractAddr)
	slot := common.BigToHash(big.NewInt(int64(slotIndex)))

	value, err := client.StorageAt(ctx, addr, slot, nil)
	if err != nil {
		return false
	}

	expectedValue, _ := challenge.ValidationConfig["expected_value"].(string)
	return common.Bytes2Hex(value) == expectedValue || flag == common.Bytes2Hex(value)
}

func (s *ContestScoreService) validateBalanceFlag(ctx context.Context, challenge *model.Challenge, _ string, _ uint) bool {
	rpcURL, _ := challenge.ValidationConfig["rpc_url"].(string)
	targetAddr, _ := challenge.ValidationConfig["target_address"].(string)
	if rpcURL == "" || targetAddr == "" {
		return false
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return false
	}
	defer client.Close()

	addr := common.HexToAddress(targetAddr)
	balance, err := client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return false
	}

	expectedBalance, _ := challenge.ValidationConfig["expected_balance"].(float64)
	return balance.Cmp(big.NewInt(int64(expectedBalance))) == 0
}

func (s *ContestScoreService) updateDynamicScore(ctx context.Context, contestID, challengeID uint, contestChallenge *model.ContestChallenge) {
	solveCount, _ := s.contestSubmissionRepo.CountSolves(ctx, contestID, challengeID)

	basePoints := float64(contestChallenge.Points)
	minPoints := basePoints
	if contestChallenge.Challenge.MinPoints > 0 {
		minPoints = math.Min(basePoints, float64(contestChallenge.Challenge.MinPoints))
	}
	decay := contestChallenge.Challenge.DecayFactor
	newPoints := int(math.Max(minPoints, basePoints-decay*float64(solveCount)*basePoints))
	_ = s.contestChallengeRepo.UpdateCurrentPoints(ctx, contestID, challengeID, newPoints)
}
