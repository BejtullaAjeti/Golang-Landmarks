package handlers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"log"
	"mime/multipart"
	"net/http"
)

func uploadFileToS3(file multipart.File, fileName string) error {
	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(file); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	fileBytes := buffer.Bytes()
	fileType := http.DetectContentType(fileBytes)

	// Calculate content length
	contentLength := int64(len(fileBytes))

	input := &s3.PutObjectInput{
		Bucket:        aws.String("golang-backend-photos"),
		Key:           aws.String(fileName),
		Body:          bytes.NewReader(fileBytes),
		ContentLength: &contentLength,
		ContentType:   aws.String(fileType),
	}

	// Attempt to upload file to S3
	_, err := s3Client.PutObject(context.TODO(), input)
	if err != nil {
		// Log the error for debugging purposes
		log.Printf("Failed to upload file %s to S3: %v", fileName, err)
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return nil
}
