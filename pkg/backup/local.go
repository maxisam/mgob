package backup

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/stefanprodan/mgob/pkg/config"
)

func dump(plan config.Plan, tmpPath string, ts time.Time, attempt float64) (string, string, error) {
	retryCount := attempt
	archive := fmt.Sprintf("%v/%v-%v.gz", tmpPath, plan.Name, ts.Unix())
	mlog := fmt.Sprintf("%v/%v-%v.log", tmpPath, plan.Name, ts.Unix())
	dumpCmd := fmt.Sprintf("mongodump --archive=%v --gzip ", archive)

	if plan.Target.Uri != "" {
		// using uri (New in version 3.4.6)
		// host/port/username/password are incompatible with uri
		// https://docs.mongodb.com/manual/reference/program/mongodump/#cmdoption-mongodump-uri
		dumpCmd += fmt.Sprintf(`--uri "%v" `, plan.Target.Uri)
	} else {
		// use older host/port
		dumpCmd += fmt.Sprintf("--host %v --port %v ", plan.Target.Host, plan.Target.Port)

		if plan.Target.Username != "" && plan.Target.Password != "" {
			dumpCmd += fmt.Sprintf(`-u "%v" -p "%v" `, plan.Target.Username, plan.Target.Password)
		}
	}

	if plan.Target.Database != "" {
		dumpCmd += fmt.Sprintf("--db %v ", plan.Target.Database)
	}

	if plan.Target.Params != "" {
		dumpCmd += fmt.Sprintf("%v", plan.Target.Params)
	}

	log.Debugf("dump cmd: %v", strings.Replace(dumpCmd, fmt.Sprintf(`-p "%v"`, plan.Target.Password), "-p xxxx", -1))
	output, err := sh.Command("/bin/sh", "-c", dumpCmd).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		// Try and clean up tmp file after an error
		os.Remove(archive)
		retryCount++
		time.Sleep(time.Duration(plan.Retry.BackoffFactor * float32(math.Pow(2, retryCount)) * float32(time.Second)))
		if retryCount <= float64(plan.Retry.Attempts) {
			log.Debugf("retrying dump: %v", retryCount)
			return dump(plan, tmpPath, ts, retryCount)
		} else {

			return "", "", errors.Wrapf(err, "mongodump log %v", ex)
		}
	}
	logToFile(mlog, output)

	return archive, mlog, nil
}

func logToFile(file string, data []byte) error {
	if len(data) > 0 {
		err := ioutil.WriteFile(file, data, 0644)
		if err != nil {
			return errors.Wrapf(err, "writing log %v failed", file)
		}
	}

	return nil
}

func applyRetention(path string, retention int) error {
	gz := fmt.Sprintf("cd %v && rm -f $(ls -1t *.gz *.gz.encrypted | tail -n +%v)", path, retention+1)
	err := sh.Command("/bin/sh", "-c", gz).Run()
	if err != nil {
		return errors.Wrapf(err, "removing old gz files from %v failed", path)
	}

	log.Debug("apply retention")
	log := fmt.Sprintf("cd %v && rm -f $(ls -1t *.log | tail -n +%v)", path, retention+1)
	err = sh.Command("/bin/sh", "-c", log).Run()
	if err != nil {
		return errors.Wrapf(err, "removing old log files from %v failed", path)
	}

	return nil
}

// TmpCleanup remove files older than one day
func TmpCleanup(path string) error {
	rm := fmt.Sprintf("find %v -not -name \"mgob.db\" -mtime +%v -type f -delete", path, 1)
	err := sh.Command("/bin/sh", "-c", rm).Run()
	if err != nil {
		return errors.Wrapf(err, "%v cleanup failed", path)
	}

	return nil
}
