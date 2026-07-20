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

/**
 * Maps the wire `tablePlayType` (Go domain `ZhengPlayType` constants) to its
 * combo category. Kept in sync with `ZhengEval.go`.
 */
const ZHENG_PLAY_TYPE_BY_NUMBER: Record<number, ZhengCombo> = {
  0: 'invalid',
  1: 'single',
  2: 'pair',
  3: 'triple',
  4: 'straight',
  5: 'pairRun',
  6: 'bomb',
  7: 'jokerBomb',
};

/** Highest rank strength across the cards (joker bomb reaches the max, 14). */
function zhengMaxStrength(cards: Card[]): number {
  let max = -1;
  for (const c of cards) {
    const s = zhengRankStrength(c);
    if (s > max) max = s;
  }
  return max;
}

/**
 * A specific reason a selection cannot be played:
 * - `invalidType`: the cards do not form any legal combination type.
 * - `wrongType`: a legal type, but not the same type as the current table play.
 * - `wrongCount`: the right type, but a different number of cards than the table play.
 * - `tooWeak`: the right type and count, but the rank is not strictly higher.
 * - `needBomb`: a normal play cannot beat a four-of-a-kind bomb on the table.
 * - `unbeatable`: a joker bomb is on the table and cannot be beaten.
 */
export type ZhengInvalidReason = 'invalidType' | 'wrongType' | 'wrongCount' | 'tooWeak' | 'needBomb' | 'unbeatable';

/**
 * Returns the specific reason the selected cards cannot be played against the
 * current table play, or `null` when the selection is a legal play. Mirrors the
 * Go domain `zhengIsPlayable`: on a lead any legal combo is allowed; a joker
 * bomb beats everything; a four-of-a-kind bomb cuts any non-bomb and compares
 * by rank against another bomb; otherwise a normal play must match the table's
 * type and card count and be strictly higher in rank.
 */
export function zhengInvalidReason(
  cards: Card[],
  tableCards: Card[],
  tablePlayType: number,
): ZhengInvalidReason | null {
  const play = classifyZhengCombo(cards);
  if (play === 'invalid') return 'invalidType';

  // Lead: any legal combination is playable.
  if (tableCards.length === 0) return null;

  const tableType = ZHENG_PLAY_TYPE_BY_NUMBER[tablePlayType] ?? classifyZhengCombo(tableCards);

  // Joker bomb beats everything (two joker bombs cannot coexist).
  if (play === 'jokerBomb') return null;
  if (tableType === 'jokerBomb') return 'unbeatable';

  // A four-of-a-kind bomb cuts any non-bomb; bomb vs bomb compares by rank.
  if (play === 'bomb') {
    if (tableType !== 'bomb') return null;
    return zhengMaxStrength(cards) > zhengMaxStrength(tableCards) ? null : 'tooWeak';
  }
  if (tableType === 'bomb') return 'needBomb';

  // Normal play: type and count must match, and the rank must be strictly higher.
  if (play !== tableType) return 'wrongType';
  if (cards.length !== tableCards.length) return 'wrongCount';
  return zhengMaxStrength(cards) > zhengMaxStrength(tableCards) ? null : 'tooWeak';
}
