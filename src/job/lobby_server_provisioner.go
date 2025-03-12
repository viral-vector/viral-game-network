package job

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
)

/**
 * Creates Servers for lobbys
 */
func Job_Lobby_Server_Provisioner() {
	// fmt.Println("Job_Lobby_Server_Provisioner")

	lobbies, _, err := repository.AllLobby(-1, 1)
	if err != nil {
		fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	for _, lobby := range lobbies {
		// Get the label for the lobby
		label := strings.Split(lobby.ID.String(), ":")[1]

		fmt.Println("Job_Lobby_Server_Provisioner: @ Looking ", label)

		// Get the server pod
		_, lobpod, _ := k8.LocateServerPod(label)
		if lobpod == nil && lobby.Lobby_Server != nil {
			fmt.Println("Job_Lobby_Server_Provisioner: @ Delete", lobby.Lobby_Server.ID.String())
			repository.DelServer(lobby.Lobby_Server.ID.String())
			lobby.Lobby_Server = nil
		}

		// Do we need a new server?
		if lobby.Lobby_Server == nil && len(lobby.Lobby_Users) >= 0 {
			// Check if we have a k8-Lock on the lobby
			if cache_lock, _ := cache.Get("provisioner-lock-" + label); cache_lock != nil {
				continue
			}
			go func(lobby dbtype.Lobby, label string) {
				// k8-Lock the lobby
				cache.Set("provisioner-lock-"+label, "true", time.Second*15)
				// Create the server pod
				err := Lobby_Server_Provisioner_PUT(lobby, label)
				if err != nil {
					fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
				}
				// k8-Unlock the lobby
				cache.Del("k8-lock-" + label)
			}(lobby, label)
		}
	}
}

// Job_Lobby_Server_Provisioner_PUT
func Lobby_Server_Provisioner_PUT(lobby dbtype.Lobby, label string) error {
	fmt.Println("Job_Lobby_Server_Provisioner: @ Serving ", label)

	// Get the server pod
	_, lobpod, err := k8.LocateServerPod(label)
	if lobpod != nil {
		fmt.Println("Job_Lobby_Server_Provisioner: @ Pod Exisits ->", lobpod.Status.Phase)
		return err
	}

	// Pick Open Node/Port
	node, sPort, err := k8.FindOpenNodePort()

	aPort, err := strconv.ParseInt(os.Getenv("GAME_PORT"), 10, 32)
	if err != nil {
		fmt.Println("Job_Lobby_Server_Provisioner: @ Errors ", err)
		return fmt.Errorf("error converting string to int: %s", err)
	}

	// Format the command
	gmd := "node index.js" // os.Getenv("GAME_COMMANDS")
	cmd := []string{}
	for _, str := range strings.Split(gmd, " ") {
		str = strings.Replace(str, "${GAME_NAME}", os.Getenv("GAME_NAME"), -1)		
		str = strings.Replace(str, "${GAME_PORT}", strconv.Itoa(int(aPort)), -1)

		str = strings.Replace(str, "${NODE_HOST}", "", -1)
		str = strings.Replace(str, "${NODE_PORT}", strconv.Itoa(int(sPort)), -1)

		str = strings.Replace(str, "${VNET_HOST}", os.Getenv("VNET_HOST"), -1)
		str = strings.Replace(str, "${VNET_PORT}", os.Getenv("VNET_PORT"), -1)
		str = strings.Replace(str, "${VNET_KEY}", os.Getenv("VNET_KEY"), -1)

		str = strings.Replace(str, "${LOBBY_MAX_PLAYERS}", os.Getenv("LOBBY_MAX_PLAYERS"), -1)
		str = strings.Replace(str, "${LOBBY_ID}", lobby.ID.String(), -1)
		str = strings.Replace(str, "${LOBBY_NAME}", lobby.Name, -1)
		cmd = append(cmd, str)
	}

	// Prepare ENVS
	env := map[string]string{
		"GAME_NAME": os.Getenv("GAME_NAME"),
		"GAME_PORT": strconv.Itoa(int(aPort)),
		"NODE_HOST": "",
		"NODE_PORT": strconv.Itoa(int(sPort)),
		"VNET_HOST": os.Getenv("VNET_HOST"),
		"VNET_PORT": os.Getenv("VNET_PORT"),
		"VNET_KEY":  os.Getenv("VNET_KEY"),
		"LOBBY_MAX_PLAYERS": os.Getenv("LOBBY_MAX_PLAYERS"),
		"LOBBY_ID":   lobby.ID.String(),
		"LOBBY_NAME": lobby.Name,
	}

	// Create the server pod
	image := os.Getenv("GAME_DOKIMAGE")
	lobpod, err = k8.CreateServerPod(label, node, int32(sPort), int32(aPort), image, cmd, env)
	if err != nil {
		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	// Create A Server Db Entry
	external_address := k8.GetNodeExternalIP(node)

	server, err := repository.PutServer(&dbtype.Server{
		Name:    lobby.Name,
		Guid:    lobpod.Labels["app"],
		Status:  string(lobpod.Status.Phase),
		Address: external_address,
		Port:    int32(sPort),
	}, &lobby)
	if err != nil {
		// Delete the server pod if we failed to create the server entry
		k8.DeleteServerPod(label)

		return fmt.Errorf("Job_Lobby_Server_Provisioner:  %s", err)
	}

	// Link Lobby & Server
	err = repository.LinkLobbyServer(&lobby, server)
	if err != nil {
		// Delete the server pod if we failed to create the server entry
		k8.DeleteServerPod(label)
		// Delete the server entry if we failed to create the server entry
		repository.DelServer(server.ID.String())
	}

	fmt.Println("Job_Lobby_Server_Provisioner: Server/Pod Created ", lobpod.Status.Phase)

	return nil
}
