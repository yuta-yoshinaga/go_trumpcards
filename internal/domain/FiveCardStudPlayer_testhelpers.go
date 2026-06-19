//go:build test

package domain

// This file contains test helper methods for FiveCardStudPlayer.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetHoleCards 伏せ札設定（テスト用）
func (p *FiveCardStudPlayer) SetHoleCards(cards []*Card) { p.holeCards = cards }

// SetDoorCards 表向き札設定（テスト用）
func (p *FiveCardStudPlayer) SetDoorCards(cards []*Card) { p.doorCards = cards }

// SetBestHand ベストハンド設定（テスト用）
func (p *FiveCardStudPlayer) SetBestHand(cards []*Card) { p.bestHand = cards }

// SetPlayStyle プレイスタイル設定（テスト用）
func (p *FiveCardStudPlayer) SetPlayStyle(style FiveCardStudPlayStyle) { p.playStyle = style }

// SetTotalHands 総ハンド数設定（テスト用）
func (p *FiveCardStudPlayer) SetTotalHands(n int) { p.totalHands = n }

// SetVPIPCount VPIP対象ハンド数設定（テスト用）
func (p *FiveCardStudPlayer) SetVPIPCount(n int) { p.vpipCount = n }

// SetPFRCount PFR対象ハンド数設定（テスト用）
func (p *FiveCardStudPlayer) SetPFRCount(n int) { p.pfrCount = n }
