package gen

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type Uploader struct {
	prefix string
	bucket string
	client *s3manager.Uploader
}

func NewUploader(bucket string, prefix string) (*Uploader, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	})
	if err != nil {
		return nil, err
	}
	return &Uploader{
		bucket: bucket,
		prefix: prefix,
		client: s3manager.NewUploader(sess),
	}, nil
}

func (u Uploader) Upload(assets []Asset) error {
	p := mpb.New(
		mpb.WithWidth(60),
		mpb.WithRefreshRate(100*time.Millisecond),
	)

	for _, asset := range assets {
		key := u.prefix + path.Base(asset.BuildPath())

		reader, size, err := read(asset.BuildPath())
		if err != nil {
			return err
		}
		bar := p.AddBar(size, mpb.BarStyle("[=>-]"),
			mpb.PrependDecorators(
				decor.OnComplete(
					decor.Name(fmt.Sprintf("  %-20v", asset.Name())),
					fmt.Sprintf("✅ %-20v", asset.Name()),
				),
			),
			mpb.AppendDecorators(
				decor.CountersKibiByte("% .2f / % .2f"),
			),
		)

		if err := u.upload(key, bar.ProxyReader(reader)); err != nil {
			return err
		}
		asset.SetS3Key(key)
	}

	p.Wait()
	return nil
}

func read(file string) (io.Reader, int64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, stat.Size(), nil
}

func (u Uploader) upload(key string, reader io.ReadCloser) error {
	_, err := u.client.Upload(&s3manager.UploadInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	return err
}
