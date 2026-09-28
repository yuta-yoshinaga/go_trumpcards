//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSjavs_UnmarshalLegacySessionWithCpuDifficulty(t *testing.T) {
	legacy := []byte(`{"pl":[{},{},{},{}],"cfg":{"cd":0},"ph":0}`)
	game := domain.NewDefaultSjavs()
	require.NoError(t, json.Unmarshal(legacy, game))
	require.Equal(t, domain.DefaultSjavsConfig(), game.GetConfig())
}
