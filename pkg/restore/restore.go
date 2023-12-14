package restore

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	backup "github.com/stefanprodan/mgob/pkg/backup"
	"github.com/stefanprodan/mgob/pkg/config"
)

func Run(plan config.Plan, conf *config.AppConfig, modules *config.ModuleConfig, backupPath string) (backup.Result, error) {
	t1 := time.Now()

	log.WithField("plan", plan.Name).Debugf("Running restore for plan %v, backupPath %v", plan.Name, backupPath)
	restoreCmd := backup.BuildRestoreCmd(backupPath, plan.Target, plan.Target)
	log.WithField("plan", plan.Name).Infof("Running restore command with : %v", restoreCmd)
	fi, err := os.Stat(backupPath)

	res := backup.Result{
		Plan:      plan.Name,
		Timestamp: t1.UTC(),
		Status:    500,
	}
	_, res.Name = filepath.Split(backupPath)
	if err != nil {
		return res, errors.Wrapf(err, "stat file %v failed", backupPath)
	}
	res.Size = fi.Size()
	output, err := backup.RunRestore(backupPath, plan)
	if err != nil || backup.CheckIfAnyFailure(string(output)) != nil {
		log.WithField("plan", plan.Name).Error("Restore failed")
		res.Duration = time.Since(t1)
		return res, errors.Wrapf(err, "failed to execute restore command")
	}
	log.WithField("plan", plan.Name).Debugf("Restore command output: %v", string(output))
	res.Status = 200
	res.Duration = time.Since(t1)
	client, ctx, err := backup.GetMongoClient(backup.BuildUri(plan.Validation.Database))
	if err == nil {
		defer backup.Dispose(client, ctx)
	} else {
		log.WithField("plan", plan.Name).Errorf("Failed to get mongo client: %v, depose skipped", err)
	}
	return res, nil
}
