import type { Card, MachiavelliMeld, MachiavelliResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MachiavelliPhase } from '../../types/phases';

/**
 * Returns a frontend HintResult for Machiavelli, or null if no suggestion
 * applies (no human seat, empty hand, game over, or not the human's turn).
 *
 * On the human's turn it prefers forming a new meld from hand, then laying a
 * card off onto an existing table meld, and finally drawing from the stock.
 */
export function getMachiavelliHint(state: MachiavelliResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;
  const human = state.players[humanIdx];
  if (human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.currentPlayerIdx !== humanIdx) return null;
  if (state.phase !== MachiavelliPhase.TURN) return null;

  if (findHandMeld(human.cards) !== null) {
    return { targetAction: 'newMeld', reason: 'hint.newMeld', confidence: 'strong' };
  }
  if (human.cards.some((c) => state.table.some((m) => canLayoff(c, m)))) {
    return { targetAction: 'layoff', reason: 'hint.layoff', confidence: 'moderate' };
  }
  return { targetAction: 'draw', reason: 'hint.draw', confidence: 'moderate' };
}

/**
 * Find a valid meld (set or run) within the hand and return its card indices,
 * or null if none exists. A set is 3+ cards of the same value in distinct
 * suits; a run is 3+ consecutive same-suit cards.
 */
export function findHandMeld(hand: Card[]): number[] | null {
  return findSet(hand) ?? findRun(hand);
}

/** Find 3+ same-value cards in distinct suits; returns their indices or null. */
function findSet(hand: Card[]): number[] | null {
  const byValue = new Map<number, number[]>();
  for (let i = 0; i < hand.length; i++) {
    const v = hand[i].value;
    if (!byValue.has(v)) byValue.set(v, []);
    byValue.get(v)?.push(i);
  }
  for (const indices of byValue.values()) {
    const seenSuits = new Set<string>();
    const distinct: number[] = [];
    for (const idx of indices) {
      if (!seenSuits.has(hand[idx].design)) {
        seenSuits.add(hand[idx].design);
        distinct.push(idx);
      }
    }
    if (distinct.length >= 3) return distinct;
  }
  return null;
}

/** Find 3+ consecutive same-suit cards; returns their indices or null. */
function findRun(hand: Card[]): number[] | null {
  const bySuit = new Map<string, { idx: number; value: number }[]>();
  for (let i = 0; i < hand.length; i++) {
    const s = hand[i].design;
    if (!bySuit.has(s)) bySuit.set(s, []);
    bySuit.get(s)?.push({ idx: i, value: hand[i].value });
  }
  for (const cards of bySuit.values()) {
    const seen = new Set<number>();
    const unique: { idx: number; value: number }[] = [];
    for (const c of cards) {
      if (seen.has(c.value)) continue;
      seen.add(c.value);
      unique.push(c);
    }
    unique.sort((a, b) => a.value - b.value);
    let run: number[] = unique.length > 0 ? [unique[0].idx] : [];
    for (let i = 1; i < unique.length; i++) {
      if (unique[i].value === unique[i - 1].value + 1) {
        run.push(unique[i].idx);
      } else {
        if (run.length >= 3) return run;
        run = [unique[i].idx];
      }
    }
    if (run.length >= 3) return run;
  }
  return null;
}

/** Whether a hand card can be laid off onto an existing table meld. */
export function canLayoff(card: Card, meld: MachiavelliMeld): boolean {
  if (meld.cards.length === 0) return false;
  if (meld.kind === 0) {
    // Set: same value, suit not already present.
    const value = meld.cards[0].value;
    if (card.value !== value) return false;
    return !meld.cards.some((c) => c.design === card.design);
  }
  // Run: same suit, one below the min or one above the max.
  const suit = meld.cards[0].design;
  if (card.design !== suit) return false;
  const values = meld.cards.map((c) => c.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  return card.value === min - 1 || card.value === max + 1;
}
