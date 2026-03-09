package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SevensGame 7並べゲームインタフェース
type SevensGame interface {
	// interactor が呼ぶメソッド
	SetConfig(config domain.SevensConfig)
	Reset()
	GetGameEndFlag() bool
	IsHumanTurn() bool
	HasAnyOption(playerIdx int) bool
	AutoHandleNoOption()
	CpuPlay()
	PlayerPlay(idx int) error
	PlayerPlayJoker(cardIdx, targetSuit, targetValue int) error
	GetCurrentTurn() int

	// presenter が呼ぶメソッド
	GetPlayerCnt() int
	GetPlayer(i int) *domain.SevensPlayer
	GetConfig() domain.SevensConfig
	GetTableMinVals() [5]int
	GetTableMaxVals() [5]int
	GetHumanAction() *domain.SevensCpuAction
	GetCpuActions() []*domain.SevensCpuAction
	GetTablePlaced() [5]uint16
	GetActionLog() []*domain.ActionLogEntry
}
