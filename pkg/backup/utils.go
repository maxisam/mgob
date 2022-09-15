package backup

import (
	"fmt"

	"github.com/stefanprodan/mgob/pkg/config"
)

func BuildDumpCmd(archive string, plan config.Plan) string {
	return buildCmd("mongodump", archive, plan.Target)
}

func BuildRestoreCmd(archive string, plan config.Plan) string {
	return buildCmd("mongorestore", archive, plan.Target)
}

// command: mongodump | mongorestore
func buildCmd(command string, archive string, target config.Target) string {
	cmd := fmt.Sprintf("%v --archive=%v --gzip ", command, archive)
	// using uri (New in version 3.4.6)
	// host/port/username/password are incompatible with uri
	// https://docs.mongodb.com/manual/reference/program/mongodump/#cmdoption-mongodump-uri
	// use older host/port
	if target.Uri != "" {
		cmd += fmt.Sprintf(`--uri "%v" `, target.Uri)
	} else {
		cmd += fmt.Sprintf("--host %v --port %v ", target.Host, target.Port)

		if target.Username != "" && target.Password != "" {
			cmd += fmt.Sprintf(`-u "%v" -p "%v" `, target.Username, target.Password)
		}
	}

	if target.Database != "" {
		cmd += fmt.Sprintf("--db %v ", target.Database)
	}

	if target.Params != "" {
		cmd += fmt.Sprintf("%v", target.Params)
	}
	return cmd
}

func buildUri(target config.Target) string {
	return fmt.Sprintf("mongodb://%v:%v@%v:%v", target.Username, target.Password, target.Host, target.Port)
}