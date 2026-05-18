/**
 * AI Game Concierge — per-game profile vector.
 *
 * Each vector slot is aligned with the option order declared in
 * `discoverAxes.ts` (a wire-format break in `AXES.<axis>.options`
 * cascades through every profile here). Values are integers in
 * `0..PROFILE_MAX` (= 5), where 5 means "this option describes the
 * game perfectly" and 0 means "this option is a poor fit."
 */
export interface GameProfile {
  /** Aligned with `AXES.mood.options` (quiet_focus, lively, thoughtful, quick). */
  readonly mood: readonly [number, number, number, number];
  /** Aligned with `AXES.skill.options` (beginner, intermediate, advanced, learning_rules). */
  readonly skill: readonly [number, number, number, number];
  /** Aligned with `AXES.social.options` (solo, vs_cpu, with_friends). */
  readonly social: readonly [number, number, number];
  /** Aligned with `AXES.theme.options` (casino, european, western, japanese_household). */
  readonly theme: readonly [number, number, number, number];
}

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
  /** Per-axis profile vector used by `/discover` recommendation scoring. */
  profile: GameProfile;
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
      {
        path: '/',
        labelKey: 'nav.blackjack',
        icon: '🃏',
        page: 'BlackJack',
        profile: { mood: [3, 2, 3, 5], skill: [5, 5, 3, 4], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/spanish21',
        labelKey: 'nav.spanish21',
        icon: '🇪🇸',
        page: 'Spanish21',
        profile: { mood: [3, 2, 3, 4], skill: [3, 5, 4, 2], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/baccarat',
        labelKey: 'nav.baccarat',
        icon: '💎',
        page: 'Baccarat',
        profile: { mood: [2, 3, 1, 5], skill: [5, 3, 1, 5], social: [3, 5, 1], theme: [5, 1, 1, 1] },
      },
      {
        path: '/threecard',
        labelKey: 'nav.threecard',
        icon: '🎴',
        page: 'ThreeCard',
        profile: { mood: [3, 3, 2, 5], skill: [5, 5, 2, 4], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/caribbeanstud',
        labelKey: 'nav.caribbeanstud',
        icon: '🏝️',
        page: 'CaribbeanStud',
        profile: { mood: [2, 3, 3, 3], skill: [3, 5, 3, 3], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/oasispoker',
        labelKey: 'nav.oasispoker',
        icon: '🌴',
        page: 'OasisPoker',
        profile: { mood: [2, 3, 4, 3], skill: [3, 4, 4, 3], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/texasholdembonus',
        labelKey: 'nav.texasholdembonus',
        icon: '🤠',
        page: 'TexasHoldemBonus',
        profile: { mood: [2, 3, 4, 3], skill: [3, 4, 4, 2], social: [3, 5, 2], theme: [5, 2, 1, 1] },
      },
      {
        path: '/casinoholdem',
        labelKey: 'nav.casinoholdem',
        icon: '🎰',
        page: 'CasinoHoldem',
        profile: { mood: [2, 4, 3, 4], skill: [4, 5, 3, 3], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/ultimatetexasholdem',
        labelKey: 'nav.ultimatetexasholdem',
        icon: '♠️',
        page: 'UltimateTexasHoldem',
        profile: { mood: [2, 4, 5, 2], skill: [2, 4, 5, 2], social: [3, 5, 2], theme: [5, 2, 1, 1] },
      },
      {
        path: '/paigow',
        labelKey: 'nav.paigow',
        icon: '🀄',
        page: 'PaiGow',
        profile: { mood: [4, 2, 5, 1], skill: [2, 4, 5, 1], social: [3, 5, 2], theme: [4, 1, 1, 3] },
      },
      {
        path: '/letitride',
        labelKey: 'nav.letitride',
        icon: '🎰',
        page: 'LetItRide',
        profile: { mood: [2, 3, 2, 4], skill: [4, 4, 2, 3], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/reddog',
        labelKey: 'nav.reddog',
        icon: '🐕',
        page: 'RedDog',
        profile: { mood: [3, 3, 1, 5], skill: [5, 3, 1, 5], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/casinowar',
        labelKey: 'nav.casinowar',
        icon: '⚔️',
        page: 'CasinoWar',
        profile: { mood: [3, 4, 1, 5], skill: [5, 2, 1, 5], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/dragontiger',
        labelKey: 'nav.dragontiger',
        icon: '🐉',
        page: 'DragonTiger',
        profile: { mood: [3, 4, 1, 5], skill: [5, 2, 1, 5], social: [3, 5, 2], theme: [5, 0, 0, 2] },
      },
      {
        path: '/blackjackswitch',
        labelKey: 'nav.blackjackswitch',
        icon: '🔀',
        page: 'BlackJackSwitch',
        profile: { mood: [2, 3, 4, 3], skill: [2, 4, 5, 2], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/mississippistud',
        labelKey: 'nav.mississippistud',
        icon: '🚢',
        page: 'MississippiStud',
        profile: { mood: [3, 3, 4, 3], skill: [3, 5, 4, 2], social: [3, 5, 2], theme: [5, 1, 2, 1] },
      },
      {
        path: '/highcardflush',
        labelKey: 'nav.highcardflush',
        icon: '♣️',
        page: 'HighCardFlush',
        profile: { mood: [3, 3, 2, 4], skill: [5, 4, 2, 4], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
      {
        path: '/fourcardpoker',
        labelKey: 'nav.fourcardpoker',
        icon: '🃏',
        page: 'FourCardPoker',
        profile: { mood: [3, 3, 3, 3], skill: [4, 5, 3, 3], social: [3, 5, 2], theme: [5, 1, 1, 1] },
      },
    ],
  },
  {
    labelKey: 'nav.category.poker',
    icon: '♠️',
    routes: [
      {
        path: '/poker',
        labelKey: 'nav.poker',
        icon: '🂡',
        page: 'Poker',
        profile: { mood: [2, 3, 4, 3], skill: [3, 5, 4, 3], social: [1, 3, 5], theme: [3, 1, 3, 2] },
      },
      {
        path: '/holdem',
        labelKey: 'nav.holdem',
        icon: '🤠',
        page: 'Holdem',
        profile: { mood: [2, 4, 5, 2], skill: [2, 4, 5, 2], social: [1, 3, 5], theme: [3, 1, 3, 1] },
      },
      {
        path: '/omaha',
        labelKey: 'nav.omaha',
        icon: '4️⃣',
        page: 'Omaha',
        profile: { mood: [1, 3, 5, 2], skill: [1, 3, 5, 1], social: [1, 3, 5], theme: [3, 1, 3, 1] },
      },
      {
        path: '/omahahilo',
        labelKey: 'nav.omahahilo',
        icon: '½',
        page: 'OmahaHiLo',
        profile: { mood: [1, 2, 5, 1], skill: [0, 2, 5, 0], social: [1, 3, 5], theme: [3, 1, 3, 1] },
      },
      {
        path: '/shortdeck',
        labelKey: 'nav.shortdeck',
        icon: '6️⃣',
        page: 'ShortDeck',
        profile: { mood: [2, 5, 3, 4], skill: [2, 4, 4, 2], social: [1, 3, 5], theme: [3, 1, 3, 1] },
      },
      {
        path: '/pineapple',
        labelKey: 'nav.pineapple',
        icon: '🍍',
        page: 'Pineapple',
        profile: { mood: [2, 4, 4, 3], skill: [2, 4, 4, 2], social: [1, 3, 5], theme: [3, 1, 2, 1] },
      },
      {
        path: '/crazypineapple',
        labelKey: 'nav.crazypineapple',
        icon: '🤪',
        page: 'CrazyPineapple',
        profile: { mood: [1, 5, 3, 3], skill: [2, 4, 4, 2], social: [1, 3, 5], theme: [3, 1, 2, 1] },
      },
      {
        path: '/sevencardstud',
        labelKey: 'nav.sevencardstud',
        icon: '7️⃣',
        page: 'SevenCardStud',
        profile: { mood: [2, 3, 5, 2], skill: [2, 4, 5, 2], social: [1, 3, 5], theme: [2, 1, 3, 1] },
      },
      {
        path: '/razz',
        labelKey: 'nav.razz',
        icon: '🃏',
        page: 'Razz',
        profile: { mood: [2, 2, 5, 2], skill: [1, 2, 5, 1], social: [1, 3, 5], theme: [2, 1, 3, 1] },
      },
      {
        path: '/badugi',
        labelKey: 'nav.badugi',
        icon: '🪷',
        page: 'Badugi',
        profile: { mood: [2, 2, 5, 2], skill: [1, 3, 5, 1], social: [1, 3, 5], theme: [3, 1, 1, 1] },
      },
      {
        path: '/indianpoker',
        labelKey: 'nav.indianpoker',
        icon: '🙈',
        page: 'IndianPoker',
        profile: { mood: [1, 5, 3, 4], skill: [3, 3, 3, 4], social: [0, 2, 5], theme: [3, 1, 1, 2] },
      },
      {
        path: '/videopoker',
        labelKey: 'nav.videopoker',
        icon: '📺',
        page: 'VideoPoker',
        profile: { mood: [4, 2, 3, 5], skill: [4, 5, 4, 3], social: [5, 2, 0], theme: [5, 1, 1, 1] },
      },
      {
        path: '/deuceswild',
        labelKey: 'nav.deuceswild',
        icon: '2️⃣',
        page: 'DeucesWild',
        profile: { mood: [4, 2, 4, 4], skill: [3, 5, 4, 3], social: [5, 2, 0], theme: [5, 1, 1, 1] },
      },
      {
        path: '/jokerpoker',
        labelKey: 'nav.jokerpoker',
        icon: '🤡',
        page: 'JokerPoker',
        profile: { mood: [4, 2, 3, 5], skill: [4, 5, 3, 3], social: [5, 2, 0], theme: [5, 1, 1, 1] },
      },
    ],
  },
  {
    labelKey: 'nav.category.trickTaking',
    icon: '🏆',
    routes: [
      {
        path: '/hearts',
        labelKey: 'nav.hearts',
        icon: '♥️',
        page: 'Hearts',
        profile: { mood: [3, 3, 4, 2], skill: [3, 5, 4, 3], social: [1, 3, 5], theme: [2, 3, 3, 2] },
      },
      {
        path: '/spades',
        labelKey: 'nav.spades',
        icon: '♠️',
        page: 'Spades',
        profile: { mood: [3, 3, 4, 2], skill: [3, 5, 4, 3], social: [1, 3, 5], theme: [2, 2, 4, 1] },
      },
      {
        path: '/pitch',
        labelKey: 'nav.pitch',
        icon: '🎯',
        page: 'Pitch',
        profile: { mood: [3, 3, 4, 2], skill: [3, 4, 4, 2], social: [1, 3, 5], theme: [1, 1, 5, 1] },
      },
      {
        path: '/twotenjack',
        labelKey: 'nav.twotenjack',
        icon: '🎯',
        page: 'TwoTenJack',
        profile: { mood: [3, 4, 3, 3], skill: [4, 4, 3, 2], social: [1, 3, 5], theme: [1, 1, 1, 5] },
      },
      {
        path: '/ohhell',
        labelKey: 'nav.ohhell',
        icon: '🔔',
        page: 'OhHell',
        profile: { mood: [2, 4, 4, 3], skill: [3, 4, 4, 2], social: [1, 3, 5], theme: [2, 3, 3, 2] },
      },
      {
        path: '/euchre',
        labelKey: 'nav.euchre',
        icon: '🎩',
        page: 'Euchre',
        profile: { mood: [2, 4, 3, 3], skill: [3, 5, 4, 3], social: [1, 3, 5], theme: [2, 2, 4, 1] },
      },
      {
        path: '/bridge',
        labelKey: 'nav.bridge',
        icon: '🌉',
        page: 'Bridge',
        profile: { mood: [1, 2, 5, 0], skill: [0, 2, 5, 0], social: [1, 3, 5], theme: [3, 5, 2, 1] },
      },
      {
        path: '/napoleon',
        labelKey: 'nav.napoleon',
        icon: '👑',
        page: 'Napoleon',
        profile: { mood: [2, 3, 5, 2], skill: [2, 3, 5, 1], social: [1, 3, 5], theme: [3, 5, 1, 2] },
      },
      {
        path: '/whist',
        labelKey: 'nav.whist',
        icon: '🎴',
        page: 'Whist',
        profile: { mood: [3, 3, 4, 2], skill: [3, 4, 4, 2], social: [1, 3, 5], theme: [2, 5, 1, 1] },
      },
      {
        path: '/belote',
        labelKey: 'nav.belote',
        icon: '🇫🇷',
        page: 'Belote',
        profile: { mood: [2, 3, 4, 2], skill: [2, 4, 4, 2], social: [1, 3, 5], theme: [2, 5, 1, 1] },
      },
      {
        path: '/mighty',
        labelKey: 'nav.mighty',
        icon: '👊',
        page: 'Mighty',
        profile: { mood: [2, 4, 5, 2], skill: [2, 3, 5, 2], social: [1, 3, 5], theme: [2, 2, 1, 3] },
      },
      {
        path: '/piquet',
        labelKey: 'nav.piquet',
        icon: '🎴',
        page: 'Piquet',
        profile: { mood: [2, 2, 5, 2], skill: [1, 3, 5, 1], social: [1, 3, 5], theme: [2, 5, 1, 1] },
      },
      {
        path: '/callbreak',
        labelKey: 'nav.callbreak',
        icon: '🃏',
        page: 'CallBreak',
        profile: { mood: [2, 4, 4, 3], skill: [2, 4, 4, 2], social: [1, 3, 5], theme: [2, 2, 1, 3] },
      },
      {
        path: '/tarneeb',
        labelKey: 'nav.tarneeb',
        icon: '🎴',
        page: 'Tarneeb',
        profile: { mood: [2, 4, 4, 3], skill: [3, 4, 4, 2], social: [1, 3, 5], theme: [2, 3, 1, 3] },
      },
      {
        path: '/briscola',
        labelKey: 'nav.briscola',
        icon: '🇮🇹',
        page: 'Briscola',
        profile: { mood: [3, 4, 3, 4], skill: [4, 4, 3, 3], social: [1, 3, 5], theme: [2, 5, 1, 1] },
      },
    ],
  },
  {
    labelKey: 'nav.category.matching',
    icon: '🔄',
    routes: [
      {
        path: '/oldmaid',
        labelKey: 'nav.oldmaid',
        icon: '👵',
        page: 'OldMaid',
        profile: { mood: [3, 5, 1, 4], skill: [5, 2, 1, 5], social: [0, 2, 5], theme: [2, 2, 2, 5] },
      },
      {
        path: '/doubt',
        labelKey: 'nav.doubt',
        icon: '🤥',
        page: 'Doubt',
        profile: { mood: [2, 5, 2, 3], skill: [4, 4, 3, 4], social: [0, 2, 5], theme: [2, 2, 2, 5] },
      },
      {
        path: '/durak',
        labelKey: 'nav.durak',
        icon: '🃏',
        page: 'Durak',
        profile: { mood: [2, 4, 4, 3], skill: [2, 4, 4, 2], social: [1, 3, 5], theme: [2, 3, 1, 1] },
      },
      {
        path: '/daifugo',
        labelKey: 'nav.daifugo',
        icon: '💰',
        page: 'Daifugo',
        profile: { mood: [2, 5, 3, 3], skill: [3, 5, 4, 2], social: [0, 2, 5], theme: [2, 1, 1, 5] },
      },
      {
        path: '/president',
        labelKey: 'nav.president',
        icon: '🎩',
        page: 'President',
        profile: { mood: [2, 5, 3, 3], skill: [3, 5, 4, 2], social: [0, 2, 5], theme: [2, 2, 3, 3] },
      },
      {
        path: '/cassino',
        labelKey: 'nav.cassino',
        icon: '🎣',
        page: 'Cassino',
        profile: { mood: [3, 3, 4, 2], skill: [3, 4, 3, 3], social: [1, 3, 5], theme: [2, 4, 2, 1] },
      },
      {
        path: '/sevens',
        labelKey: 'nav.sevens',
        icon: '7️⃣',
        page: 'Sevens',
        profile: { mood: [3, 4, 2, 4], skill: [5, 4, 2, 4], social: [0, 2, 5], theme: [2, 2, 1, 4] },
      },
      {
        path: '/crazyeights',
        labelKey: 'nav.crazyeights',
        icon: '8️⃣',
        page: 'CrazyEights',
        profile: { mood: [3, 5, 2, 5], skill: [5, 3, 1, 5], social: [1, 3, 5], theme: [2, 2, 3, 3] },
      },
      {
        path: '/pageone',
        labelKey: 'nav.pageone',
        icon: '📄',
        page: 'PageOne',
        profile: { mood: [3, 5, 2, 5], skill: [5, 3, 1, 5], social: [1, 3, 5], theme: [2, 1, 1, 5] },
      },
      {
        path: '/speed',
        labelKey: 'nav.speed',
        icon: '⚡',
        page: 'Speed',
        profile: { mood: [3, 5, 1, 5], skill: [5, 3, 1, 5], social: [0, 2, 5], theme: [1, 1, 2, 3] },
      },
      {
        path: '/gofish',
        labelKey: 'nav.gofish',
        icon: '🐟',
        page: 'GoFish',
        profile: { mood: [3, 4, 2, 4], skill: [5, 2, 1, 5], social: [0, 2, 5], theme: [2, 2, 3, 3] },
      },
      {
        path: '/pinochle',
        labelKey: 'nav.pinochle',
        icon: '🎯',
        page: 'Pinochle',
        profile: { mood: [2, 3, 5, 2], skill: [2, 4, 5, 1], social: [1, 3, 5], theme: [2, 3, 3, 1] },
      },
      {
        path: '/pigtail',
        labelKey: 'nav.pigtail',
        icon: '🐷',
        page: 'PigsTail',
        profile: { mood: [3, 5, 1, 5], skill: [5, 3, 1, 5], social: [0, 2, 5], theme: [2, 3, 1, 4] },
      },
      {
        path: '/war',
        labelKey: 'nav.war',
        icon: '⚔️',
        page: 'War',
        profile: { mood: [3, 4, 1, 5], skill: [5, 1, 0, 5], social: [0, 3, 5], theme: [1, 1, 2, 3] },
      },
      {
        path: '/fiftyone',
        labelKey: 'nav.fiftyone',
        icon: '5️⃣',
        page: 'FiftyOne',
        profile: { mood: [3, 4, 2, 4], skill: [4, 3, 2, 4], social: [0, 2, 5], theme: [2, 2, 1, 4] },
      },
      {
        path: '/trash',
        labelKey: 'nav.trash',
        icon: '🗑️',
        page: 'Trash',
        profile: { mood: [3, 4, 2, 4], skill: [5, 2, 1, 5], social: [0, 2, 5], theme: [2, 2, 3, 3] },
      },
      {
        path: '/spiteandmalice',
        labelKey: 'nav.spiteandmalice',
        icon: '😈',
        page: 'SpiteAndMalice',
        profile: { mood: [2, 4, 4, 3], skill: [3, 4, 4, 2], social: [0, 5, 3], theme: [2, 2, 3, 2] },
      },
      {
        path: '/skat',
        labelKey: 'nav.skat',
        icon: '🇩🇪',
        page: 'Skat',
        profile: { mood: [1, 3, 5, 1], skill: [0, 3, 5, 0], social: [1, 3, 5], theme: [2, 5, 1, 1] },
      },
      {
        path: '/shithead',
        labelKey: 'nav.shithead',
        icon: '👑',
        page: 'Shithead',
        profile: { mood: [3, 5, 2, 3], skill: [4, 4, 3, 3], social: [0, 2, 5], theme: [2, 3, 2, 2] },
      },
      {
        path: '/nertz',
        labelKey: 'nav.nertz',
        icon: '🥜',
        page: 'Nertz',
        profile: { mood: [3, 5, 2, 4], skill: [3, 4, 3, 3], social: [0, 2, 5], theme: [2, 2, 3, 2] },
      },
      {
        path: '/slapjack',
        labelKey: 'nav.slapjack',
        icon: '✋',
        page: 'Slapjack',
        profile: { mood: [3, 5, 1, 5], skill: [5, 2, 1, 5], social: [0, 2, 5], theme: [2, 2, 3, 3] },
      },
      {
        path: '/egyptianratscrew',
        labelKey: 'nav.egyptianratscrew',
        icon: '🐀',
        page: 'EgyptianRatscrew',
        profile: { mood: [2, 5, 1, 5], skill: [4, 3, 2, 4], social: [0, 2, 5], theme: [2, 2, 3, 2] },
      },
    ],
  },
  {
    labelKey: 'nav.category.solitaire',
    icon: '🏔️',
    routes: [
      {
        path: '/klondike',
        labelKey: 'nav.klondike',
        icon: '⛏️',
        page: 'Klondike',
        profile: { mood: [5, 1, 3, 3], skill: [4, 5, 3, 4], social: [5, 1, 0], theme: [3, 3, 3, 2] },
      },
      {
        path: '/freecell',
        labelKey: 'nav.freecell',
        icon: '🔲',
        page: 'FreeCell',
        profile: { mood: [5, 1, 5, 2], skill: [3, 4, 5, 2], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/eightoff',
        labelKey: 'nav.eightoff',
        icon: '🎱',
        page: 'EightOff',
        profile: { mood: [5, 1, 5, 2], skill: [2, 4, 5, 1], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/seahaventowers',
        labelKey: 'nav.seahaventowers',
        icon: '🏖️',
        page: 'SeahavenTowers',
        profile: { mood: [5, 1, 5, 2], skill: [2, 3, 5, 1], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/spider',
        labelKey: 'nav.spider',
        icon: '🕷️',
        page: 'Spider',
        profile: { mood: [5, 1, 5, 1], skill: [2, 4, 5, 1], social: [5, 1, 0], theme: [3, 3, 3, 2] },
      },
      {
        path: '/spiderette',
        labelKey: 'nav.spiderette',
        icon: '🕸️',
        page: 'Spiderette',
        profile: { mood: [5, 1, 4, 3], skill: [4, 4, 3, 3], social: [5, 1, 0], theme: [3, 3, 3, 2] },
      },
      {
        path: '/pyramid',
        labelKey: 'nav.pyramid',
        icon: '🔺',
        page: 'Pyramid',
        profile: { mood: [5, 2, 3, 4], skill: [4, 4, 3, 4], social: [5, 1, 0], theme: [3, 2, 1, 2] },
      },
      {
        path: '/gaps',
        labelKey: 'nav.gaps',
        icon: '🧩',
        page: 'Gaps',
        profile: { mood: [5, 1, 4, 3], skill: [3, 4, 4, 2], social: [5, 1, 0], theme: [2, 3, 3, 1] },
      },
      {
        path: '/tripeaks',
        labelKey: 'nav.tripeaks',
        icon: '⛰️',
        page: 'TriPeaks',
        profile: { mood: [5, 2, 2, 5], skill: [5, 4, 2, 5], social: [5, 1, 0], theme: [3, 2, 2, 2] },
      },
      {
        path: '/golf',
        labelKey: 'nav.golf',
        icon: '⛳',
        page: 'Golf',
        profile: { mood: [5, 2, 2, 5], skill: [5, 3, 2, 5], social: [5, 1, 0], theme: [3, 2, 3, 2] },
      },
      {
        path: '/memory',
        labelKey: 'nav.memory',
        icon: '🧠',
        page: 'Memory',
        profile: { mood: [4, 3, 3, 4], skill: [5, 2, 1, 5], social: [5, 1, 3], theme: [2, 2, 2, 5] },
      },
      {
        path: '/clocksolitaire',
        labelKey: 'nav.clocksolitaire',
        icon: '🕐',
        page: 'ClockSolitaire',
        profile: { mood: [5, 1, 1, 5], skill: [5, 2, 0, 5], social: [5, 1, 0], theme: [3, 3, 2, 1] },
      },
      {
        path: '/fortythieves',
        labelKey: 'nav.fortythieves',
        icon: '🏰',
        page: 'FortyThieves',
        profile: { mood: [5, 1, 5, 1], skill: [1, 3, 5, 0], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/bakersdozen',
        labelKey: 'nav.bakersdozen',
        icon: '🥖',
        page: 'BakersDozen',
        profile: { mood: [5, 1, 5, 2], skill: [3, 4, 4, 2], social: [5, 1, 0], theme: [3, 3, 2, 1] },
      },
      {
        path: '/beleagueredcastle',
        labelKey: 'nav.beleagueredcastle',
        icon: '🏯',
        page: 'BeleagueredCastle',
        profile: { mood: [5, 1, 5, 2], skill: [3, 4, 4, 2], social: [5, 1, 0], theme: [3, 4, 2, 1] },
      },
      {
        path: '/canfield',
        labelKey: 'nav.canfield',
        icon: '🎩',
        page: 'Canfield',
        profile: { mood: [5, 1, 4, 3], skill: [3, 4, 4, 3], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/yukon',
        labelKey: 'nav.yukon',
        icon: '🏔️',
        page: 'Yukon',
        profile: { mood: [5, 1, 4, 2], skill: [3, 4, 4, 2], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/russiansolitaire',
        labelKey: 'nav.russiansolitaire',
        icon: '🪆',
        page: 'RussianSolitaire',
        profile: { mood: [5, 1, 5, 1], skill: [2, 4, 5, 1], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/cruel',
        labelKey: 'nav.cruel',
        icon: '🪓',
        page: 'Cruel',
        profile: { mood: [5, 1, 3, 3], skill: [4, 3, 3, 4], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/scorpion',
        labelKey: 'nav.scorpion',
        icon: '🦂',
        page: 'Scorpion',
        profile: { mood: [5, 1, 5, 2], skill: [2, 4, 4, 2], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/accordion',
        labelKey: 'nav.accordion',
        icon: '🪗',
        page: 'Accordion',
        profile: { mood: [5, 1, 4, 3], skill: [3, 4, 4, 3], social: [5, 1, 0], theme: [2, 3, 3, 1] },
      },
      {
        path: '/pokersquares',
        labelKey: 'nav.pokersquares',
        icon: '🔢',
        page: 'PokerSquares',
        profile: { mood: [5, 1, 5, 3], skill: [3, 4, 5, 2], social: [5, 1, 0], theme: [4, 2, 2, 1] },
      },
      {
        path: '/montecarlo',
        labelKey: 'nav.montecarlo',
        icon: '🎲',
        page: 'MonteCarlo',
        profile: { mood: [5, 2, 3, 4], skill: [4, 4, 2, 4], social: [5, 1, 0], theme: [3, 3, 2, 1] },
      },
      {
        path: '/calculation',
        labelKey: 'nav.calculation',
        icon: '🧮',
        page: 'Calculation',
        profile: { mood: [5, 1, 5, 2], skill: [2, 4, 5, 1], social: [5, 1, 0], theme: [3, 3, 3, 1] },
      },
      {
        path: '/crescent',
        labelKey: 'nav.crescent',
        icon: '🌙',
        page: 'Crescent',
        profile: { mood: [5, 1, 4, 3], skill: [3, 4, 4, 3], social: [5, 1, 0], theme: [3, 4, 2, 1] },
      },
    ],
  },
  {
    labelKey: 'nav.category.rummy',
    icon: '🍸',
    routes: [
      {
        path: '/ginrummy',
        labelKey: 'nav.ginrummy',
        icon: '🫐',
        page: 'GinRummy',
        profile: { mood: [3, 3, 4, 3], skill: [3, 5, 4, 3], social: [1, 3, 5], theme: [3, 2, 4, 2] },
      },
      {
        path: '/tonk',
        labelKey: 'nav.tonk',
        icon: '🎯',
        page: 'Tonk',
        profile: { mood: [3, 4, 3, 4], skill: [4, 4, 3, 3], social: [1, 3, 5], theme: [3, 1, 3, 2] },
      },
      {
        path: '/canasta',
        labelKey: 'nav.canasta',
        icon: '🃏',
        page: 'Canasta',
        profile: { mood: [2, 3, 5, 1], skill: [1, 3, 5, 1], social: [1, 3, 5], theme: [2, 3, 3, 1] },
      },
      {
        path: '/cribbage',
        labelKey: 'nav.cribbage',
        icon: '📌',
        page: 'Cribbage',
        profile: { mood: [3, 3, 4, 3], skill: [3, 5, 4, 2], social: [1, 3, 5], theme: [2, 5, 2, 1] },
      },
      {
        path: '/sevenbridge',
        labelKey: 'nav.sevenbridge',
        icon: '7️⃣',
        page: 'SevenBridge',
        profile: { mood: [3, 4, 3, 4], skill: [4, 4, 3, 3], social: [1, 3, 5], theme: [2, 1, 1, 5] },
      },
      {
        path: '/contractrummy',
        labelKey: 'nav.contractrummy',
        icon: '📜',
        page: 'ContractRummy',
        profile: { mood: [2, 3, 5, 2], skill: [2, 4, 5, 1], social: [1, 3, 5], theme: [2, 3, 3, 1] },
      },
      {
        path: '/rummy500',
        labelKey: 'nav.rummy500',
        icon: '🥃',
        page: 'Rummy500',
        profile: { mood: [3, 3, 4, 3], skill: [3, 4, 4, 2], social: [1, 3, 5], theme: [2, 2, 4, 2] },
      },
    ],
  },
] as const;

/** Flat list of all game routes (derived from categories) for routing. */
export const gameRoutes: readonly GameRoute[] = gameCategories.flatMap((c) => c.routes);
