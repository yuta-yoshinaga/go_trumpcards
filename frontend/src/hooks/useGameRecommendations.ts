import { useMemo } from 'react';
import { gameRoutes } from '../constants/gameRoutes';
import { type RecommendationResult, recommend, type UserMood } from '../utils/recommendationScoring';

/**
 * Memoize the result of `recommend()` against the full game registry.
 *
 * Recomputation is keyed on a JSON snapshot of `mood` so a parent that
 * builds a fresh `UserMood` literal each render still gets a stable
 * memoized result as long as the values are unchanged.
 */
export function useGameRecommendations(mood: UserMood): RecommendationResult {
  const moodKey = JSON.stringify(mood);
  // biome-ignore lint/correctness/useExhaustiveDependencies: moodKey snapshots mood — adding mood would invalidate on every render.
  return useMemo(() => recommend(gameRoutes, mood), [moodKey]);
}
