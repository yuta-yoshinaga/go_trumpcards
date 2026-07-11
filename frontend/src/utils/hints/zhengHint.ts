import type { Card, ZhengResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { zhengRankStrength } from '../zhengComboValidator';

/** Backend ZhengPlayType wire values (internal/domain/ZhengEval.go). */
const PLAY_SINGLE = 1;
const PLAY_PAIR = 2;
const PLAY_TRIPLE = 3;
const PLAY_STRAIGHT = 4;
const PLAY_PAIR_RUN = 5;
const PLAY_BOMB = 6;
const PLAY_JOKER_BOMB = 7;

/** Highest strength allowed inside straights / pair runs (A = 11; 2 and jokers excluded). */
const MAX_RUN_STRENGTH = 11;

/** Counts hand cards per Zheng strength (suits are irrelevant). */
function strengthCounts(cards: Card[]): Map<number, number> {
  const counts = new Map<number, number>();
  for (const c of cards) {
    const s = zhengRankStrength(c);
    counts.set(s, (counts.get(s) ?? 0) + 1);
  }
  return counts;
}

/** Highest strength among the table cards (the play's comparison key). */
function tableKey(tableCards: Card[]): number {
  return Math.max(...tableCards.map(zhengRankStrength));
}

/** True if the hand holds both jokers (the unbeatable joker bomb). */
function hasJokerBomb(counts: Map<number, number>): boolean {
  return (counts.get(13) ?? 0) > 0 && (counts.get(14) ?? 0) > 0;
}

/** True if the hand can play a four-of-a-kind bomb stronger than minStrength. */
function hasBombAbove(counts: Map<number, number>, minStrength: number): boolean {
  for (const [s, cnt] of counts) {
    if (s <= 12 && cnt >= 4 && s > minStrength) return true;
  }
  return false;
}

/** True if the hand holds a same-size set (single/pair/triple) above the table key. */
function hasSetAbove(counts: Map<number, number>, size: number, key: number): boolean {
  for (const [s, cnt] of counts) {
    // Jokers (13/14) can only be played as singles; they never pair or triple.
    if (size > 1 && s > 12) continue;
    if (cnt >= size && s > key) return true;
  }
  return false;
}

/** True if the hand holds a run (each strength `copies` times) of `length` ranks whose top beats key. */
function hasRunAbove(counts: Map<number, number>, length: number, copies: number, key: number): boolean {
  for (let top = key + 1; top <= MAX_RUN_STRENGTH; top++) {
    const start = top - length + 1;
    if (start < 0) continue;
    let ok = true;
    for (let s = start; s <= top; s++) {
      if ((counts.get(s) ?? 0) < copies) {
        ok = false;
        break;
      }
    }
    if (ok) return true;
  }
  return false;
}

/** True if the hand can legally beat the current table play. */
function canBeatTable(cards: Card[], tableCards: Card[], tablePlayType: number): boolean {
  const counts = strengthCounts(cards);

  // The joker bomb beats everything (a second joker bomb cannot exist).
  if (tablePlayType === PLAY_JOKER_BOMB) return false;
  if (hasJokerBomb(counts)) return true;

  if (tablePlayType === PLAY_BOMB) {
    return hasBombAbove(counts, tableKey(tableCards));
  }
  // Any bomb tops any non-bomb play.
  if (hasBombAbove(counts, -1)) return true;

  const key = tableKey(tableCards);
  switch (tablePlayType) {
    case PLAY_SINGLE:
      return hasSetAbove(counts, 1, key);
    case PLAY_PAIR:
      return hasSetAbove(counts, 2, key);
    case PLAY_TRIPLE:
      return hasSetAbove(counts, 3, key);
    case PLAY_STRAIGHT:
      return hasRunAbove(counts, tableCards.length, 1, key);
    case PLAY_PAIR_RUN:
      return hasRunAbove(counts, tableCards.length / 2, 2, key);
    default:
      return false;
  }
}

/**
 * Returns a frontend HintResult for Zheng Shangyou, or null when no suggestion
 * applies. Strength follows the Zheng order 3 < ... < A < 2 < jokers with no
 * suit tiebreak: on a lead it nudges toward shedding weak cards, otherwise it
 * checks whether any legal combination can beat the table and suggests
 * playing or passing accordingly.
 */
export function getZhengHint(state: ZhengResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isFinished) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.currentTurn !== humanIdx) return null;

  // Leading: any combination is legal — shed weak cards first.
  if (state.tableCards.length === 0) {
    return { targetAction: 'play', reason: 'hint.playLow', confidence: 'moderate' };
  }

  if (canBeatTable(human.cards, state.tableCards, state.tablePlayType)) {
    return { targetAction: 'play', reason: 'hint.canPlay', confidence: 'strong' };
  }
  return { targetAction: 'pass', reason: 'hint.shouldPass', confidence: 'moderate' };
}
