package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// PitchPlayerCnt ピッチプレイヤー数 (4人, シングルハンド/カットスロート)
const PitchPlayerCnt = 4

// PitchHandSize 各プレイヤーの手札枚数
const PitchHandSize = 6

// PitchTotalTricks ラウンド毎のトリック数
const PitchTotalTricks = 6

// PitchMinBid 最小ビッド値 (パス=0 を除く有効最小値)
const PitchMinBid = 2

// PitchMaxBid 最大ビッド値 (4 = 全ポイント獲得宣言 / Smudge)
const PitchMaxBid = 4

// PitchPassBid パスを表すビッド値
const PitchPassBid = 0

// PitchPhase ゲームフェーズ
type PitchPhase int

// Pitchのフェーズ定数
const (
	// PitchPhaseBid ビッドフェーズ
	PitchPhaseBid PitchPhase = 0
	// PitchPhasePlay トリックプレイフェーズ
	PitchPhasePlay PitchPhase = 1
	// PitchPhaseTrickEnd トリック終了フェーズ
	PitchPhaseTrickEnd PitchPhase = 2
	// PitchPhaseRoundEnd ラウンド終了フェーズ
	PitchPhaseRoundEnd PitchPhase = 3
	// PitchPhaseGameEnd ゲーム終了フェーズ
	PitchPhaseGameEnd PitchPhase = 4
)

// PitchTrumpUnset トランプ未確定を表す値
const PitchTrumpUnset = 0

// PitchHint ヒント情報
type PitchHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド時nil)
	Bid       *int   // 推奨ビッド値 (プレイ時nil; パスは 0)
	Reason    string // ヒント理由キー
}

// Pitch ピッチ (Auction Pitch / Setback) ゲームクラス
type Pitch struct {
	trumpCards       *TrumpCards
	players          []*PitchPlayer
	config           PitchConfig
	phase            PitchPhase
	roundNumber      int
	trickNumber      int
	dealerIdx        int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	bidPlayerIdx     int // 現在ビッド中のプレイヤー
	currentBid       int // 最高ビッド値 (0=未ビッド/パスのみ)
	bidWinnerIdx     int // 最高ビッダーのインデックス (-1=未確定)
	trumpSuit        int // 切り札スート (PitchTrumpUnset=未確定)
	gameEndFlag      bool
	winnerIdx        int
	// roundBreakdown は直近ラウンドの 4 得点を誰が取ったか (#5584)。
	roundBreakdown PitchRoundBreakdown
	actionLogBase
}

// PitchRoundBreakdown は High / Low / Jack / Game をそれぞれ誰が取ったかを表す。
// -1 は「誰も取っていない」(切り札が出なかった、Game が同点、など)。
//
// **4 種の得点はこのゲームの骨格なのに、合計しか画面に出ていなかった** (#5584)。
// 合計だけでは、1 点差の理由が Jack を取られたからなのか Game で並ばれたからなのか
// が分からない。
type PitchRoundBreakdown struct {
	High int
	Low  int
	Jack int
	Game int
}

// PitchNoScorer は PitchRoundBreakdown の「誰も取っていない」を表す。
const PitchNoScorer = -1

// GetRoundBreakdown は直近ラウンドの得点内訳を返す。
func (p *Pitch) GetRoundBreakdown() PitchRoundBreakdown { return p.roundBreakdown }

// NewPitch コンストラクタ
func NewPitch(trumpCards *TrumpCards, players []*PitchPlayer, config PitchConfig) *Pitch {
	return &Pitch{
		trumpCards:   trumpCards,
		players:      players,
		config:       config,
		winnerIdx:    -1,
		bidWinnerIdx: -1,
		dealerIdx:    PitchPlayerCnt - 1,
		roundNumber:  0,
	}
}

// NewDefaultPitch returns Pitch with the standard 4-player setup (1 human, 3 CPU)
// and DefaultPitchConfig. Single source of truth for CUI, Web, and Worker construction.
func NewDefaultPitch() *Pitch {
	players := []*PitchPlayer{
		NewPitchPlayer(true),
		NewPitchPlayer(false),
		NewPitchPlayer(false),
		NewPitchPlayer(false),
	}
	return NewPitch(NewTrumpCards(0), players, DefaultPitchConfig())
}

// Reset ゲーム初期化
func (p *Pitch) Reset() {
	p.gameEndFlag = false
	p.winnerIdx = -1
	p.roundNumber = 1
	p.trickNumber = 0
	p.currentTrick = nil
	p.leadPlayerIdx = -1
	p.currentPlayerIdx = -1
	p.dealerIdx = PitchPlayerCnt - 1
	p.bidPlayerIdx = (p.dealerIdx + 1) % PitchPlayerCnt
	p.currentBid = 0
	p.bidWinnerIdx = -1
	p.trumpSuit = PitchTrumpUnset
	p.actionLog = nil

	for _, pl := range p.players {
		pl.bid = -1
		pl.SetRoundScore(0)
		pl.SetCumulativeScore(0)
		pl.ResetTricks()
		pl.Reset()
		pl.SetIsFinished(false)
	}

	p.trumpCards.Shuffle()
	p.dealRound()
	p.sortAllHands()

	p.phase = PitchPhaseBid
	// 新しいゲームでは誰も取っていない。ゼロ値のままだと席 0 が全部取ったように出る。
	p.roundBreakdown = PitchRoundBreakdown{
		High: PitchNoScorer, Low: PitchNoScorer, Jack: PitchNoScorer, Game: PitchNoScorer,
	}
}

// dealRound Pitch 用に各プレイヤーへ PitchHandSize 枚配る (52枚デッキから 24枚のみ使用)
func (p *Pitch) dealRound() {
	for i := 0; i < PitchHandSize; i++ {
		for j := 0; j < PitchPlayerCnt; j++ {
			card := p.trumpCards.DrawCard()
			if card == nil {
				return
			}
			p.players[j].AddCard(card)
		}
	}
}

// NextRound 次のラウンドを開始する
func (p *Pitch) NextRound() {
	if p.phase != PitchPhaseRoundEnd {
		return
	}

	p.roundNumber++
	p.trickNumber = 0
	p.currentTrick = nil
	p.leadPlayerIdx = -1
	p.currentPlayerIdx = -1
	p.dealerIdx = (p.dealerIdx + 1) % PitchPlayerCnt
	p.bidPlayerIdx = (p.dealerIdx + 1) % PitchPlayerCnt
	p.currentBid = 0
	p.bidWinnerIdx = -1
	p.trumpSuit = PitchTrumpUnset

	for _, pl := range p.players {
		pl.ResetRound()
	}

	p.trumpCards.Shuffle()
	p.dealRound()
	p.sortAllHands()

	p.phase = PitchPhaseBid
}

// PlayerBid 人間プレイヤーがビッドする (bid: 0=pass, 2/3/4)
func (p *Pitch) PlayerBid(bid int) error {
	if p.gameEndFlag {
		return ErrGameEnded
	}
	if p.phase != PitchPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(p.players)
	if humanIdx < 0 || p.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if err := p.validateBidValue(humanIdx, bid); err != nil {
		return err
	}
	p.applyBid(humanIdx, bid)
	p.advanceBid()
	return nil
}

// CpuBid 現在のビッド手番がCPUの場合にビッドする
func (p *Pitch) CpuBid() {
	if p.gameEndFlag || p.phase != PitchPhaseBid {
		return
	}
	if p.bidPlayerIdx >= PitchPlayerCnt || p.bidPlayerIdx < 0 {
		return
	}
	if p.players[p.bidPlayerIdx].GetIsHuman() {
		return
	}
	bid := p.cpuSelectBid(p.bidPlayerIdx)
	p.applyBid(p.bidPlayerIdx, bid)
	p.advanceBid()
}

// validateBidValue ビッド値が有効か検証する
func (p *Pitch) validateBidValue(playerIdx, bid int) error {
	if bid == PitchPassBid {
		// 親 (dealer) は他全員パスの状態で必ず stuck されるためパス不可。
		// 親以外はパス可能。
		if playerIdx == p.dealerIdx && p.currentBid == 0 {
			return NewDomainError(ErrInvalidPlay, "親 (dealer) は全員パスの場合パスできません")
		}
		return nil
	}
	if bid < PitchMinBid || bid > PitchMaxBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは pass(0) または %d〜%d で指定してください", PitchMinBid, PitchMaxBid))
	}
	if bid <= p.currentBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは現在の最高 %d を超える必要があります", p.currentBid))
	}
	return nil
}

// applyBid ビッド値をプレイヤーに反映し、現在の最高ビッド/勝者を更新する
func (p *Pitch) applyBid(playerIdx, bid int) {
	p.players[playerIdx].SetBid(bid)
	if bid > p.currentBid {
		p.currentBid = bid
		p.bidWinnerIdx = playerIdx
	}
	logBid := fmt.Sprintf("%d", bid)
	if bid == PitchPassBid {
		logBid = "pass"
	}
	p.appendLog(playerIdx, "bid", fmt.Sprintf("%s bids %s", playerName(p.players, playerIdx), logBid), nil)
}

// advanceBid 次のビッド手番へ進める。全員終わればプレイ開始 (stuck dealer も処理)
func (p *Pitch) advanceBid() {
	bidsDone := p.bidsCompleted()
	if bidsDone < PitchPlayerCnt {
		p.bidPlayerIdx = (p.bidPlayerIdx + 1) % PitchPlayerCnt
		// 親に到達し、かつ全員パス済みの場合 stuck 強制
		if p.bidPlayerIdx == p.dealerIdx && p.currentBid == 0 && bidsDone == PitchPlayerCnt-1 {
			p.applyBid(p.dealerIdx, PitchMinBid)
			p.appendLog(p.dealerIdx, "stuck", fmt.Sprintf("%s is stuck with %d", playerName(p.players, p.dealerIdx), PitchMinBid), nil)
			p.startPlayPhase()
		}
		return
	}
	p.startPlayPhase()
}

// bidsCompleted いくつのプレイヤーがビッド済みか
func (p *Pitch) bidsCompleted() int {
	cnt := 0
	for _, pl := range p.players {
		if pl.GetBid() != -1 {
			cnt++
		}
	}
	return cnt
}

// startPlayPhase プレイフェーズ開始: ビッド勝者がリード
func (p *Pitch) startPlayPhase() {
	if p.bidWinnerIdx < 0 {
		// 安全策: 親 stuck
		p.bidWinnerIdx = p.dealerIdx
		p.currentBid = PitchMinBid
		p.players[p.dealerIdx].SetBid(PitchMinBid)
	}
	p.leadPlayerIdx = p.bidWinnerIdx
	p.currentPlayerIdx = p.bidWinnerIdx
	p.trickNumber = 1
	p.currentTrick = nil
	p.phase = PitchPhasePlay
	p.appendLog(p.bidWinnerIdx, "bid_won",
		fmt.Sprintf("%s wins the bid at %d (will lead first card to set trump)", playerName(p.players, p.bidWinnerIdx), p.currentBid),
		nil)
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (p *Pitch) PlayerPlay(cardIndex int) error {
	if p.gameEndFlag {
		return ErrGameEnded
	}
	if p.phase != PitchPhasePlay {
		return ErrWrongPhase
	}
	if !p.players[p.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := p.players[p.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := p.validatePlay(p.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	p.playCard(p.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (p *Pitch) CpuPlay() {
	if p.gameEndFlag || p.phase != PitchPhasePlay {
		return
	}
	if p.players[p.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := p.players[p.currentPlayerIdx]
	cardIdx := p.cpuSelectPlayCard(p.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	p.playCard(p.currentPlayerIdx, played)
}

// playCard カードをプレイする共通処理
func (p *Pitch) playCard(playerIdx int, card *Card) {
	// 最初のトリックのリードカードがトランプを設定
	if p.trumpSuit == PitchTrumpUnset && len(p.currentTrick) == 0 {
		p.trumpSuit = card.GetDesign()
		p.appendLog(playerIdx, "trump_set",
			fmt.Sprintf("Trump is %s", suitName(p.trumpSuit)), nil)
	}
	p.currentTrick = append(p.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	p.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(p.players, playerIdx), cardStr(card)),
		[]*Card{card})

	if len(p.currentTrick) == PitchPlayerCnt {
		p.phase = PitchPhaseTrickEnd
	} else {
		p.currentPlayerIdx = (p.currentPlayerIdx + 1) % PitchPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
// ルール: lead suit を持っている場合は lead suit OR trump のみプレイ可。
// lead suit を持っていなければ任意のカードをプレイ可。
func (p *Pitch) validatePlay(playerIdx int, card *Card) error {
	if len(p.currentTrick) == 0 {
		return nil // リードは任意
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return nil
	}
	if p.trumpSuit != PitchTrumpUnset && card.GetDesign() == p.trumpSuit {
		return nil // トランプはいつでも合法
	}
	if p.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従うかトランプを切ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (p *Pitch) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(p.players[playerIdx], design)
}

// ResolveTrick トリックを解決して勝者を決定する
func (p *Pitch) ResolveTrick() {
	if p.phase != PitchPhaseTrickEnd || len(p.currentTrick) != PitchPlayerCnt {
		return
	}
	winnerIdx := p.trickWinner()
	trickCards := make([]*Card, len(p.currentTrick))
	for i, tc := range p.currentTrick {
		trickCards[i] = tc.Card
	}
	p.players[winnerIdx].AddTrick(trickCards)
	p.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(p.players, winnerIdx), p.trickNumber),
		trickCards)
	p.leadPlayerIdx = winnerIdx
	if p.trickNumber >= PitchTotalTricks {
		p.phase = PitchPhaseRoundEnd
		// **内訳は画面に出る時点で確定していること。**得点の確定 (ScoreRound) は
		// プレイヤーが次のラウンドへ進めたときに走るので、そこで初めて作ると、
		// ラウンド終了画面には前のラウンドの内訳 (初回はゼロ値) が出る (#5584)。
		p.roundBreakdown = p.computeRoundBreakdown()
	} else {
		p.phase = PitchPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (p *Pitch) NextTrick() {
	if p.phase != PitchPhaseTrickEnd {
		return
	}
	p.currentTrick = nil
	p.currentPlayerIdx = p.leadPlayerIdx
	p.trickNumber++
	p.phase = PitchPhasePlay
}

// trickWinner トリックの勝者を決定する (トランプ最高 > リード最高)
func (p *Pitch) trickWinner() int {
	if len(p.currentTrick) == 0 {
		return 0
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	winnerIdx := p.currentTrick[0].PlayerIdx
	winnerVal := p.currentTrick[0].Card.GetValue()
	winnerIsTrump := p.trumpSuit != PitchTrumpUnset && p.currentTrick[0].Card.GetDesign() == p.trumpSuit

	for _, tc := range p.currentTrick[1:] {
		isTrump := p.trumpSuit != PitchTrumpUnset && tc.Card.GetDesign() == p.trumpSuit
		if isTrump && !winnerIsTrump {
			winnerIdx = tc.PlayerIdx
			winnerVal = pitchRankValue(tc.Card.GetValue())
			winnerIsTrump = true
			continue
		}
		if isTrump && winnerIsTrump {
			if pitchRankValue(tc.Card.GetValue()) > pitchRankValue(winnerVal) {
				winnerIdx = tc.PlayerIdx
				winnerVal = pitchRankValue(tc.Card.GetValue())
			}
			continue
		}
		if !isTrump && !winnerIsTrump && tc.Card.GetDesign() == leadSuit {
			if pitchRankValue(tc.Card.GetValue()) > pitchRankValue(winnerVal) {
				winnerIdx = tc.PlayerIdx
				winnerVal = pitchRankValue(tc.Card.GetValue())
			}
		}
	}
	return winnerIdx
}

// pitchRankValue カード値 (1=A) を比較用ランクに変換 (A=14 > K=13 > … > 2=2)
func pitchRankValue(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// PitchHandPips は手札のゲームピップ合計を返す。
//
// **CUI は入札前に手札の得点価値を暗算させていた (#4751)。**Web は入札中に
// バッジと内訳ポップオーバーを出している。
//
// 集計は Game ポイントの計算そのものと同じ pitchPipValue を通す。**別実装に
// すると、プレビューした値と実際に数えられる値がずれる。**
func PitchHandPips(cards []*Card) int {
	total := 0
	for _, c := range cards {
		if c == nil {
			continue
		}
		total += pitchPipValue(c.GetValue())
	}
	return total
}

// pitchPipValue Game ポイント計算用のピップ値: A=4 K=3 Q=2 J=1 10=10 他=0
func pitchPipValue(v int) int {
	switch v {
	case 1:
		return 4
	case 13:
		return 3
	case 12:
		return 2
	case 11:
		return 1
	case 10:
		return 10
	}
	return 0
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (p *Pitch) ScoreRound() {
	if p.phase != PitchPhaseRoundEnd {
		return
	}
	earned := p.computeRoundPoints()

	for i := 0; i < PitchPlayerCnt; i++ {
		pl := p.players[i]
		points := earned[i]
		if i == p.bidWinnerIdx {
			if points < p.currentBid {
				pl.SetRoundScore(-p.currentBid)
				p.appendLog(i, "set_back",
					fmt.Sprintf("%s set back: bid=%d earned=%d -> %d",
						playerName(p.players, i), p.currentBid, points, -p.currentBid), nil)
			} else {
				pl.SetRoundScore(points)
				p.appendLog(i, "bid_made",
					fmt.Sprintf("%s makes bid: bid=%d earned=%d -> +%d",
						playerName(p.players, i), p.currentBid, points, points), nil)
			}
		} else {
			pl.SetRoundScore(points)
			if points > 0 {
				p.appendLog(i, "non_bidder_score",
					fmt.Sprintf("%s scores %d", playerName(p.players, i), points), nil)
			}
		}
	}
	for i := 0; i < PitchPlayerCnt; i++ {
		p.players[i].CommitRoundScore()
	}
	for i := 0; i < PitchPlayerCnt; i++ {
		p.appendLog(i, "cumulative_score",
			fmt.Sprintf("%s: total=%d", playerName(p.players, i), p.players[i].GetCumulativeScore()), nil)
	}
	p.checkGameEnd()
}

// computeRoundPoints 各プレイヤーが獲得したポイント数 (High/Low/Jack/Game) を返す
func (p *Pitch) computeRoundPoints() []int {
	points := make([]int, PitchPlayerCnt)
	bd := p.computeRoundBreakdown()
	p.roundBreakdown = bd
	for _, idx := range []int{bd.High, bd.Low, bd.Jack, bd.Game} {
		if idx != PitchNoScorer {
			points[idx]++
		}
	}
	if bd.High != PitchNoScorer {
		p.appendLog(bd.High, "score_high",
			fmt.Sprintf("%s scores High", playerName(p.players, bd.High)), nil)
	}
	if bd.Low != PitchNoScorer {
		p.appendLog(bd.Low, "score_low",
			fmt.Sprintf("%s scores Low", playerName(p.players, bd.Low)), nil)
	}
	if bd.Jack != PitchNoScorer {
		p.appendLog(bd.Jack, "score_jack",
			fmt.Sprintf("%s scores Jack", playerName(p.players, bd.Jack)), nil)
	}
	if bd.Game != PitchNoScorer {
		p.appendLog(bd.Game, "score_game",
			fmt.Sprintf("%s scores Game (%d pip)", playerName(p.players, bd.Game), p.gamePipTotal(bd.Game)), nil)
	}
	return points
}

// gamePipTotal は席 idx が取った札のピップ合計。Game ポイントの根拠を棋譜に残すため。
func (p *Pitch) gamePipTotal(idx int) int {
	total := 0
	for _, trick := range p.players[idx].GetTricksTaken() {
		for _, card := range trick {
			total += pitchPipValue(card.GetValue())
		}
	}
	return total
}

// computeRoundBreakdown は High/Low/Jack/Game をそれぞれ誰が取ったかを数える。
//
// **副作用を持たない。**ラウンドが終わった時点 (ResolveTrick) と、点を確定する
// 時点 (ScoreRound) の両方から呼ぶので、棋譜に書いたり得点を動かしたりしない
// (#5584 のレビュー指摘: 内訳を ScoreRound でしか作らないと、画面に出る
// ラウンド終了時にはまだ前のラウンドの値が残っている)。
func (p *Pitch) computeRoundBreakdown() PitchRoundBreakdown {
	bd := PitchRoundBreakdown{
		High: PitchNoScorer, Low: PitchNoScorer, Jack: PitchNoScorer, Game: PitchNoScorer,
	}
	if p.trumpSuit == PitchTrumpUnset {
		return bd
	}
	// 全カード走査でトランプの最高/最低と J of trump 所有者を特定
	highPlayer := -1
	highRank := -1
	lowPlayer := -1
	// Sentinel above any possible trump rank (A maps to pitchRankValue=14).
	const lowRankSentinel = math.MaxInt32
	lowRank := lowRankSentinel
	jackPlayer := -1
	for playerIdx, pl := range p.players {
		for _, trick := range pl.GetTricksTaken() {
			for _, card := range trick {
				if card.GetDesign() != p.trumpSuit {
					continue
				}
				rank := pitchRankValue(card.GetValue())
				if rank > highRank {
					highRank = rank
					highPlayer = playerIdx
				}
				if rank < lowRank {
					lowRank = rank
					lowPlayer = playerIdx
				}
				if card.GetValue() == 11 {
					jackPlayer = playerIdx
				}
			}
		}
	}
	if highPlayer >= 0 {
		bd.High = highPlayer
	}
	if lowPlayer >= 0 {
		bd.Low = lowPlayer
	}
	if jackPlayer >= 0 {
		bd.Jack = jackPlayer
	}
	// Game ポイント: 全カードのピップ値合計, 最大が単独なら +1
	gameTotals := make([]int, PitchPlayerCnt)
	for playerIdx, pl := range p.players {
		for _, trick := range pl.GetTricksTaken() {
			for _, card := range trick {
				gameTotals[playerIdx] += pitchPipValue(card.GetValue())
			}
		}
	}
	gameWinner := -1
	maxTotal := -1
	tied := false
	for i, total := range gameTotals {
		if total > maxTotal {
			maxTotal = total
			gameWinner = i
			tied = false
		} else if total == maxTotal {
			tied = true
		}
	}
	if !tied && gameWinner >= 0 && maxTotal > 0 {
		bd.Game = gameWinner
	}
	return bd
}

// checkGameEnd ゲーム終了判定: PointLimit 到達者がいれば終了。
// ビッダーは到達優先 (同点なら non-bidder より優先しない / シンプル化)。
func (p *Pitch) checkGameEnd() {
	bidder := p.bidWinnerIdx
	if bidder >= 0 && p.players[bidder].GetCumulativeScore() >= p.config.PointLimit {
		p.gameEndFlag = true
		p.phase = PitchPhaseGameEnd
		p.winnerIdx = bidder
		p.appendLog(-1, "game_end",
			fmt.Sprintf("%s wins the game!", playerName(p.players, p.winnerIdx)), nil)
		return
	}
	maxScore := -1 << 30
	winner := -1
	hasWinner := false
	for i := 0; i < PitchPlayerCnt; i++ {
		score := p.players[i].GetCumulativeScore()
		if score >= p.config.PointLimit {
			hasWinner = true
		}
		if score > maxScore {
			maxScore = score
			winner = i
		}
	}
	if !hasWinner {
		return
	}
	p.gameEndFlag = true
	p.phase = PitchPhaseGameEnd
	p.winnerIdx = winner
	p.appendLog(-1, "game_end",
		fmt.Sprintf("%s wins the game!", playerName(p.players, p.winnerIdx)), nil)
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (p *Pitch) GetPhase() PitchPhase { return p.phase }

// SetPhase フェーズ設定 (テスト用)
func (p *Pitch) SetPhase(phase PitchPhase) { p.phase = phase }

// GetRoundNumber ラウンド番号取得
func (p *Pitch) GetRoundNumber() int { return p.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (p *Pitch) SetRoundNumber(n int) { p.roundNumber = n }

// GetTrickNumber トリック番号取得
func (p *Pitch) GetTrickNumber() int { return p.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (p *Pitch) SetTrickNumber(n int) { p.trickNumber = n }

// GetDealerIdx ディーラーインデックス取得
func (p *Pitch) GetDealerIdx() int { return p.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (p *Pitch) SetDealerIdx(idx int) { p.dealerIdx = idx }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (p *Pitch) GetCurrentPlayerIdx() int { return p.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (p *Pitch) SetCurrentPlayerIdx(idx int) { p.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (p *Pitch) GetCurrentTrick() []*TrickCard { return p.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (p *Pitch) SetCurrentTrick(trick []*TrickCard) { p.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (p *Pitch) GetLeadPlayerIdx() int { return p.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (p *Pitch) SetLeadPlayerIdx(idx int) { p.leadPlayerIdx = idx }

// GetBidPlayerIdx 現在ビッド中のプレイヤーインデックス取得
func (p *Pitch) GetBidPlayerIdx() int { return p.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (p *Pitch) SetBidPlayerIdx(idx int) { p.bidPlayerIdx = idx }

// GetCurrentBid 現在の最高ビッド値取得
func (p *Pitch) GetCurrentBid() int { return p.currentBid }

// SetCurrentBid 現在の最高ビッド値設定 (テスト用)
func (p *Pitch) SetCurrentBid(bid int) { p.currentBid = bid }

// GetBidWinnerIdx 最高ビッダーのインデックス取得 (-1=未確定)
func (p *Pitch) GetBidWinnerIdx() int { return p.bidWinnerIdx }

// SetBidWinnerIdx 最高ビッダーのインデックス設定 (テスト用)
func (p *Pitch) SetBidWinnerIdx(idx int) { p.bidWinnerIdx = idx }

// GetTrumpSuit 切り札スート取得 (PitchTrumpUnset=未確定)
func (p *Pitch) GetTrumpSuit() int { return p.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (p *Pitch) SetTrumpSuit(suit int) { p.trumpSuit = suit }

// GetGameEndFlag ゲーム終了フラグ取得
func (p *Pitch) GetGameEndFlag() bool { return p.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1=未確定)
func (p *Pitch) GetWinnerIdx() int { return p.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (p *Pitch) GetPlayerCnt() int { return len(p.players) }

// GetPlayer プレイヤー取得
func (p *Pitch) GetPlayer(i int) *PitchPlayer {
	return getPlayer(p.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (p *Pitch) IsHumanTurn() bool {
	return isHumanTurn(p.players, p.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (p *Pitch) IsHumanBidTurn() bool {
	return isHumanTurn(p.players, p.bidPlayerIdx)
}

// GetConfig 設定取得
func (p *Pitch) GetConfig() PitchConfig { return p.config }

// SetConfig 設定変更
func (p *Pitch) SetConfig(cfg PitchConfig) { p.config = cfg }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (p *Pitch) GetValidPlayIndices(playerIdx int) []int {
	return p.getValidPlayIndices(playerIdx)
}

// --- private helpers ---

func (p *Pitch) sortAllHands() {
	for _, pl := range p.players {
		pitchSortHand(pl)
	}
}

func pitchSortHand(pl *PitchPlayer) {
	sortPlayerHand(pl, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return pitchRankValue(ci.GetValue()) < pitchRankValue(cj.GetValue())
	})
}

func (p *Pitch) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(p.players[playerIdx], func(c *Card) bool { return p.validatePlay(playerIdx, c) == nil })
}

// --- CPU AI ---

// cpuSelectBid CPUがビッドを選択する
func (p *Pitch) cpuSelectBid(playerIdx int) int {
	switch p.config.CpuDifficulty {
	case PitchCpuDifficultyHard:
		return p.cpuBidHard(playerIdx)
	case PitchCpuDifficultyNormal:
		return p.cpuBidNormal(playerIdx)
	default:
		return p.cpuBidEasy(playerIdx)
	}
}

func (p *Pitch) cpuBidEasy(playerIdx int) int {
	// 50% でランダムに 0/2/3 を選択
	choices := []int{PitchPassBid, PitchMinBid, PitchMinBid + 1}
	bid := choices[rand.Intn(len(choices))]
	if bid > 0 && bid <= p.currentBid {
		return PitchPassBid
	}
	return p.coercePassWhenForbidden(playerIdx, bid)
}

// cpuBidNormal 期待トリック数からビッドを推定
func (p *Pitch) cpuBidNormal(playerIdx int) int {
	bid := p.estimateBidFromHand(playerIdx, false)
	return p.bidWithRules(playerIdx, bid)
}

// cpuBidHard cpuBidNormal を厳しめに評価
func (p *Pitch) cpuBidHard(playerIdx int) int {
	bid := p.estimateBidFromHand(playerIdx, true)
	return p.bidWithRules(playerIdx, bid)
}

// estimateBidFromHand 手札から最大ビッドを推定する
// strict=true の場合は控えめに評価する
func (p *Pitch) estimateBidFromHand(playerIdx int, strict bool) int {
	pl := p.players[playerIdx]
	bestBid := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		ranks := []int{}
		hasJ := false
		for i := 0; i < pl.GetCardsSize(); i++ {
			c := pl.GetCard(i)
			if c.GetDesign() != suit {
				continue
			}
			ranks = append(ranks, pitchRankValue(c.GetValue()))
			if c.GetValue() == 11 {
				hasJ = true
			}
		}
		if len(ranks) == 0 {
			continue
		}
		// 4ポイント (High/Low/Jack/Game) を見積もる
		points := 0
		// High: Aがあれば確実、Kでも長ければほぼ確実
		hasA := false
		hasK := false
		for _, r := range ranks {
			if r == 14 {
				hasA = true
			}
			if r == 13 {
				hasK = true
			}
		}
		if hasA {
			points++
		} else if hasK && len(ranks) >= 3 {
			points++
		}
		// Low: 2-3 の低いトランプを取る側になる確率は手の長さに依存
		hasLow := false
		for _, r := range ranks {
			if r <= 5 {
				hasLow = true
				break
			}
		}
		if hasLow && len(ranks) >= 3 {
			points++
		}
		// Jack: J of trump を持っていれば取りやすい
		if hasJ && len(ranks) >= 2 {
			points++
		}
		// Game: 高札 + 10 を多く保持
		gameStrong := 0
		for _, r := range ranks {
			if r >= 10 {
				gameStrong++
			}
		}
		if gameStrong >= 2 {
			points++
		}
		if strict && points > 0 {
			points-- // 控えめに 1 引く
		}
		if points < 0 {
			points = 0
		}
		if points > PitchMaxBid {
			points = PitchMaxBid
		}
		if points > bestBid {
			bestBid = points
		}
	}
	return bestBid
}

// bidWithRules 推定ビッドにビッドルール (>currentBid) を適用
func (p *Pitch) bidWithRules(playerIdx, suggested int) int {
	if suggested < PitchMinBid {
		return p.coercePassWhenForbidden(playerIdx, PitchPassBid)
	}
	if suggested <= p.currentBid {
		return p.coercePassWhenForbidden(playerIdx, PitchPassBid)
	}
	if suggested > PitchMaxBid {
		suggested = PitchMaxBid
	}
	return suggested
}

// coercePassWhenForbidden 親で全員パスの場合に強制 stuck (PitchMinBid) を返す
func (p *Pitch) coercePassWhenForbidden(playerIdx, bid int) int {
	if bid == PitchPassBid && playerIdx == p.dealerIdx && p.currentBid == 0 {
		return PitchMinBid
	}
	return bid
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (p *Pitch) cpuSelectPlayCard(playerIdx int) int {
	validIndices := p.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}
	switch p.config.CpuDifficulty {
	case PitchCpuDifficultyHard:
		return p.cpuPlayHard(playerIdx, validIndices)
	case PitchCpuDifficultyNormal:
		return p.cpuPlayNormal(playerIdx, validIndices)
	default:
		return validIndices[rand.Intn(len(validIndices))]
	}
}

// cpuPlayNormal: トリックを取りに行きたい場合は最強カード, それ以外は最弱
func (p *Pitch) cpuPlayNormal(playerIdx int, validIndices []int) int {
	pl := p.players[playerIdx]
	wantWin := p.cpuWantsWin(playerIdx)
	if len(p.currentTrick) == 0 {
		// リード: 取りに行く場合 A や高トランプ, それ以外は中位
		if wantWin {
			return pickMaxCard(pl, validIndices, p.trumpSuit, false)
		}
		return pickLowestCard(pl, validIndices)
	}
	// フォロー
	leadSuit := p.currentTrick[0].Card.GetDesign()
	bestInTrickRank, trumpInTrick := p.bestRankInTrick()
	if wantWin {
		idx := pickWinningCard(pl, validIndices, leadSuit, p.trumpSuit, bestInTrickRank, trumpInTrick)
		if idx >= 0 {
			return idx
		}
	}
	return pickLowestNonTrumpFollow(pl, validIndices, leadSuit, p.trumpSuit)
}

// cpuPlayHard: cpuPlayNormal とほぼ同じだが Low トランプを温存し J of trump を最後まで残す
func (p *Pitch) cpuPlayHard(playerIdx int, validIndices []int) int {
	pl := p.players[playerIdx]
	wantWin := p.cpuWantsWin(playerIdx)
	if len(p.currentTrick) == 0 {
		if wantWin {
			// 強いトランプでリード (J of trump 以外で最大)
			return pickMaxCard(pl, validIndices, p.trumpSuit, true)
		}
		return pickLowestCard(pl, validIndices)
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	bestInTrickRank, trumpInTrick := p.bestRankInTrick()
	if wantWin {
		idx := pickWinningCard(pl, validIndices, leadSuit, p.trumpSuit, bestInTrickRank, trumpInTrick)
		if idx >= 0 {
			return idx
		}
	}
	return pickLowestNonTrumpFollow(pl, validIndices, leadSuit, p.trumpSuit)
}

// cpuWantsWin このプレイヤーがトリックを取りに行くべきか
func (p *Pitch) cpuWantsWin(playerIdx int) bool {
	if playerIdx == p.bidWinnerIdx {
		return true // ビッダーは常に取りに行く
	}
	// non-bidder はビッダーを set back させたい / Game 用に高札を取りたい
	return p.players[playerIdx].GetTrickCount() < 2
}

// bestRankInTrick 現在のトリックで最強の (rank, isTrump) を返す
func (p *Pitch) bestRankInTrick() (int, bool) {
	leadSuit := p.currentTrick[0].Card.GetDesign()
	bestRank := -1
	trumpInTrick := false
	for _, tc := range p.currentTrick {
		if p.trumpSuit != PitchTrumpUnset && tc.Card.GetDesign() == p.trumpSuit {
			trumpInTrick = true
			r := pitchRankValue(tc.Card.GetValue())
			if r > bestRank {
				bestRank = r
			}
		} else if !trumpInTrick && tc.Card.GetDesign() == leadSuit {
			r := pitchRankValue(tc.Card.GetValue())
			if r > bestRank {
				bestRank = r
			}
		}
	}
	return bestRank, trumpInTrick
}

func pickMaxCard(pl *PitchPlayer, validIndices []int, trumpSuit int, avoidJackOfTrump bool) int {
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		c := pl.GetCard(idx)
		score := pitchRankValue(c.GetValue())
		if c.GetDesign() == trumpSuit {
			score += 100
		}
		if avoidJackOfTrump && c.GetDesign() == trumpSuit && c.GetValue() == 11 {
			score -= 50
		}
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

func pickLowestCard(pl *PitchPlayer, validIndices []int) int {
	bestIdx := validIndices[0]
	bestScore := 1 << 30
	for _, idx := range validIndices {
		c := pl.GetCard(idx)
		score := pitchRankValue(c.GetValue())
		if score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// pickWinningCard トリックを取れる最小のカードを返す。取れない場合は -1
func pickWinningCard(pl *PitchPlayer, validIndices []int, leadSuit, trumpSuit, bestInTrickRank int, trumpInTrick bool) int {
	bestIdx := -1
	bestScore := 1 << 30
	for _, idx := range validIndices {
		c := pl.GetCard(idx)
		isTrump := trumpSuit != PitchTrumpUnset && c.GetDesign() == trumpSuit
		// このカードが現在のリーダーを上回るか?
		canWin := false
		if isTrump {
			if !trumpInTrick {
				canWin = true
			} else if pitchRankValue(c.GetValue()) > bestInTrickRank {
				canWin = true
			}
		} else if !trumpInTrick && c.GetDesign() == leadSuit {
			if pitchRankValue(c.GetValue()) > bestInTrickRank {
				canWin = true
			}
		}
		if !canWin {
			continue
		}
		score := pitchRankValue(c.GetValue())
		if isTrump {
			score += 100 // トランプは確実に取れるが温存可
		}
		if score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// pickLowestNonTrumpFollow フォロー時の最弱カード (トランプは温存)
func pickLowestNonTrumpFollow(pl *PitchPlayer, validIndices []int, leadSuit, trumpSuit int) int {
	bestIdx := validIndices[0]
	bestScore := 1 << 30
	for _, idx := range validIndices {
		c := pl.GetCard(idx)
		score := pitchRankValue(c.GetValue())
		if c.GetDesign() == trumpSuit {
			score += 100 // トランプ温存
		}
		// Game ポイントになる高札 (10/A) をフォロー時は捨てたくない
		if c.GetValue() == 10 || c.GetValue() == 1 {
			score += 5
		}
		if score < bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// GetHint ヒントを取得する
func (p *Pitch) GetHint() *PitchHint {
	humanIdx := findHumanIdx(p.players)
	if humanIdx < 0 {
		return nil
	}
	if p.phase == PitchPhaseBid && p.bidPlayerIdx == humanIdx {
		bid := p.cpuBidHard(humanIdx)
		// dealer stuck チェック
		bid = p.coercePassWhenForbidden(humanIdx, bid)
		if bid > 0 && bid <= p.currentBid {
			bid = PitchPassBid
		}
		bid = p.coercePassWhenForbidden(humanIdx, bid)
		reason := "bid_pass"
		if bid >= PitchMinBid {
			reason = "bid_strong"
		}
		return &PitchHint{Bid: &bid, Reason: reason}
	}
	if p.phase == PitchPhasePlay && p.currentPlayerIdx == humanIdx {
		validIndices := p.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := p.cpuPlayHard(humanIdx, validIndices)
		return &PitchHint{CardIndex: &idx, Reason: p.playHintReason(humanIdx, idx)}
	}
	return nil
}

func (p *Pitch) playHintReason(playerIdx, chosenIdx int) string {
	pl := p.players[playerIdx]
	card := pl.GetCard(chosenIdx)
	if len(p.currentTrick) == 0 {
		if playerIdx == p.bidWinnerIdx && p.trumpSuit == PitchTrumpUnset {
			return "set_trump_lead"
		}
		return "lead_strong"
	}
	leadSuit := p.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == p.trumpSuit {
		return "trump_cut"
	}
	return "discard_low"
}

// pitchJSON is the JSON wire format for Pitch.
type pitchJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*PitchPlayer    `json:"ps"`
	Config           PitchConfig       `json:"cf"`
	Phase            PitchPhase        `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	DealerIdx        int               `json:"di"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	CurrentBid       int               `json:"cb"`
	BidWinnerIdx     int               `json:"bw"`
	TrumpSuit        int               `json:"ts"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (p *Pitch) MarshalJSON() ([]byte, error) {
	return json.Marshal(pitchJSON{
		TrumpCards:       p.trumpCards,
		Players:          p.players,
		Config:           p.config,
		Phase:            p.phase,
		RoundNumber:      p.roundNumber,
		TrickNumber:      p.trickNumber,
		DealerIdx:        p.dealerIdx,
		CurrentPlayerIdx: p.currentPlayerIdx,
		CurrentTrick:     p.currentTrick,
		LeadPlayerIdx:    p.leadPlayerIdx,
		BidPlayerIdx:     p.bidPlayerIdx,
		CurrentBid:       p.currentBid,
		BidWinnerIdx:     p.bidWinnerIdx,
		TrumpSuit:        p.trumpSuit,
		GameEndFlag:      p.gameEndFlag,
		WinnerIdx:        p.winnerIdx,
		ActionLog:        p.actionLog,
	})
}

// pitchMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const pitchMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (p *Pitch) UnmarshalJSON(data []byte) error {
	var j pitchJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > pitchMaxSliceLen || len(j.CurrentTrick) > pitchMaxSliceLen ||
		len(j.ActionLog) > pitchMaxSliceLen {
		return fmt.Errorf("pitch: input array exceeds maximum allowed size")
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.players = j.Players
	if p.players == nil {
		p.players = make([]*PitchPlayer, 0)
	}
	p.config = j.Config
	p.phase = j.Phase
	p.roundNumber = j.RoundNumber
	p.trickNumber = j.TrickNumber
	p.dealerIdx = j.DealerIdx
	p.currentPlayerIdx = j.CurrentPlayerIdx
	p.currentTrick = j.CurrentTrick
	if p.currentTrick == nil {
		p.currentTrick = make([]*TrickCard, 0)
	}
	p.leadPlayerIdx = j.LeadPlayerIdx
	p.bidPlayerIdx = j.BidPlayerIdx
	p.currentBid = j.CurrentBid
	p.bidWinnerIdx = j.BidWinnerIdx
	p.trumpSuit = j.TrumpSuit
	p.gameEndFlag = j.GameEndFlag
	p.winnerIdx = j.WinnerIdx
	p.actionLog = j.ActionLog
	if p.actionLog == nil {
		p.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
