import { describe, expect, it } from 'vitest';
import {
  AXES,
  AXIS_KEYS,
  AXIS_QUESTION_OPTION_COUNTS,
  AXIS_WEIGHTS,
  PROFILE_MAX,
  SOCIAL_PENALTY,
  SOCIAL_SOLO_IDX,
  SOCIAL_SOLO_PROFILE_IDX,
  TOTAL_QUESTIONS,
} from './discoverAxes';

describe('discoverAxes constants', () => {
  it('AXIS_KEYS lists all four axes in stable order', () => {
    expect(AXIS_KEYS).toEqual(['mood', 'skill', 'social', 'theme']);
  });

  it('AXIS_WEIGHTS sums to 1.0', () => {
    const sum = Object.values(AXIS_WEIGHTS).reduce((a, b) => a + b, 0);
    expect(sum).toBeCloseTo(1, 5);
  });

  it('TOTAL_QUESTIONS is 8 (4 axes × 2 sub-questions)', () => {
    expect(TOTAL_QUESTIONS).toBe(8);
    const total = AXIS_KEYS.reduce((acc, k) => acc + AXES[k].questions.length, 0);
    expect(total).toBe(TOTAL_QUESTIONS);
  });

  it('PROFILE_MAX is 5', () => {
    expect(PROFILE_MAX).toBe(5);
  });

  it('SOCIAL_PENALTY is positive', () => {
    expect(SOCIAL_PENALTY).toBeGreaterThan(0);
  });

  it('SOCIAL_SOLO_IDX resolves to the solo option in social Q1', () => {
    expect(SOCIAL_SOLO_IDX).toBe(0);
    expect(AXES.social.questions[0].options[SOCIAL_SOLO_IDX].key).toBe('solo');
  });

  it('SOCIAL_SOLO_PROFILE_IDX points at the solo profile slot', () => {
    expect(SOCIAL_SOLO_PROFILE_IDX).toBe(0);
  });

  it('every axis has exactly 2 sub-questions', () => {
    for (const key of AXIS_KEYS) {
      expect(AXES[key].questions).toHaveLength(2);
    }
  });

  it('every axis has unique option keys across both sub-questions', () => {
    for (const key of AXIS_KEYS) {
      const optKeys = AXES[key].questions.flatMap((q) => q.options.map((o) => o.key));
      expect(new Set(optKeys).size).toBe(optKeys.length);
    }
  });

  it('every option has an option.<axis>.* i18n key (no leading `discover.` — the namespace is supplied by useTranslation)', () => {
    for (const key of AXIS_KEYS) {
      for (const q of AXES[key].questions) {
        for (const opt of q.options) {
          expect(opt.i18nKey).toMatch(new RegExp(`^option\\.${key}\\.`));
          expect(opt.i18nKey).not.toMatch(/^discover\./);
        }
      }
    }
  });

  it('every axis label and question key omits the `discover.` namespace prefix', () => {
    for (const key of AXIS_KEYS) {
      expect(AXES[key].labelI18nKey).toMatch(new RegExp(`^axis\\.${key}\\.label$`));
      expect(AXES[key].labelI18nKey).not.toMatch(/^discover\./);
      for (const q of AXES[key].questions) {
        expect(q.questionI18nKey).toMatch(new RegExp(`^axis\\.${key}\\.q[12]$`));
      }
    }
  });

  it('AXIS_QUESTION_OPTION_COUNTS matches each sub-question option count', () => {
    for (const key of AXIS_KEYS) {
      const expected = AXES[key].questions.map((q) => q.options.length);
      expect(AXIS_QUESTION_OPTION_COUNTS[key]).toEqual(expected);
    }
  });

  it('every option.profileIdx is within axis.profileLength', () => {
    for (const key of AXIS_KEYS) {
      for (const q of AXES[key].questions) {
        for (const opt of q.options) {
          expect(opt.profileIdx).toBeGreaterThanOrEqual(0);
          expect(opt.profileIdx).toBeLessThan(AXES[key].profileLength);
        }
      }
    }
  });

  it('Q1 and Q2 profile-idx ranges do not overlap (except where polarity inverts a shared slot)', () => {
    for (const key of AXIS_KEYS) {
      const q1Idxs = new Set(AXES[key].questions[0].options.map((o) => o.profileIdx));
      for (const opt of AXES[key].questions[1].options) {
        if (q1Idxs.has(opt.profileIdx)) {
          // Shared slot is only legitimate when Q2 inverts via polarity.
          expect(opt.polarity).toBe(-1);
        }
      }
    }
  });
});
