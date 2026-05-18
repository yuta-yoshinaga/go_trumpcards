/**
 * Pure recommendation scoring for the `/discover` concierge.
 *
 * Inputs: a list of game routes (each with a 4-axis profile) and a
 * `UserMood` (the survey answers). Output: ranked recommendations.
 *
 * All functions are pure — same inputs, same outputs, no side effects
 * outside an opt-in dev-mode `console.warn` for out-of-bounds inputs.
 */

import {
  AXIS_KEYS,
  AXIS_WEIGHTS,
  type AxisKey,
  PROFILE_MAX,
  SOCIAL_PENALTY,
  SOCIAL_SOLO_IDX,
} from '../constants/discoverAxes';
import type { GameRoute } from '../constants/gameRoutes';

/** A single survey answer per question — `null` denotes a skipped question. */
export type AxisAnswer = number | null;

/** The two answers a user gave for each axis (premise: 2 questions per axis). */
export interface UserMood {
  readonly mood: readonly [AxisAnswer, AxisAnswer];
  readonly skill: readonly [AxisAnswer, AxisAnswer];
  readonly social: readonly [AxisAnswer, AxisAnswer];
  readonly theme: readonly [AxisAnswer, AxisAnswer];
}

/** A scored game plus the axis that dominated the score. */
export interface ScoredGame {
  readonly game: GameRoute;
  /** Final score, clamped to [0, 1]. Higher = better match. */
  readonly score: number;
  /** Axis that contributed most to the score (UI label). */
  readonly topAxis: AxisKey;
}

/** The full recommendation result for one user mood. */
export interface RecommendationResult {
  readonly top3: readonly ScoredGame[];
  /** Mid-band pick that is profile-furthest from `top3[0]`, or null if the pool is too small. */
  readonly stretch: ScoredGame | null;
  /** Also-rans (ranks 4..10). The stretch pick is filtered out. */
  readonly also: readonly ScoredGame[];
}

/** Compute the 0..1 axis score for a profile vector against the user's answers. */
export function axisScore(profileVec: readonly number[], answers: readonly AxisAnswer[]): number {
  const valid = answers.filter((a): a is number => a !== null);
  if (valid.length === 0) return 0.5;
  let sum = 0;
  for (const idx of valid) {
    if (idx < 0 || idx >= profileVec.length) {
      if (import.meta.env.DEV) {
        console.warn(
          `[recommendationScoring] axisScore: idx ${idx} out of bounds for profile length ${profileVec.length}`,
        );
      }
      sum += 0.5;
      continue;
    }
    sum += Math.min(1, profileVec[idx] / PROFILE_MAX);
  }
  return sum / valid.length;
}

/** Compute the weighted match score for a single game against the user's mood. */
export function score(game: GameRoute, mood: UserMood): number {
  let total = 0;
  for (const key of AXIS_KEYS) {
    total += axisScore(game.profile[key], mood[key]) * AXIS_WEIGHTS[key];
  }
  // Solo-leaning user against a game that does not support solo well.
  // `mood.social[0]` is the first social answer; if the user picked solo there
  // and the game's solo affinity is weak (< 2), apply a fixed penalty.
  if (mood.social[0] === SOCIAL_SOLO_IDX && game.profile.social[SOCIAL_SOLO_IDX] < 2) {
    total -= SOCIAL_PENALTY * AXIS_WEIGHTS.social;
  }
  return Math.max(0, Math.min(1, total));
}

/** Return the axis with the largest weighted contribution for this game/mood. */
export function dominantAxis(game: GameRoute, mood: UserMood): AxisKey {
  let bestKey: AxisKey = AXIS_KEYS[0];
  let bestContribution = -Infinity;
  for (const key of AXIS_KEYS) {
    const contribution = axisScore(game.profile[key], mood[key]) * AXIS_WEIGHTS[key];
    if (contribution > bestContribution) {
      bestContribution = contribution;
      bestKey = key;
    }
  }
  return bestKey;
}

/** Sum of per-axis Euclidean distances between two profile vectors. */
export function profileDistance(a: GameRoute, b: GameRoute): number {
  let sum = 0;
  for (const key of AXIS_KEYS) {
    const va = a.profile[key];
    const vb = b.profile[key];
    let axisSum = 0;
    for (let i = 0; i < va.length; i++) {
      const diff = va[i] - vb[i];
      axisSum += diff * diff;
    }
    sum += Math.sqrt(axisSum);
  }
  return sum;
}

/**
 * Rank `games` against `mood` and return TOP3, a Stretch Pick, and Also-rans.
 *
 * Ties are broken by `game.path.localeCompare` for stable, testable output.
 * The Stretch Pick is drawn from the mid-band (sorted ranks 40..60) and is
 * the game whose profile vector is furthest from the top1 pick — designed to
 * surface "you'd never have picked this, but try it" cases.
 */
export function recommend(games: readonly GameRoute[], mood: UserMood): RecommendationResult {
  const scored: ScoredGame[] = games
    .map((game) => ({ game, score: score(game, mood), topAxis: dominantAxis(game, mood) }))
    .sort((a, b) => b.score - a.score || a.game.path.localeCompare(b.game.path));

  const top3 = scored.slice(0, 3);
  const midBand = scored.slice(40, 61);
  let stretch: ScoredGame | null = null;
  if (midBand.length > 0 && top3.length > 0) {
    const top1Game = top3[0].game;
    stretch = midBand.reduce<ScoredGame>(
      (best, current) =>
        profileDistance(current.game, top1Game) > profileDistance(best.game, top1Game) ? current : best,
      midBand[0],
    );
  }
  const also = scored.slice(3, 10).filter((g) => g.game.path !== stretch?.game.path);
  return { top3, stretch, also };
}
