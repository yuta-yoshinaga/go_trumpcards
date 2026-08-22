//go:build !js || !wasm || extra4

package domain_test

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

var _ interfaces.PutGame = (*domain.Put)(nil)
