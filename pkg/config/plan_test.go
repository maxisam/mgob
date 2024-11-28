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
	dir := filepath.Join(cwd, "../../test", "unit-test")
	return dir
}

func TestLoadPlan(t *testing.T) {
	// set env var for test
	testConnectionString := "test-connetion-string"
	testPlanName := "mongo-test"
	os.Setenv(fmt.Sprintf("%s__%s_%s", strings.ToUpper(strings.Replace(testPlanName, "-", "_", -1)), "AZURE", "CONNECTIONSTRING"), testConnectionString)
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
	targetPassword := "test-password"
	os.Setenv(fmt.Sprintf("%s__%s_%s", strings.ToUpper(strings.ReplaceAll(testPlanName, "-", "_")), "AZURE", "CONNECTIONSTRING"), testConnectionString)
	os.Setenv(fmt.Sprintf("%s__%s_%s", strings.ToUpper(strings.ReplaceAll(testPlanName, "-", "_")), "TARGET", "PASSWORD"), targetPassword)
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
			if plan.Target.Password != targetPassword {
				t.Errorf("LoadPlans(%q) returned plan with wrong target password: got %q, want %q", dir, plan.Target.Password, targetPassword)
			}
			if plan.Azure.ConnectionString != testConnectionString {
				t.Errorf("LoadPlans(%q) returned plan with wrong azure connection string: got %q, want %q", dir, plan.Azure.ConnectionString, "test-connetion-string")
			}
		}
	}
}
