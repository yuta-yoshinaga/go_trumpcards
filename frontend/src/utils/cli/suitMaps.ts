/**
 * Shared suit name/abbreviation → internal suit number maps for CLI command
 * parsers.
 *
 * Historically each `*Commands.ts` parser declared its own local `SUIT_MAP`,
 * which drifted: most games omitted the singular `club` alias that Mighty
 * accepted, so `club` worked in one game but was rejected as an invalid
 * command in others. Centralizing the canonical map here removes that
 * inconsistency and gives a single place to add future aliases.
 */

/**
 * Standard four-suit map (Spade=1, Club=2, Heart=3, Diamond=4).
 *
 * Accepts both `club` (singular) and `clubs`/`clover` so the alias set is
 * uniform across every game that uses standard suit ordering.
 */
export const STANDARD_SUIT_MAP: Readonly<Record<string, number>> = {
  spade: 1,
  spades: 1,
  s: 1,
  club: 2,
  clubs: 2,
  clover: 2,
  c: 2,
  heart: 3,
  hearts: 3,
  h: 3,
  diamond: 4,
  diamonds: 4,
  d: 4,
} as const;

/**
 * Bridge-specific suit ranking (Club=1, Diamond=2, Heart=3, Spade=4, NT=5).
 * This ordering is intentional and must not be merged with the standard map.
 */
export const BRIDGE_SUIT_MAP: Readonly<Record<string, number>> = {
  club: 1,
  clubs: 1,
  clover: 1,
  c: 1,
  diamond: 2,
  diamonds: 2,
  d: 2,
  heart: 3,
  hearts: 3,
  h: 3,
  spade: 4,
  spades: 4,
  s: 4,
  notrump: 5,
  nt: 5,
} as const;
