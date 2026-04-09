//go:build test

package domain

// This file contains test helper functions for PaiGow.
// They exist solely for cross-package test setup
// and are not part of the production game logic.

// EvalPaiGowHighHandExported exports evalPaiGowHighHand for testing.
func EvalPaiGowHighHandExported(cards []*Card) int {
	return evalPaiGowHighHand(cards)
}

// EvalPaiGowLowHandExported exports evalPaiGowLowHand for testing.
func EvalPaiGowLowHandExported(cards []*Card) int {
	return evalPaiGowLowHand(cards)
}

// ComparePaiGowHighHandsExported exports comparePaiGowHighHands for testing.
func ComparePaiGowHighHandsExported(a, b []*Card) int {
	return comparePaiGowHighHands(a, b)
}

// ComparePaiGowLowHandsExported exports comparePaiGowLowHands for testing.
func ComparePaiGowLowHandsExported(a, b []*Card) int {
	return comparePaiGowLowHands(a, b)
}

// PaiGowHouseWayExported exports paiGowHouseWay for testing.
func PaiGowHouseWayExported(cards []*Card) (high []*Card, low []*Card) {
	return paiGowHouseWay(cards)
}
