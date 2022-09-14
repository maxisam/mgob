package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stefanprodan/mgob/pkg/config"
)

func Test_buildDumpCmd_no_uri(t *testing.T) {
	plan := config.Plan{
		Name: "test",
		Target: config.Target{
			Host:     "localhost",
			Port:     27017,
			Username: "user",
			Database: "test",
			Password: "password",
			Params:   "--authenticationDatabase admin",
		},
	}

	dumpCmd := buildDumpCmd("test.gz", plan)
	assert.Equal(t, dumpCmd, `mongodump --archive=test.gz --gzip --host localhost --port 27017 -u "user" -p "password" --db test --authenticationDatabase admin`)
}

func Test_buildDumpCmd_uri(t *testing.T) {
	plan := config.Plan{
		Name: "test",
		Target: config.Target{
			Database: "test",
			Uri:      "mongodb://user:password@localhost:27017",
			Params:   "--authenticationDatabase admin",
		},
	}

	dumpCmd := buildDumpCmd("test.gz", plan)
	assert.Equal(t, dumpCmd, `mongodump --archive=test.gz --gzip --uri "mongodb://user:password@localhost:27017" --db test --authenticationDatabase admin`)
}
