/** Hi-Lo counting system constant. Must match domain `BJCounting` value. */
export const BJ_COUNTING_HILO = 0;
/** KO counting system constant. Must match domain `BJCounting` value. */
export const BJ_COUNTING_KO = 1;
/** Zen counting system constant. Must match domain `BJCounting` value. */
export const BJ_COUNTING_ZEN = 2;
/** Omega II counting system constant. Must match domain `BJCounting` value. */
export const BJ_COUNTING_OMEGA2 = 3;

/** Late surrender rule constant. Must match domain `BJSurrender` value. */
export const BJ_SURRENDER_LATE = 0;
/** Early surrender rule constant. Must match domain `BJSurrender` value. */
export const BJ_SURRENDER_EARLY = 1;
/** No surrender rule constant. Must match domain `BJSurrender` value. */
export const BJ_SURRENDER_NONE = 2;

/** Valid deck penetration percentage options. Must match domain `BJPenetration` values. */
export const BJ_VALID_PENETRATIONS = [50, 75] as const;

/** Perfect Pairs side bet type constant. Must match domain `BJSideBet` value. */
export const BJ_SIDE_BET_PERFECT_PAIRS = 1;

/** Suggested action: none. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_NONE = 0;
/** Suggested action: hit. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_HIT = 1;
/** Suggested action: stand. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_STAND = 2;
/** Suggested action: double down. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_DOUBLE = 3;
/** Suggested action: split. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_SPLIT = 4;
/** Suggested action: surrender. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_SURRENDER = 5;
/** Suggested action: decline insurance. Must match domain `BJSuggestedAction` value. */
export const BJ_SUGGEST_DECLINE_INSURANCE = 6;
/**
 * Suggested action: "Double if allowed, otherwise Stand" — basic strategy
 * notation "Ds". Distinct from `BJ_SUGGEST_DOUBLE` because the backend emits
 * this when the optimal action is double-down but the rules at this turn might
 * disallow it (post-split, low chips). The UI displays it as plain "Double"
 * via `useSuggestionLabels` in `BlackJackPage.tsx`.
 *
 * Must match domain `BJSuggestedAction` value.
 */
export const BJ_SUGGEST_DOUBLE_STAND = 7;

/** Return a CSS class with highlight ring when the action is suggested. */
export function highlightClass(base: string, isHighlighted: boolean): string {
  return isHighlighted ? `${base} ring-2 ring-white ring-offset-1` : base;
}
