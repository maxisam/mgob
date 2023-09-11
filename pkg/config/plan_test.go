package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// set env var for test
	testConnectionString := "test-connetion-string"
	testPlanName := "mongo-test"
	os.Setenv(fmt.Sprintf("%s_%s_%s", strings.ToUpper(testPlanName), "AZURE", "CONNECTIONSTRING"), testConnectionString)
	dir := getDir(t)
	plan, err := LoadPlan(dir, testPlanName)

	if err != nil {
		t.Errorf("LoadPlan(%q, %q) returned error: %v", dir, testPlanName, err)
	}

	if plan.Name != testPlanName {
		t.Errorf("LoadPlan(%q, %q) returned plan with wrong name: got %q, want %q", dir, testPlanName, plan.Name, testPlanName)
	}
	if plan.Azure.ConnectionString != testConnectionString {
		t.Errorf("LoadPlan(%q, %q) returned plan with wrong azure connection string: got %q, want %q, viper failed to load env", dir, testPlanName, plan.Azure.ConnectionString, "test-connetion-string")
	}
}

func TestLoadPlans(t *testing.T) {
	// set env var for test
	testConnectionString := "test-connetion-string"
	testPlanName := "mongo-test"
	os.Setenv(fmt.Sprintf("%s_%s_%s", strings.ToUpper(testPlanName), "AZURE", "CONNECTIONSTRING"), testConnectionString)
	dir := getDir(t)

	plans, err := LoadPlans(dir)

	if err != nil {
		t.Errorf("LoadPlans(%q) returned error: %v", dir, err)
	}

	if len(plans) == 0 {
		t.Errorf("LoadPlans(%q) returned empty list of plans", dir)
	}
	// verify plan.name = mongo-test has a azure connection string set as test-connetion-string
	for _, plan := range plans {
		if plan.Name == testPlanName {
			if plan.Azure.ConnectionString != testConnectionString {
				t.Errorf("LoadPlans(%q) returned plan with wrong azure connection string: got %q, want %q", dir, plan.Azure.ConnectionString, "test-connetion-string")
			}
		}
	}
}
