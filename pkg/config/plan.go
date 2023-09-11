package config

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
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
	StorageClass string `yaml:"storageClass" validate:"omitempty,oneof=STANDARD REDUCED_REDUNDANCY STANDARD_IA ONE-ZONE_IA INTELLIGENT_TIERING GLACIER DEEP_ARCHIVE"`
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

	// Set viper to read YAML configurations.
	viper.SetConfigType("yaml")

	// Set the paths to look for the config file in.
	viper.AddConfigPath(dir)

	// Set the name of the config file (without extension).
	viper.SetConfigName(name)

	// Try to read the config file.
	if err := viper.ReadInConfig(); err != nil {
		return plan, errors.Wrapf(err, "Reading %v failed", name)
	}

	// Unmarshal the read YAML into our struct.
	if err := viper.Unmarshal(&plan); err != nil {
		return plan, errors.Wrapf(err, "Parsing %v failed", name)
	}

	plan.Name = name

	return plan, nil
}

func LoadPlans(dir string) ([]Plan, error) {
	plans := make([]Plan, 0)

	// Use Go's standard lib to list all YAML files.
	files, err := filepath.Glob(filepath.Join(dir, "*.y*ml"))
	if err != nil {
		return nil, errors.Wrapf(err, "Reading from %v failed", dir)
	}

	for _, path := range files {
		var plan Plan

		viper.Reset() // Reset viper to ensure no overlap from previous configs.
		viper.SetConfigType("yaml")
		viper.SetConfigFile(path)

		if err := viper.ReadInConfig(); err != nil {
			return nil, errors.Wrapf(err, "Reading %v failed", path)
		}

		if err := viper.Unmarshal(&plan); err != nil {
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
