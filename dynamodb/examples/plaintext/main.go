package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/jsii-runtime-go"
)

type Item struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	table := os.Getenv("TABLE_NAME")
	ddb := dynamodb.NewFromConfig(cfg)

	http.HandleFunc("/get", func(response http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")

		imap, err := attributevalue.MarshalMap(Item{
			ID: key,
		})
		if err != nil {
			log.Printf("failed to marshal map: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		result, err := ddb.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName: jsii.String(table),
			Key:       imap,
		})
		if err != nil {
			log.Printf("failed to get item: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		item := Item{}
		err = attributevalue.UnmarshalMap(result.Item, item)
		if err != nil {
			log.Printf("failed to unmarshal map: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = response.Write([]byte(item.Value))
		if err != nil {
			log.Printf("failed to write response: %v", err)
		}
	})

	http.HandleFunc("/put", func(response http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			log.Printf("failed to read request body: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		avm, err := attributevalue.MarshalMap(Item{
			ID:    key,
			Value: string(body),
		})
		if err != nil {
			log.Printf("failed to marshal map: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = ddb.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:      avm,
			TableName: aws.String(table),
		})
		if err != nil {
			log.Printf("failed to put item: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
