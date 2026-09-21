//go:build test

package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestCuiErrorBlockTranslatesDomainKeyParams(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })
	errors := []error{
		domain.NewDomainErrorCode(domain.ErrInvalidPlay, "vira.errBidMustOutrank", map[string]string{"bidKey": "vira.bidShort.gask"}),
		domain.NewDomainErrorCode(domain.ErrInvalidPlay, "quodlibet.errContractUnavailable", map[string]string{"contractKey": "quodlibet.contractName.badNeighbour"}),
		domain.NewDomainErrorCode(domain.ErrInvalidPlay, "contractrummy.errContractSlotSet", map[string]string{"slot": "1", "size": "3"}),
		domain.NewDomainErrorCode(domain.ErrInvalidPlay, "carioca.errContractSlotRun", map[string]string{"slot": "2", "size": "4"}),
	}
	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		var all strings.Builder
		for _, err := range errors {
			var b strings.Builder
			cuiErrorBlock(&b, err)
			text := b.String()
			all.WriteString(text)
			require.NotContains(t, text, "{{")
			require.NotContains(t, text, ".Key")
			require.NotContains(t, text, "vira.bidShort.gask")
			require.NotContains(t, text, "quodlibet.contractName.badNeighbour")
		}
		if lang == "ja" {
			require.Contains(t, all.String(), "ガスク")
			require.Contains(t, all.String(), "悪い隣人")
			require.Contains(t, all.String(), "セット")
			require.Contains(t, all.String(), "ラン")
			require.NotContains(t, all.String(), "Gask")
			require.NotContains(t, all.String(), "Bad Neighbour")
			require.NotContains(t, all.String(), "set of")
			require.NotContains(t, all.String(), "run of")
		} else {
			require.Contains(t, all.String(), "Gask")
			require.Contains(t, all.String(), "Bad Neighbour")
			require.Contains(t, all.String(), "set of")
			require.Contains(t, all.String(), "run of")
			require.NotContains(t, all.String(), "ガスク")
			require.NotContains(t, all.String(), "悪い隣人")
			require.NotContains(t, all.String(), "セット")
			require.NotContains(t, all.String(), "ラン")
		}
	}
}
