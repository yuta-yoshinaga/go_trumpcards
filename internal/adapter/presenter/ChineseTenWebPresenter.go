//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ChineseTenWebPresenter 撿紅點Webプレゼンタークラス
type ChineseTenWebPresenter struct{}

// chineseTenCardOutput は 1 枚を、得点と赤札かどうかまで含めて出力する。
// 得点表はサーバーが持つ -- クライアントに 52 枚ぶんの表をもう一部持たせない。
func chineseTenCardOutput(c *domain.Card) *controller.ChineseTenWebOutputCard {
	if c == nil {
		return nil
	}
	pts := domain.ChineseTenCardPoints(c)
	return &controller.ChineseTenWebOutputCard{
		WebOutputCard: cardToOutput(c),
		Points:        pts,
		IsRed:         c.GetDesign() == domain.CardDesignHeart || c.GetDesign() == domain.CardDesignDiamond,
	}
}

func chineseTenCardsOutput(cards []*domain.Card) []*controller.ChineseTenWebOutputCard {
	out := make([]*controller.ChineseTenWebOutputCard, 0, len(cards))
	for _, c := range cards {
		if oc := chineseTenCardOutput(c); oc != nil {
			out = append(out, oc)
		}
	}
	return out
}

// Output ゲーム状態をJSON出力
func (p *ChineseTenWebPresenter) Output(c interfaces.ChineseTenGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ChineseTenWebPresenter) buildBase(c interfaces.ChineseTenGame) *controller.ChineseTenWebOutput {
	resObj := new(controller.ChineseTenWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.StockCount = c.GetStockCount()
	resObj.TieScore = domain.ChineseTenTieScore
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()
	resObj.Layout = chineseTenCardsOutput(c.GetLayout())
	resObj.PendingCard = chineseTenCardOutput(c.GetPendingCard())

	sel := c.GetSelectableIndices()
	resObj.SelectableIndices = make([]int, 0, len(sel))
	resObj.SelectableIndices = append(resObj.SelectableIndices, sel...)

	cfg := c.GetConfig()
	resObj.Config = controller.ChineseTenWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。他ゲームは HintOutput でしか設定して
	// おらず、フロントは通常の state レスポンスから読むため、どのページも
	// 呼んでいない = ヒントトグルが何も表示しない状態になっている。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = chineseTenHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。Workers はこの JSON をそのままブラウザへ返すので、ここで
// 落とさなかったものは相手の手札がそのまま見えることを意味する。**取り札は
// 両者とも公開**する -- 何が取られたかを見て残りを読むのがフィッシング系の
// 骨格で、隠すとゲームが成立しない。
func (p *ChineseTenWebPresenter) buildPlayersOutput(c interfaces.ChineseTenGame) []*controller.ChineseTenWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.ChineseTenWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.ChineseTenWebOutputCard, 0, player.GetCardsSize())
		if reveal {
			for j := range player.GetCardsSize() {
				if oc := chineseTenCardOutput(player.GetCard(j)); oc != nil {
					cards = append(cards, oc)
				}
			}
		}
		out = append(out, &controller.ChineseTenWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			Captured:  chineseTenCardsOutput(c.GetCaptured(i)),
			Score:     c.GetScore(i),
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ChineseTenWebPresenter) buildMessage(c interfaces.ChineseTenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	switch c.GetWinnerIdx() {
	case 0:
		return "you win", "chineseten.win", nil
	case -1:
		return "draw", "chineseten.draw", nil
	default:
		return "you lose", "chineseten.lose", nil
	}
}

// HintOutput ヒント情報を出力する
func (p *ChineseTenWebPresenter) HintOutput(c interfaces.ChineseTenGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = chineseTenHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *ChineseTenWebPresenter) ActionLogOutput(c interfaces.ChineseTenGame) string {
	return actionLogOutputJSON(c)
}

// chineseTenHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func chineseTenHint(c interfaces.ChineseTenGame) *controller.ChineseTenWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.ChineseTenWebOutputHint{Reason: "chineseten.hint.game_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.ChineseTenWebOutputHint{Reason: "chineseten.hint.not_your_turn"}
	}
	action := c.ChineseTenCpuDecide(0)
	if c.GetPhase() == domain.ChineseTenPhaseSelect {
		if action.LayoutIdx < 0 {
			return &controller.ChineseTenWebOutputHint{Reason: "chineseten.hint.none"}
		}
		idx := action.LayoutIdx
		return &controller.ChineseTenWebOutputHint{LayoutIndex: &idx, Reason: "chineseten.hint.select"}
	}
	if action.HandIdx < 0 {
		return &controller.ChineseTenWebOutputHint{Reason: "chineseten.hint.none"}
	}
	idx := action.HandIdx
	return &controller.ChineseTenWebOutputHint{CardIndex: &idx, Reason: "chineseten.hint.play"}
}
