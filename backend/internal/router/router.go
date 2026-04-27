package router

import (
	"github.com/chainspace/backend/internal/config"
	"github.com/chainspace/backend/internal/handler"
	"github.com/chainspace/backend/internal/middleware"
	"github.com/chainspace/backend/internal/model"
	"github.com/chainspace/backend/internal/pkg/jwt"
	"github.com/chainspace/backend/internal/repository"
	"github.com/chainspace/backend/internal/uiassets"
	"github.com/gin-gonic/gin"
)

// Router 路由管理器
type Router struct {
	engine *gin.Engine
	cfg    *config.Config
	jwt    *jwt.Manager

	// Repositories
	operationLogRepo *repository.OperationLogRepository

	// Handlers
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

// NewRouter 创建路由管理器
func NewRouter(
	cfg *config.Config,
	jwtManager *jwt.Manager,
	operationLogRepo *repository.OperationLogRepository,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	schoolHandler *handler.SchoolHandler,
	courseHandler *handler.CourseHandler,
	experimentHandler *handler.ExperimentHandler,
	contestHandler *handler.ContestHandler,
	notificationHandler *handler.NotificationHandler,
	discussionHandler *handler.DiscussionHandler,
	systemHandler *handler.SystemHandler,
	uploadHandler *handler.UploadHandler,
	agentBattleHandler *handler.AgentBattleHandler,
	websocketHandler *handler.WebSocketHandler,
	imageHandler *handler.ImageHandler,
	challengeHandler *handler.ChallengeHandler,
	envProxyHandler *handler.EnvProxyHandler,
	contestEnvProxyHandler *handler.ContestEnvProxyHandler,
	battleWorkspaceProxyHandler *handler.BattleWorkspaceProxyHandler,
) *Router {
	return &Router{
		cfg:                         cfg,
		jwt:                         jwtManager,
		operationLogRepo:            operationLogRepo,
		authHandler:                 authHandler,
		userHandler:                 userHandler,
		schoolHandler:               schoolHandler,
		courseHandler:               courseHandler,
		experimentHandler:           experimentHandler,
		contestHandler:              contestHandler,
		notificationHandler:         notificationHandler,
		discussionHandler:           discussionHandler,
		systemHandler:               systemHandler,
		uploadHandler:               uploadHandler,
		agentBattleHandler:          agentBattleHandler,
		websocketHandler:            websocketHandler,
		imageHandler:                imageHandler,
		challengeHandler:            challengeHandler,
		envProxyHandler:             envProxyHandler,
		contestEnvProxyHandler:      contestEnvProxyHandler,
		battleWorkspaceProxyHandler: battleWorkspaceProxyHandler,
	}
}

// Setup 设置路由
func (r *Router) Setup() *gin.Engine {
	if r.cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r.engine = gin.New()

	// 全局中间件
	r.engine.Use(middleware.Recovery())
	r.engine.Use(middleware.Logger())
	r.engine.Use(middleware.CORS(&r.cfg.CORS))
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.RateLimit(&r.cfg.RateLimit))

	// API路由组
	api := r.engine.Group("/api/v1")

	// 认证路由（无需登录）
	r.setupAuthRoutes(api)

	// 需要认证的路由
	authenticated := api.Group("")
	authenticated.Use(middleware.Auth(r.jwt))
	authenticated.Use(middleware.TenantIsolation())
	authenticated.Use(middleware.AuditLogger(r.operationLogRepo))

	r.setupUserRoutes(authenticated)
	r.setupSchoolRoutes(authenticated)
	r.setupCourseRoutes(authenticated)
	r.setupExperimentRoutes(authenticated)
	r.setupContestRoutes(authenticated)
	r.setupNotificationRoutes(authenticated)
	r.setupDiscussionRoutes(authenticated)
	r.setupSystemRoutes(authenticated)
	r.setupUploadRoutes(authenticated)
	r.setupAgentBattleRoutes(authenticated)
	r.setupImageRoutes(authenticated)
	r.setupChallengeRoutes(authenticated)
	r.setupWebSocketRoutes(api)

	// 实验环境代理路由（独立于认证组，iframe 无法携带 JWT）
	r.setupEnvProxyRoutes(api)

	// 健康检查
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	uiassets.Register(r.engine)

	return r.engine
}

// setupAuthRoutes 设置认证路由
func (r *Router) setupAuthRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/refresh", r.authHandler.RefreshToken)

		// 需要认证
		authRequired := auth.Group("")
		authRequired.Use(middleware.Auth(r.jwt))
		{
			authRequired.POST("/logout", r.authHandler.Logout)
			authRequired.GET("/me", r.authHandler.GetCurrentUser)
			authRequired.PUT("/me", r.authHandler.UpdateProfile)
			authRequired.PUT("/password", r.authHandler.ChangePassword)
			authRequired.POST("/reset-password", middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin), r.authHandler.ResetPassword)
		}
	}
}

// setupUserRoutes 设置用户路由
func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	users.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin))
	{
		users.GET("", r.userHandler.ListUsers)
		users.POST("", r.userHandler.CreateUser)
		users.GET("/import-template", r.userHandler.DownloadImportTemplate)
		users.GET("/:id", r.userHandler.GetUser)
		users.PUT("/:id", r.userHandler.UpdateUser)
		users.DELETE("/:id", r.userHandler.DeleteUser)
		users.PUT("/:id/status", r.userHandler.UpdateUserStatus)
		users.POST("/batch-import", r.userHandler.BatchImportStudents)
	}
}

// setupSchoolRoutes 设置学校路由
func (r *Router) setupSchoolRoutes(rg *gin.RouterGroup) {
	// 当前学校信息（学校管理员可访问自己学校）
	schoolCurrent := rg.Group("/schools/current")
	schoolCurrent.Use(middleware.RequireRoles(model.RoleSchoolAdmin))
	{
		schoolCurrent.GET("", r.schoolHandler.GetCurrentSchool)
		schoolCurrent.PUT("", r.schoolHandler.UpdateCurrentSchool)
	}

	// 学校管理（仅平台管理员）
	schools := rg.Group("/schools")
	schools.Use(middleware.RequireRoles(model.RolePlatformAdmin))
	{
		schools.GET("", r.schoolHandler.ListSchools)
		schools.POST("", r.schoolHandler.CreateSchool)
		schools.GET("/:id", r.schoolHandler.GetSchool)
		schools.PUT("/:id", r.schoolHandler.UpdateSchool)
		schools.PUT("/:id/status", r.schoolHandler.UpdateSchoolStatus)
		schools.DELETE("/:id", r.schoolHandler.DeleteSchool)
	}

	// 班级管理
	classes := rg.Group("/classes")
	classes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
	{
		classes.GET("", r.schoolHandler.ListClasses)
		classes.POST("", r.schoolHandler.CreateClass)
		classes.GET("/:id", r.schoolHandler.GetClass)
		classes.PUT("/:id", r.schoolHandler.UpdateClass)
		classes.DELETE("/:id", r.schoolHandler.DeleteClass)
		classes.GET("/:id/students", r.schoolHandler.ListClassStudents)
	}
}

// setupCourseRoutes 设置课程路由
func (r *Router) setupCourseRoutes(rg *gin.RouterGroup) {
	courses := rg.Group("/courses")
	{
		// 所有认证用户可访问
		courses.GET("", r.courseHandler.ListCourses)
		courses.GET("/my", r.courseHandler.ListMyCourses)
		courses.GET("/:id", r.courseHandler.GetCourse)
		courses.GET("/:id/chapters", r.courseHandler.ListChapters)
		courses.GET("/:id/chapters/:chapter_id/materials", r.courseHandler.ListMaterials)
		courses.POST("/join", r.courseHandler.JoinCourse)

		// 教师和管理员
		teacherRoutes := courses.Group("")
		teacherRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			teacherRoutes.POST("", r.courseHandler.CreateCourse)
			teacherRoutes.PUT("/:id", r.courseHandler.UpdateCourse)
			teacherRoutes.PUT("/:id/status", r.courseHandler.UpdateCourseStatus)
			teacherRoutes.POST("/:id/invite-code/reset", r.courseHandler.ResetInviteCode)
			teacherRoutes.DELETE("/:id", r.courseHandler.DeleteCourse)
			teacherRoutes.POST("/:id/students", r.courseHandler.AddStudents)
			teacherRoutes.DELETE("/:id/students", r.courseHandler.RemoveStudents)
			teacherRoutes.GET("/:id/students", r.courseHandler.ListCourseStudents)
			teacherRoutes.POST("/:id/students/import", r.courseHandler.ImportStudentsFromExcel)

			// 章节
			teacherRoutes.POST("/:id/chapters", r.courseHandler.CreateChapter)
			teacherRoutes.PUT("/:id/chapters/:chapter_id", r.courseHandler.UpdateChapter)
			teacherRoutes.DELETE("/:id/chapters/:chapter_id", r.courseHandler.DeleteChapter)
			teacherRoutes.PUT("/:id/chapters/sort", r.courseHandler.ReorderChapters)

			// 资料
			teacherRoutes.POST("/:id/chapters/:chapter_id/materials", r.courseHandler.CreateMaterial)
			teacherRoutes.PUT("/:id/chapters/:chapter_id/materials/:material_id", r.courseHandler.UpdateMaterial)
			teacherRoutes.DELETE("/:id/chapters/:chapter_id/materials/:material_id", r.courseHandler.DeleteMaterial)
		}

		// 学习进度（学生）
		courses.GET("/:id/progress", r.courseHandler.GetCourseProgress)
		courses.PUT("/:id/materials/:material_id/progress", r.courseHandler.UpdateMaterialProgress)
	}
}

// setupExperimentRoutes 设置实验路由
func (r *Router) setupExperimentRoutes(rg *gin.RouterGroup) {
	experiments := rg.Group("/experiments")
	{
		// 所有认证用户可访问
		experiments.GET("", r.experimentHandler.ListExperiments)
		experiments.GET("/:id", r.experimentHandler.GetExperiment)

		// 教师和管理员
		teacherRoutes := experiments.Group("")
		teacherRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			teacherRoutes.POST("", r.experimentHandler.CreateExperiment)
			teacherRoutes.PUT("/:id", r.experimentHandler.UpdateExperiment)
			teacherRoutes.PUT("/:id/publish", r.experimentHandler.PublishExperiment)
			teacherRoutes.DELETE("/:id", r.experimentHandler.DeleteExperiment)
		}
	}

	// 实验环境
	envs := rg.Group("/envs")
	{
		envs.GET("", r.experimentHandler.ListEnvs)
		envs.POST("/start", r.experimentHandler.StartEnv)
		envs.GET("/:env_id", r.experimentHandler.GetEnvStatus)
		envs.POST("/:env_id/stop", r.experimentHandler.StopEnv)
		envs.POST("/:env_id/extend", r.experimentHandler.ExtendEnv)
		envs.POST("/:env_id/pause", r.experimentHandler.PauseEnv)
		envs.POST("/:env_id/resume", r.experimentHandler.ResumeEnv)
		envs.POST("/:env_id/snapshots", r.experimentHandler.CreateSnapshot)
	}

	// 提交
	sessions := rg.Group("/experiment-sessions")
	{
		sessions.GET("", r.experimentHandler.ListSessions)
		sessions.GET("/:session_key", r.experimentHandler.GetSession)
		sessions.GET("/:session_key/messages", r.experimentHandler.ListSessionMessages)
		sessions.GET("/:session_key/logs", r.experimentHandler.GetSessionLogs)
		sessions.POST("/:session_key/messages", r.experimentHandler.SendSessionMessage)
		sessions.POST("/:session_key/join", r.experimentHandler.JoinSession)
		sessions.POST("/:session_key/leave", r.experimentHandler.LeaveSession)
		sessions.PUT("/:session_key/members/:user_id", r.experimentHandler.UpdateSessionMember)
	}

	submissions := rg.Group("/submissions")
	{
		submissions.POST("", r.experimentHandler.SubmitExperiment)
		submissions.GET("", r.experimentHandler.ListSubmissions)

		// 批改（教师）
		teacherRoutes := submissions.Group("")
		teacherRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			teacherRoutes.POST("/:id/grade", r.experimentHandler.GradeSubmission)
		}
	}
}

// setupContestRoutes 设置竞赛路由
func (r *Router) setupContestRoutes(rg *gin.RouterGroup) {
	contests := rg.Group("/contests")
	{
		// 所有认证用户可访问
		contests.GET("", r.contestHandler.ListContests)
		contests.GET("/my-records", r.contestHandler.GetMyContestRecords)
		contests.GET("/:id", r.contestHandler.GetContest)
		contests.GET("/:id/scoreboard", r.contestHandler.GetScoreboard)
		contests.GET("/:id/my-team", r.contestHandler.GetMyTeam)
		contests.POST("/:id/submit", r.contestHandler.SubmitFlag)
		contests.POST("/:id/register", r.contestHandler.RegisterContest)
		contests.GET("/:id/review-challenges", r.contestHandler.GetContestReviewChallenges)

		// 参赛者可访问的题目
		contests.GET("/:id/challenges", r.contestHandler.GetContestChallenges)
		contests.GET("/:id/challenges/:challenge_id/attachments/:attachment_index/access-url", r.contestHandler.GetChallengeAttachmentAccessURL)
		contests.POST("/:id/agent/upload", r.contestHandler.UploadAgentCode)

		// 题目环境
		contests.POST("/:id/challenges/:challenge_id/env", r.contestHandler.StartChallengeEnv)
		contests.GET("/:id/challenges/:challenge_id/env", r.contestHandler.GetChallengeEnvStatus)
		contests.DELETE("/:id/challenges/:challenge_id/env", r.contestHandler.StopChallengeEnv)

		// 管理员和教师
		adminRoutes := contests.Group("")
		adminRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			adminRoutes.POST("", r.contestHandler.CreateContest)
			adminRoutes.PUT("/:id", r.contestHandler.UpdateContest)
			adminRoutes.PUT("/:id/publish", r.contestHandler.PublishContest)
			adminRoutes.DELETE("/:id", r.contestHandler.DeleteContest)
			// 竞赛题目管理
			adminRoutes.GET("/:id/challenges/admin", r.contestHandler.ListContestChallengesAdmin)
			adminRoutes.POST("/:id/challenges", r.contestHandler.AddChallengeToContest)
			adminRoutes.DELETE("/:id/challenges/:challenge_id", r.contestHandler.RemoveChallengeFromContest)
		}
	}

	// 队伍
	teams := rg.Group("/teams")
	{
		teams.GET("/my", r.contestHandler.GetMyTeams)
		teams.POST("", r.contestHandler.CreateTeam)
		teams.POST("/join", r.contestHandler.JoinTeam)
		teams.POST("/:team_id/leave", r.contestHandler.LeaveTeam)
		teams.POST("/:team_id/invite", r.contestHandler.InviteTeamMember)
		teams.DELETE("/:team_id/members/:user_id", r.contestHandler.RemoveTeamMember)
	}
}

// setupNotificationRoutes 设置通知路由
func (r *Router) setupNotificationRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	{
		notifications.GET("", r.notificationHandler.ListNotifications)
		notifications.GET("/unread-count", r.notificationHandler.GetUnreadCount)
		notifications.POST("/read", r.notificationHandler.MarkAsRead)
		notifications.POST("/read-all", r.notificationHandler.MarkAllAsRead)
		notifications.DELETE("/:id", r.notificationHandler.DeleteNotification)
		notifications.POST("/batch-delete", r.notificationHandler.BatchDeleteNotifications)

		// 管理员
		adminRoutes := notifications.Group("")
		adminRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin))
		{
			adminRoutes.POST("/send", r.notificationHandler.SendNotification)
			adminRoutes.POST("/broadcast", r.notificationHandler.BroadcastNotification)
		}
	}
}

// setupDiscussionRoutes 设置讨论路由
func (r *Router) setupDiscussionRoutes(rg *gin.RouterGroup) {
	posts := rg.Group("/posts")
	{
		posts.GET("", r.discussionHandler.ListPosts)
		posts.GET("/:id", r.discussionHandler.GetPost)
		posts.POST("", r.discussionHandler.CreatePost)
		posts.PUT("/:id", r.discussionHandler.UpdatePost)
		posts.DELETE("/:id", r.discussionHandler.DeletePost)

		// 点赞
		posts.POST("/:id/like", r.discussionHandler.LikePost)
		posts.DELETE("/:id/like", r.discussionHandler.UnlikePost)

		// 回复
		posts.GET("/:id/replies", r.discussionHandler.ListReplies)
		posts.POST("/:id/replies", r.discussionHandler.CreateReply)
		posts.DELETE("/:id/replies/:reply_id", r.discussionHandler.DeleteReply)
		posts.POST("/:id/replies/:reply_id/like", r.discussionHandler.LikeReply)
		posts.DELETE("/:id/replies/:reply_id/like", r.discussionHandler.UnlikeReply)

		// 采纳回复（帖子作者）
		posts.POST("/:id/accept", r.discussionHandler.AcceptReply)

		// 管理员操作
		adminRoutes := posts.Group("")
		adminRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			adminRoutes.POST("/:id/pin", r.discussionHandler.PinPost)
			adminRoutes.POST("/:id/lock", r.discussionHandler.LockPost)
		}
	}
}

// setupSystemRoutes 设置系统路由
func (r *Router) setupSystemRoutes(rg *gin.RouterGroup) {
	system := rg.Group("/system")
	{
		system.GET("/health", r.systemHandler.HealthCheck)
		system.GET("/stats", middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin), r.systemHandler.GetStats)

		// 配置管理（仅平台管理员）
		configs := system.Group("/configs")
		configs.Use(middleware.RequireRoles(model.RolePlatformAdmin))
		{
			configs.GET("", r.systemHandler.ListConfigs)
			configs.GET("/:key", r.systemHandler.GetConfig)
			configs.POST("", r.systemHandler.SetConfig)
		}

		// 操作日志
		logs := system.Group("/logs")
		logs.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin))
		{
			logs.GET("", r.systemHandler.ListOperationLogs)
		}

		// 跨校申请
		crossSchool := system.Group("/cross-school")
		{
			crossSchool.GET("", r.systemHandler.ListCrossSchoolApplications)
			crossSchool.POST("", r.systemHandler.CreateCrossSchoolApplication)
			crossSchool.POST("/:id/handle", middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin), r.systemHandler.HandleCrossSchoolApplication)
		}

		// 定时任务（仅平台管理员）
		scheduler := system.Group("/scheduler")
		scheduler.Use(middleware.RequireRoles(model.RolePlatformAdmin))
		{
			scheduler.GET("/status", r.systemHandler.GetSchedulerStatus)
			scheduler.POST("/tasks/:task/run", r.systemHandler.RunSchedulerTask)
		}

		// 系统监控（仅平台管理员）
		monitor := system.Group("")
		monitor.Use(middleware.RequireRoles(model.RolePlatformAdmin))
		{
			monitor.GET("/monitor", r.systemHandler.GetSystemMonitor)
			monitor.GET("/containers", r.systemHandler.GetContainerStats)
			monitor.GET("/services", r.systemHandler.GetServiceHealth)
		}

		// 漏洞管理（仅平台管理员）
		vulns := system.Group("/vulnerabilities")
		vulns.Use(middleware.RequireRoles(model.RolePlatformAdmin))
		{
			vulns.GET("", r.systemHandler.ListVulnerabilities)
			vulns.PUT("/:id", r.systemHandler.UpdateVulnerability)
			vulns.POST("/:id/convert", r.systemHandler.ConvertVulnerability)
			vulns.POST("/:id/enrich", r.systemHandler.EnrichVulnerabilityCode)
			vulns.PUT("/:id/skip", r.systemHandler.SkipVulnerability)
			vulns.POST("/sync", r.systemHandler.SyncVulnerabilities)
		}

		// 题目公开审核
		challengeReviews := system.Group("/challenge-reviews")
		challengeReviews.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin))
		{
			challengeReviews.GET("", r.systemHandler.ListChallengePublishRequests)
			challengeReviews.POST("/:id/handle", r.systemHandler.HandleChallengePublishRequest)
		}
	}
}

// setupUploadRoutes 设置上传路由
func (r *Router) setupUploadRoutes(rg *gin.RouterGroup) {
	upload := rg.Group("/upload")
	{
		upload.POST("/avatar", r.uploadHandler.UploadAvatar)
		upload.POST("/courses/:course_id/cover", r.uploadHandler.UploadCourseCover)
		upload.POST("/courses/:course_id/chapters/:chapter_id/materials", r.uploadHandler.UploadMaterial)
		upload.POST("/experiments/assets", r.uploadHandler.UploadExperimentAsset)
		upload.POST("/experiments/:experiment_id/submission", r.uploadHandler.UploadSubmission)
		upload.POST("/challenges/:challenge_id/attachment", r.uploadHandler.UploadChallengeAttachment)
	}
}

// setupAgentBattleRoutes 设置智能体对战路由
func (r *Router) setupAgentBattleRoutes(rg *gin.RouterGroup) {
	battle := rg.Group("/agent-battle")
	{
		// 轮次管理
		battle.GET("/contests/:contest_id/rounds", r.agentBattleHandler.ListRounds)
		battle.GET("/contests/:contest_id/current-round", r.agentBattleHandler.GetCurrentRound)
		battle.GET("/contests/:contest_id/scoreboard", r.agentBattleHandler.GetScoreboard)
		battle.GET("/contests/:contest_id/status", r.agentBattleHandler.GetStatus)
		battle.GET("/contests/:contest_id/spectate", r.agentBattleHandler.GetSpectateData)
		battle.GET("/contests/:contest_id/events", r.agentBattleHandler.GetContestEvents)
		battle.GET("/contests/:contest_id/replay", r.agentBattleHandler.GetReplayData)
		battle.GET("/contests/:contest_id/config", r.agentBattleHandler.GetBattleConfig)

		// 队伍工作区（需要认证）
		workspaceRoutes := battle.Group("/contests/:contest_id/workspace")
		workspaceRoutes.Use(middleware.Auth(r.jwt))
		{
			workspaceRoutes.GET("", r.agentBattleHandler.GetMyTeamWorkspace)
			workspaceRoutes.POST("", r.agentBattleHandler.CreateTeamWorkspace)
			workspaceRoutes.DELETE("", r.agentBattleHandler.StopTeamWorkspace)
		}

		// 合约部署
		battle.POST("/contracts/deploy", r.agentBattleHandler.DeployContract)
		battle.POST("/contracts/upgrade", r.agentBattleHandler.UpgradeContract)
		battle.GET("/contests/:contest_id/teams/:team_id/contract", r.agentBattleHandler.GetContract)
		battle.GET("/contests/:contest_id/final-rank", r.agentBattleHandler.GetFinalRank)

		// 事件查询
		battle.GET("/rounds/:round_id/events", r.agentBattleHandler.ListEvents)

		// 管理员操作
		adminRoutes := battle.Group("")
		adminRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			adminRoutes.POST("/contests/:contest_id/rounds", r.agentBattleHandler.CreateRound)
		}
	}
}

// setupWebSocketRoutes 设置WebSocket路由
func (r *Router) setupWebSocketRoutes(rg *gin.RouterGroup) {
	ws := rg.Group("/ws")
	ws.Use(middleware.Auth(r.jwt))
	{
		ws.GET("/connect", r.websocketHandler.Connect)
		ws.POST("/envs/:env_id/join", r.websocketHandler.JoinEnvRoom)
		ws.POST("/envs/:env_id/leave", r.websocketHandler.LeaveEnvRoom)
		ws.POST("/sessions/:session_key/join", r.websocketHandler.JoinSessionRoom)
		ws.POST("/sessions/:session_key/leave", r.websocketHandler.LeaveSessionRoom)
		ws.POST("/contests/:contest_id/join", r.websocketHandler.JoinContestRoom)
		ws.POST("/contests/:contest_id/leave", r.websocketHandler.LeaveContestRoom)
	}
}

// setupImageRoutes 设置镜像管理路由
func (r *Router) setupImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	{
		// 教师也需要查看镜像列表（创建实验时选择镜像）
		images.GET("", r.imageHandler.ListImages)
		images.GET("/all", r.imageHandler.ListAllImages)
		images.GET("/capabilities", r.imageHandler.ListImageCapabilities)

		// 创建/更新/删除需要平台管理员权限
		adminImages := images.Group("")
		adminImages.Use(middleware.RequireRoles(model.RolePlatformAdmin))
		{
			adminImages.POST("", r.imageHandler.CreateImage)
			adminImages.PUT("/:id", r.imageHandler.UpdateImage)
			adminImages.DELETE("/:id", r.imageHandler.DeleteImage)
		}
	}
}

// setupChallengeRoutes 设置题目管理路由
func (r *Router) setupChallengeRoutes(rg *gin.RouterGroup) {
	challenges := rg.Group("/challenges")
	{
		challenges.GET("", r.challengeHandler.ListChallenges)
		challenges.GET("/:id", r.challengeHandler.GetChallenge)

		// 需要教师权限
		teacherRoutes := challenges.Group("")
		teacherRoutes.Use(middleware.RequireRoles(model.RolePlatformAdmin, model.RoleSchoolAdmin, model.RoleTeacher))
		{
			teacherRoutes.POST("", r.challengeHandler.CreateChallenge)
			teacherRoutes.PUT("/:id", r.challengeHandler.UpdateChallenge)
			teacherRoutes.DELETE("/:id", r.challengeHandler.DeleteChallenge)
			teacherRoutes.POST("/:id/publish-request", r.challengeHandler.RequestPublish)
		}
	}
}

// setupEnvProxyRoutes 设置实验环境代理路由（无需JWT，iframe无法携带token）
func (r *Router) setupEnvProxyRoutes(rg *gin.RouterGroup) {
	proxy := rg.Group("/envs")
	{
		proxy.Any("/:env_id/proxy/ide/*path", r.envProxyHandler.ProxyIDE)
		proxy.Any("/:env_id/proxy/terminal/*path", r.envProxyHandler.ProxyTerminal)
		proxy.Any("/:env_id/proxy/rpc/*path", r.envProxyHandler.ProxyRPC)
		proxy.Any("/:env_id/proxy/api_debug/*path", r.envProxyHandler.ProxyAPIDebug)
		proxy.Any("/:env_id/proxy/visualization/*path", r.envProxyHandler.ProxySim)
		proxy.Any("/:env_id/proxy/explorer/*path", r.envProxyHandler.ProxyExplorer)
		proxy.Any("/:env_id/proxy/network/*path", r.envProxyHandler.ProxyNetwork)
		proxy.Any("/:env_id/proxy/files/*path", r.envProxyHandler.ProxyFiles)
		proxy.Any("/:env_id/proxy/logs/*path", r.envProxyHandler.ProxyLogs)
		proxy.Any("/:env_id/instances/:instance_key/proxy/ide/*path", r.envProxyHandler.ProxyInstanceIDE)
		proxy.Any("/:env_id/instances/:instance_key/proxy/terminal/*path", r.envProxyHandler.ProxyInstanceTerminal)
		proxy.Any("/:env_id/instances/:instance_key/proxy/rpc/*path", r.envProxyHandler.ProxyInstanceRPC)
		proxy.Any("/:env_id/instances/:instance_key/proxy/api_debug/*path", r.envProxyHandler.ProxyInstanceAPIDebug)
		proxy.Any("/:env_id/instances/:instance_key/proxy/visualization/*path", r.envProxyHandler.ProxyInstanceSim)
		proxy.Any("/:env_id/instances/:instance_key/proxy/explorer/*path", r.envProxyHandler.ProxyInstanceExplorer)
		proxy.Any("/:env_id/instances/:instance_key/proxy/network/*path", r.envProxyHandler.ProxyInstanceNetwork)
		proxy.Any("/:env_id/instances/:instance_key/proxy/files/*path", r.envProxyHandler.ProxyInstanceFiles)
		proxy.Any("/:env_id/instances/:instance_key/proxy/logs/*path", r.envProxyHandler.ProxyInstanceLogs)
	}

	// 解题赛题目环境代理路由（无需JWT）
	contestEnv := rg.Group("/contest-envs")
	{
		contestEnv.Any("/:env_id/proxy/ide/*path", r.contestEnvProxyHandler.ProxyIDE)
		contestEnv.Any("/:env_id/proxy/terminal/*path", r.contestEnvProxyHandler.ProxyTerminal)
		contestEnv.Any("/:env_id/proxy/rpc/*path", r.contestEnvProxyHandler.ProxyRPC)
		contestEnv.Any("/:env_id/proxy/api_debug/*path", r.contestEnvProxyHandler.ProxyAPIDebug)
		contestEnv.Any("/:env_id/proxy/visualization/*path", r.contestEnvProxyHandler.ProxySim)
		contestEnv.Any("/:env_id/proxy/explorer/*path", r.contestEnvProxyHandler.ProxyExplorer)
		contestEnv.Any("/:env_id/proxy/network/*path", r.contestEnvProxyHandler.ProxyNetwork)
		contestEnv.Any("/:env_id/proxy/services/:service_key/*path", r.contestEnvProxyHandler.ProxyService)
		contestEnv.Any("/:env_id/proxy/files/*path", r.contestEnvProxyHandler.ProxyFiles)
		contestEnv.Any("/:env_id/proxy/logs/*path", r.contestEnvProxyHandler.ProxyLogs)
	}

	// 对抗赛队伍工作区代理路由（无需JWT）
	battleWs := rg.Group("/battle-workspace")
	{
		battleWs.Any("/:env_id/ide/*path", r.battleWorkspaceProxyHandler.ProxyIDE)
		battleWs.Any("/:env_id/terminal/*path", r.battleWorkspaceProxyHandler.ProxyTerminal)
		battleWs.Any("/:env_id/rpc/*path", r.battleWorkspaceProxyHandler.ProxyRPC)
		battleWs.Any("/:env_id/api_debug/*path", r.battleWorkspaceProxyHandler.ProxyAPIDebug)
		battleWs.Any("/:env_id/files/*path", r.battleWorkspaceProxyHandler.ProxyFiles)
		battleWs.Any("/:env_id/logs/*path", r.battleWorkspaceProxyHandler.ProxyLogs)
		battleWs.Any("/:env_id/explorer/*path", r.battleWorkspaceProxyHandler.ProxyExplorer)
		battleWs.Any("/:env_id/network/*path", r.battleWorkspaceProxyHandler.ProxyNetwork)
		battleWs.Any("/:env_id/visualization/*path", r.battleWorkspaceProxyHandler.ProxyVisualization)
	}
}
