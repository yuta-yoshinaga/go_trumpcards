/** What a page should do with its sweep badge after a state update. */
export type SweepCelebration =
  | { kind: 'none' }
  /** A sweep landed. `own` marks it as the local player's (or their team's). */
  | { kind: 'fire'; own: boolean }
  /** A counter dropped — a new round — so any showing badge is stale. */
  | { kind: 'clear' };

/**
 * Decide whether a sweep badge should fire, clear, or stay put.
 *
 * Sweeping the table is the highlight of the Scopa family and is easy to miss
 * amid fast CPU turns, so it is announced rather than left to the statistics
 * row. The counters reset to zero each round, which is a drop rather than a
 * sweep and must clear a stale badge instead of re-firing it.
 * @param prev - Per-player sweep counts from the previous update, or null on the first.
 * @param current - Per-player sweep counts now.
 * @param isOwn - Whether the player at an index counts as the viewer's own.
 * @returns The action to take.
 */
export function sweepCelebration(
  prev: readonly number[] | null,
  current: readonly number[],
  isOwn: (index: number) => boolean,
): SweepCelebration {
  // First update, or the player count changed: re-seed without firing, and drop
  // any badge left over from the previous shape.
  if (prev === null) return { kind: 'none' };
  if (prev.length !== current.length) return { kind: 'clear' };

  let gain = false;
  let own = false;
  let dropped = false;
  for (let i = 0; i < current.length; i += 1) {
    const delta = (current[i] as number) - (prev[i] as number);
    if (delta > 0) {
      gain = true;
      if (isOwn(i)) own = true;
    } else if (delta < 0) {
      dropped = true;
    }
  }
  if (gain) return { kind: 'fire', own };
  return dropped ? { kind: 'clear' } : { kind: 'none' };
}
