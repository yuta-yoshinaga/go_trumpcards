//go:build !js || !wasm || extra3

// Package domain ウルティ / ウルティモ (Ulti / Ultimó) のドメインモデル。
//
// Ulti はハンガリー発祥の 3 人用コントラクト・トリックテイキングゲーム。1 人のデクレアラー
// (declarer, この実装では常に人間=seat 0) が残り 2 人の CPU 連合 (coalition) と対戦する。
//
// # 簡略化ルール (本実装が採用する縮小版)
//
// 本実装は本来の Ulti (競争入札のオークションと十数種のコントラクト) を採用せず、以下に固定した
// 縮小ルールを実装する:
//
//   - デッキ: 32 枚 (A,10,K,Q,J,9,8,7 × 4 スート)。NewTrumpCardsPrsi と同一構成。
//   - トリックランク (高→低): A > 10 > K > Q > J > 9 > 8 > 7。生カード値ではなく
//     ultiTrickRank(value) の写像で比較する (A=値1 が最強、10=値10 が 2 番目)。
//   - カードポイント: A=10, 10=10, K=4, Q=3, J=2, 9/8/7=0。1 スート 29 点 × 4 = 116 点。
//     加えて最終トリック獲得に +10 点ボーナス → パルティ計算上の総計 126 点。
//   - 3 人固定。各 10 枚配り、残り 2 枚がタロン (talon)。
//   - フェーズ: Bid (デクレアラーがコントラクト+切り札を宣言) → Discard (タロン 2 枚を拾い 12 枚に
//     なった手札から 2 枚を伏せて捨てる) → Play (10 トリック) → RoundEnd → GameEnd。
//   - コントラクト (縮小版・以下の 3 種):
//     Party (パルティ): 切り札スートを指名。カードポイントの過半 (126 点中 61 点以上) を取れば勝ち。
//     勝ち→各相手から +2 / 負け→各相手へ -2。
//     Betli (ベトリ): 切り札なし。1 トリックも取らなければ勝ち。勝ち→+5 / 負け→-5。
//     Durchmarsch (ドゥルマルス): 切り札なし。全 10 トリックを取れば勝ち。勝ち→+6 / 負け→-6。
//   - 入札は非競争 (簡略化): 人間がコントラクト (Party は切り札スートも) を宣言するだけで、CPU は
//     一切上乗せしない。
//   - タロン: 宣言時にデクレアラーがタロン 2 枚を拾い (手札 12 枚)、正確に 2 枚を捨てる。捨札の点は
//     誰にも計上されない (Party の集計上も除外)。
//   - トリック: デクレアラーが第 1 トリックをリード。マストフォロー。ボイド時、Party では
//     オーバートランプ義務 (場の最強切り札を上回れるなら切らねばならない、上回れないなら任意の札)。
//     Betli / Durchmarsch (切り札なし) はボイド時に任意の札。切り札があれば最強切り札が、無ければ
//     リードスートの最強札が勝つ。Betli も A ハイの同一トリックランクを用いる (簡略化)。
//   - コイン精算: RoundEnd でコントラクト成否を判定し、デクレアラーと 2 人の連合の間でコインを移動。
//     TargetRounds ディール後、累積コイン最上位が勝者。winnerIdx = 累積コイン最上位。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// UltiPlayerCnt プレイヤー数 (人間 1 + CPU 2)
const UltiPlayerCnt = 3

// UltiHandSize 各プレイヤーの配り札枚数
const UltiHandSize = 10

// UltiTrickCount 1 ディールのトリック数
const UltiTrickCount = 10

// UltiDeckSize デッキ枚数 (32 枚デッキ)
const UltiDeckSize = 32

// UltiTalonSize タロン枚数
const UltiTalonSize = 2

// UltiDiscardSize 宣言後にデクレアラーが捨てる枚数
const UltiDiscardSize = 2

// UltiWinRounds マッチを構成するディール数 (既定)
const UltiWinRounds = 5

// UltiPartyThreshold Party の勝利に必要なカードポイント (126 点の過半)
const UltiPartyThreshold = 61

// UltiTotalPoints Party 計算上の総ポイント (116 カード点 + 10 最終トリックボーナス)
const UltiTotalPoints = 126

// UltiContract コントラクト種別
type UltiContract int

// Ulti のコントラクト定数
const (
	// UltiContractNone 未宣言
	UltiContractNone UltiContract = 0
	// UltiContractParty パルティ (切り札あり、過半のカードポイントを取る)
	UltiContractParty UltiContract = 1
	// UltiContractBetli ベトリ (切り札なし、1 トリックも取らない)
	UltiContractBetli UltiContract = 2
	// UltiContractDurchmarsch ドゥルマルス (切り札なし、全トリックを取る)
	UltiContractDurchmarsch UltiContract = 3
	// UltiContractUlti ウルティ (切り札あり、切り札の 7 で最終トリックを取る)。
	// ゲーム名の由来となるコントラクト。デクレアラーは切り札を指名し、切り札の 7 を
	// 最後まで温存して第 10 (最終) トリックをその 7 で勝たねばならない。失敗時は倍払い。
	UltiContractUlti UltiContract = 4
)

// UltiUltiStake ウルティ契約の 1 相手あたりのコイン額。
const UltiUltiStake = 4

// UltiUltiLossMultiplier ウルティ契約失敗時のコイン倍率 (倍払いルール)。
const UltiUltiLossMultiplier = 2

// UltiPhase ゲームフェーズ
type UltiPhase int

// Ulti のフェーズ定数
const (
	// UltiPhaseBid コントラクト宣言フェーズ
	UltiPhaseBid UltiPhase = 0
	// UltiPhaseDiscard タロン受け取り後の捨札フェーズ
	UltiPhaseDiscard UltiPhase = 1
	// UltiPhasePlay トリックプレイフェーズ
	UltiPhasePlay UltiPhase = 2
	// UltiPhaseTrickEnd トリック終了フェーズ
	UltiPhaseTrickEnd UltiPhase = 3
	// UltiPhaseRoundEnd ディール終了フェーズ
	UltiPhaseRoundEnd UltiPhase = 4
	// UltiPhaseGameEnd ゲーム終了フェーズ
	UltiPhaseGameEnd UltiPhase = 5
)

// UltiPhaseMin フェーズ下限 (検証用)
const UltiPhaseMin = int(UltiPhaseBid)

// UltiPhaseMax フェーズ上限 (検証用)
const UltiPhaseMax = int(UltiPhaseGameEnd)

// UltiOutcome ディール結果 (デクレアラー視点)
type UltiOutcome int

// Ulti のディール結果定数
const (
	// UltiOutcomeNone 未確定
	UltiOutcomeNone UltiOutcome = 0
	// UltiOutcomeWin デクレアラーがコントラクトを達成
	UltiOutcomeWin UltiOutcome = 1
	// UltiOutcomeLoss デクレアラーがコントラクトを失敗
	UltiOutcomeLoss UltiOutcome = 2
)

// UltiResult 人間視点のマッチ結果
type UltiResult int

// Ulti のマッチ結果定数
const (
	// UltiResultLose 敗北
	UltiResultLose UltiResult = -1
	// UltiResultNone 未確定 / 引き分け
	UltiResultNone UltiResult = 0
	// UltiResultWin 勝利
	UltiResultWin UltiResult = 1
)

// UltiHint ヒント情報
type UltiHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Ulti ウルティのゲームクラス
type Ulti struct {
	trumpCards       *TrumpCards
	players          []*UltiPlayer
	config           UltiConfig
	phase            UltiPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	declarerIdx      int          // デクレアラー (常に人間 seat 0)
	contract         UltiContract // 宣言されたコントラクト
	trumpSuit        int          // 切り札スート (Party のみ 1..4、それ以外 -1)
	talon            []*Card      // タロン (2 枚、宣言前に保持)
	talonTaken       bool         // タロンを拾ったか
	discards         []*Card      // デクレアラーが伏せて捨てた 2 枚
	playerCoins      [UltiPlayerCnt]int
	lastTrickWinner  int         // 最終トリック勝者 (-1=未確定)
	outcome          UltiOutcome // 直近ディールの結果
	result           UltiResult  // 人間視点のマッチ結果
	scored           bool        // 当該ディールの得点計算済みか (RoundEnd 突入時に一度だけ)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewUlti コンストラクタ
func NewUlti(trumpCards *TrumpCards, players []*UltiPlayer, config UltiConfig) *Ulti {
	return &Ulti{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     0,
		contract:        UltiContractNone,
		trumpSuit:       -1,
	}
}

// NewDefaultUlti 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultUlti() *Ulti {
	players := make([]*UltiPlayer, UltiPlayerCnt)
	players[0] = NewUltiPlayer(true)
	for i := 1; i < UltiPlayerCnt; i++ {
		players[i] = NewUltiPlayer(false)
	}
	return NewUlti(newUltiDeck(), players, DefaultUltiConfig())
}

// newUltiDeck Ulti 用 32 枚デッキを生成する (A,10,K,Q,J,9,8,7 × 4 スート)。
// NewTrumpCardsPrsi と同一構成 (German/Czech 32-card pack)。build-tag 無しの構成なので
// extra ワーカーからも到達可能。
func newUltiDeck() *TrumpCards {
	return NewTrumpCardsPrsi()
}

// Reset ゲーム初期化
func (g *Ulti) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerCoins = [UltiPlayerCnt]int{}
	g.result = UltiResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Ulti) NextRound() {
	if g.phase != UltiPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % UltiPlayerCnt
	g.startRound()
}

// startRound 手札を配り、宣言フェーズを開始する。
func (g *Ulti) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.lastTrickWinner = -1
	g.declarerIdx = findHumanIdx(g.players)
	if g.declarerIdx < 0 {
		g.declarerIdx = 0
	}
	g.contract = UltiContractNone
	g.trumpSuit = -1
	g.talon = nil
	g.talonTaken = false
	g.discards = nil
	g.outcome = UltiOutcomeNone
	g.scored = false
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.sortAllHands()
	g.currentPlayerIdx = g.declarerIdx
	g.phase = UltiPhaseBid
}

// deal 各プレイヤーへ UltiHandSize 枚を配り、残り 2 枚をタロンとする。
func (g *Ulti) deal() {
	for i := 0; i < UltiHandSize; i++ {
		for j := 0; j < UltiPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % UltiPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	g.talon = make([]*Card, 0, UltiTalonSize)
	for i := 0; i < UltiTalonSize; i++ {
		if c := g.trumpCards.DrawCard(); c != nil {
			g.talon = append(g.talon, c)
		}
	}
}

// --- Bidding (declaration) ---

// PlayerBid 人間デクレアラーがコントラクトを宣言する。Party のとき trumpSuit は 1..4。
func (g *Ulti) PlayerBid(contract UltiContract, trumpSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != UltiPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !ultiValidContract(contract) {
		return NewDomainError(ErrInvalidPlay, "コントラクトを選んでください (party/betli/durchmarsch/ulti)")
	}
	if ultiContractNeedsTrump(contract) && !ultiValidSuit(trumpSuit) {
		return NewDomainError(ErrInvalidPlay, "Party / Ulti では切り札スートを選んでください (1..4)")
	}
	g.applyBid(contract, trumpSuit)
	return nil
}

// CpuBid 入札は非競争のため CPU は宣言しない (インタフェース互換の no-op)。
func (g *Ulti) CpuBid() {}

// applyBid コントラクトを適用し、タロンを拾って捨札フェーズへ進む。
func (g *Ulti) applyBid(contract UltiContract, trumpSuit int) {
	g.contract = contract
	if ultiContractNeedsTrump(contract) {
		g.trumpSuit = trumpSuit
	} else {
		g.trumpSuit = -1
	}
	// タロンをデクレアラーの手札に加える (手札 12 枚)。
	for _, c := range g.talon {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.talon = make([]*Card, 0)
	g.talonTaken = true
	g.sortAllHands()
	g.appendLog(g.declarerIdx, "bid",
		fmt.Sprintf("%s declares %s (trump %s)", playerName(g.players, g.declarerIdx), ultiContractName(contract), ultiSuitName(g.trumpSuit)), nil)
	g.phase = UltiPhaseDiscard
}

// --- Discard ---

// PlayerDiscard デクレアラーがタロン受け取り後に 2 枚を伏せて捨てる。
func (g *Ulti) PlayerDiscard(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != UltiPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.declarerIdx]
	if len(cardIndices) != UltiDiscardSize {
		return NewDomainError(ErrInvalidPlay, "ちょうど 2 枚を捨ててください")
	}
	seen := map[int]bool{}
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "同じカードを 2 回選べません")
		}
		seen[idx] = true
	}
	// 大きいインデックスから削除して整合を保つ。
	sorted := append([]int(nil), cardIndices...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	g.discards = make([]*Card, 0, UltiDiscardSize)
	for _, idx := range sorted {
		g.discards = append(g.discards, player.RemoveCard(idx))
	}
	g.appendLog(g.declarerIdx, "discard",
		fmt.Sprintf("%s discards %d cards", playerName(g.players, g.declarerIdx), len(g.discards)), append([]*Card(nil), g.discards...))
	g.startPlay()
	return nil
}

// --- Play ---

// startPlay 捨札確定後、プレイフェーズを開始する (デクレアラーがリード)。
func (g *Ulti) startPlay() {
	g.sortAllHands()
	g.leadPlayerIdx = g.declarerIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber = 1
	g.currentTrick = nil
	g.phase = UltiPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Ulti) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != UltiPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Ulti) CpuPlay() {
	if g.gameEndFlag || g.phase != UltiPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *Ulti) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == UltiPlayerCnt {
		g.phase = UltiPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % UltiPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。最終トリックなら RoundEnd に入り、得点計算を発火する。
func (g *Ulti) ResolveTrick() {
	if g.phase != UltiPhaseTrickEnd || len(g.currentTrick) != UltiPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= UltiTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = UltiPhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = UltiPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Ulti) NextTrick() {
	if g.phase != UltiPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = UltiPhasePlay
}

// enterRoundEnd RoundEnd 突入時に一度だけ結果判定とコイン精算を行う (scored フラグでガード)。
func (g *Ulti) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.outcome = g.evalOutcome()
	g.applyScores(g.outcome)
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: declarer(%s) %s %s",
			g.roundNumber, playerName(g.players, g.declarerIdx), ultiContractName(g.contract), ultiOutcomeName(g.outcome)), nil)
	g.checkGameEnd()
}

// evalOutcome コントラクトの成否を判定する。
func (g *Ulti) evalOutcome() UltiOutcome {
	declTricks := g.players[g.declarerIdx].GetTrickCount()
	var won bool
	switch g.contract {
	case UltiContractBetli:
		won = declTricks == 0
	case UltiContractDurchmarsch:
		won = declTricks == UltiTrickCount
	case UltiContractUlti:
		won = g.declarerWonLastTrickWithTrumpSeven()
	default: // Party
		won = g.declarerCardPoints() >= UltiPartyThreshold
	}
	if won {
		return UltiOutcomeWin
	}
	return UltiOutcomeLoss
}

// declarerWonLastTrickWithTrumpSeven Ulti の成否判定: デクレアラーが最終トリックを
// 切り札の 7 で勝ったか。最終トリックの勝者がデクレアラーで、かつ当該トリックで
// デクレアラーが出した札が切り札の 7 のときにのみ真。
// (最終トリックの札は ResolveTrick が currentTrick を消さずに RoundEnd へ入るため参照可能。)
func (g *Ulti) declarerWonLastTrickWithTrumpSeven() bool {
	if g.lastTrickWinner != g.declarerIdx || !ultiValidSuit(g.trumpSuit) {
		return false
	}
	for _, tc := range g.currentTrick {
		if tc.PlayerIdx == g.declarerIdx {
			return tc.Card != nil && tc.Card.GetDesign() == g.trumpSuit && tc.Card.GetValue() == 7
		}
	}
	return false
}

// declarerCardPoints デクレアラーが獲得したカードポイント (+最終トリックボーナス) を返す。
func (g *Ulti) declarerCardPoints() int {
	pts := g.cardPointsOf(g.declarerIdx)
	if g.lastTrickWinner == g.declarerIdx {
		pts += 10
	}
	return pts
}

// cardPointsOf プレイヤー i が獲得したトリック中のカードポイント合計 (ボーナス無し) を返す。
func (g *Ulti) cardPointsOf(i int) int {
	if i < 0 || i >= len(g.players) {
		return 0
	}
	sum := 0
	for _, trick := range g.players[i].GetTricksTaken() {
		for _, c := range trick {
			sum += ultiCardPoints(c.GetValue())
		}
	}
	return sum
}

// applyScores ディール結果に応じて累積コインを更新する。
func (g *Ulti) applyScores(outcome UltiOutcome) {
	if outcome == UltiOutcomeNone {
		return
	}
	stake := ultiContractStake(g.contract)
	declWon := outcome == UltiOutcomeWin
	// ウルティは失敗時に倍払い。
	if g.contract == UltiContractUlti && !declWon {
		stake *= UltiUltiLossMultiplier
	}
	for i := 0; i < UltiPlayerCnt; i++ {
		if i == g.declarerIdx {
			if declWon {
				g.playerCoins[i] += stake * UltiCoalitionSize
			} else {
				g.playerCoins[i] -= stake * UltiCoalitionSize
			}
		} else {
			if declWon {
				g.playerCoins[i] -= stake
			} else {
				g.playerCoins[i] += stake
			}
		}
	}
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積コイン最上位を勝者とする。
func (g *Ulti) checkGameEnd() {
	if g.roundNumber < g.config.TargetRounds {
		return
	}
	leader, best := 0, g.playerCoins[0]
	tie := false
	for i := 1; i < UltiPlayerCnt; i++ {
		if g.playerCoins[i] > best {
			best = g.playerCoins[i]
			leader = i
			tie = false
		} else if g.playerCoins[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.winnerPlayer = leader
	g.phase = UltiPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
}

// humanResult 人間 (seat 0) の視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None、他は Lose。
func (g *Ulti) humanResult(leader int, tie bool) UltiResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return UltiResultNone
	}
	if g.playerCoins[human] == g.playerCoins[leader] {
		if tie {
			return UltiResultNone
		}
		return UltiResultWin
	}
	return UltiResultLose
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ、インタフェース互換)。
func (g *Ulti) ScoreRound() {
	if g.phase != UltiPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// --- Trick / play helpers ---

// validatePlay マストフォロー + Party のオーバートランプ義務を検証する。
func (g *Ulti) validatePlay(playerIdx int, card *Card) error {
	valid := g.getValidPlayIndices(playerIdx)
	player := g.players[playerIdx]
	for _, idx := range valid {
		if player.GetCard(idx) == card {
			return nil
		}
	}
	return NewDomainError(ErrInvalidPlay, "リードスートに従う (またはオーバートランプする) 必要があります")
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Ulti) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	all := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		all = append(all, i)
	}
	if len(g.currentTrick) == 0 {
		return all
	}
	led := g.currentTrick[0].Card.GetDesign()
	// リードスートを持っていれば必ず従う。
	follows := ultiFilter(all, func(idx int) bool {
		return player.GetCard(idx).GetDesign() == led
	})
	if len(follows) > 0 {
		return follows
	}
	// ボイド。切り札コントラクト以外は任意。
	if !ultiHasTrump(g.contract, g.trumpSuit) {
		return all
	}
	// Party / Ulti: オーバートランプ義務。場に切り札があり、それを上回れるなら上回る切り札を出す。
	highestTrumpRank, hasTrumpInTrick := g.highestTrumpInTrick()
	if !hasTrumpInTrick {
		return all
	}
	beating := ultiFilter(all, func(idx int) bool {
		c := player.GetCard(idx)
		return c.GetDesign() == g.trumpSuit && ultiTrickRank(c.GetValue()) > highestTrumpRank
	})
	if len(beating) > 0 {
		return beating
	}
	return all
}

// highestTrumpInTrick 現在のトリック中の最強切り札のトリックランクを返す。
func (g *Ulti) highestTrumpInTrick() (int, bool) {
	best, found := -1, false
	for _, tc := range g.currentTrick {
		if tc.Card.GetDesign() == g.trumpSuit {
			if r := ultiTrickRank(tc.Card.GetValue()); r > best {
				best = r
				found = true
			}
		}
	}
	return best, found
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、無ければリードスートの最強札が勝つ。
func (g *Ulti) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	led := g.currentTrick[0].Card.GetDesign()
	winIdx := g.currentTrick[0].PlayerIdx
	winCard := g.currentTrick[0].Card
	for _, tc := range g.currentTrick[1:] {
		if g.ultiBeats(tc.Card, winCard, led) {
			winIdx = tc.PlayerIdx
			winCard = tc.Card
		}
	}
	return winIdx
}

// ultiBeats cand が cur (現在の勝ち札) を上回るか。led はリードスート。
func (g *Ulti) ultiBeats(cand, cur *Card, led int) bool {
	trump := g.trumpSuit
	trumpActive := ultiHasTrump(g.contract, trump)
	candTrump := trumpActive && cand.GetDesign() == trump
	curTrump := trumpActive && cur.GetDesign() == trump
	if candTrump != curTrump {
		return candTrump // 切り札は非切り札に勝つ
	}
	if candTrump && curTrump {
		return ultiTrickRank(cand.GetValue()) > ultiTrickRank(cur.GetValue())
	}
	// どちらも非切り札。
	if cand.GetDesign() != led {
		return false // 場外の札は勝てない
	}
	if cur.GetDesign() != led {
		return true
	}
	return ultiTrickRank(cand.GetValue()) > ultiTrickRank(cur.GetValue())
}

// --- Card ranking ---

// ultiTrickRank トリックランク (A>10>K>Q>J>9>8>7)。生カード値ではなくこの写像で比較する。
func ultiTrickRank(value int) int {
	switch value {
	case 1: // A
		return 7
	case 10:
		return 6
	case 13: // K
		return 5
	case 12: // Q
		return 4
	case 11: // J
		return 3
	case 9:
		return 2
	case 8:
		return 1
	default: // 7
		return 0
	}
}

// ultiCardPoints カードポイント (A=10,10=10,K=4,Q=3,J=2, その他 0)。
func ultiCardPoints(value int) int {
	switch value {
	case 1, 10: // A, 10
		return 10
	case 13: // K
		return 4
	case 12: // Q
		return 3
	case 11: // J
		return 2
	default:
		return 0
	}
}

// ultiContractStake コントラクトごとの 1 相手あたりのコイン額。
func ultiContractStake(contract UltiContract) int {
	switch contract {
	case UltiContractBetli:
		return 5
	case UltiContractDurchmarsch:
		return 6
	case UltiContractUlti:
		return UltiUltiStake
	default: // Party
		return 2
	}
}

// ultiPlayStrength カード選択用の実効強さ。Party では切り札に大きな下駄を履かせる。
func ultiPlayStrength(c *Card, trump int, contract UltiContract) int {
	r := ultiTrickRank(c.GetValue())
	if ultiHasTrump(contract, trump) && c.GetDesign() == trump {
		r += 100
	}
	return r
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Ulti) sortAllHands() {
	for _, p := range g.players {
		ultiSortHand(p, g.trumpSuit, g.contract)
	}
}

// ultiSortHand 手札をスート→トリックランク降順にソートする (切り札は先頭にまとめる)。
func ultiSortHand(p *UltiPlayer, trump int, contract UltiContract) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	trumpActive := ultiHasTrump(contract, trump)
	sort.SliceStable(cards, func(i, j int) bool {
		ti := trumpActive && cards[i].GetDesign() == trump
		tj := trumpActive && cards[j].GetDesign() == trump
		if ti != tj {
			return ti // 切り札を先頭へ
		}
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return ultiTrickRank(cards[i].GetValue()) > ultiTrickRank(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// ultiContractName コントラクトの表示名を返す。
func ultiContractName(contract UltiContract) string {
	switch contract {
	case UltiContractParty:
		return "party"
	case UltiContractBetli:
		return "betli"
	case UltiContractDurchmarsch:
		return "durchmarsch"
	case UltiContractUlti:
		return "ulti"
	default:
		return "-"
	}
}

// ultiSuitName スートの表示名を返す。
func ultiSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "spades"
	case CardDesignClover:
		return "clubs"
	case CardDesignHeart:
		return "hearts"
	case CardDesignDiamond:
		return "diamonds"
	default:
		return "-"
	}
}

// ultiOutcomeName 結果の表示名を返す。
func ultiOutcomeName(o UltiOutcome) string {
	switch o {
	case UltiOutcomeWin:
		return "win"
	case UltiOutcomeLoss:
		return "loss"
	default:
		return "-"
	}
}

// ultiValidSuit suit が有効なスート (1..4) か。
func ultiValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// ultiValidContract contract が有効な宣言 (Party/Betli/Durchmarsch/Ulti) か。
func ultiValidContract(contract UltiContract) bool {
	return contract >= UltiContractParty && contract <= UltiContractUlti
}

// ultiContractNeedsTrump 切り札スートの指名を要するコントラクト (Party / Ulti) か。
func ultiContractNeedsTrump(contract UltiContract) bool {
	return contract == UltiContractParty || contract == UltiContractUlti
}

// ultiHasTrump 切り札が有効に機能する状況 (切り札コントラクト かつ 有効スート) か。
func ultiHasTrump(contract UltiContract, trump int) bool {
	return ultiContractNeedsTrump(contract) && ultiValidSuit(trump)
}

// ultiFilter 述語を満たすインデックスを抽出する。
func ultiFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Ulti) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Ulti) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == UltiCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart コントラクトと陣営 (デクレアラー vs 連合) を意識した戦略プレイ。
func (g *Ulti) cpuPlaySmart(playerIdx int, valid []int) int {
	isDecl := playerIdx == g.declarerIdx

	// リード。
	if len(g.currentTrick) == 0 {
		switch g.contract {
		case UltiContractBetli:
			return g.minByStrength(playerIdx, valid) // 低い札でリード
		case UltiContractDurchmarsch:
			return g.maxByStrength(playerIdx, valid) // 強い札で取りに行く
		default: // Party
			if isDecl {
				return g.maxByStrength(playerIdx, valid)
			}
			return g.minByStrength(playerIdx, valid)
		}
	}

	// フォロー。
	winnerIdx := g.trickWinner()
	winCard := g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card
	declWinning := winnerIdx == g.declarerIdx
	winners := ultiFilter(valid, func(idx int) bool {
		return g.ultiBeats(g.players[playerIdx].GetCard(idx), winCard, g.currentTrick[0].Card.GetDesign())
	})
	nonWinners := ultiFilter(valid, func(idx int) bool {
		return !g.ultiBeats(g.players[playerIdx].GetCard(idx), winCard, g.currentTrick[0].Card.GetDesign())
	})

	switch g.contract {
	case UltiContractBetli:
		// デクレアラーはトリックを取りたくない、連合はデクレアラーに取らせたい → いずれもダック。
		if len(nonWinners) > 0 {
			return g.maxByStrength(playerIdx, nonWinners) // 取らずに高い札を処理
		}
		return g.minByStrength(playerIdx, valid)
	case UltiContractDurchmarsch:
		// デクレアラーは全勝したい、連合は 1 つでも取りたい → いずれも取りに行く。
		if len(winners) > 0 {
			return g.minByStrength(playerIdx, winners)
		}
		return g.minByStrength(playerIdx, valid)
	default: // Party
		if isDecl {
			if !declWinning && len(winners) > 0 {
				return g.minByStrength(playerIdx, winners)
			}
			return g.minByStrength(playerIdx, valid)
		}
		// 連合: デクレアラーが勝っていて取れるなら取りに行く。
		if declWinning && len(winners) > 0 {
			return g.minByStrength(playerIdx, winners)
		}
		return g.minByStrength(playerIdx, valid)
	}
}

// minByStrength 実効強さが最小となるインデックスを返す。
func (g *Ulti) minByStrength(playerIdx int, indices []int) int {
	player := g.players[playerIdx]
	best := indices[0]
	bestScore := ultiPlayStrength(player.GetCard(best), g.trumpSuit, g.contract)
	for _, idx := range indices[1:] {
		if s := ultiPlayStrength(player.GetCard(idx), g.trumpSuit, g.contract); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByStrength 実効強さが最大となるインデックスを返す。
func (g *Ulti) maxByStrength(playerIdx int, indices []int) int {
	player := g.players[playerIdx]
	best := indices[0]
	bestScore := ultiPlayStrength(player.GetCard(best), g.trumpSuit, g.contract)
	for _, idx := range indices[1:] {
		if s := ultiPlayStrength(player.GetCard(idx), g.trumpSuit, g.contract); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Ulti) GetHint() *UltiHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case UltiPhaseBid:
		if g.declarerIdx != human {
			return nil
		}
		return &UltiHint{Reason: ultiBidHintReason(g.recommendContract(human))}
	case UltiPhaseDiscard:
		if g.declarerIdx != human {
			return nil
		}
		return &UltiHint{CardIndices: g.recommendDiscards(human), Reason: "discard_weak"}
	case UltiPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &UltiHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// recommendContract 手札から推奨コントラクトを選ぶ (ヒント用の簡易評価)。
func (g *Ulti) recommendContract(playerIdx int) UltiContract {
	p := g.players[playerIdx]
	high, low := 0, 0
	for i := 0; i < p.GetCardsSize(); i++ {
		r := ultiTrickRank(p.GetCard(i).GetValue())
		if r >= 5 { // A,10,K
			high++
		}
		if r <= 1 { // 8,7
			low++
		}
	}
	if high >= 8 {
		return UltiContractDurchmarsch
	}
	if low >= 8 {
		return UltiContractBetli
	}
	if g.hasProtectedTrumpSeven(playerIdx) {
		return UltiContractUlti
	}
	return UltiContractParty
}

// hasProtectedTrumpSeven ある 1 スートに 7 を含む長い保有 (5 枚以上) があるか。
// 高位の切り札で 7 を守り、最終トリックまで温存できる見込みがある手を Ulti 候補とする。
func (g *Ulti) hasProtectedTrumpSeven(playerIdx int) bool {
	p := g.players[playerIdx]
	counts := map[int]int{}
	hasSeven := map[int]bool{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		counts[c.GetDesign()]++
		if c.GetValue() == 7 {
			hasSeven[c.GetDesign()] = true
		}
	}
	for suit, n := range counts {
		if hasSeven[suit] && n >= 5 {
			return true
		}
	}
	return false
}

// recommendDiscards 捨てるべき 2 枚 (最も弱い札) のインデックスを返す。
func (g *Ulti) recommendDiscards(playerIdx int) []int {
	p := g.players[playerIdx]
	idxs := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		idxs = append(idxs, i)
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return ultiPlayStrength(p.GetCard(idxs[a]), g.trumpSuit, g.contract) <
			ultiPlayStrength(p.GetCard(idxs[b]), g.trumpSuit, g.contract)
	})
	if len(idxs) < UltiDiscardSize {
		return idxs
	}
	return idxs[:UltiDiscardSize]
}

// ultiBidHintReason 推奨コントラクトに対応するヒント理由キーを返す。
func ultiBidHintReason(contract UltiContract) string {
	switch contract {
	case UltiContractDurchmarsch:
		return "bid_durchmarsch"
	case UltiContractBetli:
		return "bid_betli"
	case UltiContractUlti:
		return "bid_ulti"
	default:
		return "bid_party"
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Ulti) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		if playerIdx == g.declarerIdx {
			return "lead_high"
		}
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	led := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	winCard := g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card
	if g.ultiBeats(card, winCard, led) {
		return "follow_win"
	}
	if card.GetDesign() != led {
		return "discard_low"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Ulti) GetPhase() UltiPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Ulti) SetPhase(phase UltiPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Ulti) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Ulti) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Ulti) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Ulti) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Ulti) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Ulti) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Ulti) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Ulti) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Ulti) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Ulti) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Ulti) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx デクレアラーインデックス取得
func (g *Ulti) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx デクレアラーインデックス設定 (テスト用)
func (g *Ulti) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract コントラクト取得
func (g *Ulti) GetContract() UltiContract { return g.contract }

// SetContract コントラクト設定 (テスト用)
func (g *Ulti) SetContract(c UltiContract) { g.contract = c }

// GetTrumpSuit 切り札スート取得 (-1=未確定/なし, 1..4)
func (g *Ulti) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Ulti) SetTrumpSuit(s int) { g.trumpSuit = s }

// GetTalonCount タロンの残り枚数取得
func (g *Ulti) GetTalonCount() int { return len(g.talon) }

// GetTalonTaken タロンを拾ったか取得
func (g *Ulti) GetTalonTaken() bool { return g.talonTaken }

// GetDiscardCount 捨札枚数取得
func (g *Ulti) GetDiscardCount() int { return len(g.discards) }

// GetPlayerCoins プレイヤー別累積コイン取得
func (g *Ulti) GetPlayerCoins() [UltiPlayerCnt]int { return g.playerCoins }

// SetPlayerCoins プレイヤー別累積コイン設定 (テスト用)
func (g *Ulti) SetPlayerCoins(s [UltiPlayerCnt]int) { g.playerCoins = s }

// GetCardPoints プレイヤー i が獲得したカードポイント (ボーナス無し) を返す。
func (g *Ulti) GetCardPoints(i int) int { return g.cardPointsOf(i) }

// GetOutcome 直近ディールの結果取得
func (g *Ulti) GetOutcome() UltiOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *Ulti) GetResult() UltiResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Ulti) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Ulti) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Ulti) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Ulti) GetPlayer(i int) *UltiPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (Bid/Discard/Play) が人間か。
func (g *Ulti) IsHumanTurn() bool {
	switch g.phase {
	case UltiPhaseBid, UltiPhaseDiscard:
		if g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
			return false
		}
		return g.players[g.declarerIdx].GetIsHuman()
	case UltiPhasePlay:
		if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
			return false
		}
		return g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// IsHumanBidTurn 現在のビッド (宣言) 手番が人間か。
func (g *Ulti) IsHumanBidTurn() bool {
	if g.phase != UltiPhaseBid {
		return false
	}
	if g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return false
	}
	return g.players[g.declarerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Ulti) GetConfig() UltiConfig { return g.config }

// SetConfig 設定変更
func (g *Ulti) SetConfig(cfg UltiConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Ulti) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != UltiPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// ultiJSON is the JSON wire format for Ulti.
type ultiJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*UltiPlayer      `json:"ps"`
	Config           UltiConfig         `json:"cf"`
	Phase            UltiPhase          `json:"ph"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	CurrentPlayerIdx int                `json:"ci"`
	CurrentTrick     []*TrickCard       `json:"ct"`
	LeadPlayerIdx    int                `json:"li"`
	DealerIdx        int                `json:"di"`
	DeclarerIdx      int                `json:"dc"`
	Contract         UltiContract       `json:"co"`
	TrumpSuit        int                `json:"ts"`
	Talon            []*Card            `json:"tl"`
	TalonTaken       bool               `json:"tt"`
	Discards         []*Card            `json:"ds"`
	PlayerCoins      [UltiPlayerCnt]int `json:"sc"`
	LastTrickWinner  int                `json:"lt"`
	Outcome          UltiOutcome        `json:"oc"`
	Result           UltiResult         `json:"rs"`
	Scored           bool               `json:"sd"`
	GameEndFlag      bool               `json:"ge"`
	WinnerPlayer     int                `json:"wp"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Ulti) MarshalJSON() ([]byte, error) {
	return json.Marshal(ultiJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		DeclarerIdx:      g.declarerIdx,
		Contract:         g.contract,
		TrumpSuit:        g.trumpSuit,
		Talon:            g.talon,
		TalonTaken:       g.talonTaken,
		Discards:         g.discards,
		PlayerCoins:      g.playerCoins,
		LastTrickWinner:  g.lastTrickWinner,
		Outcome:          g.outcome,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// ultiMaxSliceLen caps slice sizes during deserialisation.
const ultiMaxSliceLen = 5000

// errUltiOversized is the single sentinel error for oversized input arrays.
var errUltiOversized = errors.New("ulti: input array exceeds maximum allowed size")

// errUltiInvalidPlayers is returned when restored state lacks exactly UltiPlayerCnt players.
var errUltiInvalidPlayers = errors.New("ulti: invalid player count")

// errUltiInvalidTrick is returned when a restored trick card or its card is nil / out of range.
var errUltiInvalidTrick = errors.New("ulti: invalid trick card")

// errUltiInvalidCard is returned when a restored talon/discard card is nil.
var errUltiInvalidCard = errors.New("ulti: invalid card element")

// errUltiInvalidIndex is returned when a restored index field is out of range.
var errUltiInvalidIndex = errors.New("ulti: index field out of range")

// errUltiInvalidPhase is returned when a restored phase is out of range.
var errUltiInvalidPhase = errors.New("ulti: phase out of range")

// errUltiInvalidContract is returned when a restored contract value is out of range.
var errUltiInvalidContract = errors.New("ulti: contract value out of range")

// errUltiInvalidTrump is returned when a restored trump suit is out of range.
var errUltiInvalidTrump = errors.New("ulti: trump suit out of range")

// errUltiInvalidOutcome is returned when a restored outcome or result value is out of range.
var errUltiInvalidOutcome = errors.New("ulti: outcome/result value out of range")

// ultiInRange reports whether v is in [0, UltiPlayerCnt).
func ultiInRange(v int) bool { return v >= 0 && v < UltiPlayerCnt }

// ultiInRangeOrUnset reports whether v is -1 (unset) or in [0, UltiPlayerCnt).
func ultiInRangeOrUnset(v int) bool { return v == -1 || ultiInRange(v) }

// ultiValidContractVal reports whether c is a defined contract value (incl. None).
func ultiValidContractVal(c UltiContract) bool {
	return c >= UltiContractNone && c <= UltiContractUlti
}

// ultiTrumpInRangeOrUnset reports whether s is -1 (unset) or a valid suit (1..4).
func ultiTrumpInRangeOrUnset(s int) bool { return s == -1 || ultiValidSuit(s) }

// ultiCheckCards rejects a slice with any nil element.
func ultiCheckCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return errUltiInvalidCard
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Ulti) UnmarshalJSON(data []byte) error {
	var j ultiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ultiMaxSliceLen || len(j.CurrentTrick) > ultiMaxSliceLen ||
		len(j.ActionLog) > ultiMaxSliceLen || len(j.Talon) > ultiMaxSliceLen || len(j.Discards) > ultiMaxSliceLen {
		return errUltiOversized
	}
	if len(j.Players) != UltiPlayerCnt {
		return errUltiInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errUltiInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errUltiInvalidTrick
		}
		if !ultiInRange(tc.PlayerIdx) {
			return errUltiInvalidTrick
		}
	}
	if err := ultiCheckCards(j.Talon); err != nil {
		return err
	}
	if err := ultiCheckCards(j.Discards); err != nil {
		return err
	}
	// 範囲必須のインデックス [0, PlayerCnt)。
	if !ultiInRange(j.CurrentPlayerIdx) || !ultiInRange(j.DealerIdx) || !ultiInRange(j.DeclarerIdx) {
		return errUltiInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !ultiInRangeOrUnset(j.LeadPlayerIdx) || !ultiInRangeOrUnset(j.LastTrickWinner) ||
		!ultiInRangeOrUnset(j.WinnerPlayer) {
		return errUltiInvalidIndex
	}
	if int(j.Phase) < UltiPhaseMin || int(j.Phase) > UltiPhaseMax {
		return errUltiInvalidPhase
	}
	// フェーズ依存の厳格化: play 以降は declarer・lead・contract が確定していなければならない。
	if j.Phase >= UltiPhasePlay && j.Phase <= UltiPhaseRoundEnd {
		if !ultiInRange(j.LeadPlayerIdx) {
			return errUltiInvalidIndex
		}
		if !ultiValidContract(j.Contract) {
			return errUltiInvalidContract
		}
		if ultiContractNeedsTrump(j.Contract) && !ultiValidSuit(j.TrumpSuit) {
			return errUltiInvalidTrump
		}
	}
	if !ultiValidContractVal(j.Contract) {
		return errUltiInvalidContract
	}
	if !ultiTrumpInRangeOrUnset(j.TrumpSuit) {
		return errUltiInvalidTrump
	}
	if j.Outcome < UltiOutcomeNone || j.Outcome > UltiOutcomeLoss {
		return errUltiInvalidOutcome
	}
	if j.Result < UltiResultLose || j.Result > UltiResultWin {
		return errUltiInvalidOutcome
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newUltiDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.trumpSuit = j.TrumpSuit
	g.talon = j.Talon
	if g.talon == nil {
		g.talon = make([]*Card, 0)
	}
	g.talonTaken = j.TalonTaken
	g.discards = j.Discards
	if g.discards == nil {
		g.discards = make([]*Card, 0)
	}
	g.playerCoins = j.PlayerCoins
	g.lastTrickWinner = j.LastTrickWinner
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
