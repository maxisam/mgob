package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type Plan struct {
	Name       string      `yaml:"name"`
	Target     Target      `yaml:"target"`
	Scheduler  Scheduler   `yaml:"scheduler"`
	Retry      Retry       `yaml:"retry"`
	Validation *Validation `yaml:"validation"`
	Encryption *Encryption `yaml:"encryption"`
	S3         *S3         `yaml:"s3"`
	GCloud     *GCloud     `yaml:"gcloud"`
	Rclone     *Rclone     `yaml:"rclone"`
	Azure      *Azure      `yaml:"azure"`
	SFTP       *SFTP       `yaml:"sftp"`
	SMTP       *SMTP       `yaml:"smtp"`
	Slack      *Slack      `yaml:"slack"`
	Team       *Team       `yaml:"team"`
}

type Validation struct {
	Database Target `yaml:"database"`
}

type Target struct {
	Database string `yaml:"database"`
	Host     string `yaml:"host"`
	Uri      string `yaml:"uri"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	NoGzip   bool   `yaml:"noGzip"`
	Params   string `yaml:"params"`
}

type Scheduler struct {
	Cron      string `yaml:"cron"`
	Retention int    `yaml:"retention"`
	Timeout   int    `yaml:"timeout"`
}

type Retry struct {
	Attempts      int     `yaml:"attempts"`
	BackoffFactor float32 `yaml:"backoffFactor"`
}

type Encryption struct {
	Gpg *Gpg `yaml:"gpg"`
}

type Gpg struct {
	KeyServer  string   `yaml:"keyServer"`
	Recipients []string `yaml:"recipients"`
	KeyFile    string   `yaml:"keyFile"`
}

type S3 struct {
	Bucket       string `yaml:"bucket"`
	AccessKey    string `yaml:"accessKey"`
	API          string `yaml:"api"`
	SecretKey    string `yaml:"secretKey"`
	URL          string `yaml:"url"`
	KmsKeyId     string `yaml:"kmsKeyId"`
	StorageClass string `yaml:"storageClass" validate:"omitempty,oneof=STANDARD REDUCED_REDUNDANCY STANDARD_IA ONE-ZONE_IA INTELLIGENT_TIERING GLACIER DEEP_ARCHIVE`
}

type GCloud struct {
	Bucket      string `yaml:"bucket"`
	KeyFilePath string `yaml:"keyFilePath"`
}

type Rclone struct {
	Bucket         string `yaml:"bucket"`
	ConfigFilePath string `yaml:"configFilePath"`
	ConfigSection  string `yaml:"configSection"`
}

type Azure struct {
	ContainerName    string `yaml:"containerName"`
	ConnectionString string `yaml:"connectionString"`
}

type SFTP struct {
	Dir        string `yaml:"dir"`
	Host       string `yaml:"host"`
	Password   string `yaml:"password"`
	PrivateKey string `yaml:"private_key"`
	Passphrase string `yaml:"passphrase"`
	Port       int    `yaml:"port"`
	Username   string `yaml:"username"`
}

type SMTP struct {
	WarnOnly           bool     `yaml:"warnOnly"`
	TlsEnabled         bool     `yaml:"tlsEnabled"`
	InsecureSkipVerify bool     `yaml:"insecureSkipVerify"`
	Server             string   `yaml:"server"`
	Port               string   `yaml:"port"`
	Password           string   `yaml:"password"`
	Username           string   `yaml:"username"`
	From               string   `yaml:"from"`
	To                 []string `yaml:"to"`
}
type Team struct {
	WebhookURL string `yaml:"webhookUrl"`
	WarnOnly   bool   `yaml:"warnOnly"`
	ThemeColor string `yaml:"themeColor"`
}
type Slack struct {
	URL      string `yaml:"url"`
	Channel  string `yaml:"channel"`
	Username string `yaml:"username"`
	WarnOnly bool   `yaml:"warnOnly"`
}

func LoadPlan(dir string, name string) (Plan, error) {
	plan := Plan{}
	planPath := ""
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.Contains(path, name+".yml") || strings.Contains(path, name+".yaml") {
			planPath = path
		}
		return nil
	})

	if err != nil {
		return plan, errors.Wrapf(err, "Reading from %v failed", dir)
	}

	if len(planPath) < 1 {
		return plan, errors.Errorf("Plan %v not found", name)
	}

	data, err := ioutil.ReadFile(planPath)
	if err != nil {
		return plan, errors.Wrapf(err, "Reading %v failed", planPath)
	}

	if err := yaml.Unmarshal(data, &plan); err != nil {
		return plan, errors.Wrapf(err, "Parsing %v failed", planPath)
	}
	_, filename := filepath.Split(planPath)
	plan.Name = strings.TrimSuffix(filename, filepath.Ext(filename))

	return plan, nil
}

func LoadPlans(dir string) ([]Plan, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if strings.Contains(path, "yml") || strings.Contains(path, "yaml") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrapf(err, "Reading from %v failed", dir)
	}

	plans := make([]Plan, 0)

	for _, path := range files {
		var plan Plan
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "Reading %v failed", path)
		}

		if err := yaml.Unmarshal(data, &plan); err != nil {
			return nil, errors.Wrapf(err, "Parsing %v failed", path)
		}
		_, filename := filepath.Split(path)
		plan.Name = strings.TrimSuffix(filename, filepath.Ext(filename))

		duplicate := false
		for _, p := range plans {
			if p.Name == plan.Name {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}

		plans = append(plans, plan)

	}
	if len(plans) < 1 {
		return nil, errors.Errorf("No backup plans found in %v", dir)
	}

	return plans, nil
}
