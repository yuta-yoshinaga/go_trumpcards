/**
 * AI Game Concierge — axis definitions, weights, and constants.
 *
 * Single source of truth (SSoT) for the four mood axes used by the
 * `/discover` survey and the recommendation scoring function. All
 * downstream consumers (`recommendationScoring`, `urlMoodCodec`,
 * `useSurveyDraft`, `MoodQuestion`, `gameRoutes` profile typing) must
 * import the axis order, option keys, and weights from here.
 *
 * Changing the option order is a wire-format break: every game's
 * `profile` vector in `gameRoutes.ts` is indexed by these positions,
 * and the URL codec emits comma-separated integers tied to them.
 */

/** Option entry: a stable key plus its i18n key for the discover bundle. */
export interface AxisOption {
  /** Stable kebab-case identifier (do not rename without migrating data). */
  readonly key: string;
  /** i18n key in the lazy-loaded `discover` namespace. */
  readonly i18nKey: string;
}

/** Definition of one mood axis: 2 questions, ordered options. */
export interface AxisDef {
  /** Number of questions asked for this axis (premise: 2 per axis). */
  readonly questionCount: 2;
  /** i18n key for the axis label (e.g. shown in result `topAxis` chip). */
  readonly labelI18nKey: string;
  /** i18n keys for the 2 question prompts (order = question index 0, 1). */
  readonly questionI18nKeys: readonly [string, string];
  /** Ordered list of options; index is the wire format. */
  readonly options: readonly AxisOption[];
}

/** Discriminated key for the four mood axes. */
export type AxisKey = 'mood' | 'skill' | 'social' | 'theme';

/** Ordered list of axis keys used to drive survey progression and scoring. */
export const AXIS_KEYS = ['mood', 'skill', 'social', 'theme'] as const satisfies readonly AxisKey[];

/** Maximum value in a game's profile vector (profile values are 0..PROFILE_MAX). */
export const PROFILE_MAX = 5;

/** Total number of survey questions (4 axes × 2 questions). */
export const TOTAL_QUESTIONS = 8;

/** Weight each axis contributes to the final match score. Sum = 1.0. */
export const AXIS_WEIGHTS: Readonly<Record<AxisKey, number>> = {
  mood: 0.35,
  skill: 0.2,
  social: 0.3,
  theme: 0.15,
} as const;

/** Extra penalty applied when a solo-leaning user is matched against a multi-only game. */
export const SOCIAL_PENALTY = 0.5;

/** Axis definitions. See module doc for ordering rules. */
export const AXES: Readonly<Record<AxisKey, AxisDef>> = {
  mood: {
    questionCount: 2,
    labelI18nKey: 'discover.axis.mood.label',
    questionI18nKeys: ['discover.axis.mood.q1', 'discover.axis.mood.q2'],
    options: [
      { key: 'quiet_focus', i18nKey: 'discover.option.mood.quiet_focus' },
      { key: 'lively', i18nKey: 'discover.option.mood.lively' },
      { key: 'thoughtful', i18nKey: 'discover.option.mood.thoughtful' },
      { key: 'quick', i18nKey: 'discover.option.mood.quick' },
    ],
  },
  skill: {
    questionCount: 2,
    labelI18nKey: 'discover.axis.skill.label',
    questionI18nKeys: ['discover.axis.skill.q1', 'discover.axis.skill.q2'],
    options: [
      { key: 'beginner', i18nKey: 'discover.option.skill.beginner' },
      { key: 'intermediate', i18nKey: 'discover.option.skill.intermediate' },
      { key: 'advanced', i18nKey: 'discover.option.skill.advanced' },
      { key: 'learning_rules', i18nKey: 'discover.option.skill.learning_rules' },
    ],
  },
  social: {
    questionCount: 2,
    labelI18nKey: 'discover.axis.social.label',
    questionI18nKeys: ['discover.axis.social.q1', 'discover.axis.social.q2'],
    options: [
      { key: 'solo', i18nKey: 'discover.option.social.solo' },
      { key: 'vs_cpu', i18nKey: 'discover.option.social.vs_cpu' },
      { key: 'with_friends', i18nKey: 'discover.option.social.with_friends' },
    ],
  },
  theme: {
    questionCount: 2,
    labelI18nKey: 'discover.axis.theme.label',
    questionI18nKeys: ['discover.axis.theme.q1', 'discover.axis.theme.q2'],
    options: [
      { key: 'casino', i18nKey: 'discover.option.theme.casino' },
      { key: 'european', i18nKey: 'discover.option.theme.european' },
      { key: 'western', i18nKey: 'discover.option.theme.western' },
      { key: 'japanese_household', i18nKey: 'discover.option.theme.japanese_household' },
    ],
  },
} as const;

/** Index of the `solo` option inside `AXES.social.options` (used by social penalty). */
export const SOCIAL_SOLO_IDX = AXES.social.options.findIndex((o) => o.key === 'solo');

/** Number of options for each axis (used by profile vector length validation). */
export const AXIS_OPTION_COUNT: Readonly<Record<AxisKey, number>> = {
  mood: AXES.mood.options.length,
  skill: AXES.skill.options.length,
  social: AXES.social.options.length,
  theme: AXES.theme.options.length,
} as const;
