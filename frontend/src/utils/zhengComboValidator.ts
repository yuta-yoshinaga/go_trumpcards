import type { Card } from '../types/card';

/**
 * Zheng Shangyou rank strength: 3..K = 0..10, A = 11, 2 = 12,
 * small joker = 13, big joker = 14. Suits never matter.
 * Mirrors the Go domain `zhengRankStrength`.
 */
export function zhengRankStrength(card: Card): number {
  if (card.design === 'JOKER') {
    return card.value === 2 ? 14 : 13;
  }
  if (card.value === 2) return 12;
  if (card.value === 1) return 11;
  return card.value - 3;
}

/** True if every card shares one non-joker rank (jokers never pair with anything). */
function allSameNonJokerValue(cards: Card[]): boolean {
  if (cards.length === 0) return false;
  if (cards.some((c) => c.design === 'JOKER')) return false;
  return cards.every((c) => c.value === cards[0].value);
}

/** True if the cards are exactly the small + big joker (the strongest bomb). */
function isJokerBomb(cards: Card[]): boolean {
  if (cards.length !== 2) return false;
  if (cards[0].design !== 'JOKER' || cards[1].design !== 'JOKER') return false;
  const [v0, v1] = [cards[0].value, cards[1].value];
  return (v0 === 1 && v1 === 2) || (v0 === 2 && v1 === 1);
}

/** A run of 3+ consecutive ranks; 2 and jokers cannot appear, suits are irrelevant. */
function checkStraight(cards: Card[]): boolean {
  if (cards.length < 3) return false;
  if (cards.some((c) => c.design === 'JOKER' || c.value === 2)) return false;
  const strengths = cards.map(zhengRankStrength).sort((a, b) => a - b);
  for (let i = 1; i < strengths.length; i++) {
    if (strengths[i] !== strengths[i - 1] + 1) return false;
  }
  return true;
}

/** Three or more consecutive pairs (e.g. 44-55-66); 2 and jokers cannot appear. */
function checkPairRun(cards: Card[]): boolean {
  const n = cards.length;
  if (n < 6 || n % 2 !== 0) return false;
  if (cards.some((c) => c.design === 'JOKER' || c.value === 2)) return false;
  const freq = new Map<number, number>();
  for (const c of cards) {
    const s = zhengRankStrength(c);
    freq.set(s, (freq.get(s) ?? 0) + 1);
  }
  if (freq.size !== n / 2) return false;
  const strengths: number[] = [];
  for (const [s, cnt] of freq) {
    if (cnt !== 2) return false;
    strengths.push(s);
  }
  strengths.sort((a, b) => a - b);
  for (let i = 1; i < strengths.length; i++) {
    if (strengths[i] !== strengths[i - 1] + 1) return false;
  }
  return true;
}

/** The Zheng Shangyou play categories. */
export type ZhengCombo = 'single' | 'pair' | 'triple' | 'straight' | 'pairRun' | 'bomb' | 'jokerBomb' | 'invalid';

/**
 * Classifies a selection of cards into its Zheng Shangyou play type, mirroring
 * the Go domain `zhengClassifyPlay`.
 */
export function classifyZhengCombo(cards: Card[]): ZhengCombo {
  const n = cards.length;
  if (n === 0) return 'invalid';
  if (n === 1) return 'single';
  if (n === 2) {
    if (isJokerBomb(cards)) return 'jokerBomb';
    return allSameNonJokerValue(cards) ? 'pair' : 'invalid';
  }
  if (n === 3) {
    if (allSameNonJokerValue(cards)) return 'triple';
    return checkStraight(cards) ? 'straight' : 'invalid';
  }
  if (n === 4) {
    if (allSameNonJokerValue(cards)) return 'bomb';
    return checkStraight(cards) ? 'straight' : 'invalid';
  }
  if (checkPairRun(cards)) return 'pairRun';
  return checkStraight(cards) ? 'straight' : 'invalid';
}

/** True if the selected cards form any legal Zheng Shangyou combination. */
export function isValidZhengCombo(cards: Card[]): boolean {
  return cards.length > 0 && classifyZhengCombo(cards) !== 'invalid';
}
