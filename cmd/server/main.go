package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imkerbos/mxcmdb/internal/config"
	"github.com/imkerbos/mxcmdb/internal/model"
	"github.com/imkerbos/mxcmdb/internal/pkg/logger"
	"github.com/imkerbos/mxcmdb/internal/repository"
	"github.com/imkerbos/mxcmdb/internal/router"
	"github.com/imkerbos/mxcmdb/internal/service"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 设置时区（默认 Asia/Shanghai，可通过 TZ 环境变量覆盖）
	if os.Getenv("TZ") == "" {
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err == nil {
			time.Local = loc
		}
	}

	// 加载配置（日志尚未初始化，使用 stderr 输出）
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}
	cfg := config.Cfg

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Log.Sync() }()

	// 连接 PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		logger.Log.Fatalf("连接数据库失败: %v", err)
	}
	logger.Log.Info("数据库连接成功")

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Log.Fatalf("连接 Redis 失败: %v", err)
	}
	logger.Log.Info("Redis 连接成功")

	// 自动迁移
	if err := db.AutoMigrate(
		&model.User{},
		&model.SystemConfig{},
		&model.AuditLog{},
		&model.CloudAccount{},
		&model.Project{},
		&model.Asset{},
		&model.AssetTag{},
		&model.SSHKey{},
		&model.SSHKeyBinding{},
		&model.SSHKeyDeployLog{},
		&model.ProbeResult{},
		&model.Task{},
		&model.TaskResult{},
		&model.FileTask{},
		&model.FileTaskResult{},
		&model.LinuxUser{},
		&model.TerminalSession{},
		&model.Subnet{},
		&model.IPAddress{},
		&model.Playbook{},
		&model.PlaybookStep{},
		&model.PlaybookExecution{},
		&model.PlaybookExecutionResult{},
		&model.Notification{},
		&model.Approval{},
		&model.Permission{},
		&model.RolePermission{},
		&model.UserQuickAction{},
	); err != nil {
		logger.Log.Fatalf("数据库迁移失败: %v", err)
	}

	// 迁移存量资产状态：running/stopped → unknown（IDC 资产由探活机制自动更新状态）
	db.Model(&model.Asset{}).Where("source = ? AND status IN ?", "manual", []string{"running", "stopped"}).Update("status", "unknown")

	// 初始化管理员账号
	seedAdmin(db)

	// 初始化系统配置服务
	configRepo := repository.NewSystemConfigRepository(db)
	configSvc := service.NewSystemConfigService(configRepo)
	if err := configSvc.SeedDefaults(); err != nil {
		logger.Log.Errorf("初始化默认配置失败: %v", err)
	}
	if err := configSvc.LoadAll(); err != nil {
		logger.Log.Fatalf("加载系统配置失败: %v", err)
	}

	// 初始化内置剧本
	playbookRepo := repository.NewPlaybookRepository(db)
	playbookSvc := service.NewPlaybookService(playbookRepo, nil, "", configSvc, nil)
	if err := playbookSvc.SeedBuiltinPlaybooks(); err != nil {
		logger.Log.Errorf("初始化内置剧本失败: %v", err)
	}

	// 初始化审计日志服务
	auditRepo := repository.NewAuditLogRepository(db)
	auditSvc := service.NewAuditLogService(auditRepo)

	// 启动探活调度器
	assetRepo := repository.NewAssetRepository(db)
	heartbeatSvc := service.NewHeartbeatService(assetRepo, configSvc)
	heartbeatSvc.Start()

	// 初始化路由
	r := router.Setup(db, rdb, cfg, auditSvc, configSvc)

	// 启动服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		logger.Log.Infof("服务启动: http://localhost:%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("正在关闭服务...")

	// 关闭探活调度器
	heartbeatSvc.Stop()

	// 关闭审计日志服务，刷新缓冲区
	auditSvc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("服务关闭异常: %v", err)
	}
	logger.Log.Info("服务已停止")
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Errorf("生成密码哈希失败: %v", err)
		return
	}

	admin := model.User{
		Username: "admin",
		Password: string(hash),
		Nickname: "管理员",
		Role:     "admin",
		Status:   1,
	}
	if err := db.Create(&admin).Error; err != nil {
		logger.Log.Errorf("创建管理员失败: %v", err)
		return
	}
	logger.Log.Info("管理员账号已创建 (用户名: admin, 密码: admin123)")
}
