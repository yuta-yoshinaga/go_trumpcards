package presenter_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestBasraCuiPresenter_Output_PlayPhase(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	p := new(presenter.BasraCuiPresenter)
	out := p.Output(g, nil)
	assert.NotEmpty(t, out)
	assert.Contains(t, out, "[0]") // human indexed hand + indexed table
}

// **どの札で何を取れるかを見せる。**Web は選択中の札が捕獲できる場札を
// リングとチェックで示すのに、CUI はヒントを叩かない限り分からなかった (#4922)。
func TestBasraCuiPresenter_AnnotatesCaptureOptions(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetPhase(domain.BasraPhasePlay)
	g.SetCurrentTurn(0)

	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // 場の 5 を取れる
	human.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false)) // 何も取れない
	g.SetTableCards([]*domain.Card{
		domain.NewCard(domain.CardDesignClover, 5, false),
		domain.NewCard(domain.CardDesignDiamond, 3, false),
	})

	out := new(presenter.BasraCuiPresenter).Output(g, nil)
	// ♠5 は場[0] の ♣5 を取れる。
	assert.Contains(t, out, "[0]SPADE 5 → 場[0]")
	// 何も取れない札には注記が付かない (受け入れ条件2)。
	assert.NotContains(t, out, "[1]HEART 9 →")
}

// CPU の手番でも人間の手札の注記は出す (手札はもともと公開されている)。
// 一方、捕獲できる札が 1 枚も無ければ行そのものを出さない。
func TestBasraCuiPresenter_NoCaptureLineWhenNothingCanBeTaken(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetPhase(domain.BasraPhasePlay)
	g.SetCurrentTurn(0)

	human := g.GetPlayer(0)
	human.Reset()
	human.AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))
	g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignDiamond, 3, false)})

	assert.NotContains(t, new(presenter.BasraCuiPresenter).Output(g, nil), "→ 場")
}

func TestBasraCuiPresenter_Output_AllPhases(t *testing.T) {
	p := new(presenter.BasraCuiPresenter)

	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetCurrentTurn(0)
	assert.NotEmpty(t, p.Output(g, nil))

	g.SetPhase(domain.BasraPhaseGameEnd)
	assert.NotEmpty(t, p.Output(g, nil))

	// エラー出力。
	assert.NotEmpty(t, p.Output(g, errors.New("boom")))

	// 空テーブル表示。
	g.SetPhase(domain.BasraPhasePlay)
	g.SetTableCards(nil)
	assert.NotEmpty(t, p.Output(g, nil))
}

func TestBasraCuiPresenter_HintOutput(t *testing.T) {
	p := new(presenter.BasraCuiPresenter)

	t.Run("capture hint", func(t *testing.T) {
		g := domain.NewDefaultBasra()
		g.Reset()
		g.SetCurrentTurn(0)
		g.SetTableCards([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
		g.GetPlayer(0).Reset()
		g.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		assert.NotEmpty(t, p.HintOutput(g))
	})

	t.Run("no hint outside human turn", func(t *testing.T) {
		g := domain.NewDefaultBasra()
		g.Reset()
		g.SetCurrentTurn(1)
		assert.NotEmpty(t, p.HintOutput(g)) // hintNone message
	})
}

func TestBasraCuiPresenter_ActionLog(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	p := new(presenter.BasraCuiPresenter)
	assert.NotEmpty(t, p.ActionLogOutput(g))
}

// cuiLineContaining returns the first line of out that contains marker.
// 出力全体に対する Contains だと、別の行 (プレイヤー一覧など) に当たって
// 検査したい 1 行を素通りしてしまうので、行を取り出してから調べる。
func cuiLineContaining(out, marker string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, marker) {
			return line
		}
	}
	return ""
}

// basraPlayToGameEnd plays a whole game out and returns it sitting in GameEnd.
func basraPlayToGameEnd(t *testing.T) *domain.Basra {
	t.Helper()
	g := domain.NewDefaultBasra()
	g.Reset()
	for i := 0; i < 4000 && !g.GetGameEndFlag(); i++ {
		switch {
		case g.IsHumanTurn():
			valid := g.GetPlayableIndices(g.GetCurrentTurn())
			require.NotEmpty(t, valid)
			require.NoError(t, g.PlayerPlay(valid[0], nil))
		default:
			g.CpuPlay()
		}
	}
	require.True(t, g.GetGameEndFlag())
	return g
}

// #5694: Web の basra-result は勝者と最終得点を出すのに、CUI の GameEnd は
// 「ゲーム終了。nr で…」という案内だけで、誰が勝ったのかも何点だったのかも
// 画面のどこにも出ていなかった。
func TestBasraCuiPresenter_GameEndShowsWinnerAndScores(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true) // 名前と点の組を素の文字列で照合するため
	defer color.SetNoColor(orig)
	g := basraPlayToGameEnd(t)
	p := new(presenter.BasraCuiPresenter)

	out := p.Output(g, nil)

	playerLabel := func(idx int) string {
		if g.GetPlayer(idx).GetIsHuman() {
			return i18n.T("cuiPlayerYou")
		}
		return i18n.Tf("cuiPlayerCpu", "idx", strconv.Itoa(idx))
	}

	winners := g.GetWinners()
	require.NotEmpty(t, winners)
	winnerLine := cuiLineContaining(out, strings.Split(i18n.T("basra.resultWinner"), "{{")[0])
	require.NotEmpty(t, winnerLine, "the winner must be named")
	for _, w := range winners {
		assert.Contains(t, winnerLine, playerLabel(w))
	}

	scoreLine := cuiLineContaining(out, strings.Split(i18n.T("basra.resultScores"), "{{")[0])
	require.NotEmpty(t, scoreLine)
	// **名前と点を組で照合する。**行に 4 つ数字が並ぶので、単独の Contains だと
	// 別のプレイヤーの数字や捕獲枚数に当たって素通りする。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		want := fmt.Sprintf("%s %d", playerLabel(i), g.GetPlayer(i).GetScore())
		assert.Contains(t, scoreLine, want)
	}
}

// #5694: まだ終わっていない局面では出さない。
func TestBasraCuiPresenter_NoResultBeforeTheEnd(t *testing.T) {
	g := domain.NewDefaultBasra()
	g.Reset()
	g.SetPhase(domain.BasraPhasePlay)
	p := new(presenter.BasraCuiPresenter)

	out := p.Output(g, nil)

	assert.NotContains(t, out, strings.Split(i18n.T("basra.resultWinner"), "{{")[0])
	assert.NotContains(t, out, strings.Split(i18n.T("basra.resultScores"), "{{")[0])
}
