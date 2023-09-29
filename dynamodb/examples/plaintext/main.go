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
	"github.com/sirupsen/logrus"
)

type Item struct {
	ID    string `json:"id" dynamodbav:"id"`
	Value string `json:"value" dynamodbav:"value"`
}

type KeyItem struct {
	ID string `json:"id" dynamodbav:"id"`
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	table := os.Getenv("TABLE_NAME")
	ddb := dynamodb.NewFromConfig(cfg)

	logrus.Infof("caching via DynamoDB table: %s", table)

	http.HandleFunc("/get", func(response http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")

		imap, err := attributevalue.MarshalMap(KeyItem{
			ID: key,
		})
		if err != nil {
			logrus.WithError(err).Error("failed to marshal map")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		logrus.Infof("imap = %v", imap)

		result, err := ddb.GetItem(context.Background(), &dynamodb.GetItemInput{
			TableName: aws.String(table),
			Key:       imap,
		})
		if err != nil {
			logrus.WithError(err).Error("failed to get item")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		item := Item{}
		err = attributevalue.UnmarshalMap(result.Item, &item)
		if err != nil {
			logrus.WithError(err).Error("failed to unmarshal map")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = response.Write([]byte(item.Value))
		if err != nil {
			logrus.WithError(err).Error("failed to write response")
		}
	})

	http.HandleFunc("/put", func(response http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")

		body, err := io.ReadAll(request.Body)
		if err != nil {
			logrus.WithError(err).Error("failed to read request body")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		avm, err := attributevalue.MarshalMap(Item{
			ID:    key,
			Value: string(body),
		})
		if err != nil {
			logrus.WithError(err).Error("failed to marshal map")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = ddb.PutItem(context.Background(), &dynamodb.PutItemInput{
			Item:      avm,
			TableName: aws.String(table),
		})
		if err != nil {
			logrus.WithError(err).Error("failed to put item")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	logrus.WithError(http.ListenAndServe(":8080", nil)).Fatal("failed to run server")
}
