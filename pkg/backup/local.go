package backup

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/stefanprodan/mgob/pkg/config"
)

func localBackup(file string, storagePath string, mlog string, plan config.Plan) (string, error) {
	t1 := time.Now()
	planDir := fmt.Sprintf("%v/%v", storagePath, plan.Name)
	err := sh.Command("mkdir", "-p", planDir).Run()
	if err != nil {
		return "", errors.Wrapf(err, "creating dir %v in %v failed", plan.Name, storagePath)
	}
	err = sh.Command("cp", file, planDir).Run()
	if err != nil {
		return "", errors.Wrapf(err, "moving file from %v to %v failed", file, planDir)
	}
	// check if log file exists, is not always created
	if _, err := os.Stat(mlog); os.IsNotExist(err) {
		log.Debug("appears no log file was generated")
	} else {
		err = sh.Command("cp", mlog, planDir).Run()
		if err != nil {
			return "", errors.Wrapf(err, "moving file from %v to %v failed", mlog, planDir)
		}
	}
	if plan.Scheduler.Retention > 0 {
		err = applyRetention(planDir, plan.Scheduler.Retention)
		if err != nil {
			return "", errors.Wrap(err, "retention job failed")
		}
	}
	_, filename := filepath.Split(file)
	distPath := filepath.Join(planDir, filename)
	t2 := time.Now()
	msg := fmt.Sprintf("Local backup finished filename:`%v`, filepath:`%v`, Duration: %v",
		file, distPath, t2.Sub(t1))
	return msg, nil
}

func dump(plan config.Plan, tmpPath string, ts time.Time) (string, string, error) {
	retryCount := 0.0
	archive := fmt.Sprintf("%v/%v-%v.gz", tmpPath, plan.Name, ts.Unix())
	mlog := fmt.Sprintf("%v/%v-%v.log", tmpPath, plan.Name, ts.Unix())
	dumpCmd := BuildDumpCmd(archive, plan.Target)
	timeout := time.Duration(plan.Scheduler.Timeout) * time.Minute

	log.Debugf("dump cmd: %v", strings.Replace(dumpCmd, fmt.Sprintf(`-p "%v"`, plan.Target.Password), "-p xxxx", -1))
	output, retryCount, err := runDump(dumpCmd, plan.Retry, archive, retryCount, timeout)
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", "", errors.Wrapf(err, "after %v retries, mongodump log %v", retryCount, ex)
	}
	if plan.Validation != nil {
		backupResult := getDumpedDocMap(string(output))
		if isValidate, err := ValidateBackup(archive, plan, backupResult); !isValidate || err != nil {
			client, ctx, err := GetMongoClient(BuildUri(plan.Validation.Database))
			if err != nil {
				return "", "", errors.Wrapf(err, "Failed to validate backup, failed to get mongo client for cleanup")
			}
			defer Dispose(client, ctx)
			if err = cleanMongo(plan.Validation.Database.Database, client); err != nil {
				return "", "", errors.Wrapf(err, "Failed to validate backup, failed to clean mongo validation database")
			}
			return "", "", errors.Wrapf(err, "backup validation failed")
		}
	}
	logToFile(mlog, output)

	return archive, mlog, nil
}

func getDumpedDocMap(output string) map[string]string {
	result := map[string]string{}
	dbDocCapRegex := `done dumping\s\w*\.(\S*)\s\((\d*).document`
	reg := regexp.MustCompile(dbDocCapRegex)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "done dumping") {
			matches := reg.FindStringSubmatch(line)
			if reg.NumSubexp() == 2 {
				result[matches[1]] = matches[2]
				log.Debugf("dumped %v documents from %v", matches[2], matches[1])
			}
		}
	}
	return result
}

func runDump(dumpCmd string, retryPlan config.Retry, archive string, retryAttempt float64, timeout time.Duration) ([]byte, float64, error) {
	duration := float32(0)
	output, err := sh.Command("/bin/sh", "-c", dumpCmd).SetTimeout(timeout).CombinedOutput()
	if err != nil {
		// Try and clean up tmp file after an error
		os.Remove(archive)
		retryAttempt++
		if retryAttempt > float64(retryPlan.Attempts) {
			return nil, retryAttempt - 1, err
		}
		duration = retryPlan.BackoffFactor * float32(math.Pow(2, retryAttempt)) * float32(time.Second)
		time.Sleep(time.Duration(duration))
		log.Debugf("retrying dump: %v after %v second", retryAttempt, duration)
		return runDump(dumpCmd, retryPlan, archive, retryAttempt, timeout)
	}
	return output, retryAttempt, nil
}

func logToFile(file string, data []byte) error {
	if len(data) > 0 {
		err := os.WriteFile(file, data, 0644)
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
