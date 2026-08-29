//go:build test

package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupYukonCuiMockDefaults(yg *interfaces.MockYukonGame) {
	yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
	yg.On("GetMoveCount").Return(0).Maybe()
	yg.On("IsStalemate").Return(false).Maybe()
	yg.On("UndoToEscape").Return(0).Maybe()

	var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
	for i := range domain.YukonTableauCnt {
		tableau[i] = []*domain.KlondikeTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+1, false), FaceUp: true},
		}
	}
	yg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.YukonFoundationCnt][]*domain.Card
	yg.On("GetFoundation").Return(foundation).Maybe()
}

func TestYukonCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonCuiMockDefaults(yg)
		p := new(YukonCuiPresenter)

		result := p.Output(yg, nil)
		assert.Contains(t, result, "Yukon")
		assert.Contains(t, result, "Foundation")
		assert.Contains(t, result, "列0:")
		// **Yukon 固有の一括移動ルールを常時出す (#4788)。**盤面は Klondike と
		// 見分けが付かないので、Klondike の感覚だと「揃った並びしか動かせない」
		// と思い込んだままになる。
		assert.Contains(t, result, "順序に関係なく")
	})

	t.Run("does not advertise the block-move rule once the game has ended", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhaseGameClear).Maybe()
		yg.On("GetMoveCount").Return(42).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonCuiPresenter)
		assert.NotContains(t, p.Output(yg, nil), "順序に関係なく")
	})

	t.Run("with error", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		setupYukonCuiMockDefaults(yg)
		p := new(YukonCuiPresenter)

		result := p.Output(yg, assert.AnError)
		assert.Contains(t, result, assert.AnError.Error())
	})

	t.Run("stalemate", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
		yg.On("GetMoveCount").Return(5).Maybe()
		yg.On("IsStalemate").Return(true).Maybe()
		yg.On("UndoToEscape").Return(0).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonCuiPresenter)
		result := p.Output(yg, nil)
		assert.Contains(t, result, "手詰まり")
	})

	t.Run("game clear", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhaseGameClear).Maybe()
		yg.On("GetMoveCount").Return(42).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonCuiPresenter)
		result := p.Output(yg, nil)
		assert.Contains(t, result, "ゲームクリア")
	})

	t.Run("game over", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhaseGameOver).Maybe()
		yg.On("GetMoveCount").Return(10).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonCuiPresenter)
		result := p.Output(yg, nil)
		assert.Contains(t, result, "ゲームオーバー")
	})

	t.Run("foundation with cards", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhasePlaying).Maybe()
		yg.On("GetMoveCount").Return(0).Maybe()
		yg.On("IsStalemate").Return(false).Maybe()
		var tableau [domain.YukonTableauCnt][]*domain.KlondikeTableauCard
		yg.On("GetTableau").Return(tableau).Maybe()
		var foundation [domain.YukonFoundationCnt][]*domain.Card
		foundation[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		yg.On("GetFoundation").Return(foundation).Maybe()

		p := new(YukonCuiPresenter)
		result := p.Output(yg, nil)
		assert.Contains(t, result, "SPADE 1")
	})
}

func TestYukonCuiPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetHint").Return(&domain.YukonHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(YukonCuiPresenter)
		result := p.HintOutput(yg)
		assert.Contains(t, result, "ヒント")
		assert.Contains(t, result, "ファンデーション")
		// Foundation moves carry the high-priority confidence label.
		assert.Contains(t, result, "優先度: 高")
	})

	t.Run("hint to tableau", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetHint").Return(&domain.YukonHint{
			FromCol:   0,
			CardIndex: 1,
			ToZone:    "tableau",
			ToCol:     3,
		})

		p := new(YukonCuiPresenter)
		result := p.HintOutput(yg)
		assert.Contains(t, result, "タブロー列3")
		// Tableau moves carry the medium-priority confidence label.
		assert.Contains(t, result, "優先度: 中")
	})

	t.Run("no hint", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetHint").Return((*domain.YukonHint)(nil))

		p := new(YukonCuiPresenter)
		result := p.HintOutput(yg)
		assert.Contains(t, result, "ヒントはありません")
	})
}

func TestYukonCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhasePlaying)

		p := new(YukonCuiPresenter)
		result := p.ActionLogOutput(yg)
		assert.NotEmpty(t, result)
	})

	t.Run("game over", func(t *testing.T) {
		yg := new(interfaces.MockYukonGame)
		yg.On("GetPhase").Return(domain.YukonPhaseGameOver)
		yg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		p := new(YukonCuiPresenter)
		result := p.ActionLogOutput(yg)
		assert.NotEmpty(t, result)
	})
}

// 受け入れ条件そのもの: ドメインが返した拒否理由が、**画面に出る時点で**
// ロケールの言語になっていること (#6327)。ドメイン側のテストは「コードを
// 持っているか」までしか見ていないので、描画経路はここで押さえる。
//
// **`i18n.T(key)` を期待値にしない。** 未翻訳ならキーがそのまま返るので、
// 翻訳が無くても通ってしまう。実際の文言を書き、**反対の言語が漏れていない
// ことも**見る。
func TestYukonCuiPresenter_RefusalIsTranslated(t *testing.T) {
	refusal := domain.NewDomainErrorCode(domain.ErrInvalidPlay, "yukon.errNotAllFaceUp", nil)

	render := func(t *testing.T, lang string) string {
		t.Helper()
		orig := i18n.Lang()
		i18n.SetLang(lang)
		t.Cleanup(func() { i18n.SetLang(orig) })

		yg := new(interfaces.MockYukonGame)
		setupYukonCuiMockDefaults(yg)
		return new(YukonCuiPresenter).Output(yg, refusal)
	}

	t.Run("japanese", func(t *testing.T) {
		out := render(t, "ja")
		assert.Contains(t, out, "裏向きの札が残っているため自動で完成させられません")
		assert.NotContains(t, out, "still face down", "英語が漏れている")
		assert.NotContains(t, out, "yukon.err", "キーが生のまま出ている")
	})

	t.Run("english", func(t *testing.T) {
		out := render(t, "en")
		assert.Contains(t, out, "Cards are still face down")
		assert.NotContains(t, out, "裏向き", "日本語が漏れている")
		assert.NotContains(t, out, "yukon.err", "キーが生のまま出ている")
	})
}
