package migrations

import (
	"strconv"
	dbtype "viral-game-network/src/database/type"
)
var versionset = 0
var identifier = "10032025"

var migrations = []dbtype.Migration{
	// System Event 
	dbtype.Migration{
		SQL: `
		DEFINE TABLE System_Event SCHEMAFULL;
		DEFINE FIELD id ON TABLE System_Event TYPE string;
		DEFINE FIELD date_created ON TABLE System_Event TYPE string;
		DEFINE FIELD message ON TABLE System_Event TYPE string;
		DEFINE FIELD severity ON TABLE System_Event TYPE string;
		DEFINE INDEX idx_system_event_id ON TABLE System_Event COLUMNS id UNIQUE;
		DEFINE ANALYZER idx_system_event_analyzer TOKENIZERS class FILTERS ascii;
		DEFINE INDEX idx_system_event_message_analyzer ON TABLE Lobby COLUMNS message SEARCH ANALYZER idx_system_event_analyzer BM25 HIGHLIGHTS;
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
		DEFINE FIELD name ON TABLE Lobby TYPE string;
		DEFINE FIELD date_created ON TABLE Lobby TYPE string;
		DEFINE FIELD date_updated ON TABLE Lobby TYPE option<string>;
		DEFINE FIELD configs ON TABLE Lobby TYPE option<string>;
		DEFINE FIELD private ON TABLE Lobby TYPE bool DEFAULT false;
		DEFINE FIELD code ON TABLE Lobby TYPE option<string>;
		DEFINE INDEX idx_lobby_id ON TABLE Lobby COLUMNS id UNIQUE;
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
