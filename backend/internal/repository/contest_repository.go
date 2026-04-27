package repository

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

// ContestRepository 竞赛仓库
type ContestRepository struct {
	*BaseRepository
}

// NewContestRepository 创建竞赛仓库
func NewContestRepository(db *gorm.DB) *ContestRepository {
	return &ContestRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建竞赛
func (r *ContestRepository) Create(ctx context.Context, contest *model.Contest) error {
	return r.DB(ctx).Create(contest).Error
}

// Update 更新竞赛
func (r *ContestRepository) Update(ctx context.Context, contest *model.Contest) error {
	return r.DB(ctx).Save(contest).Error
}

// Delete 删除竞赛（软删除）
func (r *ContestRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Contest{}, id).Error
}

// GetByID 根据ID获取竞赛
func (r *ContestRepository) GetByID(ctx context.Context, id uint) (*model.Contest, error) {
	var contest model.Contest
	err := r.DB(ctx).Preload("School").Preload("Creator").First(&contest, id).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

// GetWithChallenges 获取竞赛及其题目
func (r *ContestRepository) GetWithChallenges(ctx context.Context, id uint) (*model.Contest, error) {
	var contest model.Contest
	err := r.DB(ctx).
		Preload("Creator").
		Preload("Challenges", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_visible = ?", true).Order("sort_order ASC")
		}).
		Preload("Challenges.Challenge").
		First(&contest, id).Error
	if err != nil {
		return nil, err
	}
	return &contest, nil
}

// List 获取竞赛列表
func (r *ContestRepository) List(ctx context.Context, schoolID uint, contestType, level, status, keyword string, isPublic *bool, page, pageSize int) ([]model.Contest, int64, error) {
	var contests []model.Contest
	var total int64

	query := r.buildContestListQuery(ctx, schoolID, contestType, level, status, keyword, isPublic)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("School").Preload("Creator").
		Order("start_time DESC").
		Find(&contests).Error
	if err != nil {
		return nil, 0, err
	}

	return contests, total, nil
}

func (r *ContestRepository) ListAll(ctx context.Context, schoolID uint, contestType, level, status, keyword string, isPublic *bool) ([]model.Contest, error) {
	var contests []model.Contest

	err := r.buildContestListQuery(ctx, schoolID, contestType, level, status, keyword, isPublic).
		Preload("School").Preload("Creator").
		Order("start_time DESC").
		Find(&contests).Error
	if err != nil {
		return nil, err
	}

	return contests, nil
}

func (r *ContestRepository) buildContestListQuery(ctx context.Context, schoolID uint, contestType, level, status, keyword string, isPublic *bool) *gorm.DB {
	query := r.DB(ctx).Model(&model.Contest{})

	if schoolID > 0 {
		query = query.Where("school_id = ? OR is_public = ? OR school_id IS NULL", schoolID, true)
	}
	if contestType != "" {
		query = query.Where("type = ?", contestType)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if keyword != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	return query
}

// UpdateStatus 更新竞赛状态

// ChallengeRepository 题目仓库
type ChallengeRepository struct {
	*BaseRepository
}

// NewChallengeRepository 创建题目仓库
func NewChallengeRepository(db *gorm.DB) *ChallengeRepository {
	return &ChallengeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建题目
func (r *ChallengeRepository) Create(ctx context.Context, challenge *model.Challenge) error {
	return r.DB(ctx).Create(challenge).Error
}

// Update 更新题目
func (r *ChallengeRepository) Update(ctx context.Context, challenge *model.Challenge) error {
	return r.DB(ctx).Save(challenge).Error
}

// Delete 删除题目（软删除）
func (r *ChallengeRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Challenge{}, id).Error
}

// GetByID 根据ID获取题目
func (r *ChallengeRepository) GetByID(ctx context.Context, id uint) (*model.Challenge, error) {
	var challenge model.Challenge
	err := r.DB(ctx).Preload("Creator").First(&challenge, id).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

// List 获取题目列表
func (r *ChallengeRepository) List(
	ctx context.Context,
	role string,
	schoolID uint,
	creatorID uint,
	category string,
	difficulty *int,
	sourceType string,
	status string,
	keyword string,
	isPublic *bool,
	page int,
	pageSize int,
) ([]model.Challenge, int64, error) {
	var challenges []model.Challenge
	var total int64

	query := r.DB(ctx).Model(&model.Challenge{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if difficulty != nil {
		query = query.Where("difficulty = ?", *difficulty)
	}
	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if false && isPublic != nil {
		if creatorID > 0 {
			// 非管理员：查看公开题目 + 自己创建的题目
			query = query.Where("is_public = ? OR creator_id = ?", *isPublic, creatorID)
		} else {
			query = query.Where("is_public = ?", *isPublic)
		}
	}
	if keyword != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	switch role {
	case model.RolePlatformAdmin:
		if isPublic != nil {
			query = query.Where("is_public = ?", *isPublic)
		}
	case model.RoleSchoolAdmin:
		if isPublic != nil {
			if *isPublic {
				query = query.Where("is_public = ?", true)
			} else {
				query = query.Where("school_id = ? AND is_public = ?", schoolID, false)
			}
		} else {
			query = query.Where("school_id = ? OR is_public = ?", schoolID, true)
		}
	case model.RoleTeacher:
		if isPublic != nil {
			if *isPublic {
				query = query.Where("is_public = ?", true)
			} else {
				query = query.Where("creator_id = ? AND is_public = ?", creatorID, false)
			}
		} else {
			query = query.Where("creator_id = ? OR is_public = ?", creatorID, true)
		}
	default:
		query = query.Where("is_public = ?", true)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Creator").
		Order("created_at DESC").
		Find(&challenges).Error
	if err != nil {
		return nil, 0, err
	}

	return challenges, total, nil
}

// IncrementSolveCount 增加解决次数
func (r *ChallengeRepository) IncrementSolveCount(ctx context.Context, id uint) error {
	return r.DB(ctx).Model(&model.Challenge{}).Where("id = ?", id).
		Update("solve_count", gorm.Expr("solve_count + 1")).Error
}

// IncrementAttemptCount 增加尝试次数
func (r *ChallengeRepository) IncrementAttemptCount(ctx context.Context, id uint) error {
	return r.DB(ctx).Model(&model.Challenge{}).Where("id = ?", id).
		Update("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

// ListPublishRequests 获取题目公开申请列表
func (r *ChallengeRepository) ListPublishRequests(ctx context.Context, schoolID uint, status string, page, pageSize int) ([]model.ChallengePublishRequest, int64, error) {
	var requests []model.ChallengePublishRequest
	var total int64

	query := r.DB(ctx).Model(&model.ChallengePublishRequest{})
	if schoolID > 0 {
		query = query.Joins("JOIN challenges ON challenges.id = challenge_publish_requests.challenge_id").
			Where("challenges.school_id = ?", schoolID)
	}
	if status != "" {
		query = query.Where("challenge_publish_requests.status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Scopes(Paginate(page, pageSize)).
		Preload("Challenge").Preload("Applicant").
		Order("challenge_publish_requests.created_at DESC").
		Find(&requests).Error
	return requests, total, err
}

// GetPublishRequestByID 根据ID获取题目公开申请
func (r *ChallengeRepository) GetPublishRequestByID(ctx context.Context, id uint) (*model.ChallengePublishRequest, error) {
	var req model.ChallengePublishRequest
	err := r.DB(ctx).Preload("Challenge").Preload("Applicant").First(&req, id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// UpdatePublishRequest 更新题目公开申请
func (r *ChallengeRepository) UpdatePublishRequest(ctx context.Context, req *model.ChallengePublishRequest) error {
	return r.DB(ctx).Save(req).Error
}

// CreatePublishRequest 创建题目公开申请
func (r *ChallengeRepository) CreatePublishRequest(ctx context.Context, req *model.ChallengePublishRequest) error {
	return r.DB(ctx).Create(req).Error
}

// ContestChallengeRepository 竞赛题目关联仓库
type ContestChallengeRepository struct {
	*BaseRepository
}

// NewContestChallengeRepository 创建竞赛题目关联仓库
func NewContestChallengeRepository(db *gorm.DB) *ContestChallengeRepository {
	return &ContestChallengeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建关联
func (r *ContestChallengeRepository) Create(ctx context.Context, cc *model.ContestChallenge) error {
	return r.DB(ctx).Create(cc).Error
}

// Update 更新关联
func (r *ContestChallengeRepository) Update(ctx context.Context, cc *model.ContestChallenge) error {
	return r.DB(ctx).Save(cc).Error
}

// Delete 删除关联
func (r *ContestChallengeRepository) Delete(ctx context.Context, contestID, challengeID uint) error {
	return r.DB(ctx).Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Delete(&model.ContestChallenge{}).Error
}

// GetByID 根据竞赛和题目ID获取关联
func (r *ContestChallengeRepository) GetByID(ctx context.Context, contestID, challengeID uint) (*model.ContestChallenge, error) {
	var cc model.ContestChallenge
	err := r.DB(ctx).Preload("Challenge").
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&cc).Error
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

// ListByContest 获取竞赛的题目列表
func (r *ContestChallengeRepository) ListByContest(ctx context.Context, contestID uint, visibleOnly bool) ([]model.ContestChallenge, error) {
	var list []model.ContestChallenge

	query := r.DB(ctx).Where("contest_id = ?", contestID)
	if visibleOnly {
		query = query.Where("is_visible = ?", true)
	}

	err := query.Preload("Challenge").Order("sort_order ASC").Find(&list).Error
	return list, err
}

// UpdateCurrentPoints 更新当前分值（动态分数）
func (r *ContestChallengeRepository) UpdateCurrentPoints(ctx context.Context, contestID, challengeID uint, points int) error {
	return r.DB(ctx).Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Update("current_points", points).Error
}

// CountByContest 统计竞赛题目数量
func (r *ContestChallengeRepository) CountByContest(ctx context.Context, contestID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ContestChallenge{}).Where("contest_id = ?", contestID).Count(&count).Error
	return count, err
}

// TeamRepository 队伍仓库
type TeamRepository struct {
	*BaseRepository
}

// NewTeamRepository 创建队伍仓库
func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建队伍
func (r *TeamRepository) Create(ctx context.Context, team *model.Team) error {
	return r.DB(ctx).Create(team).Error
}

// Update 更新队伍
func (r *TeamRepository) Update(ctx context.Context, team *model.Team) error {
	return r.DB(ctx).Save(team).Error
}

// Delete 删除队伍（软删除）
func (r *TeamRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Team{}, id).Error
}

// GetByID 根据ID获取队伍
func (r *TeamRepository) GetByID(ctx context.Context, id uint) (*model.Team, error) {
	var team model.Team
	err := r.DB(ctx).Preload("Leader").Preload("Members.User").First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetByToken 根据Token获取队伍
func (r *TeamRepository) GetByToken(ctx context.Context, token string) (*model.Team, error) {
	var team model.Team
	err := r.DB(ctx).Where("token = ?", token).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetUserTeam 获取用户在竞赛中的队伍
func (r *TeamRepository) GetUserTeam(ctx context.Context, contestID, userID uint) (*model.Team, error) {
	var team model.Team
	err := r.DB(ctx).
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("teams.contest_id = ? AND team_members.user_id = ?", contestID, userID).
		Preload("Leader").Preload("Members.User").
		First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// List 获取队伍列表
func (r *TeamRepository) List(ctx context.Context, contestID uint, status, keyword string, page, pageSize int) ([]model.Team, int64, error) {
	var teams []model.Team
	var total int64

	query := r.DB(ctx).Model(&model.Team{}).Where("contest_id = ?", contestID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Leader").Preload("Members.User").
		Order("created_at ASC").
		Find(&teams).Error
	if err != nil {
		return nil, 0, err
	}

	return teams, total, nil
}

// CountMembers 统计队伍成员数
func (r *TeamRepository) CountMembers(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

// CountParticipants 统计竞赛参与人数
func (r *TeamRepository) CountParticipants(ctx context.Context, contestID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.TeamMember{}).
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("teams.contest_id = ?", contestID).
		Count(&count).Error
	return count, err
}

// TeamMemberRepository 队伍成员仓库
type TeamMemberRepository struct {
	*BaseRepository
}

// NewTeamMemberRepository 创建队伍成员仓库
func NewTeamMemberRepository(db *gorm.DB) *TeamMemberRepository {
	return &TeamMemberRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建成员
func (r *TeamMemberRepository) Create(ctx context.Context, member *model.TeamMember) error {
	return r.DB(ctx).Create(member).Error
}

// Delete 删除成员
func (r *TeamMemberRepository) Delete(ctx context.Context, teamID, userID uint) error {
	return r.DB(ctx).Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&model.TeamMember{}).Error
}

// Exists 检查是否为成员
func (r *TeamMemberRepository) Exists(ctx context.Context, teamID, userID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count).Error
	return count > 0, err
}

// IsInAnyTeam 检查用户是否已在竞赛的任何队伍中
func (r *TeamMemberRepository) IsInAnyTeam(ctx context.Context, contestID, userID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&model.TeamMember{}).
		Joins("JOIN teams ON teams.id = team_members.team_id").
		Where("teams.contest_id = ? AND team_members.user_id = ?", contestID, userID).
		Count(&count).Error
	return count > 0, err
}

// ListByTeam 获取队伍成员列表
func (r *TeamMemberRepository) ListByTeam(ctx context.Context, teamID uint) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.DB(ctx).Where("team_id = ?", teamID).Preload("User").Find(&members).Error
	return members, err
}

// GetUserTeams 获取用户所在的所有队伍
func (r *TeamMemberRepository) GetUserTeams(ctx context.Context, userID uint) ([]model.TeamMember, error) {
	var members []model.TeamMember
	err := r.DB(ctx).Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

// ContestSubmissionRepository 竞赛提交仓库
type ContestSubmissionRepository struct {
	*BaseRepository
}

// NewContestSubmissionRepository 创建竞赛提交仓库
func NewContestSubmissionRepository(db *gorm.DB) *ContestSubmissionRepository {
	return &ContestSubmissionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建提交
func (r *ContestSubmissionRepository) Create(ctx context.Context, sub *model.ContestSubmission) error {
	return r.DB(ctx).Create(sub).Error
}

// GetByID 根据ID获取提交
func (r *ContestSubmissionRepository) GetByID(ctx context.Context, id uint) (*model.ContestSubmission, error) {
	var sub model.ContestSubmission
	err := r.DB(ctx).First(&sub, id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// HasSolved 检查是否已解决
func (r *ContestSubmissionRepository) HasSolved(ctx context.Context, contestID, challengeID uint, teamID *uint, userID uint) (bool, error) {
	var count int64
	query := r.DB(ctx).Model(&model.ContestSubmission{}).
		Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", contestID, challengeID, true)

	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// GetFirstBlood 获取一血
func (r *ContestSubmissionRepository) GetFirstBlood(ctx context.Context, contestID, challengeID uint) (*model.ContestSubmission, error) {
	var sub model.ContestSubmission
	err := r.DB(ctx).
		Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", contestID, challengeID, true).
		Order("submitted_at ASC").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetCorrectSubmission 获取当前队伍或用户的正确提交记录。
func (r *ContestSubmissionRepository) GetCorrectSubmission(ctx context.Context, contestID, challengeID uint, teamID *uint, userID uint) (*model.ContestSubmission, error) {
	var sub model.ContestSubmission
	query := r.DB(ctx).
		Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", contestID, challengeID, true)

	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Order("submitted_at ASC, id ASC").First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// List 获取提交列表
func (r *ContestSubmissionRepository) List(ctx context.Context, contestID, challengeID uint, teamID *uint, userID uint, page, pageSize int) ([]model.ContestSubmission, int64, error) {
	var subs []model.ContestSubmission
	var total int64

	query := r.DB(ctx).Model(&model.ContestSubmission{}).Where("contest_id = ?", contestID)

	if challengeID > 0 {
		query = query.Where("challenge_id = ?", challengeID)
	}
	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("submitted_at DESC").
		Find(&subs).Error
	if err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

// CountSolves 统计题目解决次数
func (r *ContestSubmissionRepository) CountSolves(ctx context.Context, contestID, challengeID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ContestSubmission{}).
		Where("contest_id = ? AND challenge_id = ? AND is_correct = ?", contestID, challengeID, true).
		Count(&count).Error
	return count, err
}

// ContestScoreRepository 竞赛分数仓库
type ContestScoreRepository struct {
	*BaseRepository
}

// NewContestScoreRepository 创建竞赛分数仓库
func NewContestScoreRepository(db *gorm.DB) *ContestScoreRepository {
	return &ContestScoreRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Upsert 创建或更新分数
func (r *ContestScoreRepository) Upsert(ctx context.Context, score *model.ContestScore) error {
	return r.DB(ctx).Save(score).Error
}

// GetByTeamOrUser 获取队伍或用户分数
func (r *ContestScoreRepository) GetByTeamOrUser(ctx context.Context, contestID uint, teamID *uint, userID uint) (*model.ContestScore, error) {
	var score model.ContestScore
	query := r.DB(ctx).Where("contest_id = ?", contestID)

	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	} else {
		query = query.Where("user_id = ? AND team_id IS NULL", userID)
	}

	err := query.First(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}

// GetScoreboard 获取排行榜
func (r *ContestScoreRepository) GetScoreboard(ctx context.Context, contestID uint, page, pageSize int) ([]model.ContestScore, int64, error) {
	var scores []model.ContestScore
	var total int64

	query := r.DB(ctx).Model(&model.ContestScore{}).Where("contest_id = ?", contestID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("Team").Preload("User").
		Order("total_score DESC, last_solve_at ASC").
		Find(&scores).Error
	if err != nil {
		return nil, 0, err
	}

	return scores, total, nil
}

// UpdateRanks 更新排名
func (r *ContestScoreRepository) UpdateRanks(ctx context.Context, contestID uint) error {
	// 使用窗口函数更新排名
	sql := `
		UPDATE contest_scores 
		SET rank = subquery.new_rank
		FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY total_score DESC, last_solve_at ASC) as new_rank
			FROM contest_scores
			WHERE contest_id = ?
		) AS subquery
		WHERE contest_scores.id = subquery.id
	`
	return r.DB(ctx).Exec(sql, contestID).Error
}

// AddScore 增加分数
func (r *ContestScoreRepository) AddScore(ctx context.Context, contestID uint, teamID *uint, userID uint, points int) error {
	// 团队赛按 team_id 聚合，个人赛按 user_id 聚合，避免同一队伍被拆成多条分数记录。
	query := r.DB(ctx).Model(&model.ContestScore{}).Where("contest_id = ?", contestID)
	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	} else {
		query = query.Where("user_id = ? AND team_id IS NULL", userID)
	}

	result := query.Updates(map[string]interface{}{
		"total_score":   gorm.Expr("total_score + ?", points),
		"solve_count":   gorm.Expr("solve_count + 1"),
		"last_solve_at": gorm.Expr("NOW()"),
	})
	if result.Error != nil {
		return result.Error
	}

	// 首次得分时创建一条新的排行榜记录。
	if result.RowsAffected == 0 {
		score := &model.ContestScore{
			ContestID:  contestID,
			TeamID:     teamID,
			UserID:     userID,
			TotalScore: points,
			SolveCount: 1,
		}
		now := time.Now()
		score.LastSolveAt = &now
		return r.DB(ctx).Create(score).Error
	}
	return nil
}

// ============ ChallengeEnv ============

// ChallengeEnvRepository 题目环境仓库
type ChallengeEnvRepository struct {
	BaseRepository
}

func NewChallengeEnvRepository(db *gorm.DB) *ChallengeEnvRepository {
	return &ChallengeEnvRepository{BaseRepository{db: db}}
}

func (r *ChallengeEnvRepository) Create(ctx context.Context, env *model.ChallengeEnv) error {
	return r.DB(ctx).Create(env).Error
}

func (r *ChallengeEnvRepository) Update(ctx context.Context, env *model.ChallengeEnv) error {
	return r.DB(ctx).Save(env).Error
}

func (r *ChallengeEnvRepository) GetActiveByUserAndChallenge(ctx context.Context, userID, contestID, challengeID uint) (*model.ChallengeEnv, error) {
	var env model.ChallengeEnv
	err := r.DB(ctx).Where("user_id = ? AND contest_id = ? AND challenge_id = ? AND status IN ?",
		userID, contestID, challengeID, model.ActiveChallengeEnvStatuses()).Order("id DESC").First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *ChallengeEnvRepository) GetByEnvID(ctx context.Context, envID string) (*model.ChallengeEnv, error) {
	var env model.ChallengeEnv
	err := r.DB(ctx).Where("env_id = ?", envID).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}
