package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
)

// AllFoursPlayerCnt All Fours プレイヤー数 (1 human vs 1 CPU)
const AllFoursPlayerCnt = 2

// AllFoursHandSize 各プレイヤーの初期手札枚数
const AllFoursHandSize = 6

// AllFoursRunDeal "run the cards" 時に各プレイヤーへ追加配布する枚数
const AllFoursRunDeal = 3

// AllFoursMaxRuns 1ディール中に "run the cards" を試行する最大回数
const AllFoursMaxRuns = 10

// AllFoursDealerIdx 親 (dealer) の固定インデックス (CPU)。非親 (elder hand) は人間。
const AllFoursDealerIdx = 1

// AllFoursNonDealerIdx 非親 (elder hand / 人間) のインデックス
const AllFoursNonDealerIdx = 0

// AllFoursTrumpUnset トランプ未確定を表す値
const AllFoursTrumpUnset = 0

// AllFoursPhase ゲームフェーズ
type AllFoursPhase int

// AllFoursのフェーズ定数
const (
	// AllFoursPhaseBeg 非親が stand / beg を選ぶフェーズ
	AllFoursPhaseBeg AllFoursPhase = 0
	// AllFoursPhaseGift 親が beg に応答するフェーズ (gift / run)
	AllFoursPhaseGift AllFoursPhase = 1
	// AllFoursPhasePlay トリックプレイフェーズ
	AllFoursPhasePlay AllFoursPhase = 2
	// AllFoursPhaseTrickEnd トリック終了フェーズ
	AllFoursPhaseTrickEnd AllFoursPhase = 3
	// AllFoursPhaseRoundEnd ラウンド終了フェーズ
	AllFoursPhaseRoundEnd AllFoursPhase = 4
	// AllFoursPhaseGameEnd ゲーム終了フェーズ
	AllFoursPhaseGameEnd AllFoursPhase = 5
)

// AllFoursHint ヒント情報
type AllFoursHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	Beg       *bool  // true=beg, false=stand (begフェーズ時)
	Run       *bool  // true=run the cards, false=gift (giftフェーズ時)
	Reason    string // ヒント理由キー
}

// AllFours All Fours (Seven Up / Old Sledge) ゲームクラス
type AllFours struct {
	trumpCards       *TrumpCards
	players          []*AllFoursPlayer
	config           AllFoursConfig
	phase            AllFoursPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	trumpSuit        int   // 切り札スート (AllFoursTrumpUnset=未確定)
	turnUp           *Card // めくり札 (provisional trump)
	runCount         int   // このディールで run the cards した回数
	giftAward        int   // gift で 1 点を得たプレイヤー (-1=gift なし)
	gameEndFlag      bool
	winnerIdx        int
	actionLogBase
}

// NewAllFours コンストラクタ
func NewAllFours(trumpCards *TrumpCards, players []*AllFoursPlayer, config AllFoursConfig) *AllFours {
	return &AllFours{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultAllFours returns AllFours with the standard 2-player setup
// (1 human elder hand, 1 CPU dealer) and DefaultAllFoursConfig.
func NewDefaultAllFours() *AllFours {
	players := []*AllFoursPlayer{
		NewAllFoursPlayer(true),  // AllFoursNonDealerIdx (elder hand)
		NewAllFoursPlayer(false), // AllFoursDealerIdx (dealer)
	}
	return NewAllFours(NewTrumpCards(0), players, DefaultAllFoursConfig())
}

// Reset ゲーム初期化
func (a *AllFours) Reset() {
	a.gameEndFlag = false
	a.winnerIdx = -1
	a.roundNumber = 1
	for _, pl := range a.players {
		pl.SetCumulativeScore(0)
		pl.ResetRound()
	}
	a.actionLog = nil
	a.startDeal()
}

// startDeal 1ディールを開始する (シャッフル → 配布 → めくり札 → begフェーズ)
func (a *AllFours) startDeal() {
	a.trickNumber = 0
	a.currentTrick = nil
	a.leadPlayerIdx = -1
	a.currentPlayerIdx = -1
	a.trumpSuit = AllFoursTrumpUnset
	a.turnUp = nil
	a.runCount = 0
	a.giftAward = -1
	for _, pl := range a.players {
		pl.ResetRound()
	}

	a.trumpCards.Shuffle()
	a.dealHands(AllFoursHandSize)
	a.turnUp = a.trumpCards.DrawCard()
	if a.turnUp != nil {
		a.trumpSuit = a.turnUp.GetDesign()
		a.appendLog(-1, "turn_up",
			fmt.Sprintf("Turn-up: %s (trump %s)", cardStr(a.turnUp), allFoursSuitGlyph(a.trumpSuit)), []*Card{a.turnUp})
	}
	a.sortAllHands()
	a.phase = AllFoursPhaseBeg
	a.appendLog(AllFoursNonDealerIdx, "beg_phase",
		fmt.Sprintf("%s to stand or beg", a.playerName(AllFoursNonDealerIdx)), nil)
}

// dealHands 各プレイヤーへ n 枚配る
func (a *AllFours) dealHands(n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < AllFoursPlayerCnt; j++ {
			card := a.trumpCards.DrawCard()
			if card == nil {
				return
			}
			a.players[j].AddCard(card)
		}
	}
}

// NextRound 次のディールを開始する
func (a *AllFours) NextRound() {
	if a.phase != AllFoursPhaseRoundEnd {
		return
	}
	a.roundNumber++
	a.startDeal()
}

// PlayerBeg 人間 (非親) が stand / beg を選ぶ (beg=true で beg)
func (a *AllFours) PlayerBeg(beg bool) error {
	if a.gameEndFlag {
		return ErrGameEnded
	}
	if a.phase != AllFoursPhaseBeg {
		return ErrWrongPhase
	}
	if !a.players[AllFoursNonDealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	a.applyBeg(beg)
	return nil
}

// CpuBeg 非親がCPUの場合に stand / beg を決める (デフォルト構成では呼ばれない)
func (a *AllFours) CpuBeg() {
	if a.gameEndFlag || a.phase != AllFoursPhaseBeg {
		return
	}
	if a.players[AllFoursNonDealerIdx].GetIsHuman() {
		return
	}
	a.applyBeg(a.cpuDecideBeg())
}

// applyBeg beg/stand の結果を反映する
func (a *AllFours) applyBeg(beg bool) {
	if !beg {
		a.appendLog(AllFoursNonDealerIdx, "stand",
			fmt.Sprintf("%s stands", a.playerName(AllFoursNonDealerIdx)), nil)
		a.startPlay()
		return
	}
	a.appendLog(AllFoursNonDealerIdx, "beg",
		fmt.Sprintf("%s begs", a.playerName(AllFoursNonDealerIdx)), nil)
	a.phase = AllFoursPhaseGift
	// 親がCPUの場合は即応答 (デフォルト構成)。人間親なら respond を待つ。
	if !a.players[AllFoursDealerIdx].GetIsHuman() {
		a.applyGiftResponse(a.cpuDecideRun())
	}
}

// PlayerRespondBeg 人間 (親) が beg に応答する (run=true で run the cards, false で gift)
func (a *AllFours) PlayerRespondBeg(run bool) error {
	if a.gameEndFlag {
		return ErrGameEnded
	}
	if a.phase != AllFoursPhaseGift {
		return ErrWrongPhase
	}
	if !a.players[AllFoursDealerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	a.applyGiftResponse(run)
	return nil
}

// applyGiftResponse 親の beg 応答 (gift or run) を反映する
func (a *AllFours) applyGiftResponse(run bool) {
	if !run {
		// Gift / take-it: 非親に 1 点 (round score) を与え、同じトランプでプレイ。
		// この 1 点は ScoreRound 時に High/Low/Jack/Game と合算され、peg-out 判定では
		// 集計の最初に加算される (gift は最も早く確定する得点とみなす)。
		a.players[AllFoursNonDealerIdx].SetRoundScore(a.players[AllFoursNonDealerIdx].GetRoundScore() + 1)
		a.giftAward = AllFoursNonDealerIdx
		a.appendLog(AllFoursNonDealerIdx, "gift",
			fmt.Sprintf("%s gifts 1 point to %s (take-it)",
				a.playerName(AllFoursDealerIdx), a.playerName(AllFoursNonDealerIdx)), nil)
		a.startPlay()
		return
	}
	a.runTheCards()
}

// runTheCards 追加で 3 枚ずつ配り、新しいめくり札を出す。
// 同じスートが続く限り上限まで再 run。デッキが尽きたら全再配布。
func (a *AllFours) runTheCards() {
	for {
		a.runCount++
		if a.runCount > AllFoursMaxRuns || a.trumpCards.GetRemainingCount() < AllFoursPlayerCnt*AllFoursRunDeal+1 {
			// デッキが尽きた / 上限到達 → 全カードを再配布する。
			a.appendLog(-1, "redeal",
				"Deck exhausted on run; re-dealing the hand", nil)
			a.redeal()
			return
		}
		a.dealHands(AllFoursRunDeal)
		newTurnUp := a.trumpCards.DrawCard()
		a.appendLog(-1, "run",
			fmt.Sprintf("Run the cards: new turn-up %s", cardStr(newTurnUp)), []*Card{newTurnUp})
		if newTurnUp == nil {
			a.redeal()
			return
		}
		if newTurnUp.GetDesign() != a.trumpSuit {
			a.turnUp = newTurnUp
			a.trumpSuit = newTurnUp.GetDesign()
			a.appendLog(-1, "trump_set",
				fmt.Sprintf("New trump %s", allFoursSuitGlyph(a.trumpSuit)), nil)
			a.sortAllHands()
			a.startPlay()
			return
		}
		// 同じスート: もう一度 run する (turn-up は脇に置く)。
		a.turnUp = newTurnUp
	}
}

// redeal 全カードを集め直し、新しいディールを開始する (同じ親のまま)。
func (a *AllFours) redeal() {
	a.trickNumber = 0
	a.currentTrick = nil
	a.leadPlayerIdx = -1
	a.currentPlayerIdx = -1
	a.trumpSuit = AllFoursTrumpUnset
	a.turnUp = nil
	a.runCount = 0
	a.giftAward = -1
	for _, pl := range a.players {
		pl.Reset()
		pl.ResetTricks()
		pl.SetRoundScore(0)
	}
	a.trumpCards.Shuffle()
	a.dealHands(AllFoursHandSize)
	a.turnUp = a.trumpCards.DrawCard()
	if a.turnUp != nil {
		a.trumpSuit = a.turnUp.GetDesign()
	}
	a.sortAllHands()
	a.phase = AllFoursPhaseBeg
	a.appendLog(AllFoursNonDealerIdx, "beg_phase",
		fmt.Sprintf("%s to stand or beg (re-deal)", a.playerName(AllFoursNonDealerIdx)), nil)
}

// startPlay プレイフェーズを開始する。非親 (elder hand) が最初にリードする。
func (a *AllFours) startPlay() {
	a.leadPlayerIdx = AllFoursNonDealerIdx
	a.currentPlayerIdx = AllFoursNonDealerIdx
	a.trickNumber = 1
	a.currentTrick = nil
	a.phase = AllFoursPhasePlay
	a.appendLog(AllFoursNonDealerIdx, "play_start",
		fmt.Sprintf("Trump is %s. %s leads.",
			allFoursSuitGlyph(a.trumpSuit), a.playerName(AllFoursNonDealerIdx)), nil)
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (a *AllFours) PlayerPlay(cardIndex int) error {
	if a.gameEndFlag {
		return ErrGameEnded
	}
	if a.phase != AllFoursPhasePlay {
		return ErrWrongPhase
	}
	if !a.players[a.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := a.players[a.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := a.validatePlay(a.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	a.playCard(a.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (a *AllFours) CpuPlay() {
	if a.gameEndFlag || a.phase != AllFoursPhasePlay {
		return
	}
	if a.players[a.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := a.players[a.currentPlayerIdx]
	cardIdx := a.cpuSelectPlayCard(a.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	a.playCard(a.currentPlayerIdx, played)
}

// playCard カードをプレイする共通処理
func (a *AllFours) playCard(playerIdx int, card *Card) {
	a.currentTrick = append(a.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	a.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", a.playerName(playerIdx), cardStr(card)), []*Card{card})
	if len(a.currentTrick) == AllFoursPlayerCnt {
		a.phase = AllFoursPhaseTrickEnd
	} else {
		a.currentPlayerIdx = (a.currentPlayerIdx + 1) % AllFoursPlayerCnt
	}
}

// validatePlay カードのプレイが有効か検証する。
// ルール: lead suit を持っている場合は lead suit OR trump のみプレイ可。
// lead suit を持っていなければ任意のカードをプレイ可。トランプはいつでも合法。
func (a *AllFours) validatePlay(playerIdx int, card *Card) error {
	if len(a.currentTrick) == 0 {
		return nil // リードは任意
	}
	leadSuit := a.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return nil
	}
	if a.trumpSuit != AllFoursTrumpUnset && card.GetDesign() == a.trumpSuit {
		return nil // トランプはいつでも合法
	}
	if a.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従うかトランプを切ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが特定のスートを持っているか
func (a *AllFours) playerHasSuit(playerIdx, design int) bool {
	pl := a.players[playerIdx]
	for i := 0; i < pl.GetCardsSize(); i++ {
		if pl.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// ResolveTrick トリックを解決して勝者を決定する
func (a *AllFours) ResolveTrick() {
	if a.phase != AllFoursPhaseTrickEnd || len(a.currentTrick) != AllFoursPlayerCnt {
		return
	}
	winnerIdx := a.trickWinner()
	trickCards := make([]*Card, len(a.currentTrick))
	for i, tc := range a.currentTrick {
		trickCards[i] = tc.Card
	}
	a.players[winnerIdx].AddTrick(trickCards)
	a.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", a.playerName(winnerIdx), a.trickNumber), trickCards)
	a.leadPlayerIdx = winnerIdx
	if a.players[AllFoursNonDealerIdx].GetCardsSize() == 0 &&
		a.players[AllFoursDealerIdx].GetCardsSize() == 0 {
		a.phase = AllFoursPhaseRoundEnd
	} else {
		a.phase = AllFoursPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (a *AllFours) NextTrick() {
	if a.phase != AllFoursPhaseTrickEnd {
		return
	}
	a.currentTrick = nil
	a.currentPlayerIdx = a.leadPlayerIdx
	a.trickNumber++
	a.phase = AllFoursPhasePlay
}

// trickWinner トリックの勝者を決定する (トランプ最高 > リード最高)
func (a *AllFours) trickWinner() int {
	if len(a.currentTrick) == 0 {
		return 0
	}
	leadSuit := a.currentTrick[0].Card.GetDesign()
	winnerIdx := a.currentTrick[0].PlayerIdx
	winnerRank := allFoursRankValue(a.currentTrick[0].Card.GetValue())
	winnerIsTrump := a.trumpSuit != AllFoursTrumpUnset && a.currentTrick[0].Card.GetDesign() == a.trumpSuit

	for _, tc := range a.currentTrick[1:] {
		isTrump := a.trumpSuit != AllFoursTrumpUnset && tc.Card.GetDesign() == a.trumpSuit
		rank := allFoursRankValue(tc.Card.GetValue())
		switch {
		case isTrump && !winnerIsTrump:
			winnerIdx, winnerRank, winnerIsTrump = tc.PlayerIdx, rank, true
		case isTrump && winnerIsTrump:
			if rank > winnerRank {
				winnerIdx, winnerRank = tc.PlayerIdx, rank
			}
		case !isTrump && !winnerIsTrump && tc.Card.GetDesign() == leadSuit:
			if rank > winnerRank {
				winnerIdx, winnerRank = tc.PlayerIdx, rank
			}
		}
	}
	return winnerIdx
}

// allFoursRankValue カード値 (1=A) を比較用ランクに変換 (A=14 > K=13 > … > 2=2)
func allFoursRankValue(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// allFoursPipValue Game ポイント計算用のピップ値: A=4 K=3 Q=2 J=1 10=10 他=0
func allFoursPipValue(v int) int {
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

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う。
//
// 集計順 (peg order): Gift → High → Low → Jack → Game。
// この順で各点を 1 つずつ加算していき、最初に PointLimit へ到達したプレイヤーが勝つ。
// 両者が同一ディールで上限へ達しても、この決定的な順序により勝者は一意に定まる。
func (a *AllFours) ScoreRound() {
	if a.phase != AllFoursPhaseRoundEnd {
		return
	}
	awards := a.pegAwards()

	// base = 今ディール開始時の cumulative スコア。gift 由来の round score は
	// awards に "gift" として含まれるため、round score から差し引いて二重計上を防ぐ。
	base := make([]int, AllFoursPlayerCnt)
	for i := 0; i < AllFoursPlayerCnt; i++ {
		base[i] = a.players[i].GetCumulativeScore()
		a.players[i].SetRoundScore(0)
	}

	limit := a.config.PointLimit
	winner := -1
	for _, aw := range awards {
		base[aw.player]++
		a.players[aw.player].SetRoundScore(a.players[aw.player].GetRoundScore() + 1)
		a.appendLog(aw.player, "score_"+aw.kind,
			fmt.Sprintf("%s scores %s", a.playerName(aw.player), aw.kind), nil)
		if winner < 0 && base[aw.player] >= limit {
			winner = aw.player
		}
	}

	for i := 0; i < AllFoursPlayerCnt; i++ {
		a.players[i].CommitRoundScore()
		a.appendLog(i, "cumulative_score",
			fmt.Sprintf("%s: total=%d", a.playerName(i), a.players[i].GetCumulativeScore()), nil)
	}

	if winner >= 0 {
		a.gameEndFlag = true
		a.phase = AllFoursPhaseGameEnd
		a.winnerIdx = winner
		a.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", a.playerName(winner)), nil)
	}
}

// allFoursAward は 1 点の獲得 (種別 + プレイヤー) を表す。
type allFoursAward struct {
	kind   string
	player int
}

// pegAwards 集計順 (Gift→High→Low→Jack→Game) に獲得した 1 点の列を返す。
// 各点は最大 1 つ。同点 Game は付与しない (慣習)。
//
// High/Low/Jack はラウンド終了時に「捕獲されたトランプ」から判定する。
// All Fours では全カードがプレイされるため、配られたトランプはすべていずれかの
// トリックに含まれる。よって捕獲済みトランプの最高/最低/J が in-play 集合と一致する。
func (a *AllFours) pegAwards() []allFoursAward {
	var awards []allFoursAward
	if a.giftAward >= 0 {
		awards = append(awards, allFoursAward{"gift", a.giftAward})
	}
	if a.trumpSuit == AllFoursTrumpUnset {
		return awards
	}

	// 1パス目: 捕獲済みトランプの最高/最低ランクと、J/Game 集計を求める。
	highRank, lowRank := -1, math.MaxInt32
	jackPlayer := -1
	gameTotals := make([]int, AllFoursPlayerCnt)
	type trumpCard struct {
		player int
		rank   int
	}
	var trumps []trumpCard
	for playerIdx, pl := range a.players {
		for _, trick := range pl.GetTricksTaken() {
			for _, card := range trick {
				gameTotals[playerIdx] += allFoursPipValue(card.GetValue())
				if card.GetDesign() != a.trumpSuit {
					continue
				}
				rank := allFoursRankValue(card.GetValue())
				trumps = append(trumps, trumpCard{playerIdx, rank})
				if rank > highRank {
					highRank = rank
				}
				if rank < lowRank {
					lowRank = rank
				}
				if card.GetValue() == 11 {
					jackPlayer = playerIdx
				}
			}
		}
	}

	highPlayer, lowPlayer := -1, -1
	for _, tc := range trumps {
		if tc.rank == highRank {
			highPlayer = tc.player
		}
		if tc.rank == lowRank {
			lowPlayer = tc.player
		}
	}

	if highPlayer >= 0 {
		awards = append(awards, allFoursAward{"high", highPlayer})
	}
	if lowPlayer >= 0 {
		awards = append(awards, allFoursAward{"low", lowPlayer})
	}
	if jackPlayer >= 0 {
		awards = append(awards, allFoursAward{"jack", jackPlayer})
	}

	gw, maxT, tied := -1, -1, false
	for i, total := range gameTotals {
		if total > maxT {
			maxT, gw, tied = total, i, false
		} else if total == maxT {
			tied = true
		}
	}
	if !tied && gw >= 0 && maxT > 0 {
		awards = append(awards, allFoursAward{"game", gw})
	}
	return awards
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (a *AllFours) GetPhase() AllFoursPhase { return a.phase }

// SetPhase フェーズ設定 (テスト用)
func (a *AllFours) SetPhase(phase AllFoursPhase) { a.phase = phase }

// GetRoundNumber ラウンド番号取得
func (a *AllFours) GetRoundNumber() int { return a.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (a *AllFours) SetRoundNumber(n int) { a.roundNumber = n }

// GetTrickNumber トリック番号取得
func (a *AllFours) GetTrickNumber() int { return a.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (a *AllFours) SetTrickNumber(n int) { a.trickNumber = n }

// GetDealerIdx ディーラー (親) インデックス取得
func (a *AllFours) GetDealerIdx() int { return AllFoursDealerIdx }

// GetNonDealerIdx 非親 (elder hand) インデックス取得
func (a *AllFours) GetNonDealerIdx() int { return AllFoursNonDealerIdx }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (a *AllFours) GetCurrentPlayerIdx() int { return a.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (a *AllFours) SetCurrentPlayerIdx(idx int) { a.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (a *AllFours) GetCurrentTrick() []*TrickCard { return a.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (a *AllFours) SetCurrentTrick(trick []*TrickCard) { a.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (a *AllFours) GetLeadPlayerIdx() int { return a.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (a *AllFours) SetLeadPlayerIdx(idx int) { a.leadPlayerIdx = idx }

// GetTrumpSuit 切り札スート取得 (AllFoursTrumpUnset=未確定)
func (a *AllFours) GetTrumpSuit() int { return a.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (a *AllFours) SetTrumpSuit(suit int) { a.trumpSuit = suit }

// GetTurnUp めくり札取得
func (a *AllFours) GetTurnUp() *Card { return a.turnUp }

// GetRunCount このディールの run 回数取得
func (a *AllFours) GetRunCount() int { return a.runCount }

// GetGameEndFlag ゲーム終了フラグ取得
func (a *AllFours) GetGameEndFlag() bool { return a.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1=未確定)
func (a *AllFours) GetWinnerIdx() int { return a.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (a *AllFours) GetPlayerCnt() int { return len(a.players) }

// GetPlayer プレイヤー取得
func (a *AllFours) GetPlayer(i int) *AllFoursPlayer {
	if i < 0 || i >= len(a.players) {
		return nil
	}
	return a.players[i]
}

// IsHumanTurn 現在の手番が人間かどうか
func (a *AllFours) IsHumanTurn() bool {
	if a.currentPlayerIdx < 0 || a.currentPlayerIdx >= len(a.players) {
		return false
	}
	return a.players[a.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (a *AllFours) GetConfig() AllFoursConfig { return a.config }

// SetConfig 設定変更
func (a *AllFours) SetConfig(cfg AllFoursConfig) { a.config = cfg }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (a *AllFours) GetValidPlayIndices(playerIdx int) []int {
	pl := a.players[playerIdx]
	var valid []int
	for i := 0; i < pl.GetCardsSize(); i++ {
		if a.validatePlay(playerIdx, pl.GetCard(i)) == nil {
			valid = append(valid, i)
		}
	}
	return valid
}

// --- private helpers ---

func (a *AllFours) sortAllHands() {
	for _, pl := range a.players {
		allFoursSortHand(pl)
	}
}

func allFoursSortHand(pl *AllFoursPlayer) {
	sortPlayerHand(pl, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return allFoursRankValue(ci.GetValue()) < allFoursRankValue(cj.GetValue())
	})
}

func (a *AllFours) playerName(idx int) string {
	if idx < 0 || idx >= len(a.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if a.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// allFoursSuitGlyph スートを Unicode グリフで返す。
func allFoursSuitGlyph(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "♠"
	case CardDesignClover:
		return "♣"
	case CardDesignHeart:
		return "♥"
	case CardDesignDiamond:
		return "♦"
	}
	return "?"
}

// --- CPU AI ---

// cpuDecideBeg 非親CPUが beg するか判断する (デフォルト構成では非親は人間)。
func (a *AllFours) cpuDecideBeg() bool {
	return a.handTrumpStrength(AllFoursNonDealerIdx) < 2
}

// cpuDecideRun 親CPUが beg に対し run するか (true) gift するか (false) 判断する。
func (a *AllFours) cpuDecideRun() bool {
	// 親のトランプが弱い場合は run して引き直す。強い場合は gift で 1 点渡しても勝ち取りに行く。
	strength := a.handTrumpStrength(AllFoursDealerIdx)
	return strength < 2
}

// handTrumpStrength 指定プレイヤーのトランプ手札の概算強さ (枚数 + 高位カード)。
func (a *AllFours) handTrumpStrength(playerIdx int) int {
	pl := a.players[playerIdx]
	strength := 0
	for i := 0; i < pl.GetCardsSize(); i++ {
		c := pl.GetCard(i)
		if c.GetDesign() != a.trumpSuit {
			continue
		}
		strength++
		if allFoursRankValue(c.GetValue()) >= 12 { // Q 以上
			strength++
		}
	}
	return strength
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (a *AllFours) cpuSelectPlayCard(playerIdx int) int {
	valid := a.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	switch a.config.CpuDifficulty {
	case AllFoursCpuDifficultyHard, AllFoursCpuDifficultyNormal:
		return a.cpuPlaySmart(playerIdx, valid)
	default:
		return valid[rand.Intn(len(valid))]
	}
}

// cpuPlaySmart トリックを取りたい場合は最弱の勝てるカード、取れない/取りたくない場合は最弱の安牌。
func (a *AllFours) cpuPlaySmart(playerIdx int, valid []int) int {
	pl := a.players[playerIdx]
	if len(a.currentTrick) == 0 {
		// リード: 高トランプがあれば High を取りに行く、なければ非トランプ高位 (Game狙い)。
		return a.cpuLead(pl, valid)
	}
	leadSuit := a.currentTrick[0].Card.GetDesign()
	bestRank, trumpInTrick := a.bestInTrick()
	// 勝てる最弱カードを探す。
	winIdx, winScore := -1, 1<<30
	for _, idx := range valid {
		c := pl.GetCard(idx)
		isTrump := a.trumpSuit != AllFoursTrumpUnset && c.GetDesign() == a.trumpSuit
		rank := allFoursRankValue(c.GetValue())
		canWin := false
		switch {
		case isTrump && !trumpInTrick:
			canWin = true
		case isTrump && trumpInTrick:
			canWin = rank > bestRank
		case !isTrump && !trumpInTrick && c.GetDesign() == leadSuit:
			canWin = rank > bestRank
		}
		if !canWin {
			continue
		}
		score := rank
		if isTrump {
			score += 100 // トランプは温存したいので勝てる中で優先度を下げる
		}
		if score < winScore {
			winScore, winIdx = score, idx
		}
	}
	// Game 点になる高位 (10/A) を相手に渡すのは避けたいので、勝てるなら取る。
	if winIdx >= 0 {
		return winIdx
	}
	// 勝てない: 最弱カードを捨てる (トランプ温存、高ピップ温存)。
	return a.cpuDiscard(pl, valid)
}

func (a *AllFours) cpuLead(pl *AllFoursPlayer, valid []int) int {
	// 高トランプ (A/K) があればリードで High を確保しに行く。
	bestIdx, bestScore := valid[0], -1
	for _, idx := range valid {
		c := pl.GetCard(idx)
		score := allFoursRankValue(c.GetValue())
		if a.trumpSuit != AllFoursTrumpUnset && c.GetDesign() == a.trumpSuit {
			score += 50
		}
		if score > bestScore {
			bestScore, bestIdx = score, idx
		}
	}
	return bestIdx
}

func (a *AllFours) cpuDiscard(pl *AllFoursPlayer, valid []int) int {
	bestIdx, bestScore := valid[0], 1<<30
	for _, idx := range valid {
		c := pl.GetCard(idx)
		score := allFoursRankValue(c.GetValue())
		if a.trumpSuit != AllFoursTrumpUnset && c.GetDesign() == a.trumpSuit {
			score += 100 // トランプ温存
		}
		if c.GetValue() == 10 || c.GetValue() == 1 {
			score += 5 // Game になる高ピップは捨てたくない
		}
		if score < bestScore {
			bestScore, bestIdx = score, idx
		}
	}
	return bestIdx
}

func (a *AllFours) bestInTrick() (int, bool) {
	leadSuit := a.currentTrick[0].Card.GetDesign()
	bestRank, trumpInTrick := -1, false
	for _, tc := range a.currentTrick {
		if a.trumpSuit != AllFoursTrumpUnset && tc.Card.GetDesign() == a.trumpSuit {
			trumpInTrick = true
			if r := allFoursRankValue(tc.Card.GetValue()); r > bestRank {
				bestRank = r
			}
		} else if !trumpInTrick && tc.Card.GetDesign() == leadSuit {
			if r := allFoursRankValue(tc.Card.GetValue()); r > bestRank {
				bestRank = r
			}
		}
	}
	return bestRank, trumpInTrick
}

// GetHint ヒントを取得する
func (a *AllFours) GetHint() *AllFoursHint {
	humanIdx := -1
	for i, pl := range a.players {
		if pl.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return nil
	}
	switch a.phase {
	case AllFoursPhaseBeg:
		if humanIdx != AllFoursNonDealerIdx {
			return nil
		}
		beg := a.handTrumpStrength(humanIdx) < 2
		reason := "beg_stand"
		if beg {
			reason = "beg_beg"
		}
		return &AllFoursHint{Beg: &beg, Reason: reason}
	case AllFoursPhaseGift:
		if humanIdx != AllFoursDealerIdx {
			return nil
		}
		run := a.handTrumpStrength(humanIdx) < 2
		reason := "gift_gift"
		if run {
			reason = "gift_run"
		}
		return &AllFoursHint{Run: &run, Reason: reason}
	case AllFoursPhasePlay:
		if a.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := a.GetValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := a.cpuSelectPlayCard(humanIdx)
		return &AllFoursHint{CardIndex: &idx, Reason: a.playHintReason(humanIdx, idx)}
	}
	return nil
}

func (a *AllFours) playHintReason(playerIdx, chosenIdx int) string {
	card := a.players[playerIdx].GetCard(chosenIdx)
	if len(a.currentTrick) == 0 {
		return "lead_strong"
	}
	leadSuit := a.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == a.trumpSuit {
		return "trump_cut"
	}
	return "discard_low"
}

// allFoursJSON is the JSON wire format for AllFours.
type allFoursJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*AllFoursPlayer `json:"ps"`
	Config           AllFoursConfig    `json:"cf"`
	Phase            AllFoursPhase     `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"li"`
	TrumpSuit        int               `json:"ts"`
	TurnUp           *Card             `json:"tu"`
	RunCount         int               `json:"rc"`
	GiftAward        int               `json:"ga"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (a *AllFours) MarshalJSON() ([]byte, error) {
	return json.Marshal(allFoursJSON{
		TrumpCards:       a.trumpCards,
		Players:          a.players,
		Config:           a.config,
		Phase:            a.phase,
		RoundNumber:      a.roundNumber,
		TrickNumber:      a.trickNumber,
		CurrentPlayerIdx: a.currentPlayerIdx,
		CurrentTrick:     a.currentTrick,
		LeadPlayerIdx:    a.leadPlayerIdx,
		TrumpSuit:        a.trumpSuit,
		TurnUp:           a.turnUp,
		RunCount:         a.runCount,
		GiftAward:        a.giftAward,
		GameEndFlag:      a.gameEndFlag,
		WinnerIdx:        a.winnerIdx,
		ActionLog:        a.actionLog,
	})
}

// allFoursMaxSliceLen caps slice sizes during deserialisation.
const allFoursMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (a *AllFours) UnmarshalJSON(data []byte) error {
	var j allFoursJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > allFoursMaxSliceLen || len(j.CurrentTrick) > allFoursMaxSliceLen ||
		len(j.ActionLog) > allFoursMaxSliceLen {
		return fmt.Errorf("allfours: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("allfours: invalid config: %w", err)
	}
	if j.Phase < AllFoursPhaseBeg || j.Phase > AllFoursPhaseGameEnd {
		return fmt.Errorf("allfours: invalid phase: %d", j.Phase)
	}
	if j.TrumpSuit < AllFoursTrumpUnset || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("allfours: invalid trump suit: %d", j.TrumpSuit)
	}
	for name, idx := range map[string]int{
		"currentPlayerIdx": j.CurrentPlayerIdx,
		"leadPlayerIdx":    j.LeadPlayerIdx,
		"winnerIdx":        j.WinnerIdx,
		"giftAward":        j.GiftAward,
	} {
		if idx < -1 || idx >= AllFoursPlayerCnt {
			return fmt.Errorf("allfours: invalid %s: %d", name, idx)
		}
	}
	a.trumpCards = j.TrumpCards
	if a.trumpCards == nil {
		a.trumpCards = NewTrumpCards(0)
	}
	a.players = j.Players
	if a.players == nil {
		a.players = make([]*AllFoursPlayer, 0)
	}
	a.config = j.Config
	a.phase = j.Phase
	a.roundNumber = j.RoundNumber
	a.trickNumber = j.TrickNumber
	a.currentPlayerIdx = j.CurrentPlayerIdx
	a.currentTrick = j.CurrentTrick
	if a.currentTrick == nil {
		a.currentTrick = make([]*TrickCard, 0)
	}
	a.leadPlayerIdx = j.LeadPlayerIdx
	a.trumpSuit = j.TrumpSuit
	a.turnUp = j.TurnUp
	a.runCount = j.RunCount
	a.giftAward = j.GiftAward
	a.gameEndFlag = j.GameEndFlag
	a.winnerIdx = j.WinnerIdx
	a.actionLog = j.ActionLog
	if a.actionLog == nil {
		a.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
