package presenter

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// AllFoursWebPresenter All Fours Webプレゼンタークラス
type AllFoursWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *AllFoursWebPresenter) Output(s interfaces.AllFoursGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, s.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**AllFours.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := s.GetHint(); hint != nil {
		resObj.Hint = &controller.AllFoursWebOutputHint{
			CardIndex: hint.CardIndex,
			Beg:       hint.Beg,
			Run:       hint.Run,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *AllFoursWebPresenter) buildBase(s interfaces.AllFoursGame) *controller.AllFoursWebOutput {
	resObj := new(controller.AllFoursWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.DealerIdx = s.GetDealerIdx()
	resObj.NonDealerIdx = s.GetNonDealerIdx()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.TrumpSuit = s.GetTrumpSuit()
	resObj.TurnUp = cardToOutput(s.GetTurnUp())
	resObj.RunCount = s.GetRunCount()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()

	cfg := s.GetConfig()
	resObj.Config = controller.AllFoursWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(s)
	resObj.RoundBreakdown = p.buildRoundBreakdown(s)

	for i := 0; i < s.GetPlayerCnt(); i++ {
		if pl := s.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			resObj.ValidPlayIndices = s.GetValidPlayIndices(i)
			break
		}
	}
	if resObj.ValidPlayIndices == nil {
		resObj.ValidPlayIndices = make([]int, 0)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *AllFoursWebPresenter) buildPlayersOutput(s interfaces.AllFoursGame) []*controller.AllFoursWebOutputPlayer {
	out := make([]*controller.AllFoursWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.AllFoursWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		})
	}
	return out
}

// allFoursRankValue はカード値 (1=A) を比較用ランクへ変換する (A=14 > K=13 > … > 2=2)。
// domain.pegAwards と同一ロジックを adapter 層で再現したもの (WASM ワーカーには載せない)。
func allFoursRankValue(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// allFoursPipValue は Game ポイント計算用のピップ値を返す (A=4 K=3 Q=2 J=1 10=10 他=0)。
func allFoursPipValue(v int) int {
	switch v {
	case 1:
		return 4
	case 13:
		return 3
	case 12:
		return 2
	case 11:
		return 1
	case 10:
		return 10
	}
	return 0
}

// buildRoundBreakdown はラウンド確定時 (ROUND_END / GAME_END) の
// High / Low / Jack / Game 得点内訳を構築する。domain.pegAwards の判定を
// adapter 層で再現し、各プレイヤーが捕獲したトランプ札とピップ合計から
// 各項目の獲得者を求める。その他のフェーズでは nil を返す。
//
// High/Low = 捕獲済みトランプの最高/最低ランク札の捕獲者。Jack = J トランプの捕獲者
// (場に出なければ -1)。Game = ピップ合計が単独最大のプレイヤー (同点・全員 0 なら -1)。
func (p *AllFoursWebPresenter) buildRoundBreakdown(s interfaces.AllFoursGame) *controller.AllFoursWebOutputRoundBreakdown {
	phase := s.GetPhase()
	if phase != domain.AllFoursPhaseRoundEnd && phase != domain.AllFoursPhaseGameEnd {
		return nil
	}
	playerCnt := s.GetPlayerCnt()
	bd := &controller.AllFoursWebOutputRoundBreakdown{
		High: controller.AllFoursWebOutputBreakdownAward{WinnerIdx: -1},
		Low:  controller.AllFoursWebOutputBreakdownAward{WinnerIdx: -1},
		Jack: controller.AllFoursWebOutputBreakdownJack{WinnerIdx: -1},
		Game: controller.AllFoursWebOutputBreakdownGame{WinnerIdx: -1, Points: make([]int, playerCnt)},
	}
	trump := s.GetTrumpSuit()
	if trump == domain.AllFoursTrumpUnset {
		return bd
	}

	highRank, lowRank := -1, math.MaxInt32
	var highCard, lowCard *domain.Card
	for i := 0; i < playerCnt; i++ {
		pl := s.GetPlayer(i)
		if pl == nil {
			continue
		}
		for _, trick := range pl.GetTricksTaken() {
			for _, card := range trick {
				if card == nil {
					continue
				}
				bd.Game.Points[i] += allFoursPipValue(card.GetValue())
				if card.GetDesign() != trump {
					continue
				}
				rank := allFoursRankValue(card.GetValue())
				if rank > highRank {
					highRank, highCard = rank, card
					bd.High.WinnerIdx = i
				}
				if rank < lowRank {
					lowRank, lowCard = rank, card
					bd.Low.WinnerIdx = i
				}
				if card.GetValue() == 11 {
					bd.Jack.WinnerIdx = i
				}
			}
		}
	}
	bd.High.Card = cardToOutput(highCard)
	bd.Low.Card = cardToOutput(lowCard)

	gw, maxTotal, tied := -1, -1, false
	for i, total := range bd.Game.Points {
		switch {
		case total > maxTotal:
			maxTotal, gw, tied = total, i, false
		case total == maxTotal:
			tied = true
		}
	}
	if !tied && gw >= 0 && maxTotal > 0 {
		bd.Game.WinnerIdx = gw
	}
	return bd
}

// buildMessage ゲーム結果メッセージを構築
func (p *AllFoursWebPresenter) buildMessage(s interfaces.AllFoursGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		winnerIdx := s.GetWinnerIdx()
		player := s.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("allfours", winnerIdx, isHuman)
	}
	switch s.GetPhase() {
	case domain.AllFoursPhaseBeg:
		return "", "allfours.begPhase", nil
	case domain.AllFoursPhaseGift:
		return "", "allfours.giftPhase", nil
	case domain.AllFoursPhasePlay:
		if len(trick) == 0 {
			return "", "allfours.playPhase.lead", nil
		}
		return "", "allfours.playPhase.follow", nil
	case domain.AllFoursPhaseTrickEnd:
		return "", "allfours.trickEnd", nil
	case domain.AllFoursPhaseRoundEnd:
		return "", "allfours.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *AllFoursWebPresenter) HintOutput(s interfaces.AllFoursGame) string {
	hint := s.GetHint()
	resObj := p.buildBase(s)
	if hint != nil {
		resObj.Hint = &controller.AllFoursWebOutputHint{
			CardIndex: hint.CardIndex,
			Beg:       hint.Beg,
			Run:       hint.Run,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *AllFoursWebPresenter) ActionLogOutput(s interfaces.AllFoursGame) string {
	return actionLogOutputJSON(s)
}
