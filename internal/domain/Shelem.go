//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ShelemPhase シェレムのゲームフェーズ
type ShelemPhase int

// Shelem のフェーズ定数
const (
	// ShelemPhaseBid 点数を競り上げる
	ShelemPhaseBid ShelemPhase = iota
	// ShelemPhaseDiscard 落札者がウィドウを取り込み、4 枚捨てて切り札を決める
	ShelemPhaseDiscard
	// ShelemPhasePlay プレイ中
	ShelemPhasePlay
	// ShelemPhaseRoundEnd ラウンド終了
	ShelemPhaseRoundEnd
	// ShelemPhaseGameEnd ゲーム終了
	ShelemPhaseGameEnd
)

// ShelemPlayerCnt プレイヤー数（4 人固定・2 対 2）
const ShelemPlayerCnt = 4

// ShelemTeamCnt チーム数
const ShelemTeamCnt = 2

// ShelemHandSize 各プレイヤーの手札枚数
//
// **12 枚 × 4 人 + ウィドウ 4 枚 = 52 枚。** issue は「52 枚 + ウィドウ 4 枚」と
// 書いているが、それでは 56 枚になり標準デッキで配れない。
const ShelemHandSize = 12

// ShelemWidowSize ウィドウ（伏せ札）の枚数
const ShelemWidowSize = 4

// ShelemTricksPerRound 1 ラウンドのトリック数
const ShelemTricksPerRound = ShelemHandSize

// ShelemHandPoints は下で定義する 1 ラウンドのカード点の合計 (100)。

// ShelemMinBid 競りの最低入札。**過半（55 点）から。**
//
// 100 点しか出回らない卓で 100 点から競り始めると、そもそも競り上げる余地が
// 無い。「相手より多く取る」と言える最小の額から始める。
const ShelemMinBid = 55

// ShelemMaxBid 競りの上限。**1 ラウンドに出回るカード点そのもの。**
//
// これを超える額は**カード点では絶対に達成できない**（プールが 100 点しか
// 無い）。CPU の見積もりが 100 を超えても、ここで頭打ちにする
// （レビュー指摘 PR #5306。Cinch も CinchMaxBid = CinchTotalPoints としている）。
const ShelemMaxBid = ShelemHandPoints

// ShelemBidStep 競り上げの刻み
const ShelemBidStep = 5

// ShelemHandPoints 1 ラウンドに出回るカード点の合計
const ShelemHandPoints = 100

// ShelemValue Shelem（全トリック独占）の成否で動く点数
const ShelemValue = 200

// shelemMaxSliceLen caps slice sizes during deserialisation.
const shelemMaxSliceLen = 1000

// Shelem シェレム ゲームクラス。
//
// イランで遊ばれるコントラクトブリッジ系の競り型トリックテイキング。4 人 2 対 2
// （向かい合う席が味方）、52 枚を **12 枚ずつ + ウィドウ 4 枚**。
//
// **issue の枚数は成立しない。** 「標準 52 枚＋ウィドウ 4 枚」と書かれているが
// それでは 56 枚になる。実際は 12 × 4 = 48 に伏せ札 4 枚を足して 52 枚ちょうど
// で、トリックは 13 ではなく **12**。
//
// **競るのはトリック数ではなく点数そのもの。** 55 から 5 刻みで 100 まで
// 競り上げ、落札者はウィドウ 4 枚を取り込んで 4 枚捨て、切り札を決める。
// 落札した点数をカード点で取り切れれば加点、届かなければ同額を失点する。
//
// **点になるのは 3 ランクだけ。** A と 10 が 10 点、5 が 5 点で、1 ラウンドの
// 合計はちょうど 100 点。**上限が 100 なのはそのため**——プールが 100 点しか
// 無い卓で 100 を超える契約は、カード点では絶対に達成できない。下限の 55 は
// 「相手より多く取る」と言える最小の額。
//
// **Shelem** は点数の代わりに宣言する全トリック独占。取り切れば +200、
// 1 つでも落とせば -200 で、通常の契約より大きく振れる。
type Shelem struct {
	trumpCards *TrumpCards
	players    []*ShelemPlayer
	config     ShelemConfig

	phase       ShelemPhase
	roundNumber int
	trickNumber int
	trumpSuit   int
	// widow は伏せられた 4 枚。落札者だけが見て取り込む。
	widow []*Card
	// declarerIdx は落札者 (-1: 未決定)、contract は落札した点数。
	declarerIdx int
	contract    int
	// shelemBid は Shelem 宣言で落札したか。
	shelemBid bool

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	bidPlayerIdx     int

	scores      [ShelemTeamCnt]int
	roundPoints [ShelemTeamCnt]int

	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewShelem コンストラクタ
func NewShelem(trumpCards *TrumpCards, players []*ShelemPlayer, config ShelemConfig) *Shelem {
	return &Shelem{trumpCards: trumpCards, players: players, config: config, declarerIdx: -1, winnerTeam: -1}
}

// NewDefaultShelem 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultShelem() *Shelem {
	players := make([]*ShelemPlayer, 0, ShelemPlayerCnt)
	for i := range ShelemPlayerCnt {
		players = append(players, NewShelemPlayer(i == 0))
	}
	return NewShelem(NewTrumpCards(0), players, DefaultShelemConfig())
}

// ShelemTeamOf 席のチーム番号。**向かい合う席が味方。**
func ShelemTeamOf(playerIdx int) int { return playerIdx % ShelemTeamCnt }

// ShelemCardPoints 札の点数。**A と 10 が 10 点、5 が 5 点。他は 0。**
//
// 1 ラウンドの合計はちょうど 100 点になる（10×4 + 10×4 + 5×4）。
func ShelemCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1, 10:
		return 10
	case 5:
		return 5
	}
	return 0
}

// Reset ゲーム全体を初期化する
func (s *Shelem) Reset() {
	s.roundNumber = 1
	s.dealerIdx = 0
	s.gameEndFlag = false
	s.winnerTeam = -1
	s.scores = [ShelemTeamCnt]int{}
	s.actionLog = nil
	for _, p := range s.players {
		p.ResetGame()
	}
	s.dealRound()
}

// dealRound 12 枚ずつ配り、4 枚をウィドウにして競りから始める
func (s *Shelem) dealRound() {
	s.phase = ShelemPhaseBid
	s.trickNumber = 0
	s.currentTrick = nil
	s.trumpSuit = 0
	s.declarerIdx = -1
	s.contract = 0
	s.shelemBid = false
	s.roundPoints = [ShelemTeamCnt]int{}
	for _, p := range s.players {
		p.ResetRound()
	}

	s.trumpCards = NewTrumpCards(0)
	s.trumpCards.Shuffle()
	for range ShelemHandSize {
		for i := range ShelemPlayerCnt {
			idx := (s.dealerIdx + 1 + i) % ShelemPlayerCnt
			if c := s.trumpCards.DrawCard(); c != nil {
				s.players[idx].AddCard(c)
			}
		}
	}
	// **残った 4 枚がウィドウ。** 12×4 + 4 = 52 でちょうど配り切る。
	s.widow = make([]*Card, 0, ShelemWidowSize)
	for range ShelemWidowSize {
		if c := s.trumpCards.DrawCard(); c != nil {
			s.widow = append(s.widow, c)
		}
	}
	s.sortAllHands()
	s.bidPlayerIdx = (s.dealerIdx + 1) % ShelemPlayerCnt
	s.currentPlayerIdx = s.bidPlayerIdx
	s.leadPlayerIdx = s.bidPlayerIdx
	s.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", s.roundNumber), nil)
}

// sortAllHands 手札をスート・ランク順に並べ替える
func (s *Shelem) sortAllHands() {
	for _, p := range s.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return shelemRank(ci) < shelemRank(cj)
		})
	}
}

// shelemRank 札の強さ。A が最強、以下 K,Q,J,10..2。
func shelemRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// --- 競り ---

// PlayerBid 人間プレイヤーが点数で入札する
func (s *Shelem) PlayerBid(bid int) error {
	if err := s.guardBid(0); err != nil {
		return err
	}
	if bid < ShelemMinBid || bid > ShelemMaxBid {
		return fmt.Errorf("bids run from %d to %d", ShelemMinBid, ShelemMaxBid)
	}
	if bid%ShelemBidStep != 0 {
		return fmt.Errorf("bids go up in steps of %d", ShelemBidStep)
	}
	if bid <= s.contract {
		return fmt.Errorf("bid must beat %d", s.contract)
	}
	s.acceptBid(0, bid, false)
	return nil
}

// PlayerBidShelem 人間プレイヤーが Shelem（全トリック独占）を宣言する
func (s *Shelem) PlayerBidShelem() error {
	if err := s.guardBid(0); err != nil {
		return err
	}
	// **Shelem はどんな点数入札にも勝つ。** 全トリックを取る宣言なので。
	s.acceptBid(0, ShelemMaxBid, true)
	return nil
}

// PlayerPass 人間プレイヤーが競りを降りる
func (s *Shelem) PlayerPass() error {
	if err := s.guardBid(0); err != nil {
		return err
	}
	// **全員が降りると契約が決まらない。** 誰も入札していない最後の 1 人は降りられない。
	if s.contract == 0 && s.activeBidders() == 1 {
		return errors.New("the last bidder standing must bid")
	}
	s.acceptPass(0)
	return nil
}

// guardBid 競りで操作できる状態かを確かめる
func (s *Shelem) guardBid(idx int) error {
	if s.gameEndFlag {
		return errors.New("game has ended")
	}
	if s.phase != ShelemPhaseBid {
		return errors.New("not the bidding phase")
	}
	if s.bidPlayerIdx != idx {
		return errors.New("not your turn to bid")
	}
	if s.players[idx].GetPassed() {
		return errors.New("you have already passed")
	}
	return nil
}

// activeBidders まだ降りていない人数
func (s *Shelem) activeBidders() int {
	n := 0
	for _, p := range s.players {
		if !p.GetPassed() {
			n++
		}
	}
	return n
}

// CpuBid 手番の CPU が入札するか降りるかを決める
func (s *Shelem) CpuBid() {
	if s.gameEndFlag || s.phase != ShelemPhaseBid || s.bidPlayerIdx == 0 {
		return
	}
	idx := s.bidPlayerIdx
	if bid, ok := s.cpuBidChoice(idx); ok {
		s.acceptBid(idx, bid, false)
		return
	}
	if s.contract == 0 && s.activeBidders() == 1 {
		// 誰も入札しないまま最後の 1 人。最低額で引き受けるしかない。
		s.acceptBid(idx, ShelemMinBid, false)
		return
	}
	s.acceptPass(idx)
}

// acceptBid 入札を記録して次の席へ回す
func (s *Shelem) acceptBid(idx, bid int, shelem bool) {
	s.players[idx].SetBid(bid)
	s.players[idx].SetDeclaredShelem(shelem)
	s.contract, s.declarerIdx, s.shelemBid = bid, idx, shelem
	if shelem {
		s.appendLog(idx, "shelem", "Shelem（全トリック独占）を宣言", nil)
	} else {
		s.appendLog(idx, "bid", fmt.Sprintf("%d 点で入札", bid), nil)
	}
	s.advanceBidding()
}

// acceptPass 降りたことを記録して次の席へ回す
func (s *Shelem) acceptPass(idx int) {
	s.players[idx].SetPassed(true)
	s.appendLog(idx, "pass", "競りを降りた", nil)
	s.advanceBidding()
}

// advanceBidding 次の入札者へ回し、決着していれば競りを閉じる
func (s *Shelem) advanceBidding() {
	// **入札がある状態で生存者が 1 人になったら決着。**
	if s.contract > 0 && s.activeBidders() <= 1 {
		s.closeBidding()
		return
	}
	// Shelem は誰も上回れないので即決着。
	if s.shelemBid {
		s.closeBidding()
		return
	}
	for i := 1; i <= ShelemPlayerCnt; i++ {
		next := (s.bidPlayerIdx + i) % ShelemPlayerCnt
		if !s.players[next].GetPassed() {
			s.bidPlayerIdx = next
			return
		}
	}
	s.closeBidding()
}

// closeBidding 落札者にウィドウを渡し、捨て札フェーズに入る
func (s *Shelem) closeBidding() {
	s.phase = ShelemPhaseDiscard
	// **ウィドウは落札者の手に入る。** 16 枚から 4 枚捨てて 12 枚に戻す。
	for _, c := range s.widow {
		s.players[s.declarerIdx].AddCard(c)
	}
	s.widow = nil
	s.sortAllHands()
	s.appendLog(s.declarerIdx, "widow",
		fmt.Sprintf("ウィドウ %d 枚を取り込んだ（契約 %d 点）", ShelemWidowSize, s.contract), nil)

	// CPU が落札したなら、その場で捨てて切り札も決める。
	if s.declarerIdx != 0 {
		s.cpuDiscardAndTrump()
	}
}

// --- ウィドウ交換と切り札 ---

// PlayerDiscard 人間の落札者が 4 枚捨てて切り札を決める
func (s *Shelem) PlayerDiscard(indices []int, suit int) error {
	if s.gameEndFlag {
		return errors.New("game has ended")
	}
	if s.phase != ShelemPhaseDiscard {
		return errors.New("not the discard phase")
	}
	if s.declarerIdx != 0 {
		return errors.New("only the declarer discards")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", suit)
	}
	if len(indices) != ShelemWidowSize {
		return fmt.Errorf("discard exactly %d cards", ShelemWidowSize)
	}
	p := s.players[0]
	seen := map[int]bool{}
	for _, i := range indices {
		if i < 0 || i >= p.GetCardsSize() {
			return fmt.Errorf("invalid card index: %d", i)
		}
		if seen[i] {
			return fmt.Errorf("duplicate card index: %d", i)
		}
		seen[i] = true
	}
	// **後ろから消す。** 前から消すと残りのインデックスがずれる。
	sorted := make([]int, 0, len(indices))
	sorted = append(sorted, indices...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	discarded := make([]*Card, 0, len(sorted))
	for _, i := range sorted {
		discarded = append(discarded, p.GetCard(i))
		p.RemoveCard(i)
	}
	s.creditDiscard(discarded)
	s.acceptTrump(suit)
	return nil
}

// creditDiscard は捨て札のカード点を落札者チームに加える。
//
// カード点は 52 枚全体で ShelemHandPoints (100) 点あり、そのうち 4 枚はウィドウに
// 伏せられている。落札者がウィドウを取り込んで 4 枚捨てたあと、その札の点を誰にも
// 加算しないと、点札が捨て札に入った分だけラウンド合計が 100 を割る (#5795)。
// ウィドウ・キティ系の通例どおり、捨てた点は落札者チームのものとして扱う。
func (s *Shelem) creditDiscard(cards []*Card) {
	if s.declarerIdx < 0 {
		return
	}
	pts := 0
	for _, c := range cards {
		pts += ShelemCardPoints(c)
	}
	s.roundPoints[ShelemTeamOf(s.declarerIdx)] += pts
}

// cpuDiscardAndTrump CPU の落札者が点にならない札を捨て、長いスートを切り札にする
func (s *Shelem) cpuDiscardAndTrump() {
	suit := s.longestSuit(s.declarerIdx)
	p := s.players[s.declarerIdx]
	discarded := make([]*Card, 0, ShelemWidowSize)
	// **切り札でなく、点にもならない札から捨てる。**
	for p.GetCardsSize() > ShelemHandSize {
		worst, worstScore := 0, 1<<30
		for i := range p.GetCardsSize() {
			c := p.GetCard(i)
			score := shelemRank(c) + ShelemCardPoints(c)*10
			if c.GetDesign() == suit {
				score += 100
			}
			if score < worstScore {
				worst, worstScore = i, score
			}
		}
		discarded = append(discarded, p.GetCard(worst))
		p.RemoveCard(worst)
	}
	s.creditDiscard(discarded)
	s.acceptTrump(suit)
}

// acceptTrump 切り札を確定させてプレイに入る
func (s *Shelem) acceptTrump(suit int) {
	s.trumpSuit = suit
	s.phase = ShelemPhasePlay
	s.leadPlayerIdx = s.declarerIdx
	s.currentPlayerIdx = s.declarerIdx
	s.sortAllHands()
	s.appendLog(s.declarerIdx, "trump", fmt.Sprintf("切り札を %d に決めた", suit), nil)
}

// longestSuit いちばん枚数の多いスート
func (s *Shelem) longestSuit(idx int) int {
	p := s.players[idx]
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		counts[p.GetCard(i).GetDesign()]++
	}
	best, bestN := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestN {
			best, bestN = suit, counts[suit]
		}
	}
	return best
}

// cpuBidChoice CPU の入札。手札の点札と長いスートの厚みで測る。
func (s *Shelem) cpuBidChoice(idx int) (int, bool) {
	p := s.players[idx]
	suit := s.longestSuit(idx)
	strength := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		strength += ShelemCardPoints(c)
		if c.GetDesign() == suit {
			strength += 3
		}
		if c.GetValue() == 1 {
			strength += 5
		}
	}
	bid := ShelemMinBid + (strength/10)*ShelemBidStep
	// **プールを超える契約は達成できない。** 見積もりが上振れても頭打ちにする。
	if bid > ShelemMaxBid {
		bid = ShelemMaxBid
	}
	if bid <= s.contract {
		return 0, false
	}
	return bid, true
}

// --- プレイ ---

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (s *Shelem) PlayerPlay(cardIndex int) error {
	if s.gameEndFlag {
		return errors.New("game has ended")
	}
	if s.phase != ShelemPhasePlay {
		return errors.New("not the play phase")
	}
	if s.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return s.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (s *Shelem) CpuPlay() {
	if s.gameEndFlag || s.phase != ShelemPhasePlay || s.currentPlayerIdx == 0 {
		return
	}
	_ = s.play(s.currentPlayerIdx, s.chooseCpuCard(s.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (s *Shelem) play(playerIdx, cardIndex int) error {
	p := s.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !s.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	s.currentTrick = append(s.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	s.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(s.currentTrick) < ShelemPlayerCnt {
		s.currentPlayerIdx = (playerIdx + 1) % ShelemPlayerCnt
		return nil
	}
	s.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか
func (s *Shelem) canPlay(playerIdx int, card *Card) bool {
	if len(s.currentTrick) == 0 {
		return true
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := s.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (s *Shelem) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(s.players) {
		return nil
	}
	p := s.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if s.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、カード点を勝者チームに加える
func (s *Shelem) resolveTrick() {
	winner := s.trickWinner()
	cards := make([]*Card, 0, len(s.currentTrick))
	pts := 0
	for _, tc := range s.currentTrick {
		cards = append(cards, tc.Card)
		pts += ShelemCardPoints(tc.Card)
	}
	s.players[winner].AddTrick(cards)
	s.roundPoints[ShelemTeamOf(winner)] += pts

	s.trickNumber++
	s.currentTrick = nil
	s.leadPlayerIdx = winner
	s.currentPlayerIdx = winner

	// **Shelem は 1 つ落とした時点で失敗が確定する。** 残りを打たせない。
	if s.shelemBid && ShelemTeamOf(winner) != ShelemTeamOf(s.declarerIdx) {
		s.finishRound()
		return
	}
	if s.trickNumber >= ShelemTricksPerRound {
		s.finishRound()
	}
}

// trickWinner 現在のトリックの勝者
func (s *Shelem) trickWinner() int {
	if len(s.currentTrick) == 0 {
		return s.leadPlayerIdx
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	bestIdx, best := s.currentTrick[0].PlayerIdx, s.currentTrick[0].Card
	for _, tc := range s.currentTrick[1:] {
		if s.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats challenger が currentBest に勝つか
func (s *Shelem) beats(challenger, currentBest *Card, leadSuit int) bool {
	cTrump := challenger.GetDesign() == s.trumpSuit
	bTrump := currentBest.GetDesign() == s.trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if challenger.GetDesign() != currentBest.GetDesign() {
		return challenger.GetDesign() == leadSuit
	}
	return shelemRank(challenger) > shelemRank(currentBest)
}

// finishRound 契約の当否で得点を確定させる
func (s *Shelem) finishRound() {
	declTeam := ShelemTeamOf(s.declarerIdx)
	other := 1 - declTeam

	if s.shelemBid {
		// **Shelem は全トリック取れたかどうかだけ。** カード点は見ない。
		took := s.teamTricks(declTeam)
		if took == ShelemTricksPerRound {
			s.scores[declTeam] += ShelemValue
			s.appendLog(-1, "score", fmt.Sprintf("Shelem 成功。チーム%d に +%d", declTeam, ShelemValue), nil)
		} else {
			s.scores[declTeam] -= ShelemValue
			s.appendLog(-1, "score",
				fmt.Sprintf("Shelem 失敗（%d/%d トリック）。チーム%d に -%d",
					took, ShelemTricksPerRound, declTeam, ShelemValue), nil)
		}
	} else {
		got := s.roundPoints[declTeam]
		if got >= s.contract {
			s.scores[declTeam] += s.contract
			s.appendLog(-1, "score",
				fmt.Sprintf("契約 %d 点を %d 点で達成。チーム%d に +%d", s.contract, got, declTeam, s.contract), nil)
		} else {
			s.scores[declTeam] -= s.contract
			s.appendLog(-1, "score",
				fmt.Sprintf("契約 %d 点に %d 点で未達。チーム%d に -%d", s.contract, got, declTeam, s.contract), nil)
		}
		// **相手チームは取ったカード点をそのまま得る。**
		s.scores[other] += s.roundPoints[other]
	}

	if s.scores[0] >= s.config.Target || s.scores[1] >= s.config.Target {
		s.finishGame()
		return
	}
	s.phase = ShelemPhaseRoundEnd
}

// teamTricks チームの獲得トリック数
func (s *Shelem) teamTricks(team int) int {
	n := 0
	for i, p := range s.players {
		if ShelemTeamOf(i) == team {
			n += p.GetTrickCount()
		}
	}
	return n
}

// TeamTricks チームの獲得トリック数
func (s *Shelem) TeamTricks(team int) int {
	if team < 0 || team >= ShelemTeamCnt {
		return 0
	}
	return s.teamTricks(team)
}

// NextRound 次のラウンドを開始する
func (s *Shelem) NextRound() {
	if s.gameEndFlag || s.phase != ShelemPhaseRoundEnd {
		return
	}
	s.roundNumber++
	s.dealerIdx = (s.dealerIdx + 1) % ShelemPlayerCnt
	s.dealRound()
}

// finishGame 規定点に達したチームの勝ち
func (s *Shelem) finishGame() {
	s.phase = ShelemPhaseGameEnd
	s.gameEndFlag = true
	switch {
	case s.scores[0] > s.scores[1]:
		s.winnerTeam = 0
	case s.scores[1] > s.scores[0]:
		s.winnerTeam = 1
	default:
		s.winnerTeam = -1
	}
	s.appendLog(-1, "result", fmt.Sprintf("最終得点 %d - %d", s.scores[0], s.scores[1]), nil)
}

// chooseCpuCard CPU の手。味方が勝っていれば点を乗せ、そうでなければ取りに行く。
func (s *Shelem) chooseCpuCard(playerIdx int) int {
	valid := s.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := s.players[playerIdx]

	if len(s.currentTrick) == 0 {
		bestIdx, bestRank := valid[0], shelemRank(p.GetCard(valid[0]))
		for _, i := range valid[1:] {
			if r := shelemRank(p.GetCard(i)); r > bestRank {
				bestIdx, bestRank = i, r
			}
		}
		return bestIdx
	}

	if s.partnerIsWinning(playerIdx) {
		// **味方が勝っているなら点札を乗せる。** カード点がそのまま契約の達否になる。
		bestIdx, bestPts := -1, -1
		for _, i := range valid {
			c := p.GetCard(i)
			if s.wouldWin(c) {
				continue
			}
			if pts := ShelemCardPoints(c); pts > bestPts {
				bestIdx, bestPts = i, pts
			}
		}
		if bestIdx >= 0 {
			return bestIdx
		}
	}
	if idx, ok := s.pickCheapestWinner(p, valid); ok {
		return idx
	}
	// 取れないなら点にならない札を捨てる。
	bestIdx, bestPts := valid[0], ShelemCardPoints(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if pts := ShelemCardPoints(p.GetCard(i)); pts < bestPts {
			bestIdx, bestPts = i, pts
		}
	}
	return bestIdx
}

// partnerIsWinning 現時点で味方がトリックを取っているか
func (s *Shelem) partnerIsWinning(playerIdx int) bool {
	if len(s.currentTrick) == 0 {
		return false
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	best, bestIdx := s.currentTrick[0].Card, s.currentTrick[0].PlayerIdx
	for _, tc := range s.currentTrick[1:] {
		if s.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx != playerIdx && ShelemTeamOf(bestIdx) == ShelemTeamOf(playerIdx)
}

// pickCheapestWinner トリックを取れる札のうち一番弱いもの
func (s *Shelem) pickCheapestWinner(p *ShelemPlayer, valid []int) (int, bool) {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !s.wouldWin(c) {
			continue
		}
		if r := shelemRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx, bestIdx >= 0
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (s *Shelem) wouldWin(c *Card) bool {
	if c == nil || len(s.currentTrick) == 0 {
		return true
	}
	leadSuit := s.currentTrick[0].Card.GetDesign()
	best := s.currentTrick[0].Card
	for _, tc := range s.currentTrick[1:] {
		if s.beats(tc.Card, best, leadSuit) {
			best = tc.Card
		}
	}
	return s.beats(c, best, leadSuit)
}

// ShelemHint ヒント情報
type ShelemHint struct {
	// CardIndex 推奨する手札のインデックス（競り・捨て札中は nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
	// Value 推奨する入札点数（それ以外は 0）
	Value int
	// Suit 推奨する切り札スート（それ以外は 0）
	Suit int
}

// GetHint 人間プレイヤーへの推奨手を返す
func (s *Shelem) GetHint() *ShelemHint {
	if s.gameEndFlag {
		return nil
	}
	if s.phase == ShelemPhaseBid && s.bidPlayerIdx == 0 && !s.players[0].GetPassed() {
		if bid, ok := s.cpuBidChoice(0); ok {
			return &ShelemHint{Reason: "shelemBid", Value: bid}
		}
		return &ShelemHint{Reason: "shelemPass"}
	}
	if s.phase == ShelemPhaseDiscard && s.declarerIdx == 0 {
		return &ShelemHint{Reason: "shelemDiscard", Suit: s.longestSuit(0)}
	}
	if !s.IsHumanTurn() || s.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := s.chooseCpuCard(0)
	reason := "shelemWinTrick"
	if s.partnerIsWinning(0) {
		reason = "shelemFeedPartner"
	}
	return &ShelemHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (s *Shelem) GetPhase() ShelemPhase { return s.phase }

// GetConfig 現在の設定
func (s *Shelem) GetConfig() ShelemConfig { return s.config }

// SetConfig 設定を差し替える
func (s *Shelem) SetConfig(c ShelemConfig) { s.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (s *Shelem) GetRoundNumber() int { return s.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (s *Shelem) GetTrickNumber() int { return s.trickNumber }

// GetTrumpSuit 切り札のスート（未決定は 0）
func (s *Shelem) GetTrumpSuit() int { return s.trumpSuit }

// GetDeclarerIdx 落札者 (-1: 未決定)
func (s *Shelem) GetDeclarerIdx() int { return s.declarerIdx }

// GetContract 落札した点数（未決定は 0）
func (s *Shelem) GetContract() int { return s.contract }

// GetShelemBid Shelem 宣言で落札したか
func (s *Shelem) GetShelemBid() bool { return s.shelemBid }

// GetWidowSize 伏せられているウィドウの枚数（落札後は 0）
func (s *Shelem) GetWidowSize() int { return len(s.widow) }

// GetScore チームの累計得点
func (s *Shelem) GetScore(team int) int {
	if team < 0 || team >= ShelemTeamCnt {
		return 0
	}
	return s.scores[team]
}

// SetScoreForTestUse チームの累計得点を設定する（復元・テスト用）
func (s *Shelem) SetScoreForTestUse(team, n int) {
	if team >= 0 && team < ShelemTeamCnt {
		s.scores[team] = n
	}
}

// GetRoundPoints チームの現ラウンドのカード点
func (s *Shelem) GetRoundPoints(team int) int {
	if team < 0 || team >= ShelemTeamCnt {
		return 0
	}
	return s.roundPoints[team]
}

// GetCurrentTrick 現在のトリック
func (s *Shelem) GetCurrentTrick() []*TrickCard { return s.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (s *Shelem) GetCurrentPlayerIdx() int { return s.currentPlayerIdx }

// GetBidPlayerIdx 競りの手番
func (s *Shelem) GetBidPlayerIdx() int { return s.bidPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (s *Shelem) GetLeadPlayerIdx() int { return s.leadPlayerIdx }

// GetDealerIdx ディーラー
func (s *Shelem) GetDealerIdx() int { return s.dealerIdx }

// GetPlayerCnt プレイヤー数
func (s *Shelem) GetPlayerCnt() int { return len(s.players) }

// GetPlayer 指定インデックスのプレイヤー
func (s *Shelem) GetPlayer(i int) *ShelemPlayer {
	if i < 0 || i >= len(s.players) {
		return nil
	}
	return s.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (s *Shelem) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1: 未確定/同点)
func (s *Shelem) GetWinnerTeam() int { return s.winnerTeam }

// IsHumanTurn 人間の手番か
func (s *Shelem) IsHumanTurn() bool {
	return !s.gameEndFlag && s.phase == ShelemPhasePlay && s.currentPlayerIdx == 0
}

// IsHumanBidTurn 人間が入札する番か
func (s *Shelem) IsHumanBidTurn() bool {
	return !s.gameEndFlag && s.phase == ShelemPhaseBid && s.bidPlayerIdx == 0 && !s.players[0].GetPassed()
}

// IsHumanDiscardTurn 人間が捨て札と切り札を決める番か
func (s *Shelem) IsHumanDiscardTurn() bool {
	return !s.gameEndFlag && s.phase == ShelemPhaseDiscard && s.declarerIdx == 0
}

// GiveUp 投了する
func (s *Shelem) GiveUp() {
	if s.gameEndFlag {
		return
	}
	s.phase = ShelemPhaseGameEnd
	s.gameEndFlag = true
	s.winnerTeam = 1
	s.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (s *Shelem) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.appendLogAt(s.trickNumber, playerIdx, actionType, detail, cards)
}

// shelemJSON is the KV snapshot format for Shelem.
type shelemJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*ShelemPlayer    `json:"pl"`
	Config           ShelemConfig       `json:"cf"`
	Phase            ShelemPhase        `json:"ph"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	TrumpSuit        int                `json:"ts"`
	Widow            []*Card            `json:"wd"`
	DeclarerIdx      int                `json:"dc"`
	Contract         int                `json:"ct"`
	ShelemBid        bool               `json:"sb"`
	CurrentTrick     []*TrickCard       `json:"tk"`
	CurrentPlayerIdx int                `json:"cp"`
	BidPlayerIdx     int                `json:"bp"`
	LeadPlayerIdx    int                `json:"lp"`
	DealerIdx        int                `json:"di"`
	Scores           [ShelemTeamCnt]int `json:"sc"`
	RoundPoints      [ShelemTeamCnt]int `json:"rp"`
	GameEndFlag      bool               `json:"ge"`
	WinnerTeam       int                `json:"wt"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (s *Shelem) MarshalJSON() ([]byte, error) {
	return json.Marshal(&shelemJSON{
		TrumpCards:       s.trumpCards,
		Players:          s.players,
		Config:           s.config,
		Phase:            s.phase,
		RoundNumber:      s.roundNumber,
		TrickNumber:      s.trickNumber,
		TrumpSuit:        s.trumpSuit,
		Widow:            s.widow,
		DeclarerIdx:      s.declarerIdx,
		Contract:         s.contract,
		ShelemBid:        s.shelemBid,
		CurrentTrick:     s.currentTrick,
		CurrentPlayerIdx: s.currentPlayerIdx,
		BidPlayerIdx:     s.bidPlayerIdx,
		LeadPlayerIdx:    s.leadPlayerIdx,
		DealerIdx:        s.dealerIdx,
		Scores:           s.scores,
		RoundPoints:      s.roundPoints,
		GameEndFlag:      s.gameEndFlag,
		WinnerTeam:       s.winnerTeam,
		ActionLog:        s.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (s *Shelem) UnmarshalJSON(data []byte) error {
	var j shelemJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < ShelemPhaseBid || j.Phase > ShelemPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札はフェーズと整合していなければならない。** 競り・捨て札の
	// あいだはまだ 0、プレイ以降は実在するスート (#5302〜#5305 と同じ穴)。
	if j.Phase == ShelemPhaseBid || j.Phase == ShelemPhaseDiscard {
		if j.TrumpSuit != 0 {
			return fmt.Errorf("trump suit %d before it was chosen", j.TrumpSuit)
		}
	} else if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	if j.TrickNumber < 0 || j.TrickNumber > ShelemTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.Contract != 0 && (j.Contract < ShelemMinBid || j.Contract > ShelemMaxBid) {
		return fmt.Errorf("invalid contract: %d", j.Contract)
	}
	if len(j.Widow) > ShelemWidowSize {
		return fmt.Errorf("widow holds %d cards", len(j.Widow))
	}
	if len(j.ActionLog) > shelemMaxSliceLen {
		return errors.New("shelem: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > ShelemPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"bid player":     j.BidPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= ShelemPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.DeclarerIdx < -1 || j.DeclarerIdx >= ShelemPlayerCnt {
		return fmt.Errorf("invalid declarer: %d", j.DeclarerIdx)
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= ShelemTeamCnt {
		return fmt.Errorf("invalid winner team: %d", j.WinnerTeam)
	}
	if j.TrumpCards != nil {
		s.trumpCards = j.TrumpCards
	}
	if len(j.Players) == ShelemPlayerCnt {
		s.players = j.Players
	}
	s.config = j.Config
	s.phase = j.Phase
	s.roundNumber = j.RoundNumber
	s.trickNumber = j.TrickNumber
	s.trumpSuit = j.TrumpSuit
	s.widow = j.Widow
	s.declarerIdx = j.DeclarerIdx
	s.contract = j.Contract
	s.shelemBid = j.ShelemBid
	s.currentTrick = j.CurrentTrick
	s.currentPlayerIdx = j.CurrentPlayerIdx
	s.bidPlayerIdx = j.BidPlayerIdx
	s.leadPlayerIdx = j.LeadPlayerIdx
	s.dealerIdx = j.DealerIdx
	s.scores = j.Scores
	s.roundPoints = j.RoundPoints
	s.gameEndFlag = j.GameEndFlag
	s.winnerTeam = j.WinnerTeam
	s.actionLog = j.ActionLog
	return nil
}
