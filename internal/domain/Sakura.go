//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// SakuraPhase はゲームフェーズ。
type SakuraPhase int

// さくらのフェーズ定数。
const (
	// SakuraPhasePlay は手札を 1 枚出す手番。
	SakuraPhasePlay SakuraPhase = 0
	// SakuraPhaseRoundEnd はラウンド終了 (結果表示。次ラウンド待ち)。
	SakuraPhaseRoundEnd SakuraPhase = 1
	// SakuraPhaseGameEnd は終局。
	SakuraPhaseGameEnd SakuraPhase = 2
)

// SakuraSeatResult は 1 席のラウンド結果。
type SakuraSeatResult struct {
	// CardPoints は獲得札の素点。
	CardPoints int `json:"cardPoints"`
	// Bonuses は成立した追加役。
	Bonuses []SakuraBonus `json:"bonuses"`
	// BonusPoints は追加役の合計点。
	BonusPoints int `json:"bonusPoints"`
	// Total は素点と追加役の合計。
	Total int `json:"total"`
}

// SakuraRoundResult は 1 ラウンドの結果。
type SakuraRoundResult struct {
	// Round はラウンド番号。
	Round int `json:"round"`
	// Winner は勝った席 (-1 = 引き分け)。
	Winner int `json:"winner"`
	// Seats は席ごとの内訳 (席番号順)。
	Seats []SakuraSeatResult `json:"seats"`
}

// SakuraHint はヒント情報。
type SakuraHint struct {
	// CardIndex は推奨する手札インデックス (-1 = なし)。
	CardIndex int `json:"cardIndex"`
	// FieldIndex は推奨する場札インデックス (-1 = 合う札なし)。
	FieldIndex int `json:"fieldIndex"`
	// Reason は理由キー。
	Reason string `json:"reason"`
}

// Sakura はさくら (肥後花) の集約ルート。
//
// **勝敗は役ではなく点数の合計で決まる。** こいこいが「役ができたら止める」
// ゲームなのに対し、さくらは手札を配り切るまで打ち、獲得した札の点数を数える ──
// 途中で終わる手段が無いので、こいこい決断に当たるフェーズを持たない。
type Sakura struct {
	players []*SakuraPlayer
	config  SakuraConfig

	phase SakuraPhase
	turn  int
	// dealer は親。ラウンドごとに 1 つずつ回る。
	dealer int
	// field は場札。
	field []*Card
	// stock は山札。
	stock []*Card
	// round は現在のラウンド番号 (1 始まり)。
	round int

	gameEndFlag bool
	// winner は終局時の勝者 (-1 = 引き分け)。
	winner int
	// lastResult は直前のラウンド結果。
	lastResult *SakuraRoundResult

	actionLogBase
}

// NewSakura はコンストラクタ。
func NewSakura(players []*SakuraPlayer, config SakuraConfig) *Sakura {
	return &Sakura{
		players: players,
		config:  config,
		phase:   SakuraPhasePlay,
		winner:  -1,
		round:   1,
	}
}

// NewDefaultSakura は既定構成 (1 human + CPU) でさくらを生成する。
func NewDefaultSakura() *Sakura {
	cfg := DefaultSakuraConfig()
	return NewSakura(NewSakuraPlayersForTable(cfg.Seats), cfg)
}

// GetPlayers は席の一覧を返す。
func (g *Sakura) GetPlayers() []*SakuraPlayer { return g.players }

// GetConfig は卓設定を返す。
func (g *Sakura) GetConfig() SakuraConfig { return g.config }

// SetConfig は卓設定を差し替える (次の Reset で反映される)。
func (g *Sakura) SetConfig(cfg SakuraConfig) { g.config = cfg }

// GetPlayerCnt は席数を返す。
func (g *Sakura) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定した席を返す (範囲外なら nil)。
func (g *Sakura) GetPlayer(i int) *SakuraPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPhase は現在のフェーズを返す。
func (g *Sakura) GetPhase() SakuraPhase { return g.phase }

// GetTurn は手番の席を返す。
func (g *Sakura) GetTurn() int { return g.turn }

// GetDealer は親の席を返す。
func (g *Sakura) GetDealer() int { return g.dealer }

// GetField は場札を返す。
func (g *Sakura) GetField() []*Card { return g.field }

// GetStockCount は山札の残り枚数を返す。
func (g *Sakura) GetStockCount() int { return len(g.stock) }

// GetRound は現在のラウンド番号を返す。
func (g *Sakura) GetRound() int { return g.round }

// GetGameEndFlag は終局したかを返す。
func (g *Sakura) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinner は終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *Sakura) GetWinner() int { return g.winner }

// GetLastResult は直前のラウンド結果を返す。
func (g *Sakura) GetLastResult() *SakuraRoundResult { return g.lastResult }

// HumanSeat は人間の席を返す (居なければ 0)。
func (g *Sakura) HumanSeat() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return 0
}

// IsHumanTurn は人間の手番かを返す。
func (g *Sakura) IsHumanTurn() bool {
	return g.phase == SakuraPhasePlay && !g.gameEndFlag &&
		g.turn >= 0 && g.turn < len(g.players) && g.players[g.turn].GetIsHuman()
}

// playerName は席の表示名を返す。
func (g *Sakura) playerName(i int) string {
	if i < 0 || i >= len(g.players) {
		return "?"
	}
	return g.players[i].GetName()
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。
//
// **席は設定から作り直す。** 設定だけ替えて席を据え置くと、席数と人数が食い違った
// まま配ることになり、その局面は保存すると復元できなくなる。
func (g *Sakura) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultSakuraConfig()
	}
	if len(g.players) != g.config.Seats {
		g.players = NewSakuraPlayersForTable(g.config.Seats)
	}
	for _, p := range g.players {
		p.ResetForRound()
		p.score = 0
		p.roundWins = 0
	}
	g.phase = SakuraPhasePlay
	g.gameEndFlag = false
	g.winner = -1
	g.lastResult = nil
	g.round = 1
	g.dealer = 0
	g.actionLog = make([]*ActionLogEntry, 0)
	g.startRound()
}

// NextRound はラウンド終了後に次のラウンドを開始する。
func (g *Sakura) NextRound() {
	if g.gameEndFlag || g.phase != SakuraPhaseRoundEnd {
		return
	}
	g.round++
	g.dealer = (g.dealer + 1) % len(g.players)
	g.startRound()
}

// startRound は配札と場札の配置を行い、プレイフェーズを開始する。
func (g *Sakura) startRound() {
	deck := buildKoiKoiDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

	for _, p := range g.players {
		p.ResetForRound()
	}
	g.field = make([]*Card, 0, SakuraFieldSize)
	g.phase = SakuraPhasePlay
	// 親から順に打つ。
	g.turn = g.dealer

	pos := 0
	for range SakuraHandSize {
		for _, p := range g.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	for range SakuraFieldSize {
		g.field = append(g.field, deck[pos])
		pos++
	}
	g.stock = append([]*Card(nil), deck[pos:]...)

	g.appendLog(-1, "deal", fmt.Sprintf("round %d dealt (field %d, stock %d)",
		g.round, len(g.field), len(g.stock)), append([]*Card(nil), g.field...))
}

// --- 捕獲 ---

// fieldMatches は場札のうち card と同じ月のインデックスを返す。
func (g *Sakura) fieldMatches(card *Card) []int {
	var out []int
	for i, c := range g.field {
		if koikoiSameMonth(c, card) {
			out = append(out, i)
		}
	}
	return out
}

// bestFieldMatch は候補のうち最も点数の高い場札インデックスを返す。
func (g *Sakura) bestFieldMatch(matches []int) int {
	best, bestVal := -1, -1
	for _, idx := range matches {
		if idx < 0 || idx >= len(g.field) {
			continue
		}
		if v := SakuraCardPoints(g.field[idx]); v > bestVal {
			best, bestVal = idx, v
		}
	}
	return best
}

// placeCard は 1 枚 (手札またはめくり札) を場と突き合わせて解決し、獲得したかを返す。
//
//   - 一致なし: 場に置く (捨て札)。
//   - 一致 1 枚: その札とともに獲得。
//   - 一致 2 枚: chosen が一致札ならそれを、そうでなければ点数の高いほうを獲得。
//   - 一致 3 枚: 場の同月 4 枚目なので、すべて獲得。
func (g *Sakura) placeCard(seat int, card *Card, chosen int) bool {
	matches := g.fieldMatches(card)
	if len(matches) == 0 {
		g.field = append(g.field, card)
		return false
	}
	var take []int
	switch {
	case len(matches) >= 3:
		take = matches
	case sakuraNeedsFieldChoice(len(matches)):
		sel := -1
		for _, idx := range matches {
			if idx == chosen {
				sel = idx
			}
		}
		if sel < 0 {
			sel = g.bestFieldMatch(matches)
		}
		take = []int{sel}
	default:
		take = []int{matches[0]}
	}
	captured := make([]*Card, 0, len(take)+1)
	captured = append(captured, card)
	for _, idx := range take {
		captured = append(captured, g.field[idx])
	}
	g.field = removeIndices(g.field, take)
	g.players[seat].AddTaken(captured...)
	return true
}

// --- Play ---

// PlayerPlay は人間が手札 handIdx を出す。fieldIdx は同月札が 2 枚ある場合に
// どちらを取るかの場札インデックス (不要なら -1)。
func (g *Sakura) PlayerPlay(handIdx, fieldIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != SakuraPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	if !g.players[g.turn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.turn]
	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	card := player.GetCard(handIdx)
	if fieldIdx >= 0 {
		if fieldIdx >= len(g.field) || !koikoiSameMonth(g.field[fieldIdx], card) {
			return NewDomainError(ErrInvalidPlay, "chosen field card does not match the played card's month")
		}
	}
	g.applyTurn(g.turn, handIdx, fieldIdx)
	return nil
}

// CpuPlay は CPU の手番を 1 つ進める。
func (g *Sakura) CpuPlay() {
	if g.gameEndFlag || g.phase != SakuraPhasePlay {
		return
	}
	p := g.players[g.turn]
	if p.GetIsHuman() || p.GetCardsSize() == 0 {
		return
	}
	handIdx, fieldIdx := g.chooseCpuPlay(g.turn)
	g.applyTurn(g.turn, handIdx, fieldIdx)
}

// applyTurn は手札を 1 枚出し、続けて山札を 1 枚めくって解決する。
//
// **山が尽きてもめくりだけを飛ばす。** 4 席だと手札 28 枚に対し山は 14 枚しか
// 残らないので、めくりを必須にすると後半の手番が成り立たなくなる。
func (g *Sakura) applyTurn(seat, handIdx, fieldIdx int) {
	player := g.players[seat]
	card := player.RemoveCard(handIdx)
	if card == nil {
		return
	}
	took := g.placeCard(seat, card, fieldIdx)
	g.appendLog(seat, "play", fmt.Sprintf("%s plays %s (%s)",
		g.playerName(seat), sakuraCardStr(card), sakuraTookWord(took)), []*Card{card})

	if len(g.stock) > 0 {
		drawn := g.stock[0]
		g.stock = g.stock[1:]
		tookDraw := g.placeCard(seat, drawn, -1)
		g.appendLog(seat, "draw", fmt.Sprintf("%s flips %s (%s)",
			g.playerName(seat), sakuraCardStr(drawn), sakuraTookWord(tookDraw)), []*Card{drawn})
	}

	g.advanceTurn()
}

// advanceTurn は手番を次に進め、全員の手札が尽きたらラウンドを終える。
func (g *Sakura) advanceTurn() {
	if g.allHandsEmpty() {
		g.finishRound()
		return
	}
	for range len(g.players) {
		g.turn = (g.turn + 1) % len(g.players)
		if g.players[g.turn].GetCardsSize() > 0 {
			return
		}
	}
	// 全員が空 (allHandsEmpty で捕まえているので通常は届かない)。
	g.finishRound()
}

// allHandsEmpty は全員の手札が空かを返す。
func (g *Sakura) allHandsEmpty() bool {
	for _, p := range g.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishRound は獲得札を集計してラウンドを終える。
func (g *Sakura) finishRound() {
	result := &SakuraRoundResult{Round: g.round, Winner: -1,
		Seats: make([]SakuraSeatResult, 0, len(g.players))}

	best, tied := -1, false
	for i, p := range g.players {
		total := p.TotalPoints()
		p.SetRoundScore(total)
		p.AddScore(total)
		result.Seats = append(result.Seats, SakuraSeatResult{
			CardPoints:  p.CardPoints(),
			Bonuses:     p.Bonuses(),
			BonusPoints: p.BonusPoints(),
			Total:       total,
		})
		switch {
		case best < 0 || total > g.players[best].GetRoundScore():
			best, tied = i, false
		case total == g.players[best].GetRoundScore():
			tied = true
		}
	}
	if !tied && best >= 0 {
		result.Winner = best
		g.players[best].AddRoundWin()
	}

	g.lastResult = result
	g.phase = SakuraPhaseRoundEnd
	if result.Winner >= 0 {
		g.appendLog(result.Winner, "round",
			fmt.Sprintf("%s takes round %d with %d points",
				g.playerName(result.Winner), g.round, result.Seats[result.Winner].Total), nil)
	} else {
		g.appendLog(-1, "round", fmt.Sprintf("round %d is a tie", g.round), nil)
	}

	if g.round >= g.config.Rounds {
		g.finishGame()
	}
}

// finishGame は通算得点で勝者を決める。
//
// **同点は勝ったラウンド数で割る。** 点数の合計だけだと、1 ラウンドの大勝と
// 何度も競り勝った積み重ねが区別できない。
func (g *Sakura) finishGame() {
	best, tied := -1, false
	for i, p := range g.players {
		if best < 0 {
			best = i
			continue
		}
		b := g.players[best]
		switch {
		case p.GetScore() > b.GetScore():
			best, tied = i, false
		case p.GetScore() < b.GetScore():
			// 現状維持。
		case p.GetRoundWins() > b.GetRoundWins():
			best, tied = i, false
		case p.GetRoundWins() == b.GetRoundWins():
			tied = true
		}
	}
	g.winner = -1
	if !tied && best >= 0 {
		g.winner = best
	}
	g.gameEndFlag = true
	g.phase = SakuraPhaseGameEnd
	if g.winner >= 0 {
		g.appendLog(g.winner, "gameEnd", fmt.Sprintf("%s wins with %d points",
			g.playerName(g.winner), g.players[g.winner].GetScore()), nil)
	} else {
		g.appendLog(-1, "gameEnd", "the game ends in a tie", nil)
	}
}

// --- CPU ---

// chooseCpuPlay は CPU の手 (手札インデックス, 場札インデックス) を選ぶ。
//
// 合わせられる札のうち、獲得できる点数がいちばん大きい組を選ぶ。合わせられる
// 札が 1 枚も無ければ、いちばん安い札を捨てる。
func (g *Sakura) chooseCpuPlay(seat int) (int, int) {
	p := g.players[seat]
	bestHand, bestField, bestGain := -1, -1, -1
	worstHand, worstVal := 0, -1
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		matches := g.fieldMatches(c)
		if len(matches) == 0 {
			if v := SakuraCardPoints(c); worstVal < 0 || v < worstVal {
				worstHand, worstVal = i, v
			}
			continue
		}
		fieldIdx := g.bestFieldMatch(matches)
		gain := SakuraCardPoints(c)
		if len(matches) >= 3 {
			for _, idx := range matches {
				gain += SakuraCardPoints(g.field[idx])
			}
			fieldIdx = -1
		} else {
			gain += SakuraCardPoints(g.field[fieldIdx])
		}
		if gain > bestGain {
			bestHand, bestField, bestGain = i, fieldIdx, gain
		}
	}
	if bestHand >= 0 {
		return bestHand, bestField
	}
	return worstHand, -1
}

// --- ヒント ---

// GetHint は人間の手番における推奨手を返す。
func (g *Sakura) GetHint() SakuraHint {
	seat := g.HumanSeat()
	if g.gameEndFlag || g.phase != SakuraPhasePlay || g.turn != seat ||
		g.players[seat].GetCardsSize() == 0 {
		return SakuraHint{CardIndex: -1, FieldIndex: -1, Reason: "none"}
	}
	handIdx, fieldIdx := g.chooseCpuPlay(seat)
	reason := "capture"
	if len(g.fieldMatches(g.players[seat].GetCard(handIdx))) == 0 {
		reason = "discard"
	}
	return SakuraHint{CardIndex: handIdx, FieldIndex: fieldIdx, Reason: reason}
}

// sakuraNeedsFieldChoice は一致枚数から「取る札を選ぶ必要があるか」を返す。
//
// **選ぶ余地があるのは 2 枚一致のときだけ。** 1 枚ならそれしか無く、3 枚なら
// 場の同月 4 枚目としてまとめて取るので、どれを押しても結果は変わらない。
// 判定をここに置くのは、画面が同じ規則を作り直すと「選べと言われたのに選択が
// 効かない」表示になるから (#5338 のレビュー指摘)。
func sakuraNeedsFieldChoice(matches int) bool { return matches == 2 }

// GetChoiceIndices は「取る札を選ぶ必要がある」手札だけを返す。
//
// GetValidFieldIndices が「合わせられるか」を答えるのに対し、こちらは
// 「選ばせるべきか」を答える。画面の確認プロンプトはこちらを見る。
func (g *Sakura) GetChoiceIndices() map[int][]int {
	out := map[int][]int{}
	for i, m := range g.GetValidFieldIndices() {
		if sakuraNeedsFieldChoice(len(m)) {
			out[i] = m
		}
	}
	return out
}

// GetValidFieldIndices は手札ごとに合わせられる場札インデックスを返す。
func (g *Sakura) GetValidFieldIndices() map[int][]int {
	out := map[int][]int{}
	seat := g.turn
	if g.phase != SakuraPhasePlay || seat < 0 || seat >= len(g.players) {
		return out
	}
	p := g.players[seat]
	for i := range p.GetCardsSize() {
		if m := g.fieldMatches(p.GetCard(i)); len(m) > 0 {
			out[i] = m
		}
	}
	return out
}

// --- 表示補助 ---

// sakuraCardStr は札の表示文字列を返す。
func sakuraCardStr(c *Card) string {
	if c == nil {
		return "-"
	}
	return fmt.Sprintf("%d月%s", c.GetDesign(), KoiKoiCardGlyph(c))
}

// sakuraTookWord は獲得したかどうかの語を返す。
func sakuraTookWord(took bool) string {
	if took {
		return "captured"
	}
	return "discarded"
}

// --- JSON ---

// sakuraJSON は Sakura の JSON 表現。
type sakuraJSON struct {
	Players     []*SakuraPlayer    `json:"pl"`
	Config      SakuraConfig       `json:"cf"`
	Phase       SakuraPhase        `json:"ph"`
	Turn        int                `json:"tn"`
	Dealer      int                `json:"dl"`
	Field       []*Card            `json:"fd"`
	Stock       []*Card            `json:"st"`
	Round       int                `json:"rd"`
	GameEndFlag bool               `json:"ge"`
	Winner      int                `json:"wn"`
	LastResult  *SakuraRoundResult `json:"lr"`
	ActionLog   []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Sakura) MarshalJSON() ([]byte, error) {
	return json.Marshal(sakuraJSON{
		Players: g.players, Config: g.config,
		Phase: g.phase, Turn: g.turn, Dealer: g.dealer,
		Field: g.field, Stock: g.stock, Round: g.round,
		GameEndFlag: g.gameEndFlag, Winner: g.winner,
		LastResult: g.lastResult, ActionLog: g.actionLog,
	})
}

// sakuraMaxSliceLen は復元時のスライス長の上限。
const sakuraMaxSliceLen = 1000

// sakuraValidateCards は札に nil や範囲外の月/インデックスが無いか検証する。
func sakuraValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("sakura: nil card in state")
		}
		m, i := c.GetDesign(), c.GetValue()
		if m < 1 || m > KoiKoiMonthCnt || i < 1 || i > KoiKoiCardsPerMonth {
			return fmt.Errorf("sakura: card out of range (month %d, index %d)", m, i)
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **札の総数が 48 枚を超えていないかまで見る。** 席ごとの上限だけを見ても、
// 同じ札を配れば点数はいくらでも積める ── 花札は同じ札が 2 枚存在しないので、
// 総数と重複の両方が不変条件になる。
func (g *Sakura) UnmarshalJSON(data []byte) error {
	var j sakuraJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > sakuraMaxSliceLen || len(j.Field) > sakuraMaxSliceLen ||
		len(j.Stock) > sakuraMaxSliceLen || len(j.ActionLog) > sakuraMaxSliceLen {
		return fmt.Errorf("sakura: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("sakura: invalid config: %w", err)
	}
	if len(j.Players) != j.Config.Seats {
		return fmt.Errorf("sakura: player count %d does not match seats %d", len(j.Players), j.Config.Seats)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("sakura: nil player in state")
		}
	}
	if j.Phase < SakuraPhasePlay || j.Phase > SakuraPhaseGameEnd {
		return fmt.Errorf("sakura: invalid phase %d", j.Phase)
	}
	if j.Turn < 0 || j.Turn >= len(j.Players) {
		return fmt.Errorf("sakura: turn out of range")
	}
	if j.Dealer < 0 || j.Dealer >= len(j.Players) {
		return fmt.Errorf("sakura: dealer out of range")
	}
	if j.Round < 1 || j.Round > j.Config.Rounds {
		return fmt.Errorf("sakura: round %d out of range", j.Round)
	}
	if j.Winner < -1 || j.Winner >= len(j.Players) {
		return fmt.Errorf("sakura: winner out of range")
	}
	if err := sakuraValidateCards(j.Field); err != nil {
		return err
	}
	if err := sakuraValidateCards(j.Stock); err != nil {
		return err
	}
	if err := sakuraValidateDeck(&j); err != nil {
		return err
	}

	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.turn = j.Turn
	g.dealer = j.Dealer
	g.field = j.Field
	g.stock = j.Stock
	g.round = j.Round
	g.gameEndFlag = j.GameEndFlag
	g.winner = j.Winner
	g.lastResult = j.LastResult
	g.actionLog = j.ActionLog
	if g.field == nil {
		g.field = make([]*Card, 0)
	}
	if g.stock == nil {
		g.stock = make([]*Card, 0)
	}
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// sakuraValidateDeck は盤上の札が花札 48 枚を 1 枚ずつしか使っていないことを検証する。
func sakuraValidateDeck(j *sakuraJSON) error {
	seen := make(map[int]bool, SakuraDeckSize)
	count := 0
	add := func(cards []*Card) error {
		for _, c := range cards {
			if c == nil {
				return fmt.Errorf("sakura: nil card in state")
			}
			key := c.GetDesign()*(KoiKoiCardsPerMonth+1) + c.GetValue()
			if seen[key] {
				return fmt.Errorf("sakura: duplicate card (month %d, index %d)", c.GetDesign(), c.GetValue())
			}
			seen[key] = true
			count++
		}
		return nil
	}
	if err := add(j.Field); err != nil {
		return err
	}
	if err := add(j.Stock); err != nil {
		return err
	}
	for _, p := range j.Players {
		if err := add(p.GetCards()); err != nil {
			return err
		}
		if err := add(p.GetTaken()); err != nil {
			return err
		}
	}
	if count > SakuraDeckSize {
		return fmt.Errorf("sakura: %d cards in play, deck holds %d", count, SakuraDeckSize)
	}
	return nil
}
