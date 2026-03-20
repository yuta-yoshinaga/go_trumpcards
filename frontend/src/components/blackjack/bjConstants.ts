/** Hi-Lo counting system constant. */
export const BJ_COUNTING_HILO = 0;
/** KO counting system constant. */
export const BJ_COUNTING_KO = 1;
/** Zen counting system constant. */
export const BJ_COUNTING_ZEN = 2;
/** Omega II counting system constant. */
export const BJ_COUNTING_OMEGA2 = 3;

/** Late surrender rule constant. */
export const BJ_SURRENDER_LATE = 0;
/** Early surrender rule constant. */
export const BJ_SURRENDER_EARLY = 1;
/** No surrender rule constant. */
export const BJ_SURRENDER_NONE = 2;

/** Valid deck penetration percentage options. */
export const BJ_VALID_PENETRATIONS = [50, 75] as const;

/** Perfect Pairs side bet type constant. */
export const BJ_SIDE_BET_PERFECT_PAIRS = 1;

/** Suggested action: none. */
export const BJ_SUGGEST_NONE = 0;
/** Suggested action: hit. */
export const BJ_SUGGEST_HIT = 1;
/** Suggested action: stand. */
export const BJ_SUGGEST_STAND = 2;
/** Suggested action: double down. */
export const BJ_SUGGEST_DOUBLE = 3;
/** Suggested action: split. */
export const BJ_SUGGEST_SPLIT = 4;
/** Suggested action: surrender. */
export const BJ_SUGGEST_SURRENDER = 5;
/** Suggested action: decline insurance. */
export const BJ_SUGGEST_DECLINE_INSURANCE = 6;
/** Suggested action: double if allowed, otherwise stand. */
export const BJ_SUGGEST_DOUBLE_STAND = 7;

/** Return a CSS class with highlight ring when the action is suggested. */
export function highlightClass(base: string, isHighlighted: boolean): string {
  return isHighlighted ? `${base} ring-2 ring-white ring-offset-1` : base;
}
