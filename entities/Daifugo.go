package entities

import (
	"math/rand"
	"sort"
)

// DaifugoPlayerCnt 大富豪プレイヤー数
const DaifugoPlayerCnt = 4

// Daifugo 大富豪ゲームクラス
type Daifugo struct {
	trumpCards     *TrumpCards
	players        []*DaifugoPlayer
	currentTurn    int
	gameEndFlag    bool
	lastPlay       []*Card // 最後に場に出されたカード
	lastPlayPlayer int     // 最後に場にカードを出したプレイヤーインデックス (-1: 場が流れた状態)
	passCount      int     // 連続パス数
	isRevolution   bool    // 革命状態フラグ
	finishedCount  int     // 上がったプレイヤー数
	rankHistory    []int   // 今回の順位 (プレイヤーインデックスのリスト)
}

// NewDaifugo コンストラクタ
func NewDaifugo(trumpCards *TrumpCards, players []*DaifugoPlayer) *Daifugo {
	return &Daifugo{
		trumpCards:     trumpCards,
		players:        players,
		currentTurn:    0,
		gameEndFlag:    false,
		lastPlay:       nil,
		lastPlayPlayer: -1,
		passCount:      0,
		isRevolution:   false,
		finishedCount:  0,
		rankHistory:    make([]int, 0),
	}
}

// Reset ゲーム初期化 (新しいゲーム開始)
func (d *Daifugo) Reset() {
	d.gameEndFlag = false
	d.lastPlay = nil
	d.lastPlayPlayer = -1
	d.passCount = 0
	d.isRevolution = false
	d.finishedCount = 0
	d.rankHistory = make([]int, 0)

	// シャッフル
	for i := 0; i < 10; i++ {
		d.trumpCards.Shuffle()
	}

	// 全プレイヤーのカードリセット
	for _, p := range d.players {
		p.Reset()
		p.SetIsFinished(false)
	}

	// 全カードを配る
	idx := 0
	for {
		card := d.trumpCards.DrawCard()
		if card == nil {
			break
		}
		d.players[idx%DaifugoPlayerCnt].AddCard(card)
		idx++
	}

	// 手札をソート
	for _, p := range d.players {
		p.SortCards()
	}

	// ダイヤの3を持っているプレイヤーが最初の手番
	d.currentTurn = 0
	for i, p := range d.players {
		for _, c := range p.GetCards() {
			if c.GetDesign() == CardDesignDiamond && c.GetValue() == 3 {
				d.currentTurn = i
				break
			}
		}
	}
}

// PlayCards カードを場に出す
func (d *Daifugo) PlayCards(playerIdx int, cardIndices []int) bool {
	if d.gameEndFlag || playerIdx != d.currentTurn {
		return false
	}

	player := d.players[playerIdx]
	if player.GetIsFinished() {
		return false
	}

	// パスの場合
	if len(cardIndices) == 0 {
		return d.Pass(playerIdx)
	}

	// 出せるかチェック
	cards := make([]*Card, 0, len(cardIndices))
	for _, idx := range cardIndices {
		cards = append(cards, player.GetCards()[idx])
	}

	if !d.CanPlay(cards) {
		return false
	}

	// カードを出す
	player.RemoveCards(cardIndices)
	d.lastPlay = cards
	d.lastPlayPlayer = playerIdx
	d.passCount = 0

	// 革命チェック (4枚以上で革命)
	if len(cards) >= 4 {
		d.isRevolution = !d.isRevolution
	}

	// 上がりチェック
	if player.GetCardsSize() == 0 {
		player.SetIsFinished(true)
		d.finishedCount++
		d.rankHistory = append(d.rankHistory, playerIdx)
		player.SetRank(d.finishedCount - 1)
	}

	// ゲーム終了チェック
	if d.finishedCount >= DaifugoPlayerCnt-1 {
		// 最後の一人を確定
		for i, p := range d.players {
			if !p.GetIsFinished() {
				p.SetIsFinished(true)
				d.rankHistory = append(d.rankHistory, i)
				p.SetRank(DaifugoPlayerCnt - 1)
				break
			}
		}
		d.gameEndFlag = true
	} else {
		d.advanceTurn()
	}

	return true
}

// Pass パスする
func (d *Daifugo) Pass(playerIdx int) bool {
	if d.gameEndFlag || playerIdx != d.currentTurn {
		return false
	}

	d.passCount++
	
	// 自分以外全員パス、または自分以外上がっている状態でパスが回ってきたら場を流す
	activePlayers := 0
	for _, p := range d.players {
		if !p.GetIsFinished() {
			activePlayers++
		}
	}

	if d.passCount >= activePlayers-1 {
		d.ClearField()
	} else {
		d.advanceTurn()
	}

	return true
}

// ClearField 場を流す
func (d *Daifugo) ClearField() {
	d.lastPlay = nil
	d.lastPlayPlayer = -1
	d.passCount = 0
	
	// 次のターンは、最後にカードを出した人から。
	// ただし、その人が上がっていたら、その次のアクティブな人から。
	if d.lastPlayPlayer != -1 && !d.players[d.lastPlayPlayer].GetIsFinished() {
		d.currentTurn = d.lastPlayPlayer
	} else {
		// すでに次のプレイヤーに進んでいるはずなので、そのままにする
		// または、現在手番の人が上がっていないことを確認
		if d.players[d.currentTurn].GetIsFinished() {
			d.advanceTurn()
		}
	}
}

// advanceTurn 次のプレイヤーへ
func (d *Daifugo) advanceTurn() {
	for i := 1; i <= DaifugoPlayerCnt; i++ {
		next := (d.currentTurn + i) % DaifugoPlayerCnt
		if !d.players[next].GetIsFinished() {
			d.currentTurn = next
			break
		}
	}
}

// CanPlay カードが出せるか判定
func (d *Daifugo) CanPlay(cards []*Card) bool {
	if len(cards) == 0 {
		return false
	}

	// 出すカードがすべて同じ強さかチェック
	strength := GetDaifugoStrength(cards[0])
	for i := 1; i < len(cards); i++ {
		if GetDaifugoStrength(cards[i]) != strength {
			return false
		}
	}

	// 最初のターンの場合
	if d.lastPlay == nil {
		return true
	}

	// 枚数が同じである必要がある
	if len(cards) != len(d.lastPlay) {
		return false
	}

	// 強さ比較
	lastStrength := GetDaifugoStrength(d.lastPlay[0])
	if d.isRevolution {
		return strength < lastStrength
	}
	return strength > lastStrength
}

// CpuPlay CPUの思考
func (d *Daifugo) CpuPlay() {
	if d.gameEndFlag || d.players[d.currentTurn].GetIsHuman() {
		return
	}

	player := d.players[d.currentTurn]
	hand := player.GetCards()

	// 可能な組み合わせを探す (strength -> indices)
	groups := make(map[int][]int)
	strengths := make([]int, 0)
	for i, c := range hand {
		s := GetDaifugoStrength(c)
		if _, ok := groups[s]; !ok {
			strengths = append(strengths, s)
		}
		groups[s] = append(groups[s], i)
	}
	sort.Ints(strengths)

	var bestPlay []int

	if d.lastPlay == nil {
		// 自分が親なら、一番弱いカード(群)を出す
		if len(strengths) > 0 {
			targetStrength := strengths[0]
			if d.isRevolution {
				targetStrength = strengths[len(strengths)-1]
			}
			// 一番弱いランクのカードを1枚だけ出す (シンプルに)
			bestPlay = []int{groups[targetStrength][0]}
		}
	} else {
		// 場にカードがある場合
		requiredCnt := len(d.lastPlay)
		lastStrength := GetDaifugoStrength(d.lastPlay[0])

		if !d.isRevolution {
			for _, s := range strengths {
				if len(groups[s]) >= requiredCnt && s > lastStrength {
					bestPlay = groups[s][:requiredCnt]
					break
				}
			}
		} else {
			for i := len(strengths) - 1; i >= 0; i-- {
				s := strengths[i]
				if len(groups[s]) >= requiredCnt && s < lastStrength {
					bestPlay = groups[s][:requiredCnt]
					break
				}
			}
		}
	}

	if bestPlay != nil {
		d.PlayCards(d.currentTurn, bestPlay)
	} else {
		d.Pass(d.currentTurn)
	}
}

// GetCurrentTurn 取得
func (d *Daifugo) GetCurrentTurn() int {
	return d.currentTurn
}

// GetPlayers 取得
func (d *Daifugo) GetPlayers() []*DaifugoPlayer {
	return d.players
}

// GetLastPlay 取得
func (d *Daifugo) GetLastPlay() []*Card {
	return d.lastPlay
}

// GetIsRevolution 取得
func (d *Daifugo) GetIsRevolution() bool {
	return d.isRevolution
}

// GetGameEndFlag 取得
func (d *Daifugo) GetGameEndFlag() bool {
	return d.gameEndFlag
}
