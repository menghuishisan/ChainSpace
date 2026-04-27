package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/handler"
	"github.com/chainspace/backend/internal/pkg/cache"
	"github.com/chainspace/backend/internal/pkg/database"
	"github.com/chainspace/backend/internal/pkg/jwt"
	"github.com/chainspace/backend/internal/pkg/k8s"
	"github.com/chainspace/backend/internal/pkg/logger"
	"github.com/chainspace/backend/internal/pkg/mq"
	"github.com/chainspace/backend/internal/pkg/task"
	"github.com/chainspace/backend/internal/pkg/websocket"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/internal/router"
	"github.com/chainspace/backend/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Runtime 封装基础设施装配完成后的应用实例。
type Runtime struct {
	config           *config.Config
	db               *gorm.DB
	redisClient      *redis.Client
	server           *http.Server
	schedulerService *service.SchedulerService
	taskManager      *task.Manager
	mqClient         *mq.Client
	authService      *service.AuthService
	wsHub            *websocket.Hub
}

// Build 统一完成入口装配。
func Build(cfg *config.Config) (*Runtime, error) {
	db, err := database.Init(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	redisClient, err := cache.Init(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("init redis: %w", err)
	}

	jwtManager := jwt.NewManager(&cfg.JWT, redisClient)
	k8sClient := newK8sClient(cfg)
	repos := newRepositories(db)

	services, err := newServices(cfg, db, redisClient, jwtManager, k8sClient, repos)
	if err != nil {
		return nil, err
	}

	handlers := newHandlers(services)
	engine := router.NewRouter(
		cfg,
		jwtManager,
		repos.operationLogRepo,
		handlers.authHandler,
		handlers.userHandler,
		handlers.schoolHandler,
		handlers.courseHandler,
		handlers.experimentHandler,
		handlers.contestHandler,
		handlers.notificationHandler,
		handlers.discussionHandler,
		handlers.systemHandler,
		handlers.uploadHandler,
		handlers.agentBattleHandler,
		handlers.websocketHandler,
		handlers.imageHandler,
		handlers.challengeHandler,
		handlers.envProxyHandler,
		handlers.contestEnvProxyHandler,
		handlers.battleWorkspaceProxyHandler,
	).Setup()

	return &Runtime{
		config:      cfg,
		db:          db,
		redisClient: redisClient,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      engine,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
		schedulerService: services.schedulerService,
		taskManager:      services.taskManager,
		mqClient:         services.mqClient,
		authService:      services.authService,
		wsHub:            services.wsHub,
	}, nil
}

// Start 启动异步组件与 HTTP 服务。
func (a *Runtime) Start() {
	go a.wsHub.Run()
	a.taskManager.Start()
	a.schedulerService.Start()

	go func() {
		logger.Info("Server starting", zap.Int("port", a.config.Server.Port))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()
}

// Shutdown 按依赖顺序关闭应用。
func (a *Runtime) Shutdown(ctx context.Context) error {
	a.schedulerService.Stop()
	a.taskManager.Stop()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	if a.mqClient != nil {
		if err := a.mqClient.Close(); err != nil {
			logger.Warn("Close RabbitMQ failed", zap.Error(err))
		}
	}

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			logger.Warn("Close redis failed", zap.Error(err))
		}
	}

	if err := database.Close(); err != nil {
		logger.Warn("Close database failed", zap.Error(err))
	}

	return nil
}

// InitPlatformAdmin 根据初始化配置补齐平台管理员账号。
func (a *Runtime) InitPlatformAdmin(ctx context.Context) {
	if a.config.Init.AdminPhone == "" || a.config.Init.AdminPassword == "" {
		return
	}

	if err := a.authService.InitPlatformAdmin(ctx, a.config.Init.AdminPhone, a.config.Init.AdminName, a.config.Init.AdminPassword, a.config.Init.AdminEmail); err != nil {
		logger.Warn("Init platform admin", zap.Error(err))
		return
	}

	logger.Info("Platform admin initialized")
}

// newK8sClient 根据配置创建可选的 Kubernetes 客户端。
func newK8sClient(cfg *config.Config) *k8s.Client {
	if !cfg.Kubernetes.Enabled {
		logger.Info("Kubernetes disabled, experiment environments will use local mode")
		return nil
	}

	k8sClient, err := k8s.NewClient(&cfg.Kubernetes)
	if err != nil {
		logger.Warn("Failed to init k8s client, running without k8s support", zap.Error(err))
		return nil
	}

	logger.Info("Kubernetes client initialized", zap.String("namespace", cfg.Kubernetes.Namespace))
	return k8sClient
}

// repositories 聚合入口装配阶段所需的仓储依赖。
type repositories struct {
	userRepo              *repository.UserRepository
	schoolRepo            *repository.SchoolRepository
	classRepo             *repository.ClassRepository
	courseRepo            *repository.CourseRepository
	courseStudentRepo     *repository.CourseStudentRepository
	chapterRepo           *repository.ChapterRepository
	materialRepo          *repository.MaterialRepository
	materialProgressRepo  *repository.MaterialProgressRepository
	experimentRepo        *repository.ExperimentRepository
	sessionRepo           *repository.ExperimentSessionRepository
	sessionMemberRepo     *repository.ExperimentSessionMemberRepository
	sessionMessageRepo    *repository.ExperimentSessionMessageRepository
	envRepo               *repository.ExperimentEnvRepository
	submissionRepo        *repository.SubmissionRepository
	dockerImageRepo       *repository.DockerImageRepository
	contestRepo           *repository.ContestRepository
	challengeRepo         *repository.ChallengeRepository
	contestChallengeRepo  *repository.ContestChallengeRepository
	teamRepo              *repository.TeamRepository
	teamMemberRepo        *repository.TeamMemberRepository
	contestSubmissionRepo *repository.ContestSubmissionRepository
	contestScoreRepo      *repository.ContestScoreRepository
	challengeEnvRepo      *repository.ChallengeEnvRepository
	notificationRepo      *repository.NotificationRepository
	postRepo              *repository.PostRepository
	replyRepo             *repository.ReplyRepository
	postLikeRepo          *repository.PostLikeRepository
	replyLikeRepo         *repository.ReplyLikeRepository
	configRepo            *repository.SystemConfigRepository
	vulnSourceRepo        *repository.VulnerabilitySourceRepository
	vulnRepo              *repository.VulnerabilityRepository
	crossSchoolRepo       *repository.CrossSchoolApplicationRepository
	operationLogRepo      *repository.OperationLogRepository
	roundRepo             *repository.AgentBattleRoundRepository
	contractRepo          *repository.AgentContractRepository
	eventRepo             *repository.AgentBattleEventRepository
	battleScoreRepo       *repository.AgentBattleScoreRepository
}

// newRepositories 统一创建第一层需要的仓储实例。
func newRepositories(db *gorm.DB) *repositories {
	return &repositories{
		userRepo:              repository.NewUserRepository(db),
		schoolRepo:            repository.NewSchoolRepository(db),
		classRepo:             repository.NewClassRepository(db),
		courseRepo:            repository.NewCourseRepository(db),
		courseStudentRepo:     repository.NewCourseStudentRepository(db),
		chapterRepo:           repository.NewChapterRepository(db),
		materialRepo:          repository.NewMaterialRepository(db),
		materialProgressRepo:  repository.NewMaterialProgressRepository(db),
		experimentRepo:        repository.NewExperimentRepository(db),
		sessionRepo:           repository.NewExperimentSessionRepository(db),
		sessionMemberRepo:     repository.NewExperimentSessionMemberRepository(db),
		sessionMessageRepo:    repository.NewExperimentSessionMessageRepository(db),
		envRepo:               repository.NewExperimentEnvRepository(db),
		submissionRepo:        repository.NewSubmissionRepository(db),
		dockerImageRepo:       repository.NewDockerImageRepository(db),
		contestRepo:           repository.NewContestRepository(db),
		challengeRepo:         repository.NewChallengeRepository(db),
		contestChallengeRepo:  repository.NewContestChallengeRepository(db),
		teamRepo:              repository.NewTeamRepository(db),
		teamMemberRepo:        repository.NewTeamMemberRepository(db),
		contestSubmissionRepo: repository.NewContestSubmissionRepository(db),
		contestScoreRepo:      repository.NewContestScoreRepository(db),
		challengeEnvRepo:      repository.NewChallengeEnvRepository(db),
		notificationRepo:      repository.NewNotificationRepository(db),
		postRepo:              repository.NewPostRepository(db),
		replyRepo:             repository.NewReplyRepository(db),
		postLikeRepo:          repository.NewPostLikeRepository(db),
		replyLikeRepo:         repository.NewReplyLikeRepository(db),
		configRepo:            repository.NewSystemConfigRepository(db),
		vulnSourceRepo:        repository.NewVulnerabilitySourceRepository(db),
		vulnRepo:              repository.NewVulnerabilityRepository(db),
		crossSchoolRepo:       repository.NewCrossSchoolApplicationRepository(db),
		operationLogRepo:      repository.NewOperationLogRepository(db),
		roundRepo:             repository.NewAgentBattleRoundRepository(db),
		contractRepo:          repository.NewAgentContractRepository(db),
		eventRepo:             repository.NewAgentBattleEventRepository(db),
		battleScoreRepo:       repository.NewAgentBattleScoreRepository(db),
	}
}

// services 聚合入口装配阶段创建出的服务实例。
type services struct {
	authService          *service.AuthService
	userService          *service.UserService
	schoolService        *service.SchoolService
	courseService        *service.CourseService
	uploadService        *service.UploadService
	contestService       *service.ContestService
	notificationService  *service.NotificationService
	discussionService    *service.DiscussionService
	challengeService     *service.ChallengeService
	vulnerabilityService *service.VulnerabilityAdminService
	systemService        *service.SystemService
	agentBattleService   *service.AgentBattleService
	workspaceService     *service.WorkspaceAccessService
	imageService         *service.ImageService
	experimentService    *service.ExperimentService
	schedulerService     *service.SchedulerService
	taskHandlerService   *service.TaskHandlerService
	taskManager          *task.Manager
	mqClient             *mq.Client
	wsHub                *websocket.Hub
}

// newServices 统一完成入口层所需服务实例和异步组件接线。
func newServices(
	cfg *config.Config,
	db *gorm.DB,
	redisClient *redis.Client,
	jwtManager *jwt.Manager,
	k8sClient *k8s.Client,
	repos *repositories,
) (*services, error) {
	authService := service.NewAuthService(repos.userRepo, repos.schoolRepo, jwtManager)
	userService := service.NewUserService(repos.userRepo, repos.schoolRepo, repos.classRepo)
	schoolService := service.NewSchoolService(repos.schoolRepo, repos.userRepo, repos.classRepo)
	courseService := service.NewCourseService(repos.courseRepo, repos.courseStudentRepo, repos.chapterRepo, repos.materialRepo, repos.materialProgressRepo, repos.userRepo, repos.classRepo)

	uploadService, err := service.NewUploadService(&cfg.MinIO, &cfg.Upload)
	if err != nil {
		return nil, fmt.Errorf("init upload service: %w", err)
	}

	challengeEnvService := service.NewChallengeEnvService()
	challengeWorkspaceSeedService := service.NewChallengeWorkspaceSeedService(k8sClient)
	contestRuntimeService := service.NewContestRuntimeService(
		repos.challengeEnvRepo,
		repos.challengeRepo,
		uploadService,
		k8sClient,
		challengeEnvService,
		challengeWorkspaceSeedService,
	)
	contestScoreService := service.NewContestScoreService(
		repos.contestRepo,
		repos.challengeRepo,
		repos.contestChallengeRepo,
		repos.teamRepo,
		repos.contestSubmissionRepo,
		repos.contestScoreRepo,
		redisClient,
	)
	contestService := service.NewContestService(
		repos.contestRepo,
		repos.challengeRepo,
		repos.contestChallengeRepo,
		repos.teamRepo,
		repos.teamMemberRepo,
		repos.contestSubmissionRepo,
		repos.contestScoreRepo,
		repos.userRepo,
		repos.dockerImageRepo,
		contestRuntimeService,
		contestScoreService,
		uploadService,
	)
	notificationService := service.NewNotificationService(repos.notificationRepo, repos.userRepo, redisClient)
	discussionService := service.NewDiscussionService(repos.postRepo, repos.replyRepo, repos.postLikeRepo, repos.replyLikeRepo)

	configCache := cache.NewConfigCache(&cache.ConfigCacheOptions{
		Redis:  redisClient,
		TTL:    5 * time.Minute,
		Prefix: "config:",
	})
	configService := service.NewConfigService(repos.configRepo, configCache)
	if err := configService.PreloadCache(context.Background()); err != nil {
		logger.Warn("Preload config cache failed", zap.Error(err))
	}

	service.SetProxyConfig(cfg.Proxy.HTTP, cfg.Proxy.HTTPS, cfg.Proxy.NoProxy)

	challengeVulnerabilityService := service.NewChallengeVulnerabilityService(
		repos.challengeRepo,
		repos.vulnRepo,
		cfg.Etherscan.APIKey,
		cfg.Proxy.HTTP,
		cfg.Proxy.HTTPS,
		cfg.Proxy.NoProxy,
	)
	challengeService := service.NewChallengeService(repos.challengeRepo, repos.dockerImageRepo)
	crossSchoolService := service.NewCrossSchoolService(repos.crossSchoolRepo)
	systemVulnerabilityService := service.NewSystemVulnerabilityService(repos.vulnSourceRepo, repos.vulnRepo)
	vulnerabilityAdminService := service.NewVulnerabilityAdminService(systemVulnerabilityService, challengeVulnerabilityService)
	systemMonitorService := service.NewSystemMonitorService(db, redisClient, uploadService, cfg.Kubernetes.Enabled)
	systemService := service.NewSystemService(
		repos.configRepo,
		repos.challengeRepo,
		repos.operationLogRepo,
		configService,
		crossSchoolService,
		systemMonitorService,
	)
	battleChainRuntimeService := service.NewBattleChainRuntimeService(repos.contestRepo, k8sClient)
	agentBattleScoreService := service.NewAgentBattleScoreService(repos.contractRepo, repos.eventRepo, repos.battleScoreRepo, repos.teamRepo)
	agentContractService := service.NewAgentContractService(repos.contractRepo, k8sClient)
	battleWorkspaceService := service.NewBattleWorkspaceService(repos.teamRepo, repos.contestRepo, repos.roundRepo, k8sClient, challengeEnvService)
	agentBattleService := service.NewAgentBattleService(
		repos.roundRepo,
		repos.contractRepo,
		repos.eventRepo,
		repos.battleScoreRepo,
		repos.contestRepo,
		repos.teamRepo,
		battleChainRuntimeService,
		agentBattleScoreService,
		agentContractService,
		battleWorkspaceService,
	)
	imageService := service.NewImageService(repos.dockerImageRepo)
	workspaceService := service.NewWorkspaceAccessService(repos.envRepo, repos.challengeEnvRepo, repos.challengeRepo, k8sClient)

	wsHub := websocket.NewHub()
	envManagerService := service.NewEnvManagerService(
		k8sClient,
		repos.sessionRepo,
		repos.sessionMemberRepo,
		repos.envRepo,
		repos.experimentRepo,
		repos.dockerImageRepo,
		repos.notificationRepo,
		uploadService,
		wsHub,
		cfg,
	)
	experimentGradingService := service.NewExperimentGradingService(repos.submissionRepo, repos.envRepo, k8sClient)
	experimentService := service.NewExperimentService(repos.experimentRepo, repos.chapterRepo, repos.sessionRepo, repos.sessionMemberRepo, repos.sessionMessageRepo, repos.envRepo, repos.submissionRepo, repos.dockerImageRepo, envManagerService, experimentGradingService)

	mqClient, err := newMQClient(cfg)
	if err != nil {
		return nil, err
	}

	taskManager := task.NewManager(db, 5)
	schedulerService := service.NewSchedulerService(envManagerService, notificationService)
	taskHandlerService := service.NewTaskHandlerService(taskManager, mqClient, envManagerService, vulnerabilityAdminService, repos.notificationRepo)
	schedulerService.SetTaskHandler(taskHandlerService)

	return &services{
		authService:          authService,
		userService:          userService,
		schoolService:        schoolService,
		courseService:        courseService,
		uploadService:        uploadService,
		contestService:       contestService,
		notificationService:  notificationService,
		discussionService:    discussionService,
		challengeService:     challengeService,
		vulnerabilityService: vulnerabilityAdminService,
		systemService:        systemService,
		agentBattleService:   agentBattleService,
		workspaceService:     workspaceService,
		imageService:         imageService,
		experimentService:    experimentService,
		schedulerService:     schedulerService,
		taskHandlerService:   taskHandlerService,
		taskManager:          taskManager,
		mqClient:             mqClient,
		wsHub:                wsHub,
	}, nil
}

// newMQClient 根据配置创建可选的 RabbitMQ 客户端。
func newMQClient(cfg *config.Config) (*mq.Client, error) {
	if cfg.RabbitMQ.Host == "" {
		return nil, nil
	}

	mqClient, err := mq.NewClient(&mq.Config{
		URL:            cfg.RabbitMQ.URL(),
		ReconnectDelay: 5 * time.Second,
		MaxRetries:     10,
	})
	if err != nil {
		logger.Warn("RabbitMQ connect failed, fallback to sync mode", zap.Error(err))
		return nil, nil
	}

	logger.Info("RabbitMQ connected")
	return mqClient, nil
}

// handlers 聚合入口装配阶段所需的 HTTP 处理器。
type handlers struct {
	authHandler                 *handler.AuthHandler
	userHandler                 *handler.UserHandler
	schoolHandler               *handler.SchoolHandler
	courseHandler               *handler.CourseHandler
	experimentHandler           *handler.ExperimentHandler
	contestHandler              *handler.ContestHandler
	notificationHandler         *handler.NotificationHandler
	discussionHandler           *handler.DiscussionHandler
	systemHandler               *handler.SystemHandler
	uploadHandler               *handler.UploadHandler
	agentBattleHandler          *handler.AgentBattleHandler
	websocketHandler            *handler.WebSocketHandler
	imageHandler                *handler.ImageHandler
	challengeHandler            *handler.ChallengeHandler
	envProxyHandler             *handler.EnvProxyHandler
	contestEnvProxyHandler      *handler.ContestEnvProxyHandler
	battleWorkspaceProxyHandler *handler.BattleWorkspaceProxyHandler
}

// newHandlers 统一创建入口层需要的处理器实例。
func newHandlers(services *services) *handlers {
	return &handlers{
		authHandler:                 handler.NewAuthHandler(services.authService),
		userHandler:                 handler.NewUserHandler(services.userService),
		schoolHandler:               handler.NewSchoolHandler(services.schoolService),
		courseHandler:               handler.NewCourseHandler(services.courseService),
		experimentHandler:           handler.NewExperimentHandler(services.experimentService),
		contestHandler:              handler.NewContestHandler(services.contestService),
		notificationHandler:         handler.NewNotificationHandler(services.notificationService),
		discussionHandler:           handler.NewDiscussionHandler(services.discussionService),
		systemHandler:               handler.NewSystemHandler(services.systemService, services.vulnerabilityService, services.schedulerService, services.taskHandlerService),
		uploadHandler:               handler.NewUploadHandler(services.uploadService),
		agentBattleHandler:          handler.NewAgentBattleHandler(services.agentBattleService),
		websocketHandler:            handler.NewWebSocketHandler(services.wsHub),
		imageHandler:                handler.NewImageHandler(services.imageService),
		challengeHandler:            handler.NewChallengeHandler(services.challengeService),
		envProxyHandler:             handler.NewEnvProxyHandler(services.workspaceService),
		contestEnvProxyHandler:      handler.NewContestEnvProxyHandler(services.workspaceService),
		battleWorkspaceProxyHandler: handler.NewBattleWorkspaceProxyHandler(services.workspaceService),
	}
}
