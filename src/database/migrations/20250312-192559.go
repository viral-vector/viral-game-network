package migrations

import (
	dbtype "viral-game-network/src/database/type"
)

func init() {
	// Append the migration to the AllMigrations slice.
	AllMigrations = append(AllMigrations, dbtype.Migration{
		Name:  "20250312-192559-1",
		SQL: `
		DEFINE TABLE System_Event SCHEMAFULL;
		DEFINE FIELD id ON TABLE System_Event TYPE string;
		DEFINE FIELD date_created ON TABLE System_Event TYPE string;
		DEFINE FIELD message ON TABLE System_Event TYPE string;
		DEFINE FIELD severity ON TABLE System_Event TYPE string;
		DEFINE INDEX idx_system_event_id ON TABLE System_Event COLUMNS id UNIQUE;
		`,
	})

	AllMigrations = append(AllMigrations, dbtype.Migration{
		Name:  "20250312-192559-2",
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
	})

	AllMigrations = append(AllMigrations, dbtype.Migration{
		Name:  "20250312-192559-3",
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
		`,
	})

	AllMigrations = append(AllMigrations, dbtype.Migration{
		Name:  "20250312-192559-4",
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
		DEFINE INDEX idx_server_guid ON TABLE Server COLUMNS UNIQUE;
		`,
	})
}
