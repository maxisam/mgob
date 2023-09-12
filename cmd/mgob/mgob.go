package main

import (
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/stefanprodan/mgob/pkg/api"
	"github.com/stefanprodan/mgob/pkg/backup"
	"github.com/stefanprodan/mgob/pkg/config"
	"github.com/stefanprodan/mgob/pkg/db"
	"github.com/stefanprodan/mgob/pkg/scheduler"
)

var (
	appConfig = &config.AppConfig{}
	modules   = &config.ModuleConfig{}
	name      = "mgob"
	version   = "v2.0.0-dev"
)

func beforeApp(c *cli.Context) error {
	level, err := log.ParseLevel(c.GlobalString("LogLevel"))
	if err != nil {
		log.Fatalf("unable to determine and set log level: %+v", err)
	}
	log.SetLevel(level)

	if c.GlobalBool("JSONLog") {
		// platforms such as Google StackDriver want logs to stdout
		log.SetOutput(os.Stdout)
		log.SetFormatter(&log.JSONFormatter{})
	}

	log.Debug("log level set to ", c.GlobalString("LogLevel"))

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Version = version
	app.Usage = "mongodb dockerized backup agent"
	app.Action = start
	app.Before = beforeApp
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "ConfigPath,c",
			Usage: "plan yml files dir",
			Value: "/config",
		},
		cli.StringFlag{
			Name:  "StoragePath,s",
			Usage: "backup storage",
			Value: "/storage",
		},
		cli.StringFlag{
			Name:  "TmpPath,t",
			Usage: "temporary backup storage",
			Value: "/tmp",
		},
		cli.StringFlag{
			Name:  "DataPath,d",
			Usage: "db dir",
			Value: "/data",
		},
		cli.IntFlag{
			Name:  "Port,p",
			Usage: "Port to bind the HTTP server on",
			Value: 8090,
		},
		cli.StringFlag{
			Name:  "Bind,b",
			Usage: "Host to bind the HTTP server on",
			Value: "",
		},
		cli.BoolFlag{
			Name:  "JSONLog,j",
			Usage: "logs in JSON format",
		},
		cli.StringFlag{
			Name:  "LogLevel,l",
			Usage: "logging threshold level: debug|info|warn|error|fatal|panic",
			Value: "info",
		},
	}
	app.Run(os.Args)
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %+v", message, err)
	}
}

func loadConfiguration(c *cli.Context) {
	appConfig.LogLevel = c.String("LogLevel")
	appConfig.JSONLog = c.Bool("JSONLog")
	appConfig.Port = c.Int("Port")
	appConfig.Host = c.String("Bind")
	appConfig.ConfigPath = c.String("ConfigPath")
	appConfig.StoragePath = c.String("StoragePath")
	appConfig.TmpPath = c.String("TmpPath")
	appConfig.DataPath = c.String("DataPath")
	appConfig.Version = version

	log.Infof("starting with config: %+v", appConfig)

	err := envconfig.Process(name, modules)
	handleErr(err, "Error processing environment configuration")

	appConfig.UseAwsCli = true
	appConfig.HasGpg = true
}

func start(c *cli.Context) error {
	log.Infof("mgob %v", version)

	// Load the configuration from the command-line flags and environment variables.
	loadConfiguration(c)

	// Check if mongodump is installed and print the version information.
	info, err := backup.CheckMongodump()
	handleErr(err, "Failed to check mongodump")
	log.Info(info)

	// Check if all required clients are installed.
	checkClients()

	// Load the backup plans from the configuration directory.
	plans, err := config.LoadPlans(appConfig.ConfigPath)
	handleErr(err, "Failed to load backup plans")

	// Open the database store for status information.
	store, err := db.Open(path.Join(appConfig.DataPath, "mgob.db"))
	handleErr(err, "Failed to open database store")
	defer store.Close()

	// Create a new status store for the scheduler.
	statusStore, err := db.NewStatusStore(store)
	handleErr(err, "Failed to create status store")

	// Create a new scheduler and start it.
	sch := scheduler.New(plans, appConfig, modules, statusStore)
	sch.Start()

	// Create a new HTTP server and start it in a separate goroutine.
	server := &api.HttpServer{
		Config:  appConfig,
		Modules: modules,
		Stats:   statusStore,
	}
	log.Infof("Starting HTTP server on port %v", appConfig.Port)
	go server.Start(appConfig.Version)

	// Wait for a SIGINT (Ctrl+C) or SIGTERM (docker stop) signal.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Infof("Shutting down (%v signal received)", sig)

	return nil
}

func checkClients() {
	checkClient("Minio Client", modules.MinioClient, backup.CheckMinioClient)
	checkClient("AWS CLI", modules.AWSClient, backup.CheckAWSClient)
	checkClient("GPG", modules.GnuPG, backup.CheckGpg)
	checkClient("Google Storage", modules.GCloudClient, backup.CheckGCloudClient)
	checkClient("Azure Storage", modules.AzureClient, backup.CheckAzureClient)
	checkClient("RClone", modules.RCloneClient, backup.CheckRCloneClient)
}

func checkClient(name string, enabled bool, checkFunc func() (string, error)) {
	if !enabled {
		log.Infof("%s is disabled.", name)
		disableConfig(name)
		return
	}

	info, err := checkFunc()
	if err != nil {
		if name == "AWS CLI" || name == "GPG" {
			log.Warn(err)
			disableConfig(name)
		} else {
			log.Fatal(err)
		}
	} else {
		log.Info(info)
	}
}

func disableConfig(name string) {
	switch name {
	case "AWS CLI":
		appConfig.UseAwsCli = false
	case "GPG":
		appConfig.HasGpg = false
	}
}
