//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestFiftyOne テスト用に手動で構築されたゲーム状態を返す
func newTestFiftyOne() *FiftyOne {
	players := []*FiftyOnePlayer{
		NewFiftyOnePlayer(true),
		NewFiftyOnePlayer(false),
		NewFiftyOnePlayer(false),
		NewFiftyOnePlayer(false),
	}
	fo := NewFiftyOne(NewTrumpCards(0), players)
	return fo
}

// setupManualGame 手動でカードを設定したゲームを返す (Reset呼ばない)
func setupManualGame() *FiftyOne {
	players := []*FiftyOnePlayer{
		NewFiftyOnePlayer(true),
		NewFiftyOnePlayer(false),
		NewFiftyOnePlayer(false),
		NewFiftyOnePlayer(false),
	}
	fo := NewFiftyOne(NewTrumpCards(0), players)

	// 手動で手札セット
	// プレイヤー0 (人間): スペードA(11), スペード10(10), スペード5(5), ハート3(3), ダイヤ2(2) → スペード26
	players[0].AddCard(NewCard(CardDesignSpade, 1, false))
	players[0].AddCard(NewCard(CardDesignSpade, 10, false))
	players[0].AddCard(NewCard(CardDesignSpade, 5, false))
	players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	players[0].AddCard(NewCard(CardDesignDiamond, 2, false))

	// プレイヤー1: ハートA(11), ハートK(10), ハート7(7), クローバー4(4), ダイヤ6(6) → ハート28
	players[1].AddCard(NewCard(CardDesignHeart, 1, false))
	players[1].AddCard(NewCard(CardDesignHeart, 13, false))
	players[1].AddCard(NewCard(CardDesignHeart, 7, false))
	players[1].AddCard(NewCard(CardDesignClover, 4, false))
	players[1].AddCard(NewCard(CardDesignDiamond, 6, false))

	// プレイヤー2: ダイヤA(11), ダイヤK(10), ダイヤ8(8), クローバー9(9), スペード3(3) → ダイヤ29
	players[2].AddCard(NewCard(CardDesignDiamond, 1, false))
	players[2].AddCard(NewCard(CardDesignDiamond, 13, false))
	players[2].AddCard(NewCard(CardDesignDiamond, 8, false))
	players[2].AddCard(NewCard(CardDesignClover, 9, false))
	players[2].AddCard(NewCard(CardDesignSpade, 3, false))

	// プレイヤー3: クローバーA(11), クローバーK(10), クローバー7(7), スペード4(4), ハート2(2) → クローバー28
	players[3].AddCard(NewCard(CardDesignClover, 1, false))
	players[3].AddCard(NewCard(CardDesignClover, 13, false))
	players[3].AddCard(NewCard(CardDesignClover, 7, false))
	players[3].AddCard(NewCard(CardDesignSpade, 4, false))
	players[3].AddCard(NewCard(CardDesignHeart, 2, false))

	// 場札: スペードK(10), ハート9(9), ダイヤQ(10), クローバー6(6), スペード8(8)
	fo.tableCards = []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignClover, 6, false),
		NewCard(CardDesignSpade, 8, false),
	}

	fo.currentTurn = 0
	fo.phase = FiftyOnePhasePlay
	fo.stopCallerIdx = -1
	fo.gameEndFlag = false
	return fo
}

func TestNewFiftyOne(t *testing.T) {
	fo := newTestFiftyOne()
	assert.NotNil(t, fo)
	assert.Equal(t, FiftyOnePlayerCnt, fo.GetPlayerCnt())
}

func TestFiftyOne_Reset(t *testing.T) {
	fo := newTestFiftyOne()
	fo.Reset()

	// 各プレイヤー5枚、場札5枚
	for i := range FiftyOnePlayerCnt {
		assert.Equal(t, FiftyOneHandSize, fo.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, FiftyOneTableSize, len(fo.GetTableCards()))
	assert.Equal(t, FiftyOnePhasePlay, fo.GetPhase())
	assert.False(t, fo.GetGameEndFlag())
	assert.Equal(t, -1, fo.GetStopCallerIdx())
	assert.Equal(t, 0, fo.GetCurrentTurn())
	assert.True(t, fo.IsHumanTurn())
}

func TestFiftyOne_ExchangeOne(t *testing.T) {
	fo := setupManualGame()

	// 人間のターン: 手札[3](ハート3) と 場札[0](スペードK) を交換
	oldHand3 := fo.GetPlayer(0).GetCard(3)
	oldTable0 := fo.GetTableCards()[0]
	assert.Equal(t, CardDesignHeart, oldHand3.GetDesign())
	assert.Equal(t, 3, oldHand3.GetValue())
	assert.Equal(t, CardDesignSpade, oldTable0.GetDesign())
	assert.Equal(t, 13, oldTable0.GetValue())

	err := fo.ExchangeOne(3, 0)
	assert.NoError(t, err)

	// 交換後: 手札[3]がスペードKに、場札[0]がハート3に
	assert.Equal(t, CardDesignSpade, fo.GetPlayer(0).GetCard(3).GetDesign())
	assert.Equal(t, 13, fo.GetPlayer(0).GetCard(3).GetValue())
	assert.Equal(t, CardDesignHeart, fo.GetTableCards()[0].GetDesign())
	assert.Equal(t, 3, fo.GetTableCards()[0].GetValue())

	// ターンが進んでいること
	assert.Equal(t, 1, fo.GetCurrentTurn())
	assert.Equal(t, "exchange_one", fo.GetLastAction())
}

func TestFiftyOne_ExchangeOne_InvalidIndex(t *testing.T) {
	fo := setupManualGame()

	assert.Error(t, fo.ExchangeOne(-1, 0))
	assert.Error(t, fo.ExchangeOne(5, 0))
	assert.Error(t, fo.ExchangeOne(0, -1))
	assert.Error(t, fo.ExchangeOne(0, 5))
}

func TestFiftyOne_ExchangeOne_NotHumanTurn(t *testing.T) {
	fo := setupManualGame()
	fo.currentTurn = 1 // CPUのターン
	assert.ErrorIs(t, fo.ExchangeOne(0, 0), ErrNotHumanTurn)
}

func TestFiftyOne_ExchangeOne_GameEnded(t *testing.T) {
	fo := setupManualGame()
	fo.gameEndFlag = true
	assert.ErrorIs(t, fo.ExchangeOne(0, 0), ErrGameEnded)
}

func TestFiftyOne_ExchangeAll(t *testing.T) {
	fo := setupManualGame()

	oldHand := make([]*Card, FiftyOneHandSize)
	for i := range FiftyOneHandSize {
		oldHand[i] = fo.GetPlayer(0).GetCard(i)
	}
	oldTable := make([]*Card, FiftyOneTableSize)
	copy(oldTable, fo.GetTableCards())

	err := fo.ExchangeAll()
	assert.NoError(t, err)

	// 手札と場札が入れ替わっていること
	for i := range FiftyOneHandSize {
		assert.Equal(t, oldTable[i].GetDesign(), fo.GetPlayer(0).GetCard(i).GetDesign())
		assert.Equal(t, oldTable[i].GetValue(), fo.GetPlayer(0).GetCard(i).GetValue())
	}
	for i := range FiftyOneTableSize {
		assert.Equal(t, oldHand[i].GetDesign(), fo.GetTableCards()[i].GetDesign())
		assert.Equal(t, oldHand[i].GetValue(), fo.GetTableCards()[i].GetValue())
	}
	assert.Equal(t, "exchange_all", fo.GetLastAction())
	assert.Equal(t, 1, fo.GetCurrentTurn())
}

func TestFiftyOne_ExchangeAll_NotHumanTurn(t *testing.T) {
	fo := setupManualGame()
	fo.currentTurn = 1
	assert.ErrorIs(t, fo.ExchangeAll(), ErrNotHumanTurn)
}

func TestFiftyOne_ExchangeAll_GameEnded(t *testing.T) {
	fo := setupManualGame()
	fo.gameEndFlag = true
	assert.ErrorIs(t, fo.ExchangeAll(), ErrGameEnded)
}

func TestFiftyOne_Stop(t *testing.T) {
	fo := setupManualGame()

	err := fo.Stop()
	assert.NoError(t, err)

	assert.Equal(t, 0, fo.GetStopCallerIdx())
	assert.Equal(t, FiftyOnePlayerCnt-1, fo.stopRemaining) // advanceTurn が1減: 4→3
	assert.Equal(t, "stop", fo.GetLastAction())
	assert.Equal(t, 1, fo.GetCurrentTurn())
}

func TestFiftyOne_Stop_NotHumanTurn(t *testing.T) {
	fo := setupManualGame()
	fo.currentTurn = 1
	assert.ErrorIs(t, fo.Stop(), ErrNotHumanTurn)
}

func TestFiftyOne_Stop_AlreadyStopped(t *testing.T) {
	fo := setupManualGame()
	fo.stopCallerIdx = 2 // 既にプレイヤー2がストップ宣言済み
	fo.stopRemaining = 1

	err := fo.Stop()
	assert.Error(t, err) // already stopped
}

func TestFiftyOne_CpuPlay(t *testing.T) {
	fo := setupManualGame()
	fo.currentTurn = 1 // CPU1のターン

	err := fo.CpuPlay()
	assert.NoError(t, err)

	// ターンが進んでいること
	assert.Equal(t, 2, fo.GetCurrentTurn())
	// 手札サイズは変わらないこと
	assert.Equal(t, FiftyOneHandSize, fo.GetPlayer(1).GetCardsSize())
	// 場札サイズは変わらないこと
	assert.Equal(t, FiftyOneTableSize, len(fo.GetTableCards()))
	// アクションログが記録されていること
	assert.Greater(t, len(fo.GetActionLog()), 0)
}

func TestFiftyOne_CpuPlay_NotCpuTurn(t *testing.T) {
	fo := setupManualGame()
	fo.currentTurn = 0 // 人間のターン
	assert.Error(t, fo.CpuPlay())
}

func TestFiftyOne_StopAndEndGame(t *testing.T) {
	fo := setupManualGame()

	// 人間がストップ宣言
	require.NoError(t, fo.Stop())
	assert.Equal(t, 0, fo.GetStopCallerIdx())
	assert.Equal(t, 3, fo.stopRemaining)

	// CPU1プレイ → 残り2
	require.NoError(t, fo.CpuPlay())
	assert.Equal(t, 2, fo.stopRemaining)
	assert.False(t, fo.GetGameEndFlag())

	// CPU2プレイ → 残り1
	require.NoError(t, fo.CpuPlay())
	assert.Equal(t, 1, fo.stopRemaining)
	assert.False(t, fo.GetGameEndFlag())

	// CPU3プレイ → 残り0 → ゲーム終了
	require.NoError(t, fo.CpuPlay())
	assert.True(t, fo.GetGameEndFlag())
	assert.Equal(t, FiftyOnePhaseGameEnd, fo.GetPhase())
	// 勝者が設定されていること
	assert.GreaterOrEqual(t, fo.GetWinnerIdx(), 0)
	assert.Less(t, fo.GetWinnerIdx(), FiftyOnePlayerCnt)
}

func TestFiftyOne_CpuStop(t *testing.T) {
	// CPUがストップ宣言するシナリオ (高得点手札を設定)
	fo := setupManualGame()
	fo.currentTurn = 1

	// CPU1に高得点手札を設定して内部ロジックでストップが発生するかテスト
	// (実際のストップ判断はAIに依存するため、cpuCallStop内部メソッドを直接テスト)
	fo.players[1].Reset()
	fo.players[1].AddCard(NewCard(CardDesignHeart, 1, false))  // 11
	fo.players[1].AddCard(NewCard(CardDesignHeart, 13, false)) // 10
	fo.players[1].AddCard(NewCard(CardDesignHeart, 12, false)) // 10
	fo.players[1].AddCard(NewCard(CardDesignHeart, 11, false)) // 10
	fo.players[1].AddCard(NewCard(CardDesignHeart, 10, false)) // 10 → 51

	// Hardモードの場合、51点ならほぼ確実にストップ
	fo.config.CpuDifficulty = FiftyOneDifficultyHard
	err := fo.CpuPlay()
	assert.NoError(t, err)

	// 51点なら必ずストップするはず
	assert.Equal(t, 1, fo.GetStopCallerIdx())
}

func TestFiftyOne_EndGame_WinnerDetermination(t *testing.T) {
	fo := setupManualGame()
	// スコア: P0=26(スペード), P1=28(ハート), P2=29(ダイヤ), P3=28(クローバー)
	fo.endGame()

	assert.True(t, fo.GetGameEndFlag())
	assert.Equal(t, FiftyOnePhaseGameEnd, fo.GetPhase())
	assert.Equal(t, 2, fo.GetWinnerIdx()) // プレイヤー2が最高29
}

func TestFiftyOne_EndGame_TieBreaker(t *testing.T) {
	// 同点の場合、インデックスの小さい方が勝つ
	fo := setupManualGame()
	// P0とP1を同スコアにする
	fo.players[0].Reset()
	fo.players[0].AddCard(NewCard(CardDesignSpade, 1, false))  // 11
	fo.players[0].AddCard(NewCard(CardDesignSpade, 10, false)) // 10
	fo.players[0].AddCard(NewCard(CardDesignSpade, 7, false))  // 7
	fo.players[0].AddCard(NewCard(CardDesignHeart, 3, false))
	fo.players[0].AddCard(NewCard(CardDesignDiamond, 2, false))
	// P0 スペード28

	fo.players[1].Reset()
	fo.players[1].AddCard(NewCard(CardDesignHeart, 1, false))  // 11
	fo.players[1].AddCard(NewCard(CardDesignHeart, 13, false)) // 10
	fo.players[1].AddCard(NewCard(CardDesignHeart, 7, false))  // 7
	fo.players[1].AddCard(NewCard(CardDesignClover, 4, false))
	fo.players[1].AddCard(NewCard(CardDesignDiamond, 6, false))
	// P1 ハート28

	// P2, P3を低スコアにする
	fo.players[2].Reset()
	fo.players[2].AddCard(NewCard(CardDesignDiamond, 2, false))
	fo.players[2].AddCard(NewCard(CardDesignClover, 3, false))
	fo.players[2].AddCard(NewCard(CardDesignSpade, 4, false))
	fo.players[2].AddCard(NewCard(CardDesignHeart, 5, false))
	fo.players[2].AddCard(NewCard(CardDesignDiamond, 6, false))

	fo.players[3].Reset()
	fo.players[3].AddCard(NewCard(CardDesignClover, 2, false))
	fo.players[3].AddCard(NewCard(CardDesignDiamond, 3, false))
	fo.players[3].AddCard(NewCard(CardDesignSpade, 4, false))
	fo.players[3].AddCard(NewCard(CardDesignHeart, 5, false))
	fo.players[3].AddCard(NewCard(CardDesignClover, 6, false))

	fo.endGame()
	assert.Equal(t, 0, fo.GetWinnerIdx()) // 同点28はインデックスの小さい方
}

func TestFiftyOne_Getters(t *testing.T) {
	fo := setupManualGame()
	assert.Equal(t, FiftyOnePhasePlay, fo.GetPhase())
	assert.False(t, fo.GetGameEndFlag())
	assert.Equal(t, 0, fo.GetCurrentTurn())
	assert.True(t, fo.IsHumanTurn())
	assert.Equal(t, -1, fo.GetWinnerIdx())
	assert.Equal(t, -1, fo.GetStopCallerIdx())
	assert.Equal(t, 0, fo.GetTurnNumber())
	assert.Equal(t, FiftyOnePlayerCnt, fo.GetPlayerCnt())
	assert.NotNil(t, fo.GetTableCards())
	assert.Equal(t, FiftyOneTableSize, len(fo.GetTableCards()))
}

func TestFiftyOne_JSON(t *testing.T) {
	fo := setupManualGame()
	require.NoError(t, fo.ExchangeOne(3, 0)) // 1つアクション

	data, err := json.Marshal(fo)
	require.NoError(t, err)

	dst := &FiftyOne{}
	require.NoError(t, json.Unmarshal(data, dst))

	assert.Equal(t, fo.GetPhase(), dst.GetPhase())
	assert.Equal(t, fo.GetCurrentTurn(), dst.GetCurrentTurn())
	assert.Equal(t, fo.GetStopCallerIdx(), dst.GetStopCallerIdx())
	assert.Equal(t, fo.GetGameEndFlag(), dst.GetGameEndFlag())
	assert.Equal(t, fo.GetWinnerIdx(), dst.GetWinnerIdx())
	assert.Equal(t, fo.GetTurnNumber(), dst.GetTurnNumber())
	assert.Equal(t, fo.GetLastAction(), dst.GetLastAction())
	assert.Equal(t, FiftyOnePlayerCnt, dst.GetPlayerCnt())
	for i := range FiftyOnePlayerCnt {
		assert.Equal(t, fo.GetPlayer(i).GetCardsSize(), dst.GetPlayer(i).GetCardsSize())
		assert.Equal(t, fo.GetPlayer(i).GetIsHuman(), dst.GetPlayer(i).GetIsHuman())
	}
	assert.Equal(t, len(fo.GetTableCards()), len(dst.GetTableCards()))
	assert.Equal(t, len(fo.GetActionLog()), len(dst.GetActionLog()))
}

func TestFiftyOne_ActionLog(t *testing.T) {
	fo := setupManualGame()
	assert.Empty(t, fo.GetActionLog())

	require.NoError(t, fo.ExchangeOne(0, 0))
	assert.Len(t, fo.GetActionLog(), 1)
	assert.Equal(t, "exchange_one", fo.GetActionLog()[0].ActionType)
	assert.Equal(t, 0, fo.GetActionLog()[0].PlayerIdx)
}

func TestFiftyOne_Config(t *testing.T) {
	fo := newTestFiftyOne()
	cfg := fo.GetConfig()
	assert.Equal(t, FiftyOneDifficultyNormal, cfg.CpuDifficulty)

	newCfg := FiftyOneConfig{CpuDifficulty: FiftyOneDifficultyHard}
	fo.SetConfig(newCfg)
	assert.Equal(t, FiftyOneDifficultyHard, fo.GetConfig().CpuDifficulty)
}

func TestFiftyOne_FullGame(t *testing.T) {
	// 完全なゲームフロー: リセット→交換→ストップ→終了
	fo := newTestFiftyOne()
	fo.Reset()

	// 何ターンかプレイ
	for range 3 {
		if fo.GetGameEndFlag() {
			break
		}
		if fo.IsHumanTurn() {
			require.NoError(t, fo.ExchangeOne(0, 0))
		} else {
			require.NoError(t, fo.CpuPlay())
		}
	}

	// ストップ宣言
	if !fo.GetGameEndFlag() {
		// 人間のターンまで進める
		for !fo.IsHumanTurn() && !fo.GetGameEndFlag() {
			require.NoError(t, fo.CpuPlay())
		}
		if !fo.GetGameEndFlag() {
			require.NoError(t, fo.Stop())
			// 残りCPUターンをプレイ
			for !fo.GetGameEndFlag() {
				if !fo.IsHumanTurn() {
					require.NoError(t, fo.CpuPlay())
				} else {
					break
				}
			}
		}
	}
}
