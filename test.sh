#bin/bash


# Used by fragmenter-sidecar
export DAEMON_URL=http://localhost:3000
export DAEMON_DATASET=test_dataset
export DAEMON_COMMIT=aj8WBYA0Waa4DqcjqGDFYxbYqL2Z04DK
export MINIO_URL=localhost:9000
export MINIO_ACCESS_KEY=iterum
export MINIO_SECRET_KEY=banaanappel
export MINIO_USE_SSL=false
export MINIO_OUTPUT_BUCKET=init-data
export MQ_BROKER_URL=amqp://iterum:sinaasappel@localhost:5672
export MQ_OUTPUT_QUEUE=output-queue

# Used by fragmenter AND fragmenter-sidecar
export FRAGMENTER_INPUT=./build/tf.sock
export FRAGMENTER_OUTPUT=./build/ff.sock

# Used by fragmenter
export ENC_FRAGMENT_SIZE_LENGTH=4


make build
fragmenter-sidecar

# python ./fragmenter/main.py