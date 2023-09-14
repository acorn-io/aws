package props

import (
	"strings"
	"testing"
)

func TestPropsValidation(t *testing.T) {
	tests := []struct {
		name        string
		props       KMSKeyStackProps
		errContains string
	}{
		{
			name: "valid",
			props: KMSKeyStackProps{
				AdminArn:          "arn:aws:iam:us-east-2:123456789012:root",
				KeySpec:           "RSA_3072",
				KeyUsage:          "ENCRYPT_DECRYPT",
				PendingWindowDays: 10,
				RemovalPolicy:     "DESTROY",
			},
		},
		{
			name: "invalid adminArn",
			props: KMSKeyStackProps{
				AdminArn:          "invalid",
				KeySpec:           "RSA_3072",
				KeyUsage:          "ENCRYPT_DECRYPT",
				PendingWindowDays: 10,
				RemovalPolicy:     "DESTROY",
			},
			errContains: "failed to parse adminArn",
		},
		{
			name: "invalid keySpec",
			props: KMSKeyStackProps{
				AdminArn:          "arn:aws:iam:us-east-2:123456789012:root",
				KeySpec:           "INVALID",
				KeyUsage:          "ENCRYPT_DECRYPT",
				PendingWindowDays: 10,
				RemovalPolicy:     "DESTROY",
			},
			errContains: "invalid key spec INVALID",
		},
		{
			name: "invalid keyUsage",
			props: KMSKeyStackProps{
				AdminArn:          "arn:aws:iam:us-east-2:123456789012:root",
				KeySpec:           "RSA_3072",
				KeyUsage:          "INVALID",
				PendingWindowDays: 10,
				RemovalPolicy:     "DESTROY",
			},
			errContains: "invalid key usage: INVALID",
		},
		{
			name: "keySpec does not support keyUsage",
			props: KMSKeyStackProps{
				AdminArn:          "arn:aws:iam:us-east-2:123456789012:root",
				KeySpec:           "SYMMETRIC_DEFAULT",
				KeyUsage:          "GENERATE_VERIFY_MAC",
				PendingWindowDays: 10,
				RemovalPolicy:     "DESTROY",
			},
			errContains: "invalid key usage GENERATE_VERIFY_MAC for key spec: SYMMETRIC_DEFAULT",
		},
		{
			name: "invalid pendingWindowDays",
			props: KMSKeyStackProps{
				AdminArn:          "arn:aws:iam:us-east-2:123456789012:root",
				KeySpec:           "RSA_3072",
				KeyUsage:          "ENCRYPT_DECRYPT",
				PendingWindowDays: 5,
				RemovalPolicy:     "DESTROY",
			},
			errContains: "pendingWindowDays must be between 7 and 30 (inclusive)",
		},
		{
			name: "invalid removalPolicy",
			props: KMSKeyStackProps{
				AdminArn:          "arn:aws:iam:us-east-2:123456789012:root",
				KeySpec:           "RSA_3072",
				KeyUsage:          "ENCRYPT_DECRYPT",
				PendingWindowDays: 10,
				RemovalPolicy:     "INVALID",
			},
			errContains: "removalPolicy must be either DESTROY or RETAIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.props.ValidateProps(); err != nil {
				if tt.errContains == "" {
					t.Errorf("unexpected error: %s", err)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err)
				}
			} else if tt.errContains != "" {
				t.Errorf("expected error to contain %q, got nil", tt.errContains)
			}
		})
	}
}
