//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// sutdaMonthGlyphs は月ごとの絵柄 (花札の 1〜10 月)。
var sutdaMonthGlyphs = []string{"", "🌸", "🌺", "🌷", "🌿", "🪻", "🦋", "🍀", "🌕", "🍶", "🍁"}

// sutdaFace は 20 枚のカルタを手続き描画するための自己記述子を返す。
//
// **1・3・8 月だけは 2 枚が別物。** 片方が「光」で、どちらを持っているかで
// 役が変わる ── ラベルに光の印を出さないと、盤面から役が読めない。
func sutdaFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	month := card.GetDesign()
	glyph := "🎴"
	if month >= 1 && month < len(sutdaMonthGlyphs) {
		glyph = sutdaMonthGlyphs[month]
	}
	label := strconv.Itoa(month)
	colour := "black"
	if domain.SutdaIsGwang(card) {
		label += "光"
		colour = "gold"
	}
	return &CardFace{Glyph: glyph, Label: label, Color: colour, Deck: "hanafuda"}
}

// SutdaWebPresenter はソッタの Web プレゼンター。
type SutdaWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *SutdaWebPresenter) Output(g interfaces.SutdaGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	if hint := g.GetHint(); hint != nil {
		resObj.HintAction = hint.Action
		resObj.HintReason = hint.Reason
	}
	return marshalOrError(resObj)
}

// buildBase は共通フィールドを構築する。
func (p *SutdaWebPresenter) buildBase(g interfaces.SutdaGame) *controller.SutdaWebOutput {
	human := p.humanIdx(g)
	resObj := new(controller.SutdaWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.HandNumber = g.GetHandNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	// **配り終えたポットは 0 になる。** ショーダウンでそのまま 0 を返すと、
	// 画面が「誰が何を取った」の隣に 0 を並べることになる。
	resObj.Pot = g.GetPot()
	if g.GetPhase() != domain.SutdaPhaseBet {
		if res := g.GetLastResult(); res != nil {
			resObj.Pot = res.Pot
		}
	}
	resObj.CurrentBet = g.GetCurrentBet()
	resObj.CallAmount = g.GetCallAmount(human)
	resObj.CanRaise = g.IsHumanTurn() && g.CanRaise(human)
	resObj.RaiseCount = g.GetRaiseCount()
	resObj.MaxRaises = domain.SutdaMaxRaises
	resObj.BetUnit = domain.SutdaBetUnit
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsShowdown = g.GetPhase() == domain.SutdaPhaseShowdown
	// **自分の役は常に見える。** 伏せているのは相手の手札であって、自分の
	// 2 枚が何の役かは最初から分かっている。
	resObj.HumanHandName = g.GetHandOf(human).Name
	resObj.LastResult = p.lastResult(g)

	cfg := g.GetConfig()
	resObj.Config = controller.SutdaWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		Seats:         cfg.Seats,
		StartChips:    cfg.StartChips,
	}
	resObj.Players = p.buildPlayersOutput(g, human)
	return resObj
}

// humanIdx は人間の席を返す (居なければ 0)。
func (p *SutdaWebPresenter) humanIdx(g interfaces.SutdaGame) int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			return i
		}
	}
	return 0
}

// lastResult は直前ハンドの結果を出力形へ直す。
func (p *SutdaWebPresenter) lastResult(g interfaces.SutdaGame) *controller.SutdaWebOutputResult {
	res := g.GetLastResult()
	if res == nil {
		return nil
	}
	names := make([]string, 0, len(res.Hands))
	for _, h := range res.Hands {
		names = append(names, h.Name)
	}
	folded := make([]bool, len(res.Folded))
	copy(folded, res.Folded)
	winners := make([]int, len(res.Winners))
	copy(winners, res.Winners)
	return &controller.SutdaWebOutputResult{
		Winners:   winners,
		Pot:       res.Pot,
		HandNames: names,
		Folded:    folded,
	}
}

// buildPlayersOutput は席の情報を構築する。
func (p *SutdaWebPresenter) buildPlayersOutput(g interfaces.SutdaGame, human int) []*controller.SutdaWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.SutdaWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		// **開くのはショーダウンだけ。** ベッティング中に相手の役が見えると、
		// 賭ける意味が無くなる。
		visible := i == human || player.IsRevealed()
		row := &controller.SutdaWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutputWithFace(player, visible, sutdaFace),
			Chips:     player.GetChips(),
			Bet:       player.GetBet(),
			Folded:    player.IsFolded(),
			Revealed:  player.IsRevealed(),
			IsDealer:  i == dealer,
		}
		if visible {
			hand := g.GetHandOf(i)
			row.HandName = hand.Name
			row.HandRank = hand.Rank
		}
		out = append(out, row)
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *SutdaWebPresenter) buildMessage(g interfaces.SutdaGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	human := p.humanIdx(g)
	if g.GetGameEndFlag() {
		if g.GetWinnerIdx() == human {
			return "", "sutda.result.humanWin", nil
		}
		return "", "sutda.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.SutdaPhaseBet:
		if g.GetCallAmount(human) > 0 {
			return "", "sutda.betPhase.toCall", nil
		}
		return "", "sutda.betPhase", nil
	case domain.SutdaPhaseShowdown:
		res := g.GetLastResult()
		if res != nil {
			for _, w := range res.Winners {
				if w == human {
					return "", "sutda.showdown.won", nil
				}
			}
		}
		return "", "sutda.showdown.lost", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *SutdaWebPresenter) HintOutput(g interfaces.SutdaGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && hint.Action != "" {
		resObj.HintAction = hint.Action
		resObj.HintReason = hint.Reason
		resObj.MessageCode = "sutda.hintRequested"
	} else {
		resObj.MessageCode = "sutda.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *SutdaWebPresenter) ActionLogOutput(g interfaces.SutdaGame) string {
	return actionLogOutputJSON(g)
}
