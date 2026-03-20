/** A single game route with its path and navigation label i18n key. */
export interface GameRoute {
  path: string;
  labelKey: string;
}

/** A category grouping related game routes with a category label i18n key. */
export interface GameCategory {
  labelKey: string;
  routes: readonly GameRoute[];
}

/** Game routes organized by category for grouped navigation display. */
export const gameCategories: readonly GameCategory[] = [
  {
    labelKey: 'nav.category.table',
    routes: [
      { path: '/', labelKey: 'nav.blackjack' },
      { path: '/baccarat', labelKey: 'nav.baccarat' },
    ],
  },
  {
    labelKey: 'nav.category.poker',
    routes: [
      { path: '/poker', labelKey: 'nav.poker' },
      { path: '/holdem', labelKey: 'nav.holdem' },
      { path: '/omaha', labelKey: 'nav.omaha' },
    ],
  },
  {
    labelKey: 'nav.category.trickTaking',
    routes: [
      { path: '/hearts', labelKey: 'nav.hearts' },
      { path: '/spades', labelKey: 'nav.spades' },
    ],
  },
  {
    labelKey: 'nav.category.matching',
    routes: [
      { path: '/oldmaid', labelKey: 'nav.oldmaid' },
      { path: '/doubt', labelKey: 'nav.doubt' },
      { path: '/daifugo', labelKey: 'nav.daifugo' },
      { path: '/sevens', labelKey: 'nav.sevens' },
      { path: '/crazyeights', labelKey: 'nav.crazyeights' },
    ],
  },
  {
    labelKey: 'nav.category.solitaire',
    routes: [
      { path: '/klondike', labelKey: 'nav.klondike' },
      { path: '/freecell', labelKey: 'nav.freecell' },
      { path: '/memory', labelKey: 'nav.memory' },
    ],
  },
  {
    labelKey: 'nav.category.rummy',
    routes: [{ path: '/ginrummy', labelKey: 'nav.ginrummy' }],
  },
] as const;

/** Flat list of all game routes (derived from categories) for routing. */
export const gameRoutes: readonly GameRoute[] = gameCategories.flatMap((c) => c.routes);
