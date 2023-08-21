package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	// Create a session that gets credential values from ~/.aws/credentials
	// and the default region from ~/.aws/config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := s3.New(sess)

	bucket := aws.String(os.Getenv("BUCKET_NAME"))

	http.HandleFunc("/get", func(response http.ResponseWriter, request *http.Request) {
		key := aws.String(request.URL.Query().Get("key"))

		output, err := svc.GetObject(&s3.GetObjectInput{
			Key:    key,
			Bucket: bucket,
		})

		if err != nil {
			log.Printf("failed to get object: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Write([]byte(output.String()))
	})

	http.HandleFunc("/put", func(response http.ResponseWriter, request *http.Request) {
		key := aws.String(request.URL.Query().Get("key"))

		body, err := io.ReadAll(request.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = svc.PutObject(&s3.PutObjectInput{Key: key, Bucket: bucket, Body: bytes.NewReader(body)})
		if err != nil {
			log.Printf("failed to put object: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
