//go:build !js || !wasm || extra3

// Package domain — トゥーペン (Toepen) のドメインモデル。
//
// オランダのパブで遊ばれるトリックテイキング。32 枚デッキ、手札 4 枚、切札なし。
// 各トリックはリードスートの最強札が取り、**最終トリックを取った者だけが失点を
// 免れる**。累計 10 失点で脱落し、最後に残った者が勝つ。
//
// # 序列 — このゲームの核心
//
//	10 (最強) > 9 > 8 > 7 > A > K > Q > J (最弱)
//
// **数札が A より強く、絵札が最弱**という、標準的な序列を丸ごと入れ替えた構成が
// トゥーペンの特徴である。issue #4411 は「J を 2 番目に高い札」としているが誤りで、
// J は**最弱**、**10 が最強**。ここを取り違えるとゲームが成立しないので、
// toepenRankOrder を単一の出典とし、TestToepen_RankOrder が全順序を固定している。
//
// # toep (ノック)
//
// 手番に関係なくいつでも「toep」を宣言でき、賭け点が 1 増える。宣言を受けた側は
// 追随するか降りるかを選ぶ。降りた者はその時点の賭け点を即座に失点として確定し、
// そのハンドから抜ける。最後まで残った者は「ノック回数 + 1」を失点する
// (最終トリックの勝者を除く)。
//
// # 貧民 (armoede) の配り直し
//
// 手札が A・K・Q・J だけ、つまり**最弱の 4 枚しか無い**場合、捨てて配り直しを
// 要求できる (Redeal)。issue はこのルールに触れていない。
//
// 卓上のルールでは他家が異議を唱えられ、外れた側がライフを失う。**本実装に異議の
// 手続きは無い。** 異議が存在するのは、人間同士では宣言が本当かどうか確かめようが
// ないからで、サーバーが手札を持っている以上そこは検証で済む。Redeal は貧民の
// ときだけ通り、嘘の宣言はそもそも成立しない。異議を残すと「必ず外れる異議」を
// 押せるだけのボタンになる。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// ToepenPlayerCnt は既定のプレイヤー数。
const ToepenPlayerCnt = 4

// ToepenMinPlayers / ToepenMaxPlayers は許容するプレイヤー数の範囲。
// pagat は 3〜8 人としている (issue の 3〜5 は狭い)。
const (
	ToepenMinPlayers = 3
	ToepenMaxPlayers = 8
)

// ToepenHandSize は配札枚数。
const ToepenHandSize = 4

// ToepenDeckSize は 32 枚 (7〜A + 絵札)。
const ToepenDeckSize = 32

// ToepenMaxLives は脱落するまでの失点。
const ToepenMaxLives = 10

// toepenValues は 32 枚デッキに含まれるカード値。
// A(1), 7, 8, 9, 10, J(11), Q(12), K(13)。
var toepenValues = []int{1, 7, 8, 9, 10, 11, 12, 13}

// toepenRankOrder はスート内の強さ。値が大きいほど強い。
//
// **10 > 9 > 8 > 7 > A > K > Q > J。** 数札が A を上回り、絵札が最弱という
// 入れ替わった序列がトゥーペンの正体で、ここが唯一の出典。
var toepenRankOrder = map[int]int{
	11: 1, // J — 最弱
	12: 2, // Q
	13: 3, // K
	1:  4, // A
	7:  5,
	8:  6,
	9:  7,
	10: 8, // 10 — 最強
}

// ToepenRankOrder はカードのスート内順位を返す (大きいほど強い)。nil は 0。
func ToepenRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	return toepenRankOrder[c.GetValue()]
}

// ToepenIsPoverty は手札が最弱 4 種 (A/K/Q/J) だけで構成されているかを返す。
//
// 「貧民」の判定であり、配り直しを要求できる条件。数札を 1 枚でも含めば偽。
func ToepenIsPoverty(hand []*Card) bool {
	if len(hand) == 0 {
		return false
	}
	for _, c := range hand {
		if c == nil {
			return false
		}
		switch c.GetValue() {
		case 1, 11, 12, 13:
		default:
			return false
		}
	}
	return true
}

// ToepenPhase はゲームフェーズ。
type ToepenPhase int

// Toepenのフェーズ定数
const (
	// ToepenPhasePlay 札を出すフェーズ
	ToepenPhasePlay ToepenPhase = iota
	// ToepenPhaseRespond toep を受けて追随/降参を選ぶフェーズ
	ToepenPhaseRespond
	// ToepenPhaseHandEnd ハンド終了 (精算済み)
	ToepenPhaseHandEnd
	// ToepenPhaseGameEnd 終局
	ToepenPhaseGameEnd
)

// newToepenDeck は 32 枚を生成する (シャッフル前)。
func newToepenDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, ToepenDeckSize)
	for _, s := range suits {
		for _, v := range toepenValues {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// toepenShuffle は Fisher-Yates。domain の shuffleCards は casino タグのファイルに
// あり extra3 ビルドから見えないため、専用名で持つ。
func toepenShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// Toepen はトゥーペンのゲームクラス。
type Toepen struct {
	players      []*ToepenPlayer
	config       ToepenConfig
	phase        ToepenPhase
	lives        []int
	folded       []bool
	eliminated   []bool
	trick        []*TrickCard
	leadSuit     int
	currentIdx   int
	leadIdx      int
	dealerIdx    int
	trickNumber  int
	stake        int // 賭け点 = ノック回数 + 1
	knockerIdx   int // toep を宣言した者 (応答フェーズ中のみ有効、それ以外 -1)
	pendingIdx   int // 応答待ちのプレイヤー (-1 なら無し)
	lastTrickWin int
	handNumber   int
	gameEndFlag  bool
	winnerIdx    int
	actionLog    []*ActionLogEntry
}

// NewToepen はコンストラクタ。
func NewToepen(players []*ToepenPlayer, config ToepenConfig) *Toepen {
	return &Toepen{
		players:      players,
		config:       config,
		lives:        make([]int, len(players)),
		folded:       make([]bool, len(players)),
		eliminated:   make([]bool, len(players)),
		knockerIdx:   -1,
		pendingIdx:   -1,
		lastTrickWin: -1,
		winnerIdx:    -1,
	}
}

// NewDefaultToepen は標準の 4 人セットアップを返す。
func NewDefaultToepen() *Toepen {
	players := make([]*ToepenPlayer, 0, ToepenPlayerCnt)
	players = append(players, NewToepenPlayer(true))
	for range ToepenPlayerCnt - 1 {
		players = append(players, NewToepenPlayer(false))
	}
	return NewToepen(players, DefaultToepenConfig())
}

// Reset はゲーム全体を初期化する。
func (t *Toepen) Reset() {
	t.lives = make([]int, len(t.players))
	t.eliminated = make([]bool, len(t.players))
	t.handNumber = 0
	t.dealerIdx = 0
	t.gameEndFlag = false
	t.winnerIdx = -1
	t.actionLog = nil
	t.startHand()
}

// startHand は 1 ハンドを配り直す。
func (t *Toepen) startHand() {
	t.phase = ToepenPhasePlay
	t.trick = nil
	t.leadSuit = -1
	t.trickNumber = 0
	t.stake = 1
	t.knockerIdx = -1
	t.pendingIdx = -1
	t.lastTrickWin = -1
	t.folded = make([]bool, len(t.players))
	for _, p := range t.players {
		p.ResetGame()
	}

	deck := newToepenDeck()
	toepenShuffle(deck)
	pos := 0
	for range ToepenHandSize {
		for i, p := range t.players {
			if t.eliminated[i] {
				continue
			}
			p.AddCard(deck[pos])
			pos++
		}
	}

	t.leadIdx = t.nextActive(t.dealerIdx)
	t.currentIdx = t.leadIdx
	t.handNumber++
	t.addLog(-1, "deal", fmt.Sprintf("hand %d dealt", t.handNumber), nil)
}

// nextActive は idx の次の、まだこのハンドに参加しているプレイヤーを返す。
func (t *Toepen) nextActive(idx int) int {
	n := len(t.players)
	for i := 1; i <= n; i++ {
		j := (idx + i) % n
		if !t.eliminated[j] && !t.folded[j] {
			return j
		}
	}
	return -1
}

// activeCount はこのハンドに残っている人数を返す。
func (t *Toepen) activeCount() int {
	c := 0
	for i := range t.players {
		if !t.eliminated[i] && !t.folded[i] {
			c++
		}
	}
	return c
}

// GetValidPlayIndices は player が出せる手札の添字を返す。
//
// **フォロー義務あり**: リードスートを持っていればそれしか出せない。切札は無い。
func (t *Toepen) GetValidPlayIndices(player int) []int {
	p := t.GetPlayer(player)
	if p == nil {
		return nil
	}
	var follow, all []int
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		all = append(all, i)
		if t.leadSuit >= 0 && c.GetDesign() == t.leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// PlayCard は player が手札 handIdx の札を出す。
func (t *Toepen) PlayCard(player, handIdx int) error {
	if t.gameEndFlag || t.phase == ToepenPhaseHandEnd {
		return fmt.Errorf("the hand is over")
	}
	if t.phase == ToepenPhaseRespond {
		return fmt.Errorf("a toep is pending a response")
	}
	if player != t.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := t.GetPlayer(player)
	if p == nil {
		return fmt.Errorf("no such player: %d", player)
	}
	valid := t.GetValidPlayIndices(player)
	if !toepenContains(valid, handIdx) {
		return fmt.Errorf("card index %d is not playable; you must follow suit", handIdx)
	}

	card := p.RemoveCard(handIdx)
	if card == nil {
		return fmt.Errorf("card index %d is empty", handIdx)
	}
	if t.leadSuit < 0 {
		t.leadSuit = card.GetDesign()
	}
	t.trick = append(t.trick, &TrickCard{PlayerIdx: player, Card: card})
	t.addLog(player, "play", "plays a card", []*Card{card})

	next := t.nextActive(player)
	if len(t.trick) >= t.activeCount() {
		t.resolveTrick()
		return nil
	}
	t.currentIdx = next
	return nil
}

// toepenContains はスライスに値が含まれるかを返す。domain には同名の containsInt が
// あるが casino タグのファイル (OpenFaceChinese.go) にあり、extra3 ビルドから見えない
// うえ非 WASM ビルドでは衝突する。
func toepenContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// resolveTrick はトリックを解決する。切札が無いので、リードスートの最強札が取る。
func (t *Toepen) resolveTrick() {
	winner := -1
	best := -1
	for _, tc := range t.trick {
		if tc.Card == nil || tc.Card.GetDesign() != t.leadSuit {
			continue
		}
		if r := ToepenRankOrder(tc.Card); r > best {
			best, winner = r, tc.PlayerIdx
		}
	}
	if winner < 0 && len(t.trick) > 0 {
		winner = t.trick[0].PlayerIdx
	}

	t.trickNumber++
	t.lastTrickWin = winner
	t.addLog(winner, "trick", fmt.Sprintf("wins trick %d", t.trickNumber), nil)
	t.trick = nil
	t.leadSuit = -1
	t.leadIdx = winner
	t.currentIdx = winner

	if t.trickNumber >= ToepenHandSize || t.handExhausted() {
		t.finishHand()
	}
}

// handExhausted は誰かの手札が尽きたかを返す。
func (t *Toepen) handExhausted() bool {
	for i, p := range t.players {
		if t.eliminated[i] || t.folded[i] {
			continue
		}
		if p.GetCardsSize() == 0 {
			return true
		}
	}
	return false
}

// Redeal は貧民 (最弱 4 種のみの手札) の player が手札を捨てて配り直しを要求する。
//
// サーバーが手札を検証するので、卓上ルールの「異議」は不要 —— 嘘の宣言は
// そもそも通らない。賭け点と親はそのままで、同じハンドを配り直す。
func (t *Toepen) Redeal(player int) error {
	if t.gameEndFlag || t.phase != ToepenPhasePlay {
		return fmt.Errorf("cannot ask for a redeal right now")
	}
	if t.trickNumber > 0 || len(t.trick) > 0 {
		return fmt.Errorf("a redeal must be asked for before any card is played")
	}
	p := t.GetPlayer(player)
	if p == nil || t.eliminated[player] || t.folded[player] {
		return fmt.Errorf("player %d is not in this hand", player)
	}
	hand := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		hand = append(hand, p.GetCard(i))
	}
	if !ToepenIsPoverty(hand) {
		return fmt.Errorf("a redeal needs a hand of nothing but A, K, Q and J")
	}

	t.addLog(player, "redeal", "asks for a redeal on a poverty hand", nil)
	// 同じ親でハンド番号も据え置いたまま配り直す。startHand は番号を進めるので
	// 戻しておく -- 配り直しは新しいハンドではない。
	t.handNumber--
	t.startHand()
	return nil
}

// CanRedeal は player が配り直しを要求できるかを返す。
func (t *Toepen) CanRedeal(player int) bool {
	if t.gameEndFlag || t.phase != ToepenPhasePlay || t.trickNumber > 0 || len(t.trick) > 0 {
		return false
	}
	p := t.GetPlayer(player)
	if p == nil || t.eliminated[player] || t.folded[player] {
		return false
	}
	hand := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		hand = append(hand, p.GetCard(i))
	}
	return ToepenIsPoverty(hand)
}

// Toep は player が賭け点を吊り上げる。
//
// 宣言者以外の参加者は追随か降参かを選ぶ。連続でノックできないよう、応答が
// 済むまで次のノックは受け付けない。
func (t *Toepen) Toep(player int) error {
	if t.gameEndFlag || t.phase != ToepenPhasePlay {
		return fmt.Errorf("cannot toep right now")
	}
	if player < 0 || player >= len(t.players) || t.eliminated[player] || t.folded[player] {
		return fmt.Errorf("player %d is not in this hand", player)
	}
	if t.activeCount() < 2 {
		return fmt.Errorf("nobody left to raise against")
	}
	t.stake++
	t.knockerIdx = player
	t.phase = ToepenPhaseRespond
	t.pendingIdx = t.nextActive(player)
	t.addLog(player, "toep", fmt.Sprintf("toeps; the stake is now %d", t.stake), nil)
	return nil
}

// Respond は toep を受けた player が追随 (stay) か降参 (fold) かを答える。
func (t *Toepen) Respond(player int, stay bool) error {
	if t.phase != ToepenPhaseRespond {
		return fmt.Errorf("no toep is pending")
	}
	if player != t.pendingIdx {
		return fmt.Errorf("it is not player %d's response", player)
	}

	if !stay {
		// 降りた者は現在の賭け点から 1 引いた額を失う。ノックに乗らなかったので
		// 直近の吊り上げぶんは負わない。
		t.folded[player] = true
		t.loseLives(player, t.stake-1)
		t.addLog(player, "fold", fmt.Sprintf("folds for %d", t.stake-1), nil)
		// 降りた者が手番だったなら手番を進める。
		if t.currentIdx == player {
			t.currentIdx = t.nextActive(player)
		}
		// 場に出ている札のうち降りた者のぶんは残す (トリックの解決には影響しない)。
	} else {
		t.addLog(player, "stay", "stays in", nil)
	}

	t.pendingIdx = t.nextRespondent(player)
	if t.pendingIdx >= 0 {
		return nil
	}

	t.phase = ToepenPhasePlay
	t.knockerIdx = -1
	if t.activeCount() <= 1 {
		t.finishHand()
		return nil
	}
	if t.currentIdx < 0 || t.folded[t.currentIdx] || t.eliminated[t.currentIdx] {
		t.currentIdx = t.nextActive(t.currentIdx)
	}
	return nil
}

// nextRespondent は player の次に応答すべき者を返す (無ければ -1)。
func (t *Toepen) nextRespondent(player int) int {
	n := len(t.players)
	for i := 1; i <= n; i++ {
		j := (player + i) % n
		if j == t.knockerIdx {
			return -1
		}
		if !t.eliminated[j] && !t.folded[j] {
			return j
		}
	}
	return -1
}

// finishHand はハンドを精算する。
//
// 免れるのは**そのハンドの勝者**ひとりだけ。通常は最終トリックを取った者だが、
// **toep で全員が降りてハンドが終わった場合は、残った 1 人が勝者**である。
// lastTrickWin だけを見ると、この場合それは -1 のままなので誰も免れず、生き残った
// 側まで賭け点を払うことになる。それでは「相手を降ろして勝つ」ことと「負ける」ことが
// 同コストになり、toep を打つ理由そのものが消える。
func (t *Toepen) finishHand() {
	t.phase = ToepenPhaseHandEnd
	t.pendingIdx = -1
	t.knockerIdx = -1

	winner := t.handWinner()
	for i := range t.players {
		if t.eliminated[i] || t.folded[i] {
			continue
		}
		if i == winner {
			continue
		}
		t.loseLives(i, t.stake)
	}
	t.addLog(-1, "hand", fmt.Sprintf("hand %d settled for %d", t.handNumber, t.stake), nil)

	remaining := 0
	last := -1
	for i := range t.players {
		if !t.eliminated[i] {
			remaining++
			last = i
		}
	}
	if remaining <= 1 {
		t.gameEndFlag = true
		t.phase = ToepenPhaseGameEnd
		t.winnerIdx = last
		t.addLog(-1, "game", "game over", nil)
	}
}

// handWinner はこのハンドで失点を免れる者を返す。
//
// 最終トリックを取った者。トリックが一度も成立しないまま (全員が降りて) ハンドが
// 終わったときは、残っている唯一のプレイヤー。どちらでもなければ -1。
func (t *Toepen) handWinner() int {
	if t.lastTrickWin >= 0 {
		return t.lastTrickWin
	}
	survivor := -1
	for i := range t.players {
		if t.eliminated[i] || t.folded[i] {
			continue
		}
		if survivor >= 0 {
			return -1 // 2 人以上残っているなら勝者は未確定
		}
		survivor = i
	}
	return survivor
}

// loseLives は player の失点を加算し、上限に達したら脱落させる。
func (t *Toepen) loseLives(player, n int) {
	if player < 0 || player >= len(t.players) || n <= 0 {
		return
	}
	t.lives[player] += n
	if t.lives[player] >= ToepenMaxLives {
		t.lives[player] = ToepenMaxLives
		t.eliminated[player] = true
		t.addLog(player, "out", "is eliminated", nil)
	}
}

// NextHand は次のハンドを開始する。
func (t *Toepen) NextHand() error {
	if t.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if t.phase != ToepenPhaseHandEnd {
		return fmt.Errorf("the hand is still in progress")
	}
	t.dealerIdx = t.nextSeat(t.dealerIdx)
	t.startHand()
	return nil
}

// nextSeat は脱落者を飛ばして次の席を返す。
func (t *Toepen) nextSeat(idx int) int {
	n := len(t.players)
	for i := 1; i <= n; i++ {
		j := (idx + i) % n
		if !t.eliminated[j] {
			return j
		}
	}
	return idx
}

// addLog は棋譜へ 1 行追加する。
func (t *Toepen) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	t.actionLog = append(t.actionLog, &ActionLogEntry{
		TurnNumber: len(t.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (t *Toepen) GetPlayers() []*ToepenPlayer { return t.players }

// GetPlayer は idx のプレイヤーを返す。範囲外は nil。
func (t *Toepen) GetPlayer(idx int) *ToepenPlayer {
	if idx < 0 || idx >= len(t.players) {
		return nil
	}
	return t.players[idx]
}

// GetLives は idx の累計失点を返す。
func (t *Toepen) GetLives(idx int) int {
	if idx < 0 || idx >= len(t.lives) {
		return 0
	}
	return t.lives[idx]
}

// SetLives は idx の累計失点を設定する (テスト用)。
func (t *Toepen) SetLives(idx, lives int) {
	if idx < 0 || idx >= len(t.lives) {
		return
	}
	t.lives[idx] = lives
}

// IsFolded は idx がこのハンドから降りたかを返す。
func (t *Toepen) IsFolded(idx int) bool {
	return idx >= 0 && idx < len(t.folded) && t.folded[idx]
}

// IsEliminated は idx が脱落したかを返す。
func (t *Toepen) IsEliminated(idx int) bool {
	return idx >= 0 && idx < len(t.eliminated) && t.eliminated[idx]
}

// GetPhase は現在のフェーズを返す。
func (t *Toepen) GetPhase() ToepenPhase { return t.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (t *Toepen) GetCurrentPlayerIdx() int { return t.currentIdx }

// GetLeadPlayerIdx はリード側のプレイヤー添字を返す。
func (t *Toepen) GetLeadPlayerIdx() int { return t.leadIdx }

// GetDealerIdx は親の添字を返す。
func (t *Toepen) GetDealerIdx() int { return t.dealerIdx }

// GetCurrentTrick は場に出ている札を返す。
func (t *Toepen) GetCurrentTrick() []*TrickCard { return t.trick }

// GetLeadSuit はリードスートを返す (未決なら -1)。
func (t *Toepen) GetLeadSuit() int { return t.leadSuit }

// GetTrickNumber は完了したトリック数を返す。
func (t *Toepen) GetTrickNumber() int { return t.trickNumber }

// GetStake は現在の賭け点を返す。
func (t *Toepen) GetStake() int { return t.stake }

// GetKnockerIdx は toep を宣言した者を返す (応答フェーズ外は -1)。
func (t *Toepen) GetKnockerIdx() int { return t.knockerIdx }

// GetPendingRespondent は応答待ちのプレイヤーを返す (無ければ -1)。
func (t *Toepen) GetPendingRespondent() int { return t.pendingIdx }

// GetLastTrickWinner は最後にトリックを取った者を返す (-1 は未確定)。
func (t *Toepen) GetLastTrickWinner() int { return t.lastTrickWin }

// GetHandNumber は現在のハンド番号 (1 始まり) を返す。
func (t *Toepen) GetHandNumber() int { return t.handNumber }

// GetGameEndFlag は終局しているかを返す。
func (t *Toepen) GetGameEndFlag() bool { return t.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す。未確定は -1。
func (t *Toepen) GetWinnerIdx() int { return t.winnerIdx }

// GetConfig はゲーム設定を返す。
func (t *Toepen) GetConfig() ToepenConfig { return t.config }

// SetConfig はゲーム設定を差し替える。
func (t *Toepen) SetConfig(c ToepenConfig) { t.config = c }

// GetActionLog は棋譜を返す。
func (t *Toepen) GetActionLog() []*ActionLogEntry { return t.actionLog }

// ---- JSON ----

// toepenJSON は KV のワイヤ形式。Worker は毎リクエストここから組み直すので、
// ここに無いものは次のリクエストでは存在しない。
type toepenJSON struct {
	Players      []*ToepenPlayer   `json:"pl"`
	Config       ToepenConfig      `json:"cf"`
	Phase        ToepenPhase       `json:"ph"`
	Lives        []int             `json:"lv"`
	Folded       []bool            `json:"fd"`
	Eliminated   []bool            `json:"el"`
	Trick        []*TrickCard      `json:"tk"`
	LeadSuit     int               `json:"ls"`
	CurrentIdx   int               `json:"ci"`
	LeadIdx      int               `json:"li"`
	DealerIdx    int               `json:"di"`
	TrickNumber  int               `json:"tn"`
	Stake        int               `json:"sk"`
	KnockerIdx   int               `json:"ki"`
	PendingIdx   int               `json:"pi"`
	LastTrickWin int               `json:"lw"`
	HandNumber   int               `json:"hn"`
	GameEndFlag  bool              `json:"ge"`
	WinnerIdx    int               `json:"wi"`
	ActionLog    []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (t *Toepen) MarshalJSON() ([]byte, error) {
	return json.Marshal(toepenJSON{
		Players:      t.players,
		Config:       t.config,
		Phase:        t.phase,
		Lives:        t.lives,
		Folded:       t.folded,
		Eliminated:   t.eliminated,
		Trick:        t.trick,
		LeadSuit:     t.leadSuit,
		CurrentIdx:   t.currentIdx,
		LeadIdx:      t.leadIdx,
		DealerIdx:    t.dealerIdx,
		TrickNumber:  t.trickNumber,
		Stake:        t.stake,
		KnockerIdx:   t.knockerIdx,
		PendingIdx:   t.pendingIdx,
		LastTrickWin: t.lastTrickWin,
		HandNumber:   t.handNumber,
		GameEndFlag:  t.gameEndFlag,
		WinnerIdx:    t.winnerIdx,
		ActionLog:    t.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// Worker はこれを KV の未検証バイト列に対して毎リクエスト実行する。添字は信用せず
// 丸め、真偽値スライスは人数ぶんに揃える。長さが足りないと後段の添字アクセスが
// 範囲外になる。
func (t *Toepen) UnmarshalJSON(data []byte) error {
	var j toepenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) == 0 {
		return fmt.Errorf("toepen: no players in snapshot")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("toepen: %w", err)
	}
	t.players = j.Players
	t.config = j.Config
	t.phase = j.Phase
	t.trick = j.Trick
	t.leadSuit = j.LeadSuit
	t.trickNumber = j.TrickNumber
	t.stake = j.Stake
	t.handNumber = j.HandNumber
	t.gameEndFlag = j.GameEndFlag
	t.actionLog = j.ActionLog

	n := len(t.players)
	t.currentIdx = toepenClampSeat(j.CurrentIdx, n)
	t.leadIdx = toepenClampSeat(j.LeadIdx, n)
	t.knockerIdx = toepenClampSeat(j.KnockerIdx, n)
	t.pendingIdx = toepenClampSeat(j.PendingIdx, n)
	t.lastTrickWin = toepenClampSeat(j.LastTrickWin, n)
	t.winnerIdx = toepenClampSeat(j.WinnerIdx, n)
	t.dealerIdx = toepenClampSeat(j.DealerIdx, n)
	if t.dealerIdx < 0 {
		t.dealerIdx = 0
	}
	if t.stake < 1 {
		t.stake = 1
	}

	t.lives = make([]int, n)
	copy(t.lives, j.Lives)
	t.folded = make([]bool, n)
	copy(t.folded, j.Folded)
	t.eliminated = make([]bool, n)
	copy(t.eliminated, j.Eliminated)
	return nil
}

// toepenClampSeat は範囲外のプレイヤー添字を -1 に丸める。
func toepenClampSeat(idx, n int) int {
	if idx < 0 || idx >= n {
		return -1
	}
	return idx
}

// ---- CPU ----

// ToepenCpuAction は CPU が選んだ手。
type ToepenCpuAction struct {
	// Toep が真なら賭け点を吊り上げる。
	Toep bool
	// Fold が真なら降りる (応答フェーズのみ)。
	Fold bool
	// HandIdx は出す札の手札添字 (それ以外は -1)。
	HandIdx int
}

// toepenStrongRank は「強い」と見なす下限順位。10/9 に相当する。
const toepenStrongRank = 7

// ToepenCpuDecide は idx の CPU が取る手を決める。
//
// 応答フェーズでは、強い札が残っていれば追随し、無ければ降りる。降りるコストは
// 現在の賭け点 −1 で、最後まで残って負けるより安いことがある。
//
// プレイフェーズでは、取れるなら最強の合法札で取りに行き、取れないなら最弱を捨てる。
// 最終トリックだけが失点を免れるので、最後のトリックは特に取りに行く。
func (t *Toepen) ToepenCpuDecide(idx int) ToepenCpuAction {
	p := t.GetPlayer(idx)
	if p == nil || p.GetCardsSize() == 0 {
		return ToepenCpuAction{HandIdx: -1}
	}

	if t.phase == ToepenPhaseRespond {
		return ToepenCpuAction{HandIdx: -1, Fold: !t.hasStrongCard(idx)}
	}

	valid := t.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return ToepenCpuAction{HandIdx: -1}
	}

	// 場の最強を上回れる最弱の札を探す。上回れないなら最弱を捨てる。
	best := t.currentTrickBest()
	winIdx, winRank := -1, 0
	lowIdx, lowRank := valid[0], 0
	for _, i := range valid {
		c := p.GetCard(i)
		r := ToepenRankOrder(c)
		if lowRank == 0 || r < lowRank {
			lowIdx, lowRank = i, r
		}
		if t.leadSuit >= 0 && c.GetDesign() != t.leadSuit {
			continue // 切札は無いので、リード以外は取れない
		}
		if r > best && (winIdx == -1 || r < winRank) {
			winIdx, winRank = i, r
		}
	}
	if winIdx >= 0 {
		return ToepenCpuAction{HandIdx: winIdx}
	}
	return ToepenCpuAction{HandIdx: lowIdx}
}

// hasStrongCard は idx が上位ランクの札を残しているかを返す。
func (t *Toepen) hasStrongCard(idx int) bool {
	p := t.GetPlayer(idx)
	if p == nil {
		return false
	}
	for i := range p.GetCardsSize() {
		if ToepenRankOrder(p.GetCard(i)) >= toepenStrongRank {
			return true
		}
	}
	return false
}

// currentTrickBest は場に出ている札のうちリードスート最強の順位を返す。
func (t *Toepen) currentTrickBest() int {
	best := 0
	for _, tc := range t.trick {
		if tc.Card == nil || tc.Card.GetDesign() != t.leadSuit {
			continue
		}
		if r := ToepenRankOrder(tc.Card); r > best {
			best = r
		}
	}
	return best
}
