package elasticache

import (
	"os"
	"testing"
)

func TestResourceID(t *testing.T) {
	err := os.Setenv("ACORN_EXTERNAL_ID", "totally-real-and-cool-external-id-123")
	if err != nil {
		t.Fatal(err)
	}

	id := ResourceID("Redis", "Sng")
	if len(*id) == 0 || len(*id) > 40 {
		t.Errorf("invalid ID %s", *id)
	}

	idAgain := ResourceID("Redis", "Sng")
	if *id != *idAgain {
		t.Error("expected matching IDs")
	}
}
