/** 1024px is the LG_BREAKPOINT boundary where the two-column desktop layout begins. */
export const NARROW_DESKTOP_WIDTH = 1024;

export const DESKTOP_VIEWPORTS = [
  { width: 1280, height: 800 },
  { width: NARROW_DESKTOP_WIDTH, height: 800 },
] as const;

/**
 * Temporarily skip only vertical assertions for listed games at 1024px. Some failures are deal-dependent
 * and can pass on some runs (for example, /baloot), so test.fail() would mark those passing runs red.
 * The list therefore does not shrink automatically; remove entries manually once #11811 is fixed.
 */
export const KNOWN_1024_FAILURES: Record<'board' | 'hand' | 'action', ReadonlySet<string>> = {
  board: new Set([
    '/braid',
    '/colorado',
    '/crescent',
    '/grandfathersclock',
    '/kingalbert',
    '/pyramid',
    '/shelem',
    '/slyfox',
  ]),
  hand: new Set<string>(),
  action: new Set(['/rams']),
};

/**
 * Games whose board cards already overflow sideways on develop, at the listed widths; only the
 * horizontal rule is skipped for them. Remove entries once #11811 is fixed. Do not add entries.
 */
export const KNOWN_HORIZONTAL_FAILURES: ReadonlyMap<string, ReadonlySet<number>> = new Map([
  // The play area is wider than its column at both widths (scrollWidth 1140 > clientWidth 1040 at 1280).
  ['/cruel', new Set([1280, 1024])],
]);

/** Minimum visible horizontal card width, passed into page.evaluate callbacks. */
export const VISIBLE_PX_X = 20;

/** Minimum visible vertical card height, passed into page.evaluate callbacks. */
export const VISIBLE_PX_Y = 52;
