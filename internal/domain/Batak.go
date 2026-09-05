package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// BatakPlayerCnt Batak のプレイヤー数
const BatakPlayerCnt = 4

// BatakHandSize 各プレイヤーの手札枚数
const BatakHandSize = 13

// BatakPhase ゲームフェーズ
type BatakPhase int

// Batak のフェーズ定数
const (
	// BatakPhaseBid ビッドフェーズ
	BatakPhaseBid BatakPhase = 0
	// BatakPhasePlay トリックプレイフェーズ
	BatakPhasePlay BatakPhase = 1
	// BatakPhaseTrickEnd トリック終了フェーズ
	BatakPhaseTrickEnd BatakPhase = 2
	// BatakPhaseRoundEnd ラウンド終了フェーズ
	BatakPhaseRoundEnd BatakPhase = 3
	// BatakPhaseGameEnd ゲーム終了フェーズ
	BatakPhaseGameEnd BatakPhase = 4
)

// BatakHint ヒント情報
type BatakHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド時 nil)
	Bid       *int   // 推奨ビッド値 (プレイ時 nil)
	Reason    string // ヒント理由キー
}

// Batak Batak ゲームクラス
//
// ルール上の差分メモ (トルコの入札制トリックテイキング Batak):
//   - 競り (auction) で 1 人だけが親 (declarer) になり、親と子で採点規則が非対称。
//   - スコアは素の整数。親は bid 達成時 +bid、未達時 -bid。子は獲得トリック数がそのまま加点 (+tricks)。
//   - ビッドは 5〜13、パスは BatakPassBid (=0)。競りは 1 周のみ。
//   - 親が第 1 トリックをリードする。
//   - リード時にスペードを切れる条件はブレイク済み or 手札がスペードのみ。
//   - リードスートに従えない場合は「必ずトランプ (スペード) を切らなければならない」。
//     ボイドかつスペードを持たない場合のみ任意のカードを捨てられる。
//   - ゲームは MaxRounds に達した時点で終了し、累積スコアが最大のプレイヤーが勝者。
type Batak struct {
	trumpCards       *TrumpCards
	players          []*BatakPlayer
	config           BatakConfig
	phase            BatakPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	spadesBroken     bool
	leadPlayerIdx    int
	bidPlayerIdx     int
	gameEndFlag      bool
	winnerIdx        int
	declarerIdx      int // 親 (最高ビッドの席、-1 = 未確定)
	highBid          int // 現在の最高ビッド (0 = 未宣言)
	bidStartIdx      int // ビッド開始席 (Reset で 0、NextRound でインクリメント)
	actionLogBase
}

// NewBatak コンストラクタ
func NewBatak(trumpCards *TrumpCards, players []*BatakPlayer, config BatakConfig) *Batak {
	return &Batak{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		declarerIdx: -1,
		highBid:     0,
		bidStartIdx: 0,
	}
}

// NewDefaultBatak は 4 人 (1 人間 + 3 CPU) の標準セットアップを返す。
// CUI / Web / Worker 共通の構築 SSoT。
func NewDefaultBatak() *Batak {
	players := []*BatakPlayer{
		NewBatakPlayer(true),
		NewBatakPlayer(false),
		NewBatakPlayer(false),
		NewBatakPlayer(false),
	}
	return NewBatak(NewTrumpCards(0), players, DefaultBatakConfig())
}

// Reset ゲーム初期化
func (cb *Batak) Reset() {
	cb.gameEndFlag = false
	cb.winnerIdx = -1
	cb.roundNumber = 1
	cb.trickNumber = 0
	cb.spadesBroken = false
	cb.currentTrick = nil
	cb.leadPlayerIdx = 0
	cb.currentPlayerIdx = -1
	cb.bidStartIdx = 0
	cb.bidPlayerIdx = cb.bidStartIdx
	cb.declarerIdx = -1
	cb.highBid = 0
	cb.actionLog = nil

	for _, p := range cb.players {
		p.bid = -1
		p.SetRoundScore(0)
		p.ResetTricks()
		p.Reset()
		p.SetIsFinished(false)
	}
	// 累積スコアはラウンド跨ぎで保持する値だが、Reset はゲーム開始時のみ呼ばれるので 0 に戻す
	for _, p := range cb.players {
		p.SetCumulativeScore(0)
	}

	cb.trumpCards.Shuffle()
	dealAllCards(cb.trumpCards, cb.players)
	cb.sortAllHands()

	cb.phase = BatakPhaseBid
}

// NextRound 次のラウンドを開始する
func (cb *Batak) NextRound() {
	if cb.phase != BatakPhaseRoundEnd {
		return
	}

	cb.roundNumber++
	cb.trickNumber = 0
	cb.spadesBroken = false
	cb.currentTrick = nil
	cb.leadPlayerIdx = 0
	cb.currentPlayerIdx = -1
	cb.bidStartIdx = (cb.bidStartIdx + 1) % BatakPlayerCnt
	cb.bidPlayerIdx = cb.bidStartIdx
	cb.declarerIdx = -1
	cb.highBid = 0

	for _, p := range cb.players {
		p.ResetRound()
	}

	cb.trumpCards.Shuffle()
	dealAllCards(cb.trumpCards, cb.players)
	cb.sortAllHands()

	cb.phase = BatakPhaseBid
}

// MinLegalBid は現在の状況で発言可能な最小ビッドを返す。
// パス以外で宣言できる最小値 (max(BatakMinBid, highBid+1)) を返し、
// それが BatakMaxBid (13) を超える場合は BatakPassBid (0, パスのみ可能) を返す。
func (cb *Batak) MinLegalBid() int {
	minBid := cb.highBid + 1
	if minBid < BatakMinBid {
		minBid = BatakMinBid
	}
	if minBid > BatakMaxBid {
		return BatakPassBid
	}
	return minBid
}

// PlayerBid 人間プレイヤーがビッドする
func (cb *Batak) PlayerBid(bid int) error {
	if cb.gameEndFlag {
		return ErrGameEnded
	}
	if cb.phase != BatakPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := findHumanIdx(cb.players)
	if humanIdx < 0 || cb.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}

	if bid == BatakPassBid {
		cb.players[humanIdx].SetBid(BatakPassBid)
		cb.appendLog(humanIdx, "bid", fmt.Sprintf("%s passes", playerName(cb.players, humanIdx)), nil)
	} else {
		minLegal := cb.MinLegalBid()
		if minLegal == BatakPassBid || bid < minLegal || bid > BatakMaxBid {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは %d または %d〜%d で指定してください", BatakPassBid, minLegal, BatakMaxBid))
		}
		cb.players[humanIdx].SetBid(bid)
		cb.highBid = bid
		cb.appendLog(humanIdx, "bid", fmt.Sprintf("%s bids %d", playerName(cb.players, humanIdx), bid), nil)
	}

	cb.bidPlayerIdx = (cb.bidPlayerIdx + 1) % BatakPlayerCnt
	cb.checkBidComplete()
	return nil
}

// CpuBid 現在のビッドプレイヤーが CPU の場合にビッドする
func (cb *Batak) CpuBid() {
	if cb.gameEndFlag || cb.phase != BatakPhaseBid {
		return
	}
	if cb.players[cb.bidPlayerIdx].GetIsHuman() {
		return
	}

	minLegal := cb.MinLegalBid()
	estimated := cb.cpuSelectBid(cb.bidPlayerIdx)

	var bid int
	if minLegal == BatakPassBid || estimated < minLegal {
		bid = BatakPassBid
	} else {
		bid = estimated
	}

	cb.players[cb.bidPlayerIdx].SetBid(bid)
	if bid == BatakPassBid {
		cb.appendLog(cb.bidPlayerIdx, "bid", fmt.Sprintf("%s passes", playerName(cb.players, cb.bidPlayerIdx)), nil)
	} else {
		cb.highBid = bid
		cb.appendLog(cb.bidPlayerIdx, "bid", fmt.Sprintf("%s bids %d", playerName(cb.players, cb.bidPlayerIdx), bid), nil)
	}

	cb.bidPlayerIdx = (cb.bidPlayerIdx + 1) % BatakPlayerCnt
	cb.checkBidComplete()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (cb *Batak) PlayerPlay(cardIndex int) error {
	if cb.gameEndFlag {
		return ErrGameEnded
	}
	if cb.phase != BatakPhasePlay {
		return ErrWrongPhase
	}
	if !cb.players[cb.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := cb.players[cb.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := cb.validatePlay(cb.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	cb.playCard(cb.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行
func (cb *Batak) CpuPlay() {
	if cb.gameEndFlag || cb.phase != BatakPhasePlay {
		return
	}
	if cb.players[cb.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := cb.players[cb.currentPlayerIdx]
	cardIdx := cb.cpuSelectPlayCard(cb.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	cb.playCard(cb.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (cb *Batak) ResolveTrick() {
	if cb.phase != BatakPhaseTrickEnd || len(cb.currentTrick) != BatakPlayerCnt {
		return
	}

	winnerIdx := cb.trickWinner()
	trickCards := make([]*Card, len(cb.currentTrick))
	for i, tc := range cb.currentTrick {
		trickCards[i] = tc.Card
	}

	cb.players[winnerIdx].AddTrick(trickCards)
	cb.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", playerName(cb.players, winnerIdx), cb.trickNumber), trickCards)

	cb.leadPlayerIdx = winnerIdx

	if cb.trickNumber >= BatakHandSize {
		cb.phase = BatakPhaseRoundEnd
	} else {
		cb.phase = BatakPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (cb *Batak) NextTrick() {
	if cb.phase != BatakPhaseTrickEnd {
		return
	}
	cb.currentTrick = nil
	cb.currentPlayerIdx = cb.leadPlayerIdx
	cb.trickNumber++
	cb.phase = BatakPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
//
// スコアは素の整数で計算する。
//   - 親 (i == declarerIdx): tricks >= bid なら +bid、下回れば -bid
//   - 子 (それ以外): +tricks (獲得トリック数がそのまま加点)
func (cb *Batak) ScoreRound() {
	if cb.phase != BatakPhaseRoundEnd {
		return
	}

	for i := 0; i < BatakPlayerCnt; i++ {
		p := cb.players[i]
		tricks := p.GetTrickCount()
		bid := p.GetBid()

		var score int
		if i == cb.declarerIdx {
			if tricks >= bid {
				score = bid
			} else {
				score = -bid
			}
		} else {
			score = tricks
		}
		p.SetRoundScore(score)

		cb.appendLog(i, "round_score", fmt.Sprintf("%s: bid=%d tricks=%d round=%d",
			playerName(cb.players, i), bid, tricks, score), nil)
	}

	// 累積スコアに加算
	for i := 0; i < BatakPlayerCnt; i++ {
		cb.players[i].CommitRoundScore()
	}

	// スコアログ
	for i := 0; i < BatakPlayerCnt; i++ {
		cb.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			playerName(cb.players, i), cb.players[i].GetCumulativeScore()), nil)
	}

	cb.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (cb *Batak) GetPhase() BatakPhase { return cb.phase }

// SetPhase フェーズ設定 (テスト用)
func (cb *Batak) SetPhase(phase BatakPhase) { cb.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (cb *Batak) GetRoundNumber() int { return cb.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (cb *Batak) SetRoundNumber(n int) { cb.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (cb *Batak) GetTrickNumber() int { return cb.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (cb *Batak) SetTrickNumber(n int) { cb.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (cb *Batak) GetCurrentPlayerIdx() int { return cb.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (cb *Batak) SetCurrentPlayerIdx(idx int) { cb.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (cb *Batak) GetCurrentTrick() []*TrickCard { return cb.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (cb *Batak) SetCurrentTrick(trick []*TrickCard) { cb.currentTrick = trick }

// GetSpadesBroken スペードブレイク状態取得
func (cb *Batak) GetSpadesBroken() bool { return cb.spadesBroken }

// SetSpadesBroken スペードブレイク状態設定 (テスト用)
func (cb *Batak) SetSpadesBroken(broken bool) { cb.spadesBroken = broken }

// GetGameEndFlag ゲーム終了フラグ取得
func (cb *Batak) GetGameEndFlag() bool { return cb.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (cb *Batak) GetWinnerIdx() int { return cb.winnerIdx }

// GetDeclarerIdx 親 (デクレアラー) インデックス取得 (-1 = 未確定)
func (cb *Batak) GetDeclarerIdx() int { return cb.declarerIdx }

// SetDeclarerIdx 親インデックス設定 (テスト用)
func (cb *Batak) SetDeclarerIdx(idx int) { cb.declarerIdx = idx }

// GetHighBid 現在の最高ビッド取得 (0 = 未宣言)
func (cb *Batak) GetHighBid() int { return cb.highBid }

// SetHighBid 現在の最高ビッド設定 (テスト用)
func (cb *Batak) SetHighBid(bid int) { cb.highBid = bid }

// GetBidStartIdx ビッド開始席インデックス取得
func (cb *Batak) GetBidStartIdx() int { return cb.bidStartIdx }

// SetBidStartIdx ビッド開始席インデックス設定 (テスト用)
func (cb *Batak) SetBidStartIdx(idx int) { cb.bidStartIdx = idx }

// GetPlayerCnt プレイヤー数取得
func (cb *Batak) GetPlayerCnt() int { return len(cb.players) }

// GetPlayer プレイヤー取得
func (cb *Batak) GetPlayer(i int) *BatakPlayer {
	return getPlayer(cb.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (cb *Batak) GetLeadPlayerIdx() int { return cb.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (cb *Batak) SetLeadPlayerIdx(idx int) { cb.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (cb *Batak) GetBidPlayerIdx() int { return cb.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (cb *Batak) SetBidPlayerIdx(idx int) { cb.bidPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (cb *Batak) IsHumanTurn() bool {
	return isHumanTurn(cb.players, cb.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (cb *Batak) IsHumanBidTurn() bool {
	if cb.phase != BatakPhaseBid {
		return false
	}
	return isHumanTurn(cb.players, cb.bidPlayerIdx)
}

// GetConfig 設定取得
func (cb *Batak) GetConfig() BatakConfig { return cb.config }

// SetConfig 設定変更
func (cb *Batak) SetConfig(cfg BatakConfig) { cb.config = cfg }

// --- Private methods ---

// checkBidComplete 全員がビッドしたかチェックし、親を決定してプレイフェーズに移行
func (cb *Batak) checkBidComplete() {
	for _, p := range cb.players {
		if p.GetBid() == -1 {
			return
		}
	}

	bestBid := 0
	bestIdx := -1
	for i, p := range cb.players {
		if p.GetBid() > bestBid {
			bestBid = p.GetBid()
			bestIdx = i
		}
	}

	if bestIdx == -1 {
		// 全員パスの場合:
		// 最後に発言した席 (bidStartIdx+3)%4 が最低額 BatakMinBid で強制的に親になる。
		// 再配りにしないのは、決定的でテストできる形を選んだため。
		forcedIdx := (cb.bidStartIdx + 3) % BatakPlayerCnt
		cb.players[forcedIdx].SetBid(BatakMinBid)
		cb.declarerIdx = forcedIdx
		cb.highBid = BatakMinBid
	} else {
		cb.declarerIdx = bestIdx
	}

	cb.leadPlayerIdx = cb.declarerIdx
	cb.phase = BatakPhasePlay
	cb.startPlayPhase()
}

// startPlayPhase プレイフェーズ開始。リードプレイヤーはラウンド先頭は 0、
// それ以降は前のトリックの勝者が引き継ぐ (leadPlayerIdx として保存済み)。
func (cb *Batak) startPlayPhase() {
	cb.trickNumber = 1
	cb.currentTrick = nil
	cb.currentPlayerIdx = cb.leadPlayerIdx
}

// playCard カードをプレイする共通処理
func (cb *Batak) playCard(playerIdx int, card *Card) {
	cb.currentTrick = append(cb.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	if card.GetDesign() == CardDesignSpade {
		cb.spadesBroken = true
	}

	cb.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(cb.players, playerIdx), cardStr(card)), []*Card{card})

	if len(cb.currentTrick) == BatakPlayerCnt {
		cb.phase = BatakPhaseTrickEnd
	} else {
		cb.currentPlayerIdx = (cb.currentPlayerIdx + 1) % BatakPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する
func (cb *Batak) validatePlay(playerIdx int, card *Card) error {
	if len(cb.currentTrick) == 0 {
		// リード: スペード未ブレイクの場合、スペードでリードできない (他にカードがある場合)
		if !cb.spadesBroken && card.GetDesign() == CardDesignSpade {
			if cb.playerHasNonSpade(playerIdx) {
				return NewDomainError(ErrInvalidPlay, "スペードはまだブレイクされていません")
			}
		}
		return nil
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()

	// フォロースート優先
	if cb.playerHasSuit(playerIdx, leadSuit) {
		if card.GetDesign() != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		return nil
	}

	// ボイド: スペード (トランプ) を持っている場合は必ず切る必要がある
	if leadSuit != CardDesignSpade && cb.playerHasSuit(playerIdx, CardDesignSpade) {
		if card.GetDesign() != CardDesignSpade {
			return NewDomainError(ErrInvalidPlay, "リードスートが無い場合はスペードで切らなければなりません")
		}
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (cb *Batak) playerHasSuit(playerIdx, design int) bool {
	return handHasSuit(cb.players[playerIdx], design)
}

// playerHasNonSpade プレイヤーがスペード以外のカードを持っているか
func (cb *Batak) playerHasNonSpade(playerIdx int) bool {
	p := cb.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() != CardDesignSpade {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する (スペードがトランプ)
func (cb *Batak) trickWinner() int {
	return ResolveTrickWinner(cb.currentTrick, CardDesignSpade, nil)
}

// checkGameEnd ゲーム終了判定: MaxRounds 到達でゲーム終了、最高スコアが勝者
func (cb *Batak) checkGameEnd() {
	if cb.roundNumber < cb.config.MaxRounds {
		return
	}

	cb.gameEndFlag = true
	cb.phase = BatakPhaseGameEnd

	maxScore := cb.players[0].GetCumulativeScore()
	cb.winnerIdx = 0
	for i := 1; i < BatakPlayerCnt; i++ {
		if cb.players[i].GetCumulativeScore() > maxScore {
			maxScore = cb.players[i].GetCumulativeScore()
			cb.winnerIdx = i
		}
	}
	cb.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(cb.players, cb.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする (スート → 値)
func (cb *Batak) sortAllHands() {
	for _, p := range cb.players {
		batakSortHand(p)
	}
}

// batakSortHand プレイヤーの手札をスート → 値の順にソートする
func batakSortHand(p *BatakPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// GetHint ヒントを取得する
func (cb *Batak) GetHint() *BatakHint {
	if cb.IsHumanBidTurn() {
		humanIdx := cb.bidPlayerIdx
		estimated := cb.cpuBidHard(humanIdx)
		minLegal := cb.MinLegalBid()

		var bid int
		var reason string
		if minLegal == BatakPassBid || estimated < minLegal {
			bid = BatakPassBid
			reason = "pass_weak_hand"
		} else {
			bid = estimated
			reason = "strategic_bid"
		}
		return &BatakHint{Bid: &bid, Reason: reason}
	}
	if cb.phase == BatakPhasePlay && cb.currentPlayerIdx == 0 {
		validIndices := cb.getValidPlayIndices(0)
		if len(validIndices) == 0 {
			return nil
		}
		idx := cb.cpuPlayHard(0, validIndices)
		return &BatakHint{CardIndex: &idx, Reason: cb.playHintReason(idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (cb *Batak) playHintReason(chosenIdx int) string {
	player := cb.players[0]
	card := player.GetCard(chosenIdx)

	if len(cb.currentTrick) == 0 {
		if player.GetTrickCount() < player.GetBid() {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == CardDesignSpade {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// cpuSelectBid CPU がビッドを選択する
func (cb *Batak) cpuSelectBid(playerIdx int) int {
	switch cb.config.CpuDifficulty {
	case BatakCpuDifficultyHard:
		return cb.cpuBidHard(playerIdx)
	case BatakCpuDifficultyNormal:
		return cb.cpuBidNormal(playerIdx)
	default:
		return cb.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy 大雑把なビッド見積り (パスしうる)
func (cb *Batak) cpuBidEasy(playerIdx int) int {
	player := cb.players[playerIdx]
	count := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() == CardDesignSpade || c.GetValue() == 1 || c.GetValue() == 13 {
			count++
		}
	}
	if count < 5 {
		return BatakPassBid
	}
	bid := BatakMinBid + rand.Intn(3)
	return cb.clampBid(bid)
}

// cpuBidNormal カードの強さに基づくビッド (半トリック単位の整数演算)
func (cb *Batak) cpuBidNormal(playerIdx int) int {
	return cb.estimateHandTricks(playerIdx, false)
}

// cpuBidHard 戦略的なビッド (半トリック単位の整数演算)
func (cb *Batak) cpuBidHard(playerIdx int) int {
	return cb.estimateHandTricks(playerIdx, true)
}

// estimateHandTricks は手札の期待獲得トリック数を半トリック単位 (1点 = 0.5トリック) で推定する。
// 切り札の長さ・絵札、サイドスートの A/K、ボイド/シングルトンによるラフ力を評価する。
func (cb *Batak) estimateHandTricks(playerIdx int, isHard bool) int {
	player := cb.players[playerIdx]
	suitCounts := [5]int{}
	hasAce := [5]bool{}
	hasKing := [5]bool{}
	hasQueen := [5]bool{}
	hasJack := [5]bool{}

	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		d := c.GetDesign()
		v := c.GetValue()
		suitCounts[d]++
		switch v {
		case 1:
			hasAce[d] = true
		case 13:
			hasKing[d] = true
		case 12:
			hasQueen[d] = true
		case 11:
			hasJack[d] = true
		}
	}
	spadeCount := suitCounts[CardDesignSpade]

	halfTricks := 0

	// 1. スペード (切り札) の絵札
	if hasAce[CardDesignSpade] {
		halfTricks += 2
	}
	if hasKing[CardDesignSpade] {
		if spadeCount >= 2 {
			halfTricks += 2
		} else {
			halfTricks += 1
		}
	}
	if hasQueen[CardDesignSpade] {
		if spadeCount >= 3 {
			halfTricks += 2
		} else if spadeCount >= 2 {
			halfTricks += 1
		}
	}
	if hasJack[CardDesignSpade] && spadeCount >= 3 {
		halfTricks += 1
	}

	// 2. 切り札の長さ
	// 4枚を超える分はほぼ確実にトリック、3〜4枚でもラフ・支配力がある
	if spadeCount == 3 {
		halfTricks += 1
	} else if spadeCount >= 4 {
		halfTricks += 2 + (spadeCount-4)*2
	}

	// 3. サイドスートの絵札
	for d := CardDesignClover; d <= CardDesignDiamond; d++ {
		cnt := suitCounts[d]
		if hasAce[d] {
			halfTricks += 2
			if hasKing[d] && cnt >= 2 {
				halfTricks += 2 // A+K が揃っていれば 2 トリック期待
			}
		} else if hasKing[d] {
			if cnt >= 2 {
				halfTricks += 1 // ガード付き K
			}
		}

		if hasQueen[d] && cnt >= 3 && (hasAce[d] || hasKing[d]) {
			halfTricks += 1
		}
	}

	// 4. ラフ力 (短スート)
	// 切り札が3枚以上ある場合、ボイドやシングルトンを刈れる
	if spadeCount >= 3 {
		for d := CardDesignClover; d <= CardDesignDiamond; d++ {
			if suitCounts[d] == 0 {
				halfTricks += 2 // ボイド刈り
			} else if suitCounts[d] == 1 && !hasAce[d] {
				// シングルトン: Hard はスペード4枚以上の余裕がある時のみ評価
				if !isHard || spadeCount >= 4 {
					halfTricks += 1
				}
			}
		}
	}

	bid := (halfTricks + 1) / 2
	return cb.clampBid(bid)
}

// clampBid ビッド値を BatakHandSize (13) 以下に収める
func (cb *Batak) clampBid(bid int) int {
	if bid > BatakHandSize {
		return BatakHandSize
	}
	return bid
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選択する
func (cb *Batak) cpuSelectPlayCard(playerIdx int) int {
	validIndices := cb.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch cb.config.CpuDifficulty {
	case BatakCpuDifficultyHard:
		return cb.cpuPlayHard(playerIdx, validIndices)
	case BatakCpuDifficultyNormal:
		return cb.cpuPlayNormal(playerIdx, validIndices)
	default:
		return cb.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (cb *Batak) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ビッドに近づくようにプレイ
func (cb *Batak) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := cb.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(cb.currentTrick) == 0 {
		if tricks < bid {
			return pickHighest(player, validIndices, nil)
		}
		return pickLowest(player, validIndices, nil)
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()
	highestInTrick := 0
	for _, tc := range cb.currentTrick {
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
	}

	leadSuitIndices := filterByDesign(player, validIndices, leadSuit)
	if len(leadSuitIndices) > 0 {
		if tricks < bid {
			over := filterAbove(player, leadSuitIndices, highestInTrick, nil)
			if len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		}
		return pickLowest(player, leadSuitIndices, nil)
	}

	// ボイド: validatePlay により残ったカードはルール上有効なものだけ。
	// スペードを必ず切るルールがある場合、validIndices はスペードのみのはず。
	if tricks < bid {
		return pickLowest(player, validIndices, nil)
	}
	return pickLowest(player, validIndices, nil)
}

// cpuPlayHard 高度な戦略プレイ
func (cb *Batak) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := cb.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(cb.currentTrick) == 0 {
		if tricks < bid {
			bestIdx := validIndices[0]
			bestScore := -1
			for _, idx := range validIndices {
				card := player.GetCard(idx)
				score := card.GetValue()
				if card.GetDesign() == CardDesignSpade {
					score += 100
				}
				if score > bestScore {
					bestScore = score
					bestIdx = idx
				}
			}
			return bestIdx
		}
		// 十分なトリック: 最も低い非スペードを出す
		bestIdx := validIndices[0]
		bestVal := player.GetCard(validIndices[0]).GetValue()
		bestIsSpade := player.GetCard(validIndices[0]).GetDesign() == CardDesignSpade
		for _, idx := range validIndices[1:] {
			card := player.GetCard(idx)
			isSpade := card.GetDesign() == CardDesignSpade
			if bestIsSpade && !isSpade {
				bestIdx = idx
				bestVal = card.GetValue()
				bestIsSpade = false
			} else if isSpade == bestIsSpade && card.GetValue() < bestVal {
				bestIdx = idx
				bestVal = card.GetValue()
			}
		}
		return bestIdx
	}

	leadSuit := cb.currentTrick[0].Card.GetDesign()
	highestSpadeInTrick, hasSpadeInTrick, highestInTrick := cb.summariseTrick(leadSuit)

	leadSuitIndices := filterByDesign(player, validIndices, leadSuit)
	if len(leadSuitIndices) > 0 {
		if tricks < bid && (!hasSpadeInTrick || leadSuit == CardDesignSpade) {
			threshold := highestInTrick
			if leadSuit == CardDesignSpade {
				threshold = highestSpadeInTrick
			}
			over := filterAbove(player, leadSuitIndices, threshold, nil)
			if len(over) > 0 {
				return pickLowest(player, over, nil)
			}
		}
		return pickLowest(player, leadSuitIndices, nil)
	}

	// ボイドかつ validIndices にスペードが含まれている場合は最小のスペードでカットを試みる。
	spadeIndices := filterByDesign(player, validIndices, CardDesignSpade)
	if len(spadeIndices) > 0 {
		if tricks < bid {
			if hasSpadeInTrick {
				if over := filterAbove(player, spadeIndices, highestSpadeInTrick, nil); len(over) > 0 {
					return pickLowest(player, over, nil)
				}
			} else {
				return pickLowest(player, spadeIndices, nil)
			}
		}
		// 余裕がある場合は最小スペードを温存気味に出す
		return pickLowest(player, spadeIndices, nil)
	}

	// スペードを持たないボイド: 最も高い不要カードを捨てる
	bestIdx := validIndices[0]
	bestScore := -1
	for _, idx := range validIndices {
		card := player.GetCard(idx)
		score := card.GetValue()
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// summariseTrick 現在のトリックの状態 (最高スペード、スペード有無、リードスート最高) を返す
func (cb *Batak) summariseTrick(leadSuit int) (highestSpade int, hasSpade bool, highestLead int) {
	for _, tc := range cb.currentTrick {
		if tc.Card.GetDesign() == CardDesignSpade {
			hasSpade = true
			if tc.Card.GetValue() > highestSpade {
				highestSpade = tc.Card.GetValue()
			}
		}
		if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestLead {
			highestLead = tc.Card.GetValue()
		}
	}
	return
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (cb *Batak) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(cb.players[playerIdx], func(c *Card) bool { return cb.validatePlay(playerIdx, c) == nil })
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web 用)
func (cb *Batak) GetValidPlayIndices(playerIdx int) []int {
	return cb.getValidPlayIndices(playerIdx)
}

// batakJSON is the JSON wire format for Batak.
type batakJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*BatakPlayer    `json:"ps"`
	Config           BatakConfig       `json:"cf"`
	Phase            BatakPhase        `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	SpadesBroken     bool              `json:"sb"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	DeclarerIdx      int               `json:"di"`
	HighBid          int               `json:"hb"`
	BidStartIdx      int               `json:"bs"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (cb *Batak) MarshalJSON() ([]byte, error) {
	return json.Marshal(batakJSON{
		TrumpCards:       cb.trumpCards,
		Players:          cb.players,
		Config:           cb.config,
		Phase:            cb.phase,
		RoundNumber:      cb.roundNumber,
		TrickNumber:      cb.trickNumber,
		CurrentPlayerIdx: cb.currentPlayerIdx,
		CurrentTrick:     cb.currentTrick,
		SpadesBroken:     cb.spadesBroken,
		LeadPlayerIdx:    cb.leadPlayerIdx,
		BidPlayerIdx:     cb.bidPlayerIdx,
		GameEndFlag:      cb.gameEndFlag,
		WinnerIdx:        cb.winnerIdx,
		DeclarerIdx:      cb.declarerIdx,
		HighBid:          cb.highBid,
		BidStartIdx:      cb.bidStartIdx,
		ActionLog:        cb.actionLog,
	})
}

// batakMaxSliceLen caps small fixed-size slice fields during deserialisation
// (players: max 4, currentTrick: max 4) to prevent excessive memory allocation.
const batakMaxSliceLen = 1000

// batakMaxActionLogLen caps the ActionLog slice during deserialisation.
// ActionLog grows by ~70 entries per round (52 plays + 4 bids + 13 trick winners +
// 1 score entry); 5000 accommodates ~71 rounds while still bounding DoS risk.
const batakMaxActionLogLen = 5000

// UnmarshalJSON implements json.Unmarshaler.
func (cb *Batak) UnmarshalJSON(data []byte) error {
	var j batakJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > batakMaxSliceLen || len(j.CurrentTrick) > batakMaxSliceLen ||
		len(j.ActionLog) > batakMaxActionLogLen {
		return fmt.Errorf("batak: input array exceeds maximum allowed size")
	}
	cb.trumpCards = j.TrumpCards
	if cb.trumpCards == nil {
		cb.trumpCards = NewTrumpCards(0)
	}
	cb.players = j.Players
	if cb.players == nil {
		cb.players = make([]*BatakPlayer, 0)
	}
	cb.config = j.Config
	cb.phase = j.Phase
	cb.roundNumber = j.RoundNumber
	cb.trickNumber = j.TrickNumber
	cb.currentPlayerIdx = j.CurrentPlayerIdx
	cb.currentTrick = j.CurrentTrick
	if cb.currentTrick == nil {
		cb.currentTrick = make([]*TrickCard, 0)
	}
	cb.spadesBroken = j.SpadesBroken
	cb.leadPlayerIdx = j.LeadPlayerIdx
	cb.bidPlayerIdx = j.BidPlayerIdx
	cb.gameEndFlag = j.GameEndFlag
	cb.winnerIdx = j.WinnerIdx
	cb.declarerIdx = j.DeclarerIdx
	cb.highBid = j.HighBid
	cb.bidStartIdx = j.BidStartIdx
	cb.actionLog = j.ActionLog
	if cb.actionLog == nil {
		cb.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
