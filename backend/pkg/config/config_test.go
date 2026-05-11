package config

import (
	"testing"
	"time"
)

func TestLoadInternalAppliesExplicitEnvOverrides(t *testing.T) {
	t.Setenv("IMAGE2API_ENV", "dev")
	t.Setenv("IMAGE2API_API_PORT", "18180")
	t.Setenv("IMAGE2API_ADMIN_PORT", "18188")
	t.Setenv("IMAGE2API_OPENAI_PORT", "18200")
	t.Setenv("IMAGE2API_WS_PORT", "18280")
	t.Setenv("IMAGE2API_PPROF_PORT", "18590")
	t.Setenv("IMAGE2API_REDIS_DB", "7")
	t.Setenv("IMAGE2API_REDIS_POOL_SIZE", "12")
	t.Setenv("IMAGE2API_JWT_ACCESS_TTL", "45m")
	t.Setenv("IMAGE2API_JWT_REFRESH_TTL", "720h")
	t.Setenv("IMAGE2API_NODE_ID", "4")
	t.Setenv("IMAGE2API_LOG_MAX_SIZE_MB", "64")
	t.Setenv("IMAGE2API_LOG_MAX_AGE_DAYS", "9")
	t.Setenv("IMAGE2API_LOG_COMPRESS", "false")
	t.Setenv("IMAGE2API_CORS_ORIGINS", "http://a.test, http://b.test")

	cfg, err := loadInternal()
	if err != nil {
		t.Fatalf("loadInternal() error = %v", err)
	}

	if cfg.Server.APIPort != 18180 || cfg.Server.AdminPort != 18188 || cfg.Server.OpenAIPort != 18200 || cfg.Server.WSPort != 18280 || cfg.Server.PprofPort != 18590 {
		t.Fatalf("server ports not overridden: %+v", cfg.Server)
	}
	if cfg.Redis.DB != 7 || cfg.Redis.PoolSize != 12 {
		t.Fatalf("redis overrides not applied: %+v", cfg.Redis)
	}
	if cfg.JWT.AccessTTL != 45*time.Minute || cfg.JWT.RefreshTTL != 720*time.Hour {
		t.Fatalf("jwt ttl overrides not applied: %+v", cfg.JWT)
	}
	if cfg.Snowflake.NodeID != 4 {
		t.Fatalf("node id not overridden: %d", cfg.Snowflake.NodeID)
	}
	if cfg.Logger.MaxSizeMB != 64 || cfg.Logger.MaxAgeDays != 9 || cfg.Logger.Compress {
		t.Fatalf("logger overrides not applied: %+v", cfg.Logger)
	}
	if len(cfg.CORS.Origins) != 2 || cfg.CORS.Origins[0] != "http://a.test" || cfg.CORS.Origins[1] != "http://b.test" {
		t.Fatalf("cors origins not split: %#v", cfg.CORS.Origins)
	}
}

func TestLoadInternalRejectsInvalidEnvOverride(t *testing.T) {
	t.Setenv("IMAGE2API_ENV", "dev")
	t.Setenv("IMAGE2API_API_PORT", "not-a-port")

	if _, err := loadInternal(); err == nil {
		t.Fatal("loadInternal() error = nil, want invalid port error")
	}
}
