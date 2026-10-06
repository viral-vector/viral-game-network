package k8_k3d

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	data, _ := os.ReadFile(log)
	for _, want := range []string{"cluster delete viral-game-network --all", "cluster create viral-game-network", "30000-30002:30000-30002@server:0:direct", "kubeconfig write viral-game-network"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("missing command %q in %s", want, data)
		}
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
