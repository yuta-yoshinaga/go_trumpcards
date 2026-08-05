//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PochWebPresenter ポッホWebプレゼンタークラス
type PochWebPresenter struct{}

func pochCardsOutput(cards []*domain.Card) []*controller.WebOutputCard {
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
func (p *PochWebPresenter) Output(c interfaces.PochGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *PochWebPresenter) buildBase(c interfaces.PochGame) *controller.PochWebOutput {
	resObj := new(controller.PochWebOutput)
	resObj.Phase = int(c.GetPhase())
	// 人間の手番でないときは null ではなく空スライス。
	resObj.ValidPlays = c.PochValidPlays(0)
	if resObj.ValidPlays == nil {
		resObj.ValidPlays = make([]int, 0)
	}
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.PaySuit = c.GetPaySuit()
	resObj.BetTarget = c.GetBetTarget()
	resObj.PochenWinner = c.GetPochenWinner()
	resObj.PochenPot = c.GetPochenPot()
	resObj.StopsSuit = c.GetStopsSuit()
	resObj.StopsRank = c.GetStopsRank()
	resObj.DealNo = c.GetDealNumber()
	resObj.DealWinner = c.GetDealWinner()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerIdx = c.GetWinnerIdx()
	resObj.PlayedPile = pochCardsOutput(c.GetPlayedPile())

	if t := c.GetTurnUp(); t != nil {
		resObj.TurnUp = cardToOutput(t)
	}

	// **9 区画は全部送る。**取られなかった区画のチップは持ち越されるので、
	// 「今いくら乗っているか」がそのまま次のディールの狙いどころになる。
	board := c.GetBoard()
	resObj.Pools = make([]*controller.PochWebOutputPool, 0, domain.PochPoolCount)
	for i := range domain.PochPoolCount {
		pool := domain.PochPool(i)
		resObj.Pools = append(resObj.Pools, &controller.PochWebOutputPool{
			Name: pool.String(), Chips: board.Get(pool),
		})
	}

	// 第 1 段階は配った直後に自動で解決するので、**結果を送らないと何が
	// 起きたのか画面から読めない。**
	awards := c.GetStakingAwards()
	resObj.StakingAwards = make([]*controller.PochWebOutputAward, 0, len(awards))
	for _, a := range awards {
		if a == nil {
			continue
		}
		resObj.StakingAwards = append(resObj.StakingAwards, &controller.PochWebOutputAward{
			Pool: a.Pool.String(), Player: a.Player, Chips: a.Chips,
		})
	}

	cfg := c.GetConfig()
	resObj.TargetDeals = cfg.TargetDeals
	resObj.Config = controller.PochWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = pochHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せるが、**枚数とチップは公開**する。出し切った人は他家から
// 残り札 1 枚につき 1 チップ受け取るので、枚数はそのまま負債額である。
func (p *PochWebPresenter) buildPlayersOutput(c interfaces.PochGame) []*controller.PochWebOutputPlayer {
	players := c.GetPlayers()
	out := make([]*controller.PochWebOutputPlayer, 0, len(players))
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
		out = append(out, &controller.PochWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			Chips:     player.GetChips(),
			Bet:       player.GetBet(),
			Folded:    player.IsFolded(),
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PochWebPresenter) buildMessage(c interfaces.PochGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerIdx() == 0 {
		return "you finish with the most chips", "poch.win", nil
	}
	return "you finish behind", "poch.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *PochWebPresenter) HintOutput(c interfaces.PochGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = pochHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *PochWebPresenter) ActionLogOutput(c interfaces.PochGame) string {
	return actionLogOutputJSON(c)
}

// pochHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func pochHint(c interfaces.PochGame) *controller.PochWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.PochWebOutputHint{Reason: "poch.hint.game_end"}
	}
	phase := c.GetPhase()
	if phase != domain.PochPhasePochen && phase != domain.PochPhaseStops {
		return &controller.PochWebOutputHint{Reason: "poch.hint.deal_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.PochWebOutputHint{Reason: "poch.hint.not_your_turn"}
	}
	action := c.PochCpuDecide(0)
	switch action.Type {
	case "bet":
		return &controller.PochWebOutputHint{Action: "bet", Reason: "poch.hint.bet"}
	case "fold":
		return &controller.PochWebOutputHint{Action: "fold", Reason: "poch.hint.fold"}
	default:
		if action.HandIdx < 0 {
			return &controller.PochWebOutputHint{Reason: "poch.hint.none"}
		}
		idx := action.HandIdx
		return &controller.PochWebOutputHint{Action: "play", CardIndex: &idx, Reason: "poch.hint.play"}
	}
}
