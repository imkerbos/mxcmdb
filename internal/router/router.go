package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imkerbos/mxcmdb/internal/config"
	"github.com/imkerbos/mxcmdb/internal/handler"
	"github.com/imkerbos/mxcmdb/internal/middleware"
	"github.com/imkerbos/mxcmdb/internal/pkg/ws"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"github.com/imkerbos/mxcmdb/internal/service"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Setup 初始化路由
func Setup(db *gorm.DB, rdb *redis.Client, cfg *config.Config, auditSvc *service.AuditLogService, configSvc *service.SystemConfigService) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

	// WebSocket Hub
	wsHub := ws.NewHub()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	cloudAccountRepo := repository.NewCloudAccountRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	sshKeyRepo := repository.NewSSHKeyRepository(db)
	probeRepo := repository.NewProbeResultRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	fileTaskRepo := repository.NewFileTaskRepository(db)
	linuxUserRepo := repository.NewLinuxUserRepository(db)
	termSessionRepo := repository.NewTerminalSessionRepository(db)
	ipamRepo := repository.NewIPAMRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	playbookRepo := repository.NewPlaybookRepository(db)

	// Services
	captchaSvc := service.NewCaptchaService(rdb, cfg.Captcha.Length, cfg.Captcha.Expire)
	mfaSvc := service.NewMFAService(userRepo, cfg.MFA.Issuer)
	authSvc := service.NewAuthService(
		userRepo, mfaSvc, configSvc, cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessExpire)*time.Second,
		time.Duration(cfg.JWT.RefreshExpire)*time.Second,
	)
	userSvc := service.NewUserService(userRepo)
	cloudAccountSvc := service.NewCloudAccountService(cloudAccountRepo, cfg.Encryption.MasterKey)
	assetSvc := service.NewAssetService(assetRepo, cfg.Encryption.MasterKey, configSvc, projectRepo, sshKeyRepo)
	assetSvc.SetDetailRepos(probeRepo, linuxUserRepo, termSessionRepo)
	sshKeySvc := service.NewSSHKeyService(sshKeyRepo, assetRepo, cfg.Encryption.MasterKey, configSvc)
	probeSvc := service.NewProbeService(probeRepo, assetRepo, cfg.Encryption.MasterKey, configSvc)
	taskSvc := service.NewTaskService(taskRepo, assetRepo, cfg.Encryption.MasterKey, configSvc, wsHub)
	fileDistSvc := service.NewFileDistributionService(fileTaskRepo, assetRepo, cfg.Encryption.MasterKey, configSvc)
	linuxUserSvc := service.NewLinuxUserService(linuxUserRepo, assetRepo, cfg.Encryption.MasterKey, configSvc)
	terminalSvc := service.NewTerminalService(termSessionRepo, assetRepo, cfg.Encryption.MasterKey, configSvc)
	terminalSvc.CleanupStaleSessions()
	ipamSvc := service.NewIPAMService(ipamRepo)
	projectSvc := service.NewProjectService(projectRepo, probeRepo)
	playbookSvc := service.NewPlaybookService(playbookRepo, assetRepo, cfg.Encryption.MasterKey, configSvc, wsHub)
	notificationRepo := repository.NewNotificationRepository(db)
	notificationSvc := service.NewNotificationService(notificationRepo, userRepo)
	probeSvc.SetNotificationService(notificationSvc)
	sshKeySvc.SetNotificationService(notificationSvc)
	approvalRepo := repository.NewApprovalRepository(db)
	approvalSvc := service.NewApprovalService(approvalRepo, userRepo, notificationSvc)
	permRepo := repository.NewPermissionRepository(db)
	permSvc := service.NewPermissionService(permRepo)
	_ = permSvc.SeedDefaults()
	auditRepo := repository.NewAuditLogRepository(db)
	quickActionRepo := repository.NewUserQuickActionRepository(db)
	dashboardSvc := service.NewDashboardService(assetRepo, auditRepo, termSessionRepo, projectRepo, quickActionRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc, captchaSvc, mfaSvc)
	userHandler := handler.NewUserHandler(userSvc)
	configHandler := handler.NewSystemConfigHandler(configSvc)
	auditHandler := handler.NewAuditLogHandler(auditSvc)
	cloudAccountHandler := handler.NewCloudAccountHandler(cloudAccountSvc)
	assetHandler := handler.NewAssetHandler(assetSvc)
	sshKeyHandler := handler.NewSSHKeyHandler(sshKeySvc)
	probeHandler := handler.NewProbeHandler(probeSvc)
	taskHandler := handler.NewTaskHandler(taskSvc, wsHub)
	fileDistHandler := handler.NewFileDistributionHandler(fileDistSvc, configSvc)
	linuxUserHandler := handler.NewLinuxUserHandler(linuxUserSvc)
	terminalHandler := handler.NewTerminalHandler(terminalSvc)
	ipamHandler := handler.NewIPAMHandler(ipamSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	playbookHandler := handler.NewPlaybookHandler(playbookSvc, wsHub)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)
	approvalHandler := handler.NewApprovalHandler(approvalSvc)
	permHandler := handler.NewPermissionHandler(permSvc)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	// API v1
	api := r.Group("/api/v1")
	api.Use(middleware.Audit(auditSvc))
	{
		// 公开接口
		auth := api.Group("/auth")
		{
			auth.GET("/captcha", authHandler.GetCaptcha)
			auth.POST("/login", authHandler.Login)
		}

		// 需认证接口
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// 用户信息 & MFA
			protected.GET("/user/info", authHandler.GetUserInfo)
			protected.GET("/user/mfa/setup", authHandler.SetupMFA)
			protected.POST("/user/mfa/bind", authHandler.BindMFA)
			protected.PUT("/user/password", userHandler.ChangePassword)

			// 用户管理
			users := protected.Group("/users")
			users.Use(middleware.RequireRole("admin", "operator"))
			{
				users.GET("", userHandler.List)
			}
			usersAdmin := protected.Group("/users")
			usersAdmin.Use(middleware.RequireRole("admin"))
			{
				usersAdmin.POST("", userHandler.Create)
				usersAdmin.PUT("/:id", userHandler.Update)
				usersAdmin.DELETE("/:id", userHandler.Delete)
				usersAdmin.PUT("/:id/password", userHandler.ResetPassword)
				usersAdmin.PUT("/:id/status", userHandler.ToggleStatus)
			}

			// 项目管理
			projects := protected.Group("/projects")
			{
				projects.GET("", projectHandler.List)
				projects.GET("/all", projectHandler.ListAll)
				projects.GET("/:id", projectHandler.GetByID)
				projects.GET("/:id/summary", projectHandler.GetSummary)
				projects.GET("/:id/assets", projectHandler.ListAssets)
			}
			projectsWrite := protected.Group("/projects")
			projectsWrite.Use(middleware.RequireRole("admin", "operator"))
			{
				projectsWrite.POST("", projectHandler.Create)
				projectsWrite.PUT("/:id", projectHandler.Update)
				projectsWrite.DELETE("/:id", projectHandler.Delete)
			}

			// 资产管理
			assets := protected.Group("/assets")
			{
				assets.GET("", assetHandler.List)
				assets.GET("/all", assetHandler.ListAll)
				assets.GET("/stats", assetHandler.Stats)
				assets.GET("/:id", assetHandler.GetByID)
				assets.GET("/:id/detail", assetHandler.GetDetail)
				assets.GET("/:id/probe", probeHandler.GetLatest)
				assets.GET("/:id/probe/history", probeHandler.ListHistory)
			}
			assetsWrite := protected.Group("/assets")
			assetsWrite.Use(middleware.RequireRole("admin", "operator"))
			{
				assetsWrite.POST("", assetHandler.Create)
				assetsWrite.POST("/import", assetHandler.BatchImport)
				assetsWrite.POST("/test-connection", assetHandler.TestConnection)
				assetsWrite.DELETE("/batch", assetHandler.BatchDelete)
				assetsWrite.PUT("/:id", assetHandler.Update)
				assetsWrite.DELETE("/:id", assetHandler.Delete)
			}

			// 云账号管理
			cloudAccounts := protected.Group("/cloud-accounts")
			cloudAccounts.Use(middleware.RequireRole("admin", "operator"))
			{
				cloudAccounts.GET("", cloudAccountHandler.List)
				cloudAccounts.GET("/:id", cloudAccountHandler.GetByID)
				cloudAccounts.POST("", cloudAccountHandler.Create)
				cloudAccounts.PUT("/:id", cloudAccountHandler.Update)
				cloudAccounts.DELETE("/:id", cloudAccountHandler.Delete)
			}

			// SSH Key 管理
			sshkeys := protected.Group("/sshkeys")
			sshkeys.Use(middleware.RequireRole("admin", "operator"))
			{
				sshkeys.GET("", sshKeyHandler.List)
				sshkeys.POST("", sshKeyHandler.Create)
				sshkeys.DELETE("/:id", sshKeyHandler.Delete)
				sshkeys.POST("/:id/deploy", sshKeyHandler.Deploy)
				sshkeys.POST("/:id/revoke", sshKeyHandler.Revoke)
				sshkeys.POST("/:id/rotate", sshKeyHandler.Rotate)
				sshkeys.GET("/:id/bindings", sshKeyHandler.ListBindings)
				sshkeys.GET("/:id/download", sshKeyHandler.Download)
				sshkeys.GET("/:id/deploy-logs", sshKeyHandler.ListDeployLogs)
				sshkeys.POST("/departure-cleanup", sshKeyHandler.DepartureCleanup)
			}

			// 资产探针
			probe := protected.Group("/probe")
			probe.Use(middleware.RequireRole("admin", "operator"))
			{
				probe.POST("/execute", probeHandler.Execute)
			}

			// 批量任务
			tasks := protected.Group("/tasks")
			tasks.Use(middleware.RequireRole("admin", "operator"))
			{
				tasks.POST("/execute", taskHandler.Execute)
				tasks.GET("", taskHandler.List)
				tasks.GET("/:id", taskHandler.GetByID)
				tasks.GET("/:id/results", taskHandler.GetResults)
			}

			// 文件分发
			files := protected.Group("/files")
			files.Use(middleware.RequireRole("admin", "operator"))
			{
				files.POST("/distribute", fileDistHandler.Distribute)
				files.GET("/tasks", fileDistHandler.List)
				files.GET("/tasks/:id/results", fileDistHandler.GetResults)
			}

			// Linux 用户管理
			linuxUsers := protected.Group("/linux-users")
			linuxUsers.Use(middleware.RequireRole("admin", "operator"))
			{
				linuxUsers.GET("", linuxUserHandler.List)
				linuxUsers.POST("", linuxUserHandler.Create)
				linuxUsers.DELETE("/:username", linuxUserHandler.Delete)
			}

			// 终端会话
			terminal := protected.Group("/terminal")
			{
				terminal.GET("/sessions", terminalHandler.ListSessions)
				terminal.GET("/sessions/active", terminalHandler.ListActiveSessions)
				terminal.GET("/sessions/:id/recording", terminalHandler.GetRecording)
				terminal.DELETE("/sessions/:id", terminalHandler.KillSession)
			}

			// 审计日志
			audit := protected.Group("/audit")
			audit.Use(middleware.RequireRole("admin", "operator"))
			{
				audit.GET("/logs", auditHandler.List)
				audit.GET("/logs/:id", auditHandler.GetByID)
			}

			// IPAM
			ipam := protected.Group("/ipam")
			ipam.Use(middleware.RequireRole("admin", "operator"))
			{
				ipam.GET("/subnets", ipamHandler.ListSubnets)
				ipam.POST("/subnets", ipamHandler.CreateSubnet)
				ipam.PUT("/subnets/:id", ipamHandler.UpdateSubnet)
				ipam.DELETE("/subnets/:id", ipamHandler.DeleteSubnet)
				ipam.GET("/subnets/:id/ips", ipamHandler.ListIPs)
				ipam.POST("/ips/:ipId/allocate", ipamHandler.AllocateIP)
				ipam.POST("/ips/:ipId/release", ipamHandler.ReleaseIP)
			}

			// 运维剧本
			playbooks := protected.Group("/playbooks")
			{
				playbooks.GET("", playbookHandler.List)
				playbooks.GET("/:id", playbookHandler.GetByID)
			}
			playbooksWrite := protected.Group("/playbooks")
			playbooksWrite.Use(middleware.RequireRole("admin", "operator"))
			{
				playbooksWrite.POST("", playbookHandler.Create)
				playbooksWrite.PUT("/:id", playbookHandler.Update)
				playbooksWrite.DELETE("/:id", playbookHandler.Delete)
				playbooksWrite.POST("/execute", playbookHandler.Execute)
				playbooksWrite.GET("/executions", playbookHandler.ListExecutions)
				playbooksWrite.GET("/executions/:id/results", playbookHandler.GetExecutionResults)
			}

			// 审批工单
			approvals := protected.Group("/approvals")
			{
				approvals.GET("", approvalHandler.List)
				approvals.GET("/mine", approvalHandler.ListMine)
				approvals.GET("/pending-count", approvalHandler.CountPending)
				approvals.GET("/:id", approvalHandler.GetByID)
				approvals.POST("", approvalHandler.Create)
				approvals.PUT("/:id/cancel", approvalHandler.Cancel)
			}
			approvalsAdmin := protected.Group("/approvals")
			approvalsAdmin.Use(middleware.RequireRole("admin"))
			{
				approvalsAdmin.PUT("/:id/review", approvalHandler.Review)
			}

			// 通知
			notifications := protected.Group("/notifications")
			{
				notifications.GET("", notificationHandler.List)
				notifications.GET("/unread-count", notificationHandler.UnreadCount)
				notifications.PUT("/:id/read", notificationHandler.MarkRead)
				notifications.PUT("/read-all", notificationHandler.MarkAllRead)
				notifications.DELETE("/:id", notificationHandler.Delete)
			}

			// 仪表盘
			dashboard := protected.Group("/dashboard")
			{
				dashboard.GET("/stats", dashboardHandler.GetStats)
				dashboard.GET("/activities", dashboardHandler.GetRecentActivities)
				dashboard.GET("/quick-actions", dashboardHandler.GetQuickActions)
				dashboard.PUT("/quick-actions", dashboardHandler.UpdateQuickActions)
			}

			// 权限管理
			permissions := protected.Group("/permissions")
			{
				permissions.GET("/mine", permHandler.GetMyPermissions)
			}
			permissionsAdmin := protected.Group("/permissions")
			permissionsAdmin.Use(middleware.RequireRole("admin"))
			{
				permissionsAdmin.GET("", permHandler.ListPermissions)
				permissionsAdmin.GET("/roles/:role", permHandler.GetRolePermissions)
				permissionsAdmin.PUT("/roles/:role", permHandler.SetRolePermissions)
			}

			// 系统设置
			settings := protected.Group("/settings")
			settings.Use(middleware.RequireRole("admin"))
			{
				settings.GET("/configs", configHandler.List)
				settings.PUT("/configs/:key", configHandler.Update)
				settings.PUT("/configs", configHandler.BatchUpdate)
			}
		}
	}

	// WebSocket 路由（JWT 通过 query param 认证）
	wsGroup := r.Group("/ws")
	wsGroup.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		wsGroup.GET("/tasks/:id", taskHandler.WebSocket)
		wsGroup.GET("/terminal/:assetId", terminalHandler.WebSocket)
		wsGroup.GET("/playbooks/:id", playbookHandler.WebSocket)
	}

	return r
}
