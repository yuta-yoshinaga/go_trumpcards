//go:build test

package domain

// This file contains test helper methods for SevenCardStudPlayer.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetHoleCards 伏せ札設定（テスト用）
func (p *SevenCardStudPlayer) SetHoleCards(cards []*Card) { p.holeCards = cards }

// SetDoorCards 表向き札設定（テスト用）
func (p *SevenCardStudPlayer) SetDoorCards(cards []*Card) { p.doorCards = cards }

// SetBestHand ベストハンド設定（テスト用）
func (p *SevenCardStudPlayer) SetBestHand(cards []*Card) { p.bestHand = cards }

// SetPlayStyle プレイスタイル設定（テスト用）
func (p *SevenCardStudPlayer) SetPlayStyle(style SevenCardStudPlayStyle) { p.playStyle = style }

// SetTotalHands 総ハンド数設定（テスト用）
func (p *SevenCardStudPlayer) SetTotalHands(n int) { p.totalHands = n }

// SetVPIPCount VPIP対象ハンド数設定（テスト用）
func (p *SevenCardStudPlayer) SetVPIPCount(n int) { p.vpipCount = n }

// SetPFRCount PFR対象ハンド数設定（テスト用）
func (p *SevenCardStudPlayer) SetPFRCount(n int) { p.pfrCount = n }
