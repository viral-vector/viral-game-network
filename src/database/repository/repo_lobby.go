package repository

import (
	"context"
	"fmt"
	"strings"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database"
	dbtype "viral-game-network/src/database/type"
)

func AllLobby(count int, pager int, search string) ([]dbtype.Lobby, int, error) {
	lQuery := `
	SELECT *
	,array::first(SELECT * FROM ->Lobby_Application.out) AS lobby_application
	,(IF count(SELECT id FROM ->Lobby_Users.out) > 0
		{array::first(SELECT out.* FROM ->Lobby_Users WHERE user_type = 'host')} ELSE {NULL}).out AS lobby_host
	,(IF count(SELECT id FROM ->Lobby_Server.out) > 0
		{array::first(SELECT * FROM ->Lobby_Server.out)} ELSE {NULL}) AS lobby_server
	,(SELECT * FROM ->Lobby_Users.out) AS lobby_users
	FROM type::table(Lobby)
	`
	params := map[string]interface{}{}

	// Search
	if search != "" {
		lQuery = fmt.Sprintf("%s WHERE [name, guid] ?~ $search", lQuery)
		params["search"] = fmt.Sprintf("%s", search)
	}

	// Ordering
	lQuery = fmt.Sprintf("%s ORDER BY date_created DESC", lQuery)

	// Limit and Pagination
	if count > -1 {
		lQuery = fmt.Sprintf("%s LIMIT $ct START $pg", lQuery)
		params["ct"] = count
		params["pg"] = (pager - 1) * count
	}

	// Get all lobbies.
	lobbies, err := database.Query[dbtype.Lobby](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all lobbies.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table(Lobby) GROUP ALL;",
		map[string]interface{}{},
	)
	if err != nil || len(total) == 0 {
		return lobbies, 0, err
	}

	return lobbies, total[0].Total, nil
}

func AllLobbyNotRunning(count int, pager int) ([]dbtype.Lobby, int, error) {
	lQuery := `
	SELECT * FROM (
		SELECT *
		,array::first(SELECT * FROM ->Lobby_Application.out) AS lobby_application
		,(IF count(SELECT id FROM ->Lobby_Server.out) > 0
			{array::first(SELECT id, name, guid, status FROM ->Lobby_Server.out)} ELSE {NULL}) AS lobby_server
		FROM type::table(Lobby)
	)
	WHERE lobby_server.status NOTINSIDE ['Running', 'Online']
	ORDER BY date_created DESC
	`
	params := map[string]interface{}{}

	if count > -1 {
		lQuery = fmt.Sprintf("%s LIMIT $ct START $pg", lQuery)
		params["ct"] = count
		params["pg"] = (pager - 1) * count
	}

	// Get all lobbies.
	lobbies, err := database.Query[dbtype.Lobby](fmt.Sprintf("%s;", lQuery), params)
	if err != nil {
		return nil, 0, err
	}

	// Count all lobbies.
	total, err := database.Query[dbtype.Total](
		"SELECT count() AS total FROM type::table(Lobby) GROUP ALL;",
		map[string]interface{}{},
	)
	if err != nil || len(total) == 0 {
		return lobbies, 0, err
	}

	return lobbies, total[0].Total, nil
}

func GetLobby(id string) (*dbtype.Lobby, error) {
	// Get lobby by ID.
	lobbies, err := database.Query[dbtype.Lobby](`
	SELECT *
	,array::first(SELECT * FROM ->Lobby_Application.out) AS lobby_application
	,(IF count(SELECT id FROM ->Lobby_Users.out) > 0
		{array::first(SELECT out.* FROM ->Lobby_Users WHERE user_type = 'host')} ELSE {NULL}).out AS lobby_host
	,(IF count(SELECT id FROM ->Lobby_Server.out) > 0
		{array::first(SELECT * FROM ->Lobby_Server.out)} ELSE {NULL}) AS lobby_server
	,(SELECT * FROM ->Lobby_Users.out) AS lobby_users
	FROM type::record($id)
	ORDER BY date_created DESC;`,
		map[string]interface{}{
			"id": id,
		})
	if err != nil {
		return nil, err
	}
	if len(lobbies) == 0 {
		return nil, nil
	}
	return &lobbies[0], nil
}

// IsLobbyMember checks the indexed membership edge without loading the lobby's
// full roster and relationships for every live chat message.
func IsLobbyMember(lobby, user string) (bool, error) {
	if !strings.HasPrefix(lobby, "Lobby:") || !strings.HasPrefix(user, "User:") {
		return false, fmt.Errorf("invalid lobby or user ID")
	}
	rows, err := database.Query[dbtype.Total](`
		SELECT count() AS total FROM Lobby_Users
		WHERE in=type::record($lobby) AND out=type::record($user) GROUP ALL;
	`, map[string]interface{}{"lobby": lobby, "user": user})
	if err != nil {
		return false, err
	}
	return len(rows) > 0 && rows[0].Total > 0, nil
}

// LobbyPatch includes editable fields only. Pointers distinguish omitted values
// from explicit false/empty values in JSON and admin form submissions.
type LobbyPatch struct {
	Name              *string             `json:"name" form:"name"`
	Private           *bool               `json:"private" form:"private"`
	Code              *string             `json:"code" form:"code"`
	Lobby_Application *dbtype.Application `json:"lobby_application" form:"lobby_application"`
}

func PatchLobby(id string, patch *LobbyPatch) (*dbtype.Lobby, error) {
	if !strings.HasPrefix(id, "Lobby:") || len(id) <= len("Lobby:") || patch == nil {
		return nil, fmt.Errorf("invalid lobby edit")
	}
	var updated *dbtype.Lobby
	held, err := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+strings.TrimPrefix(id, "Lobby:"), 60*time.Second, func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		updated, err = patchLobbyTransaction(id, patch)
		return err
	})
	if err != nil {
		return nil, err
	}
	if !held {
		return nil, fmt.Errorf("lobby busy; retry editing")
	}
	return updated, nil
}

func patchLobbyTransaction(id string, patch *LobbyPatch) (*dbtype.Lobby, error) {
	current, err := GetLobby(id)
	if err != nil {
		return nil, err
	}
	if current == nil || patch == nil {
		return nil, fmt.Errorf("lobby not found")
	}
	fields := map[string]interface{}{"date_updated": time.Now().UTC().Format(time.RFC3339)}
	if patch.Name != nil {
		name := strings.TrimSpace(*patch.Name)
		if name == "" || len(name) > 128 {
			return nil, fmt.Errorf("invalid lobby name")
		}
		fields["name"] = name
	}
	if patch.Private != nil {
		current.Private = *patch.Private
		fields["private"] = *patch.Private
	}
	if patch.Code != nil {
		current.Code = *patch.Code
		fields["code"] = *patch.Code
	}
	if current.Private && current.Code == "" {
		return nil, fmt.Errorf("private lobby requires a code")
	}
	query := `BEGIN TRANSACTION; UPDATE type::record($lobby) MERGE $fields;`
	params := map[string]interface{}{"lobby": id, "fields": fields}
	if patch.Lobby_Application != nil {
		if patch.Lobby_Application.ID == nil || patch.Lobby_Application.ID.Table != "Application" {
			return nil, fmt.Errorf("missing application")
		}
		app, err := GetApplication(patch.Lobby_Application.ModelID())
		if err != nil {
			return nil, err
		}
		if app == nil || app.ID == nil {
			return nil, fmt.Errorf("application not found")
		}
		if current.Lobby_Server != nil && (current.Lobby_Application == nil || app.ModelID() != current.Lobby_Application.ModelID()) {
			return nil, fmt.Errorf("stop the lobby server before changing application")
		}
		params["application"] = app.ModelID()
		query += ` DELETE FROM Lobby_Application WHERE in=type::record($lobby);
   RELATE (type::record($lobby))->Lobby_Application->(type::record($application));`
	}
	query += ` COMMIT TRANSACTION;`
	if _, err := database.Query[any](query, params); err != nil {
		return nil, err
	}
	return GetLobby(id)
}

// JoinLobby checks capacity and the invitation inside the same transaction as
// membership replacement. Existing members keep their role when retrying a join.
func JoinLobby(id string, user *dbtype.User, code string) (*dbtype.Lobby, error) {
	if !strings.HasPrefix(id, "Lobby:") || len(id) <= len("Lobby:") {
		return nil, fmt.Errorf("invalid lobby ID")
	}
	var joined *dbtype.Lobby
	held, err := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+strings.TrimPrefix(id, "Lobby:"), 60*time.Second, func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		var err error
		joined, err = joinLobbyTransaction(id, user, code)
		return err
	})
	if err != nil {
		return nil, err
	}
	if !held {
		return nil, fmt.Errorf("lobby busy; retry joining")
	}
	return joined, nil
}

func joinLobbyTransaction(id string, user *dbtype.User, code string) (*dbtype.Lobby, error) {
	if user == nil || user.ID == nil {
		return nil, fmt.Errorf("missing user")
	}
	_, err := database.Query[any](`
 BEGIN TRANSACTION;
 LET $room = SELECT * FROM ONLY type::record($lobby);
 IF $room = NONE { THROW 'Lobby not found'; };
 LET $members = SELECT * FROM Lobby_Users WHERE in=type::record($lobby);
 IF array::len($members[WHERE out=type::record($user)]) = 0 {
  IF $room.private AND $room.code != $code { THROW 'Invalid lobby code'; };
  LET $app = array::first(SELECT * FROM (type::record($lobby))->Lobby_Application.out);
  IF $app = NONE { THROW 'Lobby application not found'; };
  LET $capacity = type::int($app.lobby_max_players);
  IF $capacity <= 0 OR array::len($members) >= $capacity { THROW 'Lobby is full'; };
  DELETE FROM Lobby_Users WHERE out=type::record($user);
  RELATE (type::record($lobby))->Lobby_Users->(type::record($user)) CONTENT {user_type:'user'};
  UPDATE type::record($lobby) SET date_updated=$now;
 };
 COMMIT TRANSACTION;`, map[string]interface{}{
		"lobby": id, "user": user.ModelID(), "code": code, "now": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	return GetLobby(id)
}

func PutLobby(body *dbtype.Lobby, app *dbtype.Application, user *dbtype.User) (*dbtype.Lobby, error) {
	if body == nil || app == nil || app.ID == nil {
		return nil, fmt.Errorf("missing lobby or application")
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || len(name) > 128 {
		return nil, fmt.Errorf("invalid lobby name")
	}
	if app.ID.Table != "Application" || (user != nil && (user.ID == nil || user.ID.Table != "User")) {
		return nil, fmt.Errorf("invalid application or host")
	}
	if body.Private && body.Code == "" {
		return nil, fmt.Errorf("private lobby requires a code")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	guid := database.GetUUID()
	id := "Lobby:l" + strings.ReplaceAll(guid, "-", "")
	query := `BEGIN TRANSACTION;
		CREATE type::record($lobby) CONTENT $fields;
		RELATE (type::record($lobby))->Lobby_Application->(type::record($app));`
	params := map[string]interface{}{
		"lobby": id, "app": app.ModelID(),
		"fields": map[string]interface{}{
			"name": name, "guid": guid, "date_created": now, "date_updated": now,
			"private": body.Private, "code": body.Code,
		},
	}
	if user != nil {
		params["user"] = user.ModelID()
		query += `
			DELETE FROM Lobby_Host WHERE out=type::record($user);
			DELETE FROM Lobby_Users WHERE out=type::record($user);
			RELATE (type::record($lobby))->Lobby_Host->(type::record($user));
			RELATE (type::record($lobby))->Lobby_Users->(type::record($user)) CONTENT {user_type:'host'};`
	}
	query += ` COMMIT TRANSACTION;`
	if _, err := database.Query[any](query, params); err != nil {
		return nil, err
	}
	return GetLobby(id)
}

func DelLobby(id string, lobby *dbtype.Lobby) error {
	return database.Delete(*lobby.ID)
}

func LinkLobbyApplication(lobby *dbtype.Lobby, app *dbtype.Application) error {
	return database.Relate(lobby.ID, app.ID, "Lobby_Application", map[string]interface{}{})
}

func LinkLobbyHost(lobby *dbtype.Lobby, user *dbtype.User) error {
	if err := database.Relate(lobby.ID, user.ID, "Lobby_Host", map[string]interface{}{}); err != nil {
		return err
	}
	return LinkLobbyUser(lobby, user, "host")
}

func SwapLobbyHost(lobby *dbtype.Lobby, user *dbtype.User) error {
	if lobby == nil || lobby.ID == nil || lobby.ID.Table != "Lobby" || user == nil || user.ID == nil || user.ID.Table != "User" {
		return fmt.Errorf("missing lobby or user")
	}
	held, err := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+strings.TrimPrefix(lobby.ModelID(), "Lobby:"), 60*time.Second, func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := database.Query[any](`
			BEGIN TRANSACTION;
			LET $member = SELECT * FROM Lobby_Users WHERE in=type::record($lobby) AND out=type::record($user);
			IF array::len($member) = 0 { THROW 'New host must already belong to the lobby'; };
			UPDATE Lobby_Users SET user_type='user' WHERE in=type::record($lobby) AND user_type='host';
			UPDATE Lobby_Users SET user_type='host' WHERE in=type::record($lobby) AND out=type::record($user);
			DELETE FROM Lobby_Host WHERE in=type::record($lobby);
			RELATE (type::record($lobby))->Lobby_Host->(type::record($user));
			UPDATE type::record($lobby) SET date_updated=$now;
			COMMIT TRANSACTION;
		`, map[string]interface{}{"lobby": lobby.ModelID(), "user": user.ModelID(), "now": time.Now().UTC().Format(time.RFC3339)})
		return err
	})
	if err != nil {
		return err
	}
	if !held {
		return fmt.Errorf("lobby busy; retry host transfer")
	}
	return nil
}

func LinkLobbyUser(lobby *dbtype.Lobby, user *dbtype.User, user_type string) error {
	if lobby == nil || lobby.ID == nil || user == nil || user.ID == nil {
		return fmt.Errorf("missing lobby or user")
	}
	_, err := database.Query[any](`
 BEGIN TRANSACTION;
 LET $existing = SELECT * FROM Lobby_Users WHERE in=type::record($lobby) AND out=type::record($user);
 IF array::len($existing) = 0 {
  DELETE FROM Lobby_Users WHERE out=type::record($user);
  RELATE (type::record($lobby))->Lobby_Users->(type::record($user)) CONTENT {user_type:$role};
 };
 COMMIT TRANSACTION;`, map[string]interface{}{"lobby": lobby.ModelID(), "user": user.ModelID(), "role": user_type})
	return err
}

func UnlinkLobbyUser(lobby *dbtype.Lobby, user *dbtype.User) error {
	_, err := database.Query[any](
		`DELETE FROM Lobby_Users WHERE in=type::record($lobby) AND out=type::record($user) RETURN *;`,
		map[string]interface{}{
			"lobby": lobby.ID.String(),
			"user":  user.ID.String(),
		},
	)
	return err
}

func UnlinkLobbyAllUsers(id string) error {
	_, err := database.Query[any](
		`DELETE FROM Lobby_Users WHERE in=type::record($lobby) RETURN *;`,
		map[string]interface{}{
			"lobby": id,
		},
	)
	return err
}

func LinkLobbyServer(lobby *dbtype.Lobby, server *dbtype.Server) error {
	return database.Relate(lobby.ID, server.ID, "Lobby_Server", map[string]interface{}{
		"date_created": time.Now().UTC().Format(time.RFC3339),
	})
}
