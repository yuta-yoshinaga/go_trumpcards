package presenter

import (
	"sort"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HeartsWebPresenter ハーツWebプレゼンタークラス
type HeartsWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *HeartsWebPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	resObj := p.buildBase(h)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(h, h.GetCurrentTrick(), lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *HeartsWebPresenter) buildBase(h interfaces.HeartsGame) *controller.HeartsWebOutput {
	resObj := new(controller.HeartsWebOutput)
	resObj.Phase = int(h.GetPhase())
	resObj.RoundNumber = h.GetRoundNumber()
	resObj.TrickNumber = h.GetTrickNumber()
	resObj.CurrentPlayerIdx = h.GetCurrentPlayerIdx()
	resObj.HeartsBroken = h.GetHeartsBroken()
	resObj.PassDirection = int(h.GetPassDirection())
	resObj.GameEndFlag = h.GetGameEndFlag()
	resObj.WinnerIdx = h.GetWinnerIdx()
	resObj.LeadPlayerIdx = h.GetLeadPlayerIdx()

	// 設定
	cfg := h.GetConfig()
	resObj.Config = controller.HeartsWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
		OmnibusJD:     cfg.OmnibusJD,
	}

	resObj.CurrentTrick = trickCardsToOutput(h.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(h)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *HeartsWebPresenter) buildPlayersOutput(h interfaces.HeartsGame) []*controller.HeartsWebOutputPlayer {
	out := make([]*controller.HeartsWebOutputPlayer, 0)
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		pObj := &controller.HeartsWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
			PenaltyCards:    heartsPenaltyCardsOutput(player),
		}
		out = append(out, pObj)
	}
	return out
}

// heartsQueenValue はスペードのクイーン (Q♠) のカード値。
const heartsQueenValue = 12

// heartsPenaltyCardsOutput はプレイヤーが獲得済みのトリックからペナルティ
// カード (全ハート + Q♠。オムニバスのJ♦はボーナスであり含まない) を抽出し、
// 表示用に整列 (ハート昇順→Q♠) して WebOutputCard スライスへ変換する。
// nil ではなく空スライスを返すため JSON は常に `[]` となる。
func heartsPenaltyCardsOutput(player *domain.HeartsPlayer) []*controller.WebOutputCard {
	penalties := make([]*domain.Card, 0)
	for _, trick := range player.GetTricksTaken() {
		for _, card := range trick {
			if isHeartsPenaltyCard(card) {
				penalties = append(penalties, card)
			}
		}
	}
	sort.SliceStable(penalties, func(i, j int) bool {
		di, dj := penalties[i].GetDesign(), penalties[j].GetDesign()
		if di != dj {
			return di == domain.CardDesignHeart // ハートを先頭に
		}
		return penalties[i].GetValue() < penalties[j].GetValue()
	})
	return cardsToOutputOrEmpty(penalties)
}

// isHeartsPenaltyCard はカードがハーツのペナルティカード (ハートまたはQ♠) か判定する。
func isHeartsPenaltyCard(card *domain.Card) bool {
	if card == nil {
		return false
	}
	if card.GetDesign() == domain.CardDesignHeart {
		return true
	}
	return card.GetDesign() == domain.CardDesignSpade && card.GetValue() == heartsQueenValue
}

// buildMessage ゲーム結果メッセージを構築
func (p *HeartsWebPresenter) buildMessage(h interfaces.HeartsGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if h.GetGameEndFlag() {
		winnerIdx := h.GetWinnerIdx()
		player := h.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("hearts", winnerIdx, isHuman)
	}
	switch h.GetPhase() {
	case domain.HeartsPhasePass:
		return "", "hearts.passPhase", nil
	case domain.HeartsPhasePlay:
		if len(trick) == 0 {
			return "", "hearts.playPhase.lead", nil
		}
		return "", "hearts.playPhase.follow", nil
	case domain.HeartsPhaseTrickEnd:
		return "", "hearts.trickEnd", nil
	case domain.HeartsPhaseRoundEnd:
		return "", "hearts.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *HeartsWebPresenter) HintOutput(h interfaces.HeartsGame) string {
	hint := h.GetHint()
	resObj := p.buildBase(h)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *HeartsWebPresenter) ActionLogOutput(h interfaces.HeartsGame) string {
	return actionLogOutputJSON(h)
}
