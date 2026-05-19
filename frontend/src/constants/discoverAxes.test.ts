import { describe, expect, it } from 'vitest';
import {
  AXES,
  AXIS_KEYS,
  AXIS_OPTION_COUNT,
  AXIS_WEIGHTS,
  PROFILE_MAX,
  SOCIAL_PENALTY,
  SOCIAL_SOLO_IDX,
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

  it('TOTAL_QUESTIONS equals sum of questionCounts', () => {
    const total = AXIS_KEYS.reduce((acc, k) => acc + AXES[k].questionCount, 0);
    expect(TOTAL_QUESTIONS).toBe(total);
    expect(TOTAL_QUESTIONS).toBe(8);
  });

  it('PROFILE_MAX is 5', () => {
    expect(PROFILE_MAX).toBe(5);
  });

  it('SOCIAL_PENALTY is positive', () => {
    expect(SOCIAL_PENALTY).toBeGreaterThan(0);
  });

  it('SOCIAL_SOLO_IDX resolves to the solo option', () => {
    expect(SOCIAL_SOLO_IDX).toBe(0);
    expect(AXES.social.options[SOCIAL_SOLO_IDX].key).toBe('solo');
  });

  it('every axis has 2 questions', () => {
    for (const key of AXIS_KEYS) {
      expect(AXES[key].questionCount).toBe(2);
      expect(AXES[key].questionI18nKeys).toHaveLength(2);
    }
  });

  it('every axis has unique option keys', () => {
    for (const key of AXIS_KEYS) {
      const optKeys = AXES[key].options.map((o) => o.key);
      expect(new Set(optKeys).size).toBe(optKeys.length);
    }
  });

  it('every option has an option.<axis>.* i18n key (no leading `discover.` — the namespace is supplied by useTranslation)', () => {
    for (const key of AXIS_KEYS) {
      for (const opt of AXES[key].options) {
        expect(opt.i18nKey).toMatch(new RegExp('^option\\.' + key + '\\.'));
        expect(opt.i18nKey).not.toMatch(/^discover\./);
      }
    }
  });

  it('every axis label and question key omits the `discover.` namespace prefix', () => {
    for (const key of AXIS_KEYS) {
      expect(AXES[key].labelI18nKey).toMatch(new RegExp('^axis\\.' + key + '\\.label$'));
      expect(AXES[key].labelI18nKey).not.toMatch(/^discover\./);
      for (const qk of AXES[key].questionI18nKeys) {
        expect(qk).toMatch(new RegExp('^axis\\.' + key + '\\.q[12]$'));
      }
    }
  });

  it('AXIS_OPTION_COUNT matches AXES.options.length', () => {
    for (const key of AXIS_KEYS) {
      expect(AXIS_OPTION_COUNT[key]).toBe(AXES[key].options.length);
    }
  });

  it('mood/skill/theme have 4 options, social has 3', () => {
    expect(AXES.mood.options).toHaveLength(4);
    expect(AXES.skill.options).toHaveLength(4);
    expect(AXES.social.options).toHaveLength(3);
    expect(AXES.theme.options).toHaveLength(4);
  });
});
