import type { Card } from '../types/card';

/** Sheng Ji combo kinds, matching `domain.ShengJiComboKind`. */
export const SHENGJI_COMBO = {
  None: 0,
  Single: 1,
  Pair: 2,
  Tractor: 3,
} as const;

/** A combo formed by the selected cards. */
export interface ShengJiComboEval {
  /** One of {@link SHENGJI_COMBO}. */
  kind: number;
  /** The strongest card strength in the selection. */
  rank: number;
  /** Number of selected cards. */
  size: number;
  /** Whether the selected cards belong to the trump group. */
  trump: boolean;
}

const NO_TRUMP = 0;
const TIER_PLAIN = 0;
const TIER_TRUMP_SUIT = 1;
const TIER_OFF_LEVEL = 2;
const TIER_TRUMP_LEVEL = 3;
const TIER_BLACK_JOKER = 4;
const TIER_RED_JOKER = 5;

function isJoker(card: Card): boolean {
  return card.design === 'JOKER';
}

function naturalRank(card: Card): number {
  if (isJoker(card)) return 0;
  return card.value === 1 ? 14 : card.value;
}

function isLevelCard(card: Card, level: number): boolean {
  return !isJoker(card) && naturalRank(card) === level;
}

function trumpDesign(trumpSuit: number): Card['design'] | null {
  return ({ 1: 'SPADE', 2: 'CLOVER', 3: 'HEART', 4: 'DIAMOND' } as const)[trumpSuit as 1 | 2 | 3 | 4] ?? null;
}

/** Mirrors `ShengJiIsTrump`: all level cards and jokers are trumps too. */
function isTrump(card: Card, level: number, trumpSuit: number): boolean {
  return (
    isJoker(card) ||
    isLevelCard(card, level) ||
    (trumpDesign(trumpSuit) !== null && card.design === trumpDesign(trumpSuit))
  );
}

function strength(card: Card, level: number, trumpSuit: number): number {
  if (isJoker(card)) return (card.value >= 2 ? TIER_RED_JOKER : TIER_BLACK_JOKER) * 100;
  if (isLevelCard(card, level)) {
    return (
      (trumpDesign(trumpSuit) !== null && card.design === trumpDesign(trumpSuit) ? TIER_TRUMP_LEVEL : TIER_OFF_LEVEL) *
      100
    );
  }
  return (isTrump(card, level, trumpSuit) ? TIER_TRUMP_SUIT * 100 : TIER_PLAIN * 100) + naturalRank(card);
}

function sequencePosition(card: Card, level: number, trumpSuit: number): number {
  const value = strength(card, level, trumpSuit);
  if (isJoker(card) || isLevelCard(card, level)) return value;
  return naturalRank(card) > level ? value - 1 : value;
}

function cardKey(card: Card): string {
  return `${card.design}:${card.value}`;
}

function allPaired(cards: readonly Card[]): boolean {
  if (cards.length % 2 !== 0) return false;
  const counts = new Map<string, number>();
  for (const card of cards) counts.set(cardKey(card), (counts.get(cardKey(card)) ?? 0) + 1);
  return [...counts.values()].every((count) => count === 2);
}

function pairsAreConsecutive(cards: readonly Card[], level: number, trumpSuit: number): boolean {
  const positions: number[] = [];
  const seen = new Set<string>();
  for (const card of cards) {
    const key = cardKey(card);
    if (!seen.has(key)) {
      seen.add(key);
      positions.push(sequencePosition(card, level, trumpSuit));
    }
  }
  positions.sort((a, b) => a - b);
  return positions.every((position, index) => {
    if (index === 0) return true;
    const previous = positions[index - 1];
    return previous !== undefined && position === previous + 1;
  });
}

/** Classifies selected cards using the rules implemented by `ShengJiEvaluate`. */
export function shengjiEvaluate(cards: readonly Card[], level: number, trumpSuit: number): ShengJiComboEval | null {
  if (cards.length === 0) return null;
  const firstCard = cards[0];
  if (firstCard === undefined) return null;
  const firstTrump = isTrump(firstCard, level, trumpSuit);
  const firstSuit = firstTrump ? NO_TRUMP : firstCard.design;
  if (
    cards.some(
      (card) =>
        isTrump(card, level, trumpSuit) !== firstTrump ||
        (isTrump(card, level, trumpSuit) ? NO_TRUMP : card.design) !== firstSuit,
    )
  ) {
    return null;
  }
  const rank = Math.max(...cards.map((card) => strength(card, level, trumpSuit)));
  if (cards.length === 1) return { kind: SHENGJI_COMBO.Single, rank, size: 1, trump: firstTrump };
  if (!allPaired(cards)) return null;
  if (cards.length === 2) return { kind: SHENGJI_COMBO.Pair, rank, size: 2, trump: firstTrump };
  if (!pairsAreConsecutive(cards, level, trumpSuit)) return null;
  return { kind: SHENGJI_COMBO.Tractor, rank, size: cards.length, trump: firstTrump };
}
