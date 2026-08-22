import type { Card, PutResponse, PutTrickCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PutPhase } from '../../types/phases';

/**
 * Strength of a single Put card. Mirrors the backend `PutCardStrength` in
 * `internal/domain/Put.go` exactly.
 *
 * **Suits are irrelevant and all 52 cards are used.** The order is
 * `3 > 2 > A > K > Q > J > 10 > 9 > 8 > 7 > 6 > 5 > 4`, so the 3 is the
 * strongest card and the 4 the weakest. Nil cards return 0.
 *
 * This is the one place the clone source (Truco) differs most: Truco ranks
 * four suit-specific *matadores* (1♠ > 1♣ > 7♠ > 7♦) above a common ladder and
 * plays a 40-card Spanish deck with 8/9/10 removed. Carrying that here made the
 * in-app hint call an 8, 9 or 10 worthless when they are really the 5th, 6th
 * and 7th strongest ranks.
 */
export function putCardStrength(c: Card | undefined): number {
  if (!c) return 0;
  switch (c.value) {
    case 3:
      return 13;
    case 2:
      return 12;
    case 1: // A
      return 11;
    default:
      // K(13)=10, Q(12)=9, J(11)=8, 10=7, ... 4=1 — monotonically decreasing.
      return c.value >= 4 && c.value <= 13 ? c.value - 3 : 0;
  }
}

/** Highest strength among the given hand. Mirrors backend `handTopStrength`. */
function handTopStrength(cards: Card[]): number {
  let top = 0;
  for (const c of cards) {
    const s = putCardStrength(c);
    if (s > top) top = s;
  }
  return top;
}

/** Index of the weakest card in hand. Mirrors backend `cpuWeakestIdx`. */
function weakestIdx(cards: Card[]): number {
  let bestIdx = 0;
  let bestStrength = putCardStrength(cards[0]);
  for (let i = 1; i < cards.length; i++) {
    const s = putCardStrength(cards[i]);
    if (s < bestStrength) {
      bestStrength = s;
      bestIdx = i;
    }
  }
  return bestIdx;
}

/**
 * Index of the card the heuristic recommends playing. Mirrors backend
 * `cpuSelectPlayCard`: when leading, dump the weakest card; when following,
 * play the cheapest card that still beats the opponent, else dump the weakest.
 */
function selectPlayIdx(cards: Card[], trick: PutTrickCard[]): number {
  const n = cards.length;
  if (n <= 1) return 0;
  if (trick.length === 0) return weakestIdx(cards);
  const oppStrength = putCardStrength(trick[0]?.card);
  let bestIdx = -1;
  let bestStrength = 0;
  for (let i = 0; i < n; i++) {
    const s = putCardStrength(cards[i]);
    if (s > oppStrength && (bestIdx < 0 || s < bestStrength)) {
      bestIdx = i;
      bestStrength = s;
    }
  }
  return bestIdx >= 0 ? bestIdx : weakestIdx(cards);
}

/** Reason key for the recommended play. Mirrors backend `playHintReason`. */
function playReason(card: Card | undefined, trick: PutTrickCard[]): string {
  if (trick.length === 0) {
    return putCardStrength(card) >= 11 ? 'hintReason.leadStrong' : 'hintReason.leadLow';
  }
  const opp = trick[0]?.card;
  return putCardStrength(card) > putCardStrength(opp) ? 'hintReason.followWin' : 'hintReason.followDump';
}

/**
 * Put hint heuristic for the human player (the `isHuman` seat). Mirrors the
 * backend `Put.GetHint`:
 *   - PLAY: recommend declaring Put with a strong hand (top strength ≥ 11),
 *     otherwise recommend the card `cpuSelectPlayCard` would play, with a
 *     lead/follow reason.
 *   - RESPOND: accept a pending call with a strong hand (top strength ≥ 10),
 *     otherwise decline.
 * Returns null outside the human's decision turns, when the game has ended, or
 * when the hand is empty.
 */
export function getPutHint(state: PutResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0) return null;
  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;
  const trick = state.currentTrick ?? [];

  if (state.phase === PutPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    if (state.canDeclarePut && handTopStrength(human.cards) >= 11) {
      return { targetAction: 'put', reason: 'hintReason.call', confidence: 'strong' };
    }
    const idx = selectPlayIdx(human.cards, trick);
    const reason = playReason(human.cards[idx], trick);
    const confidence = reason === 'hintReason.leadStrong' || reason === 'hintReason.followWin' ? 'strong' : 'moderate';
    return { targetAction: 'play', reason, confidence };
  }

  if (state.phase === PutPhase.RESPOND) {
    if (state.responderIdx !== humanIdx) return null;
    if (handTopStrength(human.cards) >= 10) {
      return { targetAction: 'accept', reason: 'hintReason.accept', confidence: 'strong' };
    }
    return { targetAction: 'decline', reason: 'hintReason.decline', confidence: 'moderate' };
  }

  return null;
}
