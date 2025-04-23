package repository

import (
	"log"
	"fmt"
	"time"
	"sort"
	"viral-game-network/src/cache"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func DelSystemConfigsCache() error {
	cache.Del("system-configs")
	return nil
}

func GetSystemConfigs() ([]dbtype.SystemConfig, error) {
	// Check if we have cache
	if cached, _ := cache.Get[[]dbtype.SystemConfig]("system-configs"); cached != nil {
		return cached, nil
	}

	lQuery := `
	SELECT *
	FROM type::table(System_Config);
	`
	// Get Mapping
	configMap := dbtype.SystemConfig{}.KeyValMap()
	mappedMap, _ := database.Query[dbtype.SystemConfig](fmt.Sprintf("%s%s", lQuery, `;`), 
	map[string]interface{}{
		
	})

	// GLobal datetime
	datetime := time.Now().UTC().Format(time.RFC3339)

	var finalMap []dbtype.SystemConfig
	for _, config := range configMap {
		// create dummy
		temp := config
		temp.Date_Created = datetime
		// check for existing
		for _, item := range mappedMap {
			if item.Key == config.Key {
				temp.Date_Created = item.Date_Created
				temp.Date_Updated = item.Date_Updated
				temp.ID = item.ID
				temp.Val = item.Val
				temp.Version = item.Version
				break
			}
		}
		finalMap = append(finalMap, temp)
	}

	sort.Slice(finalMap, func(i, j int) bool {
		return finalMap[i].Key < finalMap[j].Key
	})
	
	cache.Set[[]dbtype.SystemConfig]("system-configs", finalMap, time.Minute*30)
	
	return finalMap, nil
}

func PopSystemConfigs(configs *[]dbtype.SystemConfig) ([]dbtype.SystemConfig, error) {	
	datetime := time.Now().UTC().Format(time.RFC3339)

	var err error
	for _, config := range *configs {
		config.Date_Updated = datetime
		if config.ID != nil {
			_, err = database.Update[dbtype.SystemConfig](*config.ID, &config)
		}else {
			_, err = database.Upsert[dbtype.SystemConfig](&config)
		}
		if err != nil{
			log.Panicf("Repository: PopSystemConfigs Error: %s", err)
		}
	}

	cache.Del("system-configs")

	return GetSystemConfigs()
}

func GetConfigValue(key string) *string {
	configs, _ := GetSystemConfigs() 
	if configs != nil {
		for _, config := range configs {
			if config.Key == key {
				return &config.Val
			}
		}
	}
	return nil
}

func GetConfig(id string) (*dbtype.SystemConfig, error) {
	// Get config.
	configs, err := database.Query[dbtype.SystemConfig](
		`SELECT * FROM type::record($id);`,
		map[string]interface{}{
			"id": id,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("Config not found")
	}

	return &configs[0], nil
}

func SelConfig(key string) (*dbtype.SystemConfig, error) {
	// Sel config.
	configs, err := database.Query[dbtype.SystemConfig](
		`SELECT * FROM type::table(System_Config) WHERE key=$key;`,
		map[string]interface{}{
			"key": key,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, fmt.Errorf("Config not found")
	}

	return &configs[0], nil
}

func DelConfig(id string, config *dbtype.SystemConfig) error {
	return database.Delete(*config.ID)
}