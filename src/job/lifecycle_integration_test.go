//go:build integration

package job

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
	"viral-game-network/tests/support"
)

func TestProvisionRunningAndFailedServerLifecycle(t *testing.T) {
	support.Storage(t)
	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}, Status: v1.NodeStatus{Addresses: []v1.NodeAddress{{Type: v1.NodeExternalIP, Address: "203.0.113.10"}}}}
	client := fake.NewSimpleClientset(node)
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "node index.js", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	lobby, err := repository.PutLobby(&dbtype.Lobby{Name: "Room"}, application, nil)
	if err != nil {
		t.Fatal(err)
	}
	label := strings.TrimPrefix(lobby.ModelID(), "Lobby:")
	Job_Lobby_Server_Provisioner()
	lobby, err = repository.GetLobby(lobby.ModelID())
	if err != nil || lobby.Lobby_Server == nil {
		t.Fatalf("server not linked: %+v %v", lobby, err)
	}
	pod, err := client.CoreV1().Pods("vgn-app").Get(context.Background(), label, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if pod.Spec.Containers[0].Image != "game:1" || pod.Labels["server"] == "" || k8.GetPodHostPort(pod) != lobby.Lobby_Server.Port {
		t.Fatalf("incorrect provisioning: %+v", pod)
	}
	variables := map[string]string{}
	for _, variable := range pod.Spec.Containers[0].Env {
		variables[variable.Name] = variable.Value
	}
	if variables["LOBBY_ID"] != lobby.ModelID() || variables["GAME_PORT"] != "4000" || variables["NODE_HOST"] != "203.0.113.10" {
		t.Fatal(variables)
	}
	pod.Spec.NodeName = node.Name
	pod.Status.Phase = v1.PodRunning
	pod, err = client.CoreV1().Pods("vgn-app").Update(context.Background(), pod, metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := RUN_Lobby_Server_Provisioner(lobby, label); err != nil {
		t.Fatal(err)
	}
	if servers, total, err := repository.AllServer(10, 1, ""); err != nil || total != 1 || len(servers) != 1 {
		t.Fatalf("duplicate server: %+v %d %v", servers, total, err)
	}
	Job_Lobby_Server_Stewardship()
	lobby, err = repository.GetLobby(lobby.ModelID())
	if err != nil || lobby.Lobby_Server.Status != "Running" || lobby.Lobby_Server.Address != "203.0.113.10" {
		t.Fatalf("running status not persisted: %+v %v", lobby, err)
	}
	pod.Status.Phase = v1.PodFailed
	pod, err = client.CoreV1().Pods("vgn-app").Update(context.Background(), pod, metav1.UpdateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	Job_Lobby_Server_Stewardship()
	if _, err := client.CoreV1().Pods("vgn-app").Get(context.Background(), label, metav1.GetOptions{}); err == nil {
		t.Fatal("failed pod retained")
	}
	if servers, total, err := repository.AllServer(10, 1, ""); err != nil || total != 0 || len(servers) != 0 {
		t.Fatalf("failed server retained: %+v %d %v", servers, total, err)
	}
}

func TestLobbyExpirationKeepsFreshRoomsAndDeletesExpiredRooms(t *testing.T) {
	support.Storage(t)
	app, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "node index.js", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	lobby, err := repository.PutLobby(&dbtype.Lobby{Name: "Room"}, app, nil)
	if err != nil {
		t.Fatal(err)
	}
	Job_Lobby_Stewardship()
	if current, err := repository.GetLobby(lobby.ModelID()); err != nil || current == nil {
		t.Fatal("fresh lobby removed", err)
	}
	lobby.Date_Updated = time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	RUN_Lobby_Stewardship(lobby)
	if current, err := repository.GetLobby(lobby.ModelID()); err != nil || current != nil {
		t.Fatal("expired lobby retained", err)
	}
}

func TestMonitoringRecordsClusterEventsOnce(t *testing.T) {
	redis := support.Storage(t)
	now := time.Now().UTC().Truncate(time.Second)
	event := &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "warning", Namespace: "vgn-app", UID: "event-1"}, FirstTimestamp: metav1.NewTime(now), EventTime: metav1.NewMicroTime(now), Type: "Warning", Message: "Image pull failed", InvolvedObject: v1.ObjectReference{Name: "game"}}
	client := fake.NewSimpleClientset(event)
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	Job_Monitoring()
	Job_Monitoring()
	redis.FastForward(6 * time.Second)
	Job_Monitoring()
	events, total, err := repository.AllSystemEvents(10, 1)
	if err != nil || total != 1 || len(events) != 1 {
		t.Fatalf("duplicate monitoring events: %+v %d %v", events, total, err)
	}
	if events[0].Severity != "warning" || events[0].Ref_Target != "game" || events[0].Message != "Image pull failed" {
		t.Fatal(events[0])
	}
	if cached, err := cache.Get[string]("monitoring-cluster:last_event"); err != nil || cached != now.Format(time.RFC3339) {
		t.Fatalf("event cursor: %s %v", cached, err)
	}
}
