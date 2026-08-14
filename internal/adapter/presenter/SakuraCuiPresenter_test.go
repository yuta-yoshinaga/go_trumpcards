package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestSakuraCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := newSakuraForPresenter()
	// **点数の表示は配りに任せない。** 光札が手札に来るかは配り次第なので、
	// 確かめたい札を自分で足してから描画する。
	g.GetPlayer(0).AddCard(domain.NewCard(domain.SakuraCurtainMonth, domain.SakuraCurtainIndex, false))
	out := new(presenter.SakuraCuiPresenter).Output(g, nil)

	assert.NotEmpty(t, out)
	assert.Contains(t, out, i18n.T("sakura.helpTitle"))
	// 場札も人間の手札もインデックス付きで並ぶ。
	assert.Contains(t, out, "[0]")
	assert.Contains(t, out, "[6]", "手札 7 枚ぶんの番号が出ていない")
	// **点数が札に添えられている。** 合計で競うゲームなので、点が読めないと打てない。
	assert.Contains(t, out, "(20)", "光札の点数が出ていない")
	assert.Contains(t, out, "(1)", "カス札の点数が出ていない")
}

// ラベルが生キーのまま出ていないこと (ロケールが引けているか)。
func TestSakuraCuiPresenter_Output_UsesTranslations(t *testing.T) {
	g := newSakuraForPresenter()
	out := new(presenter.SakuraCuiPresenter).Output(g, nil)
	assert.NotContains(t, out, "sakura.", "i18n キーが生のまま出ている")
	for _, key := range []string{"sakura.fieldLine", "sakura.playerLine", "sakura.promptPlay"} {
		assert.NotEqual(t, key, i18n.T(key), "%s の訳が無い", key)
	}
}

func TestSakuraCuiPresenter_Output_Error(t *testing.T) {
	g := newSakuraForPresenter()
	out := new(presenter.SakuraCuiPresenter).Output(g, errors.New("boom"))
	assert.Contains(t, out, "boom")
}

func TestSakuraCuiPresenter_Output_RoundEnd(t *testing.T) {
	g := domain.NewSakura(domain.NewSakuraPlayersForTable(2), domain.SakuraConfig{Seats: 2, Rounds: 2})
	g.Reset()
	for range 500 {
		if g.GetPhase() != domain.SakuraPhasePlay {
			break
		}
		if g.IsHumanTurn() {
			h := g.GetHint()
			require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
			continue
		}
		g.CpuPlay()
	}
	require.Equal(t, domain.SakuraPhaseRoundEnd, g.GetPhase())

	out := new(presenter.SakuraCuiPresenter).Output(g, nil)
	res := g.GetLastResult()
	require.NotNil(t, res)
	if res.Winner >= 0 {
		// 勝利行は素点と追加役の内訳つきで、合計点まで出す。
		seat := res.Seats[res.Winner]
		assert.Contains(t, out, strconv.Itoa(seat.Total))
		assert.Contains(t, out, strconv.Itoa(seat.CardPoints))
		assert.NotContains(t, out, i18n.T("sakura.roundDraw"))
	} else {
		assert.Contains(t, out, i18n.T("sakura.roundDraw"))
	}
}

func TestSakuraCuiPresenter_Output_GameEnd(t *testing.T) {
	g := domain.NewSakura(domain.NewSakuraPlayersForTable(2), domain.SakuraConfig{Seats: 2, Rounds: 1})
	g.Reset()
	sakuraPlayOut(t, g)

	out := new(presenter.SakuraCuiPresenter).Output(g, nil)
	if g.GetWinner() >= 0 {
		assert.Contains(t, out, "勝ち")
	} else {
		assert.Contains(t, out, i18n.T("sakura.gameDraw"))
	}
}

// 追加役が席の行に出る。
func TestSakuraCuiPresenter_ShowsBonuses(t *testing.T) {
	g := newSakuraForPresenter()
	before := new(presenter.SakuraCuiPresenter).Output(g, nil)
	assert.NotContains(t, before, i18n.T("sakura.bonus.sakuraSake"))

	g.GetPlayer(0).AddTaken(
		domain.NewCard(domain.SakuraCurtainMonth, domain.SakuraCurtainIndex, false),
		domain.NewCard(domain.SakuraMoonMonth, domain.SakuraMoonIndex, false),
	)
	after := new(presenter.SakuraCuiPresenter).Output(g, nil)
	assert.Contains(t, after, i18n.T("sakura.bonus.sakuraSake"))
	assert.Contains(t, after, "(30)")
}

func TestSakuraCuiPresenter_HintOutput(t *testing.T) {
	g := newSakuraForPresenter()
	out := new(presenter.SakuraCuiPresenter).HintOutput(g)
	assert.Contains(t, out, "HINT")
	// 理由は訳された文で出る (キーのままではない)。
	assert.True(t,
		strings.Contains(out, i18n.T("sakura.hintReasonCapture")) ||
			strings.Contains(out, i18n.T("sakura.hintReasonDiscard")),
		"ヒントの理由が訳されていない: %s", out)
}

func TestSakuraCuiPresenter_HintOutput_None(t *testing.T) {
	g := newSakuraForPresenter()
	for range 500 {
		if !g.IsHumanTurn() || g.GetPhase() != domain.SakuraPhasePlay {
			break
		}
		h := g.GetHint()
		require.NoError(t, g.PlayerPlay(h.CardIndex, h.FieldIndex))
	}
	require.False(t, g.IsHumanTurn() && g.GetPhase() == domain.SakuraPhasePlay)
	assert.Contains(t, new(presenter.SakuraCuiPresenter).HintOutput(g), i18n.T("sakura.hintNone"))
}

func TestSakuraCuiPresenter_ActionLogOutput(t *testing.T) {
	g := newSakuraForPresenter()
	sakuraPlayOut(t, g)
	out := new(presenter.SakuraCuiPresenter).ActionLogOutput(g)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "deal")
}
