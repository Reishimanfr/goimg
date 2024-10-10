package worker

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Worker struct {
	Ovh *session.Session
	Svc *s3.S3
}

// Initializes a new OVH worker
func New(accessKey, secretKey, region, endpoint string) (*Worker, error) {
	s, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	s3Client := s3.New(s)

	return &Worker{
		Ovh: s,
		Svc: s3Client,
	}, nil
}
