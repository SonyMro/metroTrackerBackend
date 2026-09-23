package redis

import (
	"testing"
	"time"
)

func TestMapaCacheTTL(t *testing.T) {
	if mapaCacheTTL != time.Hour {
		t.Fatalf("mapaCacheTTL = %v, want %v", mapaCacheTTL, time.Hour)
	}
}
