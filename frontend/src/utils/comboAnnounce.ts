/** What a combo live region should say, or `null` for "say nothing". */
export interface ComboAnnouncement {
  /** i18n key, relative to the game's namespace. */
  key: 'comboAnnounce' | 'comboEnded';
  /** Interpolation count; only meaningful for `comboAnnounce`. */
  count: number;
}

/**
 * Decide what a combo live region should announce for a chain transition.
 *
 * The badge only renders at `combo >= 2`, so it *disappears* when the chain
 * breaks — and a removed element cannot be announced. Screen-reader users got
 * neither the streak nor its end, which is the whole point of the badge
 * (#5520). Keeping the region mounted and changing its text fixes both.
 *
 * Returns `null` when there is nothing to say: no chain now and none before.
 */
export function comboAnnouncement(combo: number, previous: number): ComboAnnouncement | null {
  if (combo >= 2) return { key: 'comboAnnounce', count: combo };
  // Only worth announcing a break if there was a streak to break.
  if (previous >= 2) return { key: 'comboEnded', count: 0 };
  return null;
}
