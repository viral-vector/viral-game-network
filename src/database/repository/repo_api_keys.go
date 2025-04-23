package repository

import (
	"fmt"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

var cacheKey string = "system-configs-api-keys"

func DelApiKeyCache() error {
	cache.Del("system-configs-api-keys")
	return nil
}

func GetApiKeys() ([]dbtype.SystemConfig, error) {
	// Check if we have cache
	if cached, _ := cache.Get[[]dbtype.SystemConfig](cacheKey); cached != nil {
		return cached, nil
	}

	lQuery := `
	SELECT *
	FROM type::table(System_Config)
	WHERE string::starts_with(key, "API_KEY_")
	`
	// Get Mapping
	apiKeys, _ := database.Query[dbtype.SystemConfig](fmt.Sprintf("%s;", lQuery), 
	map[string]interface{}{
		
	})
	cache.Set[[]dbtype.SystemConfig](cacheKey, apiKeys, time.Minute*30)
	
	return apiKeys, nil
}

func AddApiKey(name string, key string) error {
	if name == "" {
		return fmt.Errorf("Name name must be set")
	}
	// Gen key
	if key == "" {
		return fmt.Errorf("Key name must be set")
	}
	datetime := time.Now().UTC().Format(time.RFC3339)
	_, err := database.Create[dbtype.SystemConfig](&dbtype.SystemConfig{
		Key   		 : "API_KEY_"+name,
		Val   		 : key,
		Type  		 : "text",
		Version   	 : "0",
		Date_Created : datetime,
		Date_Updated : datetime,
	})

	return err
}

func GetKeyValue(key string) *string {
	keys, _ := GetApiKeys() 
	if keys != nil {
		for _, key := range keys {
			fmt.Println(key)
		}
	}
	return nil
}