package store

import (
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Client struct {
	Bucket string
	prefix string

	client *s3manager.Uploader
}

func New(bucket string, prefix string) (*Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		Bucket: bucket,
		prefix: prefix,
		client: s3manager.NewUploader(sess),
	}, nil
}

func (store Client) Exists(key string) bool {
	_, err := store.client.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(store.Bucket),
		Key:    aws.String(key),
	})
	return err != nil
}

func (store Client) Upload(file string) (string, error) {
	key := store.prefix + path.Base(file)

	if !store.Exists(key) {
		return key, nil
	}

	return key, nil
}

func (store Client) upload(key string, reader io.Reader) error {

	_, err := store.client.Upload(&s3manager.UploadInput{
		Bucket: aws.String(store.Bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	return err
}

func (store Client) BatchUpload(names []string, files []string) error {
	fmt.Println("publishing:")

	progress := NewProgressReader(files)
	for i, file := range files {
		key := store.prefix + path.Base(file)
		reader, err := progress.Next(names[i])
		if err != nil {
			return err
		}
		if err := store.upload(key, reader); err != nil {
			return err
		}
	}

	return nil
}
