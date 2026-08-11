//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// TarabishPhase タラビッシュのゲームフェーズ
type TarabishPhase int

// Tarabishのフェーズ定数
const (
	// TarabishPhaseBid 表向きの札を切り札として取るかを順に選ぶ
	TarabishPhaseBid TarabishPhase = iota
	// TarabishPhasePlay プレイ中
	TarabishPhasePlay
	// TarabishPhaseRoundEnd ラウンド終了（集計済み、次ラウンド待ち）
	TarabishPhaseRoundEnd
	// TarabishPhaseGameEnd ゲーム終了
	TarabishPhaseGameEnd
)

// TarabishPlayerCnt プレイヤー数（4 人固定・2 対 2）
const TarabishPlayerCnt = 4

// TarabishTeamCnt チーム数
const TarabishTeamCnt = 2

// TarabishHandSize 各プレイヤーの手札枚数
const TarabishHandSize = 9

// TarabishFirstDealSize 切り札を決める前に配る枚数
const TarabishFirstDealSize = 6

// TarabishTricksPerRound 1 ラウンドのトリック数
const TarabishTricksPerRound = TarabishHandSize

// TarabishLastTrickBonus 最終トリックの加点
const TarabishLastTrickBonus = 10

// TarabishBellaBonus ベラ（切り札の K+Q）の申告点
const TarabishBellaBonus = 20

// TarabishRun3Bonus 3 枚のランの申告点
const TarabishRun3Bonus = 20

// TarabishRun4Bonus 4 枚以上のランの申告点
const TarabishRun4Bonus = 50

// TarabishRunMinLen ランとして認められる最短の長さ
const TarabishRunMinLen = 3

// tarabishMaxSliceLen caps slice sizes during deserialisation.
const tarabishMaxSliceLen = 1000

// Tarabish タラビッシュ ゲームクラス。
//
// カナダ・ケープブレトン島のレバノン系移民の間で育った**ジャス系**トリック
// テイキング。36 枚 (A,6,7,8,9,10,J,Q,K × 4 スート) を 4 人に 9 枚ずつ配る
// **2 対 2 のパートナー戦**で、向かい合う席が味方 (0+2 対 1+3)。
//
// 切り札の序列がこの系統の特徴:
//
//	切り札: J(Jass)=20 > 9(Menel)=14 > A=11 > 10=10 > K=4 > Q=3 > 8,7,6=0
//	非切り札: A=11 > 10=10 > K=4 > Q=3 > J=2 > 9,8,7,6=0
//
// **強さの順序と点数の順序が切り札で食い違う**のがジャス系の勘所で、切り札の
// J と 9 は点数でも強さでも A を追い越す。カード点の合計は 62 (切り札) +
// 30×3 (非切り札) = **152 点**、これに最終トリックの 10 点が乗って 162 点。
//
// 配り終えたら 1 枚を表向きにし、**そのスートを切り札として引き受けるか**を
// 順に選ぶ。全員が断ったら**親が引き受ける** (dealer is stuck)。
//
// 申告 (メルド) は配り札から自動で判定する:
//
//   - **ラン**: 同スートの連続 3 枚で 20 点、4 枚以上で 50 点
//   - **ベラ**: 切り札の K+Q で 20 点
//
// issue #5237 の仕様案に対して補ったもの:
//
//   - **ランの札順を定義していない。** 連続の判定は A,K,Q,J,10,9,8,7,6 の
//     並び（点数ではなく札の並び）で行う。切り札の J/9 が強い序列はトリックの
//     勝敗にのみ効き、ランの連続性には効かない。両者を混ぜると J と 9 が
//     飛び地になってランが成立しなくなる。
//   - **メルドは申告制でなく自動。** issue は「各プレイヤーは申告する」と書くが、
//     CPU 3 人ぶんの申告 UI を作っても選択の余地が無い（申告しない理由が無い）ため、
//     配り直後に自動で確定させている。
//
// Belote / Jass も同じ点数表を持つが、どちらも `extra3` タグ付きの別ゲームで
// 関数が非公開のため、この repo の慣習どおり点数表はゲームごとに持つ。
type Tarabish struct {
	trumpCards *TrumpCards
	players    []*TarabishPlayer
	config     TarabishConfig

	phase       TarabishPhase
	roundNumber int
	trickNumber int
	trumpSuit   int
	// upCard 切り札候補として表向きにした 1 枚
	upCard *Card
	// trumpTakerIdx 切り札を引き受けたプレイヤー (-1: 未決定)
	trumpTakerIdx int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int

	// scores はチームごとの累計点、roundPoints は現ラウンドのカード点。
	scores      [TarabishTeamCnt]int
	roundPoints [TarabishTeamCnt]int

	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewTarabish コンストラクタ
func NewTarabish(trumpCards *TrumpCards, players []*TarabishPlayer, config TarabishConfig) *Tarabish {
	return &Tarabish{trumpCards: trumpCards, players: players, config: config, trumpTakerIdx: -1, winnerTeam: -1}
}

// NewDefaultTarabish 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultTarabish() *Tarabish {
	players := make([]*TarabishPlayer, 0, TarabishPlayerCnt)
	for i := range TarabishPlayerCnt {
		players = append(players, NewTarabishPlayer(i == 0))
	}
	return NewTarabish(NewTrumpCardsShortDeck(), players, DefaultTarabishConfig())
}

// TarabishTeamOf 席のチーム番号。**向かい合う席が味方。**
func TarabishTeamOf(playerIdx int) int { return playerIdx % TarabishTeamCnt }

// Reset ゲーム全体を初期化する
func (t *Tarabish) Reset() {
	t.roundNumber = 1
	t.dealerIdx = 0
	t.gameEndFlag = false
	t.winnerTeam = -1
	t.scores = [TarabishTeamCnt]int{}
	t.actionLog = nil
	for _, p := range t.players {
		p.ResetGame()
	}
	t.dealRound()
}

// dealRound 1 ラウンド分を配り、切り札候補を表向きにして選択フェーズに入る
func (t *Tarabish) dealRound() {
	t.phase = TarabishPhaseBid
	t.trickNumber = 0
	t.currentTrick = nil
	t.roundPoints = [TarabishTeamCnt]int{}
	t.trumpTakerIdx = -1
	t.trumpSuit = 0
	for _, p := range t.players {
		p.ResetRound()
	}

	// **36 枚 (A,6-K)。** Jass も同じ構成だがあちらは extra3 タグ内の非公開関数
	// なので、untagged な共通コンストラクタを使う。
	t.trumpCards = NewTrumpCardsShortDeck()
	t.trumpCards.Shuffle()

	// **配りは 2 段構え。** 9 枚 × 4 人 = 36 枚はデッキ全部なので、先に配り切ると
	// 切り札候補としてめくる 1 枚が残らない。本場の手順どおり、まず 6 枚ずつ
	// 配って 25 枚目を表向きにし、切り札が決まってから残りを配る。
	// **その表向きの札は親の手札に入る** ので、最終的に全員 9 枚になる。
	for range TarabishFirstDealSize {
		for i := range TarabishPlayerCnt {
			idx := (t.dealerIdx + 1 + i) % TarabishPlayerCnt
			if c := t.trumpCards.DrawCard(); c != nil {
				t.players[idx].AddCard(c)
			}
		}
	}
	t.upCard = t.trumpCards.DrawCard()
	t.leadPlayerIdx = (t.dealerIdx + 1) % TarabishPlayerCnt
	t.currentPlayerIdx = t.leadPlayerIdx
	t.sortAllHands()
	t.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", t.roundNumber), nil)
}

// sortAllHands 手札をスートごと・札の並び順に整える
func (t *Tarabish) sortAllHands() {
	for _, p := range t.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return tarabishSequenceRank(ci) < tarabishSequenceRank(cj)
		})
	}
}

// TakeTrump 人間プレイヤーが表向きの札のスートを切り札として引き受ける
func (t *Tarabish) TakeTrump() error {
	if t.gameEndFlag {
		return errors.New("game has ended")
	}
	if t.phase != TarabishPhaseBid {
		return errors.New("not the bidding phase")
	}
	if t.currentPlayerIdx != 0 {
		return errors.New("not your turn to decide")
	}
	t.acceptTrump(0)
	return nil
}

// PassTrump 人間プレイヤーが見送る。以降は CPU が順に選ぶ。
func (t *Tarabish) PassTrump() error {
	if t.gameEndFlag {
		return errors.New("game has ended")
	}
	if t.phase != TarabishPhaseBid {
		return errors.New("not the bidding phase")
	}
	if t.currentPlayerIdx != 0 {
		return errors.New("not your turn to decide")
	}
	// **親は見送れない (dealer is stuck)。** 見送りを許すと誰も切り札を
	// 決めないまま手番が回り続ける。
	if t.dealerIdx == 0 {
		return errors.New("the dealer must take the trump")
	}
	t.appendLog(0, "pass", "切り札を見送った", nil)
	t.advanceBid()
	return nil
}

// CpuBid 手番の CPU が引き受けるか見送るかを決める
func (t *Tarabish) CpuBid() {
	if t.gameEndFlag || t.phase != TarabishPhaseBid || t.currentPlayerIdx == 0 {
		return
	}
	idx := t.currentPlayerIdx
	if t.cpuWantsTrump(idx) {
		t.acceptTrump(idx)
		return
	}
	t.appendLog(idx, "pass", "切り札を見送った", nil)
	t.advanceBid()
}

// acceptTrump 切り札を確定させ、メルドを数えてプレイに入る
func (t *Tarabish) acceptTrump(idx int) {
	t.trumpTakerIdx = idx
	if t.upCard != nil {
		t.trumpSuit = t.upCard.GetDesign()
	}
	t.appendLog(idx, "trump", fmt.Sprintf("切り札を引き受けた（%s）", tarabishCardLabel(t.upCard)), nil)
	t.completeDeal()
	t.countMelds()
	t.sortAllHands()
	t.phase = TarabishPhasePlay
	t.leadPlayerIdx = (t.dealerIdx + 1) % TarabishPlayerCnt
	t.currentPlayerIdx = t.leadPlayerIdx
}

// completeDeal 切り札が決まったあとに残りを配る。
//
// **表向きにした 1 枚は親の手札に入る。** 6 枚 × 4 = 24、表向き 1 枚、残り 11 枚。
// 親は表向きの 1 枚 + 2 枚、他の 3 人は 3 枚ずつで、合計 24+1+2+9 = 36 枚。
// 全員がちょうど 9 枚になる。
func (t *Tarabish) completeDeal() {
	if t.upCard != nil {
		t.players[t.dealerIdx].AddCard(t.upCard)
	}
	for i := range TarabishPlayerCnt {
		idx := (t.dealerIdx + 1 + i) % TarabishPlayerCnt
		need := TarabishHandSize - t.players[idx].GetCardsSize()
		for range need {
			if c := t.trumpCards.DrawCard(); c != nil {
				t.players[idx].AddCard(c)
			}
		}
	}
}

// tarabishCardLabel 表示用のカード名。nil でも落ちない。
func tarabishCardLabel(c *Card) string {
	if c == nil {
		return "-"
	}
	return cardStr(c)
}

// advanceBid 次の席へ選択を回す。
//
// **回ってきた先が CPU の親なら、その場で引き受けさせる (dealer is stuck)。**
// 入札は親の左隣から始まって親で終わるので、全員が見送ると最後は必ず親になる。
// ここで止めないと誰も切り札を決めないまま手番だけが回り続ける。
//
// **人間が親のときは代わりに決めない。** 選択の余地が無くても、引き受けたのが
// 自分だと分かるように操作させる。PassTrump は親のときに拒否するので、
// 押せる手は「引き受ける」だけになる。
func (t *Tarabish) advanceBid() {
	t.currentPlayerIdx = (t.currentPlayerIdx + 1) % TarabishPlayerCnt
	if t.currentPlayerIdx == t.dealerIdx && t.dealerIdx != 0 {
		t.acceptTrump(t.dealerIdx)
	}
}

// cpuWantsTrump CPU が引き受けるか。切り札候補のスートに強い札が要る。
func (t *Tarabish) cpuWantsTrump(idx int) bool {
	if t.upCard == nil {
		return true
	}
	suit := t.upCard.GetDesign()
	strength := 0
	p := t.players[idx]
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() != suit {
			continue
		}
		switch c.GetValue() {
		case 11: // J (Jass)
			strength += 4
		case 9: // Menel
			strength += 3
		case 1, 10: // A, 10
			strength += 2
		default:
			strength++
		}
	}
	return strength >= 6
}

// countMelds 配り札からラン・ベラを判定してメルド点を確定させる
func (t *Tarabish) countMelds() {
	for i, p := range t.players {
		runPts, runLen := tarabishRunPoints(p)
		bella := tarabishHasBella(p, t.trumpSuit)
		pts := runPts
		if bella {
			pts += TarabishBellaBonus
		}
		p.SetRunLen(runLen)
		p.SetHasBella(bella)
		p.SetMeldPoints(pts)
		if pts > 0 {
			t.appendLog(i, "meld", fmt.Sprintf("メルド %d 点", pts), nil)
		}
	}
}

// tarabishRunPoints 手札の最長ランの点数と長さを返す。
//
// **連続の判定は札の並び (A,K,Q,J,10,9,8,7,6) で行う。** 切り札の J/9 が強い
// のはトリックの勝敗の話で、ランの連続性には関係しない。混ぜると J と 9 が
// 飛び地になってランが崩れる。
func tarabishRunPoints(p *TarabishPlayer) (int, int) {
	bySuit := map[int][]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], tarabishSequenceRank(c))
	}

	best := 0
	for _, ranks := range bySuit {
		sort.Ints(ranks)
		run := 1
		for i := 1; i < len(ranks); i++ {
			if ranks[i] == ranks[i-1]+1 {
				run++
			} else {
				run = 1
			}
			if run > best {
				best = run
			}
		}
	}
	switch {
	case best >= 4:
		return TarabishRun4Bonus, best
	case best == TarabishRunMinLen:
		return TarabishRun3Bonus, best
	default:
		return 0, 0
	}
}

// tarabishHasBella 切り札の K と Q を両方持っているか
func tarabishHasBella(p *TarabishPlayer, trumpSuit int) bool {
	k, q := false, false
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c.GetDesign() != trumpSuit {
			continue
		}
		switch c.GetValue() {
		case 13:
			k = true
		case 12:
			q = true
		}
	}
	return k && q
}

// tarabishSequenceRank ランの連続判定に使う札の並び順。6 が最小、A が最大。
func tarabishSequenceRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return CardValueMax + 1 // A は K の上
	}
	return c.GetValue()
}

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (t *Tarabish) PlayerPlay(cardIndex int) error {
	if t.gameEndFlag {
		return errors.New("game has ended")
	}
	if t.phase != TarabishPhasePlay {
		return errors.New("not the play phase")
	}
	if t.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return t.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (t *Tarabish) CpuPlay() {
	if t.gameEndFlag || t.phase != TarabishPhasePlay || t.currentPlayerIdx == 0 {
		return
	}
	_ = t.play(t.currentPlayerIdx, t.chooseCpuCard(t.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (t *Tarabish) play(playerIdx, cardIndex int) error {
	p := t.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !t.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	t.currentTrick = append(t.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	t.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(t.currentTrick) < TarabishPlayerCnt {
		t.currentPlayerIdx = (playerIdx + 1) % TarabishPlayerCnt
		return nil
	}
	t.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか。リードなら何でも出せる。
func (t *Tarabish) canPlay(playerIdx int, card *Card) bool {
	if len(t.currentTrick) == 0 {
		return true
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := t.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (t *Tarabish) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(t.players) {
		return nil
	}
	p := t.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if t.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決し、カード点を勝者チームに加える
func (t *Tarabish) resolveTrick() {
	winner := t.trickWinner()
	cards := make([]*Card, 0, len(t.currentTrick))
	pts := 0
	for _, tc := range t.currentTrick {
		cards = append(cards, tc.Card)
		pts += TarabishCardPoints(tc.Card, t.trumpSuit)
	}
	t.players[winner].AddTrick(cards)
	t.roundPoints[TarabishTeamOf(winner)] += pts

	t.trickNumber++
	t.currentTrick = nil
	t.leadPlayerIdx = winner
	t.currentPlayerIdx = winner

	if t.trickNumber >= TarabishTricksPerRound {
		// **最終トリックに 10 点。** 152 + 10 = 162 がラウンドの総点。
		t.roundPoints[TarabishTeamOf(winner)] += TarabishLastTrickBonus
		t.appendLog(winner, "last", fmt.Sprintf("最終トリック +%d", TarabishLastTrickBonus), nil)
		t.finishRound()
	}
}

// TarabishCardPoints 札の点数。**切り札だけ序列が変わる。**
//
//	切り札: J=20, 9=14, A=11, 10=10, K=4, Q=3, 他=0
//	非切り札: A=11, 10=10, K=4, Q=3, J=2, 他=0
func TarabishCardPoints(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11: // J = Jass
			return 20
		case 9: // Menel
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

// finishRound カード点にメルド点を足してチーム得点を確定させる
func (t *Tarabish) finishRound() {
	for i, p := range t.players {
		if m := p.GetMeldPoints(); m > 0 {
			t.roundPoints[TarabishTeamOf(i)] += m
		}
	}
	for team := range TarabishTeamCnt {
		t.scores[team] += t.roundPoints[team]
		t.appendLog(-1, "score", fmt.Sprintf("チーム%d に %d 点", team, t.roundPoints[team]), nil)
	}

	if t.scores[0] >= t.config.Target || t.scores[1] >= t.config.Target {
		t.finishGame()
		return
	}
	t.phase = TarabishPhaseRoundEnd
}

// NextRound 次のラウンドを開始する
func (t *Tarabish) NextRound() {
	if t.gameEndFlag || t.phase != TarabishPhaseRoundEnd {
		return
	}
	t.roundNumber++
	t.dealerIdx = (t.dealerIdx + 1) % TarabishPlayerCnt
	t.dealRound()
}

// finishGame 目標点に達したチームの勝ち。両方超えたら高いほう。
func (t *Tarabish) finishGame() {
	t.phase = TarabishPhaseGameEnd
	t.gameEndFlag = true
	switch {
	case t.scores[0] > t.scores[1]:
		t.winnerTeam = 0
	case t.scores[1] > t.scores[0]:
		t.winnerTeam = 1
	default:
		t.winnerTeam = -1
	}
	t.appendLog(-1, "result", fmt.Sprintf("最終得点 %d - %d", t.scores[0], t.scores[1]), nil)
}

// trickWinner 現在のトリックの勝者
func (t *Tarabish) trickWinner() int {
	if len(t.currentTrick) == 0 {
		return t.leadPlayerIdx
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	bestIdx, best := t.currentTrick[0].PlayerIdx, t.currentTrick[0].Card
	for _, tc := range t.currentTrick[1:] {
		if tarabishBeats(tc.Card, best, leadSuit, t.trumpSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// tarabishBeats challenger が currentBest に勝つか
func tarabishBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cTrump := challenger.GetDesign() == trumpSuit
	bTrump := currentBest.GetDesign() == trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if cTrump {
		return tarabishTrumpRank(challenger) > tarabishTrumpRank(currentBest)
	}
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if currentBest.GetDesign() != leadSuit {
		return true
	}
	return tarabishPlainRank(challenger) > tarabishPlainRank(currentBest)
}

// isTrump その札が切り札か
func (t *Tarabish) isTrump(c *Card) bool {
	return c != nil && c.GetDesign() == t.trumpSuit
}

// tarabishTrumpRank 切り札の強さ。**J(Jass) が最強、9(Menel) が次。**
//
// **10 を明示的に扱わないと K/Q に負ける。** 素の GetValue() に落とすと
// 10 < 12 < 13 なので、点数表 (J20 > 9=14 > A11 > 10=10 > K4 > Q3) と序列が
// 食い違う。Belote.go の beloteTrumpRank も同じ理由で 10 を特別扱いしている。
func tarabishTrumpRank(c *Card) int {
	switch c.GetValue() {
	case 11: // J = Jass
		return 100
	case 9: // Menel
		return 99
	case 1: // A
		return 98
	case 10:
		return 97
	}
	return c.GetValue()
}

// tarabishPlainRank 非切り札の強さ。A が最強、以下 K,Q,J,10,...,6。
func tarabishPlainRank(c *Card) int {
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// chooseCpuCard CPU の手を選ぶ。取れるなら取り、無理なら安い札を捨てる。
func (t *Tarabish) chooseCpuCard(playerIdx int) int {
	valid := t.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := t.players[playerIdx]

	if len(t.currentTrick) == 0 {
		// リードは点の低い札から。**切り札は同点でも後回しにする** —— 点だけで
		// 選ぶと 0 点の切り札 8/7/6 が 0 点の非切り札より先に出てしまい、
		// 「温存する」という意図と裏返る。
		bestIdx := valid[0]
		bestPts, bestTrump := TarabishCardPoints(p.GetCard(bestIdx), t.trumpSuit), t.isTrump(p.GetCard(bestIdx))
		for _, i := range valid[1:] {
			c := p.GetCard(i)
			pts, isTrump := TarabishCardPoints(c, t.trumpSuit), t.isTrump(c)
			if pts < bestPts || (pts == bestPts && bestTrump && !isTrump) {
				bestIdx, bestPts, bestTrump = i, pts, isTrump
			}
		}
		return bestIdx
	}

	// 味方が勝っているなら点を乗せる。そうでなければ取りに行く。
	if t.partnerIsWinning(playerIdx) {
		bestIdx, bestPts := valid[0], -1
		for _, i := range valid {
			if pts := TarabishCardPoints(p.GetCard(i), t.trumpSuit); pts > bestPts && !t.wouldWin(p.GetCard(i)) {
				bestIdx, bestPts = i, pts
			}
		}
		if bestPts >= 0 {
			return bestIdx
		}
	}
	if idx, ok := t.pickWinning(p, valid); ok {
		return idx
	}
	bestIdx, bestPts := valid[0], TarabishCardPoints(p.GetCard(valid[0]), t.trumpSuit)
	for _, i := range valid[1:] {
		if pts := TarabishCardPoints(p.GetCard(i), t.trumpSuit); pts < bestPts {
			bestIdx, bestPts = i, pts
		}
	}
	return bestIdx
}

// partnerIsWinning 現時点で味方がトリックを取っているか
func (t *Tarabish) partnerIsWinning(playerIdx int) bool {
	if len(t.currentTrick) == 0 {
		return false
	}
	leader := t.currentTrick[0].PlayerIdx
	best, bestIdx := t.currentTrick[0].Card, leader
	leadSuit := t.currentTrick[0].Card.GetDesign()
	for _, tc := range t.currentTrick[1:] {
		if tarabishBeats(tc.Card, best, leadSuit, t.trumpSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx != playerIdx && TarabishTeamOf(bestIdx) == TarabishTeamOf(playerIdx)
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (t *Tarabish) wouldWin(c *Card) bool {
	if c == nil || len(t.currentTrick) == 0 {
		return true
	}
	leadSuit := t.currentTrick[0].Card.GetDesign()
	best := t.currentTrick[0].Card
	for _, tc := range t.currentTrick[1:] {
		if tarabishBeats(tc.Card, best, leadSuit, t.trumpSuit) {
			best = tc.Card
		}
	}
	return tarabishBeats(c, best, leadSuit, t.trumpSuit)
}

// pickWinning トリックを取れる札のうち一番安いもの
func (t *Tarabish) pickWinning(p *TarabishPlayer, valid []int) (int, bool) {
	bestIdx, bestPts := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !t.wouldWin(c) {
			continue
		}
		if pts := TarabishCardPoints(c, t.trumpSuit); bestIdx < 0 || pts < bestPts {
			bestIdx, bestPts = i, pts
		}
	}
	return bestIdx, bestIdx >= 0
}

// TarabishHint ヒント情報
type TarabishHint struct {
	// CardIndex 推奨する手札のインデックス（ビッド中は nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
}

// GetHint 人間プレイヤーへの推奨手を返す
func (t *Tarabish) GetHint() *TarabishHint {
	if t.gameEndFlag {
		return nil
	}
	if t.phase == TarabishPhaseBid && t.currentPlayerIdx == 0 {
		if t.cpuWantsTrump(0) {
			return &TarabishHint{Reason: "tarabishTakeTrump"}
		}
		return &TarabishHint{Reason: "tarabishPassTrump"}
	}
	if !t.IsHumanTurn() || t.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := t.chooseCpuCard(0)
	reason := "tarabishWinTrick"
	if t.partnerIsWinning(0) {
		reason = "tarabishFeedPartner"
	}
	return &TarabishHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (t *Tarabish) GetPhase() TarabishPhase { return t.phase }

// GetConfig 現在の設定
func (t *Tarabish) GetConfig() TarabishConfig { return t.config }

// SetConfig 設定を差し替える
func (t *Tarabish) SetConfig(c TarabishConfig) { t.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (t *Tarabish) GetRoundNumber() int { return t.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (t *Tarabish) GetTrickNumber() int { return t.trickNumber }

// GetTrumpSuit 切り札のスート
func (t *Tarabish) GetTrumpSuit() int { return t.trumpSuit }

// GetUpCard 切り札候補として表向きにした 1 枚
func (t *Tarabish) GetUpCard() *Card { return t.upCard }

// GetTrumpTakerIdx 切り札を引き受けたプレイヤー (-1: 未決定)
func (t *Tarabish) GetTrumpTakerIdx() int { return t.trumpTakerIdx }

// GetScore チームの累計得点
func (t *Tarabish) GetScore(team int) int {
	if team < 0 || team >= TarabishTeamCnt {
		return 0
	}
	return t.scores[team]
}

// SetScoreForTestUse チームの累計得点を設定する（復元・テスト用）
func (t *Tarabish) SetScoreForTestUse(team, n int) {
	if team >= 0 && team < TarabishTeamCnt {
		t.scores[team] = n
	}
}

// GetRoundPoints チームの現ラウンドのカード点
func (t *Tarabish) GetRoundPoints(team int) int {
	if team < 0 || team >= TarabishTeamCnt {
		return 0
	}
	return t.roundPoints[team]
}

// GetCurrentTrick 現在のトリック
func (t *Tarabish) GetCurrentTrick() []*TrickCard { return t.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (t *Tarabish) GetCurrentPlayerIdx() int { return t.currentPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (t *Tarabish) GetLeadPlayerIdx() int { return t.leadPlayerIdx }

// GetDealerIdx ディーラー
func (t *Tarabish) GetDealerIdx() int { return t.dealerIdx }

// GetPlayerCnt プレイヤー数
func (t *Tarabish) GetPlayerCnt() int { return len(t.players) }

// GetPlayer 指定インデックスのプレイヤー
func (t *Tarabish) GetPlayer(i int) *TarabishPlayer {
	if i < 0 || i >= len(t.players) {
		return nil
	}
	return t.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (t *Tarabish) GetGameEndFlag() bool { return t.gameEndFlag }

// GetWinnerTeam 勝利チーム (-1: 未確定または同点)
func (t *Tarabish) GetWinnerTeam() int { return t.winnerTeam }

// IsHumanTurn 人間の手番か
func (t *Tarabish) IsHumanTurn() bool {
	return !t.gameEndFlag && t.phase == TarabishPhasePlay && t.currentPlayerIdx == 0
}

// IsHumanBidTurn 人間が切り札の選択をする番か
func (t *Tarabish) IsHumanBidTurn() bool {
	return !t.gameEndFlag && t.phase == TarabishPhaseBid && t.currentPlayerIdx == 0
}

// GiveUp 投了する
func (t *Tarabish) GiveUp() {
	if t.gameEndFlag {
		return
	}
	t.phase = TarabishPhaseGameEnd
	t.gameEndFlag = true
	t.winnerTeam = 1
	t.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (t *Tarabish) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	t.appendLogAt(t.trickNumber, playerIdx, actionType, detail, cards)
}

// tarabishJSON is the KV snapshot format for Tarabish.
type tarabishJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*TarabishPlayer    `json:"pl"`
	Config           TarabishConfig       `json:"cf"`
	Phase            TarabishPhase        `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	TrumpSuit        int                  `json:"ts"`
	UpCard           *Card                `json:"uc"`
	TrumpTakerIdx    int                  `json:"tt"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	CurrentPlayerIdx int                  `json:"cp"`
	LeadPlayerIdx    int                  `json:"lp"`
	DealerIdx        int                  `json:"di"`
	Scores           [TarabishTeamCnt]int `json:"sc"`
	RoundPoints      [TarabishTeamCnt]int `json:"rp"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerTeam       int                  `json:"wt"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (t *Tarabish) MarshalJSON() ([]byte, error) {
	return json.Marshal(&tarabishJSON{
		TrumpCards:       t.trumpCards,
		Players:          t.players,
		Config:           t.config,
		Phase:            t.phase,
		RoundNumber:      t.roundNumber,
		TrickNumber:      t.trickNumber,
		TrumpSuit:        t.trumpSuit,
		UpCard:           t.upCard,
		TrumpTakerIdx:    t.trumpTakerIdx,
		CurrentTrick:     t.currentTrick,
		CurrentPlayerIdx: t.currentPlayerIdx,
		LeadPlayerIdx:    t.leadPlayerIdx,
		DealerIdx:        t.dealerIdx,
		Scores:           t.scores,
		RoundPoints:      t.roundPoints,
		GameEndFlag:      t.gameEndFlag,
		WinnerTeam:       t.winnerTeam,
		ActionLog:        t.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (t *Tarabish) UnmarshalJSON(data []byte) error {
	var j tarabishJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < TarabishPhaseBid || j.Phase > TarabishPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.TrickNumber < 0 || j.TrickNumber > TarabishTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if len(j.ActionLog) > tarabishMaxSliceLen {
		return errors.New("tarabish: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > TarabishPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= TarabishPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.TrumpTakerIdx < -1 || j.TrumpTakerIdx >= TarabishPlayerCnt {
		return fmt.Errorf("invalid trump taker: %d", j.TrumpTakerIdx)
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= TarabishTeamCnt {
		return fmt.Errorf("invalid winner team: %d", j.WinnerTeam)
	}
	if j.TrumpCards != nil {
		t.trumpCards = j.TrumpCards
	}
	if len(j.Players) == TarabishPlayerCnt {
		t.players = j.Players
	}
	t.config = j.Config
	t.phase = j.Phase
	t.roundNumber = j.RoundNumber
	t.trickNumber = j.TrickNumber
	t.trumpSuit = j.TrumpSuit
	t.upCard = j.UpCard
	t.trumpTakerIdx = j.TrumpTakerIdx
	t.currentTrick = j.CurrentTrick
	t.currentPlayerIdx = j.CurrentPlayerIdx
	t.leadPlayerIdx = j.LeadPlayerIdx
	t.dealerIdx = j.DealerIdx
	t.scores = j.Scores
	t.roundPoints = j.RoundPoints
	t.gameEndFlag = j.GameEndFlag
	t.winnerTeam = j.WinnerTeam
	t.actionLog = j.ActionLog
	return nil
}
