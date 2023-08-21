package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)

	bucket := aws.String(os.Getenv("BUCKET_NAME"))

	http.HandleFunc("/get", func(response http.ResponseWriter, request *http.Request) {
		key := aws.String(request.URL.Query().Get("key"))

		output, err := client.GetObject(context.Background(), &s3.GetObjectInput{
			Key:    key,
			Bucket: bucket,
		})
		if err != nil {
			log.Printf("failed to get object: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		value, err := io.ReadAll(output.Body)
		if err != nil {
			log.Printf("failed to read value body: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Write(value)
	})

	http.HandleFunc("/put", func(response http.ResponseWriter, request *http.Request) {
		key := aws.String(request.URL.Query().Get("key"))

		body, err := io.ReadAll(request.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = client.PutObject(context.Background(), &s3.PutObjectInput{Key: key, Bucket: bucket, Body: bytes.NewReader(body)})
		if err != nil {
			log.Printf("failed to put object: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
