package model

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3/s3manager"
    "os"
    "github.com/satori/go.uuid"
    "fmt"
 )
 
 func S3Upload(fileToUpload string) {
    sess := session.New(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("aws/profile", "default"),
	})

	bucketName := "surgenews"
	keyName := uuid.NewV4().String() + ".3gp"

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
	     u.PartSize = 32 * 1024 * 1024 // 64MB per part
	})

	file,_:=os.Open(fileToUpload)
	ct:= "audio/3gpp2"
	upParams := &s3manager.UploadInput{
	    Bucket: &bucketName,
	    Key:    &keyName,
	    Body:   file,
	    ContentType: &ct,
	}
	
	// Perform upload with options different than the those in the Uploader.
	result, err := uploader.Upload(upParams, func(u *s3manager.Uploader) {
		u.LeavePartsOnError = true  
	})
	fmt.Println("here")
	if err == nil {
		fmt.Printf("%v+", result)
	}
	fmt.Println("there")
	fmt.Println(err)
 }