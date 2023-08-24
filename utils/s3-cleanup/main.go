package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	acornCf "github.com/acorn-io/aws/utils/cdk-runner/pkg/aws/cloudformation"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

func main() {
	stackName := os.Getenv("ACORN_EXTERNAL_ID")
	event := os.Getenv("ACORN_EVENT")

	if event != "delete" {
		logrus.Error("unexpected event")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// try to build the necessary clients
	cfClient, s3Client, err := buildClients(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	// get the stack
	stack, err := acornCf.GetStack(&acornCf.Client{Ctx: ctx, Client: cfClient}, stackName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			// no s3-cleanup is necessary
			return
		}

		logrus.Fatal(err)
	}

	// check for deletion protection
	if stack.DeletionProtection && os.Getenv(acornCf.DeletionProtectionEnvKey) == "true" {
		logrus.Warnf("Stack %s has deletion protection enabled. Buckets will not be emptied.", stackName)
		return
	}

	// get the stack resource so we can search for buckets
	resources, err := cfClient.DescribeStackResources(ctx, &cloudformation.DescribeStackResourcesInput{StackName: aws.String(stackName)})
	if err != nil {
		logrus.WithError(err).Fatal("failed to describe stack resources")
	}

	// try to empty the buckets
	err = emptyBuckets(ctx, s3Client, resources.StackResources)
	if err != nil {
		logrus.WithError(err).Fatal("failed to empty buckets")
	}
}

func emptyBuckets(ctx context.Context, client *s3.Client, resources []types.StackResource) error {
	for _, resource := range resources {
		if resource.ResourceType != nil && *resource.ResourceType == "AWS::S3::Bucket" {
			logrus.WithField("name", *resource.PhysicalResourceId).Info("emptying bucket")
			err := listAndDelete(ctx, client, resource, nil, nil)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func listAndDelete(ctx context.Context, client *s3.Client, bucket types.StackResource, nextKeyMarker *string, nextVersionIdMarker *string) error {
	// list objects
	listResult, err := client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{Bucket: bucket.PhysicalResourceId, KeyMarker: nextKeyMarker, VersionIdMarker: nextVersionIdMarker})
	if err != nil {
		return fmt.Errorf("failed to list bucket objects: %w", err)
	}

	// Delete all versions of listed objects
	for _, version := range listResult.Versions {
		_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket:    bucket.PhysicalResourceId,
			Key:       version.Key,
			VersionId: version.VersionId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}

	// Delete all deletion markers
	for _, marker := range listResult.DeleteMarkers {
		_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket:    bucket.PhysicalResourceId,
			Key:       marker.Key,
			VersionId: marker.VersionId,
		})
		if err != nil {
			return fmt.Errorf("failed to delete deletion marker: %w", err)
		}
	}

	if listResult.IsTruncated {
		// do it again if the list results were truncated
		return listAndDelete(ctx, client, bucket, listResult.NextKeyMarker, listResult.NextVersionIdMarker)
	}

	return nil
}

func buildClients(ctx context.Context) (*cloudformation.Client, *s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, nil, err
	}

	return cloudformation.NewFromConfig(cfg), s3.NewFromConfig(cfg), nil
}
