package backup

import (
	"context"
	"fmt"
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

// Responsible to restore backup from one source
// using mongorestore
// Responsibilities
// - Download backup from one source
// - ValidateBackup backup using mongorestore
// - Testing restoring using queries defined by plan
func ValidateBackup(archive string, plan config.Plan, backupResult map[string]string) (bool, error) {
	output, err := runRestore(archive, plan)
	if err != nil {
		return false, errors.Wrapf(err, "failed to restore backup")
	}
	client, ctx, err := getMongoClient(buildUri(plan.Validation.Database))
	defer dispose(client, ctx)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get mongo client")
	}
	collectionNames, err := getRestoreCollectionNames(plan.Validation.Database.Database, client)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get collection names")
	}
	_, err = runCheck(backupResult, collectionNames, output)
	if err != nil {
		return false, errors.Wrapf(err, "failed to run validation check")
	}
	err = cleanMongo(plan.Validation.Database.Database, client)
	if err != nil {
		return false, errors.Wrapf(err, "failed to clean mongo validation database")
	}

	return true, nil
}

func dispose(client *mongo.Client, ctx context.Context) {
	if err := client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func getRestoreCollectionNames(databaseName string, client *mongo.Client) ([]string, error) {
	collectionNames, err := client.Database(databaseName).ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get collection names in database %v", databaseName)
	}
	log.Infof("Validation: collection names %v", strings.Join(collectionNames, ","))
	return collectionNames, nil
}

func runCheck(backupResult map[string]string, collectionNames []string, output []byte) (int, error) {
	checkCount := 0
	for _, collectionName := range collectionNames {
		if _, ok := backupResult[collectionName]; ok {
			checkCount++
		} else {
			return 0, errors.New(fmt.Sprintf("Collection %v not found in backup", collectionName))
		}
	}
	return checkCount, nil
}

func runRestore(archive string, plan config.Plan) ([]byte, error) {
	restoreCmd := BuildRestoreCmd(archive, plan.Target, plan.Validation.Database)
	log.Infof("Validation: restore backup with : %v", restoreCmd)
	output, err := sh.Command("/bin/sh", "-c", restoreCmd).SetTimeout(time.Duration(plan.Scheduler.Timeout) * time.Minute).CombinedOutput()
	if err != nil {
		ex := ""
		if len(output) > 0 {
			ex = strings.Replace(string(output), "\n", " ", -1)
		}
		return nil, errors.Wrapf(err, "mongorestore log %v", ex)
	}

	log.Infof("restore output: %v", string(output))
	return output, nil
}

func cleanMongo(dbName string, client *mongo.Client) error {
	err := client.Database(dbName).Drop(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

func getMongoClient(uri string) (*mongo.Client, context.Context, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Infof("Validation: connect to %v", uri)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, ctx, errors.Wrapf(err, "failed to connect to mongo with uri %v", uri)
	}
	return client, ctx, nil
}
