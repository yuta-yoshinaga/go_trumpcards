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
    '/baccaratbanque',
    '/bakersgame',
    '/baloot',
    '/barbu',
    '/bisley',
    '/braid',
    '/bridge',
    '/colorado',
    '/crescent',
    '/daifugo',
    '/diplomat',
    '/doubt',
    '/fortyandeight',
    '/fortythieves',
    '/freecell',
    '/goofspiel',
    '/gostop',
    '/grandfathersclock',
    '/kingalbert',
    '/klaberjass',
    '/klondike',
    '/koikoi',
    '/laughandliedown',
    '/loba',
    '/machiavelli',
    '/missmilligan',
    '/nainjaune',
    '/napoleon',
    '/napoleonssquare',
    '/omi',
    '/penguin',
    '/popejoan',
    '/pyramid',
    '/rankandfile',
    '/rikken',
    '/rummy500',
    '/saliclaw',
    '/scopone',
    '/shelem',
    '/slyfox',
    '/stalactites',
    '/sultan',
    '/tongits',
    '/trex',
    '/whitehead',
    '/windmill',
    '/yaniv',
    '/zwicker',
  ]),
  hand: new Set(['/goofspiel', '/scopone']),
  action: new Set([
    '/estimation',
    '/israeliwhist',
    '/minibridge',
    '/openfacechinese',
    '/polignac',
    '/saliclaw',
    '/sergeantmajor',
  ]),
};

/**
 * Games whose board cards already overflow sideways on develop, at the listed widths; only the
 * horizontal rule is skipped for them. Remove entries once #11811 is fixed. Do not add entries.
 */
export const KNOWN_HORIZONTAL_FAILURES: ReadonlyMap<string, ReadonlySet<number>> = new Map([
  ['/tripeaks', new Set([1024])],
]);

/** Minimum visible horizontal card width, passed into page.evaluate callbacks. */
export const VISIBLE_PX_X = 20;

/** Minimum visible vertical card height, passed into page.evaluate callbacks. */
export const VISIBLE_PX_Y = 52;
