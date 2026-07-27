//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// IndianPokerWebPresenter インディアンポーカーWebプレゼンタークラス
type IndianPokerWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (iwp *IndianPokerWebPresenter) Output(ip interfaces.IndianPokerGame, lastErr error) string {
	resObj := iwp.buildOutput(ip, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をIndianPokerWebOutputに変換
func (iwp *IndianPokerWebPresenter) buildOutput(ip interfaces.IndianPokerGame, lastErr error) *controller.IndianPokerWebOutput {
	resObj := new(controller.IndianPokerWebOutput)
	resObj.Phase = ip.GetPhase()
	resObj.Pot = ip.GetPot()
	resObj.DealerIdx = ip.GetDealerIdx()
	resObj.CurrentTurn = ip.GetCurrentTurn()
	resObj.GameEndFlag = ip.GetGameEndFlag()
	resObj.LastBet = ip.GetLastBet()
	resObj.MinRaise = ip.GetMinRaise()
	resObj.RaiseCount = ip.GetRaiseCount()
	resObj.HandCount = ip.GetHandCount()

	cfg := ip.GetConfig()
	resObj.BettingLimit = int(cfg.BettingLimit)
	resObj.Ante = cfg.Ante
	_, resObj.MaxBetAmount = domain.CalculateBettingLimits(cfg.BettingLimit, ip.GetPot(), ip.GetLastBet())

	resObj.SidePots = iwp.buildSidePotsOutput(ip)
	resObj.Players = iwp.buildPlayersOutput(ip)
	resObj.CpuActions = iwp.buildCpuActionsOutput(ip)
	resObj.RoundResults = iwp.buildRoundResultsOutput(ip)

	resObj.Message, resObj.MessageCode, resObj.MessageParams = iwp.buildMessage(ip, lastErr)

	// メタAI情報
	if profile := ip.GetHumanProfile(); profile != nil {
		resObj.MetaAI = &controller.IndianPokerWebOutputMetaAI{
			Enabled:        true,
			GamesPlayed:    profile.GamesPlayed,
			BluffRate:      profile.BluffRate(1),
			FoldRate:       profile.FoldRate(),
			HesitationMean: profile.HesitationMean,
		}
		d := profile.Export()
		resObj.Profile = &d
	}

	return resObj
}

// buildSidePotsOutput サイドポット情報を構築
func (iwp *IndianPokerWebPresenter) buildSidePotsOutput(ip interfaces.IndianPokerGame) []*controller.IndianPokerWebOutputSidePot {
	out := make([]*controller.IndianPokerWebOutputSidePot, 0)
	for _, sp := range ip.GetSidePots() {
		out = append(out, &controller.IndianPokerWebOutputSidePot{
			Amount:          sp.Amount,
			EligiblePlayers: sp.EligiblePlayers,
		})
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築
func (iwp *IndianPokerWebPresenter) buildPlayersOutput(ip interfaces.IndianPokerGame) []*controller.IndianPokerWebOutputPlayer {
	out := make([]*controller.IndianPokerWebOutputPlayer, 0)
	isShowdown := ip.GetPhase() == domain.IndianPokerPhaseShowdown || ip.GetPhase() == domain.IndianPokerPhaseEnd
	for i := 0; i < ip.GetPlayerCnt(); i++ {
		player := ip.GetPlayer(i)
		pObj := &controller.IndianPokerWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Chips:         player.GetChips(),
			CurrentBet:    player.GetCurrentBet(),
			Folded:        player.GetFolded(),
			AllIn:         player.GetAllIn(),
			PlayStyleName: player.GetPlayStyleName(),
		}

		// カード表示ロジック:
		// - 人間プレイヤー: ベッティング中はnull (自分のカードが見えない)、ショーダウン時は表示
		// - CPUプレイヤー: 常に表示 (インディアンポーカーでは他人のカードが見える)
		if player.GetCardsSize() > 0 {
			if player.GetIsHuman() {
				if isShowdown {
					pObj.Card = cardToOutput(player.GetCard(0))
				}
				// else: Card remains nil (human can't see own card during betting)
			} else {
				pObj.Card = cardToOutput(player.GetCard(0))
			}
		}

		out = append(out, pObj)
	}
	return out
}

// buildCpuActionsOutput CPU行動記録を構築
func (iwp *IndianPokerWebPresenter) buildCpuActionsOutput(ip interfaces.IndianPokerGame) []*controller.IndianPokerWebOutputCpuAction {
	out := make([]*controller.IndianPokerWebOutputCpuAction, 0)
	for _, action := range ip.GetCpuActions() {
		out = append(out, &controller.IndianPokerWebOutputCpuAction{
			PlayerIdx: action.PlayerIdx,
			Action:    action.Action,
			Amount:    action.Amount,
		})
	}
	return out
}

// buildRoundResultsOutput ラウンド結果を構築
func (iwp *IndianPokerWebPresenter) buildRoundResultsOutput(ip interfaces.IndianPokerGame) []*controller.IndianPokerWebOutputResult {
	out := make([]*controller.IndianPokerWebOutputResult, 0)
	for _, r := range ip.GetRoundResults() {
		result := &controller.IndianPokerWebOutputResult{
			PlayerIdx: r.PlayerIdx,
			CardRank:  r.CardRank,
			WonAmount: r.WonAmount,
		}
		if r.Card != nil {
			result.Card = cardToOutput(r.Card)
		}
		out = append(out, result)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (iwp *IndianPokerWebPresenter) buildMessage(ip interfaces.IndianPokerGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if ip.GetGameEndFlag() {
		msg, code := iwp.buildResultMessage(ip)
		return msg, code, nil
	}
	return "", "", nil
}

// buildResultMessage builds the end-of-round message and its i18n code
func (iwp *IndianPokerWebPresenter) buildResultMessage(ip interfaces.IndianPokerGame) (string, string) {
	results := ip.GetRoundResults()
	if len(results) == 0 {
		return "", "indianpoker.result.gameOver"
	}

	for _, r := range results {
		if ip.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "", "indianpoker.result.win"
			}
		}
	}

	// Human not in results (folded)
	for i := 0; i < ip.GetPlayerCnt(); i++ {
		if ip.GetPlayer(i).GetIsHuman() && ip.GetPlayer(i).GetFolded() {
			return "", "indianpoker.result.folded"
		}
	}

	return "", "indianpoker.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (iwp *IndianPokerWebPresenter) ActionLogOutput(ip interfaces.IndianPokerGame) string {
	return actionLogOutputJSON(ip)
}
