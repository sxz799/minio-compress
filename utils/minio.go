package utils

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
)

var minioClient *minio.Client
var bucketName = "inzone-property"

func InitMinioClient() {
	endpoint := "10.160.160.137:9000"
	accessKeyID := "property"
	secretAccessKey := "Property@123"
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
	_, err := minioClient.PutObject(context.Background(), bucketName, objectName, fileBuffer, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func ListFiles(prefix string) <-chan minio.ObjectInfo {
	objectCh := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		WithMetadata: true,
		Recursive:    true,
		Prefix:       prefix,
	})
	return objectCh
}
