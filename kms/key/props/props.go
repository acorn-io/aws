package props

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type KMSKeyStackProps struct {
	StackProps        awscdk.StackProps
	Tags              map[string]string      `json:"tags"`
	KeyName           string                 `json:"keyName"`
	AdminArn          string                 `json:"adminArn"`
	KeyAlias          string                 `json:"keyAlias"`
	Description       string                 `json:"description"`
	Enabled           bool                   `json:"enabled"`
	EnableKeyRotation bool                   `json:"enableKeyRotation"`
	KeySpec           string                 `json:"keySpec"`
	KeyUsage          string                 `json:"keyUsage"`
	PendingWindowDays int                    `json:"pendingWindowDays"`
	KeyPolicy         map[string]interface{} `json:"keyPolicy"`
	RemovalPolicy     string                 `json:"removalPolicy"`
}

// Source: https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2/awskms@v2.96.0#KeySpec
var validKeySpecsAndUsages = map[string][]awskms.KeyUsage{
	"SYMMETRIC_DEFAULT": {awskms.KeyUsage_ENCRYPT_DECRYPT},
	"RSA_2048":          {awskms.KeyUsage_ENCRYPT_DECRYPT, awskms.KeyUsage_SIGN_VERIFY},
	"RSA_3072":          {awskms.KeyUsage_ENCRYPT_DECRYPT, awskms.KeyUsage_SIGN_VERIFY},
	"RSA_4096":          {awskms.KeyUsage_ENCRYPT_DECRYPT, awskms.KeyUsage_SIGN_VERIFY},
	"ECC_NIST_P256":     {awskms.KeyUsage_SIGN_VERIFY},
	"ECC_NIST_P384":     {awskms.KeyUsage_SIGN_VERIFY},
	"ECC_NIST_P521":     {awskms.KeyUsage_SIGN_VERIFY},
	"ECC_SECG_P256K1":   {awskms.KeyUsage_SIGN_VERIFY},
	"HMAC_224":          {awskms.KeyUsage_GENERATE_VERIFY_MAC},
	"HMAC_256":          {awskms.KeyUsage_GENERATE_VERIFY_MAC},
	"HMAC_384":          {awskms.KeyUsage_GENERATE_VERIFY_MAC},
	"HMAC_512":          {awskms.KeyUsage_GENERATE_VERIFY_MAC},
}

func (ksp *KMSKeyStackProps) SetDefaults() {
	if ksp.KeyName == "" {
		ksp.KeyName = os.Getenv("ACORN_EXTERNAL_ID")
	}
	if ksp.KeySpec == "" {
		ksp.KeySpec = "SYMMETRIC_DEFAULT"
	}
	if ksp.KeyUsage == "" {
		ksp.KeyUsage = "ENCRYPT_DECRYPT"
	}
	if ksp.PendingWindowDays == 0 {
		ksp.PendingWindowDays = 30
	}
	if ksp.Description == "" {
		ksp.Description = "Acorn created KMS Key"
	}
	if ksp.RemovalPolicy == "" {
		ksp.RemovalPolicy = "DESTROY"
	}
}

func (ksp *KMSKeyStackProps) ValidateProps() error {
	var errs []error
	if len(ksp.AdminArn) > 0 {
		if _, err := arn.Parse(ksp.AdminArn); err != nil {
			errs = append(errs, fmt.Errorf("failed to parse adminArn: %w", err))
		}
	}
	if _, _, err := ksp.GetKeySpecAndUsage(); err != nil {
		errs = append(errs, err)
	}
	if ksp.PendingWindowDays < 7 || ksp.PendingWindowDays > 30 {
		errs = append(errs, fmt.Errorf("pendingWindowDays must be between 7 and 30 (inclusive)"))
	}
	if ksp.RemovalPolicy != "DESTROY" && ksp.RemovalPolicy != "RETAIN" {
		errs = append(errs, fmt.Errorf("removalPolicy must be either DESTROY or RETAIN"))
	}
	return errors.Join(errs...)
}

func (ksp *KMSKeyStackProps) GetKeySpecAndUsage() (awskms.KeySpec, awskms.KeyUsage, error) {
	var kmsUsage awskms.KeyUsage
	switch ksp.KeyUsage {
	case "ENCRYPT_DECRYPT":
		kmsUsage = awskms.KeyUsage_ENCRYPT_DECRYPT
	case "SIGN_VERIFY":
		kmsUsage = awskms.KeyUsage_SIGN_VERIFY
	case "GENERATE_VERIFY_MAC":
		kmsUsage = awskms.KeyUsage_GENERATE_VERIFY_MAC
	default:
		return "", "", fmt.Errorf("invalid key usage: %s", ksp.KeyUsage)
	}

	if usages, ok := validKeySpecsAndUsages[ksp.KeySpec]; ok {
		if slices.Contains(usages, kmsUsage) {
			return awskms.KeySpec(ksp.KeySpec), kmsUsage, nil
		}

		var supportedUsages []string
		for _, u := range usages {
			supportedUsages = append(supportedUsages, string(u))
		}
		return "", "", fmt.Errorf("invalid key usage %s for key spec: %s, supported usages: %s", ksp.KeyUsage, ksp.KeySpec, strings.Join(supportedUsages, ", "))
	}
	return "", "", fmt.Errorf("invalid key spec %s, supported key specs: %s", ksp.KeySpec, strings.Join(maps.Keys(validKeySpecsAndUsages), ", "))
}
