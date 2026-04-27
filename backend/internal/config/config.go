package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	MinIO      MinIOConfig      `mapstructure:"minio"`
	RabbitMQ   RabbitMQConfig   `mapstructure:"rabbitmq"`
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
	Log        LogConfig        `mapstructure:"log"`
	CORS       CORSConfig       `mapstructure:"cors"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Experiment ExperimentConfig `mapstructure:"experiment"`
	Contest    ContestConfig    `mapstructure:"contest"`
	Upload     UploadConfig     `mapstructure:"upload"`
	Init       InitConfig       `mapstructure:"init"`
	Etherscan  EtherscanConfig  `mapstructure:"etherscan"`
	Proxy      ProxyConfig      `mapstructure:"proxy"`
}

// EtherscanConfig Etherscan API配置
type EtherscanConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// ProxyConfig 外部网络代理配置（用于访问 SlowMist、BlockSec、Etherscan、Sourcify 等外部站点）
type ProxyConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	HTTP    string `mapstructure:"http"`     // http://host:port，留空则不启用
	HTTPS   string `mapstructure:"https"`    // https://host:port，留空则不启用
	NoProxy string `mapstructure:"no_proxy"` // 跳过代理的地址，用逗号分隔
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Name            string        `mapstructure:"name"`
	SSLMode         string        `mapstructure:"sslmode"`
	Timezone        string        `mapstructure:"timezone"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level"`
}

// DSN 返回数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode, c.Timezone,
	)
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// Addr 返回Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpire  time.Duration `mapstructure:"access_expire"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
	Issuer        string        `mapstructure:"issuer"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
	PublicURL string `mapstructure:"public_url"`
}

// RabbitMQConfig RabbitMQ配置
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

// URL 返回RabbitMQ连接URL
func (c *RabbitMQConfig) URL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.VHost)
}

// KubernetesConfig Kubernetes配置
type KubernetesConfig struct {
	Enabled          bool                     `mapstructure:"enabled"`
	InCluster        bool                     `mapstructure:"in_cluster"`
	Kubeconfig       string                   `mapstructure:"kubeconfig"`
	Namespace        string                   `mapstructure:"namespace"`
	ImageRegistry    string                   `mapstructure:"image_registry"`
	ImagePullSecret  string                   `mapstructure:"image_pull_secret"`
	DefaultTimeout   time.Duration            `mapstructure:"default_timeout"`
	MaxTimeout       time.Duration            `mapstructure:"max_timeout"`
	ResourceDefaults KubernetesResourceConfig `mapstructure:"resource_defaults"`
}

// KubernetesResourceConfig K8s资源默认配置
type KubernetesResourceConfig struct {
	CPURequest    string `mapstructure:"cpu_request"`
	CPULimit      string `mapstructure:"cpu_limit"`
	MemoryRequest string `mapstructure:"memory_request"`
	MemoryLimit   string `mapstructure:"memory_limit"`
	Storage       string `mapstructure:"storage"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowedOrigins   []string      `mapstructure:"allowed_origins"`
	AllowedMethods   []string      `mapstructure:"allowed_methods"`
	AllowedHeaders   []string      `mapstructure:"allowed_headers"`
	ExposedHeaders   []string      `mapstructure:"exposed_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

// ExperimentConfig 实验配置
type ExperimentConfig struct {
	DefaultTimeout        time.Duration `mapstructure:"default_timeout"`
	MaxExtendTimes        int           `mapstructure:"max_extend_times"`
	ExtendDuration        time.Duration `mapstructure:"extend_duration"`
	SnapshotBeforeTimeout time.Duration `mapstructure:"snapshot_before_timeout"`
	CleanupInterval       time.Duration `mapstructure:"cleanup_interval"`
	MaxUserEnvs           int           `mapstructure:"max_user_envs"`   // 每用户最大并发环境数
	MaxSchoolEnvs         int           `mapstructure:"max_school_envs"` // 每学校最大并发环境数
}

// ContestConfig CTF竞赛配置
type ContestConfig struct {
	FlagPrefix             string        `mapstructure:"flag_prefix"`
	FlagSuffix             string        `mapstructure:"flag_suffix"`
	DynamicFlagSalt        string        `mapstructure:"dynamic_flag_salt"`
	MaxTeamSize            int           `mapstructure:"max_team_size"`
	ScoreboardFreezeBefore time.Duration `mapstructure:"scoreboard_freeze_before"`
}

// UploadConfig 文件上传配置
type UploadConfig struct {
	MaxSize      int      `mapstructure:"max_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
	TempDir      string   `mapstructure:"temp_dir"`
}

// InitConfig 平台初始化配置
type InitConfig struct {
	AdminPhone    string `mapstructure:"admin_phone"`
	AdminName     string `mapstructure:"admin_name"`
	AdminPassword string `mapstructure:"admin_password"`
	AdminEmail    string `mapstructure:"admin_email"`
}

var globalConfig *Config

// Load 加载配置、应用默认值并执行基础校验。
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("../../configs")
	}

	// 设置环境变量
	v.SetEnvPrefix("CHAINSPACE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		// 配置文件不存在时使用默认值和环境变量
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// setDefaults 设置开发与本地运行所需的基础默认值。
func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.port", 3000)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", "180s")
	v.SetDefault("server.write_timeout", "180s")
	v.SetDefault("server.shutdown_timeout", "10s")

	// Database
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "chainspace")
	v.SetDefault("database.password", "123456")
	v.SetDefault("database.name", "chainspace")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.timezone", "Asia/Shanghai")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", "1h")
	v.SetDefault("database.log_level", "silent")

	// Redis
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.password", "chainspace123")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 100)
	v.SetDefault("redis.min_idle_conns", 10)

	// JWT
	v.SetDefault("jwt.secret", "f8b7a9d3e5c8b1f9e7d2c4a8b9f7e6d5c3b8a9e7f8d6c5b4a7f9e8d7c6b5a8f9")
	v.SetDefault("jwt.access_expire", "15m")
	v.SetDefault("jwt.refresh_expire", "168h")
	v.SetDefault("jwt.issuer", "chainspace")

	// MinIO
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "chainspace")
	v.SetDefault("minio.secret_key", "chainspace123")
	v.SetDefault("minio.bucket", "chainspace")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.region", "cn-north-1")
	v.SetDefault("minio.public_url", "http://localhost:9000")

	// RabbitMQ
	v.SetDefault("rabbitmq.host", "localhost")
	v.SetDefault("rabbitmq.port", 5672)
	v.SetDefault("rabbitmq.user", "chainspace")
	v.SetDefault("rabbitmq.password", "chainspace123")
	v.SetDefault("rabbitmq.vhost", "/")

	// Kubernetes
	v.SetDefault("kubernetes.enabled", true)
	v.SetDefault("kubernetes.in_cluster", true)
	v.SetDefault("kubernetes.namespace", "chainspace-labs")
	v.SetDefault("kubernetes.default_timeout", "4h")
	v.SetDefault("kubernetes.max_timeout", "24h")
	v.SetDefault("kubernetes.resource_defaults.cpu_request", "0.5")
	v.SetDefault("kubernetes.resource_defaults.cpu_limit", "2")
	v.SetDefault("kubernetes.resource_defaults.memory_request", "512Mi")
	v.SetDefault("kubernetes.resource_defaults.memory_limit", "4Gi")
	v.SetDefault("kubernetes.resource_defaults.storage", "10Gi")

	// Log
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "file")
	v.SetDefault("log.file_path", "./logs/app.log")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)

	// CORS
	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"})
	v.SetDefault("cors.exposed_headers", []string{"Content-Length", "Content-Type"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", "12h")

	// Rate Limit
	v.SetDefault("rate_limit.enabled", true)
	v.SetDefault("rate_limit.requests_per_second", 100)
	v.SetDefault("rate_limit.burst", 200)

	// Experiment
	v.SetDefault("experiment.default_timeout", "4h")
	v.SetDefault("experiment.max_extend_times", 3)
	v.SetDefault("experiment.extend_duration", "2h")
	v.SetDefault("experiment.snapshot_before_timeout", "10m")
	v.SetDefault("experiment.cleanup_interval", "5m")
	v.SetDefault("experiment.max_user_envs", 3)
	v.SetDefault("experiment.max_school_envs", 100)

	// Contest
	v.SetDefault("contest.flag_prefix", "flag{")
	v.SetDefault("contest.flag_suffix", "}")
	v.SetDefault("contest.dynamic_flag_salt", "chainspace-flag-salt")
	v.SetDefault("contest.max_team_size", 5)
	v.SetDefault("contest.scoreboard_freeze_before", "1h")

	// Upload
	v.SetDefault("upload.max_size", 100)
	v.SetDefault("upload.allowed_types", []string{
		"image/jpeg", "image/png", "image/gif",
		"application/pdf", "application/zip",
		"text/plain", "text/markdown",
	})
	v.SetDefault("upload.temp_dir", "./tmp/uploads")

	// Init
	v.SetDefault("init.admin_phone", "13800000001")
	v.SetDefault("init.admin_name", "平台管理员")
	v.SetDefault("init.admin_password", "Admin@123456")
	v.SetDefault("init.admin_email", "admin@chainspace.com")

	// Proxy
	v.SetDefault("proxy.enabled", false)
	v.SetDefault("proxy.http", "")
	v.SetDefault("proxy.https", "")
	v.SetDefault("proxy.no_proxy", "")
}

// validate 对基础层依赖的关键配置进行统一校验，避免运行时才暴露问题。
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("server.port must be greater than 0")
	}
	if cfg.Server.ReadTimeout <= 0 || cfg.Server.WriteTimeout <= 0 || cfg.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("server timeouts must be greater than 0")
	}
	if cfg.Server.Mode != "debug" && cfg.Server.Mode != "release" {
		return fmt.Errorf("server.mode must be debug or release")
	}

	if cfg.Database.Host == "" || cfg.Database.User == "" || cfg.Database.Name == "" {
		return fmt.Errorf("database host, user and name are required")
	}
	if cfg.Database.Port <= 0 {
		return fmt.Errorf("database.port must be greater than 0")
	}

	if cfg.Redis.Host == "" || cfg.Redis.Port <= 0 {
		return fmt.Errorf("redis host and port are required")
	}

	if cfg.JWT.Secret == "" || len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("jwt.secret must be at least 32 characters")
	}
	if cfg.JWT.AccessExpire <= 0 || cfg.JWT.RefreshExpire <= 0 {
		return fmt.Errorf("jwt token expiration must be greater than 0")
	}

	if cfg.RateLimit.Enabled && (cfg.RateLimit.RequestsPerSecond <= 0 || cfg.RateLimit.Burst <= 0) {
		return fmt.Errorf("rate_limit requests_per_second and burst must be greater than 0 when enabled")
	}

	if cfg.Upload.MaxSize <= 0 {
		return fmt.Errorf("upload.max_size must be greater than 0")
	}
	if cfg.Experiment.MaxUserEnvs <= 0 || cfg.Experiment.MaxSchoolEnvs <= 0 {
		return fmt.Errorf("experiment max envs must be greater than 0")
	}
	if cfg.Contest.MaxTeamSize <= 0 {
		return fmt.Errorf("contest.max_team_size must be greater than 0")
	}
	if cfg.Kubernetes.Enabled && cfg.Kubernetes.Namespace == "" {
		return fmt.Errorf("kubernetes.namespace is required when kubernetes is enabled")
	}

	return nil
}
