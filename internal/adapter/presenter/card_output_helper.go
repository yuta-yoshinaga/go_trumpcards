package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// cardToOutput カードを共通WebOutputCardに変換 (nil → nil)
func cardToOutput(card *domain.Card) *controller.WebOutputCard {
	if card == nil {
		return nil
	}
	return &controller.WebOutputCard{
		Design: cardDesignToString(card.GetDesign()),
		Value:  card.GetValue(),
	}
}

// CardFace は非52枚デッキ（タロット・花札・カブ札・ウィザード等）の札を
// 手続き的に描画するための自己記述子。専用PNGアートを持たない札について、
// フロントエンドの手続き描画パス（CardFace.tsx）が必要とする情報を運ぶ。
// 詳細は ADR-0033 を参照。
type CardFace struct {
	Glyph string // 中央に描画する記号（例 "✦"）
	Label string // 隅のランク／名称ラベル（例 "Wizard", "XXI"）
	Color string // 色調トークン（例 "red", "black", "purple", "green"）
	Deck  string // デッキ系統ID（例 "wizard"）。非空なら手続き描画へ切り替える
}

// faceProvider はドメインの Card を CardFace 記述子に対応付ける。標準52枚
// デッキの札には nil を返し、その札は従来どおりPNGパスで描画される。
type faceProvider func(card *domain.Card) *CardFace

// cardToOutputWithFace はカードを WebOutputCard に変換し、faceProvider が
// 記述子を返した札には手続き描画用フィールド（Glyph/Label/Color/Deck）を
// 付与する (nil → nil)。fp が nil を返す標準札は cardToOutput と同一出力。
func cardToOutputWithFace(card *domain.Card, fp faceProvider) *controller.WebOutputCard {
	out := cardToOutput(card)
	if out == nil || fp == nil {
		return out
	}
	if face := fp(card); face != nil {
		out.Glyph = face.Glyph
		out.Label = face.Label
		out.Color = face.Color
		out.Deck = face.Deck
	}
	return out
}

// cardsToOutputWithFace はカードスライスを faceProvider 付きで変換する
// (nil → 空スライス)。
func cardsToOutputWithFace(cards []*domain.Card, fp faceProvider) []*controller.WebOutputCard {
	result := make([]*controller.WebOutputCard, 0, len(cards))
	for _, c := range cards {
		result = append(result, cardToOutputWithFace(c, fp))
	}
	return result
}

// playerCardsToOutputWithFace は cardHolder のカードを faceProvider 付きで
// WebOutputCard スライスに変換する。shouldShow が false の場合は空スライス。
// playerCardsToOutput の非52枚デッキ版。
func playerCardsToOutputWithFace(holder cardHolder, shouldShow bool, fp faceProvider) []*controller.WebOutputCard {
	if !shouldShow {
		return make([]*controller.WebOutputCard, 0)
	}
	cards := make([]*controller.WebOutputCard, 0, holder.GetCardsSize())
	for i := 0; i < holder.GetCardsSize(); i++ {
		cards = append(cards, cardToOutputWithFace(holder.GetCard(i), fp))
	}
	return cards
}

// cardHolder はインデックスベースでカードを取得できるオブジェクトの共通インターフェース
type cardHolder interface {
	GetCardsSize() int
	GetCard(i int) *domain.Card
}

// playerCardsToOutput cardHolder のカードを WebOutputCard スライスに変換する。
// shouldShow が false の場合は空スライスを返す。
func playerCardsToOutput(holder cardHolder, shouldShow bool) []*controller.WebOutputCard {
	if !shouldShow {
		return make([]*controller.WebOutputCard, 0)
	}
	cards := make([]*controller.WebOutputCard, 0, holder.GetCardsSize())
	for i := 0; i < holder.GetCardsSize(); i++ {
		cards = append(cards, cardToOutput(holder.GetCard(i)))
	}
	return cards
}

// cardsToOutput カードスライスを共通WebOutputCardスライスに変換 (nil → nil)
func cardsToOutput(cards []*domain.Card) []*controller.WebOutputCard {
	if cards == nil {
		return nil
	}
	result := make([]*controller.WebOutputCard, len(cards))
	for i, c := range cards {
		result[i] = cardToOutput(c)
	}
	return result
}

// buildTrickCards はトリックカードスライスをマッパー関数で変換する汎用ヘルパー
func buildTrickCards[TC any, OUT any](trick []TC, mapper func(TC) OUT) []OUT {
	out := make([]OUT, 0, len(trick))
	for _, tc := range trick {
		out = append(out, mapper(tc))
	}
	return out
}

// cardsToOutputOrEmpty カードスライスを共通WebOutputCardスライスに変換 (nil → 空スライス)
func cardsToOutputOrEmpty(cards []*domain.Card) []*controller.WebOutputCard {
	if cards == nil {
		return make([]*controller.WebOutputCard, 0)
	}
	return cardsToOutput(cards)
}
