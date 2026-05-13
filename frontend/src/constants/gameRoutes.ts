/** A single game route with its path and navigation label i18n key. */
export interface GameRoute {
  path: string;
  labelKey: string;
  /** Emoji icon displayed next to the game name in navigation. */
  icon: string;
  /**
   * PascalCase component name without the `Page` suffix. Used by `App.tsx` to
   * `import.meta.glob`-resolve the matching `pages/<page>Page.tsx` module.
   * Defaults can't be inferred from `path` alone because some routes don't
   * follow the simple "/foo → FooPage" convention (e.g., `/` → BlackJackPage,
   * `/pigtail` → PigsTailPage, `/spiteandmalice` → SpiteAndMalicePage).
   */
  page: string;
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
      { path: '/', labelKey: 'nav.blackjack', icon: '🃏', page: 'BlackJack' },
      { path: '/spanish21', labelKey: 'nav.spanish21', icon: '🇪🇸', page: 'Spanish21' },
      { path: '/baccarat', labelKey: 'nav.baccarat', icon: '💎', page: 'Baccarat' },
      { path: '/threecard', labelKey: 'nav.threecard', icon: '🎴', page: 'ThreeCard' },
      { path: '/caribbeanstud', labelKey: 'nav.caribbeanstud', icon: '🏝️', page: 'CaribbeanStud' },
      { path: '/oasispoker', labelKey: 'nav.oasispoker', icon: '🌴', page: 'OasisPoker' },
      { path: '/texasholdembonus', labelKey: 'nav.texasholdembonus', icon: '🤠', page: 'TexasHoldemBonus' },
      {
        path: '/ultimatetexasholdem',
        labelKey: 'nav.ultimatetexasholdem',
        icon: '♠️',
        page: 'UltimateTexasHoldem',
      },
      { path: '/paigow', labelKey: 'nav.paigow', icon: '🀄', page: 'PaiGow' },
      { path: '/letitride', labelKey: 'nav.letitride', icon: '🎰', page: 'LetItRide' },
      { path: '/reddog', labelKey: 'nav.reddog', icon: '🐕', page: 'RedDog' },
      { path: '/casinowar', labelKey: 'nav.casinowar', icon: '⚔️', page: 'CasinoWar' },
      { path: '/dragontiger', labelKey: 'nav.dragontiger', icon: '🐉', page: 'DragonTiger' },
      { path: '/blackjackswitch', labelKey: 'nav.blackjackswitch', icon: '🔀', page: 'BlackJackSwitch' },
      {
        path: '/mississippistud',
        labelKey: 'nav.mississippistud',
        icon: '🚢',
        page: 'MississippiStud',
      },
    ],
  },
  {
    labelKey: 'nav.category.poker',
    icon: '♠️',
    routes: [
      { path: '/poker', labelKey: 'nav.poker', icon: '🂡', page: 'Poker' },
      { path: '/holdem', labelKey: 'nav.holdem', icon: '🤠', page: 'Holdem' },
      { path: '/omaha', labelKey: 'nav.omaha', icon: '4️⃣', page: 'Omaha' },
      { path: '/omahahilo', labelKey: 'nav.omahahilo', icon: '½', page: 'OmahaHiLo' },
      { path: '/shortdeck', labelKey: 'nav.shortdeck', icon: '6️⃣', page: 'ShortDeck' },
      { path: '/pineapple', labelKey: 'nav.pineapple', icon: '🍍', page: 'Pineapple' },
      { path: '/crazypineapple', labelKey: 'nav.crazypineapple', icon: '🤪', page: 'CrazyPineapple' },
      { path: '/sevencardstud', labelKey: 'nav.sevencardstud', icon: '7️⃣', page: 'SevenCardStud' },
      { path: '/razz', labelKey: 'nav.razz', icon: '🃏', page: 'Razz' },
      { path: '/badugi', labelKey: 'nav.badugi', icon: '🪷', page: 'Badugi' },
      { path: '/indianpoker', labelKey: 'nav.indianpoker', icon: '🙈', page: 'IndianPoker' },
      { path: '/videopoker', labelKey: 'nav.videopoker', icon: '📺', page: 'VideoPoker' },
      { path: '/deuceswild', labelKey: 'nav.deuceswild', icon: '2️⃣', page: 'DeucesWild' },
      { path: '/jokerpoker', labelKey: 'nav.jokerpoker', icon: '🤡', page: 'JokerPoker' },
    ],
  },
  {
    labelKey: 'nav.category.trickTaking',
    icon: '🏆',
    routes: [
      { path: '/hearts', labelKey: 'nav.hearts', icon: '♥️', page: 'Hearts' },
      { path: '/spades', labelKey: 'nav.spades', icon: '♠️', page: 'Spades' },
      { path: '/pitch', labelKey: 'nav.pitch', icon: '🎯', page: 'Pitch' },
      { path: '/twotenjack', labelKey: 'nav.twotenjack', icon: '🎯', page: 'TwoTenJack' },
      { path: '/ohhell', labelKey: 'nav.ohhell', icon: '🔔', page: 'OhHell' },
      { path: '/euchre', labelKey: 'nav.euchre', icon: '🎩', page: 'Euchre' },
      { path: '/bridge', labelKey: 'nav.bridge', icon: '🌉', page: 'Bridge' },
      { path: '/napoleon', labelKey: 'nav.napoleon', icon: '👑', page: 'Napoleon' },
      { path: '/whist', labelKey: 'nav.whist', icon: '🎴', page: 'Whist' },
      { path: '/belote', labelKey: 'nav.belote', icon: '🇫🇷', page: 'Belote' },
      { path: '/mighty', labelKey: 'nav.mighty', icon: '👊', page: 'Mighty' },
    ],
  },
  {
    labelKey: 'nav.category.matching',
    icon: '🔄',
    routes: [
      { path: '/oldmaid', labelKey: 'nav.oldmaid', icon: '👵', page: 'OldMaid' },
      { path: '/doubt', labelKey: 'nav.doubt', icon: '🤥', page: 'Doubt' },
      { path: '/durak', labelKey: 'nav.durak', icon: '🃏', page: 'Durak' },
      { path: '/daifugo', labelKey: 'nav.daifugo', icon: '💰', page: 'Daifugo' },
      { path: '/president', labelKey: 'nav.president', icon: '🎩', page: 'President' },
      { path: '/cassino', labelKey: 'nav.cassino', icon: '🎣', page: 'Cassino' },
      { path: '/sevens', labelKey: 'nav.sevens', icon: '7️⃣', page: 'Sevens' },
      { path: '/crazyeights', labelKey: 'nav.crazyeights', icon: '8️⃣', page: 'CrazyEights' },
      { path: '/pageone', labelKey: 'nav.pageone', icon: '📄', page: 'PageOne' },
      { path: '/speed', labelKey: 'nav.speed', icon: '⚡', page: 'Speed' },
      { path: '/gofish', labelKey: 'nav.gofish', icon: '🐟', page: 'GoFish' },
      { path: '/pinochle', labelKey: 'nav.pinochle', icon: '🎯', page: 'Pinochle' },
      { path: '/pigtail', labelKey: 'nav.pigtail', icon: '🐷', page: 'PigsTail' },
      { path: '/war', labelKey: 'nav.war', icon: '⚔️', page: 'War' },
      { path: '/fiftyone', labelKey: 'nav.fiftyone', icon: '5️⃣', page: 'FiftyOne' },
      { path: '/trash', labelKey: 'nav.trash', icon: '🗑️', page: 'Trash' },
      { path: '/spiteandmalice', labelKey: 'nav.spiteandmalice', icon: '😈', page: 'SpiteAndMalice' },
      { path: '/skat', labelKey: 'nav.skat', icon: '🇩🇪', page: 'Skat' },
      { path: '/shithead', labelKey: 'nav.shithead', icon: '👑', page: 'Shithead' },
      { path: '/nertz', labelKey: 'nav.nertz', icon: '🥜', page: 'Nertz' },
      { path: '/slapjack', labelKey: 'nav.slapjack', icon: '✋', page: 'Slapjack' },
      { path: '/egyptianratscrew', labelKey: 'nav.egyptianratscrew', icon: '🐀', page: 'EgyptianRatscrew' },
    ],
  },
  {
    labelKey: 'nav.category.solitaire',
    icon: '🏔️',
    routes: [
      { path: '/klondike', labelKey: 'nav.klondike', icon: '⛏️', page: 'Klondike' },
      { path: '/freecell', labelKey: 'nav.freecell', icon: '🔲', page: 'FreeCell' },
      { path: '/spider', labelKey: 'nav.spider', icon: '🕷️', page: 'Spider' },
      { path: '/spiderette', labelKey: 'nav.spiderette', icon: '🕸️', page: 'Spiderette' },
      { path: '/pyramid', labelKey: 'nav.pyramid', icon: '🔺', page: 'Pyramid' },
      { path: '/tripeaks', labelKey: 'nav.tripeaks', icon: '⛰️', page: 'TriPeaks' },
      { path: '/golf', labelKey: 'nav.golf', icon: '⛳', page: 'Golf' },
      { path: '/memory', labelKey: 'nav.memory', icon: '🧠', page: 'Memory' },
      { path: '/clocksolitaire', labelKey: 'nav.clocksolitaire', icon: '🕐', page: 'ClockSolitaire' },
      { path: '/fortythieves', labelKey: 'nav.fortythieves', icon: '🏰', page: 'FortyThieves' },
      { path: '/bakersdozen', labelKey: 'nav.bakersdozen', icon: '🥖', page: 'BakersDozen' },
      { path: '/canfield', labelKey: 'nav.canfield', icon: '🎩', page: 'Canfield' },
      { path: '/yukon', labelKey: 'nav.yukon', icon: '🏔️', page: 'Yukon' },
      {
        path: '/russiansolitaire',
        labelKey: 'nav.russiansolitaire',
        icon: '🪆',
        page: 'RussianSolitaire',
      },
      { path: '/scorpion', labelKey: 'nav.scorpion', icon: '🦂', page: 'Scorpion' },
      { path: '/accordion', labelKey: 'nav.accordion', icon: '🪗', page: 'Accordion' },
      { path: '/pokersquares', labelKey: 'nav.pokersquares', icon: '🔢', page: 'PokerSquares' },
      { path: '/montecarlo', labelKey: 'nav.montecarlo', icon: '🎲', page: 'MonteCarlo' },
      { path: '/calculation', labelKey: 'nav.calculation', icon: '🧮', page: 'Calculation' },
      { path: '/crescent', labelKey: 'nav.crescent', icon: '🌙', page: 'Crescent' },
    ],
  },
  {
    labelKey: 'nav.category.rummy',
    icon: '🍸',
    routes: [
      { path: '/ginrummy', labelKey: 'nav.ginrummy', icon: '🫐', page: 'GinRummy' },
      { path: '/tonk', labelKey: 'nav.tonk', icon: '🎯', page: 'Tonk' },
      { path: '/canasta', labelKey: 'nav.canasta', icon: '🃏', page: 'Canasta' },
      { path: '/cribbage', labelKey: 'nav.cribbage', icon: '📌', page: 'Cribbage' },
      { path: '/sevenbridge', labelKey: 'nav.sevenbridge', icon: '7️⃣', page: 'SevenBridge' },
      {
        path: '/contractrummy',
        labelKey: 'nav.contractrummy',
        icon: '📜',
        page: 'ContractRummy',
      },
    ],
  },
] as const;

/** Flat list of all game routes (derived from categories) for routing. */
export const gameRoutes: readonly GameRoute[] = gameCategories.flatMap((c) => c.routes);
