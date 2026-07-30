//go:build !js || !wasm || extra2

// Package domain — ラフ・アンド・ライダウン (Laugh and Lie Down) のドメインモデル。
//
// 16-17 世紀イングランドの 5 人用フィッシングゲーム。現存最古級の記録のひとつで、
// 実装は David Parlett の再構成 (parlettgames.uk/histocs/laughand.html、pagat が
// 参照している版) に従う。
//
// # issue #4396 の仕様案との相違
//
// issue は**フィッシングの機構そのものを取り違えている**ので、原典を採る。
//
//   - issue は「残りを場札の**山**として中央に置く」とするが、実際は
//     **12 枚を表向きに広げる**。伏せた山札は存在しない
//   - issue は「場札の**一番上**と同ランク」とするが、実際は**場のどの札とでも**
//     ランクが合えば取れる。さらに 1 枚で**1 枚または 3 枚**を取れる
//   - issue は「出せない場合は山から 1 枚引いて…ペアが揃わなければ脱落」とするが、
//     **引く手順は存在しない**。取れなくなった人は**手札を全部場に置いて**降りる。
//     置かれた札は他家が取れるようになる —— これが緊張の源で、ただの脱落ではない
//   - issue は「最後まで手札を使い切った最後のプレイヤーがポットを総取り」と
//     するが、実際は**最後まで手札が残っていた人**が 5 を取り、残りは
//     **取得枚数の 8 との過不足**を 2 枚 1 で清算する
//
// # ポットの算術が精算表を裏付ける
//
// ポットは親 3 + 他 4 人 2 = **11**。52 枚すべてが誰かの取り札になるので、
// 過不足の合計は必ず (52 - 8*5) / 2 = **6**。「最後の 1 人に 5」と合わせて
// ちょうど 11 になる。issue の「総取り」ではこの一致は起きない。
//
// # 親が残りを取る理由
//
// 終局時、最後の 1 人の手札と場の残りは**親の取り札**に入る。親だけ 1 多く
// 出しているので、その埋め合わせとして内部的に筋が通っている (原典どおり)。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// LaughAndLieDownPlayerCnt はプレイヤー数 (5 人固定)。
const LaughAndLieDownPlayerCnt = 5

// LaughAndLieDownHandSize は各自への配札枚数。
const LaughAndLieDownHandSize = 8

// LaughAndLieDownLayoutSize は表向きに広げる場札の枚数。
const LaughAndLieDownLayoutSize = 52 - LaughAndLieDownHandSize*LaughAndLieDownPlayerCnt

// LaughAndLieDownDealerAnte は親の掛け金。
const LaughAndLieDownDealerAnte = 3

// LaughAndLieDownAnte は親以外の掛け金。
const LaughAndLieDownAnte = 2

// LaughAndLieDownLastInBonus は最後まで手札が残っていた人が受け取る額。
const LaughAndLieDownLastInBonus = 5

// LaughAndLieDownPot はポットの総額 (親 3 + 他 4 人 2)。
const LaughAndLieDownPot = LaughAndLieDownDealerAnte + LaughAndLieDownAnte*(LaughAndLieDownPlayerCnt-1)

// LaughAndLieDownPhase はゲームフェーズ。
type LaughAndLieDownPhase int

// Laugh and Lie Down のフェーズ定数
const (
	// LaughAndLieDownPhasePlay 手番進行中
	LaughAndLieDownPhasePlay LaughAndLieDownPhase = iota
	// LaughAndLieDownPhaseGameEnd 終局
	LaughAndLieDownPhaseGameEnd
)

// newLaughAndLieDownDeck は 52 枚を生成する (シャッフル前)。
func newLaughAndLieDownDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, 52)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// lldShuffle は Fisher-Yates。domain の shuffleCards は casino タグのファイルに
// あり extra2 ビルドから見えないため、専用名で持つ。
func lldShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// LaughAndLieDown はラフ・アンド・ライダウンのゲームクラス。
type LaughAndLieDown struct {
	players []*LaughAndLieDownPlayer
	config  LaughAndLieDownConfig
	phase   LaughAndLieDownPhase
	// layout は表向きの場札。取れない人が置いた手札もここに積まれる。
	layout []*Card
	// won[i] は i が取った札。枚数だけが精算に効く。
	won [][]*Card
	// laidDown[i] は i が「取れなくなって手札を場に置いた」か。
	laidDown   []bool
	currentIdx int
	dealerIdx  int
	// lastInIdx は最後まで手札が残っていた人。誰も残らなかったときは -1。
	lastInIdx   int
	scores      []int
	gameEndFlag bool
	actionLog   []*ActionLogEntry
}

// NewLaughAndLieDown はコンストラクタ。
func NewLaughAndLieDown(players []*LaughAndLieDownPlayer, config LaughAndLieDownConfig) *LaughAndLieDown {
	return &LaughAndLieDown{
		players:   players,
		config:    config,
		won:       make([][]*Card, len(players)),
		laidDown:  make([]bool, len(players)),
		scores:    make([]int, len(players)),
		lastInIdx: -1,
	}
}

// NewDefaultLaughAndLieDown は標準の 5 人セットアップを返す。
func NewDefaultLaughAndLieDown() *LaughAndLieDown {
	players := make([]*LaughAndLieDownPlayer, 0, LaughAndLieDownPlayerCnt)
	players = append(players, NewLaughAndLieDownPlayer(true))
	for range LaughAndLieDownPlayerCnt - 1 {
		players = append(players, NewLaughAndLieDownPlayer(false))
	}
	return NewLaughAndLieDown(players, DefaultLaughAndLieDownConfig())
}

// Reset はゲームを初期化する。
func (l *LaughAndLieDown) Reset() {
	l.phase = LaughAndLieDownPhasePlay
	l.gameEndFlag = false
	l.actionLog = nil
	l.lastInIdx = -1
	l.dealerIdx = 0
	l.laidDown = make([]bool, len(l.players))
	l.won = make([][]*Card, len(l.players))
	l.scores = make([]int, len(l.players))
	for i := range l.won {
		l.won[i] = make([]*Card, 0, 8)
	}
	for _, p := range l.players {
		p.ResetGame()
	}

	deck := newLaughAndLieDownDeck()
	lldShuffle(deck)
	pos := 0
	for range LaughAndLieDownHandSize {
		for _, p := range l.players {
			p.AddCard(deck[pos])
			pos++
		}
	}
	// **12 枚を表向きに広げる。**伏せた山札は無い。
	l.layout = append([]*Card(nil), deck[pos:]...)

	l.currentIdx = (l.dealerIdx + 1) % len(l.players)
	l.addLog(-1, "deal", "cards dealt", nil)
	l.skipPlayersWhoCannotCapture()
}

// layoutCountOfRank は場にあるそのランクの枚数を返す。
func (l *LaughAndLieDown) layoutCountOfRank(rank int) int {
	n := 0
	for _, c := range l.layout {
		if c != nil && c.GetValue() == rank {
			n++
		}
	}
	return n
}

// GetValidPlayIndices は player が出せる手札の添字を返す。
//
// 場に同ランクの札が 1 枚でもあれば出せる。「場札の一番上」ではなく**場のどこか**
// と合えばよい。
func (l *LaughAndLieDown) GetValidPlayIndices(player int) []int {
	p := l.GetPlayer(player)
	if p == nil {
		return nil
	}
	var out []int
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		if l.layoutCountOfRank(c.GetValue()) > 0 {
			out = append(out, i)
		}
	}
	return out
}

// CanTakeThree は handIdx の札で 3 枚取りができるかを返す。
//
// 原典は「1 枚で**1 枚または 3 枚**を取る」。3 枚取りは場に 3 枚以上あるときだけ
// 選べて、取得枚数を伸ばせる代わりに他家の取り札も減らす。
func (l *LaughAndLieDown) CanTakeThree(player, handIdx int) bool {
	p := l.GetPlayer(player)
	if p == nil || handIdx < 0 || handIdx >= p.GetCardsSize() {
		return false
	}
	c := p.GetCard(handIdx)
	if c == nil {
		return false
	}
	return l.layoutCountOfRank(c.GetValue()) >= 3
}

// PlayCard は player が手札 handIdx の札を出し、場から takeCount 枚を取る。
//
// takeCount は 1 か 3 のみ。3 は場に 3 枚以上あるときだけ。
func (l *LaughAndLieDown) PlayCard(player, handIdx, takeCount int) error {
	if l.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if player != l.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if takeCount != 1 && takeCount != 3 {
		return fmt.Errorf("you may take one or three cards, not %d", takeCount)
	}
	p := l.GetPlayer(player)
	if p == nil {
		return fmt.Errorf("no such player: %d", player)
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := p.GetCard(handIdx)
	if card == nil {
		return fmt.Errorf("card index %d is empty", handIdx)
	}
	rank := card.GetValue()
	available := l.layoutCountOfRank(rank)
	if available == 0 {
		return fmt.Errorf("no card of that rank is on the table")
	}
	if takeCount == 3 && available < 3 {
		return fmt.Errorf("only %d card(s) of that rank are on the table", available)
	}

	p.RemoveCard(handIdx)
	taken := l.removeFromLayout(rank, takeCount)
	l.won[player] = append(l.won[player], card)
	l.won[player] = append(l.won[player], taken...)
	l.addLog(player, "capture", fmt.Sprintf("captures %d card(s)", len(taken)+1), append([]*Card{card}, taken...))

	l.advance()
	return nil
}

// removeFromLayout は場から rank の札を n 枚取り除いて返す。
func (l *LaughAndLieDown) removeFromLayout(rank, n int) []*Card {
	taken := make([]*Card, 0, n)
	rest := make([]*Card, 0, len(l.layout))
	for _, c := range l.layout {
		if c != nil && c.GetValue() == rank && len(taken) < n {
			taken = append(taken, c)
			continue
		}
		rest = append(rest, c)
	}
	l.layout = rest
	return taken
}

// advance は手番を進め、取れない人を降ろし、終了条件を見る。
func (l *LaughAndLieDown) advance() {
	l.currentIdx = (l.currentIdx + 1) % len(l.players)
	l.skipPlayersWhoCannotCapture()
}

// skipPlayersWhoCannotCapture は、手番の人が取れないあいだ降ろし続ける。
//
// 「取れなければ降りる」は選択ではなく強制なので、ボタンを出さずここで解決する。
// **降りた人の手札は場に積まれ、他家が取れるようになる** —— 単なる脱落ではなく、
// 場に獲物を撒く行為である。
func (l *LaughAndLieDown) skipPlayersWhoCannotCapture() {
	for range len(l.players) * 2 {
		if l.checkGameEnd() {
			return
		}
		p := l.GetPlayer(l.currentIdx)
		if p == nil {
			return
		}
		if p.GetCardsSize() > 0 && len(l.GetValidPlayIndices(l.currentIdx)) > 0 {
			return
		}
		if p.GetCardsSize() > 0 {
			l.lieDown(l.currentIdx)
		}
		l.currentIdx = (l.currentIdx + 1) % len(l.players)
	}
}

// lieDown は player の手札を全部場に置いて降ろす。
func (l *LaughAndLieDown) lieDown(player int) {
	p := l.GetPlayer(player)
	if p == nil {
		return
	}
	thrown := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if c := p.GetCard(i); c != nil {
			thrown = append(thrown, c)
		}
	}
	p.Reset()
	l.layout = append(l.layout, thrown...)
	l.laidDown[player] = true
	p.SetIsFinished(true)
	l.addLog(player, "liedown", fmt.Sprintf("lies down, adding %d card(s) to the table", len(thrown)), thrown)
}

// playersHoldingCards は手札が残っている人数と、その最後の 1 人を返す。
func (l *LaughAndLieDown) playersHoldingCards() (int, int) {
	n, last := 0, -1
	for i, p := range l.players {
		if p.GetCardsSize() > 0 {
			n++
			last = i
		}
	}
	return n, last
}

// checkGameEnd は「手札を持つ人が 1 人以下」なら終局させる。
func (l *LaughAndLieDown) checkGameEnd() bool {
	if l.gameEndFlag {
		return true
	}
	n, last := l.playersHoldingCards()
	if n > 1 {
		return false
	}
	l.finish(last)
	return true
}

// finish は残り札を親へ渡し、精算する。
func (l *LaughAndLieDown) finish(lastIn int) {
	l.lastInIdx = lastIn

	// 最後の 1 人の手札と場の残りは**親の取り札**へ。親だけ 1 多く出している
	// ことの埋め合わせで、原典どおり。
	residue := make([]*Card, 0, len(l.layout)+LaughAndLieDownHandSize)
	if lastIn >= 0 {
		p := l.GetPlayer(lastIn)
		for i := range p.GetCardsSize() {
			if c := p.GetCard(i); c != nil {
				residue = append(residue, c)
			}
		}
		p.Reset()
		p.SetIsFinished(true)
	}
	residue = append(residue, l.layout...)
	l.layout = nil
	l.won[l.dealerIdx] = append(l.won[l.dealerIdx], residue...)
	if len(residue) > 0 {
		l.addLog(l.dealerIdx, "residue", fmt.Sprintf("takes the %d leftover card(s)", len(residue)), residue)
	}

	l.settle()
	l.phase = LaughAndLieDownPhaseGameEnd
	l.gameEndFlag = true
}

// settle は精算する。掛け金を引き、最後の 1 人に 5、残りは取得枚数の 8 との
// 過不足を 2 枚 1 で清算する。
func (l *LaughAndLieDown) settle() {
	for i := range l.players {
		ante := LaughAndLieDownAnte
		if i == l.dealerIdx {
			ante = LaughAndLieDownDealerAnte
		}
		score := -ante
		if i == l.lastInIdx {
			score += LaughAndLieDownLastInBonus
		}
		// 「8 枚に対して 2 枚ごとに 1」。奇数枚の端数は切り捨てる (原典どおり
		// ポットに端数が残ることがある)。
		score += (len(l.won[i]) - LaughAndLieDownHandSize) / 2
		l.scores[i] = score
	}
}

// ---- CPU ----

// LaughAndLieDownCpuAction は CPU が選んだ手。
type LaughAndLieDownCpuAction struct {
	// HandIdx は出す札の手札添字 (出せないときは -1)。
	HandIdx int
	// TakeCount は場から取る枚数 (1 か 3)。
	TakeCount int
}

// LaughAndLieDownCpuDecide は idx の CPU が取る手を決める。
//
// 取得枚数がそのまま精算になるので、**3 枚取れるならそれを選ぶ**。同じ 1 枚取りの
// 中では、場の残り枚数が少ないランクから片付ける —— 残しておくと降りた人の手札で
// 増える可能性がある側を先に確保する。
func (l *LaughAndLieDown) LaughAndLieDownCpuDecide(idx int) LaughAndLieDownCpuAction {
	valid := l.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return LaughAndLieDownCpuAction{HandIdx: -1, TakeCount: 1}
	}
	p := l.GetPlayer(idx)
	best, bestCount := -1, 0
	for _, i := range valid {
		n := l.layoutCountOfRank(p.GetCard(i).GetValue())
		if n >= 3 {
			return LaughAndLieDownCpuAction{HandIdx: i, TakeCount: 3}
		}
		if best == -1 || n < bestCount {
			best, bestCount = i, n
		}
	}
	return LaughAndLieDownCpuAction{HandIdx: best, TakeCount: 1}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (l *LaughAndLieDown) GetPlayers() []*LaughAndLieDownPlayer { return l.players }

// GetPlayer は idx のプレイヤーを返す。
func (l *LaughAndLieDown) GetPlayer(idx int) *LaughAndLieDownPlayer {
	if idx < 0 || idx >= len(l.players) {
		return nil
	}
	return l.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (l *LaughAndLieDown) GetPhase() LaughAndLieDownPhase { return l.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (l *LaughAndLieDown) GetCurrentPlayerIdx() int { return l.currentIdx }

// GetLayout は表向きの場札を返す。
func (l *LaughAndLieDown) GetLayout() []*Card { return l.layout }

// GetWonCount は idx の取得枚数を返す。
func (l *LaughAndLieDown) GetWonCount(idx int) int {
	if idx < 0 || idx >= len(l.won) {
		return 0
	}
	return len(l.won[idx])
}

// IsLaidDown は idx が降りているかを返す。
func (l *LaughAndLieDown) IsLaidDown(idx int) bool {
	if idx < 0 || idx >= len(l.laidDown) {
		return false
	}
	return l.laidDown[idx]
}

// GetScore は idx の収支を返す。
func (l *LaughAndLieDown) GetScore(idx int) int {
	if idx < 0 || idx >= len(l.scores) {
		return 0
	}
	return l.scores[idx]
}

// GetDealerIdx は親の添字を返す。
func (l *LaughAndLieDown) GetDealerIdx() int { return l.dealerIdx }

// GetLastInIdx は最後まで手札が残っていた人の添字を返す (-1: 該当なし)。
func (l *LaughAndLieDown) GetLastInIdx() int { return l.lastInIdx }

// GetGameEndFlag は終局しているかを返す。
func (l *LaughAndLieDown) GetGameEndFlag() bool { return l.gameEndFlag }

// GetConfig はゲーム設定を返す。
func (l *LaughAndLieDown) GetConfig() LaughAndLieDownConfig { return l.config }

// SetConfig はゲーム設定をセットする。
func (l *LaughAndLieDown) SetConfig(c LaughAndLieDownConfig) { l.config = c }

// GetActionLog は棋譜を返す。
func (l *LaughAndLieDown) GetActionLog() []*ActionLogEntry { return l.actionLog }

// SetLayoutForTest はテスト用に場札を差し替える。
func (l *LaughAndLieDown) SetLayoutForTest(cards []*Card) { l.layout = cards }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (l *LaughAndLieDown) SetCurrentPlayerForTest(idx int) { l.currentIdx = idx }

// addLog は棋譜に 1 件追加する。
func (l *LaughAndLieDown) addLog(player int, action, detail string, cards []*Card) {
	l.actionLog = append(l.actionLog, &ActionLogEntry{
		TurnNumber: len(l.actionLog) + 1,
		PlayerIdx:  player,
		ActionType: action,
		Detail:     detail,
		Cards:      cards,
	})
}

// laughAndLieDownJSON is the JSON wire format for LaughAndLieDown.
type laughAndLieDownJSON struct {
	Players   []*LaughAndLieDownPlayer `json:"pl"`
	Config    LaughAndLieDownConfig    `json:"cfg"`
	Phase     LaughAndLieDownPhase     `json:"ph"`
	Layout    []*Card                  `json:"lo"`
	Won       [][]*Card                `json:"wn"`
	LaidDown  []bool                   `json:"ld"`
	Current   int                      `json:"cur"`
	Dealer    int                      `json:"dl"`
	LastIn    int                      `json:"li"`
	Scores    []int                    `json:"sc"`
	GameEnd   bool                     `json:"ge"`
	ActionLog []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (l *LaughAndLieDown) MarshalJSON() ([]byte, error) {
	return json.Marshal(laughAndLieDownJSON{
		Players: l.players, Config: l.config, Phase: l.phase, Layout: l.layout,
		Won: l.won, LaidDown: l.laidDown, Current: l.currentIdx, Dealer: l.dealerIdx,
		LastIn: l.lastInIdx, Scores: l.scores, GameEnd: l.gameEndFlag, ActionLog: l.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を
// 検証する。長さの揃っていないスライスをそのまま使うと添字外れで落ちる。
func (l *LaughAndLieDown) UnmarshalJSON(data []byte) error {
	var raw laughAndLieDownJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != LaughAndLieDownPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", LaughAndLieDownPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < LaughAndLieDownPhasePlay || raw.Phase > LaughAndLieDownPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	l.players = raw.Players
	l.config = raw.Config
	l.phase = raw.Phase
	l.layout = raw.Layout
	l.gameEndFlag = raw.GameEnd
	l.actionLog = raw.ActionLog

	l.won = make([][]*Card, len(l.players))
	copy(l.won, raw.Won)
	for i := range l.won {
		if l.won[i] == nil {
			l.won[i] = make([]*Card, 0, 8)
		}
	}
	l.laidDown = make([]bool, len(l.players))
	copy(l.laidDown, raw.LaidDown)
	l.scores = make([]int, len(l.players))
	copy(l.scores, raw.Scores)

	l.currentIdx = clampLldIdx(raw.Current, len(l.players))
	l.dealerIdx = clampLldIdx(raw.Dealer, len(l.players))
	l.lastInIdx = raw.LastIn
	if l.lastInIdx < -1 || l.lastInIdx >= len(l.players) {
		l.lastInIdx = -1
	}
	return nil
}

// clampLldIdx は席番号を 0..n-1 に収める。
func clampLldIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}
