package dbtype

import "github.com/fxamacker/cbor/v2"

// MarshalCBOR writes stored fields without the relationships projected by
// repository queries. JSON responses and CBOR query decoding retain them.
func (n Lobby) MarshalCBOR() ([]byte, error) {
	type storedLobby Lobby
	record := storedLobby(n)
	record.Lobby_Application = nil
	record.Lobby_Host = nil
	record.Lobby_Users = nil
	record.Lobby_Server = nil
	return cbor.Marshal(record)
}

func (n Server) MarshalCBOR() ([]byte, error) {
	type storedServer Server
	record := storedServer(n)
	record.Lobby = nil
	return cbor.Marshal(record)
}

// Settings presentation metadata is supplied by KeyValMap, not stored in the database.
func (n SystemConfig) MarshalCBOR() ([]byte, error) {
	type storedConfig SystemConfig
	record := storedConfig(n)
	record.Name = ""
	record.Options = nil
	record.ReadOnly = false
	record.SortOrder = 0
	return cbor.Marshal(record)
}
