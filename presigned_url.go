package main

import(
	"time"
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error){
	client := s3.NewPresignClient(s3Client)
	input := s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}
	req, err := client.PresignGetObject(context.TODO(), &input, s3.WithPresignExpires(expireTime))
	if err != nil{
		return "", err
	}

	return req.URL, nil
}