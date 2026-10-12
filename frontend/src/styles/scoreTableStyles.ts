const LARGE_DESKTOP_SCORE_TABLE_CLASS = [
  'w-full table-fixed text-xs text-ds-text-muted',
  '[&_th]:px-1 [&_td]:px-1 [&_th]:tabular-nums [&_td]:tabular-nums',
  '[&_tbody_th:first-child]:truncate [&_tbody_td:first-child]:truncate',
  '[&_thead_th:first-child]:w-1/4',
].join(' ');

/**
 * Returns score table classes for the current viewport.
 * `mobileClass` contains the base/mobile classes; pass optional spacing separately so it is
 * explicit and preserved in both viewport variants.
 */
export function scoreTableClass(isLargeDesktop: boolean, mobileClass: string, extraClass = ''): string {
  if (!isLargeDesktop) return [mobileClass, extraClass].filter(Boolean).join(' ');
  return [LARGE_DESKTOP_SCORE_TABLE_CLASS, extraClass].filter(Boolean).join(' ');
}
