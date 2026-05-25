/**
 * AI Game Concierge — axis definitions, weights, and constants.
 *
 * Single source of truth (SSoT) for the four mood axes used by the
 * `/discover` survey and the recommendation scoring function. All
 * downstream consumers (`recommendationScoring`, `urlMoodCodec`,
 * `useSurveyDraft`, `MoodQuestion`, `gameRoutes` profile typing) must
 * import the axis order, sub-question option keys, and weights from
 * here.
 *
 * Each axis is split into two sub-questions that probe **different**
 * dimensions of the axis — the first question is "broad slot" and the
 * second drills into an independent facet. Both sub-questions feed into
 * the same flat profile vector via `option.profileIdx`, so adding /
 * reordering options is a wire-format break.
 */

/** One selectable option within a sub-question. */
export interface SubQuestionOption {
  /** Stable kebab-case identifier (do not rename without migrating data). */
  readonly key: string;
  /**
   * i18n key resolved within the `discover` namespace — do NOT include
   * a leading `discover.` prefix, since callers already pass the
   * namespace via `useTranslation('discover')`.
   */
  readonly i18nKey: string;
  /** Index into the axis's flat profile vector that this option scores against. */
  readonly profileIdx: number;
  /**
   * Scoring polarity. Omitted (the common case) → score = `profile[idx]/MAX`.
   * `-1` → score = `1 - profile[idx]/MAX`, used when one sub-question turns
   * a single profile slot into a binary preference (e.g. `learning_rules`
   * vs `prefer_familiar`).
   */
  readonly polarity?: -1;
}

/** One of the two sub-questions on an axis. */
export interface SubQuestion {
  /**
   * i18n key for the question prompt, resolved within the `discover`
   * namespace (do NOT prefix with `discover.`).
   */
  readonly questionI18nKey: string;
  /** Ordered options; the option index is the wire-format integer. */
  readonly options: readonly SubQuestionOption[];
}

/** Definition of one mood axis: label + 2 sub-questions + profile vector length. */
export interface AxisDef {
  /**
   * i18n key for the axis label (e.g. shown in result `topAxis` chip).
   * Resolved within the `discover` namespace — do NOT include a leading
   * `discover.` prefix.
   */
  readonly labelI18nKey: string;
  /** Length of the flat profile vector that every game stores for this axis. */
  readonly profileLength: number;
  /** Premise: exactly 2 sub-questions per axis (question index 0 = Q1, 1 = Q2). */
  readonly questions: readonly [SubQuestion, SubQuestion];
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

/** Axis definitions. See module doc for ordering / wire-format rules. */
export const AXES: Readonly<Record<AxisKey, AxisDef>> = {
  mood: {
    labelI18nKey: 'axis.mood.label',
    profileLength: 4,
    questions: [
      {
        questionI18nKey: 'axis.mood.q1',
        options: [
          { key: 'quiet_focus', i18nKey: 'option.mood.quiet_focus', profileIdx: 0 },
          { key: 'lively', i18nKey: 'option.mood.lively', profileIdx: 1 },
        ],
      },
      {
        questionI18nKey: 'axis.mood.q2',
        options: [
          { key: 'thoughtful', i18nKey: 'option.mood.thoughtful', profileIdx: 2 },
          { key: 'quick', i18nKey: 'option.mood.quick', profileIdx: 3 },
        ],
      },
    ],
  },
  skill: {
    labelI18nKey: 'axis.skill.label',
    profileLength: 4,
    questions: [
      {
        questionI18nKey: 'axis.skill.q1',
        options: [
          { key: 'beginner', i18nKey: 'option.skill.beginner', profileIdx: 0 },
          { key: 'intermediate', i18nKey: 'option.skill.intermediate', profileIdx: 1 },
          { key: 'advanced', i18nKey: 'option.skill.advanced', profileIdx: 2 },
        ],
      },
      {
        questionI18nKey: 'axis.skill.q2',
        options: [
          { key: 'learning_rules', i18nKey: 'option.skill.learning_rules', profileIdx: 3 },
          // Inverts profile[3]: a "prefer familiar" user scores high on games
          // whose `learning_rules` value is low (i.e. they don't demand
          // novel-rule appetite). Single slot, two opposite answers.
          { key: 'prefer_familiar', i18nKey: 'option.skill.prefer_familiar', profileIdx: 3, polarity: -1 },
        ],
      },
    ],
  },
  social: {
    labelI18nKey: 'axis.social.label',
    profileLength: 5,
    questions: [
      {
        questionI18nKey: 'axis.social.q1',
        options: [
          { key: 'solo', i18nKey: 'option.social.solo', profileIdx: 0 },
          { key: 'vs_cpu', i18nKey: 'option.social.vs_cpu', profileIdx: 1 },
          { key: 'with_friends', i18nKey: 'option.social.with_friends', profileIdx: 2 },
        ],
      },
      {
        questionI18nKey: 'axis.social.q2',
        options: [
          { key: 'casual_play', i18nKey: 'option.social.casual_play', profileIdx: 3 },
          { key: 'serious_play', i18nKey: 'option.social.serious_play', profileIdx: 4 },
        ],
      },
    ],
  },
  theme: {
    labelI18nKey: 'axis.theme.label',
    profileLength: 6,
    questions: [
      {
        questionI18nKey: 'axis.theme.q1',
        options: [
          { key: 'casino', i18nKey: 'option.theme.casino', profileIdx: 0 },
          { key: 'european', i18nKey: 'option.theme.european', profileIdx: 1 },
          { key: 'western', i18nKey: 'option.theme.western', profileIdx: 2 },
          { key: 'japanese_household', i18nKey: 'option.theme.japanese_household', profileIdx: 3 },
        ],
      },
      {
        questionI18nKey: 'axis.theme.q2',
        options: [
          { key: 'showy', i18nKey: 'option.theme.showy', profileIdx: 4 },
          { key: 'subdued', i18nKey: 'option.theme.subdued', profileIdx: 5 },
        ],
      },
    ],
  },
} as const;

// `SOCIAL_SOLO_IDX` and `SOCIAL_SOLO_PROFILE_IDX` happen to both be 0 today
// because `solo` is the first option in social Q1 AND the first slot in the
// social profile vector — but they are distinct concepts: the first is the
// answer index the user submits, the second is the profile slot the scorer
// reads. Reorder the options and they diverge.

/** Option index of `solo` within social Q1 — compared against the user's submitted answer. */
const SOLO_OPT = AXES.social.questions[0].options.findIndex((o) => o.key === 'solo');
if (SOLO_OPT < 0) throw new Error('discoverAxes: social Q1 is missing the `solo` option');
export const SOCIAL_SOLO_IDX = SOLO_OPT;

/** Profile vector slot that stores a game's solo affinity — read by the solo penalty. */
const SOLO_PROFILE = AXES.social.questions[0].options[SOLO_OPT].profileIdx;
export const SOCIAL_SOLO_PROFILE_IDX = SOLO_PROFILE;

/**
 * Per-question option counts. Used to validate draft / URL input —
 * indexed as `[axis][questionIndex]`.
 */
export const AXIS_QUESTION_OPTION_COUNTS: Readonly<Record<AxisKey, readonly [number, number]>> = {
  mood: [AXES.mood.questions[0].options.length, AXES.mood.questions[1].options.length],
  skill: [AXES.skill.questions[0].options.length, AXES.skill.questions[1].options.length],
  social: [AXES.social.questions[0].options.length, AXES.social.questions[1].options.length],
  theme: [AXES.theme.questions[0].options.length, AXES.theme.questions[1].options.length],
} as const;
