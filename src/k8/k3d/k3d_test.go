package k8_k3d

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClusterCommands(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "commands")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$COMMAND_LOG\"\nif [ \"$1 $2\" = 'cluster list' ]; then printf 'viral-game-network 1/1\\n'; fi\n"
	if err := os.WriteFile(filepath.Join(dir, "k3d"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("COMMAND_LOG", log)
	running, err := CheckCluster()
	if err != nil || !running {
		t.Fatalf("%v %v", running, err)
	}
	if err := CreateCluster(30000, 30002); err != nil {
		t.Fatal(err)
	}
	if err := DeleteCluster(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(log)
	for _, want := range []string{"cluster delete viral-game-network\n", "cluster create viral-game-network", "30000-30002:30000-30002@server:0:direct", "kubeconfig write viral-game-network"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("missing command %q in %s", want, data)
		}
	}
}

func TestInvalidPortRangeDoesNotInvokeClusterCLI(t *testing.T) {
	log := fakeK3d(t, "")
	for _, ports := range [][2]int32{{0, 30000}, {30000, 65536}, {30002, 30000}, {-1, -1}} {
		if err := CreateCluster(ports[0], ports[1]); err == nil {
			t.Fatalf("invalid port range accepted: %v", ports)
		}
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("invalid request reached the cluster CLI", err)
	}
}

func TestClusterCommandDeadlineStopsHungCLI(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "k3d"), []byte("#!/bin/sh\nwhile :; do :; done\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	started := time.Now()
	if _, err := runK3d(20*time.Millisecond, "cluster", "list"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("hung command did not return a deadline error: %v", err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("hung CLI continued after the deadline")
	}
}

func TestClusterCommandFailure(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "k3d"), []byte("#!/bin/sh\necho failed >&2\nexit 1\n"), 0700)
	t.Setenv("PATH", dir)
	if _, err := CheckCluster(); err == nil {
		t.Fatal("failed cluster check accepted")
	}
	if err := WrtiteKubeConfig(); err == nil {
		t.Fatal("failed config write accepted")
	}
	if err := DeleteCluster(); err == nil {
		t.Fatal("failed deletion accepted")
	}
}

func fakeK3d(t *testing.T, output string) string {
	t.Helper()
	dir := t.TempDir()
	log := filepath.Join(dir, "commands")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$COMMAND_LOG\"\nif [ \"$1 $2\" = 'cluster list' ]; then printf '%s' \"$CLUSTER_OUTPUT\"; fi\n"
	if err := os.WriteFile(filepath.Join(dir, "k3d"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("COMMAND_LOG", log)
	t.Setenv("CLUSTER_OUTPUT", output)
	return log
}

func TestCreatingClusterDoesNotDeleteAnyCluster(t *testing.T) {
	log := fakeK3d(t, "")
	if err := CreateCluster(30000, 30002); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "cluster delete") {
		t.Fatalf("creation deletes existing clusters: %s", data)
	}
}

func TestClusterDeletionOnlyTargetsProject(t *testing.T) {
	log := fakeK3d(t, "")
	if err := DeleteCluster(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cluster delete viral-game-network\n" {
		t.Fatalf("deletion escaped the project: %s", data)
	}
}

func TestClusterCheckRequiresExactName(t *testing.T) {
	for _, output := range []string{
		"NAME SERVERS AGENTS\nviral-game-network-other 1/1 0/0\n",
		"NAME SERVERS AGENTS\nother-viral-game-network 1/1 0/0\n",
		"",
	} {
		t.Run(output, func(t *testing.T) {
			fakeK3d(t, output)
			if found, err := CheckCluster(); err != nil || found {
				t.Fatalf("unrelated cluster matched: %v %v", found, err)
			}
		})
	}
}
