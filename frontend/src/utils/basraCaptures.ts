import type { Card } from '../types/card';

/**
 * Pure Basra (Bastra) capture logic, mirroring the backend domain rule
 * (`internal/domain/Basra.go`) so the web GUI can preview — before the human
 * commits a play — exactly which table cards a selected hand card would capture.
 *
 * Card values follow the shared {@link Card} convention (`value`): Ace = 1,
 * number cards 2–10, Jack = 11, Queen = 12, King = 13.
 */

/** The value of a Jack — Basra's board-sweeping card. */
const BASRA_JACK_VALUE = 11;

/** True when the card is a Jack (the sweep card in Basra). */
export function isBasraJack(card: Card | null | undefined): boolean {
  return !!card && card.value === BASRA_JACK_VALUE;
}

/** True when the card is a face card (J/Q/K); faces never join numeric sum groups. */
function isBasraFace(card: Card): boolean {
  return card.value >= 11;
}

/**
 * Recursively searches `avail` (indices into `tableCards`) for the first subset of
 * two or more cards whose values sum to `target`. Mirrors `basraSubsetSum`.
 */
function subsetSum(tableCards: Card[], avail: number[], start: number, target: number, acc: number[]): number[] | null {
  if (target === 0) {
    return acc.length >= 2 ? [...acc] : null;
  }
  for (let i = start; i < avail.length; i++) {
    const v = tableCards[avail[i]].value;
    if (v > target) continue;
    const res = subsetSum(tableCards, avail, i + 1, target - v, [...acc, avail[i]]);
    if (res !== null) return res;
  }
  return null;
}

/**
 * Finds one capturing group for a number card of value `target`: a same-rank single
 * is preferred, otherwise a subset of two or more number cards summing to `target`.
 * `used` marks table indices already claimed by earlier groups. Mirrors
 * `basraFindOneGroup`. Returns `null` when no further group exists.
 */
function findOneGroup(tableCards: Card[], target: number, used: boolean[]): number[] | null {
  for (let i = 0; i < tableCards.length; i++) {
    if (used[i] || isBasraFace(tableCards[i])) continue;
    if (tableCards[i].value === target) return [i];
  }
  const avail: number[] = [];
  for (let i = 0; i < tableCards.length; i++) {
    if (used[i] || isBasraFace(tableCards[i])) continue;
    if (tableCards[i].value < target) avail.push(i);
  }
  return subsetSum(tableCards, avail, 0, target, []);
}

/**
 * Returns the maximal set of table indices the played card would capture, mirroring
 * `Basra.basraFindCaptures`:
 *  - a Jack sweeps every non-Jack table card;
 *  - a Queen/King captures same-rank table cards only (no sums);
 *  - a number card captures same-rank cards and any table subset summing to its value,
 *    greedily extracting multiple groups.
 *
 * The result is sorted ascending. An empty array means the card captures nothing and
 * would trail.
 */
export function basraFindCaptures(playedCard: Card, tableCards: Card[]): number[] {
  if (isBasraJack(playedCard)) {
    const out: number[] = [];
    for (let i = 0; i < tableCards.length; i++) {
      if (!isBasraJack(tableCards[i])) out.push(i);
    }
    return out;
  }
  const pv = playedCard.value;
  if (isBasraFace(playedCard)) {
    const out: number[] = [];
    for (let i = 0; i < tableCards.length; i++) {
      if (tableCards[i].value === pv) out.push(i);
    }
    return out;
  }
  const used = new Array<boolean>(tableCards.length).fill(false);
  const captured: number[] = [];
  for (;;) {
    const group = findOneGroup(tableCards, pv, used);
    if (group === null) break;
    for (const idx of group) {
      used[idx] = true;
      captured.push(idx);
    }
  }
  captured.sort((a, b) => a - b);
  return captured;
}

/** The action the Basra footer button will perform for the current selection. */
export type BasraActionKind = 'idle' | 'sweep' | 'capture' | 'capturable' | 'trail';

/** Resolved action + the number of table cards it would capture. */
export interface BasraAction {
  kind: BasraActionKind;
  /** Table cards captured by the action (0 for `idle`/`trail`/`capturable`-preview). */
  count: number;
}

/**
 * Resolves what playing the selected hand card will do, mirroring `Basra.applyPlay`:
 *  - no card selected → `idle`;
 *  - a Jack → `sweep` (captures every non-Jack table card; selection is ignored);
 *  - a non-Jack with table cards selected → `capture` (captures the selection);
 *  - a non-Jack with a capture available but nothing selected → `capturable` (the
 *    human still needs to tap the highlighted cards; pressing now would trail);
 *  - otherwise → `trail`.
 *
 * @param card selected hand card, or `null` when none is chosen
 * @param captures maximal capturable table indices for `card` ({@link basraFindCaptures})
 * @param selectedTableIndices table indices the human has toggled for capture
 */
export function resolveBasraAction(card: Card | null, captures: number[], selectedTableIndices: number[]): BasraAction {
  if (!card) return { kind: 'idle', count: 0 };
  if (isBasraJack(card)) return { kind: 'sweep', count: captures.length };
  if (selectedTableIndices.length > 0) return { kind: 'capture', count: selectedTableIndices.length };
  if (captures.length > 0) return { kind: 'capturable', count: captures.length };
  return { kind: 'trail', count: 0 };
}
