package backup

import (
	"github.com/pkg/errors"
	"github.com/stefanprodan/mgob/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
)

func Shutdown(plan config.Plan) (bool, error) {
	client, ctx, err := getMongoClient(BuildUri(plan.Target))
	defer dispose(client, ctx)
	if err != nil {
		return false, errors.Wrapf(err, "failed to get mongo client")
	}
	result := client.Database("admin").RunCommand(ctx, bson.D{{"shutdown", 1}, {"timeoutSecs", plan.Target.ShutdownTimeoutSecs}})
	if result.Err() != nil {
		return false, errors.Wrapf(result.Err(), "failed to shutdown mongo database")
	}
	return true, nil
}
