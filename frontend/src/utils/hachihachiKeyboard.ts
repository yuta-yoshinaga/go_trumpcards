/**
 * Digit-key resolution for Hachi-Hachi.
 *
 * The number keys address two different rows depending on what is pending: the
 * hand normally, but the field once a hand card matches two field cards and the
 * player still owes a choice. Keeping that decision here — rather than inline in
 * the page — means both the click and keyboard paths run the same rules and the
 * rules can be exercised without mounting the board.
 */

/** Field indices a hand card can capture, keyed by hand index. */
export type HachiHachiCaptureOptions = Record<number, number[]>;

/** What activating index `n` should do. */
export type HachiHachiAction =
  | { kind: 'none' }
  | { kind: 'select'; handIndex: number }
  | { kind: 'play'; handIndex: number; fieldIndex?: number };

/**
 * Field indices the currently-selected hand card can capture. Empty when
 * nothing is selected, which is also how "no choice is owed" is represented.
 * @param captureOptions - Per-hand-card capture options from the backend.
 * @param handIndex - Currently selected hand index, or null.
 * @returns The capturable field indices.
 */
export function hachiHachiPendingCandidates(
  captureOptions: HachiHachiCaptureOptions,
  handIndex: number | null,
): number[] {
  if (handIndex === null) return [];
  return captureOptions[handIndex] ?? [];
}

/**
 * Resolves activating index `n`, whether it came from a click or a digit key.
 *
 * With a two-way match pending the index names a field card, and anything the
 * selection cannot capture is ignored. Otherwise it names a hand card: a card
 * with more than one match only gets selected (the field choice comes next),
 * while every other card is played immediately since the backend resolves the
 * capture on its own.
 * @param captureOptions - Per-hand-card capture options from the backend.
 * @param handIndex - Currently selected hand index, or null.
 * @param index - The activated index.
 * @returns The action to dispatch.
 */
export function hachiHachiAction(
  captureOptions: HachiHachiCaptureOptions,
  handIndex: number | null,
  index: number,
): HachiHachiAction {
  const pending = hachiHachiPendingCandidates(captureOptions, handIndex);
  if (pending.length > 1) {
    if (handIndex === null || !pending.includes(index)) return { kind: 'none' };
    return { kind: 'play', handIndex, fieldIndex: index };
  }
  const options = captureOptions[index] ?? [];
  if (options.length > 1) return { kind: 'select', handIndex: index };
  return { kind: 'play', handIndex: index };
}
