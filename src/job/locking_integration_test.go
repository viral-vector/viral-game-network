//go:build integration

package job

import (
	"context"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/tests/support"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

func TestProvisioningSerializesPortReservation(t *testing.T) {
	support.Storage(t)
	client := fake.NewSimpleClientset(&v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}})
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	app, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "server", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := repository.PutLobby(&dbtype.Lobby{Name: "First"}, app, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.PutLobby(&dbtype.Lobby{Name: "Second"}, app, nil)
	if err != nil {
		t.Fatal(err)
	}
	creating, release := make(chan struct{}), make(chan struct{})
	client.PrependReactor("create", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		if action.(ktesting.CreateAction).GetObject().(*v1.Pod).Name == strings.TrimPrefix(first.ModelID(), "Lobby:") {
			close(creating)
			<-release
		}
		return false, nil, nil
	})
	finished := make(chan error, 1)
	go func() { finished <- RUN_Lobby_Server_Provisioner(first, strings.TrimPrefix(first.ModelID(), "Lobby:")) }()
	select {
	case <-creating:
	case <-time.After(3 * time.Second):
		close(release)
		<-finished
		t.Fatal("first provision did not reserve a port")
	}
	err = RUN_Lobby_Server_Provisioner(second, strings.TrimPrefix(second.ModelID(), "Lobby:"))
	close(release)
	if firstErr := <-finished; firstErr != nil {
		t.Fatal(firstErr)
	}
	if err == nil || !strings.Contains(err.Error(), "allocation busy") {
		t.Fatal("concurrent provisioning did not defer allocation", err)
	}
	if err := RUN_Lobby_Server_Provisioner(second, strings.TrimPrefix(second.ModelID(), "Lobby:")); err != nil {
		t.Fatal(err)
	}
	first, _ = repository.GetLobby(first.ModelID())
	second, _ = repository.GetLobby(second.ModelID())
	if first.Lobby_Server == nil || second.Lobby_Server == nil || first.Lobby_Server.Port == second.Lobby_Server.Port {
		t.Fatal("pending pods share a host port")
	}
}

func TestLifecycleJobsShareLobbyLock(t *testing.T) {
	lobby, server, client, pod := lifecycleFixture(t)
	lobby.Date_Updated = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	if _, err := database.Update(*lobby.ID, lobby); err != nil {
		t.Fatal(err)
	}
	pod.Status.Phase = v1.PodFailed
	if _, err := client.CoreV1().Pods("vgn-app").Update(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	held, err := cache.WithLock(context.Background(), "lobby-lifecycle-lock-"+pod.Name, time.Second, func(context.Context) error {
		Job_Lobby_Server_Provisioner()
		Job_Lobby_Server_Stewardship()
		Job_Lobby_Stewardship()
		return nil
	})
	if !held || err != nil {
		t.Fatal(held, err)
	}
	if _, err := client.CoreV1().Pods("vgn-app").Get(context.Background(), pod.Name, metav1.GetOptions{}); err != nil {
		t.Fatal("stewardship ignored active lifecycle lock", err)
	}
	current, err := repository.GetLobby(lobby.ModelID())
	if err != nil || current.Lobby_Server == nil || current.Lobby_Server.ModelID() != server.ModelID() {
		t.Fatal("locked server state changed", current, err)
	}
}
