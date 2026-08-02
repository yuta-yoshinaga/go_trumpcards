//go:build !js || !wasm || casino

// Package domain ドッペルコップ (Doppelkopf) のドメインモデル。
//
// Doppelkopf はドイツで最も広くプレイされるトリックテイキングゲームの一つ。
// 各カードを 2 枚ずつ含む 48 枚デッキ (9,10,J,Q,K,A × 4 スート × 2) を 4 人で争う。
// クラブの Q (♣Q) 2 枚を持つ 2 人が秘密裏に Re チームを組み、240 点のカード
// ポイントを Kontra チームと奪い合う。♥10 (Dulle) が最強牌となる独特の切り札
// 体系を持つ。
//
// 切り札 (強い順): ♥10(Dulle) > Q♣>Q♠>Q♥>Q♦ > J♣>J♠>J♥>J♦ > A♦>10♦>K♦>9♦
// フェイル札 (各スート, 強い順): A > 10 > K > 9 (♥ は 10 が切り札のため A>K>9)
// カードポイント: A=11, 10=10, K=4, Q=3, J=2, 9=0 (合計 240)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// DoppelkopfPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const DoppelkopfPlayerCnt = 4

// DoppelkopfHandSize 各プレイヤーへ配る手札枚数 (48 / 4)
const DoppelkopfHandSize = 12

// DoppelkopfTrickCount 1 ラウンドのトリック数
const DoppelkopfTrickCount = 12

// DoppelkopfTotalPoints 1 ラウンドのカードポイント合計
const DoppelkopfTotalPoints = 240

// DoppelkopfPhase ゲームフェーズ
type DoppelkopfPhase int

// Doppelkopf のフェーズ定数
const (
	// DoppelkopfPhasePlay トリックプレイフェーズ
	DoppelkopfPhasePlay DoppelkopfPhase = 0
	// DoppelkopfPhaseTrickEnd トリック終了フェーズ (解決済み・次トリック待ち)
	DoppelkopfPhaseTrickEnd DoppelkopfPhase = 1
	// DoppelkopfPhaseRoundEnd ラウンド終了フェーズ
	DoppelkopfPhaseRoundEnd DoppelkopfPhase = 2
	// DoppelkopfPhaseGameEnd ゲーム終了フェーズ
	DoppelkopfPhaseGameEnd DoppelkopfPhase = 3
)

// DoppelkopfHint ヒント情報
type DoppelkopfHint struct {
	CardIndices []int  // 推奨カードインデックス
	Reason      string // ヒント理由キー
}

// Doppelkopf ドッペルコップのゲームクラス
type Doppelkopf struct {
	trumpCards       *TrumpCards
	players          []*DoppelkopfPlayer
	config           DoppelkopfConfig
	phase            DoppelkopfPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	reTeam           [DoppelkopfPlayerCnt]bool // Re チームのメンバー
	soloRe           bool                      // 1 人が ♣Q を 2 枚持つソロ Re
	teamsRevealed    bool                      // ラウンド終了時にチームを公開したか
	reAnnounced      bool                      // Re 宣言済みか
	kontraAnnounced  bool                      // Kontra 宣言済みか
	roundRePts       int                       // 直近ラウンドの Re チーム得点
	roundReWon       bool                      // 直近ラウンドで Re が勝ったか
	roundGamePts     int                       // 直近ラウンドのゲームポイント (倍率込み)
	gameEndFlag      bool
	winnerIdx        int // ゲーム勝者 (-1 = 未確定)
	actionLog        []*ActionLogEntry
}

// NewDoppelkopf コンストラクタ
func NewDoppelkopf(trumpCards *TrumpCards, players []*DoppelkopfPlayer, config DoppelkopfConfig) *Doppelkopf {
	return &Doppelkopf{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
	}
}

// NewDefaultDoppelkopf 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultDoppelkopf() *Doppelkopf {
	cfg := DefaultDoppelkopfConfig()
	players := make([]*DoppelkopfPlayer, DoppelkopfPlayerCnt)
	players[0] = NewDoppelkopfPlayer(true, cfg.StartChips)
	for i := 1; i < DoppelkopfPlayerCnt; i++ {
		players[i] = NewDoppelkopfPlayer(false, cfg.StartChips)
	}
	return NewDoppelkopf(NewTrumpCardsPinochle(), players, cfg)
}

// Reset ゲーム初期化: チップを開始値へ戻し最初のラウンドを開始する。
func (g *Doppelkopf) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	for _, p := range g.players {
		p.SetChips(g.config.StartChips)
	}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Doppelkopf) NextRound() {
	if g.phase != DoppelkopfPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % DoppelkopfPlayerCnt
	g.startRound()
}

// startRound 手札を配り、Re チームを割り当ててプレイフェーズを開始する。
func (g *Doppelkopf) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.reTeam = [DoppelkopfPlayerCnt]bool{}
	g.soloRe = false
	g.teamsRevealed = false
	g.reAnnounced = false
	g.kontraAnnounced = false
	g.roundRePts = 0
	g.roundReWon = false
	g.roundGamePts = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	dealAllCards(g.trumpCards, g.players)
	g.assignTeams()
	g.sortAllHands()

	g.leadPlayerIdx = (g.dealerIdx + 1) % DoppelkopfPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = DoppelkopfPhasePlay
}

// assignTeams ♣Q の保持状況から Re チームを決定する。2 人が各 1 枚なら 2 対 2、
// 1 人が 2 枚持つ場合はソロ Re (その 1 人 対 残り 3 人)。
func (g *Doppelkopf) assignTeams() {
	for i, p := range g.players {
		cnt := 0
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetDesign() == CardDesignClover && c.GetValue() == 12 {
				cnt++
			}
		}
		if cnt >= 1 {
			g.reTeam[i] = true
		}
		if cnt == 2 {
			g.soloRe = true
		}
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Doppelkopf) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != DoppelkopfPhasePlay {
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

// PlayerAnnounce 人間プレイヤーが自チームの宣言 (Re または Kontra) を行い倍率を上げる。
// 第 1 トリック中のみ可能。
func (g *Doppelkopf) PlayerAnnounce() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != DoppelkopfPhasePlay {
		return ErrWrongPhase
	}
	human := g.findHumanIdx()
	if human < 0 {
		return ErrNotHumanTurn
	}
	if !g.canAnnounce(human) {
		return NewDomainError(ErrInvalidPlay, "宣言できる時間ではありません")
	}
	g.applyAnnounce(human)
	return nil
}

// canAnnounce playerIdx が宣言可能か (第 1 トリック中、かつ自チームが未宣言)。
func (g *Doppelkopf) canAnnounce(playerIdx int) bool {
	if g.phase != DoppelkopfPhasePlay || g.trickNumber > 1 {
		return false
	}
	if g.reTeam[playerIdx] {
		return !g.reAnnounced
	}
	return !g.kontraAnnounced
}

// applyAnnounce playerIdx の自チーム宣言を反映する。
func (g *Doppelkopf) applyAnnounce(playerIdx int) {
	if g.reTeam[playerIdx] {
		g.reAnnounced = true
		g.appendLog(playerIdx, "announce_re", fmt.Sprintf("%s announces Re", g.playerName(playerIdx)), nil)
	} else {
		g.kontraAnnounced = true
		g.appendLog(playerIdx, "announce_kontra", fmt.Sprintf("%s announces Kontra", g.playerName(playerIdx)), nil)
	}
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Doppelkopf) CpuPlay() {
	if g.gameEndFlag || g.phase != DoppelkopfPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	idx := g.currentPlayerIdx
	// CPU は得点期待値が高い局面で自チームを宣言する (簡易: 強い手札のとき)。
	if g.canAnnounce(idx) && g.cpuShouldAnnounce(idx) {
		g.applyAnnounce(idx)
	}
	player := g.players[idx]
	cardIdx := g.cpuSelectPlayCard(idx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *Doppelkopf) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == DoppelkopfPlayerCnt {
		g.phase = DoppelkopfPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % DoppelkopfPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *Doppelkopf) ResolveTrick() {
	if g.phase != DoppelkopfPhaseTrickEnd || len(g.currentTrick) != DoppelkopfPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += dkCardPoints(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pts)", g.playerName(winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	// Clear the resolved trick so a spurious second ResolveTrick call cannot
	// double-count its points (defensive — NextTrick also clears it).
	g.currentTrick = nil
	if g.trickNumber >= DoppelkopfTrickCount {
		g.phase = DoppelkopfPhaseRoundEnd
	} else {
		g.phase = DoppelkopfPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Doppelkopf) NextTrick() {
	if g.phase != DoppelkopfPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = DoppelkopfPhasePlay
}

// ScoreRound ラウンドの得点を確定し、チップ精算とゲーム終了判定を行う。
func (g *Doppelkopf) ScoreRound() {
	if g.phase != DoppelkopfPhaseRoundEnd {
		return
	}
	g.teamsRevealed = true

	rePts := g.teamPoints(true)
	kontraPts := DoppelkopfTotalPoints - rePts
	reWon := rePts >= 121 // Re は 121 点以上で勝利 (Kontra は 120 点で勝利)
	loserPts := kontraPts
	if !reWon {
		loserPts = rePts
	}
	gamePts := dkGamePoints(loserPts, g.loserTookNoTrick(reWon)) * g.announceMultiplier()

	g.roundRePts = rePts
	g.roundReWon = reWon
	g.roundGamePts = gamePts
	g.settleChips(reWon, gamePts)

	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: Re %d pts (%s, %d game pts)",
			g.roundNumber, rePts, dkOutcomeStr(reWon), gamePts), nil)

	if w := g.chipLeaderAtTarget(); w >= 0 {
		g.gameEndFlag = true
		g.winnerIdx = w
		g.phase = DoppelkopfPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", g.playerName(w)), nil)
	}
}

// announceMultiplier 宣言による倍率 (Re 宣言 ×2、Kontra 宣言 ×2、両方で ×4)。
func (g *Doppelkopf) announceMultiplier() int {
	mult := 1
	if g.reAnnounced {
		mult *= 2
	}
	if g.kontraAnnounced {
		mult *= 2
	}
	return mult
}

// settleChips チップ精算 (ゼロサム)。2 対 2 では勝者各 +gamePts、敗者各 -gamePts。
// ソロ Re では Re が ±3 倍、残り 3 人が ∓1 倍。
func (g *Doppelkopf) settleChips(reWon bool, gamePts int) {
	unit := g.config.BaseChips * gamePts
	reSign := 1
	if !reWon {
		reSign = -1
	}
	if g.soloRe {
		for i := range g.players {
			if g.reTeam[i] {
				g.players[i].AddChips(reSign * unit * 3)
			} else {
				g.players[i].AddChips(-reSign * unit)
			}
		}
		return
	}
	for i := range g.players {
		if g.reTeam[i] {
			g.players[i].AddChips(reSign * unit)
		} else {
			g.players[i].AddChips(-reSign * unit)
		}
	}
}

// --- Scoring helpers ---

// teamPoints re=true なら Re チーム、false なら Kontra チームの獲得カードポイント。
func (g *Doppelkopf) teamPoints(re bool) int {
	pts := 0
	for i := range g.players {
		if g.reTeam[i] == re {
			pts += dkTrickPoints(g.players[i].GetTricksTaken())
		}
	}
	return pts
}

// loserTookNoTrick 敗北側が 1 トリックも取れなかったか (シュバルツ判定用)。
func (g *Doppelkopf) loserTookNoTrick(reWon bool) bool {
	for i := range g.players {
		loserSide := g.reTeam[i] != reWon
		if loserSide && g.players[i].GetTrickCount() > 0 {
			return false
		}
	}
	return true
}

// chipLeaderAtTarget 目標チップに到達した最上位プレイヤーを返す (-1 = なし)。
func (g *Doppelkopf) chipLeaderAtTarget() int {
	best, bestChips := -1, g.config.TargetChips-1
	for i, p := range g.players {
		if p.GetChips() >= g.config.TargetChips && p.GetChips() > bestChips {
			best, bestChips = i, p.GetChips()
		}
	}
	return best
}

// --- Trick / play helpers ---

// validatePlay マストフォロー (リードスートに従う) を検証する。
func (g *Doppelkopf) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := dkSuitID(g.currentTrick[0].Card)
	if dkSuitID(card) != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーがスート ID (切り札含む) のカードを持っているか。
func (g *Doppelkopf) playerHasSuit(playerIdx, suitID int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if dkSuitID(p.GetCard(i)) == suitID {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。最強切り札 (Dulle 含む)、なければリード
// スートの最強札が勝つ。同強の 2 枚は先に出した方が勝つ (Dulle 同点規則)。
func (g *Doppelkopf) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := dkSuitID(g.currentTrick[0].Card)
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := dkStrength(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		// 勝負に絡むのは「切り札」または「リードスートに従った札」のみ。
		// フェイルがリードされても切り札を出せば勝てる (切り札 > フェイル)。
		if !dkIsTrump(tc.Card) && dkSuitID(tc.Card) != leadSuit {
			continue
		}
		// 厳密に強い場合のみ更新 → 同強なら先出しが勝つ。
		if s := dkStrength(tc.Card); s > winnerStrength {
			winnerIdx = tc.PlayerIdx
			winnerStrength = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Doppelkopf) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Doppelkopf) sortAllHands() {
	for _, p := range g.players {
		dkSortHand(p)
	}
}

// dkSortHand 手札を切り札を先頭に、スートごと強い順にソートする。
func dkSortHand(p *DoppelkopfPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		si, sj := dkSuitID(cards[i]), dkSuitID(cards[j])
		if si != sj {
			return si < sj
		}
		return dkStrength(cards[i]) > dkStrength(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Doppelkopf) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する。
func (g *Doppelkopf) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Doppelkopf) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopStrength 現在のトリック勝者 winnerIdx の札の強さを返す。防御的に、
// 勝者がトリック内に見つからない場合は極小値を返す (負インデックスでのパニック回避)。
func (g *Doppelkopf) trickTopStrength(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return dkStrength(g.currentTrick[idx].Card)
}

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)。
func (g *Doppelkopf) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// --- Card classification ---

// dkIsTrump 切り札 (ダイヤ全札, 全 Q, 全 J, ♥10=Dulle) か。
func dkIsTrump(card *Card) bool {
	if card.GetDesign() == CardDesignDiamond || card.GetValue() == 11 || card.GetValue() == 12 {
		return true
	}
	return card.GetDesign() == CardDesignHeart && card.GetValue() == 10
}

// dkTrumpSuit 切り札を表すスート ID。
const dkTrumpSuit = 0

// dkSuitID トリック上のスート ID。切り札は共通 ID、フェイル札はスート定数。
func dkSuitID(card *Card) int {
	if dkIsTrump(card) {
		return dkTrumpSuit
	}
	return card.GetDesign()
}

// dkStrength トリックでの強さ。すべての切り札はフェイル札より強い。
//
//	♥10 > Q♣>Q♠>Q♥>Q♦ > J♣>J♠>J♥>J♦ > A♦>10♦>K♦>9♦ > (フェイル) A>10>K>9
func dkStrength(card *Card) int {
	const trumpBase = 100
	v := card.GetValue()
	if card.GetDesign() == CardDesignHeart && v == 10 { // Dulle
		return trumpBase + 50
	}
	if v == 12 { // Queen
		return trumpBase + 40 + dkTrumpSuitOrder(card.GetDesign())
	}
	if v == 11 { // Jack
		return trumpBase + 30 + dkTrumpSuitOrder(card.GetDesign())
	}
	if card.GetDesign() == CardDesignDiamond { // diamond trump (A/10/K/9)
		return trumpBase + dkFailRank(v)
	}
	return dkFailRank(v)
}

// dkTrumpSuitOrder Q/J の切り札内スート順位 (♣>♠>♥>♦)。
func dkTrumpSuitOrder(design int) int {
	switch design {
	case CardDesignClover:
		return 3
	case CardDesignSpade:
		return 2
	case CardDesignHeart:
		return 1
	default: // Diamond
		return 0
	}
}

// dkFailRank フェイル/ダイヤ非絵札の強さ順位。A>10>K>9。
func dkFailRank(value int) int {
	switch value {
	case 1: // Ace
		return 3
	case 10:
		return 2
	case 13: // King
		return 1
	default: // 9
		return 0
	}
}

// dkCardPoints カードポイント。A=11,10=10,K=4,Q=3,J=2,9=0。
func dkCardPoints(value int) int {
	switch value {
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
	default:
		return 0
	}
}

// dkTrickPoints 取得トリック群の合計カードポイント。
func dkTrickPoints(tricks [][]*Card) int {
	pts := 0
	for _, t := range tricks {
		for _, c := range t {
			pts += dkCardPoints(c.GetValue())
		}
	}
	return pts
}

// dkGamePoints ゲームポイント。勝利 +1、敗北側 <90/<60/<30 で各 +1、無トリックで +1。
func dkGamePoints(loserPoints int, loserNoTrick bool) int {
	pts := 1
	if loserPoints < 90 {
		pts++
	}
	if loserPoints < 60 {
		pts++
	}
	if loserPoints < 30 {
		pts++
	}
	if loserNoTrick {
		pts++
	}
	return pts
}

// dkOutcomeStr 勝敗の表示文字列。
func dkOutcomeStr(reWon bool) string {
	if reWon {
		return "Re wins"
	}
	return "Kontra wins"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Doppelkopf) GetPhase() DoppelkopfPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Doppelkopf) SetPhase(phase DoppelkopfPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Doppelkopf) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Doppelkopf) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Doppelkopf) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Doppelkopf) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Doppelkopf) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Doppelkopf) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Doppelkopf) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Doppelkopf) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Doppelkopf) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Doppelkopf) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Doppelkopf) GetDealerIdx() int { return g.dealerIdx }

// IsRe playerIdx が Re チームか
func (g *Doppelkopf) IsRe(playerIdx int) bool {
	if playerIdx < 0 || playerIdx >= len(g.reTeam) {
		return false
	}
	return g.reTeam[playerIdx]
}

// SetReTeam Re チーム設定 (テスト用)
func (g *Doppelkopf) SetReTeam(team [DoppelkopfPlayerCnt]bool) { g.reTeam = team }

// IsSoloRe ソロ Re かどうか
func (g *Doppelkopf) IsSoloRe() bool { return g.soloRe }

// AreTeamsRevealed チームが公開済みか
func (g *Doppelkopf) AreTeamsRevealed() bool { return g.teamsRevealed }

// IsReAnnounced Re 宣言済みか
func (g *Doppelkopf) IsReAnnounced() bool { return g.reAnnounced }

// IsKontraAnnounced Kontra 宣言済みか
func (g *Doppelkopf) IsKontraAnnounced() bool { return g.kontraAnnounced }

// CanHumanAnnounce 人間プレイヤーが今宣言できるか
func (g *Doppelkopf) CanHumanAnnounce() bool {
	human := g.findHumanIdx()
	return human >= 0 && g.canAnnounce(human)
}

// GetRoundRePoints 直近ラウンドの Re チーム得点取得
func (g *Doppelkopf) GetRoundRePoints() int { return g.roundRePts }

// GetRoundReWon 直近ラウンドで Re が勝ったか取得
func (g *Doppelkopf) GetRoundReWon() bool { return g.roundReWon }

// GetRoundGamePoints 直近ラウンドのゲームポイント取得
func (g *Doppelkopf) GetRoundGamePoints() int { return g.roundGamePts }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Doppelkopf) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1=未確定)
func (g *Doppelkopf) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Doppelkopf) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Doppelkopf) GetPlayer(i int) *DoppelkopfPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間かどうか。
func (g *Doppelkopf) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Doppelkopf) GetConfig() DoppelkopfConfig { return g.config }

// SetConfig 設定変更
func (g *Doppelkopf) SetConfig(cfg DoppelkopfConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Doppelkopf) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Doppelkopf) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != DoppelkopfPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨プレイを返す。
func (g *Doppelkopf) GetHint() *DoppelkopfHint {
	human := g.findHumanIdx()
	if human < 0 || g.phase != DoppelkopfPhasePlay || g.currentPlayerIdx != human {
		return nil
	}
	valid := g.getValidPlayIndices(human)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuPlaySmart(human, valid)
	return &DoppelkopfHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Doppelkopf) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := dkSuitID(g.currentTrick[0].Card)
	if dkSuitID(card) != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := g.trickTopStrength(winnerIdx)
	if dkStrength(card) > topStrength {
		return "follow_win"
	}
	return "follow_duck"
}

// --- CPU AI ---

// cpuShouldAnnounce CPU が宣言すべきか (簡易: 切り札を多く持つとき)。
func (g *Doppelkopf) cpuShouldAnnounce(playerIdx int) bool {
	if g.config.CpuDifficulty == DoppelkopfCpuDifficultyEasy {
		return false
	}
	player := g.players[playerIdx]
	trumpCnt := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		if dkIsTrump(player.GetCard(i)) {
			trumpCnt++
		}
	}
	return trumpCnt >= 8
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Doppelkopf) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == DoppelkopfCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点とチーム関係を意識した戦略プレイ。
func (g *Doppelkopf) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	if len(g.currentTrick) == 0 {
		return g.minBy(player, valid, func(c *Card) int {
			return dkCardPoints(c.GetValue())*100 + dkStrength(c)
		})
	}

	leadSuit := dkSuitID(g.currentTrick[0].Card)
	winnerIdx := g.trickWinner()
	topStrength := g.trickTopStrength(winnerIdx)
	partnerWinning := g.reTeam[playerIdx] == g.reTeam[winnerIdx]
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += dkCardPoints(tc.Card.GetValue())
	}

	var follows []int
	for _, idx := range valid {
		if dkSuitID(player.GetCard(idx)) == leadSuit {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		if partnerWinning {
			return g.maxBy(player, valid, func(c *Card) int {
				if dkIsTrump(c) {
					return -dkStrength(c)
				}
				return dkCardPoints(c.GetValue())*100 - dkStrength(c)
			})
		}
		return g.minBy(player, valid, func(c *Card) int {
			return dkCardPoints(c.GetValue())*100 + dkStrength(c)
		})
	}

	winners := dkFilter(follows, func(idx int) bool {
		return dkStrength(player.GetCard(idx)) > topStrength
	})

	if partnerWinning {
		nonWinners := dkFilter(follows, func(idx int) bool {
			return dkStrength(player.GetCard(idx)) < topStrength
		})
		if len(nonWinners) > 0 {
			return g.maxBy(player, nonWinners, func(c *Card) int {
				return dkCardPoints(c.GetValue())*100 - dkStrength(c)
			})
		}
		return g.minBy(player, follows, func(c *Card) int { return dkStrength(c) })
	}

	if trickPts > 0 && len(winners) > 0 {
		return g.minBy(player, winners, func(c *Card) int { return dkStrength(c) })
	}
	return g.minBy(player, follows, func(c *Card) int {
		return dkCardPoints(c.GetValue())*100 + dkStrength(c)
	})
}

// minBy score が最小となるインデックスを返す。
func (g *Doppelkopf) minBy(player *DoppelkopfPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxBy score が最大となるインデックスを返す。
func (g *Doppelkopf) maxBy(player *DoppelkopfPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// dkFilter 述語を満たすインデックスを抽出する。
func dkFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- JSON ---

// doppelkopfJSON is the JSON wire format for Doppelkopf.
type doppelkopfJSON struct {
	TrumpCards       *TrumpCards               `json:"tc"`
	Players          []*DoppelkopfPlayer       `json:"ps"`
	Config           DoppelkopfConfig          `json:"cf"`
	Phase            DoppelkopfPhase           `json:"ph"`
	RoundNumber      int                       `json:"rn"`
	TrickNumber      int                       `json:"tn"`
	CurrentPlayerIdx int                       `json:"ci"`
	CurrentTrick     []*TrickCard              `json:"ct"`
	LeadPlayerIdx    int                       `json:"li"`
	DealerIdx        int                       `json:"di"`
	ReTeam           [DoppelkopfPlayerCnt]bool `json:"rt"`
	SoloRe           bool                      `json:"sr"`
	TeamsRevealed    bool                      `json:"tv"`
	ReAnnounced      bool                      `json:"ra"`
	KontraAnnounced  bool                      `json:"ka"`
	RoundRePts       int                       `json:"rp"`
	RoundReWon       bool                      `json:"rw"`
	RoundGamePts     int                       `json:"rg"`
	GameEndFlag      bool                      `json:"ge"`
	WinnerIdx        int                       `json:"wi"`
	ActionLog        []*ActionLogEntry         `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Doppelkopf) MarshalJSON() ([]byte, error) {
	return json.Marshal(doppelkopfJSON{
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
		ReTeam:           g.reTeam,
		SoloRe:           g.soloRe,
		TeamsRevealed:    g.teamsRevealed,
		ReAnnounced:      g.reAnnounced,
		KontraAnnounced:  g.kontraAnnounced,
		RoundRePts:       g.roundRePts,
		RoundReWon:       g.roundReWon,
		RoundGamePts:     g.roundGamePts,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		ActionLog:        g.actionLog,
	})
}

// doppelkopfMaxSliceLen caps slice sizes during deserialisation.
const doppelkopfMaxSliceLen = 1000

// errDoppelkopfOversized is the single sentinel error for oversized input arrays.
var errDoppelkopfOversized = errors.New("doppelkopf: input array exceeds maximum allowed size")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Doppelkopf) UnmarshalJSON(data []byte) error {
	var j doppelkopfJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > doppelkopfMaxSliceLen || len(j.CurrentTrick) > doppelkopfMaxSliceLen ||
		len(j.ActionLog) > doppelkopfMaxSliceLen {
		return errDoppelkopfOversized
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsPinochle()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*DoppelkopfPlayer, 0)
	}
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
	g.reTeam = j.ReTeam
	g.soloRe = j.SoloRe
	g.teamsRevealed = j.TeamsRevealed
	g.reAnnounced = j.ReAnnounced
	g.kontraAnnounced = j.KontraAnnounced
	g.roundRePts = j.RoundRePts
	g.roundReWon = j.RoundReWon
	g.roundGamePts = j.RoundGamePts
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
