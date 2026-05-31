package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestScopaImplementsInterface(t *testing.T) {
	var _ interfaces.ScopaGame = (*domain.Scopa)(nil)
}
