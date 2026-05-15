import type { Card, TarneebResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TarneebPhase } from '../../types/phases';

/** High card threshold for bid estimation (Q, K, A). */
const HIGH_CARD_VALUE = 12;
/** Suit labels mirroring `CardDesign*` in `internal/domain/Card.go`. */
const SUITS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;

/** Returns a frontend HintResult for Tarneeb, or null if no suggestion. */
export function getTarneebHint(state: TarneebResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === TarneebPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards, state.highestBid, state.config.minBid);
  }

  if (state.phase === TarneebPhase.TRUMP_DECLARATION) {
    if (state.bidWinnerIdx !== humanIdx) return null;
    return getTrumpHint(human.cards);
  }

  if (state.phase === TarneebPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid count from high cards + longest suit. Returns "bid:0" (pass) when below minimum. */
function getBidHint(cards: Card[], currentHigh: number, minBid: number): HintResult {
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE || c.value === 1).length;
  const suitCounts = new Map<string, number>();
  for (const c of cards) suitCounts.set(c.design, (suitCounts.get(c.design) ?? 0) + 1);
  const longest = Math.max(...Array.from(suitCounts.values()));

  let estimate = highCards + Math.max(0, longest - 4);
  if (estimate < minBid) {
    return { targetAction: 'bid:0', reason: 'hint.bidPass', confidence: 'moderate' };
  }
  if (estimate <= currentHigh) estimate = currentHigh + 1;
  if (estimate > 13) {
    return { targetAction: 'bid:0', reason: 'hint.bidPass', confidence: 'moderate' };
  }
  const confidence = highCards >= 4 ? 'strong' : 'moderate';
  return { targetAction: `bid:${estimate}`, reason: 'hint.bidEstimate', confidence };
}

/** Suggest trump = longest suit (ties broken by total card value). */
function getTrumpHint(cards: Card[]): HintResult {
  const counts = new Map<string, { len: number; value: number }>();
  for (const c of cards) {
    const e = counts.get(c.design) ?? { len: 0, value: 0 };
    e.len += 1;
    e.value += c.value === 1 ? 14 : c.value;
    counts.set(c.design, e);
  }
  let bestSuit = SUITS[0];
  let bestLen = -1;
  let bestValue = -1;
  for (const suit of SUITS) {
    const e = counts.get(suit) ?? { len: 0, value: 0 };
    if (e.len > bestLen || (e.len === bestLen && e.value > bestValue)) {
      bestSuit = suit;
      bestLen = e.len;
      bestValue = e.value;
    }
  }
  return {
    targetAction: `trump:${bestSuit}`,
    reason: 'hint.trumpLongest',
    confidence: bestLen >= 5 ? 'strong' : 'moderate',
  };
}

/**
 * Hint for play phase:
 *  - Lead → strongest non-trump if possible (preserve trumps).
 *  - Following with led suit → follow.
 *  - Void → if partner is winning, discard low; else trump if available; else discard.
 */
function getPlayHint(cards: Card[], state: TarneebResponse): HintResult {
  const trick = state.currentTrick;

  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);
  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }
  const trumpSuit = state.trumpSuit;
  const trumpDesign = trumpSuit > 0 ? SUITS[trumpSuit - 1] : null;
  if (trumpDesign && cards.some((c) => c.design === trumpDesign)) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
