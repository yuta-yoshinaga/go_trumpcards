import type { ScopaScoreDetail } from '../types/card';

/**
 * i18n suffix identifying a single-winner Scopa scoring category.
 * Mirrors the domain award categories: most cards (carte), most coins
 * (denari), most sevens (primiera-equivalent) and the seven of coins
 * (settebello).
 */
export type ScopaCategoryKey = 'cards' | 'denari' | 'primiera' | 'settebello';

/** One row of the Scopa round-end score breakdown for an award category. */
export interface ScopaBreakdownRow {
  /** i18n suffix key for the category label. */
  key: ScopaCategoryKey;
  /** Winning player index, or -1 when the category is tied / unawarded. */
  winner: number;
  /** Points awarded for this category (0 when tied / unawarded). */
  points: number;
}

/** Points granted per single-winner category (mirrors the domain constants, all 1). */
const CATEGORY_POINT = 1;

/**
 * Returns the sole player index holding the strict maximum count, or -1 when
 * the maximum is shared (tie) or every count is zero. Mirrors the backend
 * `uniqueMaxIndex` used by the Scopa scoring routine.
 */
function uniqueMaxIndex(counts: Record<number, number>): number {
  let best = -1;
  let bestVal = 0;
  let tie = false;
  for (const [k, v] of Object.entries(counts)) {
    const idx = Number(k);
    if (best === -1 || v > bestVal) {
      best = idx;
      bestVal = v;
      tie = false;
    } else if (v === bestVal) {
      tie = true;
    }
  }
  if (tie || bestVal === 0) {
    return -1;
  }
  return best;
}

/** Builds the winner/points pair for a category given its winning index. */
function award(winner: number): { winner: number; points: number } {
  return { winner, points: winner >= 0 ? CATEGORY_POINT : 0 };
}

/**
 * Derives the four single-winner Scopa award categories (carte, denari,
 * primiera, settebello) from a round-end score detail. Each row names the
 * winning player index (or -1 for a tie / unawarded category) and the points
 * granted. Scopa sweeps and per-player totals are rendered separately because
 * they are per-player counts rather than single-winner awards.
 */
export function scopaScoreBreakdown(detail: ScopaScoreDetail): ScopaBreakdownRow[] {
  return [
    { key: 'cards', ...award(uniqueMaxIndex(detail.cards)) },
    { key: 'denari', ...award(uniqueMaxIndex(detail.diamonds)) },
    { key: 'primiera', ...award(uniqueMaxIndex(detail.sevens)) },
    { key: 'settebello', ...award(detail.hasSetteBello) },
  ];
}
