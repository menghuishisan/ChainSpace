package model

// SystemConfig 系统配置
type SystemConfig struct {
	BaseModel
	Key         string `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Type        string `gorm:"size:20;default:string" json:"type"`
	Description string `gorm:"type:text" json:"description"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`
	Group       string `gorm:"size:50;default:general" json:"group"`
}

// TableName 表名
func (SystemConfig) TableName() string {
	return "system_configs"
}

// 配置类型常量
const (
	ConfigTypeString = "string"
	ConfigTypeInt    = "int"
	ConfigTypeBool   = "bool"
	ConfigTypeJSON   = "json"
)

// 配置组常量
const (
	ConfigGroupGeneral    = "general"
	ConfigGroupSecurity   = "security"
	ConfigGroupExperiment = "experiment"
	ConfigGroupContest    = "contest"
	ConfigGroupUpload     = "upload"
)

// 常用配置Key
const (
	ConfigKeyPlatformName       = "platform_name"
	ConfigKeyPlatformLogo       = "platform_logo"
	ConfigKeyAllowRegistration  = "allow_registration"
	ConfigKeyDefaultExpTimeout  = "default_experiment_timeout"
	ConfigKeyMaxExpExtendTimes = "max_experiment_extend_times"
	ConfigKeyUploadMaxSize      = "upload_max_size"
	ConfigKeyUploadAllowedTypes = "upload_allowed_types"
)

// OperationLog 操作日志
type OperationLog struct {
	BaseModel
	UserID       uint    `gorm:"index;not null" json:"user_id"`
	SchoolID     *uint   `gorm:"index" json:"school_id"`
	Module       string  `gorm:"size:50;not null" json:"module"`
	Action       string  `gorm:"size:50;not null" json:"action"`
	TargetType   string  `gorm:"size:50" json:"target_type"`
	TargetID     *uint   `json:"target_id"`
	Description  string  `gorm:"type:text" json:"description"`
	RequestIP    string  `gorm:"size:50" json:"request_ip"`
	UserAgent    string  `gorm:"size:500" json:"user_agent"`
	RequestData  JSONMap `gorm:"type:jsonb" json:"request_data"`
	ResponseCode int     `gorm:"default:0" json:"response_code"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 表名
func (OperationLog) TableName() string {
	return "operation_logs"
}
