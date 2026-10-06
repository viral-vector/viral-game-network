package struct_merge

import "testing"

func TestMergePatch(t *testing.T) {
	type profile struct {
		Name   string
		Count  int
		Active bool
		Labels []string
		hidden string
	}
	original := profile{Name: "old", Count: 3, Active: true, Labels: []string{"old"}, hidden: "private"}
	patch := profile{Name: "new", Labels: []string{}, hidden: "changed"}
	if err := Merge(&original, &patch); err != nil {
		t.Fatal(err)
	}
	if original.Name != "new" || original.Count != 3 || !original.Active || original.hidden != "private" || len(original.Labels) != 0 {
		t.Fatalf("unexpected merged profile: %+v", original)
	}
	if patch.Name != "new" || patch.Count != 0 {
		t.Fatal("source was mutated")
	}
}

func TestMergeRejectsInvalidInputs(t *testing.T) {
	type profile struct{ Name string }
	if err := Merge[profile](nil, &profile{}); err == nil {
		t.Fatal("nil destination accepted")
	}
	if err := Merge[profile](&profile{}, nil); err == nil {
		t.Fatal("nil source accepted")
	}
	n := 1
	if err := Merge(&n, &n); err == nil {
		t.Fatal("scalar input accepted")
	}
}
