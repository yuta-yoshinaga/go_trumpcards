import { describe, expect, it } from 'vitest';
import {
  AXIS_KEYS,
  AXIS_WEIGHTS,
  PROFILE_MAX,
  SOCIAL_PENALTY,
  SOCIAL_SOLO_IDX,
  SOCIAL_SOLO_PROFILE_IDX,
} from '../constants/discoverAxes';
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

/** Convenience: a "neutral 3" profile that fills both extended dimensions. */
const NEUTRAL_PROFILE: GameProfile = {
  mood: [3, 3, 3, 3],
  skill: [3, 3, 3, 3],
  social: [3, 3, 3, 3, 3],
  theme: [3, 3, 3, 3, 3, 3],
};

describe('axisScore', () => {
  it('returns 0.5 when every answer is null (skip)', () => {
    expect(axisScore([5, 4, 3, 2], 'mood', [null, null])).toBe(0.5);
  });

  it('Q1 + Q2 average resolves through option.profileIdx', () => {
    // mood Q1 option 0 = quiet_focus (profileIdx 0); Q2 option 1 = quick (profileIdx 3).
    // profile = [4, 3, 3, 5] → (4/5 + 5/5)/2 = 0.9
    expect(axisScore([4, 3, 3, 5], 'mood', [0, 1])).toBeCloseTo(0.9, 5);
  });

  it('inverts a polarity:-1 option (skill prefer_familiar)', () => {
    // skill Q2 option 1 = prefer_familiar (profileIdx 3, polarity -1).
    // profile[3] = 5 → inverted score = 1 - 1.0 = 0.0
    expect(axisScore([0, 0, 0, 5], 'skill', [null, 1])).toBe(0);
  });

  it('positive polarity on the same shared slot (skill learning_rules)', () => {
    // skill Q2 option 0 = learning_rules (profileIdx 3, positive). profile[3]=5 → 1.0
    expect(axisScore([0, 0, 0, 5], 'skill', [null, 0])).toBe(1);
  });

  it('clamps a match value above PROFILE_MAX at 1.0', () => {
    expect(axisScore([PROFILE_MAX + 2, 0, 0, 0], 'mood', [0, null])).toBeGreaterThanOrEqual(1);
  });

  it('returns 0.5 for an out-of-bounds option index', () => {
    expect(axisScore([3, 3, 3, 3], 'mood', [99, null])).toBe(0.5);
  });

  it('mixes valid and skip — skips are dropped, not coerced to zero', () => {
    expect(axisScore([5, 0, 0, 0], 'mood', [0, null])).toBeCloseTo(1, 5);
  });
});

describe('score', () => {
  const blackJack = makeGame(
    '/',
    {
      mood: [4, 3, 3, 5],
      skill: [5, 5, 4, 3],
      social: [3, 5, 2, 5, 5],
      theme: [5, 1, 1, 1, 5, 1],
    },
    'BlackJack',
  );

  it('computes the weighted match score across all four axes', () => {
    // Q1=quiet_focus(profileIdx 0)=4, Q2=quick(profileIdx 3)=5 → (4/5+5/5)/2 = 0.9
    // skill Q1=beginner(idx 0)=5, Q2=skipped → 5/5 = 1.0
    // social Q1=vs_cpu(idx 1)=5, Q2=serious_play(idx 4)=5 → (5/5+5/5)/2 = 1.0
    // theme  Q1=casino(idx 0)=5, Q2=showy(idx 4)=5 → (5/5+5/5)/2 = 1.0
    // total = 0.9*0.35 + 1.0*0.20 + 1.0*0.30 + 1.0*0.15 = 0.965
    const mood = makeMood({ mood: [0, 1], skill: [0, null], social: [1, 1], theme: [0, 0] });
    expect(score(blackJack, mood)).toBeCloseTo(0.965, 3);
  });

  it('applies solo penalty when user picks solo Q1 and game social[solo] < 2', () => {
    const game = makeGame(
      '/g',
      {
        mood: [0, 0, 0, 0],
        skill: [0, 0, 0, 0],
        // social[SOCIAL_SOLO_PROFILE_IDX]=0 (< 2) → penalty fires.
        social: [0, 5, 5, 5, 5],
        theme: [0, 0, 0, 0, 0, 0],
      },
      'G',
    );
    const soloUser = makeMood({ social: [SOCIAL_SOLO_IDX, null] });
    // social: profile[0]=0 → 0.0 * 0.30 = 0.000
    // mood/skill/theme skipped (0.5) → 0.175 + 0.100 + 0.075
    // raw = 0.350, penalty = 0.5 * 0.30 = 0.150, final = 0.200.
    expect(score(game, soloUser)).toBeCloseTo(0.35 - SOCIAL_PENALTY * AXIS_WEIGHTS.social, 5);
  });

  it('does not apply solo penalty when Q1 is not solo', () => {
    const game = makeGame(
      '/g',
      {
        mood: [0, 0, 0, 0],
        skill: [0, 0, 0, 0],
        social: [0, 5, 5, 5, 5],
        theme: [0, 0, 0, 0, 0, 0],
      },
      'G',
    );
    // Q1=1 (vs_cpu) → not solo → no penalty.
    const cpuUser = makeMood({ social: [1, null] });
    // social: profile[1]=5 → 1.0 * 0.30 = 0.30; others skipped = 0.175 + 0.100 + 0.075
    // total = 0.650
    expect(score(game, cpuUser)).toBeCloseTo(0.65, 5);
  });

  it('does not penalize a solo user against a solo-friendly game', () => {
    const klondike = makeGame(
      '/klondike',
      {
        mood: [5, 1, 3, 3],
        skill: [4, 5, 3, 4],
        social: [5, 1, 0, 4, 1],
        theme: [3, 3, 3, 2, 2, 4],
      },
      'Klondike',
    );
    const soloUser = makeMood({ social: [SOCIAL_SOLO_IDX, null] });
    // social[SOCIAL_SOLO_PROFILE_IDX] = 5, not < 2 → penalty does NOT fire.
    let expected = 0;
    for (const k of AXIS_KEYS) expected += axisScore(klondike.profile[k], k, soloUser[k]) * AXIS_WEIGHTS[k];
    expect(score(klondike, soloUser)).toBeCloseTo(Math.max(0, Math.min(1, expected)), 5);
  });

  it('clamps result to the [0, 1] range', () => {
    const game = makeGame(
      '/edge',
      {
        mood: [0, 0, 0, 0],
        skill: [0, 0, 0, 0],
        social: [0, 0, 0, 0, 0],
        theme: [0, 0, 0, 0, 0, 0],
      },
      'Edge',
    );
    const result = score(game, makeMood({ social: [SOCIAL_SOLO_IDX, null] }));
    expect(result).toBeGreaterThanOrEqual(0);
    expect(result).toBeLessThanOrEqual(1);
  });
});

describe('dominantAxis', () => {
  it('returns the axis with the largest weighted contribution', () => {
    // High mood profile, neutral elsewhere; user matches mood strongly.
    const game = makeGame(
      '/x',
      {
        mood: [5, 0, 0, 0],
        skill: [3, 3, 3, 3],
        social: [3, 3, 3, 3, 3],
        theme: [3, 3, 3, 3, 3, 3],
      },
      'X',
    );
    expect(dominantAxis(game, makeMood({ mood: [0, null] }))).toBe('mood');
  });

  it('returns "social" when only social is answered against a social-perfect game', () => {
    const game = makeGame(
      '/y',
      {
        mood: [0, 0, 0, 5],
        skill: [0, 0, 0, 5],
        social: [5, 5, 5, 5, 5],
        theme: [0, 0, 0, 5, 0, 0],
      },
      'Y',
    );
    expect(dominantAxis(game, makeMood({ social: [0, null] }))).toBe('social');
  });

  it('returns "skill" when only skill is answered against a skill-perfect game', () => {
    const game = makeGame(
      '/z',
      {
        mood: [3, 3, 3, 3],
        skill: [5, 5, 5, 5],
        social: [3, 3, 3, 3, 3],
        theme: [3, 3, 3, 3, 3, 3],
      },
      'Z',
    );
    expect(dominantAxis(game, makeMood({ skill: [0, null] }))).toBe('skill');
  });
});

describe('profileDistance', () => {
  it('returns 0 for identical profiles', () => {
    const a = makeGame('/a', NEUTRAL_PROFILE);
    const b = makeGame('/b', NEUTRAL_PROFILE);
    expect(profileDistance(a, b)).toBeCloseTo(0, 5);
  });

  it('returns a positive value for differing profiles', () => {
    const a = makeGame('/a', {
      mood: [5, 0, 0, 0],
      skill: [5, 0, 0, 0],
      social: [5, 0, 0, 5, 0],
      theme: [5, 0, 0, 0, 5, 0],
    });
    const b = makeGame('/b', {
      mood: [0, 0, 0, 5],
      skill: [0, 0, 0, 5],
      social: [0, 0, 5, 0, 5],
      theme: [0, 0, 0, 5, 0, 5],
    });
    expect(profileDistance(a, b)).toBeGreaterThan(0);
  });
});

describe('recommend', () => {
  function fillerGames(): GameRoute[] {
    const games: GameRoute[] = [];
    for (let i = 0; i < 60; i++) {
      games.push(makeGame(`/game${String(i).padStart(3, '0')}`, NEUTRAL_PROFILE));
    }
    return games;
  }

  it('returns top3 sorted by score then by path.localeCompare', () => {
    const perfect: GameProfile = {
      mood: [5, 5, 5, 5],
      skill: [5, 5, 5, 5],
      social: [5, 5, 5, 5, 5],
      theme: [5, 5, 5, 5, 5, 5],
    };
    const a = makeGame('/za', perfect);
    const b = makeGame('/ab', perfect);
    const mood = makeMood({ mood: [0, 0], skill: [0, 0], social: [0, 0], theme: [0, 0] });
    const result = recommend([a, b], mood);
    expect(result.top3[0].game.path).toBe('/ab');
    expect(result.top3[1].game.path).toBe('/za');
  });

  it('falls back to alphabetical when every answer is skip', () => {
    const games = [makeGame('/zebra', NEUTRAL_PROFILE), makeGame('/apple', NEUTRAL_PROFILE)];
    const result = recommend(games, makeMood());
    expect(result.top3.map((g) => g.game.path)).toEqual(['/apple', '/zebra']);
  });

  it('picks a stretch from the mid-band that is furthest from top1', () => {
    const games = fillerGames();
    games[30] = makeGame('/game030', {
      mood: [0, 0, 0, 5],
      skill: [0, 0, 0, 5],
      social: [0, 0, 5, 0, 5],
      theme: [0, 0, 0, 5, 0, 5],
    });
    const result = recommend(games, makeMood());
    expect(result.stretch).not.toBeNull();
    expect(result.stretch?.game.path).toBe('/game030');
  });

  it('returns stretch=null when the game list is too short for a mid-band', () => {
    const games = [makeGame('/a', NEUTRAL_PROFILE), makeGame('/b', NEUTRAL_PROFILE)];
    const result = recommend(games, makeMood());
    expect(result.stretch).toBeNull();
  });

  it('also-rans and the stretch pick come from disjoint rank bands', () => {
    const games = fillerGames();
    games[30] = makeGame('/game030', {
      mood: [0, 0, 0, 5],
      skill: [0, 0, 0, 5],
      social: [0, 0, 5, 0, 5],
      theme: [0, 0, 0, 5, 0, 5],
    });
    const result = recommend(games, makeMood());
    const alsoPaths = result.also.map((g) => g.game.path);
    expect(result.stretch).not.toBeNull();
    expect(alsoPaths).not.toContain(result.stretch?.game.path);
  });
});
