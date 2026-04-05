/** Adaptation level based on games played. */
export type AdaptationLevel = 'learning' | 'adapting' | 'adapted';

/** CPU strategy style derived from behavior rates. */
export type StrategyStyle = 'aggressive' | 'defensive' | 'balanced' | 'cautious' | 'observing';

/** Derive adaptation level from games played count. */
export function deriveAdaptationLevel(gamesPlayed: number): AdaptationLevel {
  if (gamesPlayed < 5) return 'learning';
  if (gamesPlayed < 15) return 'adapting';
  return 'adapted';
}

/** Derive strategy style from behavior rates. Works with any meta-AI type. */
export function deriveStrategyStyle(rates: {
  bluffRate?: number;
  foldRate?: number;
  edgePickRate?: number;
}): StrategyStyle {
  // For OldMaid (edgePickRate only)
  if (rates.edgePickRate !== undefined && rates.bluffRate === undefined) {
    return rates.edgePickRate > 0.5 ? 'cautious' : 'balanced';
  }
  const bluffRate = rates.bluffRate ?? 0;
  const foldRate = rates.foldRate ?? 0;
  if (bluffRate > 0.3) return 'aggressive';
  if (foldRate > 0.6) return 'defensive';
  if (bluffRate < 0.1 && foldRate < 0.3) return 'observing';
  return 'balanced';
}
