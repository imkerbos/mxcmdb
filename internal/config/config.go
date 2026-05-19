package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config 应用配置（仅基础设施配置，业务配置存数据库）
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Captcha    CaptchaConfig    `mapstructure:"captcha"`
	MFA        MFAConfig        `mapstructure:"mfa"`
	Log        LogConfig        `mapstructure:"log"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	MasterKey string `mapstructure:"master_key"`
}

// ServerConfig 服务配置
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessExpire  int64  `mapstructure:"access_expire"`
	RefreshExpire int64  `mapstructure:"refresh_expire"`
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Length  int  `mapstructure:"length"`
	Expire  int  `mapstructure:"expire"` // seconds
}

// MFAConfig MFA 配置
type MFAConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Issuer  string `mapstructure:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DSN 返回 PostgreSQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// Cfg 全局配置实例
var Cfg *Config

// Init 加载配置文件
func Init() error {
	path := "configs/config.yaml"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		path = p
	}

	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	// 环境变量覆盖（部署时通过 .env 注入敏感配置，优先级高于 YAML）
	bindEnvOverrides()

	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}
	return nil
}

// bindEnvOverrides 绑定环境变量，支持通过 .env 覆盖 YAML 中的配置
func bindEnvOverrides() {
	overrides := map[string]string{
		"server.port":          "MXCMDB_SERVER_PORT",
		"server.mode":          "MXCMDB_SERVER_MODE",
		"database.host":        "MXCMDB_DB_HOST",
		"database.port":        "MXCMDB_DB_PORT",
		"database.user":        "MXCMDB_DB_USER",
		"database.password":    "MXCMDB_DB_PASSWORD",
		"database.dbname":      "MXCMDB_DB_NAME",
		"database.sslmode":     "MXCMDB_DB_SSLMODE",
		"redis.host":           "MXCMDB_REDIS_HOST",
		"redis.port":           "MXCMDB_REDIS_PORT",
		"redis.password":       "MXCMDB_REDIS_PASSWORD",
		"redis.db":             "MXCMDB_REDIS_DB",
		"jwt.secret":           "MXCMDB_JWT_SECRET",
		"jwt.access_expire":    "MXCMDB_JWT_ACCESS_EXPIRE",
		"jwt.refresh_expire":   "MXCMDB_JWT_REFRESH_EXPIRE",
		"log.level":            "MXCMDB_LOG_LEVEL",
		"log.format":           "MXCMDB_LOG_FORMAT",
		"encryption.master_key": "MXCMDB_ENCRYPTION_KEY",
	}
	for key, env := range overrides {
		_ = viper.BindEnv(key, env)
	}
}
