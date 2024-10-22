package utils

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"os"
)

var minioClient *minio.Client
var bucketName string
var filePrefix string

func InitMinioClient() {

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		endpoint = "10.160.160.137:9000"
	}

	accessKeyID := os.Getenv("MINIO_ACCESS_KEY_ID")
	if accessKeyID == "" {
		accessKeyID = "property"
	}
	secretAccessKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		secretAccessKey = "Property@123"
	}

	bucketName = os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		bucketName = "inzone-property"
	}

	filePrefix = os.Getenv("MINIO_FILE_PREFIX")
	if filePrefix == "" {
		filePrefix = "2024/09"
	}

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func GetFile(objectName string) (*minio.Object, error) {
	return minioClient.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
}

func UploadFile(objectName, contentType string, fileBuffer *bytes.Buffer) error {
	_, err := minioClient.PutObject(context.Background(), bucketName, objectName, fileBuffer, int64(fileBuffer.Len()), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func ListFiles() <-chan minio.ObjectInfo {
	objectCh := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		WithMetadata: true,
		Recursive:    true,
		Prefix:       filePrefix,
	})
	return objectCh
}
