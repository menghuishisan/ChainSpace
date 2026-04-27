package repository

import (
	"context"

	"github.com/chainspace/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SystemConfigRepository 系统配置仓库
type SystemConfigRepository struct {
	*BaseRepository
}

// NewSystemConfigRepository 创建系统配置仓库
func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Get 获取配置
func (r *SystemConfigRepository) Get(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.DB(ctx).Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Set 设置配置（UPSERT: key 存在则只更新 value，不存在则创建全部字段）
func (r *SystemConfigRepository) Set(ctx context.Context, key, value, configType, description, group string, isPublic bool) error {
	config := model.SystemConfig{
		Key:         key,
		Value:       value,
		Type:        configType,
		Description: description,
		Group:       group,
		IsPublic:    isPublic,
	}

	updateCols := []string{"value", "updated_at"}
	if description != "" {
		updateCols = append(updateCols, "description")
	}
	if group != "" {
		updateCols = append(updateCols, "group")
	}

	return r.DB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns(updateCols),
	}).Create(&config).Error
}

// List 获取配置列表
func (r *SystemConfigRepository) List(ctx context.Context, group string, publicOnly bool) ([]model.SystemConfig, error) {
	var configs []model.SystemConfig

	query := r.DB(ctx).Model(&model.SystemConfig{})
	if group != "" {
		query = query.Where("\"group\" = ?", group)
	}
	if publicOnly {
		query = query.Where("is_public = ?", true)
	}

	err := query.Order("\"group\" ASC, key ASC").Find(&configs).Error
	return configs, err
}

// Delete 删除配置
func (r *SystemConfigRepository) Delete(ctx context.Context, key string) error {
	return r.DB(ctx).Where("key = ?", key).Delete(&model.SystemConfig{}).Error
}

// VulnerabilitySourceRepository 漏洞来源仓库
type VulnerabilitySourceRepository struct {
	*BaseRepository
}

// NewVulnerabilitySourceRepository 创建漏洞来源仓库
func NewVulnerabilitySourceRepository(db *gorm.DB) *VulnerabilitySourceRepository {
	return &VulnerabilitySourceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建漏洞来源
func (r *VulnerabilitySourceRepository) Create(ctx context.Context, vuln *model.VulnerabilitySource) error {
	return r.DB(ctx).Create(vuln).Error
}

// Update 更新漏洞来源
func (r *VulnerabilitySourceRepository) Update(ctx context.Context, vuln *model.VulnerabilitySource) error {
	return r.DB(ctx).Save(vuln).Error
}

// GetByID 根据ID获取漏洞来源
func (r *VulnerabilitySourceRepository) GetByID(ctx context.Context, id uint) (*model.VulnerabilitySource, error) {
	var vuln model.VulnerabilitySource
	err := r.DB(ctx).First(&vuln, id).Error
	if err != nil {
		return nil, err
	}
	return &vuln, nil
}

// ListActive 获取所有活跃的漏洞来源
func (r *VulnerabilitySourceRepository) ListActive(ctx context.Context) ([]model.VulnerabilitySource, error) {
	var sources []model.VulnerabilitySource
	err := r.DB(ctx).Where("is_active = ?", true).Find(&sources).Error
	return sources, err
}

// GetByType 根据来源类型获取漏洞来源。
func (r *VulnerabilitySourceRepository) GetByType(ctx context.Context, sourceType string) (*model.VulnerabilitySource, error) {
	var source model.VulnerabilitySource
	err := r.DB(ctx).Where("type = ?", sourceType).First(&source).Error
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// CrossSchoolApplicationRepository 跨校申请仓库
type CrossSchoolApplicationRepository struct {
	*BaseRepository
}

// NewCrossSchoolApplicationRepository 创建跨校申请仓库
func NewCrossSchoolApplicationRepository(db *gorm.DB) *CrossSchoolApplicationRepository {
	return &CrossSchoolApplicationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建申请
func (r *CrossSchoolApplicationRepository) Create(ctx context.Context, app *model.CrossSchoolApplication) error {
	return r.DB(ctx).Create(app).Error
}

// Update 更新申请
func (r *CrossSchoolApplicationRepository) Update(ctx context.Context, app *model.CrossSchoolApplication) error {
	return r.DB(ctx).Save(app).Error
}

// GetByID 根据ID获取申请
func (r *CrossSchoolApplicationRepository) GetByID(ctx context.Context, id uint) (*model.CrossSchoolApplication, error) {
	var app model.CrossSchoolApplication
	err := r.DB(ctx).
		Preload("FromSchool").Preload("ToSchool").
		Preload("Applicant").Preload("Reviewer").
		First(&app, id).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// List 获取申请列表
func (r *CrossSchoolApplicationRepository) List(ctx context.Context, fromSchoolID, toSchoolID uint, appType, status string, page, pageSize int) ([]model.CrossSchoolApplication, int64, error) {
	var apps []model.CrossSchoolApplication
	var total int64

	query := r.DB(ctx).Model(&model.CrossSchoolApplication{})

	if fromSchoolID > 0 {
		query = query.Where("from_school_id = ?", fromSchoolID)
	}
	if toSchoolID > 0 {
		query = query.Where("to_school_id = ?", toSchoolID)
	}
	if appType != "" {
		query = query.Where("type = ?", appType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("FromSchool").Preload("ToSchool").
		Preload("Applicant").
		Order("created_at DESC").
		Find(&apps).Error
	if err != nil {
		return nil, 0, err
	}

	return apps, total, nil
}

// OperationLogRepository 操作日志仓库
type OperationLogRepository struct {
	*BaseRepository
}

// NewOperationLogRepository 创建操作日志仓库
func NewOperationLogRepository(db *gorm.DB) *OperationLogRepository {
	return &OperationLogRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建日志
func (r *OperationLogRepository) Create(ctx context.Context, log *model.OperationLog) error {
	return r.DB(ctx).Create(log).Error
}

// List 获取日志列表
func (r *OperationLogRepository) List(ctx context.Context, schoolID, userID uint, module, action string, page, pageSize int) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := r.DB(ctx).Model(&model.OperationLog{})

	if schoolID > 0 {
		query = query.Where("school_id = ?", schoolID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Preload("User").
		Order("created_at DESC").
		Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteOld 删除旧日志
func (r *OperationLogRepository) DeleteOld(ctx context.Context, days int) error {
	return r.DB(ctx).Where("created_at < NOW() - INTERVAL '? days'", days).Delete(&model.OperationLog{}).Error
}

// VulnerabilityRepository 漏洞仓库
type VulnerabilityRepository struct {
	*BaseRepository
}

// NewVulnerabilityRepository 创建漏洞仓库
func NewVulnerabilityRepository(db *gorm.DB) *VulnerabilityRepository {
	return &VulnerabilityRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建漏洞
func (r *VulnerabilityRepository) Create(ctx context.Context, vuln *model.Vulnerability) error {
	return r.DB(ctx).Create(vuln).Error
}

// Update 更新漏洞
func (r *VulnerabilityRepository) Update(ctx context.Context, vuln *model.Vulnerability) error {
	return r.DB(ctx).Save(vuln).Error
}

// UpdateFields 按字段更新漏洞（不覆盖未指定字段）
func (r *VulnerabilityRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.DB(ctx).Model(&model.Vulnerability{}).Where("id = ?", id).Updates(fields).Error
}

// GetByID 根据ID获取漏洞
func (r *VulnerabilityRepository) GetByID(ctx context.Context, id uint) (*model.Vulnerability, error) {
	var vuln model.Vulnerability
	err := r.DB(ctx).Preload("Source").Preload("Converted").First(&vuln, id).Error
	if err != nil {
		return nil, err
	}
	return &vuln, nil
}

// GetByExternalID 根据外部ID获取漏洞
func (r *VulnerabilityRepository) GetByExternalID(ctx context.Context, sourceID uint, externalID string) (*model.Vulnerability, error) {
	var vuln model.Vulnerability
	err := r.DB(ctx).Where("source_id = ? AND external_id = ?", sourceID, externalID).First(&vuln).Error
	if err != nil {
		return nil, err
	}
	return &vuln, nil
}

// List 获取漏洞列表。
func (r *VulnerabilityRepository) List(ctx context.Context, keyword, status, category, severity, chain string, page, pageSize int) ([]model.Vulnerability, int64, error) {
	var vulns []model.Vulnerability
	var total int64

	query := r.DB(ctx).Model(&model.Vulnerability{})

	if keyword != "" {
		pattern := "%" + keyword + "%"
		query = query.Where(
			"title ILIKE ? OR description ILIKE ? OR technique ILIKE ? OR contract_address ILIKE ? OR attack_tx_hash ILIKE ?",
			pattern,
			pattern,
			pattern,
			pattern,
			pattern,
		)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if chain != "" {
		query = query.Where("chain ILIKE ?", "%"+chain+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Scopes(Paginate(page, pageSize)).
		Order("conversion_score DESC, attack_date DESC, created_at DESC").
		Find(&vulns).Error
	if err != nil {
		return nil, 0, err
	}

	return vulns, total, nil
}

// ListUnconverted 获取未转化的漏洞列表
func (r *VulnerabilityRepository) ListUnconverted(ctx context.Context, limit int) ([]model.Vulnerability, error) {
	var vulns []model.Vulnerability
	err := r.DB(ctx).
		Where("status IN ? AND converted_id IS NULL", []string{model.VulnStatusDiscovered, model.VulnStatusEnriched}).
		Limit(limit).
		Order("conversion_score DESC, amount DESC").
		Find(&vulns).Error
	return vulns, err
}
