//go:build !js || !wasm || casino

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CourtPieceWebPresenter Court Piece Web プレゼンタークラス
type CourtPieceWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *CourtPieceWebPresenter) Output(t interfaces.CourtPieceGame, lastErr error) string {
	resObj := p.buildBase(t)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(t, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**CourtPiece.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := t.GetHint(); hint != nil {
		resObj.Hint = &controller.CourtPieceWebOutputHint{
			CardIndex: hint.CardIndex,
			TrumpSuit: hint.TrumpSuit,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *CourtPieceWebPresenter) buildBase(t interfaces.CourtPieceGame) *controller.CourtPieceWebOutput {
	resObj := new(controller.CourtPieceWebOutput)
	resObj.Phase = int(t.GetPhase())
	resObj.RoundNumber = t.GetRoundNumber()
	resObj.TrickNumber = t.GetTrickNumber()
	resObj.CurrentPlayerIdx = t.GetCurrentPlayerIdx()
	resObj.CallerIdx = t.GetCallerIdx()
	resObj.TrumpSuit = t.GetTrumpSuit()
	resObj.ConsecutiveWins = t.GetConsecutiveWins()
	resObj.LastWinnerTeam = t.GetLastWinnerTeam()
	resObj.LastRoundCourt = t.IsLastRoundCourt()
	resObj.GameEndFlag = t.GetGameEndFlag()
	resObj.WinnerTeam = t.GetWinnerTeam()
	resObj.LeadPlayerIdx = t.GetLeadPlayerIdx()

	cfg := t.GetConfig()
	resObj.Config = controller.CourtPieceWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.TeamScores = []int{t.GetTeamScore(0), t.GetTeamScore(1)}
	resObj.CurrentTrick = trickCardsToOutput(t.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(t)
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *CourtPieceWebPresenter) buildPlayersOutput(t interfaces.CourtPieceGame) []*controller.CourtPieceWebOutputPlayer {
	out := make([]*controller.CourtPieceWebOutputPlayer, 0)
	for i := 0; i < t.GetPlayerCnt(); i++ {
		player := t.GetPlayer(i)
		pObj := &controller.CourtPieceWebOutputPlayer{
			ID:              i,
			IsHuman:         player.GetIsHuman(),
			Team:            player.GetTeam(),
			CardCount:       player.GetCardsSize(),
			Cards:           playerCardsToOutput(player, player.GetIsHuman()),
			RoundScore:      player.GetRoundScore(),
			CumulativeScore: player.GetCumulativeScore(),
			TrickCount:      player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *CourtPieceWebPresenter) buildMessage(t interfaces.CourtPieceGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if t.GetGameEndFlag() {
		winnerTeam := t.GetWinnerTeam()
		humanTeam := -1
		for i := 0; i < t.GetPlayerCnt(); i++ {
			if t.GetPlayer(i).GetIsHuman() {
				humanTeam = t.GetPlayer(i).GetTeam()
				break
			}
		}
		if winnerTeam == humanTeam {
			return "", "courtpiece.gameEndHumanWin", map[string]string{"team": strconv.Itoa(winnerTeam)}
		}
		return "", "courtpiece.gameEndCpuWin", map[string]string{"team": strconv.Itoa(winnerTeam)}
	}
	switch t.GetPhase() {
	case domain.CourtPiecePhaseTrumpDeclaration:
		return "", "courtpiece.trumpPhase", nil
	case domain.CourtPiecePhasePlay:
		if len(t.GetCurrentTrick()) == 0 {
			return "", "courtpiece.playPhase.lead", nil
		}
		return "", "courtpiece.playPhase.follow", nil
	case domain.CourtPiecePhaseTrickEnd:
		return "", "courtpiece.trickEnd", nil
	case domain.CourtPiecePhaseRoundEnd:
		return "", "courtpiece.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報を JSON 出力する
func (p *CourtPieceWebPresenter) HintOutput(t interfaces.CourtPieceGame) string {
	hint := t.GetHint()
	resObj := p.buildBase(t)
	if hint != nil {
		resObj.Hint = &controller.CourtPieceWebOutputHint{
			CardIndex: hint.CardIndex,
			TrumpSuit: hint.TrumpSuit,
			Reason:    hint.Reason,
		}
	}
	// **「頼んだヒントか」をフロントが見分けられるようにする。**ページは
	// `isRequestedHint` でこのコードを見てからバナーを出すので (#4605)、
	// 付けないと押しても何も出ない。`hintAvailable` は画面のラベルとして
	// 既に使われているため別キーにする (#4483)。
	if hint != nil {
		resObj.MessageCode = "courtPiece.hintRequested"
	} else {
		resObj.MessageCode = "courtPiece.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力
func (p *CourtPieceWebPresenter) ActionLogOutput(t interfaces.CourtPieceGame) string {
	return actionLogOutputJSON(t)
}
