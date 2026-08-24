//go:build !js || !wasm || solo

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// QuodlibetWebPresenter はクオドリベットの Web プレゼンター。
type QuodlibetWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *QuodlibetWebPresenter) Output(g interfaces.QuodlibetGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// 受動ヒントは Output でも埋める (#4483)。
	if hint := g.GetHint(); hint != nil {
		if len(hint.CardIndices) > 0 {
			resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		}
		resObj.HintContract = hint.Contract
	}
	return marshalOrError(resObj)
}

// buildBase は共通フィールドを構築する。
func (p *QuodlibetWebPresenter) buildBase(g interfaces.QuodlibetGame) *controller.QuodlibetWebOutput {
	contract := g.GetCurrentContract()
	resObj := new(controller.QuodlibetWebOutput)
	resObj.Phase = g.GetPhase()
	resObj.DealNumber = g.GetDealNumber()
	resObj.TotalDeals = domain.QuodlibetTotalDeals
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.CurrentContract = contract
	resObj.CurrentContractName = domain.QuodlibetContractName(contract)
	resObj.IsShedding = domain.QuodlibetIsSheddingContract(contract)
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.TrickCount = domain.QuodlibetHandSize
	resObj.CurrentPlayerIdx = g.GetCurrentTurn()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsContractPhase = g.GetPhase() == domain.QuodlibetPhaseSelectContract
	resObj.HintContract = -1

	avail := g.GetAvailableContracts()
	resObj.AvailableContracts = make([]int, 0, len(avail))
	resObj.AvailableContractNames = make([]string, 0, len(avail))
	for _, c := range avail {
		resObj.AvailableContracts = append(resObj.AvailableContracts, c)
		resObj.AvailableContractNames = append(resObj.AvailableContractNames, domain.QuodlibetContractName(c))
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.LastTrick = trickCardsToOutput(g.GetLastTrick())
	resObj.PlayableIndices, resObj.CanPass = p.playableIndices(g)
	resObj.TablePlaced = p.tablePlaced(g)
	resObj.Stack = p.stack(g)
	resObj.Winners = quodlibetIntsOrEmpty(g.GetWinners())
	resObj.LastDeal = quodlibetDealToOutput(g.GetLastDealDetail())
	resObj.DealHistory = make([]*controller.QuodlibetWebOutputDeal, 0)
	for _, d := range g.GetDealHistory() {
		if out := quodlibetDealToOutput(d); out != nil {
			resObj.DealHistory = append(resObj.DealHistory, out)
		}
	}

	cfg := g.GetConfig()
	resObj.Config = controller.QuodlibetWebOutputConfig{
		CpuDifficulty:      int(cfg.CpuDifficulty),
		AutoSelectContract: cfg.AutoSelectContract,
	}
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// quodlibetIntsOrEmpty は nil を空スライスに直す (JSON で null を出さない)。
//
// **同名の共通ヘルパは extra4 タグの中にある。** そちらを呼ぶと solo の
// TinyGo ビルドだけが「undefined」で落ちる ── ホストの `go build ./...` は
// どのファイルも `!js || !wasm` を満たすので絶対に落ちない。
func quodlibetIntsOrEmpty(v []int) []int {
	if v == nil {
		return make([]int, 0)
	}
	return v
}

// quodlibetDealToOutput は 1 ディールの罰点内訳を出力形へ直す。
func quodlibetDealToOutput(d *domain.QuodlibetDealDetail) *controller.QuodlibetWebOutputDeal {
	if d == nil {
		return nil
	}
	points := make([]int, domain.QuodlibetPlayerCnt)
	for i := 0; i < domain.QuodlibetPlayerCnt; i++ {
		points[i] = d.Points[i]
	}
	return &controller.QuodlibetWebOutputDeal{
		Contract:     d.Contract,
		ContractName: domain.QuodlibetContractName(d.Contract),
		Round:        d.Round,
		DealerIdx:    d.DealerIdx,
		Points:       points,
	}
}

// playableIndices は人間が出せる札と、パスできるかを返す。
func (p *QuodlibetWebPresenter) playableIndices(g interfaces.QuodlibetGame) ([]int, bool) {
	if g.GetPhase() != domain.QuodlibetPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0), false
	}
	idx := quodlibetIntsOrEmpty(g.GetPlayableIndices(g.GetCurrentTurn()))
	// **パスできるのは合法手が 1 枚も無いときだけ。** 出せるのに降りられると
	// 四分も小食いも「早く出し切る」遊びが成立しない。
	canPass := domain.QuodlibetIsSheddingContract(g.GetCurrentContract()) && len(idx) == 0
	return idx, canPass
}

// tablePlaced は小食いの場を、スートごとの位インデックス列にほどく。
func (p *QuodlibetWebPresenter) tablePlaced(g interfaces.QuodlibetGame) [][]int {
	placed := g.GetTablePlaced()
	out := make([][]int, 0, domain.CardDesignDiamond)
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		row := make([]int, 0)
		for i := 0; i < domain.QuodlibetHandSize; i++ {
			if placed[suit]&(uint16(1)<<uint(i)) != 0 {
				row = append(row, i)
			}
		}
		out = append(out, row)
	}
	return out
}

// stack は四分の重ねを出力形へ直す。
func (p *QuodlibetWebPresenter) stack(g interfaces.QuodlibetGame) []*controller.WebOutputCard {
	cards := g.GetStack()
	out := make([]*controller.WebOutputCard, 0, len(cards))
	for _, c := range cards {
		out = append(out, cardToOutput(c))
	}
	return out
}

// buildPlayersOutput は席の情報を構築する。
func (p *QuodlibetWebPresenter) buildPlayersOutput(g interfaces.QuodlibetGame) []*controller.QuodlibetWebOutputPlayer {
	dealer := g.GetDealerIdx()
	contract := g.GetCurrentContract()
	human := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			human = i
			break
		}
	}
	out := make([]*controller.QuodlibetWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		// **第 3 の輪では手札の見え方そのものが規則。** 「開いたズボン」は
		// 自分の手札だけが見えず、「狩猟」は全員の手札が見える。
		visible := domain.QuodlibetHandVisibility(contract, human, i)
		out = append(out, &controller.QuodlibetWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, visible),
			TrickCount: player.GetTrickCount(),
			Penalty:    player.GetPenalty(),
			DealPoints: player.GetDealPoints(),
			OutRank:    player.GetOutRank(),
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage はフェーズ / 結果メッセージを構築する。
func (p *QuodlibetWebPresenter) buildMessage(g interfaces.QuodlibetGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		code, params := domain.ErrorMessageCode(lastErr)
		return lastErr.Error(), code, params
	}
	if g.GetGameEndFlag() {
		winners := g.GetWinners()
		if len(winners) == 1 && winners[0] == 0 {
			return "", "quodlibet.result.humanWin", nil
		}
		if len(winners) > 1 {
			return "", "quodlibet.result.draw", nil
		}
		return "", "quodlibet.result.cpuWin", nil
	}
	switch g.GetPhase() {
	case domain.QuodlibetPhaseSelectContract:
		return "", "quodlibet.selectContract", nil
	case domain.QuodlibetPhasePlay:
		if domain.QuodlibetIsSheddingContract(g.GetCurrentContract()) {
			return "", "quodlibet.shedPhase", nil
		}
		return "", "quodlibet.playPhase", nil
	case domain.QuodlibetPhaseDealEnd:
		return "", "quodlibet.dealEnd", nil
	}
	return "", "", nil
}

// HintOutput はヒント情報を JSON 出力する。
func (p *QuodlibetWebPresenter) HintOutput(g interfaces.QuodlibetGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil && (len(hint.CardIndices) > 0 || hint.Contract >= 0) {
		if len(hint.CardIndices) > 0 {
			resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		}
		resObj.HintContract = hint.Contract
		resObj.MessageCode = "quodlibet.hintRequested"
	} else {
		resObj.MessageCode = "quodlibet.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *QuodlibetWebPresenter) ActionLogOutput(g interfaces.QuodlibetGame) string {
	return actionLogOutputJSON(g)
}
