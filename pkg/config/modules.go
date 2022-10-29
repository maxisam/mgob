package config

type ModuleConfig struct {
	AWSClient    bool `envconfig:"en_aws_cli" default:"false"`
	AzureClient  bool `envconfig:"en_azure" default:"true"`
	GCloudClient bool `envconfig:"en_gcloud" default:"true"`
	GnuPG        bool `envconfig:"en_gpg" default:"true"`
	MinioClient  bool `envconfig:"en_minio" default:"true"`
	RCloneClient bool `envconfig:"en_rclone" default:"true"`
}
