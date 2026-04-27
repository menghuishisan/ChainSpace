package repository

import (
	"context"
	"time"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
)

func preloadExperimentDetail(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Chapter").
		Preload("Creator").
		Preload("Workspace").
		Preload("Workspace.Tools", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Topology").
		Preload("Topology.ExposedEntries", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Tools", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("InitScripts", func(tx *gorm.DB) *gorm.DB { return tx.Order("scope_type ASC, scope_key ASC, sort_order ASC") }).
		Preload("Collaboration").
		Preload("Collaboration.Roles", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Collaboration.Roles.NodeAssignments", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Collaboration.Roles.ToolAssignments", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Nodes", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Nodes.Ports", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Nodes.Tools", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Services", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Services.Ports", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Services.EnvVars", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Assets", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") }).
		Preload("Checkpoints", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") })
}

func preloadExperimentEnvDetail(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Experiment", preloadExperimentDetail).
		Preload("Session").
		Preload("Session.Members").
		Preload("Session.Members.User").
		Preload("User").
		Preload("RuntimeInstances").
		Preload("RuntimeInstances.Tools", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") })
}

func preloadSubmissionDetail(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Experiment", preloadExperimentDetail).
		Preload("Student").
		Preload("Grader").
		Preload("CheckResults", func(tx *gorm.DB) *gorm.DB { return tx.Order("sort_order ASC") })
}

// ExperimentRepository 实验仓库
type ExperimentRepository struct {
	*BaseRepository
}

// NewExperimentRepository 创建实验仓库
func NewExperimentRepository(db *gorm.DB) *ExperimentRepository {
	return &ExperimentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建实验
func (r *ExperimentRepository) Create(ctx context.Context, exp *model.Experiment) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Workspace.Tools", "Topology.ExposedEntries", "Collaboration.Roles.NodeAssignments", "Collaboration.Roles.ToolAssignments", "Nodes.Ports", "Nodes.Tools", "Services.Ports", "Services.EnvVars").Create(exp).Error; err != nil {
			return err
		}
		return persistExperimentRelations(tx, exp)
	})
}

// Update 更新实验
func (r *ExperimentRepository) Update(ctx context.Context, exp *model.Experiment) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Experiment{}).
			Where("id = ?", exp.ID).
			Updates(map[string]interface{}{
				"school_id":        exp.SchoolID,
				"chapter_id":       exp.ChapterID,
				"creator_id":       exp.CreatorID,
				"title":            exp.Title,
				"description":      exp.Description,
				"type":             exp.Type,
				"mode":             exp.Mode,
				"difficulty":       exp.Difficulty,
				"max_score":        exp.MaxScore,
				"pass_score":       exp.PassScore,
				"auto_grade":       exp.AutoGrade,
				"grading_strategy": exp.GradingStrategy,
				"estimated_time":   exp.EstimatedTime,
				"sort_order":       exp.SortOrder,
				"status":           exp.Status,
				"start_time":       exp.StartTime,
				"end_time":         exp.EndTime,
				"allow_late":       exp.AllowLate,
				"late_deduction":   exp.LateDeduction,
			}).Error; err != nil {
			return err
		}
		return persistExperimentRelations(tx, exp)
	})
}

// Delete 删除实验（软删除）
func (r *ExperimentRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.Experiment{}, id).Error
}

// GetByID 根据ID获取实验
func (r *ExperimentRepository) GetByID(ctx context.Context, id uint) (*model.Experiment, error) {
	var exp model.Experiment
	err := preloadExperimentDetail(r.DB(ctx)).First(&exp, id).Error
	if err != nil {
		return nil, err
	}
	return &exp, nil
}

// List 获取实验列表
func (r *ExperimentRepository) List(ctx context.Context, schoolID, courseID, chapterID uint, expType, status, keyword string, page, pageSize int) ([]model.Experiment, int64, error) {
	var exps []model.Experiment
	var total int64

	query := r.DB(ctx).
		Model(&model.Experiment{}).
		Joins("JOIN chapters ON chapters.id = experiments.chapter_id AND chapters.deleted_at IS NULL")

	if schoolID > 0 {
		query = query.Where("experiments.school_id = ?", schoolID)
	}
	if courseID > 0 {
		query = query.Where("chapters.course_id = ?", courseID)
	}
	if chapterID > 0 {
		query = query.Where("experiments.chapter_id = ?", chapterID)
	}
	if expType != "" {
		query = query.Where("experiments.type = ?", expType)
	}
	if status != "" {
		query = query.Where("experiments.status = ?", status)
	}
	if keyword != "" {
		query = query.Where("experiments.title ILIKE ? OR experiments.description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Scopes(preloadExperimentDetail).
		Order("experiments.sort_order ASC, experiments.created_at DESC").
		Find(&exps).Error
	if err != nil {
		return nil, 0, err
	}

	return exps, total, nil
}

// ListByChapter 获取章节的实验列表
func (r *ExperimentRepository) ListByChapter(ctx context.Context, chapterID uint, status string) ([]model.Experiment, error) {
	var exps []model.Experiment

	query := r.DB(ctx).Where("chapter_id = ?", chapterID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := preloadExperimentDetail(query.Order("sort_order ASC")).Find(&exps).Error
	return exps, err
}

// CountByChapter 统计章节实验数
func (r *ExperimentRepository) CountByChapter(ctx context.Context, chapterID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Experiment{}).Where("chapter_id = ?", chapterID).Count(&count).Error
	return count, err
}

// ExperimentEnvRepository 实验环境仓库
type ExperimentEnvRepository struct {
	*BaseRepository
}

// NewExperimentEnvRepository 创建实验环境仓库
func NewExperimentEnvRepository(db *gorm.DB) *ExperimentEnvRepository {
	return &ExperimentEnvRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建实验环境
func (r *ExperimentEnvRepository) Create(ctx context.Context, env *model.ExperimentEnv) error {
	return r.DB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(env).Error
}

// Update 更新实验环境
func (r *ExperimentEnvRepository) Update(ctx context.Context, env *model.ExperimentEnv) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ExperimentEnv{}).
			Where("id = ?", env.ID).
			Updates(map[string]interface{}{
				"env_id":               env.EnvID,
				"experiment_id":        env.ExperimentID,
				"session_id":           env.SessionID,
				"user_id":              env.UserID,
				"school_id":            env.SchoolID,
				"status":               env.Status,
				"session_mode":         env.SessionMode,
				"primary_instance_key": env.PrimaryInstanceKey,
				"started_at":           env.StartedAt,
				"expires_at":           env.ExpiresAt,
				"extend_count":         env.ExtendCount,
				"snapshot_at":          env.SnapshotAt,
				"snapshot_url":         env.SnapshotURL,
				"error_message":        env.ErrorMessage,
			}).Error; err != nil {
			return err
		}
		return persistExperimentRuntimeRelations(tx, env)
	})
}

// GetByID 根据ID获取实验环境
func (r *ExperimentEnvRepository) GetByID(ctx context.Context, id uint) (*model.ExperimentEnv, error) {
	var env model.ExperimentEnv
	err := preloadExperimentEnvDetail(r.DB(ctx)).First(&env, id).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// GetByEnvID 根据EnvID获取实验环境
func (r *ExperimentEnvRepository) GetByEnvID(ctx context.Context, envID string) (*model.ExperimentEnv, error) {
	var env model.ExperimentEnv
	err := preloadExperimentEnvDetail(r.DB(ctx)).Where("env_id = ?", envID).First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

// GetActiveByUser 获取用户活跃的实验环境
func (r *ExperimentEnvRepository) GetActiveByUser(ctx context.Context, userID, experimentID uint) (*model.ExperimentEnv, error) {
	var env model.ExperimentEnv
	err := r.DB(ctx).
		Where("user_id = ? AND experiment_id = ? AND status IN ?", userID, experimentID, model.ActiveExperimentEnvStatuses()).
		Scopes(preloadExperimentEnvDetail).
		First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *ExperimentEnvRepository) GetReusableCollaborationEnv(ctx context.Context, experimentID uint) (*model.ExperimentEnv, error) {
	var env model.ExperimentEnv
	err := r.DB(ctx).
		Where("experiment_id = ? AND status IN ?", experimentID, model.ActiveExperimentEnvStatuses()).
		Where("session_mode = ?", model.ExperimentModeCollaboration).
		Scopes(preloadExperimentEnvDetail).
		Order("created_at ASC").
		First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *ExperimentEnvRepository) CountActiveBySession(ctx context.Context, sessionID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ExperimentEnv{}).
		Where("session_id = ? AND status IN ?", sessionID, model.ActiveExperimentEnvStatuses()).
		Count(&count).Error
	return count, err
}

func (r *ExperimentEnvRepository) GetFirstActiveBySession(ctx context.Context, sessionID uint, excludeEnvID string) (*model.ExperimentEnv, error) {
	var env model.ExperimentEnv
	query := r.DB(ctx).
		Where("session_id = ? AND status IN ?", sessionID, model.ActiveExperimentEnvStatuses())
	if excludeEnvID != "" {
		query = query.Where("env_id <> ?", excludeEnvID)
	}
	err := query.Scopes(preloadExperimentEnvDetail).Order("created_at ASC").First(&env).Error
	if err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *ExperimentEnvRepository) ListBySession(ctx context.Context, sessionID uint) ([]model.ExperimentEnv, error) {
	var envs []model.ExperimentEnv
	err := r.DB(ctx).
		Where("session_id = ?", sessionID).
		Scopes(preloadExperimentEnvDetail).
		Order("created_at ASC").
		Find(&envs).Error
	return envs, err
}

// List 获取实验环境列表
func (r *ExperimentEnvRepository) List(ctx context.Context, schoolID, experimentID, userID uint, status string, page, pageSize int) ([]model.ExperimentEnv, int64, error) {
	var envs []model.ExperimentEnv
	var total int64

	query := r.DB(ctx).Model(&model.ExperimentEnv{})

	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if experimentID > 0 {
		query = query.Where("experiment_id = ?", experimentID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Scopes(preloadExperimentEnvDetail).
		Order("created_at DESC").
		Find(&envs).Error
	if err != nil {
		return nil, 0, err
	}

	return envs, total, nil
}

// ListExpiring 获取即将过期的实验环境
func (r *ExperimentEnvRepository) ListExpiring(ctx context.Context, minutes int) ([]model.ExperimentEnv, error) {
	var envs []model.ExperimentEnv
	expiryThreshold := time.Now().Add(time.Duration(minutes) * time.Minute)
	err := r.DB(ctx).
		Where("status = ? AND expires_at <= ? AND expires_at > NOW()", model.EnvStatusRunning, expiryThreshold).
		Scopes(preloadExperimentEnvDetail).
		Find(&envs).Error
	return envs, err
}

// ListExpired 获取已过期的实验环境
func (r *ExperimentEnvRepository) ListExpired(ctx context.Context) ([]model.ExperimentEnv, error) {
	var envs []model.ExperimentEnv
	err := r.DB(ctx).
		Where("status = ? AND expires_at <= NOW()", model.EnvStatusRunning).
		Scopes(preloadExperimentEnvDetail).
		Find(&envs).Error
	return envs, err
}

// UpdateStatus 更新环境状态
func (r *ExperimentEnvRepository) UpdateStatus(ctx context.Context, envID string, status string) error {
	return r.DB(ctx).Model(&model.ExperimentEnv{}).Where("env_id = ?", envID).Update("status", status).Error
}

// ListByStatus 根据状态获取实验环境列表
func (r *ExperimentEnvRepository) ListByStatus(ctx context.Context, status string) ([]model.ExperimentEnv, error) {
	var envs []model.ExperimentEnv
	err := r.DB(ctx).Where("status = ?", status).Find(&envs).Error
	return envs, err
}

// UpdateExpireTime 更新过期时间
func (r *ExperimentEnvRepository) UpdateExpireTime(ctx context.Context, envID string, expiresAt interface{}) error {
	return r.DB(ctx).Model(&model.ExperimentEnv{}).Where("env_id = ?", envID).
		Updates(map[string]interface{}{
			"expires_at":   expiresAt,
			"extend_count": gorm.Expr("extend_count + 1"),
		}).Error
}

// CountUserSnapshots 统计用户在指定实验中的快照数量
func (r *ExperimentEnvRepository) CountUserSnapshots(ctx context.Context, userID, experimentID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ExperimentEnv{}).
		Where("user_id = ? AND experiment_id = ? AND snapshot_url IS NOT NULL AND snapshot_url != ''", userID, experimentID).
		Count(&count).Error
	return count, err
}

// CountActiveByUser 统计用户活跃环境数量
func (r *ExperimentEnvRepository) CountActiveByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ExperimentEnv{}).
		Where("user_id = ? AND status IN ?", userID, []string{model.EnvStatusRunning, model.EnvStatusCreating}).
		Count(&count).Error
	return count, err
}

// CountActiveBySchool 统计学校活跃环境数量
func (r *ExperimentEnvRepository) CountActiveBySchool(ctx context.Context, schoolID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.ExperimentEnv{}).
		Where("school_id = ? AND status IN ?", schoolID, []string{model.EnvStatusRunning, model.EnvStatusCreating}).
		Count(&count).Error
	return count, err
}

// SubmissionRepository 提交仓库
type SubmissionRepository struct {
	*BaseRepository
}

type ExperimentSessionRepository struct {
	*BaseRepository
}

func NewExperimentSessionRepository(db *gorm.DB) *ExperimentSessionRepository {
	return &ExperimentSessionRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ExperimentSessionRepository) Create(ctx context.Context, session *model.ExperimentSession) error {
	return r.DB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(session).Error
}

func (r *ExperimentSessionRepository) Update(ctx context.Context, session *model.ExperimentSession) error {
	return r.DB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Updates(session).Error
}

func (r *ExperimentSessionRepository) GetByID(ctx context.Context, id uint) (*model.ExperimentSession, error) {
	var session model.ExperimentSession
	err := r.DB(ctx).
		Preload("Experiment").
		Preload("Members").
		Preload("Members.User").
		Preload("Messages", func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at ASC") }).
		Preload("Messages.User").
		First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExperimentSessionRepository) GetBySessionKey(ctx context.Context, sessionKey string) (*model.ExperimentSession, error) {
	var session model.ExperimentSession
	err := r.DB(ctx).
		Preload("Experiment").
		Preload("Members").
		Preload("Members.User").
		Preload("Messages", func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at ASC") }).
		Preload("Messages.User").
		Where("session_key = ?", sessionKey).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExperimentSessionRepository) GetReusableByExperiment(ctx context.Context, experimentID uint) (*model.ExperimentSession, error) {
	var session model.ExperimentSession
	err := r.DB(ctx).
		Preload("Experiment").
		Preload("Members").
		Preload("Members.User").
		Preload("Messages", func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at ASC") }).
		Preload("Messages.User").
		Where("experiment_id = ? AND mode = ? AND status IN ?", experimentID, model.ExperimentModeCollaboration, model.ActiveExperimentEnvStatuses()).
		Order("created_at ASC").
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ExperimentSessionRepository) UpdateCounters(ctx context.Context, sessionID uint, memberCount int, primaryEnvID string, status string, startedAt *time.Time, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"current_member_count": memberCount,
	}
	if primaryEnvID != "" {
		updates["primary_env_id"] = primaryEnvID
	}
	if status != "" {
		updates["status"] = status
	}
	if startedAt != nil {
		updates["started_at"] = startedAt
	}
	if expiresAt != nil {
		updates["expires_at"] = expiresAt
	}
	return r.DB(ctx).Model(&model.ExperimentSession{}).Where("id = ?", sessionID).Updates(updates).Error
}

func (r *ExperimentSessionRepository) List(ctx context.Context, schoolID, experimentID uint, status string, page, pageSize int) ([]model.ExperimentSession, int64, error) {
	var sessions []model.ExperimentSession
	var total int64

	query := r.DB(ctx).Model(&model.ExperimentSession{})
	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if experimentID > 0 {
		query = query.Where("experiment_id = ?", experimentID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Scopes(Paginate(page, pageSize)).
		Preload("Experiment").
		Preload("Members").
		Preload("Members.User").
		Preload("Messages", func(tx *gorm.DB) *gorm.DB { return tx.Order("created_at DESC").Limit(50) }).
		Preload("Messages.User").
		Order("created_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

type ExperimentSessionMemberRepository struct {
	*BaseRepository
}

func NewExperimentSessionMemberRepository(db *gorm.DB) *ExperimentSessionMemberRepository {
	return &ExperimentSessionMemberRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ExperimentSessionMemberRepository) Create(ctx context.Context, member *model.ExperimentSessionMember) error {
	return r.DB(ctx).Create(member).Error
}

func (r *ExperimentSessionMemberRepository) GetBySessionAndUser(ctx context.Context, sessionID, userID uint) (*model.ExperimentSessionMember, error) {
	var member model.ExperimentSessionMember
	err := r.DB(ctx).Where("session_id = ? AND user_id = ?", sessionID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *ExperimentSessionMemberRepository) Update(ctx context.Context, member *model.ExperimentSessionMember) error {
	return r.DB(ctx).Save(member).Error
}

type ExperimentSessionMessageRepository struct {
	*BaseRepository
}

func NewExperimentSessionMessageRepository(db *gorm.DB) *ExperimentSessionMessageRepository {
	return &ExperimentSessionMessageRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ExperimentSessionMessageRepository) Create(ctx context.Context, message *model.ExperimentSessionMessage) error {
	return r.DB(ctx).Create(message).Error
}

func (r *ExperimentSessionMessageRepository) ListBySession(ctx context.Context, sessionID uint, page, pageSize int) ([]model.ExperimentSessionMessage, int64, error) {
	var messages []model.ExperimentSessionMessage
	var total int64
	query := r.DB(ctx).Model(&model.ExperimentSessionMessage{}).Where("session_id = ?", sessionID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Scopes(Paginate(page, pageSize)).
		Preload("User").
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

// NewSubmissionRepository 创建提交仓库
func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建提交
func (r *SubmissionRepository) Create(ctx context.Context, sub *model.Submission) error {
	return r.DB(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(sub).Error
}

// Update 更新提交
func (r *SubmissionRepository) Update(ctx context.Context, sub *model.Submission) error {
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Submission{}).
			Where("id = ?", sub.ID).
			Updates(map[string]interface{}{
				"experiment_id":  sub.ExperimentID,
				"student_id":     sub.StudentID,
				"school_id":      sub.SchoolID,
				"env_id":         sub.EnvID,
				"content":        sub.Content,
				"file_url":       sub.FileURL,
				"snapshot_url":   sub.SnapshotURL,
				"score":          sub.Score,
				"auto_score":     sub.AutoScore,
				"manual_score":   sub.ManualScore,
				"feedback":       sub.Feedback,
				"status":         sub.Status,
				"submitted_at":   sub.SubmittedAt,
				"graded_at":      sub.GradedAt,
				"grader_id":      sub.GraderID,
				"is_late":        sub.IsLate,
				"attempt_number": sub.AttemptNumber,
			}).Error; err != nil {
			return err
		}
		if err := tx.Where("submission_id = ?", sub.ID).Delete(&model.SubmissionCheckResult{}).Error; err != nil {
			return err
		}
		for index := range sub.CheckResults {
			sub.CheckResults[index].SubmissionID = sub.ID
		}
		if len(sub.CheckResults) > 0 {
			if err := tx.Create(&sub.CheckResults).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID 根据ID获取提交
func (r *SubmissionRepository) GetByID(ctx context.Context, id uint) (*model.Submission, error) {
	var sub model.Submission
	err := preloadSubmissionDetail(r.DB(ctx)).First(&sub, id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// GetLatestByStudent 获取学生最新提交
func (r *SubmissionRepository) GetLatestByStudent(ctx context.Context, experimentID, studentID uint) (*model.Submission, error) {
	var sub model.Submission
	err := r.DB(ctx).
		Where("experiment_id = ? AND student_id = ?", experimentID, studentID).
		Scopes(preloadSubmissionDetail).
		Order("submitted_at DESC").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// List 获取提交列表
func (r *SubmissionRepository) List(ctx context.Context, schoolID, experimentID, studentID uint, status string, page, pageSize int) ([]model.Submission, int64, error) {
	var subs []model.Submission
	var total int64

	query := r.DB(ctx).Model(&model.Submission{})

	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if experimentID > 0 {
		query = query.Where("experiment_id = ?", experimentID)
	}
	if studentID > 0 {
		query = query.Where("student_id = ?", studentID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Scopes(preloadSubmissionDetail).
		Order("submitted_at DESC").
		Find(&subs).Error
	if err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

// CountByExperiment 统计实验提交数
func (r *SubmissionRepository) CountByExperiment(ctx context.Context, experimentID uint) (int64, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Submission{}).Where("experiment_id = ?", experimentID).Count(&count).Error
	return count, err
}

// CountAttempts 统计学生提交次数
func (r *SubmissionRepository) CountAttempts(ctx context.Context, experimentID, studentID uint) (int, error) {
	var count int64
	err := r.DB(ctx).Model(&model.Submission{}).
		Where("experiment_id = ? AND student_id = ?", experimentID, studentID).
		Count(&count).Error
	return int(count), err
}

// DockerImageRepository Docker镜像仓库
type DockerImageRepository struct {
	*BaseRepository
}

// NewDockerImageRepository 创建Docker镜像仓库
func NewDockerImageRepository(db *gorm.DB) *DockerImageRepository {
	return &DockerImageRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建镜像
func (r *DockerImageRepository) Create(ctx context.Context, image *model.DockerImage) error {
	return r.DB(ctx).Create(image).Error
}

// Update 更新镜像
func (r *DockerImageRepository) Update(ctx context.Context, image *model.DockerImage) error {
	return r.DB(ctx).Save(image).Error
}

// Delete 删除镜像（软删除）
func (r *DockerImageRepository) Delete(ctx context.Context, id uint) error {
	return r.DB(ctx).Delete(&model.DockerImage{}, id).Error
}

// GetByID 根据ID获取镜像
func (r *DockerImageRepository) GetByID(ctx context.Context, id uint) (*model.DockerImage, error) {
	var image model.DockerImage
	err := r.DB(ctx).First(&image, id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// GetByName 根据名称获取镜像
func (r *DockerImageRepository) GetByName(ctx context.Context, name string) (*model.DockerImage, error) {
	var image model.DockerImage
	err := r.DB(ctx).Where("name = ?", name).First(&image).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// List 获取镜像列表
func (r *DockerImageRepository) List(ctx context.Context, category, status, keyword string, isBuiltIn *bool, page, pageSize int) ([]model.DockerImage, int64, error) {
	var images []model.DockerImage
	var total int64

	query := r.DB(ctx).Model(&model.DockerImage{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if isBuiltIn != nil {
		query = query.Where("is_built_in = ?", *isBuiltIn)
	}
	if keyword != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("category ASC, name ASC").
		Find(&images).Error
	if err != nil {
		return nil, 0, err
	}

	return images, total, nil
}

// ListAll 获取所有可用镜像
func (r *DockerImageRepository) ListAll(ctx context.Context) ([]model.DockerImage, error) {
	var images []model.DockerImage
	err := r.DB(ctx).Where("status = ?", model.StatusActive).Order("category ASC, name ASC").Find(&images).Error
	return images, err
}

func persistExperimentRelations(tx *gorm.DB, exp *model.Experiment) error {
	var workspaceIDs []uint
	if err := tx.Model(&model.ExperimentWorkspace{}).Where("experiment_id = ?", exp.ID).Pluck("id", &workspaceIDs).Error; err != nil {
		return err
	}
	if len(workspaceIDs) > 0 {
		if err := tx.Unscoped().Where("workspace_id IN ?", workspaceIDs).Delete(&model.ExperimentWorkspaceTool{}).Error; err != nil {
			return err
		}
	}

	var topologyIDs []uint
	if err := tx.Model(&model.ExperimentTopology{}).Where("experiment_id = ?", exp.ID).Pluck("id", &topologyIDs).Error; err != nil {
		return err
	}
	if len(topologyIDs) > 0 {
		if err := tx.Unscoped().Where("topology_id IN ?", topologyIDs).Delete(&model.ExperimentTopologyExposedEntry{}).Error; err != nil {
			return err
		}
	}

	var collaborationIDs []uint
	if err := tx.Model(&model.ExperimentCollaboration{}).Where("experiment_id = ?", exp.ID).Pluck("id", &collaborationIDs).Error; err != nil {
		return err
	}
	if len(collaborationIDs) > 0 {
		var roleIDs []uint
		if err := tx.Model(&model.ExperimentRoleBinding{}).Where("collaboration_id IN ?", collaborationIDs).Pluck("id", &roleIDs).Error; err != nil {
			return err
		}
		if len(roleIDs) > 0 {
			if err := tx.Unscoped().Where("role_binding_id IN ?", roleIDs).Delete(&model.ExperimentRoleBindingNode{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("role_binding_id IN ?", roleIDs).Delete(&model.ExperimentRoleBindingTool{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Unscoped().Where("collaboration_id IN ?", collaborationIDs).Delete(&model.ExperimentRoleBinding{}).Error; err != nil {
			return err
		}
	}

	var nodeIDs []uint
	if err := tx.Model(&model.ExperimentNode{}).Where("experiment_id = ?", exp.ID).Pluck("id", &nodeIDs).Error; err != nil {
		return err
	}
	if len(nodeIDs) > 0 {
		if err := tx.Unscoped().Where("node_id IN ?", nodeIDs).Delete(&model.ExperimentNodePort{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("node_id IN ?", nodeIDs).Delete(&model.ExperimentNodeTool{}).Error; err != nil {
			return err
		}
	}

	var serviceIDs []uint
	if err := tx.Model(&model.ExperimentService{}).Where("experiment_id = ?", exp.ID).Pluck("id", &serviceIDs).Error; err != nil {
		return err
	}
	if len(serviceIDs) > 0 {
		if err := tx.Unscoped().Where("service_id IN ?", serviceIDs).Delete(&model.ExperimentServicePort{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("service_id IN ?", serviceIDs).Delete(&model.ExperimentServiceEnvVar{}).Error; err != nil {
			return err
		}
	}

	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentWorkspace{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentTopology{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentTool{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentInitScript{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentCollaboration{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentNode{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentService{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentAsset{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("experiment_id = ?", exp.ID).Delete(&model.ExperimentCheckpoint{}).Error; err != nil {
		return err
	}

	if exp.Workspace != nil {
		exp.Workspace.ID = 0
		exp.Workspace.ExperimentID = exp.ID
		for index := range exp.Workspace.Tools {
			exp.Workspace.Tools[index].ID = 0
			exp.Workspace.Tools[index].WorkspaceID = 0
		}
		if err := tx.Create(exp.Workspace).Error; err != nil {
			return err
		}
	}

	if exp.Topology != nil {
		exp.Topology.ID = 0
		exp.Topology.ExperimentID = exp.ID
		for index := range exp.Topology.ExposedEntries {
			exp.Topology.ExposedEntries[index].ID = 0
			exp.Topology.ExposedEntries[index].TopologyID = 0
		}
		if err := tx.Create(exp.Topology).Error; err != nil {
			return err
		}
	}

	for index := range exp.Tools {
		exp.Tools[index].ID = 0
		exp.Tools[index].ExperimentID = exp.ID
	}
	if len(exp.Tools) > 0 {
		if err := tx.Create(&exp.Tools).Error; err != nil {
			return err
		}
	}

	for index := range exp.InitScripts {
		exp.InitScripts[index].ID = 0
		exp.InitScripts[index].ExperimentID = exp.ID
	}
	if len(exp.InitScripts) > 0 {
		if err := tx.Create(&exp.InitScripts).Error; err != nil {
			return err
		}
	}

	if exp.Collaboration != nil {
		exp.Collaboration.ID = 0
		exp.Collaboration.ExperimentID = exp.ID
		for index := range exp.Collaboration.Roles {
			exp.Collaboration.Roles[index].ID = 0
			exp.Collaboration.Roles[index].CollaborationID = 0
			for nodeIndex := range exp.Collaboration.Roles[index].NodeAssignments {
				exp.Collaboration.Roles[index].NodeAssignments[nodeIndex].ID = 0
				exp.Collaboration.Roles[index].NodeAssignments[nodeIndex].RoleBindingID = 0
			}
			for toolIndex := range exp.Collaboration.Roles[index].ToolAssignments {
				exp.Collaboration.Roles[index].ToolAssignments[toolIndex].ID = 0
				exp.Collaboration.Roles[index].ToolAssignments[toolIndex].RoleBindingID = 0
			}
		}
		if err := tx.Create(exp.Collaboration).Error; err != nil {
			return err
		}
	}

	for index := range exp.Nodes {
		exp.Nodes[index].ID = 0
		exp.Nodes[index].ExperimentID = exp.ID
		for portIndex := range exp.Nodes[index].Ports {
			exp.Nodes[index].Ports[portIndex].ID = 0
			exp.Nodes[index].Ports[portIndex].NodeID = 0
		}
		for toolIndex := range exp.Nodes[index].Tools {
			exp.Nodes[index].Tools[toolIndex].ID = 0
			exp.Nodes[index].Tools[toolIndex].NodeID = 0
		}
		if err := tx.Create(&exp.Nodes[index]).Error; err != nil {
			return err
		}
	}

	for index := range exp.Services {
		exp.Services[index].ID = 0
		exp.Services[index].ExperimentID = exp.ID
		for portIndex := range exp.Services[index].Ports {
			exp.Services[index].Ports[portIndex].ID = 0
			exp.Services[index].Ports[portIndex].ServiceID = 0
		}
		for envVarIndex := range exp.Services[index].EnvVars {
			exp.Services[index].EnvVars[envVarIndex].ID = 0
			exp.Services[index].EnvVars[envVarIndex].ServiceID = 0
		}
		if err := tx.Create(&exp.Services[index]).Error; err != nil {
			return err
		}
	}

	for index := range exp.Assets {
		exp.Assets[index].ID = 0
		exp.Assets[index].ExperimentID = exp.ID
	}
	if len(exp.Assets) > 0 {
		if err := tx.Create(&exp.Assets).Error; err != nil {
			return err
		}
	}

	for index := range exp.Checkpoints {
		exp.Checkpoints[index].ID = 0
		exp.Checkpoints[index].ExperimentID = exp.ID
	}
	if len(exp.Checkpoints) > 0 {
		if err := tx.Create(&exp.Checkpoints).Error; err != nil {
			return err
		}
	}
	return nil
}

func persistExperimentRuntimeRelations(tx *gorm.DB, env *model.ExperimentEnv) error {
	var runtimeIDs []uint
	if err := tx.Model(&model.ExperimentRuntimeInstance{}).Where("experiment_env_id = ?", env.ID).Pluck("id", &runtimeIDs).Error; err != nil {
		return err
	}
	if len(runtimeIDs) > 0 {
		if err := tx.Unscoped().Where("runtime_instance_id IN ?", runtimeIDs).Delete(&model.ExperimentRuntimeTool{}).Error; err != nil {
			return err
		}
	}
	if err := tx.Unscoped().Where("experiment_env_id = ?", env.ID).Delete(&model.ExperimentRuntimeInstance{}).Error; err != nil {
		return err
	}
	for index := range env.RuntimeInstances {
		env.RuntimeInstances[index].ID = 0
		env.RuntimeInstances[index].ExperimentEnvID = env.ID
		if env.RuntimeInstances[index].Status == "" {
			env.RuntimeInstances[index].Status = model.EnvStatusPending
		}
		for toolIndex := range env.RuntimeInstances[index].Tools {
			env.RuntimeInstances[index].Tools[toolIndex].ID = 0
			env.RuntimeInstances[index].Tools[toolIndex].RuntimeInstanceID = 0
		}
		if err := tx.Create(&env.RuntimeInstances[index]).Error; err != nil {
			return err
		}
	}
	return nil
}
