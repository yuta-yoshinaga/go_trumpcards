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
