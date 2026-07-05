import type { Card, PanResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PanPhase } from '../../types/phases';

/**
 * Panguingue (Pan) rank order. The deck omits 8/9/10, so a rope (run) treats
 * 7 and J (value 11) as consecutive. Values: A=1, 2-7, J=11, Q=12, K=13.
 */
const PAN_RANK_ORDER = [1, 2, 3, 4, 5, 6, 7, 11, 12, 13] as const;

/** Position of a card value within the Pan rank order, or -1 if absent. */
export function panRankIndex(value: number): number {
  return PAN_RANK_ORDER.indexOf(value as (typeof PAN_RANK_ORDER)[number]);
}

/**
 * Returns a frontend HintResult for Panguingue, or null if no suggestion
 * applies (no human seat, empty hand, game over, or not the human's turn).
 */
export function getPanHint(state: PanResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;
  const human = state.players[humanIdx];
  if (human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.currentPlayerIdx !== humanIdx) return null;

  if (state.phase === PanPhase.DRAW) {
    return getDrawHint(human.cards, state.discardTop);
  }
  if (state.phase === PanPhase.PLAY) {
    return getPlayHint(human.cards);
  }
  return null;
}

/** Draw phase: take the discard top when it fits the hand, otherwise draw from stock. */
function getDrawHint(hand: Card[], discardTop: Card | null): HintResult {
  if (discardTop && fitsWithHand(discardTop, hand)) {
    return { targetAction: 'drawDiscard', reason: 'hint.drawFromDiscard', confidence: 'strong' };
  }
  return { targetAction: 'drawStock', reason: 'hint.drawFromStock', confidence: 'moderate' };
}

/** Play phase: lay down a meld when one exists in hand, otherwise discard. */
function getPlayHint(hand: Card[]): HintResult {
  if (hasMeld(hand)) {
    return { targetAction: 'meld', reason: 'hint.meld', confidence: 'strong' };
  }
  return { targetAction: 'discard', reason: 'hint.discard', confidence: 'moderate' };
}

/** True when the hand already contains at least one valid 3-card set or rope. */
export function hasMeld(hand: Card[]): boolean {
  // Set: three or more of the same rank.
  const byValue = new Map<number, number>();
  for (const c of hand) byValue.set(c.value, (byValue.get(c.value) ?? 0) + 1);
  for (const count of byValue.values()) {
    if (count >= 3) return true;
  }

  // Rope: three consecutive ranks (per PAN_RANK_ORDER) in the same suit.
  const bySuit = new Map<string, number[]>();
  for (const c of hand) {
    const arr = bySuit.get(c.design) ?? [];
    arr.push(panRankIndex(c.value));
    bySuit.set(c.design, arr);
  }
  for (const idxs of bySuit.values()) {
    const uniq = [...new Set(idxs.filter((i) => i >= 0))].sort((a, b) => a - b);
    let run = 1;
    for (let i = 1; i < uniq.length; i++) {
      run = uniq[i] === uniq[i - 1] + 1 ? run + 1 : 1;
      if (run >= 3) return true;
    }
  }
  return false;
}

/** A discard-top card fits when it can extend a set or a rope already in hand. */
function fitsWithHand(card: Card, hand: Card[]): boolean {
  // Set: another card of the same rank.
  if (hand.some((c) => c.value === card.value)) return true;

  // Rope: same suit within two positions in the Pan rank order.
  const pos = panRankIndex(card.value);
  if (pos < 0) return false;
  for (const c of hand) {
    if (c.design !== card.design) continue;
    const diff = Math.abs(panRankIndex(c.value) - pos);
    if (diff === 1 || diff === 2) return true;
  }
  return false;
}
