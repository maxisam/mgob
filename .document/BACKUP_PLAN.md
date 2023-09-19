# Backup Plan

## Standard Backup Plan Configuration

```yaml
scheduler:
  cron: "0 6,18 */1 * *" # run every day at 6:00 and 18:00 UTC
  retention: 14 # Retains 14 local backups
  timeout: 60 # Operation timeout: 60 minutes
target:
  host: "172.18.7.21" # Mongod IP or host name
  port: 27017 # Mongodb port
  database: "test" # Database name, leave blank to backup all databases
  username: "admin" # Username, leave blank if auth is not enabled
  password: "secret" # Password
  params: "--ssl --authenticationDatabase admin" # Additional mongodump params, leave blank if not needed
  noGzip: false # Disable gzip compression (false means compression is enabled)

retry:
  attempts: 3 # number of retries
  backoffFactor: 60 # backoff factor  * (2 ^ attemptCount) seconds

validation:
  database:
    host: "127.0.0.1"
    port: 27017
    noGzip: false
    database: test_restore # Database name for restore operation
# Encryption (optional)
encryption:
  # At the time being, only gpg asymmetric encryption is supported
  # Public key file or at least one recipient is mandatory
  gpg:
    # optional path to a public key file, only the first key is used.
    keyFile: /secret/mgob-key/key.pub
    # optional key server, defaults to hkps://keys.openpgp.org
    keyServer: hkps://keys.openpgp.org
    # optional list of recipients, they will be looked up on key server
    recipients:
      - example@example.com
# S3 upload (optional)
s3:
  url: "https://play.minio.io:9000"
  bucket: "backup"
  # accessKey and secretKey are optional for AWS, if your Docker image has awscli
  accessKey: "Q3AM3UQ867SPQQA43P2F"
  secretKey: "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
  # Optional, only used for AWS (when awscli is present)
  # The customer-managed AWS Key Management  Service (KMS) key ID that should be used to
  # server-side encrypt the backup in S3
  #kmsKeyId:
  # Optional, only used for AWS (when awscli is present)
  # Valid choices are: STANDARD | REDUCED_REDUNDANCY | STANDARD_IA  |  ONE-
  #     ZONE_IA  |  INTELLIGENT_TIERING  |  GLACIER | DEEP_ARCHIVE.
  # Defaults to 'STANDARD'
  #storageClass: STANDARD
  # For Minio and AWS use S3v4 for GCP use S3v2
  api: "S3v4"
# GCloud upload (optional)
gcloud:
  bucket: "backup"
  keyFilePath: /path/to/service-account.json
# Azure blob storage upload (optional)
azure:
  containerName: "backup"
  connectionString: "DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net"
# Rclone upload (optional)
rclone:
  bucket: "my-backup-bucket"
  # See https://rclone.org/docs/ for details on how to configure rclone
  configFilePath: /etc/rclone.conf
  configSection: "myrclonesection"
# SFTP upload (optional)
sftp:
  host: sftp.company.com
  port: 2022
  username: user
  password: secret
  # you can also specify path to a private key and a passphrase
  private_key: /etc/ssh/ssh_host_rsa_key
  passphrase: secretpassphrase
  # dir must exist on the SFTP server
  dir: backup
# Email notifications (optional)
smtp:
  server: smtp.company.com
  port: 465
  username: user
  password: secret
  from: mgob@company.com
  to:
    - devops@company.com
    - alerts@company.com
  # 'true' to notify only on failures
  warnOnly: false

# Slack notifications (optional)
slack:
  url: https://hooks.slack.com/services/xxxx/xxx/xx
  channel: devops-alerts
  username: mgob
  # 'true' to notify only on failures
  warnOnly: false
team:
  webhookUrl: https://xxx/services/xxxx/xxx/xx
  warnOnly: false
  themeColor: "#f6c344"
```

## Database connection (target)

### ReplicaSet

```yaml
target:
  host: "mongo-0.mongo.db,mongo-1.mongo.db,mongo-2.mongo.db"
  port: 27017
  database: "test"
```

### Sharded cluster with authentication and SSL example:

```yaml
target:
  host: "mongos-0.db,mongos-1.db"
  port: 27017
  database: "test"
  username: "admin"
  password: "secret"
  params: "--ssl --authenticationDatabase admin"
```

### Uri usage

After mongodb version 3.4.6

With uri being set, host/port/username/password/database will be ignored. [Read more](https://www.mongodb.com/docs/database-tools/mongodump/#std-option-mongodump.--uri)

```yaml
target:
  uri: "mongodb://admin:secret@localhost:27017/test?authSource=admin&ssl=true"
```
