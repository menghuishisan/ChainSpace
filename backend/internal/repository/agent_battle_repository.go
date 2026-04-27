package repository

import (
	"context"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// AgentBattleRoundRepository 智能体对战轮次仓库
type AgentBattleRoundRepository struct {
	*BaseRepository
}

// NewAgentBattleRoundRepository 创建智能体对战轮次仓库
func NewAgentBattleRoundRepository(db *gorm.DB) *AgentBattleRoundRepository {
	return &AgentBattleRoundRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建轮次
func (r *AgentBattleRoundRepository) Create(ctx context.Context, round *model.AgentBattleRound) error {
	return r.DB(ctx).Create(round).Error
}

// Update 更新轮次
func (r *AgentBattleRoundRepository) Update(ctx context.Context, round *model.AgentBattleRound) error {
	return r.DB(ctx).Save(round).Error
}

// GetByID 根据ID获取轮次
func (r *AgentBattleRoundRepository) GetByID(ctx context.Context, id uint) (*model.AgentBattleRound, error) {
	var round model.AgentBattleRound
	err := r.DB(ctx).Preload("Contest").First(&round, id).Error
	if err != nil {
		return nil, err
	}
	return &round, nil
}

// GetCurrentRound 获取当前轮次
func (r *AgentBattleRoundRepository) GetCurrentRound(ctx context.Context, contestID uint) (*model.AgentBattleRound, error) {
	var round model.AgentBattleRound
	err := r.DB(ctx).Where("contest_id = ? AND status IN ?", contestID,
		[]string{model.RoundStatusPending, model.RoundStatusRunning}).
		Order("round_number DESC").
		First(&round).Error
	if err != nil {
		return nil, err
	}
	return &round, nil
}

// GetLatestRound 获取最新轮次，供比赛已结束后的观战/回放场景兜底使用。
func (r *AgentBattleRoundRepository) GetLatestRound(ctx context.Context, contestID uint) (*model.AgentBattleRound, error) {
	var round model.AgentBattleRound
	err := r.DB(ctx).
		Where("contest_id = ?", contestID).
		Order("round_number DESC").
		First(&round).Error
	if err != nil {
		return nil, err
	}
	return &round, nil
}

// List 获取轮次列表
func (r *AgentBattleRoundRepository) List(ctx context.Context, contestID uint, page, pageSize int) ([]model.AgentBattleRound, int64, error) {
	var rounds []model.AgentBattleRound
	var total int64

	query := r.DB(ctx).Model(&model.AgentBattleRound{}).Where("contest_id = ?", contestID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("round_number DESC").
		Find(&rounds).Error
	if err != nil {
		return nil, 0, err
	}

	return rounds, total, nil
}

// UpdateStatus 更新轮次状态
func (r *AgentBattleRoundRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	return r.DB(ctx).Model(&model.AgentBattleRound{}).Where("id = ?", id).Update("status", status).Error
}

// AgentContractRepository 智能体合约仓库
type AgentContractRepository struct {
	*BaseRepository
}

// NewAgentContractRepository 创建智能体合约仓库
func NewAgentContractRepository(db *gorm.DB) *AgentContractRepository {
	return &AgentContractRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建合约
func (r *AgentContractRepository) Create(ctx context.Context, contract *model.AgentContract) error {
	return r.DB(ctx).Create(contract).Error
}

// Update 更新合约
func (r *AgentContractRepository) Update(ctx context.Context, contract *model.AgentContract) error {
	return r.DB(ctx).Save(contract).Error
}

// GetByID 根据ID获取合约
func (r *AgentContractRepository) GetByID(ctx context.Context, id uint) (*model.AgentContract, error) {
	var contract model.AgentContract
	err := r.DB(ctx).Preload("Team").Preload("User").First(&contract, id).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

// GetByTeamAndRound 获取队伍在某轮次的合约
func (r *AgentContractRepository) GetByTeamAndRound(ctx context.Context, teamID, roundID uint) (*model.AgentContract, error) {
	var contract model.AgentContract
	err := r.DB(ctx).Where("team_id = ? AND round_id = ?", teamID, roundID).First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

// ListByRound 获取轮次的所有合约
func (r *AgentContractRepository) ListByRound(ctx context.Context, roundID uint) ([]model.AgentContract, error) {
	var contracts []model.AgentContract
	err := r.DB(ctx).Where("round_id = ?", roundID).
		Preload("Team").
		Find(&contracts).Error
	return contracts, err
}

// ListByContest 获取竞赛的所有合约
func (r *AgentContractRepository) ListByContest(ctx context.Context, contestID uint) ([]model.AgentContract, error) {
	var contracts []model.AgentContract
	err := r.DB(ctx).Where("contest_id = ?", contestID).
		Preload("Team").
		Find(&contracts).Error
	return contracts, err
}

// GetLatestByContestAndTeam 获取队伍在竞赛中的最新合约，用于展示当前队伍的智能体状态。
func (r *AgentContractRepository) GetLatestByContestAndTeam(ctx context.Context, contestID, teamID uint) (*model.AgentContract, error) {
	var contract model.AgentContract
	err := r.DB(ctx).
		Where("contest_id = ? AND team_id = ?", contestID, teamID).
		Order("round_id DESC, version DESC, created_at DESC").
		First(&contract).Error
	if err != nil {
		return nil, err
	}
	return &contract, nil
}

// UpdateStatus 更新合约状态
func (r *AgentContractRepository) UpdateStatus(ctx context.Context, id uint, status, address string) error {
	updates := map[string]interface{}{"status": status}
	if address != "" {
		updates["contract_address"] = address
	}
	return r.DB(ctx).Model(&model.AgentContract{}).Where("id = ?", id).Updates(updates).Error
}

// AgentBattleEventRepository 对战事件仓库
type AgentBattleEventRepository struct {
	*BaseRepository
}

// NewAgentBattleEventRepository 创建对战事件仓库
func NewAgentBattleEventRepository(db *gorm.DB) *AgentBattleEventRepository {
	return &AgentBattleEventRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建事件
func (r *AgentBattleEventRepository) Create(ctx context.Context, event *model.AgentBattleEvent) error {
	return r.DB(ctx).Create(event).Error
}

// BatchCreate 批量创建事件
func (r *AgentBattleEventRepository) BatchCreate(ctx context.Context, events []model.AgentBattleEvent) error {
	return r.DB(ctx).CreateInBatches(events, 100).Error
}

// List 获取事件列表
func (r *AgentBattleEventRepository) List(ctx context.Context, roundID uint, eventType string, page, pageSize int) ([]model.AgentBattleEvent, int64, error) {
	var events []model.AgentBattleEvent
	var total int64

	query := r.DB(ctx).Model(&model.AgentBattleEvent{}).Where("round_id = ?", roundID)
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("block_number ASC, tx_index ASC").
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// AgentBattleScoreRepository 对战分数仓库
type AgentBattleScoreRepository struct {
	*BaseRepository
}

// NewAgentBattleScoreRepository 创建对战分数仓库
func NewAgentBattleScoreRepository(db *gorm.DB) *AgentBattleScoreRepository {
	return &AgentBattleScoreRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Upsert 创建或更新分数
func (r *AgentBattleScoreRepository) Upsert(ctx context.Context, score *model.AgentBattleScore) error {
	return r.DB(ctx).Save(score).Error
}

// GetByTeamAndRound 获取队伍在某轮次的分数
func (r *AgentBattleScoreRepository) GetByTeamAndRound(ctx context.Context, teamID, roundID uint) (*model.AgentBattleScore, error) {
	var score model.AgentBattleScore
	err := r.DB(ctx).Where("team_id = ? AND round_id = ?", teamID, roundID).First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// GetScoreboard 获取轮次排行榜
func (r *AgentBattleScoreRepository) GetScoreboard(ctx context.Context, roundID uint) ([]model.AgentBattleScore, error) {
	var scores []model.AgentBattleScore
	err := r.DB(ctx).Where("round_id = ?", roundID).
		Preload("Team").
		Order("score DESC, rank ASC, token_balance DESC").
		Find(&scores).Error
	return scores, err
}

// UpdateScore 更新分数
func (r *AgentBattleScoreRepository) UpdateScore(ctx context.Context, teamID, roundID uint, balance int64, attacks, defenses int) error {
	return r.DB(ctx).Model(&model.AgentBattleScore{}).
		Where("team_id = ? AND round_id = ?", teamID, roundID).
		Updates(map[string]interface{}{
			"final_balance":   balance,
			"attack_success":  gorm.Expr("attack_success + ?", attacks),
			"defense_success": gorm.Expr("defense_success + ?", defenses),
		}).Error
}
