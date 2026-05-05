// Package config 加载 KleinAI 全局配置。
// 优先级：环境变量 > config.${KLEIN_ENV}.yaml > config.yaml。
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       App       `mapstructure:"app"`
	Server    Server    `mapstructure:"server"`
	MySQL     MySQL     `mapstructure:"mysql"`
	Redis     Redis     `mapstructure:"redis"`
	JWT       JWT       `mapstructure:"jwt"`
	Logger    Logger    `mapstructure:"logger"`
	Snowflake Snowflake `mapstructure:"snowflake"`
	CORS      CORS      `mapstructure:"cors"`
	RateLimit RateLimit `mapstructure:"ratelimit"`
	Pool      Pool      `mapstructure:"pool"`
	Provider  Provider  `mapstructure:"provider"`
	Billing   Billing   `mapstructure:"billing"`
	CDN       CDN       `mapstructure:"cdn"`
	AESKey    string    `mapstructure:"-"` // 来自环境变量
}

type App struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type Server struct {
	APIPort         int           `mapstructure:"api_port"`
	AdminPort       int           `mapstructure:"admin_port"`
	OpenAIPort      int           `mapstructure:"openai_port"`
	WSPort          int           `mapstructure:"ws_port"`
	PprofPort       int           `mapstructure:"pprof_port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type MySQL struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
}

type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWT struct {
	Secret        string        `mapstructure:"-"`
	RefreshSecret string        `mapstructure:"-"`
	AccessTTL     time.Duration `mapstructure:"access_ttl"`
	RefreshTTL    time.Duration `mapstructure:"refresh_ttl"`
}

type Logger struct {
	Level      string `mapstructure:"level"`
	Dir        string `mapstructure:"dir"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
	Console    bool   `mapstructure:"console"`
}

type Snowflake struct {
	NodeID int64 `mapstructure:"node_id"`
}

type CORS struct {
	Origins []string `mapstructure:"origins"`
}

type RateLimit struct {
	IPPerMinute     int `mapstructure:"ip_per_minute"`
	UserPerMinute   int `mapstructure:"user_per_minute"`
	APIKeyPerMinute int `mapstructure:"apikey_per_minute"`
}

type Pool struct {
	Strategy           string `mapstructure:"strategy"`
	CooldownSeconds    int    `mapstructure:"cooldown_seconds"`
	FailThreshold      int    `mapstructure:"fail_threshold"`
	HealthCheckSeconds int    `mapstructure:"health_check_seconds"`
}

type Provider struct {
	OpenAIBase     string        `mapstructure:"openai_base"`
	GrokBase       string        `mapstructure:"grok_base"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	Retry          int           `mapstructure:"retry"`
}

type Billing struct {
	PointUnit int64 `mapstructure:"point_unit"`
}

type CDN struct {
	Base string `mapstructure:"base"`
}

var (
	cfg     *Config
	once    sync.Once
	loadErr error
)

// Load 读取配置（线程安全，仅生效一次）。
func Load() (*Config, error) {
	once.Do(func() {
		cfg, loadErr = loadInternal()
	})
	return cfg, loadErr
}

// MustLoad 失败直接 panic（仅 cmd/* 入口使用）。
func MustLoad() *Config {
	c, err := Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}
	return c
}

// Get 返回已加载的配置；若未 Load 会 panic。
func Get() *Config {
	if cfg == nil {
		panic("config not loaded")
	}
	return cfg
}

func loadInternal() (*Config, error) {
	env := strings.TrimSpace(os.Getenv("KLEIN_ENV"))
	if env == "" {
		env = "dev"
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}

	v.SetConfigName("config." + env)
	if err := v.MergeInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !asErr(err, &notFound) {
			return nil, fmt.Errorf("merge env config: %w", err)
		}
	}

	v.SetEnvPrefix("KLEIN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	out := &Config{}
	if err := v.Unmarshal(out); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := applyEnvOverrides(out); err != nil {
		return nil, err
	}

	if env == "prod" {
		if err := validateProd(out); err != nil {
			return nil, err
		}
	}

	out.App.Env = env
	return out, nil
}

func validateProd(c *Config) error {
	if c.MySQL.DSN == "" {
		return fmt.Errorf("KLEIN_DB_DSN is required in prod")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("KLEIN_REDIS_ADDR is required in prod")
	}
	if len(c.JWT.Secret) < 32 || len(c.JWT.RefreshSecret) < 32 {
		return fmt.Errorf("KLEIN_JWT_SECRET / KLEIN_JWT_REFRESH_SECRET must be >= 32 bytes")
	}
	if len(c.AESKey) < 32 {
		return fmt.Errorf("KLEIN_AES_KEY must be >= 32 bytes")
	}
	return nil
}

func applyEnvOverrides(out *Config) error {
	mapString := func(target *string, key string) {
		if val, ok := os.LookupEnv(key); ok && val != "" {
			*target = val
		}
	}
	mapInt := func(target *int, key string) error {
		val, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(val) == "" {
			return nil
		}
		n, err := strconv.Atoi(strings.TrimSpace(val))
		if err != nil {
			return fmt.Errorf("parse %s: %w", key, err)
		}
		*target = n
		return nil
	}
	mapInt64 := func(target *int64, key string) error {
		val, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(val) == "" {
			return nil
		}
		n, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		if err != nil {
			return fmt.Errorf("parse %s: %w", key, err)
		}
		*target = n
		return nil
	}
	mapBool := func(target *bool, key string) error {
		val, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(val) == "" {
			return nil
		}
		b, err := strconv.ParseBool(strings.TrimSpace(val))
		if err != nil {
			return fmt.Errorf("parse %s: %w", key, err)
		}
		*target = b
		return nil
	}
	mapDuration := func(target *time.Duration, key string) error {
		val, ok := os.LookupEnv(key)
		if !ok || strings.TrimSpace(val) == "" {
			return nil
		}
		d, err := time.ParseDuration(strings.TrimSpace(val))
		if err != nil {
			return fmt.Errorf("parse %s: %w", key, err)
		}
		*target = d
		return nil
	}

	mapString(&out.App.Name, "KLEIN_APP_NAME")
	mapString(&out.MySQL.DSN, "KLEIN_DB_DSN")
	mapString(&out.Redis.Addr, "KLEIN_REDIS_ADDR")
	mapString(&out.Redis.Password, "KLEIN_REDIS_PASSWORD")
	mapString(&out.JWT.Secret, "KLEIN_JWT_SECRET")
	mapString(&out.JWT.RefreshSecret, "KLEIN_JWT_REFRESH_SECRET")
	mapString(&out.AESKey, "KLEIN_AES_KEY")
	mapString(&out.Provider.OpenAIBase, "KLEIN_OPENAI_BASE")
	mapString(&out.Provider.GrokBase, "KLEIN_GROK_BASE")
	mapString(&out.Logger.Dir, "KLEIN_LOG_DIR")
	mapString(&out.Logger.Level, "KLEIN_LOG_LEVEL")
	mapString(&out.CDN.Base, "KLEIN_CDN_BASE")

	for _, item := range []struct {
		target *int
		key    string
	}{
		{&out.Server.APIPort, "KLEIN_API_PORT"},
		{&out.Server.AdminPort, "KLEIN_ADMIN_PORT"},
		{&out.Server.OpenAIPort, "KLEIN_OPENAI_PORT"},
		{&out.Server.WSPort, "KLEIN_WS_PORT"},
		{&out.Server.PprofPort, "KLEIN_PPROF_PORT"},
		{&out.MySQL.MaxOpenConns, "KLEIN_MYSQL_MAX_OPEN_CONNS"},
		{&out.MySQL.MaxIdleConns, "KLEIN_MYSQL_MAX_IDLE_CONNS"},
		{&out.Redis.DB, "KLEIN_REDIS_DB"},
		{&out.Redis.PoolSize, "KLEIN_REDIS_POOL_SIZE"},
		{&out.Logger.MaxSizeMB, "KLEIN_LOG_MAX_SIZE_MB"},
		{&out.Logger.MaxAgeDays, "KLEIN_LOG_MAX_AGE_DAYS"},
		{&out.RateLimit.IPPerMinute, "KLEIN_RATELIMIT_IP_PER_MINUTE"},
		{&out.RateLimit.UserPerMinute, "KLEIN_RATELIMIT_USER_PER_MINUTE"},
		{&out.RateLimit.APIKeyPerMinute, "KLEIN_RATELIMIT_APIKEY_PER_MINUTE"},
		{&out.Pool.CooldownSeconds, "KLEIN_POOL_COOLDOWN_SECONDS"},
		{&out.Pool.FailThreshold, "KLEIN_POOL_FAIL_THRESHOLD"},
		{&out.Pool.HealthCheckSeconds, "KLEIN_POOL_HEALTH_CHECK_SECONDS"},
		{&out.Provider.Retry, "KLEIN_PROVIDER_RETRY"},
	} {
		if err := mapInt(item.target, item.key); err != nil {
			return err
		}
	}

	for _, item := range []struct {
		target *time.Duration
		key    string
	}{
		{&out.Server.ReadTimeout, "KLEIN_READ_TIMEOUT"},
		{&out.Server.WriteTimeout, "KLEIN_WRITE_TIMEOUT"},
		{&out.Server.ShutdownTimeout, "KLEIN_SHUTDOWN_TIMEOUT"},
		{&out.MySQL.ConnMaxLifetime, "KLEIN_MYSQL_CONN_MAX_LIFETIME"},
		{&out.MySQL.SlowThreshold, "KLEIN_MYSQL_SLOW_THRESHOLD"},
		{&out.JWT.AccessTTL, "KLEIN_JWT_ACCESS_TTL"},
		{&out.JWT.RefreshTTL, "KLEIN_JWT_REFRESH_TTL"},
		{&out.Provider.RequestTimeout, "KLEIN_PROVIDER_REQUEST_TIMEOUT"},
	} {
		if err := mapDuration(item.target, item.key); err != nil {
			return err
		}
	}

	if err := mapInt64(&out.Snowflake.NodeID, "KLEIN_NODE_ID"); err != nil {
		return err
	}
	if err := mapInt64(&out.Billing.PointUnit, "KLEIN_BILLING_POINT_UNIT"); err != nil {
		return err
	}
	if err := mapBool(&out.Logger.Compress, "KLEIN_LOG_COMPRESS"); err != nil {
		return err
	}
	if err := mapBool(&out.Logger.Console, "KLEIN_LOG_CONSOLE"); err != nil {
		return err
	}
	mapString(&out.Pool.Strategy, "KLEIN_POOL_STRATEGY")

	if origins := os.Getenv("KLEIN_CORS_ORIGINS"); origins != "" {
		out.CORS.Origins = splitAndTrim(origins, ",")
	}
	return nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// asErr 是 errors.As 的薄包装，避免 import cycle。
func asErr(err error, target any) bool {
	return errors.As(err, target)
}

// IsProd 判断是否生产环境。
func (c *Config) IsProd() bool { return c.App.Env == "prod" }

// IsDev 判断是否开发环境。
func (c *Config) IsDev() bool { return c.App.Env == "dev" || c.App.Env == "local" }
