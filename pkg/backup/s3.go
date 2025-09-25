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

func s3Upload(file string, plan config.Plan, useAwsCli bool, localDir string) (string, error) {

	s3Url, err := url.Parse(plan.S3.URL)

	if err != nil {
		return "", errors.Wrapf(err, "invalid S3 url for plan %v: %s", plan.Name, plan.S3.URL)
	}

	if plan.S3.Sync {
		if len(localDir) == 0 {
			return "", errors.Errorf("S3 sync enabled for plan %v but no local storage directory is configured", plan.Name)
		}

		if statErr := ensureDirExists(localDir); statErr != nil {
			return "", statErr
		}
	}

	if useAwsCli && strings.HasSuffix(s3Url.Hostname(), "amazonaws.com") {
		return awsUpload(file, plan, localDir)
	}

	return minioUpload(file, plan, localDir)
}

func ensureDirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "S3 sync source %v is not accessible", path)
	}

	if !info.IsDir() {
		return errors.Errorf("S3 sync source %v is not a directory", path)
	}

	return nil
}

func awsUpload(file string, plan config.Plan, localDir string) (string, error) {

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

	encrypt := ""
	if len(plan.S3.KmsKeyId) > 0 {
		encrypt = fmt.Sprintf(" --sse aws:kms --sse-kms-key-id %v", plan.S3.KmsKeyId)
	}

	storage := ""
	if len(plan.S3.StorageClass) > 0 {
		storage = fmt.Sprintf(" --storage-class %v", plan.S3.StorageClass)
	}
	upload := ""
	source := file
	if plan.S3.Sync {
		source = localDir
		destination := fmt.Sprintf("s3://%v/%v/", plan.S3.Bucket, plan.Name)
		upload = fmt.Sprintf("aws --quiet s3 sync %v %v --delete%v%v", localDir, destination, encrypt, storage)
	} else {
		fileName := filepath.Base(file)
		destination := fmt.Sprintf("s3://%v/%v", plan.S3.Bucket, fileName)
		upload = fmt.Sprintf("aws --quiet s3 cp %v %v%v%v", file, destination, encrypt, storage)
	}

	result, err := sh.Command("/bin/sh", "-c", upload).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if len(result) > 0 {
		output += strings.Replace(string(result), "\n", " ", -1)
	}
	if err != nil {
		return "", errors.Wrapf(err, "S3 uploading %v to %v/%v failed %v", source, plan.Name, plan.S3.Bucket, output)
	}

	if strings.Contains(output, "<ERROR>") {
		return "", errors.Errorf("S3 upload failed %v", output)
	}

	return strings.Replace(output, "\n", " ", -1), nil
}

func minioUpload(file string, plan config.Plan, localDir string) (string, error) {

	// Try the new mc alias set command first
	register := fmt.Sprintf("mc alias set %v %v %v %v --api %v",
		plan.Name, plan.S3.URL, plan.S3.AccessKey, plan.S3.SecretKey, plan.S3.API)

	result, err := sh.Command("/bin/sh", "-c", register).CombinedOutput()
	output := ""
	if len(result) > 0 {
		output = strings.Replace(string(result), "\n", " ", -1)
	}

	// If the new command fails, fallback to the old mc config host add
	if err != nil {
		registerFallback := fmt.Sprintf("mc config host add %v %v %v %v --api %v",
			plan.Name, plan.S3.URL, plan.S3.AccessKey, plan.S3.SecretKey, plan.S3.API)

		result, err = sh.Command("/bin/sh", "-c", registerFallback).CombinedOutput()
		if len(result) > 0 {
			output = strings.Replace(string(result), "\n", " ", -1)
		}
		if err != nil {
			return "", errors.Wrapf(err, "mc alias set and mc config host add both failed for plan %v: %s", plan.Name, output)
		}
	}

	if plan.S3.CreateBucketIfNeeded {
		err := minioCreateBucket(plan)
		if err != nil {
			return "", err
		}
	}

	upload := ""
	source := file
	if plan.S3.Sync {
		source = localDir
		destination := fmt.Sprintf("%v/%v/%v", plan.Name, plan.S3.Bucket, plan.Name)
		upload = fmt.Sprintf("mc --quiet mirror --overwrite --remove %v %v", localDir, destination)
	} else {
		fileName := filepath.Base(file)
		upload = fmt.Sprintf("mc --quiet cp %v %v/%v/%v",
			file, plan.Name, plan.S3.Bucket, fileName)
	}

	result, err = sh.Command("/bin/sh", "-c", upload).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	output = ""
	if len(result) > 0 {
		output = strings.Replace(string(result), "\n", " ", -1)
	}

	if err != nil {
		return "", errors.Wrapf(err, "S3 uploading %v to %v/%v failed %v", source, plan.Name, plan.S3.Bucket, output)
	}

	if strings.Contains(output, "<ERROR>") {
		return "", errors.Errorf("S3 upload failed %v", output)
	}

	return strings.Replace(output, "\n", " ", -1), nil
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
