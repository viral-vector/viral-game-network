package migrations

import (
	"strconv"
	dbtype "viral-game-network/src/database/type"
)
var versionset = 0
var identifier = "22222025"

var migrations = []dbtype.Migration{
	// System Config 
	dbtype.Migration{
		SQL: `
		DEFINE TABLE System_Config SCHEMAFULL;
		DEFINE FIELD id ON TABLE System_Config TYPE string;
		DEFINE FIELD key ON TABLE System_Config TYPE string;
		DEFINE FIELD val ON TABLE System_Config TYPE string;
		DEFINE FIELD type ON TABLE System_Config TYPE string;
		DEFINE FIELD version ON TABLE System_Config TYPE string;
		DEFINE FIELD date_created ON TABLE System_Config TYPE string;
		DEFINE FIELD date_updated ON TABLE System_Config TYPE string;
		DEFINE INDEX idx_system_config_id ON TABLE System_Config COLUMNS id UNIQUE;
		DEFINE INDEX idx_system_config_key ON TABLE System_Config COLUMNS key UNIQUE;
		`,
	},
	// System Event 
	dbtype.Migration{
		SQL: `
		DEFINE TABLE System_Event SCHEMAFULL;
		DEFINE FIELD id ON TABLE System_Event TYPE string;
		DEFINE FIELD ref_id ON TABLE System_Event TYPE option<string>;
		DEFINE FIELD ref_source ON TABLE System_Event TYPE option<string>;
		DEFINE FIELD ref_target ON TABLE System_Event TYPE option<string>;
		DEFINE FIELD date_created ON TABLE System_Event TYPE string;
		DEFINE FIELD message ON TABLE System_Event TYPE string;
		DEFINE FIELD severity ON TABLE System_Event TYPE string;
		DEFINE INDEX idx_system_event_id ON TABLE System_Event COLUMNS id UNIQUE;
		DEFINE INDEX idx_system_event_ref_id ON TABLE System_Event COLUMNS ref_id UNIQUE;
		DEFINE INDEX idx_system_event_ref_source ON TABLE System_Event COLUMNS ref_source;
		DEFINE ANALYZER idx_system_event_analyzer TOKENIZERS class FILTERS ascii;
		DEFINE INDEX idx_system_event_message_analyzer ON TABLE System_Event COLUMNS message SEARCH ANALYZER idx_system_event_analyzer BM25 HIGHLIGHTS;
		`,
	},
	// Application 
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Application SCHEMAFULL;
		DEFINE FIELD id ON TABLE Application TYPE string;
		DEFINE FIELD guid ON TABLE Application TYPE string;
		DEFINE FIELD name ON TABLE Application TYPE string;
		DEFINE FIELD date_created ON TABLE Application TYPE string;
		DEFINE FIELD date_updated ON TABLE Application TYPE option<string>;
		DEFINE FIELD group ON TABLE Application TYPE option<string>;
		DEFINE FIELD image ON TABLE Application TYPE string;
		DEFINE FIELD port ON TABLE Application TYPE string;
		DEFINE FIELD version ON TABLE Application TYPE string;
		DEFINE FIELD command ON TABLE Application TYPE string;
		DEFINE FIELD lobby_max_players ON TABLE Application TYPE string;
		DEFINE FIELD lobby_max_persist ON TABLE Application TYPE string;
		DEFINE INDEX idx_application_id ON TABLE Application COLUMNS id UNIQUE;
		DEFINE INDEX idx_application_guid ON TABLE Application COLUMNS guid UNIQUE;
		DEFINE INDEX idx_application_name ON TABLE Application COLUMNS name;
		DEFINE INDEX idx_application_group ON TABLE Application COLUMNS group;
		`,
	},
	// User 
	dbtype.Migration{
		SQL: `
		DEFINE TABLE User SCHEMAFULL;
		DEFINE FIELD id ON TABLE User TYPE string;
		DEFINE FIELD guid ON TABLE User TYPE string;
		DEFINE FIELD name ON TABLE User TYPE string;
		DEFINE FIELD date_created ON TABLE User TYPE string;
		DEFINE FIELD date_updated ON TABLE User TYPE option<string>;
		DEFINE FIELD date_last_login ON TABLE User TYPE option<string>;
		DEFINE FIELD configs ON TABLE User TYPE option<string>;
		DEFINE INDEX idx_user_id ON TABLE User COLUMNS id UNIQUE;
		DEFINE INDEX idx_user_guid ON TABLE User COLUMNS guid UNIQUE;
		DEFINE INDEX idx_user_name ON TABLE User COLUMNS name UNIQUE;
		`,
	},
	// Lobby 
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Lobby SCHEMAFULL;
		DEFINE FIELD id ON TABLE Lobby TYPE string;
		DEFINE FIELD guid ON TABLE Lobby TYPE string;
		DEFINE FIELD name ON TABLE Lobby TYPE string;
		DEFINE FIELD date_created ON TABLE Lobby TYPE string;
		DEFINE FIELD date_updated ON TABLE Lobby TYPE option<string>;
		DEFINE FIELD configs ON TABLE Lobby TYPE option<string>;
		DEFINE FIELD private ON TABLE Lobby TYPE bool DEFAULT false;
		DEFINE FIELD code ON TABLE Lobby TYPE option<string>;
		DEFINE INDEX idx_lobby_id ON TABLE Lobby COLUMNS id UNIQUE;
		DEFINE INDEX idx_lobby_guid ON TABLE Lobby COLUMNS guid UNIQUE;
		DEFINE INDEX idx_lobby_name ON TABLE Lobby COLUMNS name;
		DEFINE ANALYZER idx_lobby_analyzer TOKENIZERS class FILTERS ascii;
		DEFINE INDEX idx_lobby_name_analyzer ON TABLE Lobby COLUMNS name SEARCH ANALYZER idx_lobby_analyzer BM25 HIGHLIGHTS;

		`,
	},
	// Server
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Server SCHEMAFULL;
		DEFINE FIELD id ON TABLE Server TYPE string;
		DEFINE FIELD guid ON TABLE Server TYPE string;
		DEFINE FIELD name ON TABLE Server TYPE string;
		DEFINE FIELD status ON TABLE Server TYPE string;
		DEFINE FIELD date_created ON TABLE Server TYPE string;
		DEFINE FIELD date_updated ON TABLE Server TYPE option<string>;
		DEFINE FIELD address ON TABLE Server TYPE option<string>;
		DEFINE FIELD port ON TABLE Server TYPE option<number>;
		DEFINE FIELD configs ON TABLE Server TYPE option<string>;
		DEFINE INDEX idx_server_id ON TABLE Server COLUMNS id UNIQUE;
		DEFINE INDEX idx_server_guid ON TABLE Server COLUMNS guid UNIQUE;
		`,
	},
	// Lobby_Application
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Lobby_Application SCHEMAFULL TYPE RELATION IN Lobby OUT Application ENFORCED ;
		DEFINE INDEX idx_lobby_application_in ON TABLE Lobby_Application COLUMNS in UNIQUE;
		DEFINE INDEX idx_lobby_application_out ON TABLE Lobby_Application COLUMNS out UNIQUE;
		`,
	},
	// Lobby_Host
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Lobby_Host SCHEMAFULL TYPE RELATION IN Lobby OUT User ENFORCED ;
		DEFINE INDEX idx_lobby_host_in ON TABLE Lobby_Host COLUMNS in UNIQUE;
		DEFINE INDEX idx_lobby_host_out ON TABLE Lobby_Host COLUMNS out UNIQUE;
		`,
	},
	// Lobby_Server
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Lobby_Server SCHEMAFULL TYPE RELATION IN Lobby OUT Server ENFORCED;
		DEFINE INDEX idx_lobby_server_in ON TABLE Lobby_Server COLUMNS in UNIQUE;
		DEFINE INDEX idx_lobby_server_out ON TABLE Lobby_Server COLUMNS out UNIQUE;
		`,
	},
	// Lobby_Users
	dbtype.Migration{
		SQL: `
		DEFINE TABLE Lobby_Users SCHEMAFULL TYPE RELATION IN Lobby OUT User ENFORCED;
		DEFINE INDEX idx_lobby_users_host_in ON TABLE Lobby_Users COLUMNS in UNIQUE;
		DEFINE INDEX idx_lobby_users_host_out ON TABLE Lobby_Users COLUMNS out UNIQUE;
		`,
	},
}

func init() {
	for _, migration := range migrations {
		versionset += 1
		migration.Name = identifier + "-" + strconv.Itoa(versionset)
		Migrations_Add(migration)
	}
}
