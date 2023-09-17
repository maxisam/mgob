# Restore

There are couple of ways to restore a backup.

## Use MongoDB tools

Browse `mgob-host:8090/storage` to identify the backup you want to restore.
Login to your MongoDB server and download the archive using `curl` and restore the backup with `mongorestore` command line.

```bash
curl -o /tmp/mongo-test-1494056760.gz http://mgob-host:8090/storage/mongo-test/mongo-test-1494056760.gz
mongorestore --gzip --archive=/tmp/mongo-test-1494056760.gz --drop
```

## Use mgob container

You can also restore a backup from within mgob container.
Exec into mgob, identify the backup you want to restore and use `mongorestore` to connect to your MongoDB server.

```bash
docker exec -it mgob sh
ls /storage/mongo-test
mongorestore --gzip --archive=/storage/mongo-test/mongo-test-1494056760.gz --host mongohost:27017 --drop
```

## Use on-demand api

Read more about [On-Demand Restoration](./ON_DEMAND_OPERATION.md#on-demand-restoration)
