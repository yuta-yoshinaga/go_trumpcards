package presenter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestWebRankingPlayerNamesAreLocalized(t *testing.T) {
	i18n.SetLang("ja")
	t.Cleanup(func() { i18n.SetLang("ja") })

	tests := []struct {
		name string
		ja   func() string
		en   func() string
	}{
		{
			name: "BigTwo",
			ja: func() string {
				players := []*domain.BigTwoPlayer{domain.NewBigTwoPlayer(true), domain.NewBigTwoPlayer(false)}
				players[0].SetRank(1)
				players[1].SetRank(2)
				return (&BigTwoWebPresenter{}).buildResultMessage(domain.NewBigTwo(domain.NewTrumpCards(0), players, domain.DefaultBigTwoConfig()))
			},
			en: func() string {
				players := []*domain.BigTwoPlayer{domain.NewBigTwoPlayer(true), domain.NewBigTwoPlayer(false)}
				players[0].SetRank(1)
				players[1].SetRank(2)
				return (&BigTwoWebPresenter{}).buildResultMessage(domain.NewBigTwo(domain.NewTrumpCards(0), players, domain.DefaultBigTwoConfig()))
			},
		},
		{
			name: "President",
			ja: func() string {
				players := []*domain.PresidentPlayer{domain.NewPresidentPlayer(true), domain.NewPresidentPlayer(false)}
				players[0].SetRank(1)
				players[1].SetRank(2)
				return (&PresidentWebPresenter{}).buildResultMessage(domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig()))
			},
			en: func() string {
				players := []*domain.PresidentPlayer{domain.NewPresidentPlayer(true), domain.NewPresidentPlayer(false)}
				players[0].SetRank(1)
				players[1].SetRank(2)
				return (&PresidentWebPresenter{}).buildResultMessage(domain.NewPresident(domain.NewTrumpCards(0), players, domain.DefaultPresidentConfig()))
			},
		},
		{
			name: "Shithead",
			ja: func() string {
				players := []*domain.ShitheadPlayer{domain.NewShitheadPlayer(true), domain.NewShitheadPlayer(false)}
				players[0].SetRank(1)
				players[1].SetRank(2)
				return (&ShitheadWebPresenter{}).buildResultMessage(domain.NewShithead(domain.NewTrumpCards(0), players, domain.DefaultShitheadConfig()))
			},
			en: func() string {
				players := []*domain.ShitheadPlayer{domain.NewShitheadPlayer(true), domain.NewShitheadPlayer(false)}
				players[0].SetRank(1)
				players[1].SetRank(2)
				return (&ShitheadWebPresenter{}).buildResultMessage(domain.NewShithead(domain.NewTrumpCards(0), players, domain.DefaultShitheadConfig()))
			},
		},
		{
			name: "KoiKoi",
			ja: func() string {
				players := []*domain.KoiKoiPlayer{domain.NewKoiKoiPlayer(true), domain.NewKoiKoiPlayer(false)}
				players[0].AddScore(1)
				players[1].AddScore(2)
				return (&KoiKoiWebPresenter{}).buildResultMessage(domain.NewKoiKoi(players, domain.DefaultKoiKoiConfig()))
			},
			en: func() string {
				players := []*domain.KoiKoiPlayer{domain.NewKoiKoiPlayer(true), domain.NewKoiKoiPlayer(false)}
				players[0].AddScore(1)
				players[1].AddScore(2)
				return (&KoiKoiWebPresenter{}).buildResultMessage(domain.NewKoiKoi(players, domain.DefaultKoiKoiConfig()))
			},
		},
		{
			name: "HachiHachi",
			ja: func() string {
				players := []*domain.HachiHachiPlayer{domain.NewHachiHachiPlayer(true), domain.NewHachiHachiPlayer(false)}
				players[0].AddScore(1)
				players[1].AddScore(2)
				return (&HachiHachiWebPresenter{}).buildResultMessage(domain.NewHachiHachi(players, domain.DefaultHachiHachiConfig()))
			},
			en: func() string {
				players := []*domain.HachiHachiPlayer{domain.NewHachiHachiPlayer(true), domain.NewHachiHachiPlayer(false)}
				players[0].AddScore(1)
				players[1].AddScore(2)
				return (&HachiHachiWebPresenter{}).buildResultMessage(domain.NewHachiHachi(players, domain.DefaultHachiHachiConfig()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i18n.SetLang("ja")
			ja := tt.ja()
			require.Contains(t, ja, "あなた")
			require.NotContains(t, ja, "You")

			i18n.SetLang("en")
			en := tt.en()
			require.Contains(t, en, "You")
			require.NotContains(t, en, "あなた")
		})
	}
}

func TestWebPlayerNameDoesNotAddTerminalStyling(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("ja") })
	i18n.SetLang("en")
	name := webPlayerName(true, 0)
	require.Equal(t, "You", strings.TrimSpace(name))
	require.NotContains(t, name, "\x1b[")
}
