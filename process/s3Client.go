package process

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Client struct {
	s3         *s3.S3
	bucketName string
}

func newS3Client(bucketName, url, accessKey, secretKey string) (*s3Client, error) {
	region := "fra1"
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(url),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	return &s3Client{
		bucketName: bucketName,
		s3:         s3.New(sess),
	}, nil

}

func (s *s3Client) GetFile(fileName string) ([]byte, error) {
	// Download the file from S3
	params := &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fileName),
	}

	resp, err := s.s3.GetObject(params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buffer := &bytes.Buffer{}
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
