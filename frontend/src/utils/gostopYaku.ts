import type { GoStopBreakdown } from '../types/card';

/**
 * Maximum distance, in cards, from a scoring threshold at which a yaku is
 * previewed as "near-complete". A hand farther than this from every threshold
 * shows no preview, keeping the Go/Stop panel focused on realistic upgrades.
 */
export const GOSTOP_YAKU_PREVIEW_DISTANCE = 2;

/** Category of a Go-Stop yaku whose progress is unambiguous from a captured-card count. */
export type GoStopYakuCategory = 'gwang' | 'tti' | 'yeol' | 'pi';

/** A scoring combination (yaku) the player is a few cards away from completing. */
export interface GoStopNearYaku {
  /** Breakdown category the progress belongs to. */
  category: GoStopYakuCategory;
  /** i18n sub-key naming the yaku that completes at the next threshold. */
  target: string;
  /** Cards of this category captured so far. */
  current: number;
  /** Cards still needed to reach the next scoring threshold. */
  remaining: number;
}

// Bright (光/광) thresholds and the yaku each one completes. Mirrors the domain
// scoring in internal/domain/GoStop.go: 3 = 삼광, 4 = 사광, 5 = 오광.
const GWANG_THRESHOLDS: { count: number; target: string }[] = [
  { count: 3, target: 'samgwang' },
  { count: 4, target: 'sagwang' },
  { count: 5, target: 'ogwang' },
];

// Count-only threshold for the ribbon (띠), animal (열끗) and junk (피) categories.
// Named-ribbon sets (홍단/청단/초단) and 고도리 depend on specific months, which a
// plain count cannot determine, so they are intentionally excluded.
const TTI_THRESHOLD = 5;
const YEOL_THRESHOLD = 5;
const PI_THRESHOLD = 10;

/** Builds a near-yaku entry when the remaining distance is within the preview window. */
function nearEntry(
  category: GoStopYakuCategory,
  target: string,
  current: number,
  threshold: number,
): GoStopNearYaku | null {
  const remaining = threshold - current;
  if (remaining < 1 || remaining > GOSTOP_YAKU_PREVIEW_DISTANCE) return null;
  return { category, target, current, remaining };
}

/**
 * Computes the yaku a player is close to completing from a scoring breakdown,
 * using only the unambiguous captured-card counts (bright/ribbon/animal/junk).
 *
 * Thresholds match the Go-Stop domain scoring exactly; month-specific yaku
 * (고도리, 홍단, 청단, 초단) are omitted because a count cannot prove them.
 * Categories already at their maximum threshold produce no entry.
 */
export function computeNearYaku(bd: GoStopBreakdown | null): GoStopNearYaku[] {
  if (!bd) return [];
  const out: GoStopNearYaku[] = [];

  const nextGwang = GWANG_THRESHOLDS.find((th) => th.count > bd.brightCount);
  if (nextGwang) {
    const entry = nearEntry('gwang', nextGwang.target, bd.brightCount, nextGwang.count);
    if (entry) out.push(entry);
  }

  const tti = nearEntry('tti', 'tti', bd.ribbonCount, TTI_THRESHOLD);
  if (tti) out.push(tti);

  const yeol = nearEntry('yeol', 'yeol', bd.animalCount, YEOL_THRESHOLD);
  if (yeol) out.push(yeol);

  const pi = nearEntry('pi', 'pi', bd.piCount, PI_THRESHOLD);
  if (pi) out.push(pi);

  return out;
}
