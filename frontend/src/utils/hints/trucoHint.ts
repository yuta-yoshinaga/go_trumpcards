import type { Card, TrucoResponse, TrucoTrickCard } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TrucoPhase } from '../../types/phases';

/**
 * Strength of a single Truco card on the 40-card Spanish deck. Mirrors the
 * backend `TrucoCardStrength` ranking exactly: the four matadores (1♠ > 1♣ >
 * 7♠ > 7♦) sit above the common ladder (3 > 2 > 1 > K > Q > J > 7 > 6 > 5 > 4).
 * Unused ranks (8/9/10) and nil cards return 0.
 */
export function trucoCardStrength(c: Card | undefined): number {
  if (!c) return 0;
  const { value: v, design: d } = c;
  if (v === 1 && d === 'SPADE') return 14; // 1 de Espadas
  if (v === 1 && d === 'CLOVER') return 13; // 1 de Bastos
  if (v === 7 && d === 'SPADE') return 12; // 7 de Espadas
  if (v === 7 && d === 'DIAMOND') return 11; // 7 de Oros
  switch (v) {
    case 3:
      return 10;
    case 2:
      return 9;
    case 1:
      return 8;
    case 13:
      return 7;
    case 12:
      return 6;
    case 11:
      return 5;
    case 7:
      return 4;
    case 6:
      return 3;
    case 5:
      return 2;
    case 4:
      return 1;
    default:
      return 0;
  }
}

/** Highest strength among the given hand. Mirrors backend `handTopStrength`. */
function handTopStrength(cards: Card[]): number {
  let top = 0;
  for (const c of cards) {
    const s = trucoCardStrength(c);
    if (s > top) top = s;
  }
  return top;
}

/** Index of the weakest card in hand. Mirrors backend `cpuWeakestIdx`. */
function weakestIdx(cards: Card[]): number {
  let bestIdx = 0;
  let bestStrength = trucoCardStrength(cards[0]);
  for (let i = 1; i < cards.length; i++) {
    const s = trucoCardStrength(cards[i]);
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
function selectPlayIdx(cards: Card[], trick: TrucoTrickCard[]): number {
  const n = cards.length;
  if (n <= 1) return 0;
  if (trick.length === 0) return weakestIdx(cards);
  const oppStrength = trucoCardStrength(trick[0]?.card);
  let bestIdx = -1;
  let bestStrength = 0;
  for (let i = 0; i < n; i++) {
    const s = trucoCardStrength(cards[i]);
    if (s > oppStrength && (bestIdx < 0 || s < bestStrength)) {
      bestIdx = i;
      bestStrength = s;
    }
  }
  return bestIdx >= 0 ? bestIdx : weakestIdx(cards);
}

/** Reason key for the recommended play. Mirrors backend `playHintReason`. */
function playReason(card: Card | undefined, trick: TrucoTrickCard[]): string {
  if (trick.length === 0) {
    return trucoCardStrength(card) >= 11 ? 'hintReason.leadStrong' : 'hintReason.leadLow';
  }
  const opp = trick[0]?.card;
  return trucoCardStrength(card) > trucoCardStrength(opp) ? 'hintReason.followWin' : 'hintReason.followDump';
}

/**
 * Truco hint heuristic for the human player (the `isHuman` seat). Mirrors the
 * backend `Truco.GetHint`:
 *   - PLAY: recommend declaring Truco with a strong hand (top strength ≥ 11),
 *     otherwise recommend the card `cpuSelectPlayCard` would play, with a
 *     lead/follow reason.
 *   - RESPOND: accept a pending call with a strong hand (top strength ≥ 10),
 *     otherwise decline.
 * Returns null outside the human's decision turns, when the game has ended, or
 * when the hand is empty.
 */
export function getTrucoHint(state: TrucoResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0) return null;
  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;
  const trick = state.currentTrick ?? [];

  if (state.phase === TrucoPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    if (state.canDeclareTruco && handTopStrength(human.cards) >= 11) {
      return { targetAction: 'truco', reason: 'hintReason.call', confidence: 'strong' };
    }
    const idx = selectPlayIdx(human.cards, trick);
    const reason = playReason(human.cards[idx], trick);
    const confidence = reason === 'hintReason.leadStrong' || reason === 'hintReason.followWin' ? 'strong' : 'moderate';
    return { targetAction: 'play', reason, confidence };
  }

  if (state.phase === TrucoPhase.RESPOND) {
    if (state.responderIdx !== humanIdx) return null;
    if (handTopStrength(human.cards) >= 10) {
      return { targetAction: 'accept', reason: 'hintReason.accept', confidence: 'strong' };
    }
    return { targetAction: 'decline', reason: 'hintReason.decline', confidence: 'moderate' };
  }

  return null;
}
