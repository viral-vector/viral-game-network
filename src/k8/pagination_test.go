package k8

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func pagedCluster(t *testing.T) {
	t.Helper()
	previous := getClient()
	items := make([]v1.Pod, 5)
	for i := range items {
		items[i] = v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("pod-%d", i), Namespace: namespace}}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("continue") == "expired" || (r.URL.Query().Get("labelSelector") == "app=expires" && r.URL.Query().Get("limit") != "") {
			w.WriteHeader(http.StatusGone)
			json.NewEncoder(w).Encode(metav1.Status{TypeMeta: metav1.TypeMeta{Kind: "Status", APIVersion: "v1"}, Status: metav1.StatusFailure, Reason: metav1.StatusReasonExpired, Code: 410})
			return
		}
		items := items
		if r.URL.Query().Get("labelSelector") == "app=missing" {
			items = nil
		}
		start, _ := strconv.Atoi(r.URL.Query().Get("continue"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if start < 0 || start > len(items) {
			t.Error("invalid cursor", start)
			w.WriteHeader(400)
			return
		}
		end := len(items)
		if limit > 0 && start+limit < end {
			end = start + limit
		}
		result := v1.PodList{TypeMeta: metav1.TypeMeta{Kind: "PodList", APIVersion: "v1"}, Items: items[start:end]}
		if end < len(items) {
			result.Continue = strconv.Itoa(end)
		}
		json.NewEncoder(w).Encode(result)
	}))
	t.Cleanup(server.Close)
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	ConfigureClient(client)
	t.Cleanup(func() { ConfigureClient(previous) })
}

func TestPodPagerResumesOnlyTheNextPage(t *testing.T) {
	pagedCluster(t)
	for _, input := range []struct {
		name              string
		previous, desired int
		token             string
		names             []string
		next              string
	}{
		{"next", 1, 2, "2", []string{"pod-2", "pod-3"}, "4"},
		{"previous", 3, 2, "4", []string{"pod-2", "pod-3"}, "4"},
		{"jump", 1, 3, "2", []string{"pod-4"}, ""},
		{"first", 2, 1, "4", []string{"pod-0", "pod-1"}, "2"},
		{"out of range", 1, 9, "2", nil, ""},
		{"expired cursor", 1, 2, "expired", []string{"pod-2", "pod-3"}, "4"},
	} {
		t.Run(input.name, func(t *testing.T) {
			pods, total, next, err := GetAllServerPodsPager("", 2, input.previous, input.desired, input.token)
			var names []string
			for _, pod := range pods {
				names = append(names, pod.Name)
			}
			if err != nil || total != 5 || next != input.next || !reflect.DeepEqual(names, input.names) {
				t.Fatalf("names=%v total=%d next=%q error=%v; want names=%v next=%q", names, total, next, err, input.names, input.next)
			}
		})
	}
}

func TestPodPagerRejectsInvalidPagination(t *testing.T) {
	pagedCluster(t)
	for _, input := range []struct {
		limit             int64
		previous, desired int
	}{{0, 1, 1}, {-1, 1, 1}, {2, 0, 1}, {2, 1, 0}} {
		if _, _, _, err := GetAllServerPodsPager("", input.limit, input.previous, input.desired, ""); err == nil {
			t.Errorf("invalid pagination accepted: %+v", input)
		}
	}
}

func TestPodPagerEmptyListAndRepeatedExpiration(t *testing.T) {
	pagedCluster(t)
	pods, total, next, err := GetAllServerPodsPager("missing", 2, 1, 1, "")
	if err != nil || len(pods) != 0 || total != 0 || next != "" {
		t.Fatalf("empty list: pods=%v total=%d next=%q error=%v", pods, total, next, err)
	}
	_, _, _, err = GetAllServerPodsPager("expires", 2, 1, 2, "expired")
	if !apierrors.IsResourceExpired(err) {
		t.Fatalf("repeated expiration must return the Kubernetes error: %v", err)
	}
}
