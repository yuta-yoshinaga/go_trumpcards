/** 1024px is the LG_BREAKPOINT boundary where the two-column desktop layout begins. */
export const DESKTOP_VIEWPORTS = [
  { width: 1280, height: 800 },
  { width: 1024, height: 800 },
] as const;

/** Remove failures once #11811 is fixed. Do not add entries: fix newly failing games at 1024px. */
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
    '/rams',
    '/saliclaw',
    '/sergeantmajor',
  ]),
};

/**
 * Games whose board cards already overflow sideways on develop, at the listed widths; only the
 * horizontal rule is skipped for them. Remove entries once #11811 is fixed. Do not add entries.
 */
export const KNOWN_HORIZONTAL_FAILURES: ReadonlyMap<string, ReadonlySet<number>> = new Map([
  // The play area is wider than its column at both widths (scrollWidth 1140 > clientWidth 1040 at 1280).
  ['/cruel', new Set([1280, 1024])],
  ['/tripeaks', new Set([1024])],
]);

/** Minimum visible horizontal card width, passed into page.evaluate callbacks. */
export const VISIBLE_PX_X = 20;
