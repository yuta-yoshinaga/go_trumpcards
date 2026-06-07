package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestBourreImplementsInterface(t *testing.T) {
	var _ interfaces.BourreGame = (*domain.Bourre)(nil)
}
