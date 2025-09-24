package backup

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/codeskyblue/go-sh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stefanprodan/mgob/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ValidateBackup(archive string, plan config.Plan, backupResult map[string]string) (bool, error) {
	output, err := RunRestore(archive, plan)
	if err != nil {
		log.WithField("plan", plan.Name).Error("Validation: Failed to execute restore command. restore failed, cleaning up")
		return false, errors.Wrapf(err, "failed to execute restore command")
	}
	if err := CheckIfAnyFailure(string(output)); err != nil {
		return false, errors.Wrapf(err, "failed to restore backup")
	}
	client, ctx, err := GetMongoClient(BuildUri(plan.Validation.Database))
	if err != nil {
		return false, errors.Wrapf(err, "failed to get mongo client")
	}
	defer Dispose(client, ctx)
	collectionNames, viewNames, err := getRestoreCollectionNames(plan.Validation.Database.Database, client)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get collection names")
	}
	if err = checkRetoreDatabase(backupResult, collectionNames, viewNames); err != nil {
		return false, errors.Wrapf(err, "failed to run validation check against restore database")
	}
	if err = cleanMongo(plan.Validation.Database.Database, client); err != nil {
		return false, errors.Wrapf(err, "failed to clean mongo validation database")
	}

	return true, nil
}

func CheckIfAnyFailure(output string) error {
	numberOfFailedCaptureRegex := `(\d+)\sdocument\(s\)\srestored\ssuccessfully\.\s(\d+)\sdocument\(s\)\sfailed`
	reg := regexp.MustCompile(numberOfFailedCaptureRegex)
	matches := reg.FindStringSubmatch(output)
	if reg.NumSubexp() == 2 {
		if matches[1] == "0" || matches[2] != "0" {
			return errors.New(fmt.Sprintf("mongorestore failed with %v failed documents", matches[1]))
		}
	}
	return nil
}

func Dispose(client *mongo.Client, ctx context.Context) {
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func getRestoreCollectionNames(databaseName string, client *mongo.Client) ([]string, map[string]struct{}, error) {
	collectionNames, err := client.Database(databaseName).ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get collection names in database %v", databaseName)
	}
	log.Infof("Validation: collection names %v", strings.Join(collectionNames, ","))
	viewNames, err := client.Database(databaseName).ListCollectionNames(context.TODO(), bson.M{"type": "view"})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get view names in database %v", databaseName)
	}
	viewMap := make(map[string]struct{}, len(viewNames))
	if len(viewNames) > 0 {
		log.Infof("Validation: view names %v", strings.Join(viewNames, ","))
	}
	for _, viewName := range viewNames {
		viewMap[viewName] = struct{}{}
	}
	return collectionNames, viewMap, nil
}

func checkRetoreDatabase(backupResult map[string]string, collectionNames []string, viewNames map[string]struct{}) error {
	skipCollections := map[string]struct{}{
		"system.views": {},
	}
	for viewName := range viewNames {
		skipCollections[viewName] = struct{}{}
	}

	restoredCollections := make(map[string]struct{})
	for _, collectionName := range collectionNames {
		if _, skip := skipCollections[collectionName]; skip {
			log.Debugf("Validation: skipping view collection %v during restore check", collectionName)
			continue
		}
		restoredCollections[collectionName] = struct{}{}
		if _, ok := backupResult[collectionName]; !ok {
			return errors.New(fmt.Sprintf("Collection %v not found in backup", collectionName))
		}
	}

	backupCollections := make(map[string]struct{})
	for collectionName := range backupResult {
		if _, skip := skipCollections[collectionName]; skip {
			continue
		}
		backupCollections[collectionName] = struct{}{}
		if _, ok := restoredCollections[collectionName]; !ok {
			return errors.New(fmt.Sprintf("Collection %v from backup not found in restored database", collectionName))
		}
	}

	if len(restoredCollections) != len(backupCollections) {
		return errors.New(fmt.Sprintf("Backup collection count: %v and restore collection count: %v are not the same", len(backupCollections), len(restoredCollections)))
	}

	return nil
}

func RunRestore(archive string, plan config.Plan) ([]byte, error) {
	restoreCmd := BuildRestoreCmd(archive, plan.Target, plan.Validation.Database)
	log.WithField("plan", plan.Name).Infof("Validation: restore backup with : %v", restoreCmd)
	output, err := sh.Command("/bin/sh", "-c", restoreCmd).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return nil, errors.Wrapf(err, "mongorestore log %v", ex)
	}
	log.WithField("plan", plan.Name).Debugf("restore output: %v", string(output))
	return output, nil
}

func cleanMongo(dbName string, client *mongo.Client) error {
	err := client.Database(dbName).Drop(context.TODO())
	if err != nil {
		return errors.Wrapf(err, "failed to drop database %v", dbName)
	}
	return nil
}

func GetMongoClient(uri string) (*mongo.Client, context.Context, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Infof("Validation: connect to %v", uri)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, ctx, errors.Wrapf(err, "failed to connect to mongo with uri %v", uri)
	}
	return client, ctx, nil
}
