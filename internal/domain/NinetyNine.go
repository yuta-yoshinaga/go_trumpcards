package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// NinetyNinePlayerCnt ナインティナインプレイヤー数 (1 human + 2 CPU)
const NinetyNinePlayerCnt = 3

// NinetyNineHandSize ディール開始時の手札枚数 (12配り、3枚伏せ→9枚でプレイ)
const NinetyNineDealtSize = 12

// NinetyNineBurySize 伏せる枚数 (= ビッド宣言に使う)
const NinetyNineBurySize = 3

// NinetyNineTricksPerDeal 1ディールのトリック数 (12 - 3 = 9)
const NinetyNineTricksPerDeal = NinetyNineDealtSize - NinetyNineBurySize

// NinetyNinePhase ゲームフェーズ
type NinetyNinePhase int

// NinetyNineのフェーズ定数
const (
	// NinetyNinePhaseBid ビッド(3枚伏せ)フェーズ
	NinetyNinePhaseBid NinetyNinePhase = 0
	// NinetyNinePhasePlay トリックプレイフェーズ
	NinetyNinePhasePlay NinetyNinePhase = 1
	// NinetyNinePhaseTrickEnd トリック終了フェーズ
	NinetyNinePhaseTrickEnd NinetyNinePhase = 2
	// NinetyNinePhaseRoundEnd ディール終了フェーズ
	NinetyNinePhaseRoundEnd NinetyNinePhase = 3
	// NinetyNinePhaseGameEnd ゲーム終了フェーズ
	NinetyNinePhaseGameEnd NinetyNinePhase = 4
)

// NinetyNineHint ヒント情報
type NinetyNineHint struct {
	CardIndex   *int   // プレイ時の推奨カードインデックス (ビッド時nil)
	BuryIndices []int  // ビッド時の伏せる3枚の推奨インデックス (プレイ時nil)
	Reason      string // ヒント理由キー
}

// ninetyNineTrumpRotation はディール番号で巡回する切り札スート。
// 各ディールで切り札を固定し、ディールごとに ♠→♥→♣→♦ と巡回させる
// (David Parlett のバリエーションのうち「ディールごとに切り札が決まる」方式を採用)。
var ninetyNineTrumpRotation = []int{
	CardDesignSpade,
	CardDesignHeart,
	CardDesignClover,
	CardDesignDiamond,
}

// ninetyNineSuitBidValue は伏せ札のスートをビッド値へマップする。
// ♦=0, ♠=1, ♥=2, ♣=3。3枚の合計が宣言トリック数 (0-9) になる。
func ninetyNineSuitBidValue(design int) int {
	switch design {
	case CardDesignDiamond:
		return 0
	case CardDesignSpade:
		return 1
	case CardDesignHeart:
		return 2
	case CardDesignClover:
		return 3
	default:
		return 0
	}
}

// NinetyNine ナインティナインゲームクラス
type NinetyNine struct {
	trumpCards       *TrumpCards
	players          []*NinetyNinePlayer
	config           NinetyNineConfig
	phase            NinetyNinePhase
	dealNumber       int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	bidPlayerIdx     int
	dealerIdx        int
	trumpSuit        int // 切り札スート (CardDesignSpade..Diamond)
	gameEndFlag      bool
	winnerIdx        int
	actionLog        []*ActionLogEntry
}

// NewNinetyNine コンストラクタ
func NewNinetyNine(trumpCards *TrumpCards, players []*NinetyNinePlayer, config NinetyNineConfig) *NinetyNine {
	return &NinetyNine{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
		trumpSuit:  CardDesignSpade,
	}
}

// NewDefaultNinetyNine returns NinetyNine with the standard 3-player setup
// (1 human, 2 CPU) and DefaultNinetyNineConfig.
func NewDefaultNinetyNine() *NinetyNine {
	players := []*NinetyNinePlayer{
		NewNinetyNinePlayer(true),
		NewNinetyNinePlayer(false),
		NewNinetyNinePlayer(false),
	}
	return NewNinetyNine(NewTrumpCardsNinetyNine(), players, DefaultNinetyNineConfig())
}

// Reset ゲーム初期化
func (o *NinetyNine) Reset() {
	o.gameEndFlag = false
	o.winnerIdx = -1
	o.dealerIdx = 0
	o.dealNumber = 1
	o.trickNumber = 0
	o.currentTrick = nil
	o.leadPlayerIdx = -1
	o.currentPlayerIdx = -1
	o.actionLog = nil

	for _, p := range o.players {
		p.SetCumulativeScore(0)
		p.ResetRound()
	}

	o.deal()
	o.phase = NinetyNinePhaseBid
	o.bidPlayerIdx = (o.dealerIdx + 1) % NinetyNinePlayerCnt
}

// NextRound 次のディールを開始する
func (o *NinetyNine) NextRound() {
	if o.phase != NinetyNinePhaseRoundEnd {
		return
	}

	o.dealNumber++
	o.dealerIdx = (o.dealerIdx + 1) % NinetyNinePlayerCnt
	o.trickNumber = 0
	o.currentTrick = nil
	o.leadPlayerIdx = -1
	o.currentPlayerIdx = -1

	for _, p := range o.players {
		p.ResetRound()
	}

	o.deal()
	o.phase = NinetyNinePhaseBid
	o.bidPlayerIdx = (o.dealerIdx + 1) % NinetyNinePlayerCnt
}

// PlayerBid 人間プレイヤーが3枚を伏せてビッドする
func (o *NinetyNine) PlayerBid(buryIndices []int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.phase != NinetyNinePhaseBid {
		return ErrWrongPhase
	}

	humanIdx := o.findHumanIdx()
	if humanIdx < 0 || o.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if !o.players[humanIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	idxs, err := o.validateBuryIndices(humanIdx, buryIndices)
	if err != nil {
		return err
	}

	o.applyBury(humanIdx, idxs)
	o.advanceBid()
	return nil
}

// CpuBid 現在のビッドプレイヤーがCPUの場合に3枚伏せてビッドする
func (o *NinetyNine) CpuBid() {
	if o.gameEndFlag || o.phase != NinetyNinePhaseBid {
		return
	}
	if o.bidPlayerIdx < 0 || o.bidPlayerIdx >= NinetyNinePlayerCnt {
		return
	}
	if o.players[o.bidPlayerIdx].GetIsHuman() {
		return
	}

	idxs := o.cpuSelectBury(o.bidPlayerIdx)
	o.applyBury(o.bidPlayerIdx, idxs)
	o.advanceBid()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (o *NinetyNine) PlayerPlay(cardIndex int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.phase != NinetyNinePhasePlay {
		return ErrWrongPhase
	}
	if !o.players[o.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := o.players[o.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := o.validatePlay(o.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	o.playCard(o.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (o *NinetyNine) CpuPlay() {
	if o.gameEndFlag || o.phase != NinetyNinePhasePlay {
		return
	}
	if o.players[o.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := o.players[o.currentPlayerIdx]
	cardIdx := o.cpuSelectPlayCard(o.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	o.playCard(o.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (o *NinetyNine) ResolveTrick() {
	if o.phase != NinetyNinePhaseTrickEnd || len(o.currentTrick) != NinetyNinePlayerCnt {
		return
	}

	winnerIdx := o.trickWinner()
	trickCards := make([]*Card, len(o.currentTrick))
	for i, tc := range o.currentTrick {
		trickCards[i] = tc.Card
	}

	o.players[winnerIdx].AddTrick(trickCards)

	winnerName := o.playerName(winnerIdx)
	o.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, o.trickNumber), trickCards)

	o.leadPlayerIdx = winnerIdx

	if o.trickNumber >= NinetyNineTricksPerDeal {
		o.phase = NinetyNinePhaseRoundEnd
	} else {
		o.phase = NinetyNinePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (o *NinetyNine) NextTrick() {
	if o.phase != NinetyNinePhaseTrickEnd {
		return
	}
	o.currentTrick = nil
	o.currentPlayerIdx = o.leadPlayerIdx
	o.trickNumber++
	o.phase = NinetyNinePhasePlay
}

// ScoreRound ディールのスコアを確定し、ゲーム終了判定を行う。
//
// 得点ルール (David Parlett の Ninety-Nine をモデル化):
//   - 宣言トリック数とぴったり一致したプレイヤーは「成功」とみなす。
//   - 成功者には基礎点 (10 + 宣言数) を与える。
//   - さらに、そのディールで成功した人数に応じてボーナスを加算する:
//     1人だけ成功 → +30、2人成功 → +20 ずつ、3人(全員)成功 → +10 ずつ。
//     成功者が少ないほどボーナスが大きい (希少な達成を高く評価する)。
//   - 不一致のプレイヤーは0点 (マイナスはしない)。
//
// 累積スコアが TargetScore 以上に達したらゲーム終了。
func (o *NinetyNine) ScoreRound() {
	if o.phase != NinetyNinePhaseRoundEnd {
		return
	}

	successCount := 0
	for i := range NinetyNinePlayerCnt {
		if o.players[i].GetTrickCount() == o.players[i].GetBid() {
			successCount++
		}
	}

	bonus := o.successBonus(successCount)

	for i := range NinetyNinePlayerCnt {
		p := o.players[i]
		tricks := p.GetTrickCount()
		bid := p.GetBid()
		if tricks == bid {
			p.SetRoundScore(10 + bid + bonus)
			o.appendLog(i, "bid_success", fmt.Sprintf("%s declared %d, took %d: +%d (bonus +%d)",
				o.playerName(i), bid, tricks, p.GetRoundScore(), bonus), nil)
		} else {
			p.SetRoundScore(0)
			o.appendLog(i, "bid_fail", fmt.Sprintf("%s declared %d, took %d: 0",
				o.playerName(i), bid, tricks), nil)
		}
	}

	for i := range NinetyNinePlayerCnt {
		o.players[i].CommitRoundScore()
	}

	for i := range NinetyNinePlayerCnt {
		o.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			o.playerName(i), o.players[i].GetCumulativeScore()), nil)
	}

	// ゲーム終了判定: いずれかが TargetScore 以上に達したら終了
	reached := false
	for i := range NinetyNinePlayerCnt {
		if o.players[i].GetCumulativeScore() >= o.config.TargetScore {
			reached = true
			break
		}
	}
	if reached {
		o.gameEndFlag = true
		o.phase = NinetyNinePhaseGameEnd
		o.determineWinner()
	}
}

// successBonus 成功人数に応じたボーナス点を返す
func (o *NinetyNine) successBonus(successCount int) int {
	switch successCount {
	case 1:
		return 30
	case 2:
		return 20
	case 3:
		return 10
	default:
		return 0
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (o *NinetyNine) GetPhase() NinetyNinePhase { return o.phase }

// SetPhase フェーズ設定 (テスト用)
func (o *NinetyNine) SetPhase(phase NinetyNinePhase) { o.phase = phase }

// GetDealNumber 現在のディール番号取得
func (o *NinetyNine) GetDealNumber() int { return o.dealNumber }

// SetDealNumber ディール番号設定 (テスト用)
func (o *NinetyNine) SetDealNumber(n int) { o.dealNumber = n }

// GetHandSize プレイ時の手札枚数 (= 9) を返す
func (o *NinetyNine) GetHandSize() int { return NinetyNineTricksPerDeal }

// GetTrickNumber 現在のトリック番号取得
func (o *NinetyNine) GetTrickNumber() int { return o.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (o *NinetyNine) SetTrickNumber(n int) { o.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (o *NinetyNine) GetCurrentPlayerIdx() int { return o.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (o *NinetyNine) SetCurrentPlayerIdx(idx int) { o.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (o *NinetyNine) GetCurrentTrick() []*TrickCard { return o.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (o *NinetyNine) SetCurrentTrick(trick []*TrickCard) { o.currentTrick = trick }

// GetTrumpSuit 切り札スート取得
func (o *NinetyNine) GetTrumpSuit() int { return o.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (o *NinetyNine) SetTrumpSuit(suit int) { o.trumpSuit = suit }

// GetDealerIdx ディーラーインデックス取得
func (o *NinetyNine) GetDealerIdx() int { return o.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (o *NinetyNine) SetDealerIdx(idx int) { o.dealerIdx = idx }

// GetGameEndFlag ゲーム終了フラグ取得
func (o *NinetyNine) GetGameEndFlag() bool { return o.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (o *NinetyNine) GetWinnerIdx() int { return o.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (o *NinetyNine) GetPlayerCnt() int { return len(o.players) }

// GetPlayer プレイヤー取得
func (o *NinetyNine) GetPlayer(i int) *NinetyNinePlayer {
	if i < 0 || i >= len(o.players) {
		return nil
	}
	return o.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (o *NinetyNine) GetLeadPlayerIdx() int { return o.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (o *NinetyNine) SetLeadPlayerIdx(idx int) { o.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (o *NinetyNine) GetBidPlayerIdx() int { return o.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (o *NinetyNine) SetBidPlayerIdx(idx int) { o.bidPlayerIdx = idx }

// GetTargetScore ゲーム勝利に必要な累積スコアを返す
func (o *NinetyNine) GetTargetScore() int { return o.config.TargetScore }

// IsHumanTurn 現在の手番が人間かどうか
func (o *NinetyNine) IsHumanTurn() bool {
	if o.currentPlayerIdx < 0 || o.currentPlayerIdx >= len(o.players) {
		return false
	}
	return o.players[o.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (o *NinetyNine) IsHumanBidTurn() bool {
	if o.bidPlayerIdx < 0 || o.bidPlayerIdx >= len(o.players) {
		return false
	}
	return o.players[o.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (o *NinetyNine) GetConfig() NinetyNineConfig { return o.config }

// SetConfig 設定変更
func (o *NinetyNine) SetConfig(cfg NinetyNineConfig) { o.config = cfg }

// GetActionLog 棋譜取得
func (o *NinetyNine) GetActionLog() []*ActionLogEntry { return o.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (o *NinetyNine) GetValidPlayIndices(playerIdx int) []int {
	return o.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (o *NinetyNine) findHumanIdx() int {
	for i, p := range o.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// deal カードを配り、切り札を決める
func (o *NinetyNine) deal() {
	o.trumpCards = NewTrumpCardsNinetyNine()
	o.trumpCards.Shuffle()

	for range NinetyNineDealtSize {
		for j := range NinetyNinePlayerCnt {
			idx := (o.dealerIdx + 1 + j) % NinetyNinePlayerCnt
			card := o.trumpCards.DrawCard()
			if card != nil {
				o.players[idx].AddCard(card)
			}
		}
	}

	// 切り札はディール番号で巡回する固定スート (♠→♥→♣→♦)
	o.trumpSuit = ninetyNineTrumpRotation[(o.dealNumber-1)%len(ninetyNineTrumpRotation)]

	o.sortAllHands()
}

// validateBuryIndices 伏せる3枚のインデックスを検証し、昇順スライスを返す
func (o *NinetyNine) validateBuryIndices(playerIdx int, indices []int) ([]int, error) {
	if len(indices) != NinetyNineBurySize {
		return nil, NewDomainError(ErrInvalidPlay, fmt.Sprintf("伏せるカードは%d枚指定してください", NinetyNineBurySize))
	}
	player := o.players[playerIdx]
	seen := make(map[int]bool, NinetyNineBurySize)
	for _, idx := range indices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return nil, NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return nil, NewDomainError(ErrInvalidPlay, "同じカードを重複して指定できません")
		}
		seen[idx] = true
	}
	sorted := append([]int(nil), indices...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	return sorted, nil
}

// applyBury 指定した3枚を手札から取り除いて伏せ、ビッド値を確定する。
// indices は降順であること (RemoveCard でインデックスがずれないようにするため)。
func (o *NinetyNine) applyBury(playerIdx int, descIndices []int) {
	player := o.players[playerIdx]
	buried := make([]*Card, 0, NinetyNineBurySize)
	bid := 0
	for _, idx := range descIndices {
		card := player.RemoveCard(idx)
		if card == nil {
			continue
		}
		buried = append(buried, card)
		bid += ninetyNineSuitBidValue(card.GetDesign())
	}
	player.SetBuried(buried)
	player.SetBid(bid)
	o.appendLog(playerIdx, "bid", fmt.Sprintf("%s buries 3 and declares %d", o.playerName(playerIdx), bid), buried)
}

// advanceBid ビッドプレイヤーを次に進める
func (o *NinetyNine) advanceBid() {
	bidCount := 0
	for _, p := range o.players {
		if p.GetBid() >= 0 {
			bidCount++
		}
	}
	if bidCount >= NinetyNinePlayerCnt {
		o.phase = NinetyNinePhasePlay
		o.startPlayPhase()
	} else {
		o.bidPlayerIdx = (o.bidPlayerIdx + 1) % NinetyNinePlayerCnt
	}
}

// startPlayPhase プレイフェーズ開始: ディーラーの左がリード
func (o *NinetyNine) startPlayPhase() {
	o.trickNumber = 1
	o.currentTrick = nil
	o.leadPlayerIdx = (o.dealerIdx + 1) % NinetyNinePlayerCnt
	o.currentPlayerIdx = o.leadPlayerIdx
}

// playCard カードをプレイする共通処理
func (o *NinetyNine) playCard(playerIdx int, card *Card) {
	o.currentTrick = append(o.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	o.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", o.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(o.currentTrick) == NinetyNinePlayerCnt {
		o.phase = NinetyNinePhaseTrickEnd
	} else {
		o.currentPlayerIdx = (o.currentPlayerIdx + 1) % NinetyNinePlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する (must-follow)
func (o *NinetyNine) validatePlay(playerIdx int, card *Card) error {
	if len(o.currentTrick) == 0 {
		return nil
	}
	leadSuit := o.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && o.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (o *NinetyNine) playerHasSuit(playerIdx int, design int) bool {
	p := o.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する
func (o *NinetyNine) trickWinner() int {
	return ResolveTrickWinner(o.currentTrick, o.trumpSuit, ninetyNineRankValue)
}

// ninetyNineRankValue はトリックの強さ比較に使う値を返す。
// Aceは最強 (=14) として扱い、それ以外は素の値を用いる。
func ninetyNineRankValue(card *Card) int {
	v := card.GetValue()
	if v == 1 {
		return 14
	}
	return v
}

// determineWinner 最終的な勝者を決定する。
// 累積スコア最大が勝者。同点の場合はそのディールのラウンドスコアが高い方、
// それも同点なら座席番号 (インデックス) が小さい方を勝者とする (決定論的タイブレーク)。
func (o *NinetyNine) determineWinner() {
	best := 0
	for i := 1; i < NinetyNinePlayerCnt; i++ {
		if o.ninetyNineBeats(i, best) {
			best = i
		}
	}
	o.winnerIdx = best
	o.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", o.playerName(o.winnerIdx)), nil)
}

// ninetyNineBeats プレイヤー a がプレイヤー b に勝るか (タイブレーク込み)
func (o *NinetyNine) ninetyNineBeats(a, b int) bool {
	pa, pb := o.players[a], o.players[b]
	if pa.GetCumulativeScore() != pb.GetCumulativeScore() {
		return pa.GetCumulativeScore() > pb.GetCumulativeScore()
	}
	if pa.GetRoundScore() != pb.GetRoundScore() {
		return pa.GetRoundScore() > pb.GetRoundScore()
	}
	return a < b
}

// sortAllHands 全プレイヤーの手札をソートする
func (o *NinetyNine) sortAllHands() {
	for _, p := range o.players {
		ninetyNineSortHand(p)
	}
}

// ninetyNineSortHand プレイヤーの手札をスート→値の順にソートする
func ninetyNineSortHand(p *NinetyNinePlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ninetyNineRankValue(ci) < ninetyNineRankValue(cj)
	})
}

// playerName プレイヤー名を返す
func (o *NinetyNine) playerName(idx int) string {
	if idx < 0 || idx >= len(o.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if o.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (o *NinetyNine) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	o.actionLog = append(o.actionLog, &ActionLogEntry{
		TurnNumber: len(o.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (o *NinetyNine) getValidPlayIndices(playerIdx int) []int {
	player := o.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return o.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// GetHint ヒントを取得する
func (o *NinetyNine) GetHint() *NinetyNineHint {
	humanIdx := o.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}
	if o.phase == NinetyNinePhaseBid && o.bidPlayerIdx == humanIdx {
		idxs := o.cpuSelectBury(humanIdx)
		// cpuSelectBury は降順なので表示用に昇順へ
		asc := append([]int(nil), idxs...)
		sort.Ints(asc)
		return &NinetyNineHint{BuryIndices: asc, Reason: "strategic_bury"}
	}
	if o.phase == NinetyNinePhasePlay && o.currentPlayerIdx == humanIdx {
		validIndices := o.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := o.cpuPlayHard(humanIdx, validIndices)
		return &NinetyNineHint{CardIndex: &idx, Reason: o.playHintReason(humanIdx, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (o *NinetyNine) playHintReason(playerIdx int, chosenIdx int) string {
	player := o.players[playerIdx]
	card := player.GetCard(chosenIdx)

	if len(o.currentTrick) == 0 {
		if player.GetTrickCount() < player.GetBid() {
			return "lead_strong"
		}
		return "lead_low"
	}
	leadSuit := o.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == o.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectBury CPUが伏せる3枚のインデックスを選ぶ (降順で返す)。
// 難易度に応じて目標ビッドを決め、その合計スートになる3枚を手札から探す。
func (o *NinetyNine) cpuSelectBury(playerIdx int) []int {
	player := o.players[playerIdx]
	target := o.cpuTargetBid(playerIdx)

	// 目標ビッドにスート合計を一致させる3枚の組み合わせを探す。
	// 見つからない場合はできるだけ近い合計の組み合わせを選ぶ。
	n := player.GetCardsSize()
	bestCombo := []int{0, 1, 2}
	bestDiff := 1 << 30
	for a := 0; a < n; a++ {
		for b := a + 1; b < n; b++ {
			for c := b + 1; c < n; c++ {
				sum := ninetyNineSuitBidValue(player.GetCard(a).GetDesign()) +
					ninetyNineSuitBidValue(player.GetCard(b).GetDesign()) +
					ninetyNineSuitBidValue(player.GetCard(c).GetDesign())
				diff := sum - target
				if diff < 0 {
					diff = -diff
				}
				if diff < bestDiff {
					bestDiff = diff
					bestCombo = []int{a, b, c}
					if bestDiff == 0 {
						sort.Sort(sort.Reverse(sort.IntSlice(bestCombo)))
						return bestCombo
					}
				}
			}
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(bestCombo)))
	return bestCombo
}

// cpuTargetBid CPUが狙うトリック数を見積もる
func (o *NinetyNine) cpuTargetBid(playerIdx int) int {
	switch o.config.CpuDifficulty {
	case NinetyNineCpuDifficultyEasy:
		return rand.Intn(NinetyNineTricksPerDeal + 1)
	default:
		return o.estimateTricks(playerIdx)
	}
}

// estimateTricks 手札の強さからトリック獲得数を見積もる
func (o *NinetyNine) estimateTricks(playerIdx int) int {
	player := o.players[playerIdx]
	bid := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetDesign() == o.trumpSuit {
			if ninetyNineRankValue(card) >= 11 {
				bid++
			}
		} else if card.GetValue() == 1 || card.GetValue() == 13 {
			bid++
		}
	}
	if bid > NinetyNineTricksPerDeal {
		bid = NinetyNineTricksPerDeal
	}
	return bid
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (o *NinetyNine) cpuSelectPlayCard(playerIdx int) int {
	validIndices := o.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch o.config.CpuDifficulty {
	case NinetyNineCpuDifficultyHard:
		return o.cpuPlayHard(playerIdx, validIndices)
	case NinetyNineCpuDifficultyNormal:
		return o.cpuPlayNormal(playerIdx, validIndices)
	default:
		return validIndices[rand.Intn(len(validIndices))]
	}
}

// cpuPlayNormal ビッドに近づくようにプレイ
func (o *NinetyNine) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := o.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(o.currentTrick) == 0 {
		if tricks < bid {
			return o.highestCardIdx(player, validIndices)
		}
		return o.lowestCardIdx(player, validIndices)
	}
	leadSuit := o.currentTrick[0].Card.GetDesign()
	if tricks < bid {
		return o.tryWinTrick(player, validIndices, leadSuit)
	}
	return o.tryLoseTrick(player, validIndices, leadSuit)
}

// cpuPlayHard 高度な戦略プレイ
func (o *NinetyNine) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := o.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	needMore := tricks < bid
	exactBid := tricks == bid

	if len(o.currentTrick) == 0 {
		if needMore {
			return o.bestLeadCard(player, validIndices, true)
		}
		return o.bestLeadCard(player, validIndices, false)
	}
	leadSuit := o.currentTrick[0].Card.GetDesign()
	if needMore {
		return o.tryWinTrick(player, validIndices, leadSuit)
	}
	if exactBid {
		return o.tryLoseTrick(player, validIndices, leadSuit)
	}
	return o.lowestCardIdx(player, validIndices)
}

// --- CPU helper methods ---

func (o *NinetyNine) highestCardIdx(player *NinetyNinePlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if ninetyNineRankValue(player.GetCard(idx)) > ninetyNineRankValue(player.GetCard(best)) {
			best = idx
		}
	}
	return best
}

func (o *NinetyNine) lowestCardIdx(player *NinetyNinePlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if ninetyNineRankValue(player.GetCard(idx)) < ninetyNineRankValue(player.GetCard(best)) {
			best = idx
		}
	}
	return best
}

func (o *NinetyNine) bestLeadCard(player *NinetyNinePlayer, indices []int, wantWin bool) int {
	if wantWin {
		best := indices[0]
		bestScore := -1
		for _, idx := range indices {
			card := player.GetCard(idx)
			score := ninetyNineRankValue(card)
			if card.GetDesign() == o.trumpSuit {
				score += 100
			}
			if score > bestScore {
				bestScore = score
				best = idx
			}
		}
		return best
	}
	best := indices[0]
	bestVal := ninetyNineRankValue(player.GetCard(indices[0]))
	bestIsTrump := player.GetCard(indices[0]).GetDesign() == o.trumpSuit
	for _, idx := range indices[1:] {
		card := player.GetCard(idx)
		isTrump := card.GetDesign() == o.trumpSuit
		val := ninetyNineRankValue(card)
		if bestIsTrump && !isTrump {
			best = idx
			bestVal = val
			bestIsTrump = false
		} else if isTrump == bestIsTrump && val < bestVal {
			best = idx
			bestVal = val
		}
	}
	return best
}

func (o *NinetyNine) tryWinTrick(player *NinetyNinePlayer, validIndices []int, leadSuit int) int {
	highestInTrick := 0
	highestTrumpInTrick := 0
	hasTrumpInTrick := false
	for _, tc := range o.currentTrick {
		val := ninetyNineRankValue(tc.Card)
		if tc.Card.GetDesign() == leadSuit && val > highestInTrick {
			highestInTrick = val
		}
		if tc.Card.GetDesign() == o.trumpSuit {
			hasTrumpInTrick = true
			if val > highestTrumpInTrick {
				highestTrumpInTrick = val
			}
		}
	}

	hasLead := false
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == leadSuit {
			hasLead = true
			break
		}
	}

	if hasLead && !hasTrumpInTrick {
		overCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && ninetyNineRankValue(card) > highestInTrick {
				overCards = append(overCards, idx)
			}
		}
		if len(overCards) > 0 {
			return o.lowestCardIdx(player, overCards)
		}
	}

	if !hasLead {
		trumpCards := []int{}
		for _, idx := range validIndices {
			if player.GetCard(idx).GetDesign() == o.trumpSuit {
				trumpCards = append(trumpCards, idx)
			}
		}
		if len(trumpCards) > 0 {
			if hasTrumpInTrick {
				winnable := []int{}
				for _, idx := range trumpCards {
					if ninetyNineRankValue(player.GetCard(idx)) > highestTrumpInTrick {
						winnable = append(winnable, idx)
					}
				}
				if len(winnable) > 0 {
					return o.lowestCardIdx(player, winnable)
				}
			} else {
				return o.lowestCardIdx(player, trumpCards)
			}
		}
	}

	return o.lowestCardIdx(player, validIndices)
}

func (o *NinetyNine) tryLoseTrick(player *NinetyNinePlayer, validIndices []int, leadSuit int) int {
	highestInTrick := 0
	for _, tc := range o.currentTrick {
		val := ninetyNineRankValue(tc.Card)
		if tc.Card.GetDesign() == leadSuit && val > highestInTrick {
			highestInTrick = val
		}
	}

	underCards := []int{}
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		if card.GetDesign() == leadSuit && ninetyNineRankValue(card) < highestInTrick {
			underCards = append(underCards, idx)
		}
	}
	if len(underCards) > 0 {
		return o.highestCardIdx(player, underCards)
	}
	return o.lowestCardIdx(player, validIndices)
}

// --- JSON serialization ---

// ninetyNineJSON is the JSON wire format for NinetyNine.
type ninetyNineJSON struct {
	TrumpCards       *TrumpCards         `json:"tc"`
	Players          []*NinetyNinePlayer `json:"ps"`
	Config           NinetyNineConfig    `json:"cf"`
	Phase            NinetyNinePhase     `json:"ph"`
	DealNumber       int                 `json:"dn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	LeadPlayerIdx    int                 `json:"li"`
	BidPlayerIdx     int                 `json:"bi"`
	DealerIdx        int                 `json:"di"`
	TrumpSuit        int                 `json:"ts"`
	GameEndFlag      bool                `json:"ge"`
	WinnerIdx        int                 `json:"wi"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (o *NinetyNine) MarshalJSON() ([]byte, error) {
	return json.Marshal(ninetyNineJSON{
		TrumpCards:       o.trumpCards,
		Players:          o.players,
		Config:           o.config,
		Phase:            o.phase,
		DealNumber:       o.dealNumber,
		TrickNumber:      o.trickNumber,
		CurrentPlayerIdx: o.currentPlayerIdx,
		CurrentTrick:     o.currentTrick,
		LeadPlayerIdx:    o.leadPlayerIdx,
		BidPlayerIdx:     o.bidPlayerIdx,
		DealerIdx:        o.dealerIdx,
		TrumpSuit:        o.trumpSuit,
		GameEndFlag:      o.gameEndFlag,
		WinnerIdx:        o.winnerIdx,
		ActionLog:        o.actionLog,
	})
}

// ninetyNineMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const ninetyNineMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler. It hardens against malformed input
// by validating slice sizes, config, phase range, player count, indices, bids,
// and the winner index.
func (o *NinetyNine) UnmarshalJSON(data []byte) error {
	var j ninetyNineJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ninetyNineMaxSliceLen || len(j.CurrentTrick) > ninetyNineMaxSliceLen ||
		len(j.ActionLog) > ninetyNineMaxSliceLen {
		return fmt.Errorf("ninetynine: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("ninetynine: invalid config: %w", err)
	}
	if j.Phase < NinetyNinePhaseBid || j.Phase > NinetyNinePhaseGameEnd {
		return fmt.Errorf("ninetynine: phase %d out of range", j.Phase)
	}
	pCnt := len(j.Players)
	if pCnt != NinetyNinePlayerCnt {
		return fmt.Errorf("ninetynine: player count %d must be %d", pCnt, NinetyNinePlayerCnt)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("ninetynine: player %d is nil", i)
		}
		if b := p.GetBid(); b < -1 || b > NinetyNineTricksPerDeal {
			return fmt.Errorf("ninetynine: player %d bid %d out of range", i, b)
		}
	}
	if j.CurrentPlayerIdx < -1 || j.CurrentPlayerIdx >= pCnt {
		return fmt.Errorf("ninetynine: currentPlayerIdx %d out of range", j.CurrentPlayerIdx)
	}
	if j.DealerIdx < 0 || j.DealerIdx >= pCnt {
		return fmt.Errorf("ninetynine: dealerIdx %d out of range", j.DealerIdx)
	}
	if j.BidPlayerIdx < -1 || j.BidPlayerIdx >= pCnt {
		return fmt.Errorf("ninetynine: bidPlayerIdx %d out of range", j.BidPlayerIdx)
	}
	if j.LeadPlayerIdx < -1 || j.LeadPlayerIdx >= pCnt {
		return fmt.Errorf("ninetynine: leadPlayerIdx %d out of range", j.LeadPlayerIdx)
	}
	if j.TrickNumber < 0 {
		return fmt.Errorf("ninetynine: trickNumber %d must be >= 0", j.TrickNumber)
	}
	if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("ninetynine: trumpSuit %d out of range", j.TrumpSuit)
	}
	for i, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("ninetynine: currentTrick entry %d is nil", i)
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= pCnt {
			return fmt.Errorf("ninetynine: currentTrick entry %d playerIdx out of range", i)
		}
	}
	if j.GameEndFlag {
		if j.WinnerIdx < 0 || j.WinnerIdx >= pCnt {
			return fmt.Errorf("ninetynine: winnerIdx %d out of range when game ended", j.WinnerIdx)
		}
	} else if j.WinnerIdx != -1 && (j.WinnerIdx < 0 || j.WinnerIdx >= pCnt) {
		return fmt.Errorf("ninetynine: winnerIdx %d out of range", j.WinnerIdx)
	}

	o.trumpCards = j.TrumpCards
	if o.trumpCards == nil {
		o.trumpCards = NewTrumpCardsNinetyNine()
	}
	o.players = j.Players
	o.config = j.Config
	o.phase = j.Phase
	o.dealNumber = j.DealNumber
	o.trickNumber = j.TrickNumber
	o.currentPlayerIdx = j.CurrentPlayerIdx
	o.currentTrick = j.CurrentTrick
	if o.currentTrick == nil {
		o.currentTrick = make([]*TrickCard, 0)
	}
	o.leadPlayerIdx = j.LeadPlayerIdx
	o.bidPlayerIdx = j.BidPlayerIdx
	o.dealerIdx = j.DealerIdx
	o.trumpSuit = j.TrumpSuit
	o.gameEndFlag = j.GameEndFlag
	o.winnerIdx = j.WinnerIdx
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
