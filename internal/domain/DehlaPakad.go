//go:build !js || !wasm || extra

// Package domain デーラ・パカド (Dehla Pakad) のドメインモデル。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// DehlaPakadPlayerCnt はプレイヤー数。
const DehlaPakadPlayerCnt = 4

// DehlaPakadTeamCnt はチーム数。
const DehlaPakadTeamCnt = 2

// DehlaPakadHandSize は 1 人の手札枚数 (52 枚 ÷ 4 人)。
const DehlaPakadHandSize = 13

// DehlaPakadFirstBatch は切り札を決める前に配る枚数。
const DehlaPakadFirstBatch = 5

// DehlaPakadTrickCount は 1 ハンドのトリック数。
const DehlaPakadTrickCount = DehlaPakadHandSize

// DehlaPakadTenValue は「デーラ」── 争奪の的になる 10 の位。
const DehlaPakadTenValue = 10

// DehlaPakadTenCnt は 1 ハンドに存在する 10 の枚数 (各スート 1 枚)。
const DehlaPakadTenCnt = 4

// DehlaPakadStreakForKot はコートになる連勝数。
const DehlaPakadStreakForKot = 7

// フェーズ。
const (
	// DehlaPakadPhaseSelectTrump 切り札の宣言待ち (最初の 5 枚を見て決める)。
	DehlaPakadPhaseSelectTrump = "selectTrump"
	// DehlaPakadPhasePlay プレイ中。
	DehlaPakadPhasePlay = "play"
	// DehlaPakadPhaseHandEnd 1 ハンド終了。
	DehlaPakadPhaseHandEnd = "handEnd"
	// DehlaPakadPhaseGameEnd ゲーム終了。
	DehlaPakadPhaseGameEnd = "gameEnd"
)

// DehlaPakadHandResult は 1 ハンドの結果。
type DehlaPakadHandResult struct {
	// WinnerTeam は勝ったチーム。
	WinnerTeam int
	// TeamTens はチーム別に取った 10 の枚数。
	TeamTens [DehlaPakadTeamCnt]int
	// Kot はこのハンドでコートが決まったか。
	Kot bool
	// KotReason は "allTens" / "streak" / "" のいずれか。
	KotReason string
	// DealerIdx はこのハンドの親。
	DealerIdx int
	// TrumpSuit はこのハンドの切り札。
	TrumpSuit int
}

// DehlaPakad はデーラ・パカドの状態を保持する集約ルート。
type DehlaPakad struct {
	trumpCards      *TrumpCards
	players         []*DehlaPakadPlayer
	config          DehlaPakadConfig
	phase           string
	handNumber      int
	dealerIdx       int
	trumpSuit       int // -1 = 未宣言
	currentPlayer   int
	leadPlayer      int
	trickNumber     int
	currentTrick    []*TrickCard
	lastTrick       []*TrickCard
	lastTrickWinner int
	// centrePile は「まだ誰も引き取っていない」札。
	centrePile []*Card
	// prevTrickWinner は直前のトリックを取った席 (-1 = まだ無い)。
	prevTrickWinner int
	teamTens        [DehlaPakadTeamCnt]int
	teamKots        [DehlaPakadTeamCnt]int
	// streakTeam / streakCount は連勝を数える。7 連勝でコート。
	streakTeam  int
	streakCount int
	gameEndFlag bool
	winnerTeam  int
	lastResult  *DehlaPakadHandResult
	handHistory []*DehlaPakadHandResult
	actionLogBase
}

// NewDehlaPakad はコンストラクタ。
func NewDehlaPakad(trumpCards *TrumpCards, players []*DehlaPakadPlayer, config DehlaPakadConfig) *DehlaPakad {
	return &DehlaPakad{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		phase:           DehlaPakadPhaseSelectTrump,
		trumpSuit:       -1,
		lastTrickWinner: -1,
		prevTrickWinner: -1,
		streakTeam:      -1,
		winnerTeam:      -1,
	}
}

// NewDefaultDehlaPakad は標準の 4 人構成 (1 human + 3 CPU) を生成する。
func NewDefaultDehlaPakad() *DehlaPakad {
	players := make([]*DehlaPakadPlayer, DehlaPakadPlayerCnt)
	players[0] = NewDehlaPakadPlayer(true)
	for i := 1; i < DehlaPakadPlayerCnt; i++ {
		players[i] = NewDehlaPakadPlayer(false)
	}
	return NewDehlaPakad(NewTrumpCards(0), players, DefaultDehlaPakadConfig())
}

// DehlaPakadTeamOf は席のチームを返す。
//
// **相方は向かい。** 敵味方が交互に座るので、隣は必ず相手。
func DehlaPakadTeamOf(playerIdx int) int { return playerIdx % DehlaPakadTeamCnt }

// DehlaPakadNextSeat は次の手番の席を返す (反時計回り)。
func DehlaPakadNextSeat(playerIdx int) int { return (playerIdx + 1) % DehlaPakadPlayerCnt }

// Reset は新しいゲームを開始する。
func (d *DehlaPakad) Reset() {
	for _, p := range d.players {
		p.ResetDeal()
	}
	d.handNumber = 1
	// **開幕の親は席 3。** 切り札を決めるのは親の右隣なので、ここを 0 にすると
	// 人間 (席 0) は第 1 ハンドで宣言に立てず、しかも親側が勝つまで親が動かない
	// 規則と噛み合って何ハンドも一度も決められないことがある。
	d.dealerIdx = DehlaPakadPlayerCnt - 1
	d.teamKots = [DehlaPakadTeamCnt]int{}
	d.streakTeam = -1
	d.streakCount = 0
	d.gameEndFlag = false
	d.winnerTeam = -1
	d.lastResult = nil
	d.handHistory = make([]*DehlaPakadHandResult, 0)
	d.actionLog = make([]*ActionLogEntry, 0)
	d.startHand()
}

// NextHand は次のハンドを開始する。
func (d *DehlaPakad) NextHand() {
	if d.gameEndFlag || d.phase != DehlaPakadPhaseHandEnd {
		return
	}
	d.handNumber++
	d.startHand()
}

// startHand は最初の 5 枚を配り、切り札の宣言待ちにする。
func (d *DehlaPakad) startHand() {
	d.trumpSuit = -1
	d.currentTrick = nil
	d.lastTrick = nil
	d.lastTrickWinner = -1
	d.prevTrickWinner = -1
	d.centrePile = nil
	d.trickNumber = 1
	d.teamTens = [DehlaPakadTeamCnt]int{}
	for _, p := range d.players {
		p.ResetDeal()
	}
	d.trumpCards = NewTrumpCards(0)
	d.trumpCards.Shuffle()
	// **最初は 5 枚だけ。** 切り札はこの 5 枚を見て決める規則なので、
	// 13 枚配ってから訊くと別のゲームになる。
	d.dealBatch(DehlaPakadFirstBatch)
	// 反時計回りなので、親の右隣は次の席。そこが切り札を決めてリードする。
	d.leadPlayer = DehlaPakadNextSeat(d.dealerIdx)
	d.currentPlayer = d.leadPlayer
	d.phase = DehlaPakadPhaseSelectTrump
	d.appendLog(-1, "deal", fmt.Sprintf("hand %d: dealer=%d, %d cards each",
		d.handNumber, d.dealerIdx, DehlaPakadFirstBatch), nil)
}

// dealBatch は各席に n 枚ずつ配る (親の右隣から反時計回り)。
func (d *DehlaPakad) dealBatch(n int) {
	for i := 0; i < n; i++ {
		for j := 0; j < DehlaPakadPlayerCnt; j++ {
			idx := (d.dealerIdx + 1 + j) % DehlaPakadPlayerCnt
			if card := d.trumpCards.DrawCard(); card != nil {
				d.players[idx].AddCard(card)
			}
		}
	}
}

// GetTrumpChooserIdx は切り札を決める席を返す (親の右隣)。
func (d *DehlaPakad) GetTrumpChooserIdx() int { return DehlaPakadNextSeat(d.dealerIdx) }

// SelectTrump は切り札を宣言する。
func (d *DehlaPakad) SelectTrump(suit int) error {
	if d.gameEndFlag {
		return ErrGameEnded
	}
	if d.phase != DehlaPakadPhaseSelectTrump {
		return ErrWrongPhase
	}
	if !d.players[d.GetTrumpChooserIdx()].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return d.applySelectTrump(suit)
}

// applySelectTrump は切り札を確定し、残り 8 枚を配ってプレイに入る。
func (d *DehlaPakad) applySelectTrump(suit int) error {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainErrorCode(ErrInvalidPlay, "dehlapakad.errTrumpSuit", nil)
	}
	d.trumpSuit = suit
	// **残りは切り札を決めたあとに配る。** 先に 13 枚渡すと、宣言する人が
	// 手札全部を見てから決められてしまう。
	d.dealBatch(DehlaPakadHandSize - DehlaPakadFirstBatch)
	for _, p := range d.players {
		dehlaPakadSortHand(p, d.trumpSuit)
	}
	d.phase = DehlaPakadPhasePlay
	d.appendLog(d.GetTrumpChooserIdx(), "trump",
		fmt.Sprintf("player %d calls %s", d.GetTrumpChooserIdx(), DehlaPakadSuitName(suit)), nil)
	return nil
}

// CpuSelectTrump は CPU が切り札を宣言する。
func (d *DehlaPakad) CpuSelectTrump() {
	if d.gameEndFlag || d.phase != DehlaPakadPhaseSelectTrump {
		return
	}
	if d.players[d.GetTrumpChooserIdx()].GetIsHuman() {
		return
	}
	_ = d.applySelectTrump(d.cpuPickTrump(d.GetTrumpChooserIdx()))
}

// DehlaPakadSuitName はスートの安定した識別名を返す (i18n キー用)。
func DehlaPakadSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "spade"
	case CardDesignClover:
		return "club"
	case CardDesignHeart:
		return "heart"
	case CardDesignDiamond:
		return "diamond"
	default:
		return "unknown"
	}
}

// PlayerPlay は人間が 1 枚出す。
func (d *DehlaPakad) PlayerPlay(cardIndex int) error {
	if d.gameEndFlag {
		return ErrGameEnded
	}
	if d.phase != DehlaPakadPhasePlay {
		return ErrWrongPhase
	}
	if !d.players[d.currentPlayer].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return d.applyPlay(d.currentPlayer, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (d *DehlaPakad) CpuPlay() {
	if d.gameEndFlag || d.phase != DehlaPakadPhasePlay {
		return
	}
	if d.players[d.currentPlayer].GetIsHuman() {
		return
	}
	_ = d.applyPlay(d.currentPlayer, d.cpuSelectCard(d.currentPlayer))
}

// applyPlay は 1 枚出す共通処理。
func (d *DehlaPakad) applyPlay(playerIdx, cardIndex int) error {
	player := d.players[playerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "dehlapakad.errCardRange", nil)
	}
	card := player.GetCard(cardIndex)
	if err := validateCardIsPlayable(d.GetPlayableIndices(playerIdx), player, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	d.currentTrick = append(d.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: played})
	d.appendLog(playerIdx, "play",
		fmt.Sprintf("player %d plays %s", playerIdx, cardStr(played)), []*Card{played})

	if len(d.currentTrick) < DehlaPakadPlayerCnt {
		d.currentPlayer = DehlaPakadNextSeat(d.currentPlayer)
		return nil
	}
	d.resolveTrick()
	return nil
}

// GetPlayableIndices は出せる手札のインデックスを返す (リードスート必従)。
func (d *DehlaPakad) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(d.players) {
		return nil
	}
	player := d.players[playerIdx]
	all := make([]int, 0, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		all = append(all, i)
	}
	if len(d.currentTrick) == 0 {
		return all
	}
	lead := d.currentTrick[0].Card.GetDesign()
	follow := make([]int, 0, len(all))
	for _, i := range all {
		if c := player.GetCard(i); c != nil && c.GetDesign() == lead {
			follow = append(follow, i)
		}
	}
	if len(follow) == 0 {
		return all
	}
	return follow
}

// resolveTrick はトリックを解決する。
//
// **札は勝った席がすぐ持って帰るわけではない。** 同じ席が 2 トリック続けて
// 取ったときにはじめて中央の山ごと引き取る ── これがこのゲームの心臓部で、
// 「取ったのに 10 が手に入らない」局面を作る。最終トリックだけは例外で、
// 取った席が無条件に山を引き取る。
func (d *DehlaPakad) resolveTrick() {
	winner := d.trickWinner()
	for _, tc := range d.currentTrick {
		if tc != nil && tc.Card != nil {
			d.centrePile = append(d.centrePile, tc.Card)
		}
	}
	d.lastTrick = d.currentTrick
	d.lastTrickWinner = winner
	d.currentTrick = nil
	d.appendLog(winner, "trick_win",
		fmt.Sprintf("player %d wins trick %d", winner, d.trickNumber), nil)

	lastTrick := d.trickNumber >= DehlaPakadTrickCount
	if winner == d.prevTrickWinner || lastTrick {
		d.collectCentrePile(winner, lastTrick)
	}
	d.prevTrickWinner = winner

	if lastTrick {
		d.finishHand()
		return
	}
	d.trickNumber++
	d.leadPlayer = winner
	d.currentPlayer = winner
}

// collectCentrePile は中央の山を winner のチームに引き取らせる。
func (d *DehlaPakad) collectCentrePile(winner int, lastTrick bool) {
	if len(d.centrePile) == 0 {
		return
	}
	tens := 0
	for _, c := range d.centrePile {
		if c != nil && c.GetValue() == DehlaPakadTenValue {
			tens++
		}
	}
	d.teamTens[DehlaPakadTeamOf(winner)] += tens
	d.players[winner].AddTrick(d.centrePile)
	reason := "two in a row"
	if lastTrick && winner != d.prevTrickWinner {
		reason = "last trick"
	}
	d.appendLog(winner, "collect",
		fmt.Sprintf("player %d gathers %d card(s) (%s), %d ten(s)",
			winner, len(d.centrePile), reason, tens), d.centrePile)
	d.centrePile = nil
}

// trickWinner は切り札 > 台札スートで最強の札を出した席を返す。
func (d *DehlaPakad) trickWinner() int {
	if len(d.currentTrick) == 0 {
		return d.leadPlayer
	}
	lead := d.currentTrick[0].Card.GetDesign()
	winner, best := d.currentTrick[0].PlayerIdx, -1
	for _, tc := range d.currentTrick {
		if tc == nil || tc.Card == nil {
			continue
		}
		if rank := dehlaPakadWinRank(tc.Card, lead, d.trumpSuit); rank > best {
			best, winner = rank, tc.PlayerIdx
		}
	}
	return winner
}

// dehlaPakadWinRank は勝ち比べの順位を返す。切り札 > 台札 > それ以外。
func dehlaPakadWinRank(c *Card, lead, trump int) int {
	if c == nil {
		return -1
	}
	switch c.GetDesign() {
	case trump:
		return 1000 + DehlaPakadCardStrength(c)
	case lead:
		return 100 + DehlaPakadCardStrength(c)
	default:
		return -1
	}
}

// DehlaPakadCardStrength は札の強さを返す (A が最強、2 が最弱)。
func DehlaPakadCardStrength(c *Card) int {
	if c == nil {
		return -1
	}
	if c.GetValue() == 1 {
		return 14 // エース
	}
	return c.GetValue()
}

// finishHand はハンドを締めて勝敗を決める。
func (d *DehlaPakad) finishHand() {
	result := d.judgeHand()
	d.lastResult = result
	d.handHistory = append(d.handHistory, result)

	// **連勝はコートになる。** 7 連勝でもう 1 コート。
	if d.streakTeam == result.WinnerTeam {
		d.streakCount++
	} else {
		d.streakTeam, d.streakCount = result.WinnerTeam, 1
	}
	if result.Kot {
		d.teamKots[result.WinnerTeam]++
	} else if d.streakCount >= DehlaPakadStreakForKot {
		d.teamKots[result.WinnerTeam]++
		result.Kot = true
		result.KotReason = "streak"
		d.streakCount = 0
	}

	// **親が替わるかは勝ったのがどちらかで決まる。** 親側が勝てば親は右へ
	// 移り、そうでなければ据え置き ── だからこそ 7 連勝が起こりうる。
	if DehlaPakadTeamOf(d.dealerIdx) == result.WinnerTeam {
		d.dealerIdx = DehlaPakadNextSeat(d.dealerIdx)
	}

	d.phase = DehlaPakadPhaseHandEnd
	d.appendLog(-1, "handEnd", fmt.Sprintf("hand %d: team %d wins (tens %v, kot=%v)",
		d.handNumber, result.WinnerTeam, result.TeamTens, result.Kot), nil)

	for team := 0; team < DehlaPakadTeamCnt; team++ {
		if d.teamKots[team] >= d.config.TargetKots {
			d.gameEndFlag = true
			d.winnerTeam = team
			d.phase = DehlaPakadPhaseGameEnd
			d.appendLog(-1, "gameEnd", fmt.Sprintf("team %d takes the match", team), nil)
			return
		}
	}
}

// judgeHand は 10 の枚数からハンドの勝者を決める。
//
// **判定は左右非対称。** 親でないチームは 10 を 2 枚取れば勝つが、親側は 3 枚
// 要る ── つまり 2 対 2 は親でないチームの勝ち。ここを対称にすると、
// 親が回らなくなるか、逆に連勝が起きなくなる。
func (d *DehlaPakad) judgeHand() *DehlaPakadHandResult {
	res := &DehlaPakadHandResult{
		TeamTens:  d.teamTens,
		DealerIdx: d.dealerIdx,
		TrumpSuit: d.trumpSuit,
	}
	dealerTeam := DehlaPakadTeamOf(d.dealerIdx)
	other := 1 - dealerTeam

	switch {
	case d.teamTens[dealerTeam] == DehlaPakadTenCnt:
		res.WinnerTeam, res.Kot, res.KotReason = dealerTeam, true, "allTens"
	case d.teamTens[other] == DehlaPakadTenCnt:
		res.WinnerTeam, res.Kot, res.KotReason = other, true, "allTens"
	case d.teamTens[dealerTeam] >= 3:
		res.WinnerTeam = dealerTeam
	default:
		res.WinnerTeam = other
	}
	return res
}

// dehlaPakadSortHand はスート別・強い順に並べる (切り札を先頭に置く)。
func dehlaPakadSortHand(p *DehlaPakadPlayer, trump int) {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	suitKey := func(c *Card) int {
		if c.GetDesign() == trump {
			return -1
		}
		return c.GetDesign()
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if suitKey(cards[i]) != suitKey(cards[j]) {
			return suitKey(cards[i]) < suitKey(cards[j])
		}
		return DehlaPakadCardStrength(cards[i]) > DehlaPakadCardStrength(cards[j])
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// IsHumanTurn は人間が決める番かを返す。
func (d *DehlaPakad) IsHumanTurn() bool {
	if d.gameEndFlag {
		return false
	}
	switch d.phase {
	case DehlaPakadPhaseSelectTrump:
		return d.players[d.GetTrumpChooserIdx()].GetIsHuman()
	case DehlaPakadPhasePlay:
		if d.currentPlayer < 0 || d.currentPlayer >= len(d.players) {
			return false
		}
		return d.players[d.currentPlayer].GetIsHuman()
	default:
		return false
	}
}

// DehlaPakadHint は人間への推奨手。
type DehlaPakadHint struct {
	// CardIndices は勧める手札のインデックス。
	CardIndices []int
	// TrumpSuit は勧める切り札 (宣言フェーズのみ、それ以外は -1)。
	TrumpSuit int
	// Reason は理由の識別子。
	Reason string
}

// GetHint は人間への推奨手を返す。
func (d *DehlaPakad) GetHint() *DehlaPakadHint {
	human := findHumanIdx(d.players)
	if human < 0 || d.gameEndFlag {
		return &DehlaPakadHint{TrumpSuit: -1, Reason: "none"}
	}
	switch d.phase {
	case DehlaPakadPhaseSelectTrump:
		if !d.players[d.GetTrumpChooserIdx()].GetIsHuman() {
			return &DehlaPakadHint{TrumpSuit: -1, Reason: "none"}
		}
		return &DehlaPakadHint{TrumpSuit: d.smartTrumpFor(d.GetTrumpChooserIdx()), Reason: "call_longest"}
	case DehlaPakadPhasePlay:
		if d.currentPlayer != human {
			return &DehlaPakadHint{TrumpSuit: -1, Reason: "none"}
		}
		valid := d.GetPlayableIndices(human)
		if len(valid) == 0 {
			return &DehlaPakadHint{TrumpSuit: -1, Reason: "none"}
		}
		reason := "take_the_ten"
		if !d.pileHoldsATen() {
			reason = "keep_the_lead"
		}
		return &DehlaPakadHint{
			CardIndices: []int{d.smartCardFor(human, valid)},
			TrumpSuit:   -1,
			Reason:      reason,
		}
	case DehlaPakadPhaseHandEnd:
		return &DehlaPakadHint{TrumpSuit: -1, Reason: "next_hand"}
	default:
		return &DehlaPakadHint{TrumpSuit: -1, Reason: "none"}
	}
}

// pileHoldsATen は中央の山か進行中のトリックに 10 があるかを返す。
func (d *DehlaPakad) pileHoldsATen() bool {
	for _, c := range d.centrePile {
		if c != nil && c.GetValue() == DehlaPakadTenValue {
			return true
		}
	}
	for _, tc := range d.currentTrick {
		if tc != nil && tc.Card != nil && tc.Card.GetValue() == DehlaPakadTenValue {
			return true
		}
	}
	return false
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (d *DehlaPakad) GetConfig() DehlaPakadConfig { return d.config }

// SetConfig はゲーム設定を設定する。
func (d *DehlaPakad) SetConfig(c DehlaPakadConfig) { d.config = c }

// GetPhase は現在のフェーズを返す。
func (d *DehlaPakad) GetPhase() string { return d.phase }

// GetPlayerCnt は席数を返す。
func (d *DehlaPakad) GetPlayerCnt() int { return len(d.players) }

// GetPlayer は席のプレイヤーを返す。
func (d *DehlaPakad) GetPlayer(i int) *DehlaPakadPlayer {
	if i < 0 || i >= len(d.players) {
		return nil
	}
	return d.players[i]
}

// GetPlayers は全プレイヤーを返す。
func (d *DehlaPakad) GetPlayers() []*DehlaPakadPlayer { return d.players }

// GetHandNumber は現在のハンド番号を返す。
func (d *DehlaPakad) GetHandNumber() int { return d.handNumber }

// GetDealerIdx は親の席を返す。
func (d *DehlaPakad) GetDealerIdx() int { return d.dealerIdx }

// GetTrumpSuit は切り札スートを返す (-1 = 未宣言)。
func (d *DehlaPakad) GetTrumpSuit() int { return d.trumpSuit }

// GetCurrentPlayerIdx は手番の席を返す。
func (d *DehlaPakad) GetCurrentPlayerIdx() int { return d.currentPlayer }

// GetCurrentTurn は手番の席を返す (別名)。
func (d *DehlaPakad) GetCurrentTurn() int { return d.currentPlayer }

// GetLeadPlayerIdx はリードの席を返す。
func (d *DehlaPakad) GetLeadPlayerIdx() int { return d.leadPlayer }

// GetTrickNumber は現在のトリック番号を返す。
func (d *DehlaPakad) GetTrickNumber() int { return d.trickNumber }

// GetCurrentTrick は進行中のトリックを返す。
func (d *DehlaPakad) GetCurrentTrick() []*TrickCard { return d.currentTrick }

// GetLastTrick は直前に完了したトリックを返す。
func (d *DehlaPakad) GetLastTrick() []*TrickCard { return d.lastTrick }

// GetLastTrickWinner は直前トリックを取った席を返す (-1 = なし)。
func (d *DehlaPakad) GetLastTrickWinner() int { return d.lastTrickWinner }

// GetPrevTrickWinner は「2 連勝」判定に使う直前の勝者を返す。
func (d *DehlaPakad) GetPrevTrickWinner() int { return d.prevTrickWinner }

// GetCentrePile はまだ誰も引き取っていない札を返す。
func (d *DehlaPakad) GetCentrePile() []*Card { return d.centrePile }

// GetCentrePileTens は中央の山にある 10 の枚数を返す。
func (d *DehlaPakad) GetCentrePileTens() int {
	cnt := 0
	for _, c := range d.centrePile {
		if c != nil && c.GetValue() == DehlaPakadTenValue {
			cnt++
		}
	}
	return cnt
}

// GetTeamTens はチーム別に取った 10 の枚数を返す。
func (d *DehlaPakad) GetTeamTens() []int { return d.teamTens[:] }

// GetTeamKots はチーム別のコート数を返す。
func (d *DehlaPakad) GetTeamKots() []int { return d.teamKots[:] }

// GetStreakTeam は連勝中のチームを返す (-1 = なし)。
func (d *DehlaPakad) GetStreakTeam() int { return d.streakTeam }

// GetStreakCount は連勝数を返す。
func (d *DehlaPakad) GetStreakCount() int { return d.streakCount }

// GetLastResult は直前ハンドの結果を返す。
func (d *DehlaPakad) GetLastResult() *DehlaPakadHandResult { return d.lastResult }

// GetHandHistory は完了した各ハンドの結果を古い順に返す。
func (d *DehlaPakad) GetHandHistory() []*DehlaPakadHandResult { return d.handHistory }

// GetGameEndFlag は終局フラグを返す。
func (d *DehlaPakad) GetGameEndFlag() bool { return d.gameEndFlag }

// GetWinnerTeam は勝ったチームを返す (-1 = 未決)。
func (d *DehlaPakad) GetWinnerTeam() int { return d.winnerTeam }

// --- 永続化 ---

// dehlaPakadJSON is the JSON wire format for DehlaPakad.
type dehlaPakadJSON struct {
	TrumpCards      *TrumpCards             `json:"tc"`
	Players         []*DehlaPakadPlayer     `json:"pl"`
	Config          DehlaPakadConfig        `json:"cf"`
	Phase           string                  `json:"ph"`
	HandNumber      int                     `json:"hn"`
	DealerIdx       int                     `json:"di"`
	TrumpSuit       int                     `json:"ts"`
	CurrentPlayer   int                     `json:"cp"`
	LeadPlayer      int                     `json:"lp"`
	TrickNumber     int                     `json:"tn"`
	CurrentTrick    []*TrickCard            `json:"ct"`
	LastTrick       []*TrickCard            `json:"lt"`
	LastTrickWinner int                     `json:"lw"`
	CentrePile      []*Card                 `json:"pi"`
	PrevTrickWinner int                     `json:"pw"`
	TeamTens        [DehlaPakadTeamCnt]int  `json:"tt"`
	TeamKots        [DehlaPakadTeamCnt]int  `json:"tk"`
	StreakTeam      int                     `json:"st"`
	StreakCount     int                     `json:"sc"`
	GameEndFlag     bool                    `json:"ge"`
	WinnerTeam      int                     `json:"wt"`
	LastResult      *DehlaPakadHandResult   `json:"lr"`
	HandHistory     []*DehlaPakadHandResult `json:"hh"`
	ActionLog       []*ActionLogEntry       `json:"al"`
}

// dehlaPakadMaxSliceLen は復元時に受け入れるスライスの上限。
const dehlaPakadMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
//
// **非公開フィールドだけの型は MarshalJSON が無いと `{}` になる。** KV に
// 保存した盤が空で返るので、ここは必ず自前で書く。
func (d *DehlaPakad) MarshalJSON() ([]byte, error) {
	return json.Marshal(dehlaPakadJSON{
		TrumpCards:      d.trumpCards,
		Players:         d.players,
		Config:          d.config,
		Phase:           d.phase,
		HandNumber:      d.handNumber,
		DealerIdx:       d.dealerIdx,
		TrumpSuit:       d.trumpSuit,
		CurrentPlayer:   d.currentPlayer,
		LeadPlayer:      d.leadPlayer,
		TrickNumber:     d.trickNumber,
		CurrentTrick:    d.currentTrick,
		LastTrick:       d.lastTrick,
		LastTrickWinner: d.lastTrickWinner,
		CentrePile:      d.centrePile,
		PrevTrickWinner: d.prevTrickWinner,
		TeamTens:        d.teamTens,
		TeamKots:        d.teamKots,
		StreakTeam:      d.streakTeam,
		StreakCount:     d.streakCount,
		GameEndFlag:     d.gameEndFlag,
		WinnerTeam:      d.winnerTeam,
		LastResult:      d.lastResult,
		HandHistory:     d.handHistory,
		ActionLog:       d.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *DehlaPakad) UnmarshalJSON(data []byte) error {
	var j dehlaPakadJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > dehlaPakadMaxSliceLen || len(j.CurrentTrick) > dehlaPakadMaxSliceLen ||
		len(j.LastTrick) > dehlaPakadMaxSliceLen || len(j.CentrePile) > dehlaPakadMaxSliceLen ||
		len(j.HandHistory) > dehlaPakadMaxSliceLen || len(j.ActionLog) > dehlaPakadMaxSliceLen {
		return fmt.Errorf("dehlapakad: input array exceeds maximum allowed size")
	}
	if len(j.Players) != DehlaPakadPlayerCnt {
		return fmt.Errorf("dehlapakad: invalid player count %d, expected %d", len(j.Players), DehlaPakadPlayerCnt)
	}
	d.trumpCards = j.TrumpCards
	if d.trumpCards == nil {
		d.trumpCards = NewTrumpCards(0)
	}
	d.players = j.Players
	d.config = j.Config
	d.phase = j.Phase
	d.handNumber = j.HandNumber
	d.dealerIdx = j.DealerIdx
	d.trumpSuit = j.TrumpSuit
	d.currentPlayer = j.CurrentPlayer
	d.leadPlayer = j.LeadPlayer
	d.trickNumber = j.TrickNumber
	d.currentTrick = j.CurrentTrick
	d.lastTrick = j.LastTrick
	d.lastTrickWinner = j.LastTrickWinner
	d.centrePile = j.CentrePile
	d.prevTrickWinner = j.PrevTrickWinner
	d.teamTens = j.TeamTens
	d.teamKots = j.TeamKots
	d.streakTeam = j.StreakTeam
	d.streakCount = j.StreakCount
	d.gameEndFlag = j.GameEndFlag
	d.winnerTeam = j.WinnerTeam
	d.lastResult = j.LastResult
	d.handHistory = j.HandHistory
	if d.handHistory == nil {
		d.handHistory = make([]*DehlaPakadHandResult, 0)
	}
	d.actionLog = j.ActionLog
	if d.actionLog == nil {
		d.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// dehlaPakadHandResultJSON is the JSON wire format for DehlaPakadHandResult.
type dehlaPakadHandResultJSON struct {
	WinnerTeam int                    `json:"w"`
	TeamTens   [DehlaPakadTeamCnt]int `json:"t"`
	Kot        bool                   `json:"k"`
	KotReason  string                 `json:"r"`
	DealerIdx  int                    `json:"d"`
	TrumpSuit  int                    `json:"s"`
}

// MarshalJSON implements json.Marshaler.
func (r *DehlaPakadHandResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(dehlaPakadHandResultJSON{
		WinnerTeam: r.WinnerTeam,
		TeamTens:   r.TeamTens,
		Kot:        r.Kot,
		KotReason:  r.KotReason,
		DealerIdx:  r.DealerIdx,
		TrumpSuit:  r.TrumpSuit,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *DehlaPakadHandResult) UnmarshalJSON(data []byte) error {
	var j dehlaPakadHandResultJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	r.WinnerTeam = j.WinnerTeam
	r.TeamTens = j.TeamTens
	r.Kot = j.Kot
	r.KotReason = j.KotReason
	r.DealerIdx = j.DealerIdx
	r.TrumpSuit = j.TrumpSuit
	return nil
}

// dehlaPakadRandIntn は rand.Intn の薄いラッパ (n <= 0 を握りつぶす)。
func dehlaPakadRandIntn(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.Intn(n)
}
