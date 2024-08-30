package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"

	"github.com/stefanprodan/mgob/pkg/config"
)

func gCloudKeyFileAuth(keyFilePath string) error {
	if _, err := os.Stat(keyFilePath); err != nil {
		return err
	}
	register := fmt.Sprintf("gcloud auth activate-service-account --key-file=%v", keyFilePath)
	if _, err := sh.Command("/bin/sh", "-c", register).CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func gCloudUpload(file string, plan config.Plan) (string, error) {

	if len(plan.GCloud.KeyFilePath) > 0 {
		if err := gCloudKeyFileAuth(plan.GCloud.KeyFilePath); err != nil {
			return "", errors.Wrapf(err, "gcloud auth for plan %v failed", plan.Name)
		}
	}

	fileName := filepath.Base(file)

	upload := fmt.Sprintf("gsutil cp %v gs://%v/%v",
		file, plan.GCloud.Bucket, fileName)

	result, err := sh.Command("/bin/sh", "-c", upload).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	output := ""
	if len(result) > 0 {
		output = strings.Replace(string(result), "\n", " ", -1)
	}

	if err != nil {
		return "", errors.Wrapf(err, "GCloud uploading %v to gs://%v failed %v", file, plan.GCloud.Bucket, output)
	}

	if strings.Contains(output, "<ERROR>") {
		return "", errors.Errorf("GCloud upload failed %v", output)
	}

	return strings.Replace(output, "\n", " ", -1), nil
}
