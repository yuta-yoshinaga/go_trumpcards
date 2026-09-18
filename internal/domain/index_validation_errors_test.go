//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexValidationDomainErrorsHaveMessageCodes(t *testing.T) {
	tests := []struct {
		name    string
		indices []int
		code    string
	}{
		{name: "empty", indices: nil, code: "shared.errEmptyIndexList"},
		{name: "out of range", indices: []int{-1}, code: "shared.errCardIndexOutOfRange"},
		{name: "duplicate", indices: []int{1, 1}, code: "shared.errDuplicateCardIndex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateIndexList(tt.indices, 2)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidCard)
			de, ok := err.(*DomainError)
			require.True(t, ok, "expected DomainError, got %T", err)
			assert.Equal(t, tt.code, de.MessageCode())
			assert.Nil(t, de.MessageParams())
		})
	}

	assert.NoError(t, validateIndexList([]int{0, 1}, 2))
}

func TestIndexValidationCodeStillWrapsSentinel(t *testing.T) {
	err := validateIndexList([]int{2}, 2)
	assert.True(t, errors.Is(err, ErrInvalidCard))
}
