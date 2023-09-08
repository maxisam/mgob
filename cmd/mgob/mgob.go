package main

import (
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"

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
	version   = "v1.9.0-dev"
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
		cli.StringFlag{
			Name:  "Archive,a",
			Usage: "specify archive location to use mongo-restore instead of mongo-dump",
			Value: "",
		},
	}
	app.Run(os.Args)
}

func start(c *cli.Context) error {
	log.Infof("mgob %v", version)

	appConfig.LogLevel = c.String("LogLevel")
	appConfig.JSONLog = c.Bool("JSONLog")
	appConfig.Port = c.Int("Port")
	appConfig.Host = c.String("Bind")
	appConfig.ConfigPath = c.String("ConfigPath")
	appConfig.StoragePath = c.String("StoragePath")
	appConfig.TmpPath = c.String("TmpPath")
	appConfig.DataPath = c.String("DataPath")
	appConfig.Archive = c.String("Archive")
	appConfig.Version = version

	log.Infof("starting with config: %+v", appConfig)

	err := envconfig.Process(name, modules)
	if err != nil {
		log.Fatal(err.Error())
	}

	appConfig.UseAwsCli = true
	appConfig.HasGpg = true

	info, err := backup.CheckMongodump()
	if err != nil {
		log.Fatal(err)
	}
	log.Info(info)

	checkClients()

	plans, err := config.LoadPlans(appConfig.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	store, err := db.Open(path.Join(appConfig.DataPath, "mgob.db"))
	if err != nil {
		log.Fatal(err)
	}
	statusStore, err := db.NewStatusStore(store)
	if err != nil {
		log.Fatal(err)
	}

	restoreDone := checkForRestore(plans)
	if restoreDone {
		return nil
	}

	sch := scheduler.New(plans, appConfig, modules, statusStore)
	sch.Start()

	server := &api.HttpServer{
		Config:  appConfig,
		Modules: modules,
		Stats:   statusStore,
	}
	log.Infof("starting http server on port %v", appConfig.Port)
	go server.Start(appConfig.Version)

	// wait for SIGINT (Ctrl+C) or SIGTERM (docker stop)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	log.Infof("shutting down %v signal received", sig)

	return nil
}

func checkForRestore(plans []config.Plan) bool {
	var restoreDone bool
	for _, plan := range plans {
		if strings.TrimSpace(plan.Archive) == "" {
			continue
		}
		log.Info("Not-empty archive parameter found, starting Restore...")
		log.Debugf("Archive parameter: %v", plan.Archive)

		restoreCmd := backup.BuildRestoreCmd(plan.Archive, plan.Target, plan.Target)
		log.Infof("Running restore command with : %v", restoreCmd)
		output, err := sh.Command("/bin/sh", "-c", restoreCmd).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
		if err != nil {
			ex := ""
			if len(output) > 0 {
				ex = strings.Replace(string(output), "\n", " ", -1)
			}
			output = nil
			err = errors.Wrapf(err, "mongorestore log %v", ex)
			log.Errorf("Restore procedure failed with error: %v", err)
		} else {
			log.Debugf("Restore command output: %v", string(output))
			log.Info("Restore procedure finished successfully, shutting down")
		}
		restoreDone = true
	}
	return restoreDone
}

func checkClients() {
	if modules.MinioClient {
		info, err := backup.CheckMinioClient()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(info)
	} else {
		log.Info("Minio Client is disabled.")
	}

	if modules.AWSClient {
		info, err := backup.CheckAWSClient()
		if err != nil {
			log.Warn(err)
			appConfig.UseAwsCli = false
		}
		log.Info(info)
	} else {
		log.Info("AWS CLI is disabled.")
	}

	if modules.GnuPG {
		info, err := backup.CheckGpg()
		if err != nil {
			log.Warn(err)
			appConfig.HasGpg = false
		}
		log.Info(info)
	} else {
		log.Info("GPG is disabled.")
	}

	if modules.GCloudClient {
		info, err := backup.CheckGCloudClient()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(info)
	} else {
		log.Info("Google Storage is disabled.")
	}

	if modules.AzureClient {
		info, err := backup.CheckAzureClient()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(info)
	} else {
		log.Info("Azure Storage is disabled.")
	}

	if modules.RCloneClient {
		info, err := backup.CheckRCloneClient()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(info)
	} else {
		log.Info("RClone is disabled.")
	}
}
