package domain

import (
	"encoding/json"
	"fmt"
)

// DurakPhase ゲームフェーズ
type DurakPhase int

// Durakのフェーズ定数
const (
	// DurakPhaseAttack アタックフェーズ (攻撃者がカードを出す)
	DurakPhaseAttack DurakPhase = 0
	// DurakPhaseDefend ディフェンスフェーズ (防御者がカードで防ぐ)
	DurakPhaseDefend DurakPhase = 1
	// DurakPhaseBoutEnd バウト終了フェーズ (補充・次のバウトへ)
	DurakPhaseBoutEnd DurakPhase = 2
	// DurakPhaseGameEnd ゲーム終了フェーズ
	DurakPhaseGameEnd DurakPhase = 3
)

// DurakSortMode 手札ソートモード
type DurakSortMode int

// DurakSortMode定数
const (
	// DurakSortBySuit スート順 (デフォルト: 非切り札→切り札、各スート内で値順)
	DurakSortBySuit DurakSortMode = 0
	// DurakSortByValue 値順
	DurakSortByValue DurakSortMode = 1
)

// DurakTablePair テーブル上の攻撃・防御カードペア
type DurakTablePair struct {
	Attack  *Card `json:"a"`           // 攻撃カード
	Defense *Card `json:"d,omitempty"` // 防御カード (nil = 未防御)
}

// DurakActionType 行動種別定数
const (
	DurakActionAttack = 0 // 攻撃
	DurakActionDefend = 1 // 防御
	DurakActionPass   = 2 // パス
	DurakActionTake   = 3 // 引き取り
)

// DurakMaxBouts は 1 局のバウト数の上限。
//
// **山札が尽きた後、カードが場から消えるのは防御成功のときだけ**で、防御失敗なら
// 防御者が全部引き取るので手札の総数は変わらない。CPU の手選択に乱数は無く、同じ
// 局面からは必ず同じ手が選ばれるため、「山札切れ + 防御が成立しない配置」の循環に
// 一度入ると永久に出られない。20 万局に 14 局(0.007%)で実際に起きる。
//
// 通常局は 44 バウト前後、実測の最大が 78 なので、200 は健全な局に触れない。
const DurakMaxBouts = 200

// DurakCpuAction CPUまたは人間の1ターン分の行動記録
type DurakCpuAction struct {
	PlayerIdx  int   // 行動したプレイヤーインデックス
	ActionType int   // DurakActionAttack/Defend/Pass/Take
	CardIdx    int   // 使用した手札インデックス (-1 = パス/テイク)
	AttackIdx  int   // 防御時: 対象の攻撃カードインデックス (-1 = 該当なし)
	Card       *Card // 出したカード (nil = パス/テイク)
}

// durakRoundState ラウンドごとにリセットされる状態
type durakRoundState struct {
	attackerIdx int               // 現在の攻撃者インデックス
	defenderIdx int               // 現在の防御者インデックス
	currentTurn int               // 現在の手番プレイヤーインデックス
	tablePairs  []*DurakTablePair // テーブル上のカードペア
	phase       DurakPhase        // 現在のフェーズ
	gameEndFlag bool              // ゲーム終了フラグ
	loserIdx    int               // 敗者インデックス (-1 = 未確定)
	cpuActions  []*DurakCpuAction // CPU行動記録
	humanAction *DurakCpuAction   // 人間の最後の行動
	boutNumber  int               // バウト番号
	actionLogBase
}

// Durak ドゥラークゲームクラス
type Durak struct {
	trumpCards  *TrumpCards
	players     []*DurakPlayer
	config      DurakConfig
	sortMode    DurakSortMode
	trumpSuit   int     // 切り札スート
	trumpCard   *Card   // 切り札表示カード (山札の底)
	stock       []*Card // 山札
	discardPile []*Card // 捨て札
	round       durakRoundState
}

// NewDurak コンストラクタ
func NewDurak(trumpCards *TrumpCards, players []*DurakPlayer) *Durak {
	return &Durak{
		trumpCards: trumpCards,
		players:    players,
		config:     DefaultDurakConfig(),
		round: durakRoundState{
			loserIdx: -1,
		},
	}
}

// NewDefaultDurak returns Durak with the standard 4-player setup (1 human, 3 CPU)
// using the short deck. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultDurak() *Durak {
	players := []*DurakPlayer{
		NewDurakPlayer(true),
		NewDurakPlayer(false),
		NewDurakPlayer(false),
		NewDurakPlayer(false),
	}
	return NewDurak(NewTrumpCardsShortDeck(), players)
}

// ---- 公開メソッド: ゲーム操作 ----

// Reset ゲーム初期化
func (d *Durak) Reset() {
	d.round = durakRoundState{
		loserIdx: -1,
	}
	d.stock = nil
	d.discardPile = nil

	// 全プレイヤーのカードリセット
	resetPlayers(d.players, func(_ *DurakPlayer) {})

	// デッキをシャッフル
	d.trumpCards.Shuffle()

	// 各プレイヤーに6枚ずつ配る
	for range DurakHandSize {
		for _, p := range d.players {
			card := d.trumpCards.DrawCard()
			if card != nil {
				p.AddCard(card)
			}
		}
	}

	// 残りのカードを山札へ
	d.stock = make([]*Card, 0)
	for {
		card := d.trumpCards.DrawCard()
		if card == nil {
			break
		}
		d.stock = append(d.stock, card)
	}

	// 山札の底のカードで切り札スートを決定
	if len(d.stock) > 0 {
		d.trumpCard = d.stock[len(d.stock)-1]
		d.trumpSuit = d.trumpCard.GetDesign()
	} else {
		// カードが足りない場合 (2人プレイで24枚配って残り12枚)
		// 最後に配ったカードで代用
		d.trumpSuit = CardDesignSpade
		d.trumpCard = nil
	}

	// 手札ソート
	d.sortAllHands()

	// 最初の攻撃者を決定 (最小の切り札を持つプレイヤー)
	d.round.attackerIdx = d.findFirstAttacker()
	d.round.defenderIdx = d.nextActivePlayer(d.round.attackerIdx)
	d.round.currentTurn = d.round.attackerIdx
	d.round.phase = DurakPhaseAttack
}

// PlayerAttack 人間プレイヤーが攻撃カードを出す
func (d *Durak) PlayerAttack(cardIdx int) error {
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if d.round.phase != DurakPhaseAttack {
		return ErrWrongPhase
	}
	if d.round.currentTurn != d.round.attackerIdx {
		return NewDomainError(ErrWrongPhase, "not attacker's turn")
	}

	player := d.players[d.round.attackerIdx]
	if cardIdx < 0 || cardIdx >= player.GetCardsSize() {
		return ErrInvalidCard
	}

	card := player.GetCard(cardIdx)
	if !d.canAttackWith(card) {
		return NewDomainError(ErrInvalidPlay, "card rank not on table")
	}

	// 防御者の手札枚数チェック (未防御カード数 < 防御者の手札)
	undefended := d.countUndefended()
	defender := d.players[d.round.defenderIdx]
	if undefended >= defender.GetCardsSize() {
		return NewDomainError(ErrInvalidPlay, "defender has no cards left to defend")
	}

	played := player.RemoveCard(cardIdx)
	d.round.tablePairs = append(d.round.tablePairs, &DurakTablePair{Attack: played})

	d.round.humanAction = &DurakCpuAction{
		PlayerIdx:  d.round.attackerIdx,
		ActionType: DurakActionAttack,
		CardIdx:    cardIdx,
		AttackIdx:  -1,
		Card:       played,
	}
	d.round.cpuActions = nil

	d.appendLog(d.round.attackerIdx, "attack", fmt.Sprintf("attacks with %s", cardStr(played)), []*Card{played})

	// 攻撃者の手札がなくなった場合のチェック
	if player.GetCardsSize() == 0 && len(d.stock) == 0 {
		player.SetIsFinished(true)
	}

	// 防御フェーズへ
	d.round.phase = DurakPhaseDefend
	d.round.currentTurn = d.round.defenderIdx
	return nil
}

// PlayerDefend 人間プレイヤーが防御カードを出す
func (d *Durak) PlayerDefend(attackIdx, handIdx int) error {
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if d.round.phase != DurakPhaseDefend {
		return ErrWrongPhase
	}

	defender := d.players[d.round.defenderIdx]
	if attackIdx < 0 || attackIdx >= len(d.round.tablePairs) {
		return ErrInvalidCard
	}
	if d.round.tablePairs[attackIdx].Defense != nil {
		return NewDomainError(ErrInvalidPlay, "already defended")
	}
	if handIdx < 0 || handIdx >= defender.GetCardsSize() {
		return ErrInvalidCard
	}

	defCard := defender.GetCard(handIdx)
	atkCard := d.round.tablePairs[attackIdx].Attack
	if !d.canBeat(atkCard, defCard) {
		return NewDomainError(ErrInvalidPlay, "card cannot beat attack card")
	}

	played := defender.RemoveCard(handIdx)
	d.round.tablePairs[attackIdx].Defense = played

	d.round.humanAction = &DurakCpuAction{
		PlayerIdx:  d.round.defenderIdx,
		ActionType: DurakActionDefend,
		CardIdx:    handIdx,
		AttackIdx:  attackIdx,
		Card:       played,
	}
	d.round.cpuActions = nil

	d.appendLog(d.round.defenderIdx, "defend", fmt.Sprintf("defends %s with %s", cardStr(atkCard), cardStr(played)), []*Card{played})

	// 全て防御完了した場合
	if d.countUndefended() == 0 {
		// 攻撃者がまだ出せるなら攻撃フェーズへ戻る
		if d.canContinueAttack() {
			d.round.phase = DurakPhaseAttack
			d.round.currentTurn = d.round.attackerIdx
		} else {
			d.endBout(true)
		}
	}
	return nil
}

// PlayerPass 人間プレイヤーがパス (攻撃者が追加攻撃を止める)
func (d *Durak) PlayerPass() error {
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if d.round.phase != DurakPhaseAttack {
		return ErrWrongPhase
	}
	// 初回攻撃ではパスできない
	if len(d.round.tablePairs) == 0 {
		return ErrCannotPass
	}

	d.round.humanAction = &DurakCpuAction{
		PlayerIdx:  d.round.attackerIdx,
		ActionType: DurakActionPass,
		CardIdx:    -1,
		AttackIdx:  -1,
	}
	d.round.cpuActions = nil

	d.appendLog(d.round.attackerIdx, "pass", "stops attacking", nil)

	// 未防御カードがあるかチェック
	if d.countUndefended() > 0 {
		// 防御者にまだ防御の機会を与える
		d.round.phase = DurakPhaseDefend
		d.round.currentTurn = d.round.defenderIdx
	} else {
		// 全て防御済み → バウト成功
		d.endBout(true)
	}
	return nil
}

// PlayerTakeCards 人間プレイヤーがカードを引き取る (防御放棄)
func (d *Durak) PlayerTakeCards() error {
	if d.round.gameEndFlag {
		return ErrGameEnded
	}
	if !d.players[d.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if d.round.phase != DurakPhaseDefend {
		return ErrWrongPhase
	}

	d.round.humanAction = &DurakCpuAction{
		PlayerIdx:  d.round.defenderIdx,
		ActionType: DurakActionTake,
		CardIdx:    -1,
		AttackIdx:  -1,
	}
	d.round.cpuActions = nil

	d.appendLog(d.round.defenderIdx, "take", "picks up all table cards", nil)

	d.endBout(false)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行する
func (d *Durak) CpuPlay() {
	if d.round.gameEndFlag || d.players[d.round.currentTurn].GetIsHuman() {
		return
	}

	switch d.round.phase {
	case DurakPhaseAttack:
		d.cpuAttack()
	case DurakPhaseDefend:
		d.cpuDefend()
	}
}

// SortHumanHand 人間の手札をソートする
func (d *Durak) SortHumanHand(mode DurakSortMode) error {
	d.sortMode = mode
	for _, p := range d.players {
		if p.GetIsHuman() {
			switch mode {
			case DurakSortBySuit:
				p.SortCards(d.trumpSuit)
			case DurakSortByValue:
				p.SortCardsByValue(d.trumpSuit)
			}
			break
		}
	}
	return nil
}

// SetConfig ゲーム設定をセットする
func (d *Durak) SetConfig(config DurakConfig) {
	d.config = config
}

// ---- 公開メソッド: ゲッター ----

// GetGameEndFlag ゲーム終了フラグ取得
// DurakHint はサーバーが計算した推奨手。
type DurakHint struct {
	// CardIndex は推奨する手札の位置。取る/パスを勧めるときは nil。
	CardIndex *int
	// AttackIdx は防御時に狙うテーブル上の攻撃カードの位置。攻撃時は nil。
	AttackIdx *int
	// TakeCards が true なら「引き取る」を勧める (防御できる札が無い)。
	TakeCards bool
	// Reason は理由キー (i18n で引く)。
	Reason string
}

// GetHint は人間の手番での推奨手を返す。手番でなければ nil。
//
// **他のトリック系はサーバー計算の理由付きヒントを持つのに、Durak は CUI に
// hint コマンドすら無く、Web もクライアント完結の簡易ヒューリスティックだけ
// だった (#4740)。**推奨手は CPU の選択ロジック (cpuFind*) をそのまま使う。
// 別ロジックを書くと「CPU は選ばない手を人間に勧める」ことになる。
func (d *Durak) GetHint() *DurakHint {
	if d.round.gameEndFlag || !d.IsHumanTurn() {
		return nil
	}
	human := d.players[d.round.currentTurn]

	switch d.round.phase {
	case DurakPhaseAttack:
		if len(d.round.tablePairs) == 0 {
			idx := d.cpuFindWeakestAttackCard(human)
			if idx < 0 {
				return nil
			}
			return &DurakHint{CardIndex: &idx, Reason: "attack_weakest"}
		}
		idx := d.cpuFindAdditionalAttackCard(human)
		if idx < 0 {
			// 追撃できる札が無い = パスするしかない。
			return &DurakHint{Reason: "pass_no_addition"}
		}
		return &DurakHint{CardIndex: &idx, Reason: "attack_additional"}

	case DurakPhaseDefend:
		for pairIdx, pair := range d.round.tablePairs {
			if pair.Defense != nil {
				continue
			}
			if idx := d.cpuFindDefenseCard(human, pair.Attack); idx >= 0 {
				pi := pairIdx
				return &DurakHint{CardIndex: &idx, AttackIdx: &pi, Reason: "defend_beat"}
			}
			// この攻撃を返せない = 引き取るしかない。
			return &DurakHint{TakeCards: true, Reason: "take_cannot_beat"}
		}
		return nil
	}
	return nil
}

func (d *Durak) GetGameEndFlag() bool { return d.round.gameEndFlag }

// IsHumanTurn 現在の手番が人間かを返す
func (d *Durak) IsHumanTurn() bool {
	return d.players[d.round.currentTurn].GetIsHuman()
}

// GetPlayerCnt プレイヤー数を取得する
func (d *Durak) GetPlayerCnt() int { return len(d.players) }

// GetPlayer 指定インデックスのプレイヤーを取得する
func (d *Durak) GetPlayer(i int) *DurakPlayer { return d.players[i] }

// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
func (d *Durak) GetCurrentTurn() int { return d.round.currentTurn }

// GetPhase 現在のフェーズを取得する
func (d *Durak) GetPhase() DurakPhase { return d.round.phase }

// GetAttackerIdx 攻撃者インデックスを取得する
func (d *Durak) GetAttackerIdx() int { return d.round.attackerIdx }

// GetDefenderIdx 防御者インデックスを取得する
func (d *Durak) GetDefenderIdx() int { return d.round.defenderIdx }

// GetTablePairs テーブル上のカードペアを取得する
func (d *Durak) GetTablePairs() []*DurakTablePair { return d.round.tablePairs }

// GetTrumpSuit 切り札スートを取得する
func (d *Durak) GetTrumpSuit() int { return d.trumpSuit }

// GetTrumpCard 切り札カード (山札底) を取得する
func (d *Durak) GetTrumpCard() *Card { return d.trumpCard }

// GetStockCount 山札残数を取得する
func (d *Durak) GetStockCount() int { return len(d.stock) }

// GetLoserIdx 敗者インデックスを取得する (-1 = 未確定)
func (d *Durak) GetLoserIdx() int { return d.round.loserIdx }

// GetConfig ゲーム設定を取得する
func (d *Durak) GetConfig() DurakConfig { return d.config }

// GetSortMode 現在のソートモードを取得する
func (d *Durak) GetSortMode() DurakSortMode { return d.sortMode }

// GetCpuActions CPU行動記録を取得する
func (d *Durak) GetCpuActions() []*DurakCpuAction { return d.round.cpuActions }

// GetHumanAction 人間の最後の行動記録を取得する
func (d *Durak) GetHumanAction() *DurakCpuAction { return d.round.humanAction }

// GetActionLog 棋譜を取得する
func (d *Durak) GetActionLog() []*ActionLogEntry { return d.round.actionLog }

// GetBoutNumber バウト番号を取得する
func (d *Durak) GetBoutNumber() int { return d.round.boutNumber }

// HasPendingAction ペンディングアクションがあるかを返す (Durakでは常にfalse)
func (d *Durak) HasPendingAction() bool { return false }

// ---- 内部メソッド: ゲームロジック ----

// durakCardRank カードのランク値を返す (Ace = 14)
func durakCardRank(c *Card) int {
	v := c.GetValue()
	if v == 1 { // Ace
		return 14
	}
	return v
}

// canBeat 防御カードが攻撃カードに勝てるか判定
func (d *Durak) canBeat(attack, defense *Card) bool {
	atkTrump := attack.GetDesign() == d.trumpSuit
	defTrump := defense.GetDesign() == d.trumpSuit

	// 切り札 vs 切り札: ランク比較
	if atkTrump && defTrump {
		return durakCardRank(defense) > durakCardRank(attack)
	}
	// 切り札で非切り札を倒せる
	if defTrump && !atkTrump {
		return true
	}
	// 非切り札では切り札に勝てない
	if !defTrump && atkTrump {
		return false
	}
	// 同スートでランク比較
	if attack.GetDesign() == defense.GetDesign() {
		return durakCardRank(defense) > durakCardRank(attack)
	}
	// 異スート(どちらも非切り札)では勝てない
	return false
}

// canAttackWith 攻撃カードとしてテーブルに出せるか判定
func (d *Durak) canAttackWith(card *Card) bool {
	// 初回攻撃は何でも出せる
	if len(d.round.tablePairs) == 0 {
		return true
	}
	// テーブル上の全ランクを収集
	ranks := d.tableRanks()
	return ranks[durakCardRank(card)]
}

// tableRanks テーブル上の全ランクをマップで返す
func (d *Durak) tableRanks() map[int]bool {
	ranks := make(map[int]bool)
	for _, pair := range d.round.tablePairs {
		ranks[durakCardRank(pair.Attack)] = true
		if pair.Defense != nil {
			ranks[durakCardRank(pair.Defense)] = true
		}
	}
	return ranks
}

// countUndefended 未防御のカード数を返す
func (d *Durak) countUndefended() int {
	count := 0
	for _, pair := range d.round.tablePairs {
		if pair.Defense == nil {
			count++
		}
	}
	return count
}

// canContinueAttack 攻撃者がまだカードを追加できるか判定
func (d *Durak) canContinueAttack() bool {
	attacker := d.players[d.round.attackerIdx]
	if attacker.GetCardsSize() == 0 {
		return false
	}
	defender := d.players[d.round.defenderIdx]
	if defender.GetCardsSize() == 0 {
		return false
	}
	// テーブル上のペアは最大6組まで (防御者の元手札枚数まで)
	if len(d.round.tablePairs) >= DurakHandSize {
		return false
	}
	// 攻撃者の手札にテーブルランクと一致するものがあるか
	ranks := d.tableRanks()
	for i := 0; i < attacker.GetCardsSize(); i++ {
		if ranks[durakCardRank(attacker.GetCard(i))] {
			return true
		}
	}
	return false
}

// endBout バウト終了処理
func (d *Durak) endBout(defended bool) {
	defender := d.players[d.round.defenderIdx]

	if defended {
		// 防御成功: テーブルカードを捨て札へ
		for _, pair := range d.round.tablePairs {
			d.discardPile = append(d.discardPile, pair.Attack)
			if pair.Defense != nil {
				d.discardPile = append(d.discardPile, pair.Defense)
			}
		}
		d.appendLog(-1, "bout", "defense successful, cards discarded", nil)
	} else {
		// 防御失敗: 防御者がテーブルカードを全て引き取る
		for _, pair := range d.round.tablePairs {
			defender.AddCard(pair.Attack)
			if pair.Defense != nil {
				defender.AddCard(pair.Defense)
			}
		}
		d.appendLog(-1, "bout", "defender picks up all cards", nil)
	}

	d.round.tablePairs = nil

	// 山札から補充 (攻撃者→他プレイヤー→防御者の順)
	d.replenishHands()

	// 手札ソート
	d.sortAllHands()

	// 手札がなくなったプレイヤーをfinishedに
	d.checkFinished()

	// 次のバウトの攻撃者を決定
	d.round.boutNumber++
	if defended {
		// 防御成功: 防御者が次の攻撃者
		d.round.attackerIdx = d.nextActiveOrSamePlayer(d.round.defenderIdx)
	} else {
		// 防御失敗: 防御者の次のプレイヤーが攻撃者
		d.round.attackerIdx = d.nextActivePlayer(d.round.defenderIdx)
	}

	// ゲーム終了チェック
	activePlayers := d.countActivePlayers()
	if activePlayers > 1 && d.round.boutNumber >= DurakMaxBouts {
		// 循環に入っている。引き分けにはしない -- loserIdx = -1 は「全員上がり」
		// の意味なので、混ぜると別の結末と区別できなくなる。手札が最も多い
		// プレイヤーを敗者にする。同数なら座順の早い方。
		d.round.gameEndFlag = true
		d.round.phase = DurakPhaseGameEnd
		worst, worstCnt := -1, -1
		for i, p := range d.players {
			if p.GetIsFinished() {
				continue
			}
			if n := p.GetCardsSize(); n > worstCnt {
				worst, worstCnt = i, n
			}
		}
		d.round.loserIdx = worst
		d.appendLog(-1, "bout", "bout limit reached, the player holding the most cards loses", nil)
		return
	}
	if activePlayers <= 1 {
		d.round.gameEndFlag = true
		d.round.phase = DurakPhaseGameEnd
		// 残っているプレイヤーが敗者
		for i, p := range d.players {
			if !p.GetIsFinished() {
				d.round.loserIdx = i
				break
			}
		}
		// 全員上がった場合は引き分け
		if activePlayers == 0 {
			d.round.loserIdx = -1
		}
		return
	}

	d.round.defenderIdx = d.nextActivePlayer(d.round.attackerIdx)
	d.round.currentTurn = d.round.attackerIdx
	d.round.phase = DurakPhaseAttack
}

// replenishHands 山札から手札を補充する
func (d *Durak) replenishHands() {
	if len(d.stock) == 0 {
		return
	}

	// 補充順序: 攻撃者→他のアクティブプレイヤー→防御者
	order := d.replenishOrder()
	for _, idx := range order {
		p := d.players[idx]
		for p.GetCardsSize() < DurakHandSize && len(d.stock) > 0 {
			card := d.stock[0]
			d.stock = d.stock[1:]
			p.AddCard(card)
		}
	}

	// 切り札カードが山札からなくなったらnilに
	if len(d.stock) == 0 {
		d.trumpCard = nil
	}
}

// replenishOrder 補充順序を返す (攻撃者→他→防御者)
func (d *Durak) replenishOrder() []int {
	order := []int{d.round.attackerIdx}
	// 攻撃者の次から防御者の前まで
	idx := d.nextPlayerIdx(d.round.attackerIdx)
	for idx != d.round.defenderIdx {
		if !d.players[idx].GetIsFinished() || d.players[idx].GetCardsSize() > 0 {
			order = append(order, idx)
		}
		idx = d.nextPlayerIdx(idx)
	}
	order = append(order, d.round.defenderIdx)
	return order
}

// nextPlayerIdx 次のプレイヤーインデックス (循環)
func (d *Durak) nextPlayerIdx(idx int) int {
	return (idx + 1) % len(d.players)
}

// nextActivePlayer 次のアクティブプレイヤーインデックス (finishedをスキップ)
func (d *Durak) nextActivePlayer(idx int) int {
	next := d.nextPlayerIdx(idx)
	for i := 0; i < len(d.players); i++ {
		if !d.players[next].GetIsFinished() {
			return next
		}
		next = d.nextPlayerIdx(next)
	}
	return next // 全員finished (ありえないはず)
}

// nextActiveOrSamePlayer idxがアクティブならidxを返す、そうでなければ次のアクティブ
func (d *Durak) nextActiveOrSamePlayer(idx int) int {
	if !d.players[idx].GetIsFinished() {
		return idx
	}
	return d.nextActivePlayer(idx)
}

// checkFinished 手札ゼロ+山札ゼロのプレイヤーをfinishedにする
func (d *Durak) checkFinished() {
	if len(d.stock) > 0 {
		return // 山札がある間は誰もfinishedにならない
	}
	for _, p := range d.players {
		if p.GetCardsSize() == 0 && !p.GetIsFinished() {
			p.SetIsFinished(true)
		}
	}
}

// countActivePlayers アクティブ (未finished) プレイヤー数を返す
func (d *Durak) countActivePlayers() int {
	count := 0
	for _, p := range d.players {
		if !p.GetIsFinished() {
			count++
		}
	}
	return count
}

// findFirstAttacker 最小の切り札を持つプレイヤーを返す
func (d *Durak) findFirstAttacker() int {
	bestIdx := 0
	bestRank := 100
	for i, p := range d.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c.GetDesign() == d.trumpSuit {
				r := durakCardRank(c)
				if r < bestRank {
					bestRank = r
					bestIdx = i
				}
			}
		}
	}
	return bestIdx
}

// sortAllHands 全プレイヤーの手札をソート
func (d *Durak) sortAllHands() {
	for _, p := range d.players {
		if p.GetIsHuman() {
			switch d.sortMode {
			case DurakSortBySuit:
				p.SortCards(d.trumpSuit)
			case DurakSortByValue:
				p.SortCardsByValue(d.trumpSuit)
			}
		} else {
			p.SortCards(d.trumpSuit)
		}
	}
}

// appendLog 棋譜にエントリを追加
func (d *Durak) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	d.round.appendLog(playerIdx, actionType, detail, cards)
}

// ---- 内部メソッド: CPU AI ----

// cpuAttack CPUの攻撃ロジック
func (d *Durak) cpuAttack() {
	attacker := d.players[d.round.attackerIdx]

	// 初回攻撃: 最弱の非切り札カードを出す
	if len(d.round.tablePairs) == 0 {
		idx := d.cpuFindWeakestAttackCard(attacker)
		if idx < 0 {
			return // ありえないはず
		}
		played := attacker.RemoveCard(idx)
		d.round.tablePairs = append(d.round.tablePairs, &DurakTablePair{Attack: played})
		d.round.cpuActions = append(d.round.cpuActions, &DurakCpuAction{
			PlayerIdx: d.round.attackerIdx, ActionType: DurakActionAttack, CardIdx: idx, AttackIdx: -1, Card: played,
		})
		d.appendLog(d.round.attackerIdx, "attack", fmt.Sprintf("attacks with %s", cardStr(played)), []*Card{played})

		if attacker.GetCardsSize() == 0 && len(d.stock) == 0 {
			attacker.SetIsFinished(true)
		}
		d.round.phase = DurakPhaseDefend
		d.round.currentTurn = d.round.defenderIdx
		return
	}

	// 追加攻撃: テーブルランクに一致する弱いカードを出す (難易度による)
	idx := d.cpuFindAdditionalAttackCard(attacker)
	if idx < 0 {
		// パス (追加攻撃なし)
		d.round.cpuActions = append(d.round.cpuActions, &DurakCpuAction{
			PlayerIdx: d.round.attackerIdx, ActionType: DurakActionPass, CardIdx: -1, AttackIdx: -1,
		})
		d.appendLog(d.round.attackerIdx, "pass", "stops attacking", nil)

		if d.countUndefended() > 0 {
			d.round.phase = DurakPhaseDefend
			d.round.currentTurn = d.round.defenderIdx
		} else {
			d.endBout(true)
		}
		return
	}

	played := attacker.RemoveCard(idx)
	d.round.tablePairs = append(d.round.tablePairs, &DurakTablePair{Attack: played})
	d.round.cpuActions = append(d.round.cpuActions, &DurakCpuAction{
		PlayerIdx: d.round.attackerIdx, ActionType: DurakActionAttack, CardIdx: idx, AttackIdx: -1, Card: played,
	})
	d.appendLog(d.round.attackerIdx, "attack", fmt.Sprintf("attacks with %s", cardStr(played)), []*Card{played})

	if attacker.GetCardsSize() == 0 && len(d.stock) == 0 {
		attacker.SetIsFinished(true)
	}
	d.round.phase = DurakPhaseDefend
	d.round.currentTurn = d.round.defenderIdx
}

// cpuDefend CPUの防御ロジック
func (d *Durak) cpuDefend() {
	defender := d.players[d.round.defenderIdx]

	// 未防御カードを探す
	for pairIdx, pair := range d.round.tablePairs {
		if pair.Defense != nil {
			continue
		}

		// 防御可能なカードを探す
		bestIdx := d.cpuFindDefenseCard(defender, pair.Attack)
		if bestIdx < 0 {
			// 防御不能 → カード引き取り
			d.round.cpuActions = append(d.round.cpuActions, &DurakCpuAction{
				PlayerIdx: d.round.defenderIdx, ActionType: DurakActionTake, CardIdx: -1, AttackIdx: -1,
			})
			d.appendLog(d.round.defenderIdx, "take", "picks up all table cards", nil)
			d.endBout(false)
			return
		}

		played := defender.RemoveCard(bestIdx)
		pair.Defense = played
		d.round.cpuActions = append(d.round.cpuActions, &DurakCpuAction{
			PlayerIdx: d.round.defenderIdx, ActionType: DurakActionDefend, CardIdx: bestIdx, AttackIdx: pairIdx, Card: played,
		})
		d.appendLog(d.round.defenderIdx, "defend", fmt.Sprintf("defends %s with %s", cardStr(pair.Attack), cardStr(played)), []*Card{played})
	}

	// 全て防御完了
	if d.canContinueAttack() {
		d.round.phase = DurakPhaseAttack
		d.round.currentTurn = d.round.attackerIdx
	} else {
		d.endBout(true)
	}
}

// cpuFindWeakestAttackCard 最弱のカードのインデックスを返す (非切り札優先)
func (d *Durak) cpuFindWeakestAttackCard(p *DurakPlayer) int {
	bestIdx := -1
	bestRank := 100
	bestIsTrump := true

	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		isTrump := c.GetDesign() == d.trumpSuit
		r := durakCardRank(c)

		// 非切り札を優先、同グループ内では弱い順
		if bestIdx < 0 ||
			(!isTrump && bestIsTrump) ||
			(isTrump == bestIsTrump && r < bestRank) {
			bestIdx = i
			bestRank = r
			bestIsTrump = isTrump
		}
	}
	return bestIdx
}

// cpuFindAdditionalAttackCard 追加攻撃カードのインデックスを返す (-1 = パス)
func (d *Durak) cpuFindAdditionalAttackCard(p *DurakPlayer) int {
	ranks := d.tableRanks()
	defender := d.players[d.round.defenderIdx]
	undefended := d.countUndefended()

	// 防御者の手札が足りなければ追加しない
	if undefended >= defender.GetCardsSize() {
		return -1
	}
	// テーブル上限
	if len(d.round.tablePairs) >= DurakHandSize {
		return -1
	}

	// 難易度による攻撃意欲
	switch d.config.CpuDifficulty {
	case DurakDifficultyEasy:
		// Easyは追加攻撃しない
		return -1
	case DurakDifficultyHard:
		// Hardは積極的に追加攻撃
		return d.cpuFindMatchingCard(p, ranks, true)
	default:
		// Normalは非切り札のみ追加攻撃
		return d.cpuFindMatchingCard(p, ranks, false)
	}
}

// cpuFindMatchingCard テーブルランクに一致するカードのインデックスを返す
func (d *Durak) cpuFindMatchingCard(p *DurakPlayer, ranks map[int]bool, includeTrumps bool) int {
	bestIdx := -1
	bestRank := 100

	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		r := durakCardRank(c)
		if !ranks[r] {
			continue
		}
		isTrump := c.GetDesign() == d.trumpSuit
		if isTrump && !includeTrumps {
			continue
		}
		if r < bestRank {
			bestRank = r
			bestIdx = i
		}
	}
	return bestIdx
}

// cpuFindDefenseCard 防御可能な最弱カードのインデックスを返す (-1 = 防御不能)
func (d *Durak) cpuFindDefenseCard(p *DurakPlayer, attack *Card) int {
	bestIdx := -1
	bestRank := 100
	bestIsTrump := true

	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if !d.canBeat(attack, c) {
			continue
		}

		isTrump := c.GetDesign() == d.trumpSuit
		r := durakCardRank(c)

		// 非切り札を優先、同グループ内では弱い順
		if bestIdx < 0 ||
			(!isTrump && bestIsTrump) ||
			(isTrump == bestIsTrump && r < bestRank) {
			bestIdx = i
			bestRank = r
			bestIsTrump = isTrump
		}
	}

	return bestIdx
}

// ---- JSON シリアライズ ----

// durakJSON is the JSON wire format for Durak.
type durakJSON struct {
	Players     []*DurakPlayer    `json:"ps"`
	Config      DurakConfig       `json:"cf"`
	SortMode    DurakSortMode     `json:"sm"`
	TrumpSuit   int               `json:"ts"`
	TrumpCard   *Card             `json:"tc,omitempty"`
	Stock       []*Card           `json:"st"`
	DiscardPile []*Card           `json:"dp"`
	AttackerIdx int               `json:"ai"`
	DefenderIdx int               `json:"di"`
	CurrentTurn int               `json:"ct"`
	TablePairs  []*DurakTablePair `json:"tp"`
	Phase       DurakPhase        `json:"ph"`
	GameEndFlag bool              `json:"ge"`
	LoserIdx    int               `json:"li"`
	BoutNumber  int               `json:"bn"`
	ActionLog   []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (d *Durak) MarshalJSON() ([]byte, error) {
	return json.Marshal(durakJSON{
		Players:     d.players,
		Config:      d.config,
		SortMode:    d.sortMode,
		TrumpSuit:   d.trumpSuit,
		TrumpCard:   d.trumpCard,
		Stock:       d.stock,
		DiscardPile: d.discardPile,
		AttackerIdx: d.round.attackerIdx,
		DefenderIdx: d.round.defenderIdx,
		CurrentTurn: d.round.currentTurn,
		TablePairs:  d.round.tablePairs,
		Phase:       d.round.phase,
		GameEndFlag: d.round.gameEndFlag,
		LoserIdx:    d.round.loserIdx,
		BoutNumber:  d.round.boutNumber,
		ActionLog:   d.round.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Durak) UnmarshalJSON(data []byte) error {
	var j durakJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.players = j.Players
	d.config = j.Config
	d.sortMode = j.SortMode
	d.trumpSuit = j.TrumpSuit
	d.trumpCard = j.TrumpCard
	d.stock = j.Stock
	d.discardPile = j.DiscardPile
	d.round.attackerIdx = j.AttackerIdx
	d.round.defenderIdx = j.DefenderIdx
	d.round.currentTurn = j.CurrentTurn
	d.round.tablePairs = j.TablePairs
	d.round.phase = j.Phase
	d.round.gameEndFlag = j.GameEndFlag
	d.round.loserIdx = j.LoserIdx
	d.round.boutNumber = j.BoutNumber
	d.round.actionLog = j.ActionLog
	return nil
}
