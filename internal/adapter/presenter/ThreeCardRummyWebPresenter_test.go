//go:build test && (!js || !wasm || casino)

package presenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// ディーラーの札は 3 枚とも DIAMOND で揃えてある。プレイヤー側は
// SPADE/HEART/CLOVER しか使わないので、JSON 文字列に "DIAMOND" が
// 一度でも現れたらディーラーの手が漏れたということ。
//
// 同じスートでも 6/9/12 は連番でも同ランクでもないので「役」(0 点) には
// ならない。素点は 6+9+10=25 点。
var (
	threeCardRummyDealerCards = []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 6, true),
		domain.NewCard(domain.CardDesignDiamond, 9, true),
		domain.NewCard(domain.CardDesignDiamond, 12, true),
	}
	// プレイヤーは異スート・非連番 (2/4/7) なので役にならず、素点 13 点。
	threeCardRummyPlayerCards = []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 2, true),
		domain.NewCard(domain.CardDesignHeart, 4, true),
		domain.NewCard(domain.CardDesignClover, 7, true),
	}
)

const (
	threeCardRummyPlayerScore = 13
	threeCardRummyDealerScore = 25
)

// newThreeCardRummyDealtGame は「配り終えた」局面のゲームを組み立てる。
// Reset()/Shuffle() の結果には一切依存しない（配り依存のテストはフレークする）。
func newThreeCardRummyDealtGame(phase int) *domain.ThreeCardRummy {
	g := domain.NewDefaultThreeCardRummy()
	g.SetPhase(phase)
	g.SetPlayerHand(threeCardRummyPlayerCards)
	g.SetDealerHand(threeCardRummyDealerCards)
	g.SetPlayerScore(threeCardRummyPlayerScore)
	g.SetDealerScore(threeCardRummyDealerScore)
	g.SetAnteBet(100)
	g.SetLowBonusBet(20)
	return g
}

func parseThreeCardRummyOutput(t *testing.T, jsonStr string) *controller.ThreeCardRummyWebOutput {
	t.Helper()
	var out controller.ThreeCardRummyWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	require.NoError(t, err)
	return &out
}

// assertNoDealerCardLeak は生の JSON にディーラーの実札が現れないことを見る。
// マスクされた配列を見るだけでは、実札を別のフィールドに載せてしまう漏れ方を
// 検出できない。
func assertNoDealerCardLeak(t *testing.T, raw string) {
	t.Helper()
	assert.NotContains(t, raw, `"DIAMOND"`, "dealer's suit leaked into the JSON")
	for _, c := range threeCardRummyDealerCards {
		token := fmt.Sprintf(`"value":%d`, c.GetValue())
		assert.NotContains(t, raw, token, "dealer card value %d leaked into the JSON", c.GetValue())
	}
}

func TestThreeCardRummyWebPresenter_Output_BetPhase_NoDeal(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := domain.NewDefaultThreeCardRummy()

	raw := p.Output(g, nil)
	result := parseThreeCardRummyOutput(t, raw)
	assert.Equal(t, domain.ThreeCardRummyPhaseBet, result.Phase)
	assert.Equal(t, domain.ThreeCardRummyDefaultChips, result.Chips)
	// nil の手は null ではなく空配列で出す（フロントが length を読む）。
	assert.NotNil(t, result.PlayerHand)
	assert.NotNil(t, result.DealerHand)
	assert.Empty(t, result.PlayerHand)
	assert.Empty(t, result.DealerHand)
	assert.Empty(t, result.Message)
	assert.Empty(t, result.MessageCode)
	assert.Contains(t, raw, `"playerHand":[]`)
	assert.Contains(t, raw, `"dealerHand":[]`)
}

func TestThreeCardRummyWebPresenter_Output_AllFieldsPresent(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseEnd)
	g.SetChips(1234)
	g.SetPlayBet(100)
	g.SetResult(domain.GameResultWin)
	g.SetGameEndFlag(true)
	g.SetDealerQualified(true)
	g.SetAntePayout(200)
	g.SetPlayPayout(300)
	g.SetAnteBonusPayout(400)
	g.SetLowBonusPayout(500)

	result := parseThreeCardRummyOutput(t, p.Output(g, nil))
	assert.Len(t, result.PlayerHand, domain.ThreeCardRummyHandSize)
	assert.Len(t, result.DealerHand, domain.ThreeCardRummyHandSize)
	assert.Equal(t, domain.ThreeCardRummyPhaseEnd, result.Phase)
	assert.Equal(t, 1234, result.Chips)
	assert.Equal(t, 100, result.AnteBet)
	assert.Equal(t, 20, result.LowBonusBet)
	assert.Equal(t, 100, result.PlayBet)
	assert.Equal(t, int(domain.GameResultWin), result.Result)
	assert.Equal(t, 200, result.AntePayout)
	assert.Equal(t, 300, result.PlayPayout)
	assert.Equal(t, 400, result.AnteBonusPayout)
	assert.Equal(t, 500, result.LowBonusPayout)
	assert.Equal(t, 1400, result.TotalPayout)
	assert.True(t, result.DealerQualified)
	assert.Equal(t, threeCardRummyPlayerScore, result.PlayerScore)
	assert.Equal(t, threeCardRummyDealerScore, result.DealerScore)
}

// 勝負が付くまでディーラーの 3 枚は 1 枚も見せない。見えていたら合計が数えられ、
// play/fold の判断が意味を失う（カリビアンスタッドと違いアップカードは無い）。
func TestThreeCardRummyWebPresenter_Output_DealerHandFullyMaskedBeforeEnd(t *testing.T) {
	phases := map[string]int{
		"bet":    domain.ThreeCardRummyPhaseBet,
		"action": domain.ThreeCardRummyPhaseAction,
	}
	for name, phase := range phases {
		t.Run(name, func(t *testing.T) {
			p := new(ThreeCardRummyWebPresenter)
			g := newThreeCardRummyDealtGame(phase)

			raw := p.Output(g, nil)
			result := parseThreeCardRummyOutput(t, raw)
			assert.Equal(t, phase, result.Phase)
			require.Len(t, result.DealerHand, domain.ThreeCardRummyHandSize)
			for i, c := range result.DealerHand {
				assert.Equal(t, "", c.Design, "dealer card %d must be blank", i)
				assert.Equal(t, 0, c.Value, "dealer card %d must have no value", i)
			}
			assertNoDealerCardLeak(t, raw)
		})
	}
}

func TestThreeCardRummyWebPresenter_Output_DealerHandRevealedAtEnd(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseEnd)

	raw := p.Output(g, nil)
	result := parseThreeCardRummyOutput(t, raw)
	require.Len(t, result.DealerHand, domain.ThreeCardRummyHandSize)
	for i, c := range result.DealerHand {
		assert.Equal(t, "DIAMOND", c.Design, "dealer card %d must be revealed", i)
		assert.Equal(t, threeCardRummyDealerCards[i].GetValue(), c.Value, "dealer card %d value", i)
	}
	assert.Contains(t, raw, `"DIAMOND"`)
}

// プレイヤーの手は自分の札なので、どのフェーズでも伏せない。
func TestThreeCardRummyWebPresenter_Output_PlayerHandNeverMasked(t *testing.T) {
	wantDesigns := []string{"SPADE", "HEART", "CLOVER"}
	for _, phase := range []int{
		domain.ThreeCardRummyPhaseBet,
		domain.ThreeCardRummyPhaseAction,
		domain.ThreeCardRummyPhaseEnd,
	} {
		t.Run(fmt.Sprintf("phase%d", phase), func(t *testing.T) {
			p := new(ThreeCardRummyWebPresenter)
			g := newThreeCardRummyDealtGame(phase)

			result := parseThreeCardRummyOutput(t, p.Output(g, nil))
			require.Len(t, result.PlayerHand, domain.ThreeCardRummyHandSize)
			for i, c := range result.PlayerHand {
				assert.Equal(t, wantDesigns[i], c.Design, "player card %d design", i)
				assert.Equal(t, threeCardRummyPlayerCards[i].GetValue(), c.Value, "player card %d value", i)
			}
			// 配ってからは点数が出ている（Web GUI のヒントが読む）。
			assert.Equal(t, threeCardRummyPlayerScore, result.PlayerScore)
		})
	}
}

func TestThreeCardRummyWebPresenter_Output_MessageCodes(t *testing.T) {
	tests := []struct {
		name            string
		result          domain.GameResult
		playBet         int
		dealerQualified bool
		wantMessage     string
		wantCode        string
	}{
		{
			name:            "player wins",
			result:          domain.GameResultWin,
			playBet:         100,
			dealerQualified: true,
			wantMessage:     "Player wins!",
			wantCode:        "threecardrummy.result.playerWins",
		},
		{
			name:            "dealer wins",
			result:          domain.GameResultLose,
			playBet:         100,
			dealerQualified: true,
			wantMessage:     "Dealer wins!",
			wantCode:        "threecardrummy.result.dealerWins",
		},
		{
			// playBet が 0 の負けはフォールド。「ディーラーの勝ち」ではない。
			name:            "fold",
			result:          domain.GameResultLose,
			playBet:         0,
			dealerQualified: true,
			wantMessage:     "Player folded.",
			wantCode:        "threecardrummy.result.fold",
		},
		{
			name:            "push",
			result:          domain.GameResultDraw,
			playBet:         100,
			dealerQualified: true,
			wantMessage:     "Push!",
			wantCode:        "threecardrummy.result.push",
		},
		{
			// クオリファイ不成立は勝敗の文言を上書きする。
			name:            "dealer not qualified overrides the win",
			result:          domain.GameResultWin,
			playBet:         100,
			dealerQualified: false,
			wantMessage:     "Dealer does not qualify!",
			wantCode:        "threecardrummy.result.dealerNotQualified",
		},
		{
			// フォールド済み (playBet == 0) なら勝負していないので、
			// クオリファイ不成立でも上書きしない。
			name:            "fold is not overridden by an unqualified dealer",
			result:          domain.GameResultLose,
			playBet:         0,
			dealerQualified: false,
			wantMessage:     "Player folded.",
			wantCode:        "threecardrummy.result.fold",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := new(ThreeCardRummyWebPresenter)
			g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseEnd)
			g.SetGameEndFlag(true)
			g.SetResult(tt.result)
			g.SetPlayBet(tt.playBet)
			g.SetDealerQualified(tt.dealerQualified)

			result := parseThreeCardRummyOutput(t, p.Output(g, nil))
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, tt.wantCode, result.MessageCode)
		})
	}
}

// 決着が付いていない局面では結果の文言を出さない。
func TestThreeCardRummyWebPresenter_Output_NoResultMessageBeforeGameEnd(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseAction)
	g.SetResult(domain.GameResultWin)
	g.SetPlayBet(100)
	g.SetGameEndFlag(false)

	result := parseThreeCardRummyOutput(t, p.Output(g, nil))
	assert.Empty(t, result.Message)
	assert.Empty(t, result.MessageCode)
}

func TestThreeCardRummyWebPresenter_Output_Error(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseEnd)
	// エラーは結果の文言より優先される。
	g.SetGameEndFlag(true)
	g.SetResult(domain.GameResultWin)
	g.SetPlayBet(100)
	g.SetDealerQualified(true)

	result := parseThreeCardRummyOutput(t, p.Output(g, errors.New("test error")))
	assert.Equal(t, "test error", result.Message)
	assert.Empty(t, result.MessageCode)
}

func TestThreeCardRummyWebPresenter_HintOutput_MirrorsOutput(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseEnd)
	g.SetGameEndFlag(true)
	g.SetResult(domain.GameResultWin)
	g.SetPlayBet(100)
	g.SetDealerQualified(true)
	g.SetAntePayout(200)

	assert.Equal(t, p.Output(g, nil), p.HintOutput(g))

	result := parseThreeCardRummyOutput(t, p.HintOutput(g))
	assert.Equal(t, domain.ThreeCardRummyPhaseEnd, result.Phase)
	assert.Equal(t, threeCardRummyPlayerScore, result.PlayerScore)
	assert.Equal(t, "threecardrummy.result.playerWins", result.MessageCode)
}

// ヒントでもディーラーの手は伏せたまま（勝負前にヒントを叩けば見えてしまう）。
func TestThreeCardRummyWebPresenter_HintOutput_KeepsDealerHandMasked(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseAction)

	raw := p.HintOutput(g)
	result := parseThreeCardRummyOutput(t, raw)
	require.Len(t, result.DealerHand, domain.ThreeCardRummyHandSize)
	for i, c := range result.DealerHand {
		assert.Equal(t, "", c.Design, "dealer card %d must be blank", i)
	}
	assertNoDealerCardLeak(t, raw)
}

func TestThreeCardRummyWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(ThreeCardRummyWebPresenter)
	g := newThreeCardRummyDealtGame(domain.ThreeCardRummyPhaseEnd)

	assert.Contains(t, p.ActionLogOutput(g), "entries")
}
