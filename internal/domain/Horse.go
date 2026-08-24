//go:build !js || !wasm || casino

package domain

import (
	"errors"
	"fmt"
)

// HorsePhase はオーケストレータの進行段階。
//
// **種目の中の進行は各実装が持っている。** ここが持つのは「ハンドが動いて
// いるか、終わって次を待っているか」だけ。
type HorsePhase int

const (
	// HorsePhaseHand はハンドが進行中。
	HorsePhaseHand HorsePhase = iota
	// HorsePhaseHandEnd はハンドが決着して次を待っている。
	HorsePhaseHandEnd
	// HorsePhaseGameEnd は 1 人を残して全員のチップが尽きた。
	HorsePhaseGameEnd
)

// HorsePhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const HorsePhaseMax = HorsePhaseGameEnd

// horseMaxSliceLen は復元時に許すスライス長の上限。
const horseMaxSliceLen = 512

// エラー値 (`errors.Is` で判定するための番兵)。
//
// **画面に出るのはこの文字列ではない。** 返すのは `NewDomainErrorCode` で
// くるんだものにして、文言は i18n 側に持たせる ── 生の
// `horse: not allowed in this phase` が盤面の下に出ていた (実測)。
//
// **鍵は呼び出し側に literal で置く。** ヘルパ越しに渡すと
// `check-message-codes.mjs` の網に掛からない (あの正規表現は
// NewDomainErrorCode の第 2 引数がその場に書かれている形しか読まない) ので、
// 翻訳が無いまま鍵が画面に出ても誰も気付かない。
var (
	errHorseFinished   = errors.New("horse: game already finished")
	errHorseWrongPhase = errors.New("horse: not allowed in this phase")
	errHorseNoTable    = errors.New("horse: no hand in progress")
	errHorseNoDraw     = errors.New("horse: the current discipline has no draw")
)

// horseTable は H.O.R.S.E. が種目に対して必要とする操作だけを切り出したもの。
//
// **メソッドの形は同じでも、フェーズ番号は同じではない。** Holdem / Omaha /
// SevenCardStud はどれも `Reset() error` と
// `PlayerAction(action, amount, humanPlayMs int) error` を同じ並びで持つので、
// 呼び出しは 1 つのインタフェースにまとめられる。しかし**フェーズ定数を共有
// しているのは Omaha だけ**で、Stud は独自の並びを持つ:
//
//	Holdem/Omaha : ... Showdown=5, End=6, Rebuy=7
//	SevenCardStud: ... Showdown=6, End=7, Rebuy=8   (ストリートが 1 つ多い)
//
// 素直に `HoldemPhaseEnd` と比べると、**Stud のショーダウン中を「終わった」と
// 誤読して精算前の残高を回収する** ── 実測でチップが増えた。番号を外から
// 決めつけず、終端は種目ごとに持たせる。
type horseTable interface {
	Reset() error
	PlayerAction(action, amount, humanPlayMs int) error
	GetPhase() int
	GetGameEndFlag() bool
	GetCurrentTurn() int
	GetPot() int
	// GetPlayerCnt は卓に座っている人数を返す。復元した卓が席数と食い違って
	// いないかを確かめるために要る。
	GetPlayerCnt() int
}

// horseEndPhase は種目ごとの「ハンドが終わった」フェーズ番号。
type horseEndPhase struct {
	end   int
	rebuy int
}

// horseSeat は 1 席の通算成績。**チップは種目をまたいで持ち回る。**
type horseSeat struct {
	chips   int
	isHuman bool
	name    string
}

// Horse は H.O.R.S.E. の卓 (5 種目のオーケストレータ)。
//
// **ゲームの規則は 1 つも書いていない。** ホールデムもレーズも既存の実装が
// そのまま担当し、ここがやるのは 3 つだけ:
//
//  1. H-O-R-S-E の順に種目を回す
//  2. 種目をまたいでチップを持ち回す
//  3. 1 人を残して全員が飛んだら終える
//
// **チップの持ち回しが唯一の難所。** 各種目のプレイヤー型は別物なので、
// ハンドの開始時に「正本の残高」を配り、終了時に回収する。正本はここが持ち、
// 種目側の残高は 1 ハンドのあいだだけ有効な写しである。
type Horse struct {
	config HorseConfig
	seats  []*horseSeat

	phase HorsePhase
	// discipline はいまの種目。**H-O-R-S-E の順に進む。**
	discipline HorseDiscipline
	// handInDiscipline はいまの種目で何ハンド目か (1 始まり)。
	handInDiscipline int
	// handNumber は通算ハンド数。
	handNumber int

	// table はいまのハンドを進めている種目の実装。
	table horseTable
	// endPhase はいまの種目の終端フェーズ番号 (種目ごとに違う)。
	endPhase horseEndPhase
	// seatMap は種目側の席番号 → 正本の席番号。
	//
	// **飛んだ席は座らせないので、番号が詰まる。** 種目側の 0 番が正本の 0 番とは
	// 限らない。
	seatMap []int
	// harvest はハンド終了時に種目側の残高を正本へ戻す。
	harvest func()

	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewHorse は指定の設定で卓を構築する。席 0 が人間。
func NewHorse(config HorseConfig) *Horse {
	seats := make([]*horseSeat, 0, config.Seats)
	for i := range config.Seats {
		name := fmt.Sprintf("CPU%d", i)
		if i == 0 {
			name = "YOU"
		}
		seats = append(seats, &horseSeat{chips: config.InitialChips, isHuman: i == 0, name: name})
	}
	return &Horse{config: config, seats: seats, handInDiscipline: 1, handNumber: 1}
}

// NewDefaultHorse は既定の卓を構築する。
func NewDefaultHorse() *Horse { return NewHorse(DefaultHorseConfig()) }

// NewEightGame は Eight-Game Mix の卓を構築する。
//
// **オーケストレータは H.O.R.S.E. と同じもの。** 違うのは回す種目の並びだけで、
// チップの持ち回しも精算も 1 つの実装が担当する ── 8 種目ぶんの進行を別に
// 書くと、同じ規則を 2 か所で保つことになる。
func NewEightGame(config HorseConfig) *Horse {
	config.Variant = HorseVariantEightGame
	return NewHorse(config)
}

// NewDefaultEightGame は既定の Eight-Game Mix 卓を構築する。
func NewDefaultEightGame() *Horse { return NewHorse(DefaultEightGameConfig()) }

// GetVariant はこの卓が回すローテーションを返す。
func (g *Horse) GetVariant() HorseVariant { return g.config.Variant }

// GetRotation はこの卓が回す種目の並びを返す。
func (g *Horse) GetRotation() []HorseDiscipline { return HorseRotation(g.config.Variant) }

// Reset はゲームを初期化する。
func (g *Horse) Reset() {
	for i, s := range g.seats {
		s.chips = g.config.InitialChips
		s.isHuman = i == 0
	}
	g.phase = HorsePhaseHand
	g.discipline = g.rotationAt(0)
	g.handInDiscipline = 1
	g.handNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog("reset", "game reset")
	g.startHand()
}

// --- 種目の切り替え ---

// startHand はいまの種目で 1 ハンド始める。
func (g *Horse) startHand() {
	g.table, g.endPhase, g.harvest = g.buildTable()
	if g.table == nil {
		g.finish()
		return
	}
	if err := g.table.Reset(); err != nil {
		// **種目が配れないならその卓は畳む。** 席が足りない等で起きうる。
		g.finish()
		return
	}
	g.phase = HorsePhaseHand
	g.appendLog("hand", fmt.Sprintf("%s hand %d (%s)",
		HorseDisciplineLetter(g.discipline), g.handNumber, HorseDisciplineName(g.discipline)))
}

// buildTable はいまの種目の卓と、その残高を回収する関数を返す。
//
// **正本の残高を配ってから作る。** 種目側の残高は 1 ハンドのあいだだけ有効な
// 写しで、回収するまで正本は動かない。
func (g *Horse) buildTable() (horseTable, horseEndPhase, func()) {
	g.seatMap = g.aliveSeatIndexes()
	if len(g.seatMap) < HorseMinSeats {
		return nil, horseEndPhase{}, nil
	}
	n := len(g.seatMap)
	switch g.discipline {
	case HorseHoldem:
		players := NewPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewHoldem(NewTrumpCards(0), players, horseTableConfig(DefaultHoldemConfig(), n))
		return t, horseEndPhase{end: HoldemPhaseEnd, rebuy: HoldemPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorseOmahaHiLo:
		players := NewOmahaPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewOmahaHiLo(NewTrumpCards(0), players, horseTableConfig(DefaultOmahaConfig(), n))
		return t, horseEndPhase{end: HoldemPhaseEnd, rebuy: HoldemPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorseRazz:
		players := NewSevenCardStudPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewRazz(NewTrumpCards(0), players, horseStudTableConfig(DefaultRazzConfig(), n))
		return t, horseEndPhase{end: SevenCardStudPhaseEnd, rebuy: SevenCardStudPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorseStud:
		players := NewSevenCardStudPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewSevenCardStud(NewTrumpCards(0), players, horseStudTableConfig(DefaultSevenCardStudConfig(), n))
		return t, horseEndPhase{end: SevenCardStudPhaseEnd, rebuy: SevenCardStudPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorseStudHiLo:
		players := NewSevenCardStudPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewSevenCardStudHiLo(NewTrumpCards(0), players, horseStudTableConfig(DefaultSevenCardStudConfig(), n))
		return t, horseEndPhase{end: SevenCardStudPhaseEnd, rebuy: SevenCardStudPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorseNLHoldem:
		players := NewPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewHoldem(NewTrumpCards(0), players, horseLimitConfig(horseTableConfig(DefaultHoldemConfig(), n), BettingLimitNoLimit))
		return t, horseEndPhase{end: HoldemPhaseEnd, rebuy: HoldemPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorsePLOmaha:
		players := NewOmahaPlayersForTable(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewOmaha(NewTrumpCards(0), players, horseLimitConfig(horseTableConfig(DefaultOmahaConfig(), n), BettingLimitPotLimit))
		return t, horseEndPhase{end: HoldemPhaseEnd, rebuy: HoldemPhaseRebuy}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayer(i).GetChips() }) }
	case HorseTripleDraw:
		players := newHorseDeuceToSevenPlayers(n)
		g.dealChipsTo(func(i, chips int) { players[i].SetChips(chips) })
		t := NewDeuceToSeven(NewTrumpCards(0), players, horseDrawTableConfig(DefaultDeuceToSevenConfig(), n))
		return t, horseEndPhase{end: DeuceToSevenPhaseEnd, rebuy: DeuceToSevenPhaseEnd}, func() { g.collectChipsFrom(func(i int) int { return t.GetPlayers()[i].GetChips() }) }
	default:
		return nil, horseEndPhase{}, nil
	}
}

// horseLimitConfig はベッティングリミットだけを差し替える。
//
// **NLH と PLO は「同じ種目の別リミット」ではない。** Eight-Game Mix は
// リミットホールデムとノーリミットホールデムを**別の種目として**回すので、
// 卓を作るときにリミットを指定しないと、8 種目のうち 2 つが先に回した種目と
// 同じものになる。
func horseLimitConfig(cfg HoldemConfig, limit BettingLimitType) HoldemConfig {
	cfg.BettingLimit = limit
	return cfg
}

// horseDrawTableConfig は 2-7 Triple Draw の設定を席数に合わせる。
//
// **CPU の人数で卓の大きさが決まる。** ドロー系の設定は席数ではなく
// 「CPU 何人と打つか」を持っているので、正本の席数から 1 (人間) を引いて渡す。
func horseDrawTableConfig(cfg DeuceToSevenConfig, seats int) DeuceToSevenConfig {
	cfg.CpuCount = seats - 1
	return cfg
}

// newHorseDeuceToSevenPlayers は席数ぶんの 2-7 Triple Draw プレイヤーを作る。
//
// **席 0 が人間なのは他の種目と同じ。** CPU のスタイルは席順に配る ──
// 全員同じスタイルにすると、卓の 3 人が同じ手を打つ。
func newHorseDeuceToSevenPlayers(seats int) []*DeuceToSevenPlayer {
	styles := []DeuceToSevenPlayStyle{
		DeuceToSevenStyleConservative, DeuceToSevenStyleAggressive, DeuceToSevenStyleBluffer,
	}
	players := make([]*DeuceToSevenPlayer, 0, seats)
	players = append(players, NewDeuceToSevenPlayer(true, DeuceToSevenStyleBalanced))
	for i := 1; i < seats; i++ {
		players = append(players, NewDeuceToSevenPlayer(false, styles[(i-1)%len(styles)]))
	}
	return players
}

// horseTableConfig は種目の設定を「いま座らせる人数」に合わせる。
//
// **既定の設定は 4 人卓を指している。** 飛んだ席を抜いて 3 人で座らせても、
// 設定の TableSize が 4 のままだと**種目側は 4 人ぶんの席を作り直し**、こちらが
// 残高を配ったプレイヤーとは別の 4 人が打つ ── 回収は元の 3 人を読むので、
// 人間の残高がまったく動かず、総量が ±数十ずれる (実測: 卓 4 人 / 正本 3 席)。
func horseTableConfig(cfg HoldemConfig, seats int) HoldemConfig {
	cfg.TableSize = seats
	// **リバイとアドオンは種目に任せない。** 飛んだ席を座らせないのはこちらの
	// 役目で、種目側が補充するとチップが湧く。
	cfg.RebuyEnabled = false
	cfg.AddonEnabled = false
	return cfg
}

// horseStudTableConfig はスタッド系の設定を同じ理由で揃える。
func horseStudTableConfig(cfg SevenCardStudConfig, seats int) SevenCardStudConfig {
	cfg.TableSize = seats
	cfg.RebuyEnabled = false
	cfg.AddonEnabled = false
	return cfg
}

// aliveSeatIndexes はチップの残っている席の番号を並べて返す。
//
// **飛んだ席を種目に座らせない。** 各エンジンの `Reset` は残高 0 の席を
// `InitChips` まで**黙って積み直す** (単体のゲームなら卓を続けるための正しい
// 挙動)。ミックスゲームでそれをやると**チップが湧く** ── 実測で総量が
// 3000 → 3114 に増えた。座らせなければ積み直されない。
//
// **ただし人数を減らした卓は作れない。** Holdem 系の卓は 4 / 6 / 9 人のいずれかで、
// 3 人を渡すと黙って 4 人に落とされる。だから席が 1 つでも飛んだ時点で
// `startHand` は卓を作らず、そのマッチを終える ── 「誰かが飛んだら終わり」は
// ミックスゲームの区切りとしても自然で、席を欠いたまま続けるより誤魔化しが無い。
func (g *Horse) aliveSeatIndexes() []int {
	out := make([]int, 0, len(g.seats))
	for i, s := range g.seats {
		if s.chips > 0 {
			out = append(out, i)
		}
	}
	if len(out) != len(g.seats) {
		// 欠けた卓は作らない (呼び出し側が finish する)。
		return nil
	}
	return out
}

// dealChipsTo は正本の残高を種目側へ配る。
func (g *Horse) dealChipsTo(set func(i, chips int)) {
	for tableIdx, seatIdx := range g.seatMap {
		set(tableIdx, g.seats[seatIdx].chips)
	}
}

// collectChipsFrom は種目側の残高を正本へ戻す。
//
// **読むのは卓のアクセサであってスライスではない。** 復元では
// `json.Unmarshal` が卓の中のプレイヤーを丸ごと差し替えるので、作ったときの
// スライスを閉じ込めておくと**差し替え前の別人を読む** ── ハンドを打っても
// 正本の残高が 1 円も動かなかった (実測)。卓のポインタは同じままなので、
// `GetPlayer(i)` を通せば必ず現物を読む。
func (g *Horse) collectChipsFrom(get func(i int) int) {
	for tableIdx, seatIdx := range g.seatMap {
		g.seats[seatIdx].chips = get(tableIdx)
	}
}

// NextHand はハンドを閉じて次へ進む。
//
// **種目が変わるのはここだけ。** ハンドの途中で切り替わることは無い。
func (g *Horse) NextHand() error {
	if g.gameEndFlag {
		return NewDomainErrorCode(errHorseFinished, "horse.errFinished", nil)
	}
	if g.phase != HorsePhaseHandEnd {
		return NewDomainErrorCode(errHorseWrongPhase, "horse.errWrongPhase", nil)
	}
	if g.aliveSeats() < HorseMinSeats {
		g.finish()
		return nil
	}
	g.handNumber++
	if g.handInDiscipline >= g.config.HandsPerDiscipline {
		g.handInDiscipline = 1
		g.discipline = g.nextDiscipline()
	} else {
		g.handInDiscipline++
	}
	g.startHand()
	return nil
}

// rotationAt はこの卓のローテーションの i 番目の種目を返す。
func (g *Horse) rotationAt(i int) HorseDiscipline {
	rot := horseRotations[g.rotationVariant()]
	return rot[((i%len(rot))+len(rot))%len(rot)]
}

// rotationVariant は範囲内に丸めたバリアントを返す。
//
// **範囲外を H.O.R.S.E. として扱う。** 設定は Validate を通っているが、
// ゼロ値の `Horse{}` を直接組み立てる経路 (復元の途中など) でも
// ローテーションの参照が落ちないようにする。
func (g *Horse) rotationVariant() HorseVariant {
	if g.config.Variant < 0 || int(g.config.Variant) >= len(horseRotations) {
		return HorseVariantHorse
	}
	return g.config.Variant
}

// nextDiscipline はローテーション上の次の種目を返す。
//
// **番号の +1 ではない。** 種目の値は 8 つあり、H.O.R.S.E. が回すのはその
// 先頭 5 つだけ ── `(d+1) % 種目数` で進めると、バリアントごとの並びと
// 食い違う。
func (g *Horse) nextDiscipline() HorseDiscipline {
	idx := HorseRotationIndex(g.rotationVariant(), g.discipline)
	if idx < 0 {
		return g.rotationAt(0)
	}
	return g.rotationAt(idx + 1)
}

// --- 進行 ---

// PlayerAction は人間の手をいまの種目へ渡す。
//
// **規則の判定はここでは一切しない。** レイズ上限もベット額の刻みも種目ごとに
// 違うので、そのまま渡して種目に決めさせる ── ここで真似ると 5 種目ぶんの規則を
// 二重に持つことになる。
func (g *Horse) PlayerAction(action, amount, humanPlayMs int) error {
	if g.gameEndFlag {
		return NewDomainErrorCode(errHorseFinished, "horse.errFinished", nil)
	}
	if g.phase != HorsePhaseHand {
		return NewDomainErrorCode(errHorseWrongPhase, "horse.errWrongPhase", nil)
	}
	if g.table == nil {
		return NewDomainErrorCode(errHorseNoTable, "horse.errNoTable", nil)
	}
	if err := g.table.PlayerAction(action, amount, humanPlayMs); err != nil {
		return err
	}
	g.settleIfHandOver()
	return nil
}

// horseMuckTable は「負けた手を伏せるか公開するか」を訊いてくる種目。
//
// Holdem / Omaha / SevenCardStud はショーダウンで人間が勝てなかったとき、
// **その決定を待って止まる**。2-7 Triple Draw にはこの分岐が無いので、
// インタフェースは必須にせず、実装している卓にだけ訊く。
type horseMuckTable interface {
	IsMuckAvailable() bool
	ShowHand() error
}

// resolveShowdownDecision はショーダウンの待ちを解いて手を公開する。
//
// **これが無いとマッチが凍る。** 人間がコールして負けた手は
// `resolveShowdown` が END へ進めずショーダウンに留まり、マック待ちになる ──
// ところがオーケストレータにはその入力が無いので、**打てる手が 1 つも無い**
// 盤面のまま次のハンドへも進めない (実測: 単独の H.O.R.S.E. でも再現する)。
//
// ミックスゲームでは伏せる意味が無いので公開して閉じる。卓の要約しか出さない
// 画面では伏せ札と公開札の区別がそもそも現れず、CPU も履歴を読まない。
func (g *Horse) resolveShowdownDecision() {
	t, ok := g.table.(horseMuckTable)
	if !ok || !t.IsMuckAvailable() {
		return
	}
	if err := t.ShowHand(); err != nil {
		return
	}
	g.appendLog("show", fmt.Sprintf("hand %d shown down", g.handNumber))
}

// settleIfHandOver は種目のハンドが終わっていれば残高を回収する。
//
// **回収はここ 1 か所。** 種目側の残高を正本に戻す経路を増やすと、二重に
// 回収して増える経路ができる。
func (g *Horse) settleIfHandOver() {
	if g.table == nil {
		return
	}
	g.resolveShowdownDecision()
	if !g.tableHandIsOver() {
		return
	}
	if g.harvest != nil {
		g.harvest()
	}
	g.phase = HorsePhaseHandEnd
	g.appendLog("handEnd", fmt.Sprintf("hand %d settled", g.handNumber))
	if g.aliveSeats() < HorseMinSeats {
		g.finish()
	}
}

// tableHandIsOver は種目がもう手を受け付けないかを返す。
//
// **番号は種目ごとに違う** ので `g.endPhase` を見る。ここを `HoldemPhaseEnd`
// 固定で書いて 2 つ踏んだ:
//
//   - Stud のリバイ待ちを終端と認めず、ハンドが閉じないまま次の手が
//     「Game already ended」で拒まれ続けた (盤面は動かずエラーだけ返る)
//   - 逆に Holdem の番号で Stud を読むと、**ショーダウン中を終端と誤読して
//     精算前の残高を回収し、チップが増えた**
//
// 終了フラグも併せて見る ── 番号を 1 つでも取りこぼすと固まる側に倒れる。
func (g *Horse) tableHandIsOver() bool {
	switch g.table.GetPhase() {
	case g.endPhase.end, g.endPhase.rebuy:
		return true
	}
	return g.table.GetGameEndFlag()
}

// aliveSeats はまだチップが残っている席の数を返す。
func (g *Horse) aliveSeats() int {
	n := 0
	for _, s := range g.seats {
		if s.chips > 0 {
			n++
		}
	}
	return n
}

// finish はゲームを終える。
func (g *Horse) finish() {
	g.gameEndFlag = true
	g.phase = HorsePhaseGameEnd
	g.appendLog("gameEnd", fmt.Sprintf("winner seat %d", g.WinnerSeat()))
}

// WinnerSeat はチップがいちばん多い席を返す。同点なら若い席。
func (g *Horse) WinnerSeat() int {
	best, bestChips := 0, -1
	for i, s := range g.seats {
		if s.chips > bestChips {
			best, bestChips = i, s.chips
		}
	}
	return best
}

// --- 参照 ---

// GetConfig はゲーム設定を返す。
func (g *Horse) GetConfig() HorseConfig { return g.config }

// SetConfig はゲーム設定を設定する。
func (g *Horse) SetConfig(c HorseConfig) { g.config = c }

// GetPhase は現在のフェーズを返す。
func (g *Horse) GetPhase() HorsePhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Horse) GetGameEndFlag() bool { return g.gameEndFlag }

// GetDiscipline はいまの種目を返す。
func (g *Horse) GetDiscipline() HorseDiscipline { return g.discipline }

// GetDisciplineLetter はいまの種目の頭文字を返す。
func (g *Horse) GetDisciplineLetter() string { return HorseDisciplineLetter(g.discipline) }

// GetHandInDiscipline はいまの種目で何ハンド目かを返す。
func (g *Horse) GetHandInDiscipline() int { return g.handInDiscipline }

// GetHandNumber は通算ハンド数を返す。
func (g *Horse) GetHandNumber() int { return g.handNumber }

// GetSeatChips は席のチップ数を返す。
func (g *Horse) GetSeatChips(i int) int {
	if i < 0 || i >= len(g.seats) {
		return 0
	}
	return g.seats[i].chips
}

// SetSeatChips は席のチップ数を設定する。
func (g *Horse) SetSeatChips(i, chips int) {
	if i < 0 || i >= len(g.seats) {
		return
	}
	g.seats[i].chips = chips
}

// GetSeatName は席の表示名を返す。
func (g *Horse) GetSeatName(i int) string {
	if i < 0 || i >= len(g.seats) {
		return "?"
	}
	return g.seats[i].name
}

// GetSeatIsHuman は人間の席かを返す。
func (g *Horse) GetSeatIsHuman(i int) bool {
	if i < 0 || i >= len(g.seats) {
		return false
	}
	return g.seats[i].isHuman
}

// GetSeatCount は席数を返す。
func (g *Horse) GetSeatCount() int { return len(g.seats) }

// GetHumanSeat は人間の席を返す。
func (g *Horse) GetHumanSeat() int {
	for i, s := range g.seats {
		if s.isHuman {
			return i
		}
	}
	return 0
}

// GetCurrentTurn はいまの手番を**正本の席番号**で返す。ハンドが無ければ -1。
//
// **種目側の番号をそのまま出さない。** 飛んだ席は座らせないので番号が詰まって
// いて、種目の 0 番が正本の 0 番とは限らない。
func (g *Horse) GetCurrentTurn() int {
	if g.table == nil {
		return -1
	}
	return g.toCanonicalSeat(g.table.GetCurrentTurn())
}

// toCanonicalSeat は種目側の席番号を正本の席番号に直す。
func (g *Horse) toCanonicalSeat(tableIdx int) int {
	if tableIdx < 0 || tableIdx >= len(g.seatMap) {
		return -1
	}
	return g.seatMap[tableIdx]
}

// IsHumanTurn は人間の操作待ちかを返す。
func (g *Horse) IsHumanTurn() bool {
	return g.phase == HorsePhaseHand && g.table != nil && g.GetCurrentTurn() == g.GetHumanSeat()
}

// GetPot はいまの種目のポットを返す。
func (g *Horse) GetPot() int {
	if g.table == nil {
		return 0
	}
	return g.table.GetPot()
}

// GetTablePhase はいまの種目の内部フェーズを返す (Holdem のフェーズ定数)。
func (g *Horse) GetTablePhase() int {
	if g.table == nil {
		return HoldemPhaseInit
	}
	return g.table.GetPhase()
}

// GetActionLog は棋譜を返す。
func (g *Horse) GetActionLog() []*ActionLogEntry { return g.actionLog }

// appendLog は棋譜に 1 行足す。
func (g *Horse) appendLog(actionType, detail string) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  -1,
		ActionType: actionType,
		Detail:     detail,
	})
	if len(g.actionLog) > horseMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-horseMaxSliceLen:]
	}
}

// horseTableIdx は正本の席番号から卓の席番号を引く。座っていなければ -1。
func (g *Horse) horseTableIdx(seat int) int {
	for ti, si := range g.seatMap {
		if si == seat {
			return ti
		}
	}
	return -1
}

// playerCardsOf は卓の席 ti について (手札すべて, 表向きだけ) を返す。
//
// **スタッド系は「表向き」が実在する。** ホールデム系では自分の 2 枚 (オマハは 4 枚)
// はすべて伏せ札なので、表向きは空 ── 共有札は GetCommunityCards が返す。
func (g *Horse) playerCardsOf(ti int) (all, up []*Card) {
	switch t := g.table.(type) {
	case *Holdem:
		p := t.GetPlayer(ti)
		if p == nil {
			return nil, nil
		}
		for i := range p.GetCardsSize() {
			all = append(all, p.GetCard(i))
		}
		return all, nil
	case *Omaha:
		p := t.GetPlayer(ti)
		if p == nil {
			return nil, nil
		}
		for i := range p.GetCardsSize() {
			all = append(all, p.GetCard(i))
		}
		return all, nil
	case *SevenCardStud:
		p := t.GetPlayer(ti)
		if p == nil {
			return nil, nil
		}
		up = append(up, p.GetDoorCards()...)
		all = append(all, p.GetHoleCards()...)
		all = append(all, up...)
		return all, up
	case *DeuceToSeven:
		// **ドロー系は 1 枚も表を向かない。** 引いた枚数だけが公開情報で、
		// 札そのものはショーダウンまで誰にも見えない。
		p := horseDrawPlayer(t, ti)
		if p == nil {
			return nil, nil
		}
		for i := range p.GetCardsSize() {
			all = append(all, p.GetCard(i))
		}
		return all, nil
	default:
		return nil, nil
	}
}

// horseDrawPlayer は 2-7 Triple Draw の卓から席 ti のプレイヤーを返す。
//
// **スライスは毎回卓から引き直す。** 復元でプレイヤー列は丸ごと差し替わるので、
// 掴んでおくと差し替え前の別人を読む。
func horseDrawPlayer(t *DeuceToSeven, ti int) *DeuceToSevenPlayer {
	players := t.GetPlayers()
	if ti < 0 || ti >= len(players) {
		return nil
	}
	return players[ti]
}

// GetSeatCards は指定席の「その席から見えている札」を返す。
//
// **CPU の伏せ札は返さない。** 人間の席なら全部、それ以外は表向きだけ ──
// 出力に混ぜると画面から相手の手が読めてしまう。
func (g *Horse) GetSeatCards(seat int) []*Card {
	if g.table == nil || seat < 0 || seat >= len(g.seats) {
		return nil
	}
	ti := g.horseTableIdx(seat)
	if ti < 0 {
		return nil
	}
	all, up := g.playerCardsOf(ti)
	if g.seats[seat].isHuman {
		return all
	}
	return up
}

// GetCommunityCards はいまの種目の共有札を返す。スタッド系には無いので空。
func (g *Horse) GetCommunityCards() []*Card {
	switch t := g.table.(type) {
	case *Holdem:
		return t.GetCommunityCards()
	case *Omaha:
		return t.GetCommunityCards()
	default:
		return nil
	}
}

// GetToCall は人間の席がコールに要する額を返す。0 ならチェックできる。
//
// **「賭けられているか」は種目に訊く。** ここで持つと、種目側がベットを受け
// 付けた瞬間に食い違い、画面がチェックできない場面でチェックを出す。
func (g *Horse) GetToCall() int {
	ti := g.horseTableIdx(g.GetHumanSeat())
	if g.table == nil || ti < 0 {
		return 0
	}
	var lastBet, mine int
	switch t := g.table.(type) {
	case *Holdem:
		p := t.GetPlayer(ti)
		if p == nil {
			return 0
		}
		lastBet, mine = t.GetLastBet(), p.GetCurrentBet()
	case *Omaha:
		p := t.GetPlayer(ti)
		if p == nil {
			return 0
		}
		lastBet, mine = t.GetLastBet(), p.GetCurrentBet()
	case *SevenCardStud:
		p := t.GetPlayer(ti)
		if p == nil {
			return 0
		}
		lastBet, mine = t.GetLastBet(), p.GetCurrentBet()
	case *DeuceToSeven:
		p := horseDrawPlayer(t, ti)
		if p == nil {
			return 0
		}
		lastBet, mine = t.GetLastBet(), p.GetCurrentBet()
	default:
		return 0
	}
	if lastBet <= mine {
		return 0
	}
	return lastBet - mine
}

// GetMinRaise はいまの種目が受け付ける最小のレイズ幅を返す。
func (g *Horse) GetMinRaise() int {
	switch t := g.table.(type) {
	case *Holdem:
		return t.GetMinRaise()
	case *Omaha:
		return t.GetMinRaise()
	case *SevenCardStud:
		return t.GetMinRaise()
	case *DeuceToSeven:
		return t.GetMinRaise()
	default:
		return 0
	}
}

// GetSeatLiveChips は「いま画面に出すべき残高」を返す。
//
// **正本はハンドが終わるまで動かない。** 残高は開始時に卓へ配り、終了時に
// 回収する写しなので、打っている最中に正本を出すと**自分がいくら賭けたのかが
// 画面に出ない** ── ポットだけが増えて手持ちが減らないように見える。
// 卓があるあいだは卓の残高を、無ければ正本を返す。
func (g *Horse) GetSeatLiveChips(seat int) int {
	if seat < 0 || seat >= len(g.seats) {
		return 0
	}
	ti := g.horseTableIdx(seat)
	if g.table == nil || ti < 0 {
		return g.seats[seat].chips
	}
	switch t := g.table.(type) {
	case *Holdem:
		if p := t.GetPlayer(ti); p != nil {
			return p.GetChips()
		}
	case *Omaha:
		if p := t.GetPlayer(ti); p != nil {
			return p.GetChips()
		}
	case *SevenCardStud:
		if p := t.GetPlayer(ti); p != nil {
			return p.GetChips()
		}
	case *DeuceToSeven:
		if p := horseDrawPlayer(t, ti); p != nil {
			return p.GetChips()
		}
	}
	return g.seats[seat].chips
}

// --- ドロー (2-7 Triple Draw のときだけ動く) ---

// IsDrawPhase はいまの種目が引き直しを待っているかを返す。
//
// **賭ける場面と引く場面は別の入力。** これを見ずにベットの面だけ出すと、
// ドローの番で押せるボタンが 1 つも無くなり、Eight-Game Mix は 6 種目目で
// 止まる。
func (g *Horse) IsDrawPhase() bool {
	t, ok := g.table.(*DeuceToSeven)
	return ok && g.phase == HorsePhaseHand && t.GetPhase() == DeuceToSevenPhaseDraw
}

// GetDrawIndex は何回目の引き直しかを返す (1..3)。ドロー中でなければ 0。
func (g *Horse) GetDrawIndex() int {
	t, ok := g.table.(*DeuceToSeven)
	if !ok || !g.IsDrawPhase() {
		return 0
	}
	return t.GetDrawIndex()
}

// PlayerExchange は人間の引き直しをいまの種目へ渡す。
//
// **枚数も規則も種目に決めさせる。** 何枚まで引けるかはドロー系の規則で、
// ここで真似ると同じ規則を 2 か所で持つことになる。空のスライスは
// スタンドパット (引かない)。
func (g *Horse) PlayerExchange(indices []int) error {
	if g.gameEndFlag {
		return NewDomainErrorCode(errHorseFinished, "horse.errFinished", nil)
	}
	if g.phase != HorsePhaseHand {
		return NewDomainErrorCode(errHorseWrongPhase, "horse.errWrongPhase", nil)
	}
	t, ok := g.table.(*DeuceToSeven)
	if !ok {
		return NewDomainErrorCode(errHorseNoDraw, "horse.errNoDraw", nil)
	}
	if err := t.PlayerExchange(indices); err != nil {
		return err
	}
	g.settleIfHandOver()
	return nil
}
