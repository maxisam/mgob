package backup

import (
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
)

func CheckMongodump() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "mongodump --version").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "mongodump failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}

func CheckMinioClient() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "mc --version").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "mc failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}

func CheckAWSClient() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "aws --version").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "aws failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}

func CheckGpg() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "gpg --version").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "gpg failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}

func CheckGCloudClient() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "gcloud --version").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "gcloud failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}

func CheckAzureClient() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "az --version | grep 'azure-cli'").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "az failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}

func CheckRCloneClient() (string, error) {
	output, err := sh.Command("/bin/sh", "-c", "rclone version | grep 'rclone'").CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return "", errors.Wrapf(err, "rclone failed %v", ex)
	}

	return strings.Replace(string(output), "\n", " ", -1), nil
}
