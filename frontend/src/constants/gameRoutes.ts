/** Route definitions for all game pages with navigation label i18n keys. */
export const gameRoutes = [
  { path: '/', labelKey: 'nav.blackjack' },
  { path: '/poker', labelKey: 'nav.poker' },
  { path: '/oldmaid', labelKey: 'nav.oldmaid' },
  { path: '/daifugo', labelKey: 'nav.daifugo' },
  { path: '/sevens', labelKey: 'nav.sevens' },
  { path: '/doubt', labelKey: 'nav.doubt' },
  { path: '/holdem', labelKey: 'nav.holdem' },
  { path: '/omaha', labelKey: 'nav.omaha' },
  { path: '/hearts', labelKey: 'nav.hearts' },
  { path: '/spades', labelKey: 'nav.spades' },
  { path: '/memory', labelKey: 'nav.memory' },
  { path: '/klondike', labelKey: 'nav.klondike' },
  { path: '/freecell', labelKey: 'nav.freecell' },
  { path: '/baccarat', labelKey: 'nav.baccarat' },
  { path: '/crazyeights', labelKey: 'nav.crazyeights' },
  { path: '/ginrummy', labelKey: 'nav.ginrummy' },
] as const;
