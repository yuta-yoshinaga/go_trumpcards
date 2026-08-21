//go:build test

package presenter

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupSalicLawCuiMockDefaults(g *interfaces.MockSalicLawGame) {
	g.On("GetPhase").Return(domain.SalicLawPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("GetStockCount").Return(96).Maybe()
	g.On("GetOpenPiles").Return(domain.SalicLawTableauCnt).Maybe()
	queens := make([]*domain.Card, domain.SalicLawQueenCnt)
	for i := range queens {
		queens[i] = domain.NewCard(i%4+1, domain.SalicLawQueenValue, true)
	}
	g.On("GetQueens").Return(queens).Maybe()

	// 配りどおり、どの列も底が K。K だけの列は「置ける枠」として別表示に
	// なるので、既定盤は K の上に 1 枚積んでおく。
	var tableau [domain.SalicLawTableauCnt][]*domain.Card
	for i := range domain.SalicLawTableauCnt {
		tableau[i] = []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, true),
			domain.NewCard(domain.CardDesignSpade, i+2, true),
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.SalicLawFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func TestSalicLawCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawCuiMockDefaults(g)

		result := new(SalicLawCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "Salic Law")
		assert.Contains(t, result, i18n.T("saliclaw.foundationHeader"))
		assert.Contains(t, result, "列0:")
		assert.Contains(t, result, "列7:", "all eight columns are rendered")
		assert.Contains(t, result, "96")
		assert.Contains(t, result, "手数: 0")
		// **退場した Q を見せる。**8 枚が消えている理由は盤からは読めない。
		assert.Contains(t, result, i18n.T("saliclaw.queensHeader"))
	})

	// まだ K が出ていない列は「未開放」と言う。空列を「置ける枠」と読ませると
	// 唯一の置き場所を取り違える。
	t.Run("a column with no king yet says it is not open", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		g.On("GetTableau").Return([domain.SalicLawTableauCnt][]*domain.Card{})

		out := new(SalicLawCuiPresenter).Output(g, nil)
		assert.Contains(t, out, i18n.T("saliclaw.emptyPile"))
		assert.NotContains(t, out, i18n.T("saliclaw.bareKingPile"), "未開放は置ける枠ではない")
	})

	// **K だけの列には印を付ける。**このゲームで唯一の置き場所なので、
	// 他の列と同じ見た目だと探せない。
	t.Run("a bare king is marked as the one place a card can go", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTableau")
		var tableau [domain.SalicLawTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, true)}
		tableau[1] = []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, domain.CardValueMax, true),
			domain.NewCard(domain.CardDesignClover, 4, true),
		}
		g.On("GetTableau").Return(tableau)

		out := new(SalicLawCuiPresenter).Output(g, nil)
		// 印はちょうど 1 つ。全部の列に付いたら見分けにならない。
		assert.Equal(t, 1, strings.Count(out, i18n.T("saliclaw.bareKingPile")))
	})

	t.Run("stalemate shows undo-to-escape guidance", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(3)

		assert.Contains(t, new(SalicLawCuiPresenter).Output(g, nil),
			i18n.Tf("saliclaw.undoToEscape", "count", "3"))
	})

	t.Run("stalemate with no escape hides the guidance", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawCuiMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)
		g.On("UndoToEscape").Return(0)

		result := new(SalicLawCuiPresenter).Output(g, nil)
		assert.Contains(t, result, "手詰まり")
		assert.NotContains(t, result, "脱出には")
	})

	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		setupSalicLawCuiMockDefaults(g)

		assert.Contains(t, new(SalicLawCuiPresenter).Output(g, errors.New("test error")), "test error")
	})

	for _, tc := range []struct {
		name string
		val  domain.SalicLawPhase
		want string
	}{
		{"game clear", domain.SalicLawPhaseGameClear, "ゲームクリア"},
		{"game over", domain.SalicLawPhaseGameOver, "ゲームオーバー"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSalicLawGame)
			setupSalicLawCuiMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			assert.Contains(t, new(SalicLawCuiPresenter).Output(g, nil), tc.want)
		})
	}
}

func TestSalicLawCuiPresenter_HintOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hint     *domain.SalicLawHint
		contains []string
	}{
		{"tableau to a foundation",
			&domain.SalicLawHint{FromZone: "tableau", FromIdx: 1, ToZone: "foundation", ToIdx: 2},
			[]string{"タブロー列1", "基礎札2"}},
		{"onto a bare king",
			&domain.SalicLawHint{FromZone: "tableau", FromIdx: 0, ToZone: "tableau", ToIdx: 5},
			[]string{"タブロー列0", "タブロー列5"}},
		{"deal another card",
			&domain.SalicLawHint{FromZone: "stock", FromIdx: -1, ToZone: "stock", ToIdx: -1},
			[]string{i18n.T("saliclaw.hintDeal")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSalicLawGame)
			g.On("GetHint").Return(tc.hint)

			result := new(SalicLawCuiPresenter).HintOutput(g)
			for _, want := range tc.contains {
				assert.Contains(t, result, want)
			}
		})
	}

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		g.On("GetHint").Return((*domain.SalicLawHint)(nil))

		assert.Contains(t, new(SalicLawCuiPresenter).HintOutput(g), "ヒントはありません")
	})
}

func TestSalicLawCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		g.On("GetPhase").Return(domain.SalicLawPhasePlaying)

		assert.Contains(t, new(SalicLawCuiPresenter).ActionLogOutput(g), "棋譜はありません")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockSalicLawGame)
		g.On("GetPhase").Return(domain.SalicLawPhaseGameOver)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(SalicLawCuiPresenter).ActionLogOutput(g), "move")
	})
}

// #5562: 空き山の規則違反は英語の生文で返っていたので、日本語ロケールでも
// 英語のまま出ていた。**コードを名乗るようにしただけでは足りない** — 訳が
// 無ければキー文字列がそのまま画面に出る。
func TestSalicLawCuiPresenter_ErrorsAreTranslated(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	defer i18n.SetLang("ja")

	codes := []string{
		"saliclaw.errStockEmptyNoRedeal",
		"saliclaw.errPileEmpty",
		"saliclaw.errNoFoundationForCard",
		"saliclaw.errSamePile",
		"saliclaw.errKingIsTheBase",
		"saliclaw.errCannotPlaceOnPile",
		"saliclaw.errNothingToAutoComplete",
		"saliclaw.errNothingToUndo",
		"saliclaw.errNotPlaying",
		"saliclaw.errInvalidPile",
	}

	for _, lang := range []string{"ja", "en"} {
		i18n.SetLang(lang)
		for _, code := range codes {
			g := new(interfaces.MockSalicLawGame)
			setupSalicLawCuiMockDefaults(g)
			err := domain.NewDomainErrorCode(domain.ErrInvalidPlay, code, map[string]string{"pile": "3"})
			out := new(SalicLawCuiPresenter).Output(g, err)

			// **キーが画面に出ないこと。**訳の抜けはこれでしか見えない。
			assert.NotContains(t, out, code, "%s untranslated in %s", code, lang)
			// 訳文そのものが出ていること (キーが出ないだけなら空でも通る)。
			assert.Contains(t, out, i18n.Tf(code, "pile", "3"))
		}
	}

	// **日本語と英語で違う文が出ること。**両方に同じ英文を入れても上は通る。
	i18n.SetLang("ja")
	ja := i18n.T("saliclaw.errCannotPlaceOnPile")
	i18n.SetLang("en")
	assert.NotEqual(t, ja, i18n.T("saliclaw.errCannotPlaceOnPile"))
}

// コードを持たないエラー (逆シリアライズの防御など) は今までどおり素の文言で
// 出ること。全部を翻訳経路に流すと、キーでない文字列を引いて空行になる。
func TestSalicLawCuiPresenter_UncodedErrorsStillPrintTheirPhrase(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)

	g := new(interfaces.MockSalicLawGame)
	setupSalicLawCuiMockDefaults(g)
	out := new(SalicLawCuiPresenter).Output(g, errors.New("saliclaw: snapshot array exceeds maximum allowed size"))
	assert.Contains(t, out, "snapshot array exceeds maximum allowed size")
}
