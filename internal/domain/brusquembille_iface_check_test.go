//go:build !js || !wasm || classic

package domain_test

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

var _ interfaces.BrusquembilleGame = (*domain.Brusquembille)(nil)
