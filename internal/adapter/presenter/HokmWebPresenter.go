//go:build !js || !wasm || classic

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HokmWebPresenter ホクムWebプレゼンタークラス
type HokmWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *HokmWebPresenter) Output(h interfaces.HokmGame, lastErr error) string {
	resObj := p.buildBase(h)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(h, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := h.GetHint(); hint != nil {
		resObj.Hint = &controller.HokmWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *HokmWebPresenter) buildBase(h interfaces.HokmGame) *controller.HokmWebOutput {
	resObj := new(controller.HokmWebOutput)
	resObj.Phase = int(h.GetPhase())
	resObj.HandNumber = h.GetHandNumber()
	resObj.TrickNumber = h.GetTrickNumber()
	resObj.TrumpSuit = h.GetTrumpSuit()
	resObj.HakemIdx = h.GetHakemIdx()
	resObj.TricksToWin = domain.HokmTricksToWin
	resObj.LastHandKot = h.GetLastHandKot()
	resObj.LastHandWinner = h.GetLastHandWinner()
	resObj.CurrentPlayerIdx = h.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = h.GetLeadPlayerIdx()
	resObj.ValidPlays = intSliceOrEmpty(h.GetValidPlayIndices(0))
	resObj.GameEndFlag = h.GetGameEndFlag()
	resObj.WinnerTeam = h.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(h.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(h)

	scores := make([]int, 0, domain.HokmTeamCnt)
	tricks := make([]int, 0, domain.HokmTeamCnt)
	for team := 0; team < domain.HokmTeamCnt; team++ {
		scores = append(scores, h.GetScore(team))
		// **7 で即終了なので、ハンドの進捗はトリック数のほうに出る。**
		tricks = append(tricks, h.TeamTricks(team))
	}
	resObj.Scores = scores
	resObj.TeamTricks = tricks
	resObj.Config = controller.HokmWebOutputConfig{Target: h.GetConfig().Target}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *HokmWebPresenter) buildPlayersOutput(h interfaces.HokmGame) []*controller.HokmWebOutputPlayer {
	out := make([]*controller.HokmWebOutputPlayer, 0)
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		out = append(out, &controller.HokmWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			Team:       domain.HokmTeamOf(i),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			IsHakem:    i == h.GetHakemIdx(),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *HokmWebPresenter) buildMessage(h interfaces.HokmGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if h.GetGameEndFlag() {
		params := map[string]string{
			"t0": strconv.Itoa(h.GetScore(0)),
			"t1": strconv.Itoa(h.GetScore(1)),
		}
		switch h.GetWinnerTeam() {
		case 0:
			return "", "hokm.result.team0", params
		case 1:
			return "", "hokm.result.team1", params
		default:
			return "", "hokm.result.tie", params
		}
	}
	switch h.GetPhase() {
	case domain.HokmPhaseTrump:
		if h.IsHumanTrumpTurn() {
			return "", "hokm.trump.choose", nil
		}
		return "", "hokm.trump.wait", nil
	case domain.HokmPhaseHandEnd:
		// **Kot かどうかで案内を変える。** 2 点入った理由が盤面から読めない。
		code := "hokm.handEnd"
		if h.GetLastHandKot() {
			code = "hokm.handEndKot"
		}
		return "", code, map[string]string{
			"hand": strconv.Itoa(h.GetHandNumber()),
			"t0":   strconv.Itoa(h.GetScore(0)),
			"t1":   strconv.Itoa(h.GetScore(1)),
		}
	}
	return "", "hokm.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *HokmWebPresenter) HintOutput(h interfaces.HokmGame) string {
	resObj := p.buildBase(h)
	if hint := h.GetHint(); hint != nil {
		resObj.Hint = &controller.HokmWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason, Suit: hint.Suit}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *HokmWebPresenter) ActionLogOutput(h interfaces.HokmGame) string {
	return actionLogOutputJSON(h)
}
