package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GongZhuWebPresenter 拱猪Webプレゼンタークラス
type GongZhuWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GongZhuWebPresenter) Output(g interfaces.GongZhuGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, g.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**GongZhu.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}

	// 得点内訳はラウンド終了でだけ確定する。途中で出すと未確定の数字が並ぶ (#5630)。
	if g.GetPhase() == domain.GongZhuPhaseRoundEnd {
		resObj.ScoreBreakdowns = make([]*controller.GongZhuWebOutputBreakdown, 0, g.GetPlayerCnt())
		for i := 0; i < g.GetPlayerCnt(); i++ {
			d := g.ScoreBreakdownFor(i)
			resObj.ScoreBreakdowns = append(resObj.ScoreBreakdowns, &controller.GongZhuWebOutputBreakdown{
				HeartCount:        d.HeartCount,
				HeartsSum:         d.HeartsSum,
				AllHearts:         d.AllHearts,
				AceExposed:        d.AceExposed,
				HasPig:            d.HasPig,
				PigExposed:        d.PigExposed,
				HasSheep:          d.HasSheep,
				SheepExposed:      d.SheepExposed,
				HasDoubler:        d.HasDoubler,
				DoublerMultiplier: d.DoublerMultiplier,
				DoublerStandalone: d.DoublerStandalone,
				Subtotal:          d.Subtotal,
				Total:             d.Total,
			})
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *GongZhuWebPresenter) buildBase(g interfaces.GongZhuGame) *controller.GongZhuWebOutput {
	resObj := new(controller.GongZhuWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.HeartsBroken = g.GetHeartsBroken()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	ex := g.GetExposure()
	resObj.Exposed = controller.GongZhuWebOutputExposure{
		Pig:     ex.Pig,
		Sheep:   ex.Sheep,
		Ace:     ex.Ace,
		Doubler: ex.Doubler,
	}
	resObj.ExposableIndices = p.exposableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.GongZhuWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.PlayableIndices = p.playableIndices(g)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// exposableIndices 人間プレイヤーが公開できるカードのインデックスを返す
func (p *GongZhuWebPresenter) exposableIndices(g interfaces.GongZhuGame) []int {
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			idx := g.GetExposableIndices(i)
			if idx == nil {
				return make([]int, 0)
			}
			return idx
		}
	}
	return make([]int, 0)
}

// buildPlayersOutput プレイヤー情報を構築
func (p *GongZhuWebPresenter) buildPlayersOutput(g interfaces.GongZhuGame) []*controller.GongZhuWebOutputPlayer {
	out := make([]*controller.GongZhuWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.GongZhuWebOutputPlayer{
			ID:                 i,
			IsHuman:            player.GetIsHuman(),
			CardCount:          player.GetCardsSize(),
			Cards:              playerCardsToOutput(player, player.GetIsHuman()),
			CapturedPointCards: gongZhuCapturedPointCards(player),
			RoundScore:         player.GetRoundScore(),
			CumulativeScore:    player.GetCumulativeScore(),
			TrickCount:         player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

// gongZhuCapturedPointCards はプレイヤーが獲得した得点札（ハート各札・♠Q=豚・
// ♦J=羊・♣10=倍化）を WebOutputCard スライスに変換する。得点札はトリック獲得時に
// 公開される公開情報のため、全プレイヤー分をそのまま送出してよい。nil の場合は
// 空スライスを返す。
func gongZhuCapturedPointCards(player *domain.GongZhuPlayer) []*controller.WebOutputCard {
	return cardsToOutputOrEmpty(gongZhuCapturedPoints(player))
}

// buildMessage ゲーム結果メッセージを構築
func (p *GongZhuWebPresenter) buildMessage(g interfaces.GongZhuGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("gongzhu", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.GongZhuPhaseExpose:
		return "", "gongzhu.exposePhase", nil
	case domain.GongZhuPhasePlay:
		if len(trick) == 0 {
			return "", "gongzhu.playPhase.lead", nil
		}
		return "", "gongzhu.playPhase.follow", nil
	case domain.GongZhuPhaseTrickEnd:
		return "", "gongzhu.trickEnd", nil
	case domain.GongZhuPhaseRoundEnd:
		return "", "gongzhu.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *GongZhuWebPresenter) HintOutput(g interfaces.GongZhuGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if hint != nil {
		resObj.MessageCode = "gongzhu.hintRequested"
	} else {
		resObj.MessageCode = "gongzhu.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GongZhuWebPresenter) ActionLogOutput(g interfaces.GongZhuGame) string {
	return actionLogOutputJSON(g)
}

// playableIndices は人間がいま出せる手札の位置を返す。
//
// **マストフォローの可視化。**Web はどのカードが出せるかを一切示しておらず、
// プレイヤーが自力で判断するしかなかった (#4812)。判定はドメインの validatePlay
// をそのまま通す GetPlayableIndices を使う。
func (p *GongZhuWebPresenter) playableIndices(g interfaces.GongZhuGame) []int {
	if g.GetPhase() != domain.GongZhuPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}
