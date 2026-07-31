//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrexWebPresenter トリックスWebプレゼンタークラス
type TrexWebPresenter struct{}

// trexSuits はドミノの列を送る順。スート定数は 1 始まりなので明示して持つ。
var trexSuits = []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}

// Output ゲーム状態をJSON出力
func (p *TrexWebPresenter) Output(c interfaces.TrexGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TrexWebPresenter) buildBase(c interfaces.TrexGame) *controller.TrexWebOutput {
	resObj := new(controller.TrexWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.KingIdx = c.GetKingIdx()
	resObj.Contract = int(c.GetContract())
	resObj.IsTrix = c.IsTrix()
	resObj.DealNo = c.GetDealNumber()
	resObj.TotalDeals = domain.TrexTotalDeals
	resObj.TrickNo = c.GetTrickNumber()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()

	avail := c.AvailableContracts()
	resObj.AvailableContracts = make([]int, 0, len(avail))
	for _, ct := range avail {
		resObj.AvailableContracts = append(resObj.AvailableContracts, int(ct))
	}

	trick := c.GetTrick()
	resObj.Trick = make([]*controller.TrexWebOutputTrickCard, 0, len(trick))
	for _, tc := range trick {
		if tc.Card == nil {
			continue
		}
		resObj.Trick = append(resObj.Trick, &controller.TrexWebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}

	// ドミノの 4 列。J=11 起点で上下に伸びるので、範囲をそのまま送れば
	// クライアントは「次に置ける端」を数え直さずに描ける。
	resObj.Runs = make([]*controller.TrexWebOutputRun, 0, len(trexSuits))
	for _, suit := range trexSuits {
		started, low, high := c.GetSuitRun(suit)
		resObj.Runs = append(resObj.Runs, &controller.TrexWebOutputRun{
			Suit: suit, Started: started, Low: low, High: high,
		})
	}

	order := c.GetFinishOrder()
	resObj.FinishOrder = make([]int, 0, len(order))
	resObj.FinishOrder = append(resObj.FinishOrder, order...)

	valid := c.GetValidPlayIndices(0)
	resObj.ValidIndices = make([]int, 0, len(valid))
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 && c.GetPhase() == domain.TrexPhasePlay {
		resObj.ValidIndices = append(resObj.ValidIndices, valid...)
	}
	// パスできるのは「ドミノで、出せる札が 1 枚も無い」ときだけ。トリック契約
	// にパスは存在しないので、そこで押せると規則が壊れて見える。
	resObj.CanPass = !c.GetGameEndFlag() &&
		c.GetCurrentPlayerIdx() == 0 &&
		c.GetPhase() == domain.TrexPhasePlay &&
		c.IsTrix() &&
		len(valid) == 0

	cfg := c.GetConfig()
	resObj.Config = controller.TrexWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() {
		resObj.Hint = trexHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。**得点は全席公開**する -- 個人戦で 20 ディールを戦う
// ゲームなので、誰がどれだけ沈んでいるかが契約選択の判断材料そのもの。
func (p *TrexWebPresenter) buildPlayersOutput(c interfaces.TrexGame) []*controller.TrexWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.TrexWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		if reveal {
			for j := range player.GetCardsSize() {
				if card := player.GetCard(j); card != nil {
					cards = append(cards, cardToOutput(card))
				}
			}
		}
		out = append(out, &controller.TrexWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			Score:     c.GetScore(i),
			DealScore: c.GetDealScore(i),
			TricksWon: c.GetTricksWon(i),
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TrexWebPresenter) buildMessage(c interfaces.TrexGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerIdx() == 0 {
		return "you win", "trex.win", nil
	}
	return "you lose", "trex.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *TrexWebPresenter) HintOutput(c interfaces.TrexGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = trexHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *TrexWebPresenter) ActionLogOutput(c interfaces.TrexGame) string {
	return actionLogOutputJSON(c)
}

// trexHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func trexHint(c interfaces.TrexGame) *controller.TrexWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.TrexWebOutputHint{Reason: "trex.hint.game_end"}
	}
	if c.GetPhase() == domain.TrexPhaseChoose {
		// 選択フェーズの手番は王。人間が王でなければ助言することがない。
		if c.GetKingIdx() != 0 {
			return &controller.TrexWebOutputHint{Reason: "trex.hint.not_your_turn"}
		}
		action := c.TrexCpuDecide(0)
		ct := int(action.Contract)
		return &controller.TrexWebOutputHint{Contract: &ct, Reason: "trex.hint.choose"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.TrexWebOutputHint{Reason: "trex.hint.not_your_turn"}
	}
	action := c.TrexCpuDecide(0)
	if action.Pass {
		return &controller.TrexWebOutputHint{Pass: true, Reason: "trex.hint.pass"}
	}
	if action.HandIdx < 0 {
		return &controller.TrexWebOutputHint{Reason: "trex.hint.none"}
	}
	idx := action.HandIdx
	return &controller.TrexWebOutputHint{CardIndex: &idx, Reason: "trex.hint.play"}
}
