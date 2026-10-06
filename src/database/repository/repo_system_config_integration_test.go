//go:build integration

package repository

import (
	"testing"
	"viral-game-network/tests/support"
)

func TestSavingNewSystemSettingsPreservesAPIKeysAndOtherSettings(t *testing.T) {
	support.Storage(t)
	if err := AddApiKey("game", "game-api-key"); err != nil {
		t.Fatal(err)
	}
	configs, err := GetSystemConfigs()
	if err != nil {
		t.Fatal(err)
	}
	for i := range configs {
		switch configs[i].Key {
		case "VNET_NAME":
			configs[i].Val = "Test Network"
		case "VNET_TOKEN_EXPIRE":
			configs[i].Val = "45"
		}
	}
	if _, err := PopSystemConfigs(&configs); err != nil {
		t.Fatal(err)
	}
	DelApiKeyCache()
	keys, err := GetApiKeys()
	if err != nil || len(keys) != 1 || keys[0].Val != "game-api-key" {
		t.Fatalf("saving settings overwrote API keys: %+v %v", keys, err)
	}
	for _, want := range configs {
		actual, err := SelConfig(want.Key)
		if err != nil || actual.Val != want.Val {
			t.Fatalf("saving a setting overwrote another setting %s: %+v %v", want.Key, actual, err)
		}
	}
	for i := range configs {
		if configs[i].Key == "VNET_NAME" {
			configs[i].Val = "Renamed Network"
		}
	}
	if _, err := PopSystemConfigs(&configs); err != nil {
		t.Fatal("repeat save failed", err)
	}
	keys, err = GetApiKeys()
	if err != nil || len(keys) != 1 {
		t.Fatalf("repeat save lost API key: %+v %v", keys, err)
	}
}
