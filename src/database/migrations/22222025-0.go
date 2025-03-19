package migrations

import (
	"strconv"
	dbtype "viral-game-network/src/database/type"
)

func init() {
	versionset := 0
	identifier := "22222025-0"

	migrations := []dbtype.Migration{
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
		// Admins 
		dbtype.Migration{
			SQL: `
			DEFINE TABLE Admin SCHEMAFULL;
			DEFINE FIELD id ON TABLE Admin TYPE string;
			DEFINE FIELD name ON TABLE Admin TYPE string;
			DEFINE FIELD password ON TABLE Admin TYPE string;
			DEFINE FIELD email ON TABLE Admin TYPE string;
			DEFINE FIELD phone ON TABLE Admin TYPE string;
			DEFINE FIELD date_created ON TABLE Admin TYPE string;
			DEFINE FIELD date_updated ON TABLE Admin TYPE option<string>;
			DEFINE FIELD date_last_login ON TABLE Admin TYPE option<string>;
			DEFINE INDEX idx_user_id ON TABLE Admin COLUMNS id UNIQUE;
			DEFINE INDEX idx_admin_username ON TABLE Admin COLUMNS username UNIQUE;
			`,
		},
	}

	for _, migration := range migrations {
		versionset += 1
		migration.Name = identifier + "-" + strconv.Itoa(versionset)
		Migrations_Add(migration)
	}
}
