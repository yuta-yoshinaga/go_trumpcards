const LARGE_DESKTOP_SCORE_TABLE_CLASS =
  'w-full table-fixed text-xs text-ds-text-muted [&_th]:px-1 [&_td]:px-1 [&_th]:tabular-nums [&_td]:tabular-nums [&_tbody_th:first-child]:truncate [&_tbody_td:first-child]:truncate [&_thead_th:first-child]:w-1/4 [&_tbody_th:first-child]:w-1/4 [&_tbody_td:first-child]:w-1/4';

/** Returns the score table classes for the current viewport. */
export function scoreTableClass(isLargeDesktop: boolean, mobileClass: string): string {
  if (!isLargeDesktop) return mobileClass;
  return `${LARGE_DESKTOP_SCORE_TABLE_CLASS}${mobileClass.endsWith(' mt-1') ? ' mt-1' : ''}`;
}
