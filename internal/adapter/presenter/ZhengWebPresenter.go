//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// ZhengWebPresenter 争上游 Webプレゼンタークラス
type ZhengWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *ZhengWebPresenter) Output(zg interfaces.ZhengGame, lastErr error) string {
	resObj := new(controller.ZhengWebOutput)
	resObj.Players = make([]*controller.ZhengWebOutputPlayer, 0)
	resObj.CurrentTurn = zg.GetCurrentTurn()
	resObj.LastPlayPlayerIdx = zg.GetLastPlayPlayerIdx()
	resObj.GameEndFlag = zg.GetGameEndFlag()
	resObj.TablePlayType = int(zg.GetTablePlayType())

	config := zg.GetConfig()
	resObj.Config = controller.ZhengWebConfig{
		CpuDifficulty: int(config.CpuDifficulty),
	}

	resObj.TableCards = cardsToOutputOrEmpty(zg.GetTableCards())

	resObj.CpuActions = make([]*controller.ZhengWebOutputAction, 0)
	for _, action := range zg.GetCpuActions() {
		a := &controller.ZhengWebOutputAction{
			PlayerIdx:   action.PlayerIdx,
			PlayedCards: cardsToOutput(action.PlayedCards),
		}
		resObj.CpuActions = append(resObj.CpuActions, a)
	}

	humanAction := zg.GetHumanAction()
	if humanAction != nil {
		resObj.HumanAction = &controller.ZhengWebOutputAction{
			PlayerIdx:   humanAction.PlayerIdx,
			PlayedCards: cardsToOutput(humanAction.PlayedCards),
		}
	}

	for i := 0; i < zg.GetPlayerCnt(); i++ {
		player := zg.GetPlayer(i)
		if player == nil {
			continue
		}
		pObj := new(controller.ZhengWebOutputPlayer)
		pObj.ID = i
		pObj.IsHuman = player.GetIsHuman()
		pObj.IsFinished = player.GetIsFinished()
		pObj.Rank = player.GetRank()
		pObj.CardCount = player.GetCardsSize()
		pObj.Cards = playerCardsToOutput(player, player.GetIsHuman())
		resObj.Players = append(resObj.Players, pObj)
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if zg.GetGameEndFlag() {
		rankings := p.buildRankings(zg)
		// MessageParams carries only the rankings; the frontend wraps it with the
		// localised "zheng.result.rankings" template (e.g. "Game Over! {{rankings}}"),
		// so the prefix must NOT be embedded here (it would double up). Message is the
		// plain-text fallback used when the client does not resolve messageCode.
		resObj.Message = i18n.T("zheng.gameEnd") + " " + rankings
		resObj.MessageCode = "zheng.result.rankings"
		resObj.MessageParams = map[string]string{"rankings": rankings}
	}

	return marshalOrError(resObj)
}

// buildRankings ゲーム終了時の順位文字列（接頭辞なし）を生成する
func (p *ZhengWebPresenter) buildRankings(zg interfaces.ZhengGame) string {
	var b strings.Builder
	for i := 0; i < zg.GetPlayerCnt(); i++ {
		player := zg.GetPlayer(i)
		if player == nil {
			continue
		}
		rank := player.GetRank()
		if rank < 1 || rank > 4 {
			continue
		}
		var name string
		if player.GetIsHuman() {
			name = i18n.T("zheng.playerYou")
		} else {
			name = fmt.Sprintf("CPU %d", i)
		}
		fmt.Fprintf(&b, "%s:%s ", name, i18n.Tf("zheng.rankN", "rank", strconv.Itoa(rank)))
	}
	return b.String()
}

// ActionLogOutput 棋譜をJSON出力
func (p *ZhengWebPresenter) ActionLogOutput(zg interfaces.ZhengGame) string {
	return actionLogOutputJSON(zg)
}
