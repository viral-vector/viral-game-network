//go:build integration

package job

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"strings"
	"testing"
	"time"
	"viral-game-network/src/auth"
	"viral-game-network/src/cache"
	"viral-game-network/src/database"
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
	claims, err := auth.ValidateTokenFor(variables["VNET_SERVER_TOKEN"], auth.ServerAudience)
	if err != nil || claims.Subject != lobby.ModelID() {
		t.Fatal("game pod lacks a lobby-bound server token", err)
	}
	if _, exists := variables["VNET_TOKEN_KEY"]; exists {
		t.Fatal("game pod received private signing key")
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
	if recent, err := repository.SelSystemEventsInFrame(60); err != nil || len(recent) != 1 {
		t.Fatalf("cluster event missing from live feed: %+v %v", recent, err)
	}
}

func TestMonitoringRetainsEventsSharingTimestampAndModernEvents(t *testing.T) {
	support.Storage(t)
	now := time.Now().UTC().Truncate(time.Second)
	first := &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "vgn-app", UID: "first"}, FirstTimestamp: metav1.NewTime(now), EventTime: metav1.NewMicroTime(now), Message: "first"}
	client := fake.NewSimpleClientset(first)
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	RUN_Monitoring_ClusterEvents()
	for _, event := range []*v1.Event{
		{ObjectMeta: metav1.ObjectMeta{Name: "second", Namespace: "vgn-app", UID: "second"}, FirstTimestamp: metav1.NewTime(now), Message: "same timestamp"},
		{ObjectMeta: metav1.ObjectMeta{Name: "modern", Namespace: "vgn-app", UID: "modern"}, EventTime: metav1.NewMicroTime(now), Message: "modern event"},
	} {
		if _, err := client.CoreV1().Events("vgn-app").Create(context.Background(), event, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	}
	RUN_Monitoring_ClusterEvents()
	if events, total, err := repository.AllSystemEvents(10, 1); err != nil || total != 3 || len(events) != 3 {
		t.Fatalf("events dropped at cursor boundary: %+v %d %v", events, total, err)
	}
}

func TestMonitoringRetriesFailedDatabaseWrite(t *testing.T) {
	support.Storage(t)
	now := time.Now().UTC().Truncate(time.Second)
	client := fake.NewSimpleClientset(&v1.Event{ObjectMeta: metav1.ObjectMeta{Name: "retry", Namespace: "vgn-app", UID: "retry"}, FirstTimestamp: metav1.NewTime(now), Message: "retry"})
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	if _, err := database.Query[any](`DEFINE FIELD OVERWRITE message ON TABLE System_Event TYPE string ASSERT $value != 'retry';`, nil); err != nil {
		t.Fatal(err)
	}
	RUN_Monitoring_ClusterEvents()
	if _, err := database.Query[any](`DEFINE FIELD OVERWRITE message ON TABLE System_Event TYPE string;`, nil); err != nil {
		t.Fatal(err)
	}
	RUN_Monitoring_ClusterEvents()
	if events, total, err := repository.AllSystemEvents(10, 1); err != nil || total != 1 || len(events) != 1 {
		t.Fatalf("failed write was not retried: %+v %d %v", events, total, err)
	}
}

func lifecycleFixture(t *testing.T) (*dbtype.Lobby, *dbtype.Server, *fake.Clientset, *v1.Pod) {
	t.Helper()
	support.Storage(t)
	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "worker"}, Status: v1.NodeStatus{Addresses: []v1.NodeAddress{{Type: v1.NodeExternalIP, Address: "203.0.113.10"}}}}
	application, err := repository.PutApplication(&dbtype.Application{Name: "Game", Guid: "game", Image: "game", Version: "1", Port: "4000", Command: "node index.js", Lobby_Max_Players: "4", Lobby_Max_Persist: "30"})
	if err != nil {
		t.Fatal(err)
	}
	lobby, err := repository.PutLobby(&dbtype.Lobby{Name: "Room"}, application, nil)
	if err != nil {
		t.Fatal(err)
	}
	label := strings.TrimPrefix(lobby.ModelID(), "Lobby:")
	server, err := repository.PutServer(&dbtype.Server{Name: "Room", Guid: label, Status: "Pending", Address: "203.0.113.10", Port: 30000}, lobby)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.LinkLobbyServer(lobby, server); err != nil {
		t.Fatal(err)
	}
	lobby.Lobby_Server = server
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: label, Namespace: "vgn-app"}, Spec: v1.PodSpec{NodeName: node.Name, Containers: []v1.Container{{Name: "game", Ports: []v1.ContainerPort{{HostPort: 30000}}}}}, Status: v1.PodStatus{Phase: v1.PodRunning}}
	client := fake.NewSimpleClientset(node, pod)
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	return lobby, server, client, pod
}

func TestProvisionerPreservesServersWhenPodLookupFails(t *testing.T) {
	_, server, client, _ := lifecycleFixture(t)
	client.PrependReactor("get", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("temporary API failure")
	})
	Job_Lobby_Server_Provisioner()
	actual, err := repository.GetServer(server.ModelID())
	if err != nil || actual == nil {
		t.Fatal("pod lookup failure removed the existing server", err)
	}
	for _, action := range client.Actions() {
		if action.GetVerb() == "create" && action.GetResource().Resource == "pods" {
			t.Fatal("pod lookup failure triggered provisioning")
		}
	}
}

func TestRunningStewardshipPreservesOnlineStatusAndPersistsAddressChanges(t *testing.T) {
	lobby, server, client, pod := lifecycleFixture(t)
	server.Status = "Online"
	if _, err := repository.SetServer(server.ModelID(), server); err != nil {
		t.Fatal(err)
	}
	node, err := client.CoreV1().Nodes().Get(context.Background(), "worker", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	node.Status.Addresses[0].Address = "203.0.113.20"
	if _, err := client.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	pod.Spec.Containers[0].Ports[0].HostPort = 30001
	if _, err := client.CoreV1().Pods("vgn-app").Update(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	Job_Lobby_Server_Stewardship()
	actual, err := repository.GetServer(server.ModelID())
	if err != nil || actual.Status != "Online" || actual.Address != "203.0.113.20" || actual.Port != 30001 {
		t.Fatalf("healthy server state changed or address was not saved: %+v %v", actual, err)
	}
	if current, err := repository.GetLobby(lobby.ModelID()); err != nil || current.Lobby_Server == nil {
		t.Fatal("healthy server detached", err)
	}
}

func TestOrphanPendingPodIsPurgedBeforeScheduling(t *testing.T) {
	support.Storage(t)
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "orphan", Namespace: "vgn-app"}, Status: v1.PodStatus{Phase: v1.PodPending}}
	client := fake.NewSimpleClientset(pod)
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	Job_Lobby_Server_Stewardship()
	if _, err := client.CoreV1().Pods("vgn-app").Get(context.Background(), pod.Name, metav1.GetOptions{}); err == nil {
		t.Fatal("orphan pending pod retained because it has no node")
	}
}

func TestFailedPodDeletionPreservesServerForRetry(t *testing.T) {
	_, server, client, pod := lifecycleFixture(t)
	pod.Status.Phase = v1.PodFailed
	if _, err := client.CoreV1().Pods("vgn-app").Update(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	client.PrependReactor("delete", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("temporary API failure")
	})
	Job_Lobby_Server_Stewardship()
	if actual, err := repository.GetServer(server.ModelID()); err != nil || actual == nil {
		t.Fatal("failed pod deletion discarded the server record needed for retry", err)
	}
}

func TestCompletedPodIsPurged(t *testing.T) {
	_, server, client, pod := lifecycleFixture(t)
	pod.Status.Phase = v1.PodSucceeded
	if _, err := client.CoreV1().Pods("vgn-app").Update(context.Background(), pod, metav1.UpdateOptions{}); err != nil {
		t.Fatal(err)
	}
	Job_Lobby_Server_Stewardship()
	if _, err := client.CoreV1().Pods("vgn-app").Get(context.Background(), pod.Name, metav1.GetOptions{}); err == nil {
		t.Fatal("completed pod retained")
	}
	if actual, err := repository.GetServer(server.ModelID()); err != nil || actual != nil {
		t.Fatal("completed server retained", err)
	}
}

func TestUnchangedServerStateDoesNotFloodLobbyNotifications(t *testing.T) {
	lobby, _, _, _ := lifecycleFixture(t)
	Job_Lobby_Server_Stewardship()
	Job_Lobby_Server_Stewardship()
	messages, err := cache.List("lobby:" + lobby.ModelID() + ":channel")
	if err != nil || len(messages) != 1 {
		t.Fatalf("unchanged status generated duplicate messages: %d %v", len(messages), err)
	}
}

func TestLobbyJobsHandleADeletedApplication(t *testing.T) {
	lobby, _, _, _ := lifecycleFixture(t)
	if err := database.Delete(*lobby.Lobby_Application.ID); err != nil {
		t.Fatal(err)
	}
	current, err := repository.GetLobby(lobby.ModelID())
	if err != nil {
		t.Fatal(err)
	}
	if current.Lobby_Application != nil && current.Lobby_Application.ID != nil {
		t.Fatal("deleted application still linked")
	}
	if err := RUN_Lobby_Server_Provisioner(current, strings.TrimPrefix(lobby.ModelID(), "Lobby:")); err == nil {
		t.Fatal("provisioned a lobby with no application")
	}
	RUN_Lobby_Stewardship(current)
}

func TestMonitoringReconcilesMoreThanOneHundredEvents(t *testing.T) {
	support.Storage(t)
	var objects []runtime.Object
	for i := 0; i < 121; i++ {
		objects = append(objects, &v1.Event{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("event-%d", i), Namespace: "vgn-app", UID: types.UID(fmt.Sprintf("event-%d", i))}, Message: "retained event"})
	}
	k8.ConfigureClient(fake.NewSimpleClientset(objects...))
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	for i := 0; i < 2; i++ {
		if err := runMonitoringClusterEvents(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if events, total, err := repository.AllSystemEvents(200, 1); err != nil || len(events) != 121 || total != 121 {
		t.Fatalf("large event batch lost or duplicated: count=%d total=%d error=%v", len(events), total, err)
	}
}

func TestMonitoringHonorsExistingLease(t *testing.T) {
	support.Storage(t)
	client := fake.NewSimpleClientset()
	k8.ConfigureClient(client)
	t.Cleanup(func() { k8.ConfigureClient(nil) })
	held, err := cache.WithLock(context.Background(), "monitoring-cluster-events", time.Second, func(context.Context) error {
		Job_Monitoring()
		return nil
	})
	if !held || err != nil {
		t.Fatal("lease setup failed", err)
	}
	if len(client.Actions()) != 0 {
		t.Fatal("monitoring ran while another worker held the lease")
	}
}
