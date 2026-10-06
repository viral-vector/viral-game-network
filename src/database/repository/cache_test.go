package repository

import (
	"testing"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/tests/support"
)

func TestCachedConfigurationAndInvalidation(t *testing.T) {
	server := support.Redis(t)
	cache.Set("system-configs", []dbtype.SystemConfig{{Key: "VNET_NAME", Val: "test"}}, time.Hour)
	configs, err := GetSystemConfigs()
	if err != nil || len(configs) != 1 || configs[0].Val != "test" {
		t.Fatalf("%v %v", configs, err)
	}
	if value := GetConfigValue("VNET_NAME"); value == nil || *value != "test" {
		t.Fatal(value)
	}
	if value := GetConfigValue("unknown"); value != nil {
		t.Fatal(value)
	}
	if err := DelSystemConfigsCache(); err != nil || server.Exists("system-configs") {
		t.Fatal("config cache not invalidated", err)
	}
	cache.Set(cacheKey, []dbtype.SystemConfig{{Key: "API_KEY_test", Val: "allowed"}}, time.Hour)
	keys, err := GetApiKeys()
	if err != nil || len(keys) != 1 || keys[0].Val != "allowed" {
		t.Fatalf("%v %v", keys, err)
	}
	if err := DelApiKeyCache(); err != nil || server.Exists(cacheKey) {
		t.Fatal("key cache not invalidated", err)
	}
}
