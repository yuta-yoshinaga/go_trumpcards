//go:build test

package domain

// This file contains test helper methods for FollowTheQueenPlayer.
// They exist solely for cross-package test setup and are not part of the production game logic.

// SetHoleCards 伏せ札設定（テスト用）
func (p *FollowTheQueenPlayer) SetHoleCards(cards []*Card) { p.holeCards = cards }

// SetDoorCards 表向き札設定（テスト用）
func (p *FollowTheQueenPlayer) SetDoorCards(cards []*Card) { p.doorCards = cards }

// SetBestHand ベストハンド設定（テスト用）
func (p *FollowTheQueenPlayer) SetBestHand(cards []*Card) { p.bestHand = cards }

// SetPlayStyle プレイスタイル設定（テスト用）
func (p *FollowTheQueenPlayer) SetPlayStyle(style FollowTheQueenPlayStyle) { p.playStyle = style }

// SetTotalHands 総ハンド数設定（テスト用）
func (p *FollowTheQueenPlayer) SetTotalHands(n int) { p.totalHands = n }

// SetVPIPCount VPIP対象ハンド数設定（テスト用）
func (p *FollowTheQueenPlayer) SetVPIPCount(n int) { p.vpipCount = n }

// SetPFRCount PFR対象ハンド数設定（テスト用）
func (p *FollowTheQueenPlayer) SetPFRCount(n int) { p.pfrCount = n }
