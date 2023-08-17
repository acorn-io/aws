package acorn

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	acrnCfnClient "github.com/acorn-io/aws/utils/cdk-runner/pkg/aws/cloudformation"
	"github.com/acorn-io/aws/utils/cdk-runner/pkg/aws/utils"
	apiv1 "github.com/acorn-io/runtime/pkg/apis/api.acorn.io/v1"
	internalv1 "github.com/acorn-io/runtime/pkg/apis/internal.acorn.io/v1"
	kclient "github.com/acorn-io/runtime/pkg/k8sclient"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	stackNameEnvKey = "ACORN_EXTERNAL_ID"
	acornEventKey   = "ACORN_EVENT"
	acornNameKey    = "ACORN_NAME"
	acornProjectKey = "ACORN_PROJECT"
)

type StackWatcher struct {
	Context   context.Context
	CfnClient *cloudformation.Client
	KClient   k8sClient.WithWatch
	Stack     *Stack
}

type Stack struct {
	Name          string
	Transitioning bool
	CurrentCounts Counts
	PrevCounts    Counts
}

type Counts struct {
	ReadyCount      int
	FailedResources int
	TotalCount      int
	DeletedCount    int
}

func StartEventWatcher(ctx context.Context, stackName string) {
	client, err := acrnCfnClient.NewClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	k8sClient, err := kclient.Default()
	if err != nil {
		logrus.Fatal(err)
	}

	awsClient := client.Client

	if err := utils.WaitForClientRole(ctx); err != nil {
		logrus.Fatal(err)
	}

	sw := &StackWatcher{
		Stack: &Stack{
			Name:          stackName,
			CurrentCounts: Counts{},
			PrevCounts:    Counts{},
		},
		Context:   ctx,
		CfnClient: awsClient,
		KClient:   k8sClient,
	}

	if sw.Stack.Name == "" {
		logrus.Fatalf("Missing %s environment variable pointing to Cloud Formation StackName", stackNameEnvKey)
	}

	if err := sw.watchStack(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Stack %s is ready", sw.Stack.Name)

}

func (sw *StackWatcher) watchStack() error {

	for {
		select {
		case <-sw.Context.Done():
			return nil
		default:
			stackResources, err := sw.CfnClient.DescribeStackResources(sw.Context, &cloudformation.DescribeStackResourcesInput{
				StackName: aws.String(sw.Stack.Name),
			})
			if err != nil && strings.Contains(err.Error(), "does not exist") {
				logrus.Infof("Waiting for Stack %s to be created", sw.Stack.Name)
				time.Sleep(10 * time.Second)
				continue
			} else if err != nil {
				logrus.Error(err)
				time.Sleep(10 * time.Second)
				continue
			}

			sw.Stack.CurrentCounts.ReadyCount = 0
			sw.Stack.CurrentCounts.FailedResources = 0
			sw.Stack.CurrentCounts.DeletedCount = 0
			sw.Stack.CurrentCounts.TotalCount = len(stackResources.StackResources)
			for _, resource := range stackResources.StackResources {
				if resource.ResourceStatus == types.ResourceStatusCreateComplete || resource.ResourceStatus == types.ResourceStatusUpdateComplete {
					sw.Stack.CurrentCounts.ReadyCount++
				}
				if resource.ResourceStatus == types.ResourceStatusCreateFailed || resource.ResourceStatus == types.ResourceStatusUpdateFailed || resource.ResourceStatus == types.ResourceStatusDeleteFailed {
					sw.Stack.CurrentCounts.FailedResources++
				}
				if resource.ResourceStatus == types.ResourceStatusDeleteComplete {
					sw.Stack.CurrentCounts.DeletedCount++
				}
			}
			if sw.Stack.CurrentCounts.ReadyCount < sw.Stack.CurrentCounts.TotalCount || sw.Stack.CurrentCounts.FailedResources > 0 {
				sw.Stack.Transitioning = true
			}

			if sw.Stack.ready() && sw.Stack.Transitioning {
				sw.Stack.Transitioning = false
				sw.emit()
				sw.Stack.PrevCounts = sw.Stack.CurrentCounts
				continue
			} else if sw.Stack.ready() {
				continue
			}

			sw.emit()
			sw.Stack.PrevCounts = sw.Stack.CurrentCounts
			time.Sleep(30 * time.Second)
		}
	}
}

func (s *Stack) ready() bool {
	return s.CurrentCounts.TotalCount == s.CurrentCounts.ReadyCount
}

func (s *Stack) countsEqual() bool {
	return s.CurrentCounts.ReadyCount == s.PrevCounts.ReadyCount && s.CurrentCounts.FailedResources == s.PrevCounts.FailedResources && s.CurrentCounts.TotalCount == s.PrevCounts.TotalCount && s.CurrentCounts.DeletedCount == s.PrevCounts.DeletedCount
}

func (sw *StackWatcher) emit() {
	if sw.Stack.countsEqual() {
		return
	}

	e := &apiv1.Event{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Event",
			APIVersion: "api.acorn.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    os.Getenv(acornProjectKey),
			GenerateName: "se-",
		},

		Type:        getEventPhrase(sw.Stack.Transitioning),
		AppName:     os.Getenv(acornNameKey),
		Severity:    "info",
		Description: sw.Stack.message(),
		Resource: &internalv1.EventResource{
			Kind: "app",
			Name: os.Getenv(acornNameKey),
		},
	}
	if err := sw.KClient.Create(sw.Context, e); err != nil {
		logrus.Error(err)
	}

	fmt.Print(sw.Stack.message())
}

func getEventPhrase(t bool) string {
	event := os.Getenv(acornEventKey)
	if !t {
		return fmt.Sprintf("Service%sd", strings.ToTitle(event))
	}

	eventPhrase := map[string]string{
		"update": "ServiceUpdating",
		"create": "ServiceCreating",
		"delete": "ServiceDeleting",
	}
	return eventPhrase[event]
}

func (s *Stack) message() string {
	action := "Provisioning"
	suffix := ""
	if s.CurrentCounts.FailedResources > 0 {
		suffix = fmt.Sprintf(" (%d failed, stack: %s)", s.CurrentCounts.FailedResources, s.Name)
	}

	value := s.CurrentCounts.ReadyCount
	if os.Getenv(acornEventKey) == "delete" {
		action = "Deleting"
		value = s.CurrentCounts.DeletedCount
	}
	return fmt.Sprintf("CloudFormation: %s resources (%d/%d ready)%s\n", action, value, s.CurrentCounts.TotalCount, suffix)
}
