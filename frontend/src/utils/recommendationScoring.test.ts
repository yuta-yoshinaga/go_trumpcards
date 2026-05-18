import { describe, expect, it } from 'vitest';
import { AXIS_KEYS, AXIS_WEIGHTS, PROFILE_MAX, SOCIAL_PENALTY, SOCIAL_SOLO_IDX } from '../constants/discoverAxes';
import type { GameProfile, GameRoute } from '../constants/gameRoutes';
import { axisScore, dominantAxis, profileDistance, recommend, score, type UserMood } from './recommendationScoring';

/** Construct a GameRoute-shaped fixture for testing. */
function makeGame(path: string, profile: GameProfile, page = 'Test'): GameRoute {
  return { path, labelKey: `nav.${path.slice(1) || 'home'}`, icon: '🂠', page, profile };
}

/** Construct a UserMood, defaulting any unspecified axis to two skips. */
function makeMood(partial: Partial<UserMood> = {}): UserMood {
  return {
    mood: partial.mood ?? [null, null],
    skill: partial.skill ?? [null, null],
    social: partial.social ?? [null, null],
    theme: partial.theme ?? [null, null],
  };
}

describe('axisScore', () => {
  it('returns 0.5 when every answer is null (skip)', () => {
    expect(axisScore([5, 4, 3, 2], [null, null])).toBe(0.5);
  });

  it('averages normalized matches for valid answers', () => {
    // profile = [4, 3, 3, 5], answers = [0, 3] -> matches = [4/5, 5/5] = [0.8, 1.0]
    expect(axisScore([4, 3, 3, 5], [0, 3])).toBeCloseTo(0.9, 5);
  });

  it('clamps a match value above PROFILE_MAX at 1.0', () => {
    // Defensive: data integrity test elsewhere guards against >5, but the
    // function still clamps to keep recommendations stable.
    expect(axisScore([PROFILE_MAX + 2, 0, 0, 0], [0])).toBe(1);
  });

  it('returns 0.5 for an out-of-bounds index', () => {
    expect(axisScore([3, 3, 3, 3], [99])).toBe(0.5);
  });

  it('mixes valid and skip — skips are dropped, not coerced to zero', () => {
    expect(axisScore([5, 0, 0, 0], [0, null])).toBeCloseTo(1, 5);
  });
});

describe('score', () => {
  const blackJack = makeGame(
    '/',
    {
      mood: [4, 3, 3, 5],
      skill: [5, 5, 4, 3],
      social: [3, 5, 2],
      theme: [5, 1, 1, 1],
    },
    'BlackJack',
  );

  it('matches the design-doc worked example for the vs_cpu mood', () => {
    const mood = makeMood({ mood: [0, 3], skill: [0, null], social: [1, 1], theme: [0, 0] });
    // mood   = 0.90 * 0.35 = 0.315
    // skill  = 1.00 * 0.20 = 0.200
    // social = 1.00 * 0.30 = 0.300
    // theme  = 1.00 * 0.15 = 0.150
    // total  = 0.965; no penalty (social[0] = 1, not SOCIAL_SOLO_IDX)
    expect(score(blackJack, mood)).toBeCloseTo(0.965, 3);
  });

  it('applies solo penalty when user is solo and game social[SOCIAL_SOLO_IDX] < 2', () => {
    // Zero-profile game so the only non-skip contribution is from social.
    // game.social[0] = 0 (< 2) → penalty fires; everything else is skipped (0.5 neutral).
    const game = makeGame(
      '/g',
      { mood: [0, 0, 0, 0], skill: [0, 0, 0, 0], social: [0, 5, 5], theme: [0, 0, 0, 0] },
      'G',
    );
    const soloUser = makeMood({ social: [SOCIAL_SOLO_IDX, SOCIAL_SOLO_IDX] });

    // social   = 0.00 * 0.30 = 0.000
    // mood     = 0.50 * 0.35 = 0.175 (skipped)
    // skill    = 0.50 * 0.20 = 0.100 (skipped)
    // theme    = 0.50 * 0.15 = 0.075 (skipped)
    // raw      = 0.350
    // penalty  = SOCIAL_PENALTY * AXIS_WEIGHTS.social = 0.5 * 0.30 = 0.150
    // final    = 0.200
    expect(score(game, soloUser)).toBeCloseTo(0.35 - SOCIAL_PENALTY * AXIS_WEIGHTS.social, 5);
  });

  it('does not apply solo penalty when user did not answer solo first', () => {
    const game = makeGame(
      '/g',
      { mood: [0, 0, 0, 0], skill: [0, 0, 0, 0], social: [0, 5, 5], theme: [0, 0, 0, 0] },
      'G',
    );
    // First social answer is NOT solo → penalty does not fire even though game social[0] = 0.
    const cpuUser = makeMood({ social: [1, 1] });
    // social = 5/5,5/5 mean=1.0 -> 0.30; mood/skill/theme skipped -> 0.175 + 0.10 + 0.075 = 0.350
    // total = 0.650
    expect(score(game, cpuUser)).toBeCloseTo(0.65, 5);
  });

  it('does not penalize a solo user against a solo-friendly game', () => {
    const klondike = makeGame(
      '/klondike',
      { mood: [5, 1, 3, 3], skill: [4, 5, 3, 4], social: [5, 1, 0], theme: [3, 3, 3, 2] },
      'Klondike',
    );
    const soloUser = makeMood({ social: [SOCIAL_SOLO_IDX, null] });
    // social[SOCIAL_SOLO_IDX] = 5, not < 2 → penalty does NOT fire.
    // Verify by computing the raw weighted sum and confirming equality.
    let expected = 0;
    for (const k of AXIS_KEYS) expected += axisScore(klondike.profile[k], soloUser[k]) * AXIS_WEIGHTS[k];
    expect(score(klondike, soloUser)).toBeCloseTo(Math.max(0, Math.min(1, expected)), 5);
  });

  it('clamps result to the [0, 1] range', () => {
    const game = makeGame(
      '/edge',
      { mood: [0, 0, 0, 0], skill: [0, 0, 0, 0], social: [0, 0, 0], theme: [0, 0, 0, 0] },
      'Edge',
    );
    const result = score(game, makeMood({ social: [SOCIAL_SOLO_IDX, SOCIAL_SOLO_IDX] }));
    expect(result).toBeGreaterThanOrEqual(0);
    expect(result).toBeLessThanOrEqual(1);
  });
});

describe('dominantAxis', () => {
  it('returns the axis with the largest weighted contribution', () => {
    // High mood profile, neutral elsewhere; user matches mood strongly.
    const game = makeGame(
      '/x',
      { mood: [5, 0, 0, 0], skill: [3, 3, 3, 3], social: [3, 3, 3], theme: [3, 3, 3, 3] },
      'X',
    );
    expect(dominantAxis(game, makeMood({ mood: [0, 0] }))).toBe('mood');
  });

  it('returns "social" when only social is answered against a social-perfect game', () => {
    const game = makeGame(
      '/y',
      { mood: [0, 0, 0, 5], skill: [0, 0, 0, 5], social: [5, 5, 5], theme: [0, 0, 0, 5] },
      'Y',
    );
    // social: 1.00 * 0.30 = 0.300
    // mood/skill/theme: skipped (0.5) -> 0.175 / 0.100 / 0.075. social wins.
    expect(dominantAxis(game, makeMood({ social: [0, 0] }))).toBe('social');
  });

  it('returns "skill" when only skill is answered against a skill-perfect game', () => {
    const game = makeGame(
      '/z',
      { mood: [3, 3, 3, 3], skill: [5, 5, 5, 5], social: [3, 3, 3], theme: [3, 3, 3, 3] },
      'Z',
    );
    // skill: 1.00 * 0.20 = 0.200; mood (skipped): 0.5 * 0.35 = 0.175. skill wins by 0.025.
    expect(dominantAxis(game, makeMood({ skill: [0, 0] }))).toBe('skill');
  });
});

describe('profileDistance', () => {
  it('returns 0 for identical profiles', () => {
    const profile: GameProfile = {
      mood: [3, 3, 3, 3],
      skill: [3, 3, 3, 3],
      social: [3, 3, 3],
      theme: [3, 3, 3, 3],
    };
    const a = makeGame('/a', profile);
    const b = makeGame('/b', profile);
    expect(profileDistance(a, b)).toBeCloseTo(0, 5);
  });

  it('returns a positive value for differing profiles', () => {
    const a = makeGame('/a', {
      mood: [5, 0, 0, 0],
      skill: [5, 0, 0, 0],
      social: [5, 0, 0],
      theme: [5, 0, 0, 0],
    });
    const b = makeGame('/b', {
      mood: [0, 0, 0, 5],
      skill: [0, 0, 0, 5],
      social: [0, 0, 5],
      theme: [0, 0, 0, 5],
    });
    expect(profileDistance(a, b)).toBeGreaterThan(0);
  });
});

describe('recommend', () => {
  // 100 lookalike games + 1 outlier; index 50 is the outlier, designed to be
  // mid-band by score (score=0.5 — middle of the pack).
  function fillerGames(): GameRoute[] {
    const games: GameRoute[] = [];
    for (let i = 0; i < 60; i++) {
      games.push(
        makeGame(`/game${String(i).padStart(3, '0')}`, {
          mood: [3, 3, 3, 3],
          skill: [3, 3, 3, 3],
          social: [3, 3, 3],
          theme: [3, 3, 3, 3],
        }),
      );
    }
    return games;
  }

  it('returns top3 sorted by score then by path.localeCompare', () => {
    const a = makeGame('/za', {
      mood: [5, 5, 5, 5],
      skill: [5, 5, 5, 5],
      social: [5, 5, 5],
      theme: [5, 5, 5, 5],
    });
    const b = makeGame('/ab', {
      mood: [5, 5, 5, 5],
      skill: [5, 5, 5, 5],
      social: [5, 5, 5],
      theme: [5, 5, 5, 5],
    });
    const mood = makeMood({ mood: [0, 0], skill: [0, 0], social: [0, 0], theme: [0, 0] });
    const result = recommend([a, b], mood);
    // Both score equally; alphabetical path wins the tie-break.
    expect(result.top3[0].game.path).toBe('/ab');
    expect(result.top3[1].game.path).toBe('/za');
  });

  it('falls back to alphabetical when every answer is skip', () => {
    const games = [
      makeGame('/zebra', {
        mood: [3, 3, 3, 3],
        skill: [3, 3, 3, 3],
        social: [3, 3, 3],
        theme: [3, 3, 3, 3],
      }),
      makeGame('/apple', {
        mood: [3, 3, 3, 3],
        skill: [3, 3, 3, 3],
        social: [3, 3, 3],
        theme: [3, 3, 3, 3],
      }),
    ];
    const result = recommend(games, makeMood());
    expect(result.top3.map((g) => g.game.path)).toEqual(['/apple', '/zebra']);
  });

  it('picks a stretch from the mid-band that is furthest from top1', () => {
    const games = fillerGames();
    // Replace index 50 with an outlier profile that's still mid-band by score
    // (since user mood is neutral, all games score the same — mid-band = same band).
    const outlier = makeGame('/zzz-outlier', {
      mood: [0, 0, 0, 5],
      skill: [0, 0, 0, 5],
      social: [0, 0, 5],
      theme: [0, 0, 0, 5],
    });
    games[50] = outlier;
    const mood = makeMood();
    const result = recommend(games, mood);
    expect(result.stretch).not.toBeNull();
    expect(result.stretch?.game.path).toBe('/zzz-outlier');
  });

  it('returns stretch=null when the game list is too short for a mid-band', () => {
    const games = [
      makeGame('/a', {
        mood: [3, 3, 3, 3],
        skill: [3, 3, 3, 3],
        social: [3, 3, 3],
        theme: [3, 3, 3, 3],
      }),
      makeGame('/b', {
        mood: [3, 3, 3, 3],
        skill: [3, 3, 3, 3],
        social: [3, 3, 3],
        theme: [3, 3, 3, 3],
      }),
    ];
    const result = recommend(games, makeMood());
    expect(result.stretch).toBeNull();
  });

  it('also array excludes the stretch pick', () => {
    const games = fillerGames();
    games[50] = makeGame('/zzz-outlier', {
      mood: [0, 0, 0, 5],
      skill: [0, 0, 0, 5],
      social: [0, 0, 5],
      theme: [0, 0, 0, 5],
    });
    const result = recommend(games, makeMood());
    const alsoPaths = result.also.map((g) => g.game.path);
    expect(alsoPaths).not.toContain(result.stretch?.game.path);
  });
});
