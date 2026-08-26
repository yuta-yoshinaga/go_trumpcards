//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MendikotWebPresenter メンディコットWebプレゼンタークラス
type MendikotWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MendikotWebPresenter) Output(m interfaces.MendikotGame, lastErr error) string {
	resObj := p.buildBase(m)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(m, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	if hint := m.GetHint(); hint != nil {
		resObj.Hint = &controller.MendikotWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *MendikotWebPresenter) buildBase(m interfaces.MendikotGame) *controller.MendikotWebOutput {
	resObj := new(controller.MendikotWebOutput)
	resObj.Phase = int(m.GetPhase())
	resObj.HandNumber = m.GetHandNumber()
	resObj.TrickNumber = m.GetTrickNumber()
	resObj.TrumpSuit = m.GetTrumpSuit()
	resObj.TrumpChooserIdx = m.GetTrumpChooserIdx()
	resObj.TensInDeck = domain.MendikotTensInDeck
	resObj.LastHandWinner = m.GetLastHandWinner()
	resObj.LastHandKind = m.GetLastHandKind()
	resObj.CurrentPlayerIdx = m.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = m.GetLeadPlayerIdx()
	resObj.DealerIdx = m.GetDealerIdx()
	resObj.ValidPlays = intSliceOrEmpty(m.GetValidPlayIndices(0))
	resObj.WillSetTrump = m.IsHumanTurn() && m.WillSetTrump(m.GetCurrentPlayerIdx())
	resObj.GameEndFlag = m.GetGameEndFlag()
	resObj.WinnerTeam = m.GetWinnerTeam()
	resObj.CurrentTrick = trickCardsToOutput(m.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(m)

	scores := make([]int, 0, domain.MendikotTeamCnt)
	tens := make([]int, 0, domain.MendikotTeamCnt)
	tricks := make([]int, 0, domain.MendikotTeamCnt)
	for team := 0; team < domain.MendikotTeamCnt; team++ {
		scores = append(scores, m.GetScore(team))
		// **勝敗を決めるのは 10 の枚数。** トリック数は 2-2 のときだけ効く。
		tens = append(tens, m.TeamTens(team))
		tricks = append(tricks, m.TeamTricks(team))
	}
	resObj.Scores = scores
	resObj.TeamTens = tens
	resObj.TeamTricks = tricks
	resObj.Config = controller.MendikotWebOutputConfig{Target: m.GetConfig().Target}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MendikotWebPresenter) buildPlayersOutput(m interfaces.MendikotGame) []*controller.MendikotWebOutputPlayer {
	out := make([]*controller.MendikotWebOutputPlayer, 0)
	for i := 0; i < m.GetPlayerCnt(); i++ {
		player := m.GetPlayer(i)
		out = append(out, &controller.MendikotWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			Team:       domain.MendikotTeamOf(i),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Tens:       player.GetTens(),
			TrickCount: player.GetTrickCount(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MendikotWebPresenter) buildMessage(m interfaces.MendikotGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if m.GetGameEndFlag() {
		params := map[string]string{
			"t0": strconv.Itoa(m.GetScore(0)),
			"t1": strconv.Itoa(m.GetScore(1)),
		}
		switch m.GetWinnerTeam() {
		case 0:
			return "", "mendikot.result.team0", params
		case 1:
			return "", "mendikot.result.team1", params
		default:
			return "", "mendikot.result.tie", params
		}
	}
	if m.GetPhase() == domain.MendikotPhaseHandEnd {
		// **どの決まり方だったかを言う。** 勝ち点が 1/2/3 と変わる理由が
		// 盤面から読めない。
		return "", "mendikot.handEnd." + m.GetLastHandKind(), map[string]string{
			"team":   strconv.Itoa(m.GetLastHandWinner()),
			"tens0":  strconv.Itoa(m.TeamTens(0)),
			"tens1":  strconv.Itoa(m.TeamTens(1)),
			"tricks": strconv.Itoa(m.TeamTricks(m.GetLastHandWinner())),
		}
	}
	// **切り札が決まる前と後で案内が違う。**
	if m.GetTrumpSuit() == 0 {
		return "", "mendikot.play.noTrump", nil
	}
	return "", "mendikot.play", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *MendikotWebPresenter) HintOutput(m interfaces.MendikotGame) string {
	resObj := p.buildBase(m)
	if hint := m.GetHint(); hint != nil {
		resObj.Hint = &controller.MendikotWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MendikotWebPresenter) ActionLogOutput(m interfaces.MendikotGame) string {
	return actionLogOutputJSON(m)
}
