#bin/bash

export DAEMON_URL=http://localhost:3000
export DAEMON_DATASET=test_dataset
export DAEMON_COMMIT=JlJZGoF04Q5RXayMmBTzE20uYKSQysZI
export MINIO_URL=localhost:9000
export MINIO_ACCESS_KEY=iterum
export MINIO_SECRET_KEY=banaanappel
export MINIO_USE_SSL=false
export MINIO_OUTPUT_BUCKET=init-data
export MQ_BROKER_URL=amqp://iterum:sinaasappel@localhost:5672
export MQ_OUTPUT_QUEUE=output-queue

make build
fragmenter-sidecar