package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestBriscolaConfig_Validate(t *testing.T) {
	if err := domain.DefaultBriscolaConfig().Validate(); err != nil {
		t.Fatal(err)
	}
}
