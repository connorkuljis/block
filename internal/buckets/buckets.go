package buckets

import (
	"github.com/connorkuljis/block-cli/internal/tasks"
	"github.com/jmoiron/sqlx"
)

var BucketsSchema = `
	CREATE TABLE IF NOT EXISTS Buckets (
		bucket_id INTEGER PRIMARY KEY AUTOINCREMENT,
		bucket_name string
	);
`

type Bucket struct {
	BucketId   int64  `db:"bucket_id"`
	BucketName string `db:"bucket_name"`
	Tasks      []tasks.Task
}

func GetBucketByName(db *sqlx.DB, bucketName string) (Bucket, error) {
	var bucket Bucket
	q := `SELECT * FROM Buckets WHERE bucket_name = ?`

	err := db.Get(&bucket, q, bucketName)
	if err != nil {
		return bucket, err
	}

	tasks, err := tasks.GetTasksByBucketId(db, bucket.BucketId)
	if err != nil {
		return bucket, err
	}

	bucket.Tasks = tasks

	return bucket, nil
}
