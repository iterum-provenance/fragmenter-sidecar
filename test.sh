
export ITERUM_NAME="fragmenter-sidecar"
export DATA_VOLUME_PATH="./build/"
export PIPELINE_HASH="asdh37dHsf8H3fSSd24HEe35g2d4h754"

export DAEMON_URL="http://localhost:3000"
export DAEMON_DATASET="test_dataset"
export DAEMON_COMMIT_HASH="97QVlptzj81zTuzJRhm1jUhqDi9rr1dF"

export MQ_BROKER_URL="amqp://iterum:sinaasappel@localhost:5672"
export MQ_INPUT_QUEUE="INVALID"
export MQ_OUTPUT_QUEUE="frag-out"

export MINIO_URL="localhost:9000"
export MINIO_ACCESS_KEY="iterum"
export MINIO_SECRET_KEY="banaanappel"
export MINIO_USE_SSL="false"
export MINIO_OUTPUT_BUCKET="frag-out"

export FRAGMENTER_INPUT="tf.sock"
export FRAGMENTER_OUTPUT="ff.sock"

make build
fragmenter-sidecar
