//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ZwickerWebPresenter ツヴィッカーWebプレゼンタークラス
type ZwickerWebPresenter struct{}

// zwickerCardOutput は 1 枚を、取りうるマッチ値つきで出力に変換する。
//
// **値表をクライアントに持たせない。**A=1/11, J=2/12, Q=3/13, K=4/14 と
// ジョーカー 15/20/25 はサーバーの捕獲判定と同じ表から出さないと必ずずれる。
func zwickerCardOutput(c *domain.Card) *controller.ZwickerWebOutputCard {
	if c == nil {
		return nil
	}
	return &controller.ZwickerWebOutputCard{
		WebOutputCard: cardToOutput(c),
		Values:        domain.ZwickerCardValues(c),
	}
}

func zwickerCardsOutput(cards []*domain.Card) []*controller.ZwickerWebOutputCard {
	out := make([]*controller.ZwickerWebOutputCard, 0, len(cards))
	for _, c := range cards {
		if c == nil {
			continue
		}
		out = append(out, zwickerCardOutput(c))
	}
	return out
}

func zwickerPlainCards(cards []*domain.Card) []*controller.WebOutputCard {
	out := make([]*controller.WebOutputCard, 0, len(cards))
	for _, c := range cards {
		if c == nil {
			continue
		}
		out = append(out, cardToOutput(c))
	}
	return out
}

// Output ゲーム状態をJSON出力
func (p *ZwickerWebPresenter) Output(c interfaces.ZwickerGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *ZwickerWebPresenter) buildBase(c interfaces.ZwickerGame) *controller.ZwickerWebOutput {
	resObj := new(controller.ZwickerWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.StockCount = c.GetStockCount()
	resObj.TableCards = zwickerCardsOutput(c.GetTableCards())
	resObj.TeamScores = [2]int{c.GetTeamScore(0), c.GetTeamScore(1)}
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerTeam = c.GetWinnerTeam()

	cfg := c.GetConfig()
	resObj.TargetScore = cfg.TargetScore
	resObj.Config = controller.ZwickerWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	builds := c.GetBuilds()
	resObj.Builds = make([]*controller.ZwickerWebOutputBuild, 0, len(builds))
	for _, b := range builds {
		if b == nil {
			continue
		}
		resObj.Builds = append(resObj.Builds, &controller.ZwickerWebOutputBuild{
			Owner: b.Owner,
			Value: b.Value,
			Cards: zwickerPlainCards(b.Cards),
		})
	}

	if s := c.GetLastRoundScore(); s != nil {
		resObj.LastRound = &controller.ZwickerWebOutputRoundScore{
			CardPoints:   s.CardPoints,
			Cards:        s.Cards,
			MajorityTeam: s.MajorityTeam,
			Zwicks:       s.Zwicks,
			Total:        s.Total,
		}
	}

	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = zwickerHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せるが、**取った枚数と Zwick 回数は公開**する。枚数最多の 3 点
// と Zwick 1 点はどちらもそこから直に読めるので、隠しても意味がない。
func (p *ZwickerWebPresenter) buildPlayersOutput(c interfaces.ZwickerGame) []*controller.ZwickerWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.ZwickerWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.ZwickerWebOutputCard, 0, player.GetCardsSize())
		if reveal {
			for j := range player.GetCardsSize() {
				if card := player.GetCard(j); card != nil {
					cards = append(cards, zwickerCardOutput(card))
				}
			}
		}
		out = append(out, &controller.ZwickerWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			Team:          domain.ZwickerTeamOf(i),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			CapturedCount: len(player.GetCaptured()),
			Zwicks:        player.GetZwicks(),
			Hidden:        !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *ZwickerWebPresenter) buildMessage(c interfaces.ZwickerGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerTeam() == domain.ZwickerTeamOf(0) {
		return "your team wins", "zwicker.win", nil
	}
	return "the other team wins", "zwicker.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *ZwickerWebPresenter) HintOutput(c interfaces.ZwickerGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = zwickerHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *ZwickerWebPresenter) ActionLogOutput(c interfaces.ZwickerGame) string {
	return actionLogOutputJSON(c)
}

// zwickerHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func zwickerHint(c interfaces.ZwickerGame) *controller.ZwickerWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.ZwickerWebOutputHint{Reason: "zwicker.hint.game_end"}
	}
	if c.GetPhase() != domain.ZwickerPhasePlay {
		return &controller.ZwickerWebOutputHint{Reason: "zwicker.hint.round_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.ZwickerWebOutputHint{Reason: "zwicker.hint.not_your_turn"}
	}
	action := c.ZwickerCpuDecide(0)
	if action.HandIdx < 0 {
		return &controller.ZwickerWebOutputHint{Reason: "zwicker.hint.none"}
	}
	idx := action.HandIdx
	if action.Type == "take" {
		return &controller.ZwickerWebOutputHint{
			Take: true, CardIndex: &idx, Value: action.Value,
			TableIdxs: action.TableIdxs, Reason: "zwicker.hint.take",
		}
	}
	return &controller.ZwickerWebOutputHint{CardIndex: &idx, Reason: "zwicker.hint.trail"}
}
