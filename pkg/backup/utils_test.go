package backup

import (
	"testing"

	"github.com/stefanprodan/mgob/pkg/config"
	"github.com/stretchr/testify/assert"
)

func Test_buildCmd_no_uri(t *testing.T) {
	target := config.Target{
		Host:     "localhost",
		Port:     27017,
		Username: "user",
		Database: "test",
		Password: "password",
		Params:   "--authenticationDatabase admin",
	}

	dumpCmd := buildCmd("mongodump", "test.gz", target)
	assert.Equal(t, dumpCmd, `mongodump --archive=test.gz --gzip --host localhost --port 27017 -u "user" -p "password" --db test --authenticationDatabase admin`)
}

func Test_buildCmd_uri(t *testing.T) {
	target := config.Target{
		Database: "test",
		Uri:      "mongodb://user:password@localhost:27017",
		Params:   "--authenticationDatabase admin",
	}

	dumpCmd := buildCmd("mongodump", "test.gz", target)
	assert.Equal(t, dumpCmd, `mongodump --archive=test.gz --gzip --uri "mongodb://user:password@localhost:27017" --db test --authenticationDatabase admin`)
}
