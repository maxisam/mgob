package backup

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"

	"github.com/stefanprodan/mgob/pkg/config"
)

func s3Upload(file string, plan config.Plan, useAwsCli bool, storagePath string) (string, error) {

	s3Url, err := url.Parse(plan.S3.URL)

	if err != nil {
		return "", errors.Wrapf(err, "invalid S3 url for plan %v: %s", plan.Name, plan.S3.URL)
	}

	if useAwsCli && strings.HasSuffix(s3Url.Hostname(), "amazonaws.com") {
		return awsUpload(file, plan)
	}

	return minioUpload(file, plan, storagePath)
}

func awsUpload(file string, plan config.Plan) (string, error) {

	output := ""
	if len(plan.S3.AccessKey) > 0 && len(plan.S3.SecretKey) > 0 {
		// Let's use credentials given
		configure := fmt.Sprintf("aws configure set aws_access_key_id %v && aws configure set aws_secret_access_key %v",
			plan.S3.AccessKey, plan.S3.SecretKey)

		result, err := sh.Command("/bin/sh", "-c", configure).CombinedOutput()
		if len(result) > 0 {
			output += strings.Replace(string(result), "\n", " ", -1)
		}
		if err != nil {
			return "", errors.Wrapf(err, "aws configure for plan %v failed %s", plan.Name, output)
		}
	}

	fileName := filepath.Base(file)

	encrypt := ""
	if len(plan.S3.KmsKeyId) > 0 {
		encrypt = fmt.Sprintf(" --sse aws:kms --sse-kms-key-id %v", plan.S3.KmsKeyId)
	}

	storage := ""
	if len(plan.S3.StorageClass) > 0 {
		storage = fmt.Sprintf(" --storage-class %v", plan.S3.StorageClass)
	}

	upload := fmt.Sprintf("aws --quiet s3 cp %v s3://%v/%v%v%v",
		file, plan.S3.Bucket, fileName, encrypt, storage)

	result, err := sh.Command("/bin/sh", "-c", upload).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if len(result) > 0 {
		output += strings.Replace(string(result), "\n", " ", -1)
	}
	if err != nil {
		return "", errors.Wrapf(err, "S3 uploading %v to %v/%v failed %v", file, plan.Name, plan.S3.Bucket, output)
	}

	if strings.Contains(output, "<ERROR>") {
		return "", errors.Errorf("S3 upload failed %v", output)
	}

	return strings.Replace(output, "\n", " ", -1), nil
}

func minioUpload(file string, plan config.Plan, storagePath string) (string, error) {

	register := fmt.Sprintf("mc config host add %v %v %v %v --api %v",
		plan.Name, plan.S3.URL, plan.S3.AccessKey, plan.S3.SecretKey, plan.S3.API)

	result, err := sh.Command("/bin/sh", "-c", register).CombinedOutput()
	output := ""
	if len(result) > 0 {
		output = strings.Replace(string(result), "\n", " ", -1)
	}
	if err != nil {
		return "", errors.Wrapf(err, "mc config host for plan %v failed %s", plan.Name, output)
	}

	if plan.S3.CreateBucketIfNeeded {
		err := minioCreateBucket(plan)
		if err != nil {
			return "", err
		}
	}

	if plan.S3.Sync {
		syncOutput, err := minioMirror(plan, storagePath, time.Duration(plan.Scheduler.Timeout)*time.Minute)
		if err != nil {
			return "", err
		}
		return syncOutput, nil
	}

	fileName := filepath.Base(file)

	upload := fmt.Sprintf("mc --quiet cp %v %v/%v/%v",
		file, plan.Name, plan.S3.Bucket, fileName)

	result, err = sh.Command("/bin/sh", "-c", upload).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	output = ""
	if len(result) > 0 {
		output = strings.Replace(string(result), "\n", " ", -1)
	}

	if err != nil {
		return "", errors.Wrapf(err, "S3 uploading %v to %v/%v failed %v", file, plan.Name, plan.S3.Bucket, output)
	}

	if strings.Contains(output, "<ERROR>") {
		return "", errors.Errorf("S3 upload failed %v", output)
	}

	return strings.Replace(output, "\n", " ", -1), nil
}

func minioMirror(plan config.Plan, storagePath string, timeout time.Duration) (string, error) {
	if storagePath == "" {
		return "", errors.Errorf("S3 sync requested but storage path is empty for plan %v", plan.Name)
	}

	planDir := filepath.Join(storagePath, plan.Name)
	if _, err := os.Stat(planDir); err != nil {
		return "", errors.Wrapf(err, "S3 sync failed, could not access plan directory %v", planDir)
	}

	mirrorCmd := fmt.Sprintf("mc --quiet mirror --remove %v %v/%v", planDir, plan.Name, plan.S3.Bucket)

	result, err := sh.Command("/bin/sh", "-c", mirrorCmd).SetTimeout(timeout).CombinedOutput()
	output := ""
	if len(result) > 0 {
		output = strings.Replace(string(result), "\n", " ", -1)
	}

	message := fmt.Sprintf("mirror --remove %v %v/%v", planDir, plan.Name, plan.S3.Bucket)
	if output != "" {
		message = fmt.Sprintf("%s %s", message, output)
	}

	if err != nil {
		return "", errors.Wrapf(err, "S3 syncing %v to %v/%v failed %v", planDir, plan.Name, plan.S3.Bucket, output)
	}

	if strings.Contains(output, "<ERROR>") {
		return "", errors.Errorf("S3 sync failed %v", output)
	}

	return message, nil
}

func minioCreateBucket(plan config.Plan) error {
	listbucket := fmt.Sprintf("mc --quiet ls %v/%v", plan.Name, plan.S3.Bucket)

	_, err := sh.Command("/bin/sh", "-c", listbucket).CombinedOutput()

	if err == nil {
		// nothing to do
		return nil
	}

	// bucket does not seem to exist, try to create it
	createbucket := fmt.Sprintf("mc --quiet mb %v/%v", plan.Name, plan.S3.Bucket)

	result, err := sh.Command("/bin/sh", "-c", createbucket).CombinedOutput()

	if err != nil {
		output := ""
		if len(result) > 0 {
			output = strings.ReplaceAll(string(result), "\n", " ")
		}

		return errors.Wrapf(err, "S3 creation of bucket %v/%v failed %v", plan.Name, plan.S3.Bucket, output)
	}

	return nil
}
