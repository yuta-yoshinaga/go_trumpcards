//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CoinchePlayerCnt コワンシュプレイヤー数
const CoinchePlayerCnt = 4

// CoincheHandSize 各プレイヤーの手札枚数
const CoincheHandSize = 8

// CoincheTeamCnt チーム数
const CoincheTeamCnt = 2

// CoincheFirstDealSize 初回配布で配るカード枚数 (ターンアップ前)
const CoincheFirstDealSize = 5

// CoincheDixDeDerDefault Dix de Der (最終トリック) ボーナス
const CoincheDixDeDerDefault = 10

// CoincheRebeloteBonus K+Q トランプの宣言ボーナス
const CoincheRebeloteBonus = 20

// CoincheRoundCardPointsTotal 1ラウンドのカード合計点数 (Dix de Der 含まず)
const CoincheRoundCardPointsTotal = 152

// CoinchePhase ゲームフェーズ
type CoinchePhase int

// Coincheのフェーズ定数
const (
	// CoinchePhaseBid 競りフェーズ (目標点と切り札スートを競り上げる)
	CoinchePhaseBid CoinchePhase = 0
	// CoinchePhaseDouble コワンシュフェーズ (相手が倍化 / 宣言側が再倍化)
	CoinchePhaseDouble CoinchePhase = 1
	// CoinchePhasePlay トリックプレイフェーズ
	CoinchePhasePlay CoinchePhase = 2
	// CoinchePhaseTrickEnd トリック終了フェーズ
	CoinchePhaseTrickEnd CoinchePhase = 3
	// CoinchePhaseRoundEnd ラウンド終了フェーズ
	CoinchePhaseRoundEnd CoinchePhase = 4
	// CoinchePhaseGameEnd ゲーム終了フェーズ
	CoinchePhaseGameEnd CoinchePhase = 5
)

// CoincheContractPoints は宣言できる目標点。
//
// **80 から 10 刻みで 180、その上が Capot。** ベロートと違い、コワンシュは
// 「何点取るか」を競り上げるので、契約は数値そのものが順位になる。
var CoincheContractPoints = []int{80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, CoincheCapotPoints}

// CoincheCapotPoints は Capot (全 8 トリック) の契約値。
const CoincheCapotPoints = 250

// CoincheDouble は倍率の状態。
type CoincheDouble int

// 倍率の状態。
const (
	// CoincheDoubleNone 倍化なし (×1)
	CoincheDoubleNone CoincheDouble = 0
	// CoincheDoubleCoinche 相手チームが倍化 (×2)
	CoincheDoubleCoinche CoincheDouble = 1
	// CoincheDoubleSurcoinche 宣言側が再倍化 (×4)
	CoincheDoubleSurcoinche CoincheDouble = 2
)

// CoincheMultiplier は倍率状態に対応する得点倍率を返す。
func CoincheMultiplier(d CoincheDouble) int {
	switch d {
	case CoincheDoubleCoinche:
		return 2
	case CoincheDoubleSurcoinche:
		return 4
	default:
		return 1
	}
}

// CoincheHint ヒント情報
type CoincheHint struct {
	CardIndex *int   // 推奨カードインデックス (プレイ時)
	Bid       *int   // 推奨目標点 (競り時; nil = パス推奨)
	Suit      *int   // 推奨切り札スート (競り時)
	Reason    string // ヒント理由キー
}

// Coinche コワンシュゲームクラス
type Coinche struct {
	trumpCards       *TrumpCards
	players          []*CoinchePlayer
	config           CoincheConfig
	phase            CoinchePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	trumpSuit        int // 切り札スート (0 = 未確定, CardDesignSpade等)
	makerTeam        int // 契約を取ったチーム (0 or 1)
	makerPlayerIdx   int // 契約を取ったプレイヤー
	// contractPoints は落札された目標点 (0 = 未落札)。**契約は数値**なので、
	// 競りの順位も成功条件もこの 1 つの値で決まる。
	contractPoints int
	// double は倍率の状態。コワンシュ (相手が倍化) とシュルコワンシュ
	// (宣言側が再倍化) は、競りが閉じたあとの別フェーズで決める。
	double            CoincheDouble
	teamScores        [CoincheTeamCnt]int
	roundPoints       [CoincheTeamCnt]int // 当ラウンドの累計カード点数
	roundBeloteBonus  [CoincheTeamCnt]int // 当ラウンドの Belote+Rebelote ボーナス
	beloteHolderIdx   int                 // K+Q を両方持つプレイヤー (-1 = なし)
	beloteKingPlayed  bool                // 当ラウンドで K of trumps が出されたか
	beloteQueenPlayed bool                // 当ラウンドで Q of trumps が出されたか
	beloteDeclared    bool                // 当ラウンドで Belote/Rebelote が宣言済か
	lastTrickWinner   int                 // 直近の Dix de Der 用 (-1 = 未確定)
	leadPlayerIdx     int
	bidPlayerIdx      int // 現在のビッド手番
	bidPassCount      int // 連続パス数 (両ラウンド合計; 8で再配布)
	gameEndFlag       bool
	winnerTeam        int // 勝利チーム (-1 = 未確定)
	actionLogBase
}

// NewCoinche コンストラクタ
func NewCoinche(trumpCards *TrumpCards, players []*CoinchePlayer, config CoincheConfig) *Coinche {
	return &Coinche{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerTeam:      -1,
		roundNumber:     0,
		dealerIdx:       0,
		beloteHolderIdx: -1,
		lastTrickWinner: -1,
	}
}

// NewDefaultCoinche 標準4人パートナーシップ構成 (人間チーム0, CPUは交互配置)
// と DefaultCoincheConfig を組み合わせたデフォルト構築。CUI/Web/Worker 共通の SSoT。
func NewDefaultCoinche() *Coinche {
	players := []*CoinchePlayer{
		NewCoinchePlayer(true, 0),
		NewCoinchePlayer(false, 1),
		NewCoinchePlayer(false, 0),
		NewCoinchePlayer(false, 1),
	}
	return NewCoinche(NewTrumpCards32(), players, DefaultCoincheConfig())
}

// Reset ゲーム初期化
func (b *Coinche) Reset() {
	b.gameEndFlag = false
	b.winnerTeam = -1
	b.roundNumber = 1
	b.trickNumber = 0
	// **開幕は人間から話す。** 競りはディーラーの左隣から始まるので、
	// ディーラーを人間の右隣に置く。0 にすると人間 (席 0) は毎回最後に話し、
	// 先に出た宣言を上回れる点しか選べない開幕になる。
	b.dealerIdx = CoinchePlayerCnt - 1
	b.teamScores = [CoincheTeamCnt]int{}
	b.actionLog = nil
	b.trumpSuit = 0
	b.makerTeam = 0
	b.makerPlayerIdx = -1

	for _, p := range b.players {
		p.ResetRound()
	}

	b.beginRound()
}

// NextRound 次のラウンドを開始する
func (b *Coinche) NextRound() {
	if b.phase != CoinchePhaseRoundEnd {
		return
	}

	b.roundNumber++
	b.dealerIdx = (b.dealerIdx + 1) % CoinchePlayerCnt
	b.trickNumber = 0
	b.trumpSuit = 0
	b.currentTrick = nil
	b.leadPlayerIdx = -1
	b.makerPlayerIdx = -1

	for _, p := range b.players {
		p.ResetRound()
	}

	b.beginRound()
}

// beginRound ラウンドの初期処理 (配布 + ビッドフェーズ突入)
func (b *Coinche) beginRound() {
	// Reset() doesn't clear these explicitly; clearing them here keeps a
	// mid-game reset from leaking ghost trick cards into the new bid phase.
	b.currentTrick = nil
	b.leadPlayerIdx = -1
	b.roundPoints = [CoincheTeamCnt]int{}
	b.roundBeloteBonus = [CoincheTeamCnt]int{}
	b.beloteHolderIdx = -1
	b.beloteKingPlayed = false
	b.beloteQueenPlayed = false
	b.beloteDeclared = false
	b.lastTrickWinner = -1
	b.bidPassCount = 0
	b.contractPoints = 0
	b.double = CoincheDoubleNone

	b.dealAll()
	b.phase = CoinchePhaseBid
	// **競りはディーラーの左隣から。** 席 0 が人間なので、開幕は人間が
	// 最初に話すようディーラーを右隣に置いてある (Reset を参照)。
	b.bidPlayerIdx = (b.dealerIdx + 1) % CoinchePlayerCnt
	b.currentPlayerIdx = b.bidPlayerIdx
}

// dealAll 32 枚を 4 人へ 8 枚ずつ配り切る。
//
// **ターンアップも山札の残りも無い。** クローン元のベロートは 5 枚配って
// 1 枚めくり、切り札が決まってから残りを配るが、コワンシュは配り切ってから
// 目標点を競るので、伏せ札が残っていると誰の手にも無い点が生まれる。
func (b *Coinche) dealAll() {
	b.trumpCards.Shuffle()
	for range CoincheHandSize {
		for j := range CoinchePlayerCnt {
			card := b.trumpCards.DrawCard()
			if card != nil {
				b.players[j].AddCard(card)
			}
		}
	}
	b.sortAllHands()
	b.detectBeloteHolder()
}

// detectBeloteHolder K+Q of trumps を両方持つプレイヤーを記録する
func (b *Coinche) detectBeloteHolder() {
	if !b.config.EnableBeloteRebelote {
		return
	}
	for i, p := range b.players {
		hasK := false
		hasQ := false
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c.GetDesign() != b.trumpSuit {
				continue
			}
			if c.GetValue() == 13 {
				hasK = true
			} else if c.GetValue() == 12 {
				hasQ = true
			}
		}
		if hasK && hasQ {
			b.beloteHolderIdx = i
			return
		}
	}
}

// --- Bid: the contract auction ---

// PlayerBid 人間プレイヤーが目標点と切り札スートを宣言する。
//
// **今出ている契約より高い点でなければ宣言できない。** 同点や下回る宣言を
// 通すと、先に宣言した側が黙って上書きされる。
func (b *Coinche) PlayerBid(points, suit int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != CoinchePhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(b.players)
	if humanIdx < 0 || b.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if suit < CardDesignSpade || suit > CardDesignMax {
		return NewDomainError(ErrInvalidPlay, "切り札スートを指定してください")
	}
	if !coincheIsBiddablePoints(points) {
		return NewDomainError(ErrInvalidPlay, "その目標点は宣言できません")
	}
	if points <= b.contractPoints {
		return NewDomainError(ErrInvalidPlay, "今の契約を上回る点が必要です")
	}
	b.doBid(humanIdx, points, suit)
	return nil
}

// PlayerPassBid 人間プレイヤーが競りでパスする。
func (b *Coinche) PlayerPassBid() error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != CoinchePhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(b.players)
	if humanIdx < 0 || b.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	b.doPassBid(humanIdx)
	return nil
}

// CpuBid CPU が競りで宣言またはパスする。
func (b *Coinche) CpuBid() {
	if b.gameEndFlag || b.phase != CoinchePhaseBid {
		return
	}
	if b.bidPlayerIdx < 0 || b.bidPlayerIdx >= CoinchePlayerCnt {
		return
	}
	if b.players[b.bidPlayerIdx].GetIsHuman() {
		return
	}
	if points, suit, ok := b.cpuSelectBid(b.bidPlayerIdx); ok {
		b.doBid(b.bidPlayerIdx, points, suit)
		return
	}
	b.doPassBid(b.bidPlayerIdx)
}

// GetMinBid は今この席が宣言できる最小の目標点を返す (0 = もう宣言できない)。
//
// 上回れない点を選択肢に出すと、押せるのに必ず拒否される操作面ができる。
func (b *Coinche) GetMinBid() int {
	for _, pts := range CoincheContractPoints {
		if pts > b.contractPoints {
			return pts
		}
	}
	return 0
}

// GetBiddablePoints は今この席が宣言できる目標点の一覧を返す。
func (b *Coinche) GetBiddablePoints() []int {
	var out []int
	for _, pts := range CoincheContractPoints {
		if pts > b.contractPoints {
			out = append(out, pts)
		}
	}
	return out
}

// coincheIsBiddablePoints は points が契約表にある値かを返す。
func coincheIsBiddablePoints(points int) bool {
	for _, pts := range CoincheContractPoints {
		if pts == points {
			return true
		}
	}
	return false
}

// doBid 宣言を記録して手番を進める。
func (b *Coinche) doBid(playerIdx, points, suit int) {
	b.contractPoints = points
	b.trumpSuit = suit
	b.makerTeam = b.players[playerIdx].GetTeam()
	b.makerPlayerIdx = playerIdx
	// **宣言のたびに連続パス数を戻す。** 戻さないと、競りの序盤に出た
	// パスが後の宣言を追い越して競りを閉じてしまう。
	b.bidPassCount = 0
	b.appendLog(playerIdx, "bid",
		fmt.Sprintf("%s bids %d in %s", playerName(b.players, playerIdx), points, suitStr(suit)), nil)
	b.advanceBid()
}

// doPassBid パスを記録して手番を進める。
func (b *Coinche) doPassBid(playerIdx int) {
	b.bidPassCount++
	b.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes", playerName(b.players, playerIdx)), nil)
	b.advanceBid()
}

// advanceBid 競りの手番を進め、閉じる条件を判定する。
//
// **閉じるのは「宣言のあとに 3 人連続パス」。** 4 人卓なので、宣言者以外の
// 全員がパスしたということ。誰も宣言しないまま 4 人がパスしたら、
// ディーラーの左隣が最低契約 (80) を引き受ける — 配り直しにすると、
// CPU が宣言しない手札が続いたときに終わらない。
func (b *Coinche) advanceBid() {
	b.bidPlayerIdx = (b.bidPlayerIdx + 1) % CoinchePlayerCnt
	b.currentPlayerIdx = b.bidPlayerIdx

	if b.contractPoints > 0 {
		if b.bidPassCount >= CoinchePlayerCnt-1 {
			b.closeBidding()
		}
		return
	}
	if b.bidPassCount >= CoinchePlayerCnt {
		forced := (b.dealerIdx + 1) % CoinchePlayerCnt
		b.contractPoints = CoincheContractPoints[0]
		b.trumpSuit = b.cpuBestSuit(forced)
		b.makerTeam = b.players[forced].GetTeam()
		b.makerPlayerIdx = forced
		b.appendLog(forced, "forced_bid",
			fmt.Sprintf("%s must take %d in %s (all passed)",
				playerName(b.players, forced), b.contractPoints, suitStr(b.trumpSuit)), nil)
		b.closeBidding()
	}
}

// closeBidding 競りを閉じ、コワンシュフェーズへ移る。
func (b *Coinche) closeBidding() {
	b.appendLog(b.makerPlayerIdx, "contract",
		fmt.Sprintf("Contract: %d in %s by team %d",
			b.contractPoints, suitStr(b.trumpSuit), b.makerTeam), nil)
	// 切り札が決まったので並べ直す。配った直後の並びは切り札を知らない。
	b.sortAllHands()
	b.detectBeloteHolder()
	b.phase = CoinchePhaseDouble
	// 倍化を決めるのは相手チーム。人間が相手側なら人間の手番。
	b.currentPlayerIdx = b.firstOpponentOf(b.makerTeam)
}

// SetPhaseForDoubleTest は倍化フェーズへ進め、判断する席を据える (テスト用)。
func (b *Coinche) SetPhaseForDoubleTest() {
	b.phase = CoinchePhaseDouble
	b.currentPlayerIdx = b.firstOpponentOf(b.makerTeam)
}

// firstOpponentOf は team の相手チームで、ディーラーの左隣から見て最初の席。
func (b *Coinche) firstOpponentOf(team int) int {
	start := (b.dealerIdx + 1) % CoinchePlayerCnt
	for i := range CoinchePlayerCnt {
		idx := (start + i) % CoinchePlayerCnt
		if b.players[idx].GetTeam() != team {
			return idx
		}
	}
	return start
}

// --- Double: coinche / surcoinche ---

// PlayerCoinche 人間プレイヤー (守備側) が倍化する。
func (b *Coinche) PlayerCoinche() error {
	if err := b.checkDoubleTurn(false); err != nil {
		return err
	}
	b.double = CoincheDoubleCoinche
	b.appendLog(b.currentPlayerIdx, "coinche",
		fmt.Sprintf("%s coinches (x2)", playerName(b.players, b.currentPlayerIdx)), nil)
	// 倍化されたら、宣言側に再倍化の機会が回る。
	b.currentPlayerIdx = b.makerPlayerIdx
	return nil
}

// PlayerSurcoinche 人間プレイヤー (宣言側) が再倍化する。
func (b *Coinche) PlayerSurcoinche() error {
	if err := b.checkDoubleTurn(true); err != nil {
		return err
	}
	b.double = CoincheDoubleSurcoinche
	b.appendLog(b.currentPlayerIdx, "surcoinche",
		fmt.Sprintf("%s surcoinches (x4)", playerName(b.players, b.currentPlayerIdx)), nil)
	b.startPlayPhase()
	return nil
}

// PlayerDeclineDouble 人間プレイヤーが倍化せずに進める。
func (b *Coinche) PlayerDeclineDouble() error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != CoinchePhaseDouble {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(b.players)
	if humanIdx < 0 || b.currentPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	b.startPlayPhase()
	return nil
}

// checkDoubleTurn は倍化操作が今できるかを検査する。
// maker が true なら宣言側 (シュルコワンシュ)、false なら守備側 (コワンシュ)。
func (b *Coinche) checkDoubleTurn(maker bool) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != CoinchePhaseDouble {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(b.players)
	if humanIdx < 0 || b.currentPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	onMakerTeam := b.players[humanIdx].GetTeam() == b.makerTeam
	if maker {
		// シュルコワンシュは「コワンシュされた宣言側」だけ。
		if !onMakerTeam || b.double != CoincheDoubleCoinche {
			return NewDomainError(ErrInvalidPlay, "再倍化はできません")
		}
		return nil
	}
	if onMakerTeam || b.double != CoincheDoubleNone {
		return NewDomainError(ErrInvalidPlay, "倍化はできません")
	}
	return nil
}

// CpuDouble CPU の倍化判断。
func (b *Coinche) CpuDouble() {
	if b.gameEndFlag || b.phase != CoinchePhaseDouble {
		return
	}
	idx := b.currentPlayerIdx
	if idx < 0 || idx >= CoinchePlayerCnt || b.players[idx].GetIsHuman() {
		return
	}
	if b.double == CoincheDoubleNone && b.players[idx].GetTeam() != b.makerTeam && b.cpuWantsCoinche(idx) {
		b.double = CoincheDoubleCoinche
		b.appendLog(idx, "coinche", fmt.Sprintf("%s coinches (x2)", playerName(b.players, idx)), nil)
		b.currentPlayerIdx = b.makerPlayerIdx
		return
	}
	if b.double == CoincheDoubleCoinche && b.players[idx].GetTeam() == b.makerTeam && b.cpuWantsSurcoinche(idx) {
		b.double = CoincheDoubleSurcoinche
		b.appendLog(idx, "surcoinche", fmt.Sprintf("%s surcoinches (x4)", playerName(b.players, idx)), nil)
	}
	b.startPlayPhase()
}

// --- Play Phase ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Coinche) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != CoinchePhasePlay {
		return ErrWrongPhase
	}
	if !b.players[b.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := b.players[b.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := b.validatePlay(b.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	b.playCard(b.currentPlayerIdx, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行
func (b *Coinche) CpuPlay() {
	if b.gameEndFlag || b.phase != CoinchePhasePlay {
		return
	}
	if b.players[b.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := b.players[b.currentPlayerIdx]
	cardIdx := b.cpuSelectPlayCard(b.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	b.playCard(b.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (b *Coinche) ResolveTrick() {
	if b.phase != CoinchePhaseTrickEnd || len(b.currentTrick) != CoinchePlayerCnt {
		return
	}

	winnerIdx := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	trickPoints := 0
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += coincheCardPoints(tc.Card, b.trumpSuit)
	}

	b.players[winnerIdx].AddTrick(trickCards)
	b.roundPoints[b.players[winnerIdx].GetTeam()] += trickPoints

	winnerName := playerName(b.players, winnerIdx)
	b.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pts)", winnerName, b.trickNumber, trickPoints),
		trickCards)

	b.leadPlayerIdx = winnerIdx
	b.lastTrickWinner = winnerIdx

	if b.trickNumber >= CoincheHandSize {
		// Dix de Der
		b.roundPoints[b.players[winnerIdx].GetTeam()] += b.config.DixDeDer
		b.appendLog(winnerIdx, "dix_de_der",
			fmt.Sprintf("%s wins last trick +%d (Dix de Der)", winnerName, b.config.DixDeDer), nil)
		b.phase = CoinchePhaseRoundEnd
	} else {
		b.phase = CoinchePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (b *Coinche) NextTrick() {
	if b.phase != CoinchePhaseTrickEnd {
		return
	}
	b.currentTrick = nil
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber++
	b.phase = CoinchePhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (b *Coinche) ScoreRound() {
	if b.phase != CoinchePhaseRoundEnd {
		return
	}

	maker := b.makerTeam
	defender := 1 - b.makerTeam
	mult := CoincheMultiplier(b.double)

	// **契約に届いたかで勝敗が決まる。** クローン元のベロートは「カード点が
	// 多いほうの勝ち」だったが、コワンシュは宣言した目標点に届いたかどうか。
	// カード点で上回っていても契約に 1 点足りなければ落ちる。
	makerCardPts := b.roundPoints[maker]
	made := makerCardPts >= b.contractPoints

	makerTricks := 0
	for _, p := range b.players {
		if p.GetTeam() == maker {
			makerTricks += p.GetTrickCount()
		}
	}
	// Capot 契約は全 8 トリックが成功条件。点では代えられない。
	if b.contractPoints == CoincheCapotPoints {
		made = makerTricks == CoincheHandSize
	}

	if made {
		gain := (b.contractPoints + makerCardPts) * mult
		b.teamScores[maker] += gain
		b.appendLog(-1, "contract_made",
			fmt.Sprintf("Team %d makes %d (took %d, x%d): +%d",
				maker, b.contractPoints, makerCardPts, mult, gain), nil)
	} else {
		// **dedans は総取り。** 守備側が「場の総点 + 契約」を倍率込みで取り、
		// 宣言側は 0。取ったカード点は宣言側に残らない。
		// **総点は設定から導く。** Dix de Der は設定値なので、152+10 を
		// 直接書くと設定を変えたときだけ精算が静かにずれる。
		gain := (CoincheRoundCardPointsTotal + b.config.DixDeDer + b.contractPoints) * mult
		b.teamScores[defender] += gain
		b.appendLog(-1, "dedans",
			fmt.Sprintf("Team %d is dedans on %d (took %d, x%d): team %d +%d",
				maker, b.contractPoints, makerCardPts, mult, defender, gain), nil)
	}

	// **Belote/Rebelote は契約の成否と無関係。** 宣言した側のチームに残る。
	for ti := range CoincheTeamCnt {
		if bonus := b.roundBeloteBonus[ti]; bonus > 0 {
			b.teamScores[ti] += bonus
			b.appendLog(-1, "belote_bonus",
				fmt.Sprintf("Team %d keeps Belote/Rebelote: +%d", ti, bonus), nil)
		}
	}

	for ti := range CoincheTeamCnt {
		b.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d card points (total %d)",
				ti, b.roundPoints[ti], b.teamScores[ti]), nil)
	}

	b.checkGameEnd()
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (b *Coinche) GetPhase() CoinchePhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Coinche) SetPhase(p CoinchePhase) { b.phase = p }

// GetRoundNumber 現在のラウンド番号取得
func (b *Coinche) GetRoundNumber() int { return b.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (b *Coinche) SetRoundNumber(n int) { b.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (b *Coinche) GetTrickNumber() int { return b.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (b *Coinche) SetTrickNumber(n int) { b.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (b *Coinche) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (b *Coinche) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (b *Coinche) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (b *Coinche) SetCurrentTrick(trick []*TrickCard) { b.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Coinche) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (b *Coinche) GetWinnerTeam() int { return b.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (b *Coinche) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Coinche) GetPlayer(i int) *CoinchePlayer {
	return getPlayer(b.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Coinche) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (b *Coinche) SetLeadPlayerIdx(idx int) { b.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (b *Coinche) GetBidPlayerIdx() int { return b.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (b *Coinche) SetBidPlayerIdx(idx int) { b.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (b *Coinche) GetDealerIdx() int { return b.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (b *Coinche) SetDealerIdx(idx int) { b.dealerIdx = idx }

// GetTrumpSuit 切り札スート取得
func (b *Coinche) GetTrumpSuit() int { return b.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (b *Coinche) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// GetContractPoints 落札された目標点を取得 (0 = 未落札)
func (b *Coinche) GetContractPoints() int { return b.contractPoints }

// SetContractPoints 目標点設定 (テスト用)
func (b *Coinche) SetContractPoints(points int) { b.contractPoints = points }

// GetDouble 倍率の状態を取得
func (b *Coinche) GetDouble() CoincheDouble { return b.double }

// SetDouble 倍率状態設定 (テスト用)
func (b *Coinche) SetDouble(d CoincheDouble) { b.double = d }

// GetMultiplier 現在の得点倍率 (1 / 2 / 4)
func (b *Coinche) GetMultiplier() int { return CoincheMultiplier(b.double) }

// GetMakerTeam メイカーチーム取得
func (b *Coinche) GetMakerTeam() int { return b.makerTeam }

// SetMakerTeam メイカーチーム設定 (テスト用)
func (b *Coinche) SetMakerTeam(team int) { b.makerTeam = team }

// GetMakerPlayerIdx メイカープレイヤー取得
func (b *Coinche) GetMakerPlayerIdx() int { return b.makerPlayerIdx }

// SetBeloteHolderIdx Belote/Rebelote (K+Q トランプ) 所持者を設定する (テスト用)。
// 通常は dealRemainder 経由で detectBeloteHolder が呼ばれて埋まる。
func (b *Coinche) SetBeloteHolderIdx(idx int) { b.beloteHolderIdx = idx }

// GetTeamScore チームスコア取得
func (b *Coinche) GetTeamScore(team int) int {
	if team < 0 || team >= CoincheTeamCnt {
		return 0
	}
	return b.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (b *Coinche) SetTeamScore(team, score int) {
	if team >= 0 && team < CoincheTeamCnt {
		b.teamScores[team] = score
	}
}

// GetRoundPoints 当ラウンドのチーム別カード点数取得
func (b *Coinche) GetRoundPoints(team int) int {
	if team < 0 || team >= CoincheTeamCnt {
		return 0
	}
	return b.roundPoints[team]
}

// SetRoundPoints 当ラウンドのチーム別カード点数設定 (テスト用)
func (b *Coinche) SetRoundPoints(team, points int) {
	if team < 0 || team >= CoincheTeamCnt {
		return
	}
	b.roundPoints[team] = points
}

// GetRoundBeloteBonus 当ラウンドの Belote/Rebelote ボーナス取得
func (b *Coinche) GetRoundBeloteBonus(team int) int {
	if team < 0 || team >= CoincheTeamCnt {
		return 0
	}
	return b.roundBeloteBonus[team]
}

// IsHumanTurn 現在の手番が人間かどうか
func (b *Coinche) IsHumanTurn() bool {
	return isHumanTurn(b.players, b.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (b *Coinche) IsHumanBidTurn() bool {
	return isHumanTurn(b.players, b.bidPlayerIdx)
}

// GetConfig 設定取得
func (b *Coinche) GetConfig() CoincheConfig { return b.config }

// SetConfig 設定変更
func (b *Coinche) SetConfig(cfg CoincheConfig) { b.config = cfg }

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (b *Coinche) CardRankPublic(card *Card) int { return b.cardRank(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (b *Coinche) CardPointsPublic(card *Card) int { return coincheCardPoints(card, b.trumpSuit) }

// --- Ranking + scoring helpers ---

// coincheTrumpRank トランプスートのカードランク (高 = 強)
// J=8, 9=7, A=6, 10=5, K=4, Q=3, 8=2, 7=1
func coincheTrumpRank(value int) int {
	switch value {
	case 11:
		return 8
	case 9:
		return 7
	case 1:
		return 6
	case 10:
		return 5
	case 13:
		return 4
	case 12:
		return 3
	case 8:
		return 2
	case 7:
		return 1
	}
	return 0
}

// coincheNonTrumpRank 非トランプスートのカードランク (高 = 強)
// A=8, 10=7, K=6, Q=5, J=4, 9=3, 8=2, 7=1
func coincheNonTrumpRank(value int) int {
	switch value {
	case 1:
		return 8
	case 10:
		return 7
	case 13:
		return 6
	case 12:
		return 5
	case 11:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	case 7:
		return 1
	}
	return 0
}

// coincheCardPoints トランプスートを踏まえたカード点数を返す
// 切り札: J=20, 9=14, A=11, 10=10, K=4, Q=3, 8=0, 7=0
// 非切り札: A=11, 10=10, K=4, Q=3, J=2, 9=0, 8=0, 7=0
func coincheCardPoints(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11:
			return 20
		case 9:
			return 14
		case 1:
			return 11
		case 10:
			return 10
		case 13:
			return 4
		case 12:
			return 3
		}
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	}
	return 0
}

// cardRank トリック比較用ランクを返す (高い = 強い)
// 切り札スート: 200 + trumpRank, 非切り札: 100 + nonTrumpRank
func (b *Coinche) cardRank(card *Card) int {
	if card.GetDesign() == b.trumpSuit {
		return 200 + coincheTrumpRank(card.GetValue())
	}
	return 100 + coincheNonTrumpRank(card.GetValue())
}

// --- Trick play helpers ---

// startPlayPhase プレイフェーズを開始する
func (b *Coinche) startPlayPhase() {
	b.trickNumber = 1
	b.currentTrick = nil
	b.leadPlayerIdx = (b.dealerIdx + 1) % CoinchePlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
	b.phase = CoinchePhasePlay
}

// playCard カードをプレイする共通処理
func (b *Coinche) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	b.maybeDeclareBeloteRebelote(playerIdx, card)
	b.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(b.players, playerIdx), cardStr(card)), []*Card{card})

	if len(b.currentTrick) == CoinchePlayerCnt {
		b.phase = CoinchePhaseTrickEnd
	} else {
		b.currentPlayerIdx = (b.currentPlayerIdx + 1) % CoinchePlayerCnt
	}
}

// maybeDeclareBeloteRebelote K+Q トランプの宣言を自動処理する
func (b *Coinche) maybeDeclareBeloteRebelote(playerIdx int, card *Card) {
	if !b.config.EnableBeloteRebelote {
		return
	}
	if playerIdx != b.beloteHolderIdx {
		return
	}
	if card.GetDesign() != b.trumpSuit {
		return
	}
	switch card.GetValue() {
	case 13:
		b.beloteKingPlayed = true
	case 12:
		b.beloteQueenPlayed = true
	default:
		return
	}
	if b.beloteKingPlayed && b.beloteQueenPlayed && !b.beloteDeclared {
		team := b.players[playerIdx].GetTeam()
		b.roundBeloteBonus[team] += CoincheRebeloteBonus
		b.beloteDeclared = true
		b.appendLog(playerIdx, "belote_rebelote",
			fmt.Sprintf("%s declares Belote/Rebelote (+%d)",
				playerName(b.players, playerIdx), CoincheRebeloteBonus), nil)
	}
}

// validatePlay カードのプレイが Coinche のルールに従っているか検証する
// コワンシュの義務:
//
//  1. フォロースート可能ならフォロースート (リードスートのカードがある限り)
//  2. フォロースート不可かつトランプがある場合は必ずトランプを出す (obligation à couper)
//  3. トランプを出す場合、既出のトランプより強いトランプがあるなら必ずオーバートランプする (obligation à monter)
//  4. リードがトランプの場合、必ずフォロー (＋オーバートランプ可能ならする)
func (b *Coinche) validatePlay(playerIdx int, card *Card) error {
	if len(b.currentTrick) == 0 {
		return nil
	}
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()

	hasLead := b.playerHasSuit(player, leadSuit)
	cardSuit := card.GetDesign()

	if leadSuit == b.trumpSuit {
		// リードがトランプ: トランプを必ず出す。出すなら可能な限りオーバートランプ。
		if hasLead {
			if cardSuit != b.trumpSuit {
				return NewDomainError(ErrInvalidPlay, "リードスート (切り札) に従ってください")
			}
			highest := b.highestTrumpInTrick()
			canOverTrump := b.playerCanBeatTrump(player, highest)
			if canOverTrump && coincheTrumpRank(card.GetValue()) <= highest {
				return NewDomainError(ErrInvalidPlay, "オーバートランプしてください (obligation à monter)")
			}
			return nil
		}
		// リードがトランプ かつ自分はトランプを持っていない: 任意のカード可
		return nil
	}

	// リードが非トランプ
	if hasLead {
		if cardSuit != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
		return nil
	}

	// フォロースート不可
	hasTrump := b.playerHasSuit(player, b.trumpSuit)
	trickHasTrump := b.trickContainsTrump()
	partnerIdx := (playerIdx + 2) % CoinchePlayerCnt
	partnerWinning := b.partnerIsCurrentlyWinning(playerIdx, partnerIdx)

	if hasTrump && !partnerWinning {
		// トランプ義務
		if cardSuit != b.trumpSuit {
			return NewDomainError(ErrInvalidPlay, "切り札を出してください (obligation à couper)")
		}
		// オーバートランプ義務 (トリックに既に切り札が出ている場合)
		if trickHasTrump {
			highest := b.highestTrumpInTrick()
			canOverTrump := b.playerCanBeatTrump(player, highest)
			if canOverTrump && coincheTrumpRank(card.GetValue()) <= highest {
				return NewDomainError(ErrInvalidPlay, "オーバートランプしてください (obligation à monter)")
			}
		}
		return nil
	}
	// 切り札なし or パートナーが現勝者: 任意のカード可
	return nil
}

// playerHasSuit プレイヤーが特定スートを持っているか
func (b *Coinche) playerHasSuit(p *CoinchePlayer, suit int) bool {
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == suit {
			return true
		}
	}
	return false
}

// playerCanBeatTrump プレイヤーが指定ランク超えのトランプを持っているか
func (b *Coinche) playerCanBeatTrump(p *CoinchePlayer, rank int) bool {
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() == b.trumpSuit && coincheTrumpRank(c.GetValue()) > rank {
			return true
		}
	}
	return false
}

// highestTrumpInTrick 現トリック内の最強トランプランク (なければ 0)
func (b *Coinche) highestTrumpInTrick() int {
	best := 0
	for _, tc := range b.currentTrick {
		if tc.Card.GetDesign() != b.trumpSuit {
			continue
		}
		r := coincheTrumpRank(tc.Card.GetValue())
		if r > best {
			best = r
		}
	}
	return best
}

// trickContainsTrump 現トリックにトランプが含まれているか
func (b *Coinche) trickContainsTrump() bool {
	for _, tc := range b.currentTrick {
		if tc.Card.GetDesign() == b.trumpSuit {
			return true
		}
	}
	return false
}

// partnerIsCurrentlyWinning 現トリックでパートナーが現勝者か
func (b *Coinche) partnerIsCurrentlyWinning(playerIdx, partnerIdx int) bool {
	if len(b.currentTrick) == 0 {
		return false
	}
	winnerIdx := b.currentLeader()
	return winnerIdx == partnerIdx
}

// currentLeader 現在のトリック先頭時点での仮勝者を返す
func (b *Coinche) currentLeader() int {
	if len(b.currentTrick) == 0 {
		return -1
	}
	winner := b.currentTrick[0].PlayerIdx
	winnerRank := b.cardRank(b.currentTrick[0].Card)
	winnerSuit := b.currentTrick[0].Card.GetDesign()
	for _, tc := range b.currentTrick[1:] {
		suit := tc.Card.GetDesign()
		rank := b.cardRank(tc.Card)
		if suit == b.trumpSuit && winnerSuit != b.trumpSuit {
			winner = tc.PlayerIdx
			winnerRank = rank
			winnerSuit = suit
			continue
		}
		if suit == winnerSuit && rank > winnerRank {
			winner = tc.PlayerIdx
			winnerRank = rank
		}
	}
	return winner
}

// trickWinner トリックの勝者を決定する
func (b *Coinche) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	return b.currentLeader()
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (b *Coinche) GetValidPlayIndices(playerIdx int) []int {
	return b.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (b *Coinche) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(b.players[playerIdx], func(c *Card) bool { return b.validatePlay(playerIdx, c) == nil })
}

// --- Game end + bookkeeping ---

func (b *Coinche) checkGameEnd() {
	for ti := range CoincheTeamCnt {
		if b.teamScores[ti] >= b.config.TargetScore {
			b.gameEndFlag = true
			b.phase = CoinchePhaseGameEnd
			if b.teamScores[0] >= b.teamScores[1] {
				b.winnerTeam = 0
			} else {
				b.winnerTeam = 1
			}
			b.appendLog(-1, "game_end",
				fmt.Sprintf("Team %d wins the game!", b.winnerTeam), nil)
			return
		}
	}
}

// sortAllHands 全プレイヤーの手札をソートする (スート → ランク)
func (b *Coinche) sortAllHands() {
	for _, p := range b.players {
		coincheSortHand(p, b)
	}
}

func coincheSortHand(p *CoinchePlayer, b *Coinche) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		si := ci.GetDesign()
		sj := cj.GetDesign()
		if si != sj {
			return si < sj
		}
		// Strongest card first within each suit — matches the manual's display
		// example and lets the CPU's "play valid[0]" easy-lead actually pick a
		// strong card instead of the weakest.
		return b.cardRank(ci) > b.cardRank(cj)
	})
}

// --- Hints ---

// GetHint 現フェーズのヒントを返す (人間プレイヤー視点)
func (b *Coinche) GetHint() *CoincheHint {
	humanIdx := findHumanIdx(b.players)
	if humanIdx < 0 {
		return nil
	}
	switch b.phase {
	case CoinchePhaseBid:
		if b.bidPlayerIdx != humanIdx {
			return nil
		}
		points, suit, ok := b.cpuSelectBid(humanIdx)
		if !ok {
			return &CoincheHint{Reason: "pass_recommended"}
		}
		return &CoincheHint{Bid: &points, Suit: &suit, Reason: "strategic_bid"}
	case CoinchePhaseDouble:
		if b.currentPlayerIdx != humanIdx {
			return nil
		}
		// **倍率の助言は立場で変わる。** 守備側には倍化、宣言側には再倍化。
		if b.players[humanIdx].GetTeam() == b.makerTeam {
			if b.double == CoincheDoubleCoinche && b.cpuWantsSurcoinche(humanIdx) {
				return &CoincheHint{Reason: "surcoinche_recommended"}
			}
			return &CoincheHint{Reason: "decline_double"}
		}
		if b.double == CoincheDoubleNone && b.cpuWantsCoinche(humanIdx) {
			return &CoincheHint{Reason: "coinche_recommended"}
		}
		return &CoincheHint{Reason: "decline_double"}
	case CoinchePhasePlay:
		if b.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := b.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := b.cpuPlayChoose(humanIdx, valid)
		return &CoincheHint{CardIndex: &idx, Reason: b.playHintReason(idx)}
	}
	return nil
}

func (b *Coinche) playHintReason(chosenIdx int) string {
	humanIdx := findHumanIdx(b.players)
	if humanIdx < 0 {
		return ""
	}
	card := b.players[humanIdx].GetCard(chosenIdx)
	if len(b.currentTrick) == 0 {
		if card.GetDesign() == b.trumpSuit {
			return "lead_trump"
		}
		return "lead_strong"
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if card.GetDesign() == b.trumpSuit {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- CPU AI ---

// cpuBestSuit は playerIdx の手札が最も強くなる切り札スートを返す。
func (b *Coinche) cpuBestSuit(playerIdx int) int {
	bestSuit, bestScore := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if score := b.evalHandForTrump(playerIdx, suit); score > bestScore {
			bestScore, bestSuit = score, suit
		}
	}
	return bestSuit
}

// cpuSelectBid は CPU の宣言を返す (ok=false ならパス)。
//
// **手札の評価値を目標点に換算する。** 評価値がそのまま「何点取れそうか」の
// 目安になるので、そこから宣言できる最小契約以上の値を選ぶ。上回れない
// ときはパスするしかない。
func (b *Coinche) cpuSelectBid(playerIdx int) (points, suit int, ok bool) {
	suit = b.cpuBestSuit(playerIdx)
	score := b.evalHandForTrump(playerIdx, suit)

	threshold := 30
	switch b.config.CpuDifficulty {
	case CoincheCpuDifficultyEasy:
		threshold = 34
	case CoincheCpuDifficultyHard:
		threshold = 26
	}
	if score < threshold {
		return 0, 0, false
	}

	// 評価値 threshold で 80 点、以降 4 ごとに 10 点上積み。
	want := CoincheContractPoints[0] + (score-threshold)/4*10
	for _, pts := range CoincheContractPoints {
		if pts > b.contractPoints && pts >= want {
			return pts, suit, true
		}
	}
	// 望む点が既に出ている契約以下なら降りる。
	return 0, 0, false
}

// cpuWantsCoinche は守備側 CPU が倍化するかを返す。
//
// **高い契約ほど倒しやすい。** 相手が 140 以上を宣言していて、自分の手札が
// その切り札に対して強ければ倍にして獲りに行く。
func (b *Coinche) cpuWantsCoinche(playerIdx int) bool {
	if b.contractPoints < 140 {
		return false
	}
	return b.evalHandForTrump(playerIdx, b.trumpSuit) >= 26
}

// cpuWantsSurcoinche は宣言側 CPU が再倍化するかを返す。
func (b *Coinche) cpuWantsSurcoinche(playerIdx int) bool {
	return b.evalHandForTrump(playerIdx, b.trumpSuit) >= 38
}

// evalHandForTrump 仮定したトランプスートに対する手札評価値を返す
// (高い = 強い: トランプ J/9/A、長いトランプ列、外スートの A をボーナス)
func (b *Coinche) evalHandForTrump(playerIdx, trumpSuit int) int {
	p := b.players[playerIdx]
	score := 0
	trumpCount := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() == trumpSuit {
			trumpCount++
			switch c.GetValue() {
			case 11:
				score += 14 // J
			case 9:
				score += 10
			case 1:
				score += 7 // A
			case 10:
				score += 5
			case 13:
				score += 3
			case 12:
				score += 2
			}
			continue
		}
		switch c.GetValue() {
		case 1:
			score += 4 // 外スート A
		case 10:
			score += 2
		}
	}
	if trumpCount >= 4 {
		score += 6
	} else if trumpCount == 3 {
		score += 3
	}
	return score
}

// cpuSelectPlayCard CPUがプレイするカードを選ぶ
func (b *Coinche) cpuSelectPlayCard(playerIdx int) int {
	valid := b.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	switch b.config.CpuDifficulty {
	case CoincheCpuDifficultyEasy:
		return valid[rand.Intn(len(valid))]
	default:
		return b.cpuPlayChoose(playerIdx, valid)
	}
}

// cpuPlayChoose 標準ヒューリスティック:
//   - リード時: 強いトランプ (J/9) または高得点の非トランプ A/10 を優先
//   - フォロー時: 勝てるなら最弱の勝てるカード、勝てないなら最低点のカードを捨てる
func (b *Coinche) cpuPlayChoose(playerIdx int, valid []int) int {
	player := b.players[playerIdx]
	if len(b.currentTrick) == 0 {
		// リード: 強い切り札 (J/9) で相手の切り札を引き出すか、外スートの A/10 を切り出す。
		best := valid[0]
		bestScore := -1
		for _, idx := range valid {
			c := player.GetCard(idx)
			s := 0
			if c.GetDesign() == b.trumpSuit {
				switch c.GetValue() {
				case 11: // Trump J — strongest; lead it to flush opponents' trumps.
					s = 25
				case 9: // Trump 9 — second-strongest; also useful as a lead.
					s = 15
				}
			} else {
				switch c.GetValue() {
				case 1:
					s = 30
				case 10:
					s = 20
				case 13:
					s = 5
				}
			}
			if s > bestScore {
				bestScore = s
				best = idx
			}
		}
		return best
	}

	// フォロー時
	winnerIdx := b.currentLeader()
	partnerIdx := (playerIdx + 2) % CoinchePlayerCnt
	partnerWinning := winnerIdx == partnerIdx

	if partnerWinning {
		// パートナーが勝者: 最も価値が高い (= 点数高い) カードを出す
		best := valid[0]
		bestPts := -1
		for _, idx := range valid {
			pts := coincheCardPoints(player.GetCard(idx), b.trumpSuit)
			if pts > bestPts {
				bestPts = pts
				best = idx
			}
		}
		return best
	}

	// 勝てるカードがあれば最弱の勝てるカードを出す
	winnable := -1
	winnableRank := 9999
	for _, idx := range valid {
		c := player.GetCard(idx)
		if b.cardWouldWinTrick(c) {
			r := b.cardRank(c)
			if r < winnableRank {
				winnableRank = r
				winnable = idx
			}
		}
	}
	if winnable >= 0 {
		return winnable
	}

	// 勝てない: 最低点のカードを捨てる
	worst := valid[0]
	worstPts := 9999
	for _, idx := range valid {
		pts := coincheCardPoints(player.GetCard(idx), b.trumpSuit)
		if pts < worstPts {
			worstPts = pts
			worst = idx
		}
	}
	return worst
}

// cardWouldWinTrick 指定カードを今出した場合に現状の勝者を上回るか
func (b *Coinche) cardWouldWinTrick(c *Card) bool {
	if len(b.currentTrick) == 0 {
		return true
	}
	winIdx := b.currentLeader()
	var winCard *Card
	for _, tc := range b.currentTrick {
		if tc.PlayerIdx == winIdx {
			winCard = tc.Card
			break
		}
	}
	if winCard == nil {
		return true
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	cSuit := c.GetDesign()
	wSuit := winCard.GetDesign()

	if cSuit == b.trumpSuit && wSuit != b.trumpSuit {
		return true
	}
	if cSuit == wSuit {
		return b.cardRank(c) > b.cardRank(winCard)
	}
	// 現勝者がトランプの場合: 非トランプは勝てない
	if wSuit == b.trumpSuit {
		return false
	}
	// 同じ非トランプ・リードスート同士は cSuit==wSuit で扱い済み
	// c がリードスートで wSuit が非リードスートの非トランプ (起こり得ない)
	if cSuit == leadSuit {
		return b.cardRank(c) > b.cardRank(winCard)
	}
	return false
}

// --- JSON ---

// coincheJSON Coinche の JSON 表現
type coincheJSON struct {
	TrumpCards       *TrumpCards      `json:"tc"`
	Players          []*CoinchePlayer `json:"pl"`
	Config           CoincheConfig    `json:"cfg"`
	Phase            CoinchePhase     `json:"ph"`
	RoundNumber      int              `json:"rn"`
	TrickNumber      int              `json:"tn"`
	CurrentPlayerIdx int              `json:"cp"`
	CurrentTrick     []*TrickCard     `json:"ct"`
	DealerIdx        int              `json:"di"`
	TrumpSuit        int              `json:"ts"`
	// 契約は「何点取るか」そのもの。落とすと復元後に成功条件が消える。
	ContractPoints int `json:"cpt"`
	// 倍率は精算の係数。落とすと復元後に ×1 へ戻り、コワンシュ後の
	// ラウンドが静かに安くなる。
	Double            CoincheDouble       `json:"db"`
	MakerTeam         int                 `json:"mt"`
	MakerPlayerIdx    int                 `json:"mp"`
	TeamScores        [CoincheTeamCnt]int `json:"sc"`
	RoundPoints       [CoincheTeamCnt]int `json:"rp"`
	RoundBeloteBonus  [CoincheTeamCnt]int `json:"rb"`
	BeloteHolderIdx   int                 `json:"bh"`
	BeloteKingPlayed  bool                `json:"bk"`
	BeloteQueenPlayed bool                `json:"bq"`
	BeloteDeclared    bool                `json:"bd"`
	LastTrickWinner   int                 `json:"lw"`
	LeadPlayerIdx     int                 `json:"li"`
	BidPlayerIdx      int                 `json:"bi"`
	BidPassCount      int                 `json:"bp"`
	GameEndFlag       bool                `json:"ge"`
	WinnerTeam        int                 `json:"wt"`
	ActionLog         []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Coinche) MarshalJSON() ([]byte, error) {
	return json.Marshal(coincheJSON{
		TrumpCards:        b.trumpCards,
		Players:           b.players,
		Config:            b.config,
		Phase:             b.phase,
		RoundNumber:       b.roundNumber,
		TrickNumber:       b.trickNumber,
		CurrentPlayerIdx:  b.currentPlayerIdx,
		CurrentTrick:      b.currentTrick,
		DealerIdx:         b.dealerIdx,
		TrumpSuit:         b.trumpSuit,
		ContractPoints:    b.contractPoints,
		Double:            b.double,
		MakerTeam:         b.makerTeam,
		MakerPlayerIdx:    b.makerPlayerIdx,
		TeamScores:        b.teamScores,
		RoundPoints:       b.roundPoints,
		RoundBeloteBonus:  b.roundBeloteBonus,
		BeloteHolderIdx:   b.beloteHolderIdx,
		BeloteKingPlayed:  b.beloteKingPlayed,
		BeloteQueenPlayed: b.beloteQueenPlayed,
		BeloteDeclared:    b.beloteDeclared,
		LastTrickWinner:   b.lastTrickWinner,
		LeadPlayerIdx:     b.leadPlayerIdx,
		BidPlayerIdx:      b.bidPlayerIdx,
		BidPassCount:      b.bidPassCount,
		GameEndFlag:       b.gameEndFlag,
		WinnerTeam:        b.winnerTeam,
		ActionLog:         b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Coinche) UnmarshalJSON(data []byte) error {
	var j coincheJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	b.trumpCards = j.TrumpCards
	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.roundNumber = j.RoundNumber
	b.trickNumber = j.TrickNumber
	b.currentPlayerIdx = j.CurrentPlayerIdx
	b.currentTrick = j.CurrentTrick
	b.dealerIdx = j.DealerIdx
	b.trumpSuit = j.TrumpSuit
	b.contractPoints = j.ContractPoints
	b.double = j.Double
	b.makerTeam = j.MakerTeam
	b.makerPlayerIdx = j.MakerPlayerIdx
	b.teamScores = j.TeamScores
	b.roundPoints = j.RoundPoints
	b.roundBeloteBonus = j.RoundBeloteBonus
	b.beloteHolderIdx = j.BeloteHolderIdx
	b.beloteKingPlayed = j.BeloteKingPlayed
	b.beloteQueenPlayed = j.BeloteQueenPlayed
	b.beloteDeclared = j.BeloteDeclared
	b.lastTrickWinner = j.LastTrickWinner
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.bidPlayerIdx = j.BidPlayerIdx
	b.bidPassCount = j.BidPassCount
	b.gameEndFlag = j.GameEndFlag
	b.winnerTeam = j.WinnerTeam
	b.actionLog = j.ActionLog
	return nil
}
