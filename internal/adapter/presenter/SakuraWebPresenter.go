//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// sakuraCardToOutput は札に点数を添えて出力形へ変換する。
//
// **点数を札そのものに書く** (#5785)。さくらは役ではなく点数の合計で競うので、
// どの札が何点かが読めないと打ち手を決められない——CUI は最初からそう出している。
func sakuraCardToOutput(c *domain.Card) *controller.WebOutputCard {
	out := cardToOutputWithFace(c, koikoiFace)
	if out == nil {
		return nil
	}
	points := domain.SakuraCardPoints(c)
	out.Points = &points
	return out
}

// sakuraHandToOutput は手札を点数つきで返す。**人間の席だけ中身を出す。**
func sakuraHandToOutput(player *domain.SakuraPlayer) []*controller.WebOutputCard {
	if !player.GetIsHuman() {
		return make([]*controller.WebOutputCard, 0)
	}
	cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		cards = append(cards, sakuraCardToOutput(player.GetCard(i)))
	}
	return cards
}

// SakuraWebPresenter はさくら (肥後花) の Web プレゼンタークラス。
type SakuraWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *SakuraWebPresenter) Output(g interfaces.SakuraGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(g)
		resObj.MessageCode = "sakura.result.scores"
		resObj.MessageParams = map[string]string{"scores": p.encodeScoresParam(g)}
	}
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// applyHint は人間の手番であればヒント情報を出力オブジェクトへ埋める。
//
// GetHint は CPU 手番・終局・対象外フェーズでは CardIndex に -1 を返す。その場合
// Hint は設定されず、omitempty により JSON からも省かれる。
func (p *SakuraWebPresenter) applyHint(resObj *controller.SakuraWebOutput, g interfaces.SakuraGame) {
	hint := g.GetHint()
	if hint.CardIndex < 0 {
		return
	}
	resObj.Hint = &controller.SakuraWebOutputHint{
		CardIndex:  hint.CardIndex,
		FieldIndex: hint.FieldIndex,
		Reason:     hint.Reason,
	}
}

// sakuraBonusesToWeb は追加役を Web 出力型へ変換する。
func sakuraBonusesToWeb(bonuses []domain.SakuraBonus) []*controller.SakuraWebOutputBonus {
	out := make([]*controller.SakuraWebOutputBonus, 0, len(bonuses))
	for _, b := range bonuses {
		out = append(out, &controller.SakuraWebOutputBonus{
			Key:    domain.SakuraBonusName(b),
			Points: domain.SakuraBonusPoints(b),
		})
	}
	return out
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *SakuraWebPresenter) buildBase(g interfaces.SakuraGame) *controller.SakuraWebOutput {
	resObj := new(controller.SakuraWebOutput)
	resObj.Players = make([]*controller.SakuraWebOutputPlayer, 0, g.GetPlayerCnt())
	resObj.CaptureOptions = make(map[int][]int)
	resObj.ChoiceOptions = make(map[int][]int)

	fieldOut := make([]*controller.WebOutputCard, 0, len(g.GetField()))
	for _, c := range g.GetField() {
		fieldOut = append(fieldOut, sakuraCardToOutput(c))
	}
	resObj.FieldCards = fieldOut

	cfg := g.GetConfig()
	resObj.Phase = int(g.GetPhase())
	resObj.Round = g.GetRound()
	resObj.TotalRounds = cfg.Rounds
	resObj.CurrentTurn = g.GetTurn()
	resObj.Dealer = g.GetDealer()
	resObj.StockCount = g.GetStockCount()
	resObj.Winner = g.GetWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.Config = controller.SakuraWebConfigOutput{Seats: cfg.Seats, Rounds: cfg.Rounds}

	if g.GetPhase() == domain.SakuraPhasePlay && g.IsHumanTurn() {
		if opts := g.GetValidFieldIndices(); opts != nil {
			resObj.CaptureOptions = opts
		}
		if opts := g.GetChoiceIndices(); opts != nil {
			resObj.ChoiceOptions = opts
		}
	}

	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		taken := make([]*controller.WebOutputCard, 0, len(player.GetTaken()))
		for _, c := range player.GetTaken() {
			taken = append(taken, sakuraCardToOutput(c))
		}
		resObj.Players = append(resObj.Players, &controller.SakuraWebOutputPlayer{
			ID:          i,
			Name:        player.GetName(),
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       sakuraHandToOutput(player),
			Taken:       taken,
			TakenCount:  len(player.GetTaken()),
			CardPoints:  player.CardPoints(),
			Bonuses:     sakuraBonusesToWeb(player.Bonuses()),
			BonusPoints: player.BonusPoints(),
			TotalPoints: player.TotalPoints(),
			Score:       player.GetScore(),
			RoundScore:  player.GetRoundScore(),
			RoundWins:   player.GetRoundWins(),
		})
	}

	if res := g.GetLastResult(); res != nil {
		seats := make([]*controller.SakuraWebOutputSeatResult, 0, len(res.Seats))
		for _, s := range res.Seats {
			seats = append(seats, &controller.SakuraWebOutputSeatResult{
				CardPoints:  s.CardPoints,
				Bonuses:     sakuraBonusesToWeb(s.Bonuses),
				BonusPoints: s.BonusPoints,
				Total:       s.Total,
			})
		}
		resObj.LastResult = &controller.SakuraWebOutputRoundResult{
			Round:  res.Round,
			Winner: res.Winner,
			Seats:  seats,
		}
	}
	return resObj
}

// encodeScoresParam は累計得点を "0:120,1:98" 形式の文字列に詰める。
func (p *SakuraWebPresenter) encodeScoresParam(g interfaces.SakuraGame) string {
	parts := make([]string, 0, g.GetPlayerCnt())
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, player.GetScore()))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はゲーム終了時のフォールバック (英語) メッセージ。
func (p *SakuraWebPresenter) buildResultMessage(g interfaces.SakuraGame) string {
	msg := "Game over. "
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		name := "CPU"
		if player.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%d ", name, player.GetScore())
	}
	return msg
}

// HintOutput はヒント情報を JSON 出力する。
func (p *SakuraWebPresenter) HintOutput(g interfaces.SakuraGame) string {
	resObj := p.buildBase(g)
	p.applyHint(resObj, g)
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *SakuraWebPresenter) ActionLogOutput(g interfaces.SakuraGame) string {
	return actionLogOutputJSON(g)
}
