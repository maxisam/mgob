package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/stefanprodan/mgob/pkg/config"
)

func Run(plan config.Plan, conf *config.AppConfig, modules *config.ModuleConfig) (Result, error) {
	tmpPath := conf.TmpPath

	t1 := time.Now()

	archive, mlog, err := dump(plan, tmpPath, t1.UTC())
	log.WithFields(log.Fields{
		"plan":    plan.Name,
		"archive": archive,
		"mlog":    mlog,
		"err":     err,
	}).Info("new dump")

	res := Result{
		Plan:      plan.Name,
		Timestamp: t1.UTC(),
		Status:    500,
	}
	_, res.Name = filepath.Split(archive)

	if err != nil {
		return res, err
	}

	fi, err := os.Stat(archive)
	if err != nil {
		return res, errors.Wrapf(err, "stat file %v failed", archive)
	}
	res.Size = fi.Size()

	file := archive

	if plan.Encryption != nil {
		encryptedFile := fmt.Sprintf("%v.encrypted", archive)
		output, err := encrypt(archive, encryptedFile, plan, conf)
		if err != nil {
			return res, err
		} else {
			removeUnencrypted(archive, encryptedFile)
			file = encryptedFile
			log.WithField("plan", plan.Name).Infof("Encryption finished %v", output)
		}
	}

	if conf.StoragePath != "" && plan.Scheduler.Retention != 0 {
		localBackupOutput, err := localBackup(file, conf.StoragePath, mlog, plan)
		if err != nil {
			return res, err
		} else {
			log.WithField("plan", plan.Name).Infof("%v", localBackupOutput)
		}
	}

	if plan.SFTP != nil {
		sftpOutput, err := sftpUpload(file, plan)
		if err != nil {
			return res, err
		} else {
			log.WithField("plan", plan.Name).Info(sftpOutput)
		}
	}

	if plan.S3 != nil {
		s3Output, err := s3Upload(file, plan, conf.UseAwsCli)
		if err != nil {
			return res, err
		} else {
			log.WithField("plan", plan.Name).Infof("S3 upload finished %v", s3Output)
		}
	}

	if plan.GCloud != nil {
		gCloudOutput, err := gCloudUpload(file, plan)
		if err != nil {
			return res, err
		} else {
			log.WithField("plan", plan.Name).Infof("GCloud upload finished %v", gCloudOutput)
		}
	}

	if plan.Azure != nil {
		azureOutput, err := azureUpload(file, plan)
		if err != nil {
			return res, err
		} else {
			log.WithField("plan", plan.Name).Infof("Azure upload finished %v", azureOutput)
		}
	}

	if plan.Rclone != nil {
		rcloneOutput, err := rcloneUpload(file, plan)
		if err != nil {
			return res, err
		} else {
			log.WithField("plan", plan.Name).Infof("Rclone upload finished %v", rcloneOutput)
		}
	}

	output, err := cleanup(file, mlog)
	if err != nil {
		return res, err
	} else {
		log.WithField("plan", plan.Name).Infof("Clean up temp finished %v", output)
	}

	t2 := time.Now()
	res.Status = 200
	res.Duration = t2.Sub(t1)
	return res, nil
}

func cleanup(file string, mlog string) (string, error) {
	err := sh.Command("rm", file).Run()
	if err != nil {
		return "", errors.Wrapf(err, "remove file from %v failed", file)
	}
	// check if log file exists, is not always created
	if _, err := os.Stat(mlog); os.IsNotExist(err) {
		log.Debug("appears no log file was generated")
	} else {
		err = sh.Command("rm", mlog).Run()
		if err != nil {
			return "", errors.Wrapf(err, "remove file from %v failed", mlog)
		}
	}
	msg := fmt.Sprintf("Temp folder cleanup finished, `%v` is removed.", file)
	return msg, nil
}
