import type { Card } from '../types/card';

/**
 * Skat base values per game type. Null games (base 23) are intentionally omitted: their value is
 * fixed by the contract type alone and does not scale with the matador run, so they don't belong
 * in a matador-based hand estimate.
 */
const BASE = { CLOVER: 12, SPADE: 11, HEART: 10, DIAMOND: 9, GRAND: 24 } as const;

type SuitKey = 'CLOVER' | 'SPADE' | 'HEART' | 'DIAMOND';

/** Jack rank = value 11 in our deck encoding. */
function isJack(card: Card): boolean {
  return card.value === 11;
}

/** Compute matadors ("with N" or "without N") for the player's hand for a given trump suit (or Grand). */
function matadors(hand: Card[], trump: SuitKey | 'GRAND'): { with: number; without: number } {
  // Trump order for ranking: J♣, J♠, J♥, J♦, then if trump = suit add the suit cards A,10,K,Q,9,8,7.
  const heldJacks = new Set<string>();
  for (const c of hand) if (isJack(c)) heldJacks.add(c.design);
  // Build the "top of trumps" ordered list of {suit,value} from most-significant down.
  const top: { design: string; value: number }[] = [
    { design: 'CLOVER', value: 11 },
    { design: 'SPADE', value: 11 },
    { design: 'HEART', value: 11 },
    { design: 'DIAMOND', value: 11 },
  ];
  if (trump !== 'GRAND') {
    // After jacks, trump suit's A, 10, K, Q, 9, 8, 7
    for (const v of [1, 10, 13, 12, 9, 8, 7]) {
      top.push({ design: trump, value: v });
    }
  }
  const held = new Set(hand.map((c) => `${c.design}:${c.value}`));
  // "With N": held has top[0], top[1], ..., top[N-1].
  let withN = 0;
  for (const t of top) {
    if (held.has(`${t.design}:${t.value}`)) withN += 1;
    else break;
  }
  // "Without N": player is missing top[0..N-1].
  let withoutN = 0;
  for (const t of top) {
    if (!held.has(`${t.design}:${t.value}`)) withoutN += 1;
    else break;
  }
  return { with: withN, without: withoutN };
}

/** Per-game-type estimate of the hand value reachable from the matador run alone. */
export interface SkatGameValueEstimate {
  gameType: 'CLOVER' | 'SPADE' | 'HEART' | 'DIAMOND' | 'GRAND';
  base: number;
  matadors: number;
  multiplier: number;
  /**
   * Highest bid the hand can safely accept without announcing extra multipliers
   * (no Hand / Schneider / Schwarz / Ouvert). Equals `(matadors + 1 game) × base`.
   * Accepting a bid above this risks an over-bid loss.
   */
  value: number;
}

/**
 * Compute the safe-bid ceiling per game type using just the matador run (with/without —
 * whichever is larger) plus the +1 multiplier the "game" itself always adds. The returned
 * value is the highest bid the hand can justify without invoking the optional Hand /
 * Schneider / Schwarz / Ouvert multipliers.
 */
export function skatBidEstimates(hand: Card[]): SkatGameValueEstimate[] {
  const types: Array<'CLOVER' | 'SPADE' | 'HEART' | 'DIAMOND' | 'GRAND'> = [
    'CLOVER',
    'SPADE',
    'HEART',
    'DIAMOND',
    'GRAND',
  ];
  return types.map((t) => {
    const m = matadors(hand, t);
    const matadorCount = Math.max(m.with, m.without);
    const multiplier = matadorCount + 1; // +1 for "game"
    return {
      gameType: t,
      base: BASE[t],
      matadors: matadorCount,
      multiplier,
      value: multiplier * BASE[t],
    };
  });
}

/** The single best (highest game value) estimate among all game types. */
export function skatBestBidEstimate(hand: Card[]): SkatGameValueEstimate {
  const estimates = skatBidEstimates(hand);
  return estimates.reduce((best, e) => (e.value > best.value ? e : best));
}
