import type { Card } from '../types/card';
import { evaluatePartialHand, PokerHand } from './pokerSquaresUtils';

/**
 * Rough win-probability tier for a Guts declaration guide. This is a simple
 * heuristic label (not a Monte-Carlo estimate) meant to nudge the In/Out call.
 */
export type GutsGuideTier = 'high' | 'medium' | 'low';

/** A Guts declaration guide: the named hand plus a rough win-chance tier. */
export interface GutsHandGuide {
  /** i18n suffix for the hand name (`'pair'` or `'highcard'`). */
  handKey: 'pair' | 'highcard';
  /** Rough win-probability tier. */
  tier: GutsGuideTier;
  /** Estimated probability of winning after declaring In, in percent. */
  winChance: number;
}

/** Ace-high rank value: a low Ace (value 1) counts as 14, the strongest rank. */
function highRank(value: number): number {
  return value === 1 ? 14 : value;
}

function score(cards: readonly Card[]): number[] {
  const ranks = cards.map((card) => highRank(card.value)).sort((a, b) => b - a);
  return ranks[0] === ranks[1] ? [1, ranks[0]] : [0, ...ranks];
}

function compareScores(a: readonly number[], b: readonly number[]): number {
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return a[i] - b[i];
  }
  return 0;
}

function stays(cards: readonly Card[]): boolean {
  const ranks = cards.map((card) => highRank(card.value));
  return ranks[0] === ranks[1] || Math.max(...ranks) >= 11;
}

/** Estimate the win chance against CPUs following Guts' existing stay rule. */
export function estimateGutsWinChance(
  cards: readonly Card[],
  opponents: readonly { seat: number; out: boolean }[],
  humanSeat: number,
): number {
  if (cards.length !== 2) return 0;
  const deck: Card[] = [];
  for (const design of ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const) {
    for (let value = 1; value <= 13; value += 1) deck.push({ design, value });
  }
  const unavailable = new Set(cards.map((card) => `${card.design}:${card.value}`));
  const remaining = deck.filter((card) => !unavailable.has(`${card.design}:${card.value}`));
  const eligible = opponents.filter((opponent) => !opponent.out);
  if (eligible.length === 0) return 100;

  // Seed from the visible hand and table so the displayed estimate is stable.
  let seed = (cards[0].value * 7919 + cards[1].value * 104729 + humanSeat * 15485863 + eligible.length) >>> 0;
  for (const card of cards) seed = (seed * 33 + card.design.charCodeAt(0)) >>> 0;
  const random = () => {
    seed ^= seed << 13;
    seed ^= seed >>> 17;
    seed ^= seed << 5;
    return (seed >>> 0) / 4294967296;
  };
  const trials = 20000;
  let wins = 0;
  const humanScore = score(cards);
  for (let trial = 0; trial < trials; trial += 1) {
    const shuffled = [...remaining];
    for (let i = shuffled.length - 1; i > 0; i -= 1) {
      const j = Math.floor(random() * (i + 1));
      [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
    }
    let winsTrial = true;
    eligible.forEach(({ seat }, index) => {
      const hand = shuffled.slice(index * 2, index * 2 + 2);
      if (stays(hand)) {
        const comparison = compareScores(humanScore, score(hand));
        if (comparison < 0 || (comparison === 0 && seat < humanSeat)) winsTrial = false;
      }
    });
    if (winsTrial) wins += 1;
  }
  return Math.round((wins * 1000) / trials) / 10;
}

/**
 * Evaluate the human's Guts hand into a declaration guide: names the hand
 * (pair vs high card) via the shared poker partial-hand evaluator and assigns a
 * rough win-probability tier. Any pair is a strong ("high") hand; an unpaired
 * hand is "medium" when its top card is a King or Ace, otherwise "low".
 *
 * @param cards - The human's hand (typically 2 cards in Guts).
 * @returns A {@link GutsHandGuide}, or `null` for an empty hand.
 */
export function evaluateGutsGuide(
  cards: readonly Card[],
  opponents: readonly { seat: number; out: boolean }[] = [],
  humanSeat = 0,
): GutsHandGuide | null {
  if (cards.length === 0) return null;

  const made = evaluatePartialHand(cards);
  const hasPair = made !== null && made >= PokerHand.OnePair;
  if (hasPair) {
    return { handKey: 'pair', tier: 'high', winChance: estimateGutsWinChance(cards, opponents, humanSeat) };
  }

  const topRank = Math.max(...cards.map((c) => highRank(c.value)));
  const tier: GutsGuideTier = topRank >= 13 ? 'medium' : 'low';
  return { handKey: 'highcard', tier, winChance: estimateGutsWinChance(cards, opponents, humanSeat) };
}
