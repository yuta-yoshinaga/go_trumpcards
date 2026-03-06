// Counting system constants (must match domain BJCounting constants)
export const BJ_COUNTING_HILO = 0;
export const BJ_COUNTING_KO = 1;
export const BJ_COUNTING_ZEN = 2;
export const BJ_COUNTING_OMEGA2 = 3;

// Side bet type constants (must match domain BJSideBet constants)
export const BJ_SIDE_BET_PERFECT_PAIRS = 1;

// suggestedAction constants (must match domain BJSuggestedAction)
export const BJ_SUGGEST_NONE = 0;
export const BJ_SUGGEST_HIT = 1;
export const BJ_SUGGEST_STAND = 2;
export const BJ_SUGGEST_DOUBLE = 3;
export const BJ_SUGGEST_SPLIT = 4;
export const BJ_SUGGEST_SURRENDER = 5;
export const BJ_SUGGEST_DECLINE_INSURANCE = 6;
export const BJ_SUGGEST_DOUBLE_STAND = 7;

export function highlightClass(base: string, isHighlighted: boolean): string {
  return isHighlighted ? `${base} ring-2 ring-white ring-offset-1` : base;
}
