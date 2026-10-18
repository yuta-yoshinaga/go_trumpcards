package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

type finishedGoStop struct{ *domain.GoStop }

func (g finishedGoStop) GetGameEndFlag() bool { return true }

type finishedSakura struct{ *domain.Sakura }

func (g finishedSakura) GetGameEndFlag() bool { return true }

type finishedHachiHachi struct{ *domain.HachiHachi }

func (g finishedHachiHachi) GetGameEndFlag() bool { return true }

func TestActionLogToJSONWithFaceIncludesHanafudaIdentity(t *testing.T) {
	card := domain.NewCard(1, 1, false)
	entry := &domain.ActionLogEntry{Cards: []*domain.Card{card}}
	out := actionLogToJSONWithFace([]*domain.ActionLogEntry{entry}, func(*domain.Card) *CardFace {
		return &CardFace{Month: 1, Kind: "bright", Deck: "hanafuda"}
	})
	var decoded controller.ActionLogWebOutput
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	require.Len(t, decoded.Entries, 1)
	require.Len(t, decoded.Entries[0].Cards, 1)
	assert.Equal(t, 1, decoded.Entries[0].Cards[0].Month)
	assert.Equal(t, "bright", decoded.Entries[0].Cards[0].Kind)
}

func TestCuiHanafudaCardLabelsFollowLanguage(t *testing.T) {
	original := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(original) })
	card := domain.NewCard(1, 1, false)
	for _, tc := range []struct{ lang, expectedMonth string }{{"ja", "松"}, {"en", "Pine"}} {
		i18n.SetLang(tc.lang)
		assert.Contains(t, koikoiCuiCardStr(card), tc.expectedMonth)
		assert.Contains(t, hachihachiCuiCardStr(card), tc.expectedMonth)
		assert.Contains(t, gostopCuiCardStr(card), tc.expectedMonth)
		assert.Contains(t, sakuraCuiCardStr(card), tc.expectedMonth)
		entry := &domain.ActionLogEntry{DetailParams: map[string]string{"card": "旧ラベル"}, Cards: []*domain.Card{card}}
		localized := actionLogWithCardLabels([]*domain.ActionLogEntry{entry}, koikoiCuiCardStr)
		assert.Equal(t, koikoiCuiCardStr(card), localized[0].DetailParams["card"])
		assert.Contains(t, localized[0].DetailParams["card"], tc.expectedMonth)
		assert.Equal(t, "旧ラベル", entry.DetailParams["card"], "the source detail params must remain unchanged")
	}
}

func TestActionLogWithCardLabelsSkipsNilEntriesAndAbsentCardParams(t *testing.T) {
	entry := &domain.ActionLogEntry{PlayerIdx: 0, Cards: []*domain.Card{domain.NewCard(1, 1, false)}}
	entries := actionLogWithCardLabels([]*domain.ActionLogEntry{nil, entry}, gostopCuiCardStr)
	require.Len(t, entries, 1)
	assert.NotNil(t, entries[0])
	assert.NotContains(t, entries[0].DetailParams, "card")
	assert.NotPanics(t, func() { actionLogToTextWithNames(entries, nil) })
}

func TestCuiHanafudaActionLogOutputFollowsLanguage(t *testing.T) {
	original := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(original) })

	// 1. Koi-Koi
	gKoi := domain.NewDefaultKoiKoi()
	gKoi.Reset()
	gKoi.SetCurrentTurn(0)
	require.NoError(t, gKoi.PlayerPlay(0, -1))
	pKoi := new(KoiKoiCuiPresenter)
	var koiCard *domain.Card
	for _, e := range gKoi.GetActionLog() {
		if e.ActionType == "play" && len(e.Cards) > 0 {
			koiCard = e.Cards[0]
			break
		}
	}
	require.NotNil(t, koiCard)

	// 2. Sakura
	gSakura := domain.NewDefaultSakura()
	gSakura.Reset()
	hSakura := gSakura.GetHint()
	require.NoError(t, gSakura.PlayerPlay(hSakura.CardIndex, hSakura.FieldIndex))
	pSakura := new(SakuraCuiPresenter)
	var sakuraCard *domain.Card
	for _, e := range gSakura.GetActionLog() {
		if e.ActionType == "play" && len(e.Cards) > 0 {
			sakuraCard = e.Cards[0]
			break
		}
	}
	require.NotNil(t, sakuraCard)

	// 3. Hachi-Hachi
	gHachi := domain.NewDefaultHachiHachi()
	gHachi.Reset()
	gHachi.SetCurrentTurn(0)
	require.NoError(t, gHachi.PlayerPlay(0, -1))
	pHachi := new(HachiHachiCuiPresenter)
	var hachiCard *domain.Card
	for _, e := range gHachi.GetActionLog() {
		if e.ActionType == "play" && len(e.Cards) > 0 {
			hachiCard = e.Cards[0]
			break
		}
	}
	require.NotNil(t, hachiCard)

	// 4. Go-Stop
	gGoStop := domain.NewDefaultGoStop()
	gGoStop.Reset()
	gGoStop.SetCurrentTurn(0)
	require.NoError(t, gGoStop.PlayerPlay(0, -1))
	pGoStop := new(GoStopCuiPresenter)
	var gostopCard *domain.Card
	for _, e := range gGoStop.GetActionLog() {
		if e.ActionType == "play" && len(e.Cards) > 0 {
			gostopCard = e.Cards[0]
			break
		}
	}
	require.NotNil(t, gostopCard)

	// These three presenters historically hide the transcript until game end.
	empty := i18n.T("cuiActionLogEmpty") + "\n"
	assert.Equal(t, empty, pSakura.ActionLogOutput(gSakura))
	assert.Equal(t, empty, pHachi.ActionLogOutput(gHachi))
	assert.Equal(t, empty, pGoStop.ActionLogOutput(gGoStop))
	finishedSakuraGame := finishedSakura{gSakura}
	finishedHachiGame := finishedHachiHachi{gHachi}
	finishedGoStopGame := finishedGoStop{gGoStop}

	// 日本語表示の検証 (解決後の札名が含まれること)
	i18n.SetLang("ja")
	assert.Contains(t, pKoi.ActionLogOutput(gKoi), koikoiCuiCardStr(koiCard))
	assert.Contains(t, pSakura.ActionLogOutput(finishedSakuraGame), sakuraCuiCardStr(sakuraCard))
	assert.Contains(t, pHachi.ActionLogOutput(finishedHachiGame), hachihachiCuiCardStr(hachiCard))
	assert.Contains(t, pGoStop.ActionLogOutput(finishedGoStopGame), gostopCuiCardStr(gostopCard))

	// 英語表示の検証 (解決後の英語札名が含まれ、元の固定日本語札名が残らないこと)
	i18n.SetLang("en")
	enKoi := pKoi.ActionLogOutput(gKoi)
	assert.Contains(t, enKoi, koikoiCuiCardStr(koiCard))
	assert.NotContains(t, enKoi, domain.HachiHachiCardLabel(koiCard))

	enSakura := pSakura.ActionLogOutput(finishedSakuraGame)
	assert.Contains(t, enSakura, sakuraCuiCardStr(sakuraCard))
	assert.NotContains(t, enSakura, domain.HachiHachiCardLabel(sakuraCard))

	enHachi := pHachi.ActionLogOutput(finishedHachiGame)
	assert.Contains(t, enHachi, hachihachiCuiCardStr(hachiCard))
	assert.NotContains(t, enHachi, domain.HachiHachiCardLabel(hachiCard))

	enGoStop := pGoStop.ActionLogOutput(finishedGoStopGame)
	assert.Contains(t, enGoStop, gostopCuiCardStr(gostopCard))
	assert.NotContains(t, enGoStop, domain.HachiHachiCardLabel(gostopCard))
}
