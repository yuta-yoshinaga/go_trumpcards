//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// StealingBundlesWebPresenter スティーリングバンドルWebプレゼンタークラス
type StealingBundlesWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *StealingBundlesWebPresenter) Output(s interfaces.StealingBundlesGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = stealingBundlesHintOutput(s)
	return marshalOrError(resObj)
}

// stealingBundlesHintOutput はヒントを出力形に変換する。
func stealingBundlesHintOutput(s interfaces.StealingBundlesGame) *controller.StealingBundlesWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.StealingBundlesWebOutputHint{
		CardIndex: hint.CardIndex,
		VictimIdx: hint.VictimIdx,
		Reason:    hint.Reason,
	}
}

// buildBase 共通フィールドを構築
func (p *StealingBundlesWebPresenter) buildBase(s interfaces.StealingBundlesGame) *controller.StealingBundlesWebOutput {
	resObj := new(controller.StealingBundlesWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.TableCards = cardsToOutput(s.GetTableCards())
	resObj.CanCapture = s.CanCapture(0)
	resObj.DeckRemaining = s.GetDeckRemaining()
	resObj.LastCaptureIdx = s.GetLastCaptureIdx()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.TurnNumber = s.GetTurnNumber()
	resObj.PacksDealt = s.GetPacksDealt()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Players = p.buildPlayersOutput(s)
	resObj.TableMatches, resObj.StealTargets = p.buildMoveMaps(s)
	resObj.Config = controller.StealingBundlesWebOutputConfig{PlayerCnt: s.GetConfig().PlayerCnt}
	return resObj
}

// buildMoveMaps は人間の手札ごとに「取れる場札」と「奪える相手」を並べる。
//
// **どの札で何ができるかは盤面から読み切れません。** 場札のランクと全員の束の
// 一番上を突き合わせる作業を、クライアントに繰り返させないための地図です。
func (p *StealingBundlesWebPresenter) buildMoveMaps(s interfaces.StealingBundlesGame) (map[string][]int, map[string][]int) {
	tableMatches := map[string][]int{}
	stealTargets := map[string][]int{}
	human := s.GetPlayer(0)
	if human == nil {
		return tableMatches, stealTargets
	}
	for i := 0; i < human.GetCardsSize(); i++ {
		key := strconv.Itoa(i)
		if m := s.GetTableMatches(0, i); len(m) > 0 {
			tableMatches[key] = m
		}
		if t := s.GetStealTargets(0, i); len(t) > 0 {
			stealTargets[key] = t
		}
	}
	return tableMatches, stealTargets
}

// buildPlayersOutput プレイヤー情報を構築
func (p *StealingBundlesWebPresenter) buildPlayersOutput(s interfaces.StealingBundlesGame) []*controller.StealingBundlesWebOutputPlayer {
	out := make([]*controller.StealingBundlesWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		entry := &controller.StealingBundlesWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			BundleSize: player.GetBundleSize(),
		}
		// **一番上だけは全員に見えます。** そこが狙われる場所だからです。
		if top := player.GetBundleTop(); top != nil {
			entry.BundleTop = cardToOutput(top)
		}
		out = append(out, entry)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *StealingBundlesWebPresenter) buildMessage(s interfaces.StealingBundlesGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		params := map[string]string{"n": strconv.Itoa(s.GetPlayer(s.GetWinnerIdx()).GetBundleSize())}
		if s.GetWinnerIdx() == 0 {
			return "", "stealingbundles.result.you", params
		}
		params["idx"] = strconv.Itoa(s.GetWinnerIdx())
		return "", "stealingbundles.result.cpu", params
	}
	if !s.IsHumanTurn() {
		return "", "stealingbundles.waiting", nil
	}
	// **取れるときは取らねばなりません。** 置けない理由を言わないと詰まって見えます。
	if s.CanCapture(0) {
		return "", "stealingbundles.mustCapture", nil
	}
	return "", "stealingbundles.trail", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *StealingBundlesWebPresenter) HintOutput(s interfaces.StealingBundlesGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = stealingBundlesHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *StealingBundlesWebPresenter) ActionLogOutput(s interfaces.StealingBundlesGame) string {
	return actionLogOutputJSON(s)
}
