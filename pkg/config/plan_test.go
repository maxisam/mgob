package config

import (
	"os"
	"path/filepath"
	"testing"
)

func getDir(t *testing.T) string {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	dir := filepath.Join(cwd, "../../test", "gh-actions")
	return dir
}

func TestLoadPlan(t *testing.T) {
	dir := getDir(t)
	name := "mongo-test"

	plan, err := LoadPlan(dir, name)

	if err != nil {
		t.Errorf("LoadPlan(%q, %q) returned error: %v", dir, name, err)
	}

	if plan.Name != name {
		t.Errorf("LoadPlan(%q, %q) returned plan with wrong name: got %q, want %q", dir, name, plan.Name, name)
	}
}

func TestLoadPlans(t *testing.T) {
	dir := getDir(t)

	plans, err := LoadPlans(dir)

	if err != nil {
		t.Errorf("LoadPlans(%q) returned error: %v", dir, err)
	}

	if len(plans) == 0 {
		t.Errorf("LoadPlans(%q) returned empty list of plans", dir)
	}
}
