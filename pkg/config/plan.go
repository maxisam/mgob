package config

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
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
	Bucket               string `yaml:"bucket"`
	AccessKey            string `yaml:"accessKey"`
	API                  string `yaml:"api"`
	SecretKey            string `yaml:"secretKey"`
	URL                  string `yaml:"url"`
	KmsKeyId             string `yaml:"kmsKeyId"`
	StorageClass         string `yaml:"storageClass" validate:"omitempty,oneof=STANDARD REDUCED_REDUNDANCY STANDARD_IA ONE-ZONE_IA INTELLIGENT_TIERING GLACIER DEEP_ARCHIVE"`
	CreateBucketIfNeeded bool   `yaml:"createbucketifneeded"`
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
	PrivateKey string `yaml:"privateKey"`
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

	// Set the paths to look for the config file in.
	viper.AddConfigPath(dir)

	// Set the name of the config file (without extension).
	viper.SetConfigName(name)
	setupViperEnv(name)
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
	// Use Go's standard lib to list all YAML files.
	files, err := filepath.Glob(filepath.Join(dir, "*.y*ml"))
	if err != nil {
		return nil, errors.Wrapf(err, "Reading from %v failed", dir)
	}

	plans := make([]Plan, 0, len(files))
	names := make(map[string]bool)

	for _, path := range files {
		var plan Plan
		_, filename := filepath.Split(path)
		name := strings.TrimSuffix(filename, filepath.Ext(filename))

		if names[name] {
			continue // Skip duplicate plans
		}
		names[name] = true

		// Set viper to read YAML configurations.
		viper.Reset()
		viper.SetConfigFile(path)
		setupViperEnv(name)

		// Try to read the config file.
		if err := viper.ReadInConfig(); err != nil {
			return nil, errors.Wrapf(err, "Reading %v failed", path)
		}

		// Unmarshal the read YAML into our struct.
		if err := viper.Unmarshal(&plan); err != nil {
			return nil, errors.Wrapf(err, "Parsing %v failed", path)
		}

		plan.Name = name

		if log.IsLevelEnabled(logrus.DebugLevel) {
			planJSON, err := json.Marshal(plan)
			if err != nil {
				return nil, errors.Wrapf(err, "Marshaling %v failed", plan)
			}

			log.WithField("plan", plan.Name).Debugf("Loaded plan %v, plan JSON: %s", plan.Name, planJSON)
		}

		plans = append(plans, plan)
	}

	if len(plans) == 0 {
		return nil, errors.Errorf("No backup plans found in %v", dir)
	}

	return plans, nil
}

func setupViperEnv(planName string) {
	viper.SetConfigType("yaml")
	// set upper case plan name as env prefix
	envPrefix := strings.ReplaceAll(planName, "-", "_")
	viper.SetEnvPrefix(envPrefix + "_") // will be uppercased automatically
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
