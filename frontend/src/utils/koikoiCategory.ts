import type { Card } from '../types/card';

/**
 * Koi-Koi (こいこい / hanafuda) yaku categories, mirroring the Go domain's
 * `KoiKoiCategory` (光 / 種 / 短冊 / カス). See `internal/domain/KoiKoi.go`.
 */
export type KoiKoiCategory = 'bright' | 'animal' | 'ribbon' | 'kasu';

/** Display order for the four categories: Bright → Animal → Ribbon → Chaff. */
export const KOIKOI_CATEGORY_ORDER: readonly KoiKoiCategory[] = ['bright', 'animal', 'ribbon', 'kasu'];

/**
 * Classifies a captured hanafuda card into its yaku category.
 *
 * The wire `design`/`value` fields cannot carry the month (the backend collapses
 * months 5–12 to `"JOKER"` via `cardDesignToString`), so the month is not
 * recoverable on the frontend. Instead we read the `color` ink token the backend
 * derives *directly* from `domain.KoiKoiCardCategory` in `KoiKoiWebPresenter.
 * koikoiFace`: gold→bright, purple→animal, red/blue→ribbon, black→chaff. This
 * keeps the frontend classification in lockstep with the domain SSoT rather than
 * re-inventing the 48-card table.
 */
export function koikoiCategory(card: Card): KoiKoiCategory {
  switch (card.color) {
    case 'gold':
      return 'bright';
    case 'purple':
      return 'animal';
    case 'red':
    case 'blue':
      return 'ribbon';
    default:
      return 'kasu';
  }
}

/**
 * Groups captured cards by {@link koikoiCategory}, preserving input order within
 * each group. Every category key is always present (possibly with an empty array).
 */
export function groupCapturedByCategory(cards: Card[]): Record<KoiKoiCategory, Card[]> {
  const groups: Record<KoiKoiCategory, Card[]> = { bright: [], animal: [], ribbon: [], kasu: [] };
  for (const card of cards) {
    groups[koikoiCategory(card)].push(card);
  }
  return groups;
}
