package minio

import (
	"strconv"

	"github.com/iterum-provenance/fragmenter/env"

	"github.com/minio/minio-go"
)

// Config is a structure holding all relevant information regarding the minio storage used by Iterum
type Config struct {
	TargetBucket string
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	PutOptions   minio.PutObjectOptions
	GetOptions   minio.GetObjectOptions
	Client       *minio.Client
}

// NewMinioConfig initiates a new minio configuration with all its necessary information
func NewMinioConfig(endpoint, accessKey, secretAccessKey, targetBucket string, useSSL bool) Config {
	putOptions := minio.PutObjectOptions{}
	getOptions := minio.GetObjectOptions{}
	return Config{
		targetBucket,
		endpoint,
		accessKey,
		secretAccessKey,
		useSSL,
		putOptions,
		getOptions,
		nil,
	}
}

// NewMinioConfigFromEnv uses environment variables to initialize a new MinioConfig
func NewMinioConfigFromEnv() (Config, error) {
	endpoint := env.MinioURL
	accessKeyID := env.MinioAccessKey
	secretAccessKey := env.MinioSecretKey
	useSSL, sslErr := strconv.ParseBool(env.MinioUseSSL)
	targetBucket := env.MinioTargetBucket
	return NewMinioConfig(endpoint, accessKeyID, secretAccessKey, targetBucket, useSSL), sslErr
}

// Connect tries to initialize the Client element of a minio config
func (mc *Config) Connect() error {
	client, err := minio.New(mc.Endpoint, mc.AccessKey, mc.SecretKey, mc.UseSSL)
	mc.Client = client
	return err
}

// IsConnected returns whether the client of a MinioConfig is initialized
func (mc Config) IsConnected() bool {
	return mc.Client != nil
}
