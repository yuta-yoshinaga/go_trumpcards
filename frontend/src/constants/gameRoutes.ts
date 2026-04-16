/** A single game route with its path and navigation label i18n key. */
export interface GameRoute {
  path: string;
  labelKey: string;
  /** Emoji icon displayed next to the game name in navigation. */
  icon: string;
}

/** A category grouping related game routes with a category label i18n key. */
export interface GameCategory {
  labelKey: string;
  /** Emoji icon displayed next to the category name in navigation. */
  icon: string;
  routes: readonly GameRoute[];
}

/** Game routes organized by category for grouped navigation display. */
export const gameCategories: readonly GameCategory[] = [
  {
    labelKey: 'nav.category.table',
    icon: '🎰',
    routes: [
      { path: '/', labelKey: 'nav.blackjack', icon: '🃏' },
      { path: '/baccarat', labelKey: 'nav.baccarat', icon: '💎' },
      { path: '/threecard', labelKey: 'nav.threecard', icon: '🎴' },
      { path: '/caribbeanstud', labelKey: 'nav.caribbeanstud', icon: '🏝️' },
      { path: '/paigow', labelKey: 'nav.paigow', icon: '🀄' },
      { path: '/letitride', labelKey: 'nav.letitride', icon: '🎰' },
      { path: '/reddog', labelKey: 'nav.reddog', icon: '🐕' },
    ],
  },
  {
    labelKey: 'nav.category.poker',
    icon: '♠️',
    routes: [
      { path: '/poker', labelKey: 'nav.poker', icon: '🂡' },
      { path: '/holdem', labelKey: 'nav.holdem', icon: '🤠' },
      { path: '/omaha', labelKey: 'nav.omaha', icon: '4️⃣' },
      { path: '/shortdeck', labelKey: 'nav.shortdeck', icon: '6️⃣' },
      { path: '/pineapple', labelKey: 'nav.pineapple', icon: '🍍' },
      { path: '/sevencardstud', labelKey: 'nav.sevencardstud', icon: '7️⃣' },
      { path: '/indianpoker', labelKey: 'nav.indianpoker', icon: '🙈' },
      { path: '/videopoker', labelKey: 'nav.videopoker', icon: '📺' },
      { path: '/deuceswild', labelKey: 'nav.deuceswild', icon: '2️⃣' },
      { path: '/jokerpoker', labelKey: 'nav.jokerpoker', icon: '🤡' },
    ],
  },
  {
    labelKey: 'nav.category.trickTaking',
    icon: '🏆',
    routes: [
      { path: '/hearts', labelKey: 'nav.hearts', icon: '♥️' },
      { path: '/spades', labelKey: 'nav.spades', icon: '♠️' },
      { path: '/twotenjack', labelKey: 'nav.twotenjack', icon: '🎯' },
      { path: '/ohhell', labelKey: 'nav.ohhell', icon: '🔔' },
      { path: '/euchre', labelKey: 'nav.euchre', icon: '🎩' },
      { path: '/bridge', labelKey: 'nav.bridge', icon: '🌉' },
      { path: '/napoleon', labelKey: 'nav.napoleon', icon: '👑' },
      { path: '/whist', labelKey: 'nav.whist', icon: '🎴' },
    ],
  },
  {
    labelKey: 'nav.category.matching',
    icon: '🔄',
    routes: [
      { path: '/oldmaid', labelKey: 'nav.oldmaid', icon: '👵' },
      { path: '/doubt', labelKey: 'nav.doubt', icon: '🤥' },
      { path: '/durak', labelKey: 'nav.durak', icon: '🃏' },
      { path: '/daifugo', labelKey: 'nav.daifugo', icon: '💰' },
      { path: '/sevens', labelKey: 'nav.sevens', icon: '7️⃣' },
      { path: '/crazyeights', labelKey: 'nav.crazyeights', icon: '8️⃣' },
      { path: '/pageone', labelKey: 'nav.pageone', icon: '📄' },
      { path: '/speed', labelKey: 'nav.speed', icon: '⚡' },
      { path: '/gofish', labelKey: 'nav.gofish', icon: '🐟' },
      { path: '/pinochle', labelKey: 'nav.pinochle', icon: '🎯' },
      { path: '/pigtail', labelKey: 'nav.pigtail', icon: '🐷' },
      { path: '/war', labelKey: 'nav.war', icon: '⚔️' },
      { path: '/fiftyone', labelKey: 'nav.fiftyone', icon: '5️⃣' },
    ],
  },
  {
    labelKey: 'nav.category.solitaire',
    icon: '🏔️',
    routes: [
      { path: '/klondike', labelKey: 'nav.klondike', icon: '⛏️' },
      { path: '/freecell', labelKey: 'nav.freecell', icon: '🔲' },
      { path: '/spider', labelKey: 'nav.spider', icon: '🕷️' },
      { path: '/pyramid', labelKey: 'nav.pyramid', icon: '🔺' },
      { path: '/tripeaks', labelKey: 'nav.tripeaks', icon: '⛰️' },
      { path: '/golf', labelKey: 'nav.golf', icon: '⛳' },
      { path: '/memory', labelKey: 'nav.memory', icon: '🧠' },
      { path: '/clocksolitaire', labelKey: 'nav.clocksolitaire', icon: '🕐' },
      { path: '/fortythieves', labelKey: 'nav.fortythieves', icon: '🏰' },
      { path: '/canfield', labelKey: 'nav.canfield', icon: '🎩' },
      { path: '/yukon', labelKey: 'nav.yukon', icon: '🏔️' },
      { path: '/poker-squares', labelKey: 'nav.pokersquares', icon: '🔢' },
    ],
  },
  {
    labelKey: 'nav.category.rummy',
    icon: '🍸',
    routes: [
      { path: '/ginrummy', labelKey: 'nav.ginrummy', icon: '🫐' },
      { path: '/canasta', labelKey: 'nav.canasta', icon: '🃏' },
      { path: '/cribbage', labelKey: 'nav.cribbage', icon: '📌' },
    ],
  },
] as const;

/** Flat list of all game routes (derived from categories) for routing. */
export const gameRoutes: readonly GameRoute[] = gameCategories.flatMap((c) => c.routes);
