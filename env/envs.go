package env

import "os"

// DaemonURL is the remote url at which an idv daemon can be contacted
var DaemonURL = os.Getenv("DAEMON_URL")

// DaemonDataset is the name of the dataset used by this pipeline run
var DaemonDataset = os.Getenv("DAEMON_DATASET")

// DaemonCommitHash is the hash of the commit used for this pipeline run
var DaemonCommitHash = os.Getenv("DAEMON_COMMIT")

// MinioURL is the url at which the minio client can be reached
var MinioURL = os.Getenv("MINIO_URL")

// MinioAccessKey is the access key for minio
var MinioAccessKey = os.Getenv("MINIO_ACCESS_KEY")

// MinioSecretKey is the secret access key for minio
var MinioSecretKey = os.Getenv("MINIO_SECRET_KEY")

// MinioUseSSL is a string val denoting whether minio client uses SSL
var MinioUseSSL = os.Getenv("MINIO_USE_SSL")

// MinioTargetBucket is the bucket to which storage should go
var MinioTargetBucket = os.Getenv("MINIO_OUTPUT_BUCKET")

// MQBrokerURL is the url at which we can reach the message queueing system
var MQBrokerURL = os.Getenv("MQ_BROKER_URL")

// MQOutputQueue is the queue into which we push the remote fragment descriptions
var MQOutputQueue = os.Getenv("MQ_OUTPUT_QUEUE")
