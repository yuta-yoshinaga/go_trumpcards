/**
 * Background and footer theme classes for each game, organized by category.
 *
 * Every game listed in `gameRoutes.ts` (and the canonical Go registry) must have
 * an entry here. The strict `Record<GameKey, ...>` type makes a missing key a
 * compile-time error so new games cannot ship without a theme. Variants that
 * intentionally share a theme with another game declare that explicitly via
 * value reuse — e.g., `pineapple`/`crazypineapple` reuse the holdem palette but
 * keep their own keys so future divergence is a one-line change.
 */
export type GameKey =
  // Table games
  | 'blackjack'
  | 'spanish21'
  | 'baccarat'
  | 'threecard'
  | 'caribbeanstud'
  | 'oasispoker'
  | 'russianpoker'
  | 'texasholdembonus'
  | 'casinoholdem'
  | 'ultimatetexasholdem'
  | 'mississippistud'
  | 'highcardflush'
  | 'paigow'
  | 'chinesepoker'
  | 'letitride'
  | 'reddog'
  | 'casinowar'
  | 'dragontiger'
  | 'blackjackswitch'
  | 'fourcardpoker'
  // Poker
  | 'poker'
  | 'holdem'
  | 'omaha'
  | 'omahahilo'
  | 'bigo'
  | 'bigohilo'
  | 'shortdeck'
  | 'pineapple'
  | 'crazypineapple'
  | 'irishpoker'
  | 'sevencardstud'
  | 'razz'
  | 'badugi'
  | 'deucetoseven'
  | 'indianpoker'
  | 'videopoker'
  | 'deuceswild'
  | 'jokerpoker'
  // Trick-taking
  | 'hearts'
  | 'spades'
  | 'sheepshead'
  | 'doppelkopf'
  | 'mus'
  | 'tute'
  | 'sueca'
  | 'klaverjas'
  | 'manille'
  | 'marias'
  | 'sedma'
  | 'knockoutwhist'
  | 'spoilfive'
  | 'solowhist'
  | 'fortyfives'
  | 'nap'
  | 'preference'
  | 'twentynine'
  | 'courtpiece'
  | 'bezique'
  | 'ecarte'
  | 'threecardbrag'
  | 'teenpatti'
  | 'pitch'
  | 'twotenjack'
  | 'ohhell'
  | 'euchre'
  | 'bridge'
  | 'napoleon'
  | 'whist'
  | 'pinochle'
  | 'piquet'
  | 'callbreak'
  | 'tarneeb'
  | 'briscola'
  | 'schnapsen'
  | 'skat'
  | 'belote'
  | 'mighty'
  | 'fivehundred'
  // Matching/Pass
  | 'oldmaid'
  | 'doubt'
  | 'durak'
  | 'daifugo'
  | 'bigtwo'
  | 'tienlen'
  | 'president'
  | 'cassino'
  | 'scopa'
  | 'scopone'
  | 'escoba'
  | 'barbu'
  | 'macau'
  | 'mao'
  | 'sevens'
  | 'crazyeights'
  | 'pageone'
  | 'speed'
  | 'gofish'
  | 'pigtail'
  | 'war'
  | 'fiftyone'
  | 'trash'
  | 'sixcardgolf'
  | 'doudizhu'
  | 'truco'
  | 'spiteandmalice'
  | 'shithead'
  | 'nertz'
  | 'slapjack'
  | 'egyptianratscrew'
  // Solitaire
  | 'klondike'
  | 'freecell'
  | 'bakersgame'
  | 'eightoff'
  | 'penguin'
  | 'seahaventowers'
  | 'spider'
  | 'spiderette'
  | 'pyramid'
  | 'gaps'
  | 'tripeaks'
  | 'golf'
  | 'acesup'
  | 'memory'
  | 'clocksolitaire'
  | 'fortythieves'
  | 'bakersdozen'
  | 'beleagueredcastle'
  | 'canfield'
  | 'osmosis'
  | 'bristol'
  | 'bidwhist'
  | 'yukon'
  | 'russiansolitaire'
  | 'cruel'
  | 'scorpion'
  | 'wasp'
  | 'easthaven'
  | 'tichu'
  | 'bourre'
  | 'accordion'
  | 'pokersquares'
  | 'montecarlo'
  | 'calculation'
  | 'crescent'
  // Counting/Rummy
  | 'ginrummy'
  | 'conquian'
  | 'chinchon'
  | 'threethirteen'
  | 'tonk'
  | 'thirtyone'
  | 'yaniv'
  | 'gongzhu'
  | 'tressette'
  | 'canasta'
  | 'handandfoot'
  | 'burraco'
  | 'cribbage'
  | 'sevenbridge'
  | 'contractrummy'
  | 'kalooki'
  | 'rummy500';

/** Theme classes (Tailwind) applied to the page background and footer for each game. */
export interface GameThemeClasses {
  /** Background class for the outer game container. */
  bg: string;
  /** Background + border classes for the sticky footer. */
  footer: string;
}

const POKER = { bg: 'bg-game-bg-green-poker', footer: 'bg-game-bg-green-poker-dark border-white/20' } as const;
const CASINO = { bg: 'bg-game-bg-casino', footer: 'bg-game-bg-casino-dark border-white/20' } as const;
const BLUE = { bg: 'bg-game-bg-blue', footer: 'bg-game-bg-blue-dark border-white/20' } as const;
const GREEN = { bg: 'bg-game-bg-green', footer: 'bg-game-bg-green-dark border-white/20' } as const;
const BRIGHT_GREEN = {
  bg: 'bg-game-bg-green-bright',
  footer: 'bg-game-bg-green-bright-dark border-white/20',
} as const;
const SHEEPSHEAD = {
  bg: 'bg-game-bg-sheepshead',
  footer: 'bg-game-bg-sheepshead-dark border-white/20',
} as const;
const DOPPELKOPF = {
  bg: 'bg-game-bg-doppelkopf',
  footer: 'bg-game-bg-doppelkopf-dark border-white/20',
} as const;
const MUS = {
  bg: 'bg-game-bg-mus',
  footer: 'bg-game-bg-mus-dark border-white/20',
} as const;
const TUTE = {
  bg: 'bg-game-bg-tute',
  footer: 'bg-game-bg-tute-dark border-white/20',
} as const;
const SUECA = {
  bg: 'bg-game-bg-sueca',
  footer: 'bg-game-bg-sueca-dark border-white/20',
} as const;
const KLAVERJAS = {
  bg: 'bg-game-bg-klaverjas',
  footer: 'bg-game-bg-klaverjas-dark border-white/20',
} as const;
const MANILLE = {
  bg: 'bg-game-bg-manille',
  footer: 'bg-game-bg-manille-dark border-white/20',
} as const;
const MARIAS = {
  bg: 'bg-game-bg-marias',
  footer: 'bg-game-bg-marias-dark border-white/20',
} as const;
const SEDMA = {
  bg: 'bg-game-bg-sedma',
  footer: 'bg-game-bg-sedma-dark border-white/20',
} as const;
const KNOCKOUTWHIST = {
  bg: 'bg-game-bg-knockoutwhist',
  footer: 'bg-game-bg-knockoutwhist-dark border-white/20',
} as const;
const SPOILFIVE = {
  bg: 'bg-game-bg-spoilfive',
  footer: 'bg-game-bg-spoilfive-dark border-white/20',
} as const;
const SOLOWHIST = {
  bg: 'bg-game-bg-solowhist',
  footer: 'bg-game-bg-solowhist-dark border-white/20',
} as const;
const FORTYFIVES = {
  bg: 'bg-game-bg-fortyfives',
  footer: 'bg-game-bg-fortyfives-dark border-white/20',
} as const;
const NAP = {
  bg: 'bg-game-bg-nap',
  footer: 'bg-game-bg-nap-dark border-white/20',
} as const;
const TWENTYNINE = {
  bg: 'bg-game-bg-twentynine',
  footer: 'bg-game-bg-twentynine-dark border-white/20',
} as const;
const COURTPIECE = {
  bg: 'bg-game-bg-courtpiece',
  footer: 'bg-game-bg-courtpiece-dark border-white/20',
} as const;
const PREFERENCE = {
  bg: 'bg-game-bg-preference',
  footer: 'bg-game-bg-preference-dark border-white/20',
} as const;
const BEZIQUE = {
  bg: 'bg-game-bg-bezique',
  footer: 'bg-game-bg-bezique-dark border-white/20',
} as const;
const ECARTE = {
  bg: 'bg-game-bg-ecarte',
  footer: 'bg-game-bg-ecarte-dark border-white/20',
} as const;
const THREECARDBRAG = {
  bg: 'bg-game-bg-threecardbrag',
  footer: 'bg-game-bg-threecardbrag-dark border-white/20',
} as const;
const TEENPATTI = {
  bg: 'bg-game-bg-teenpatti',
  footer: 'bg-game-bg-teenpatti-dark border-white/20',
} as const;

export const gameTheme: Record<GameKey, GameThemeClasses> = {
  // Table games
  blackjack: BRIGHT_GREEN,
  spanish21: BRIGHT_GREEN,
  baccarat: CASINO,
  threecard: CASINO,
  caribbeanstud: CASINO,
  oasispoker: CASINO,
  russianpoker: CASINO,
  texasholdembonus: POKER,
  casinoholdem: POKER,
  ultimatetexasholdem: POKER,
  mississippistud: POKER,
  highcardflush: CASINO,
  paigow: CASINO,
  chinesepoker: CASINO,
  letitride: CASINO,
  reddog: CASINO,
  casinowar: CASINO,
  dragontiger: CASINO,
  blackjackswitch: BRIGHT_GREEN,
  fourcardpoker: CASINO,
  // Poker
  poker: POKER,
  holdem: POKER,
  omaha: POKER,
  omahahilo: POKER,
  bigo: POKER,
  bigohilo: POKER,
  shortdeck: POKER,
  pineapple: POKER,
  crazypineapple: POKER,
  irishpoker: POKER,
  sevencardstud: POKER,
  razz: POKER,
  badugi: POKER,
  deucetoseven: POKER,
  indianpoker: POKER,
  videopoker: CASINO,
  deuceswild: CASINO,
  jokerpoker: CASINO,
  // Trick-taking
  hearts: BLUE,
  spades: BLUE,
  sheepshead: SHEEPSHEAD,
  doppelkopf: DOPPELKOPF,
  mus: MUS,
  tute: TUTE,
  sueca: SUECA,
  klaverjas: KLAVERJAS,
  manille: MANILLE,
  marias: MARIAS,
  sedma: SEDMA,
  knockoutwhist: KNOCKOUTWHIST,
  spoilfive: SPOILFIVE,
  solowhist: SOLOWHIST,
  fortyfives: FORTYFIVES,
  nap: NAP,
  preference: PREFERENCE,
  twentynine: TWENTYNINE,
  courtpiece: COURTPIECE,
  bezique: BEZIQUE,
  ecarte: ECARTE,
  threecardbrag: THREECARDBRAG,
  teenpatti: TEENPATTI,
  pitch: BLUE,
  twotenjack: BLUE,
  ohhell: BLUE,
  euchre: BLUE,
  bridge: BLUE,
  napoleon: BLUE,
  whist: BLUE,
  pinochle: BLUE,
  piquet: BLUE,
  skat: BLUE,
  belote: BLUE,
  mighty: BLUE,
  fivehundred: BLUE,
  callbreak: BLUE,
  tarneeb: BLUE,
  briscola: BLUE,
  schnapsen: BLUE,
  // Matching/Pass
  oldmaid: GREEN,
  doubt: GREEN,
  durak: GREEN,
  daifugo: GREEN,
  bigtwo: GREEN,
  tienlen: GREEN,
  president: GREEN,
  cassino: GREEN,
  scopa: GREEN,
  scopone: GREEN,
  escoba: GREEN,
  barbu: GREEN,
  macau: GREEN,
  mao: BLUE,
  sevens: GREEN,
  crazyeights: GREEN,
  pageone: GREEN,
  speed: GREEN,
  gofish: GREEN,
  pigtail: GREEN,
  war: GREEN,
  fiftyone: GREEN,
  trash: GREEN,
  sixcardgolf: GREEN,
  doudizhu: CASINO,
  truco: CASINO,
  spiteandmalice: GREEN,
  shithead: GREEN,
  nertz: GREEN,
  slapjack: GREEN,
  egyptianratscrew: GREEN,
  // Solitaire
  klondike: CASINO,
  freecell: CASINO,
  bakersgame: CASINO,
  eightoff: CASINO,
  penguin: CASINO,
  seahaventowers: CASINO,
  spider: CASINO,
  spiderette: CASINO,
  pyramid: CASINO,
  gaps: CASINO,
  tripeaks: CASINO,
  golf: CASINO,
  acesup: GREEN,
  memory: CASINO,
  clocksolitaire: CASINO,
  fortythieves: CASINO,
  bakersdozen: CASINO,
  beleagueredcastle: CASINO,
  canfield: CASINO,
  osmosis: CASINO,
  bristol: CASINO,
  bidwhist: GREEN,
  yukon: CASINO,
  russiansolitaire: CASINO,
  cruel: CASINO,
  scorpion: CASINO,
  wasp: CASINO,
  easthaven: GREEN,
  tichu: CASINO,
  bourre: CASINO,
  accordion: CASINO,
  pokersquares: CASINO,
  montecarlo: CASINO,
  calculation: CASINO,
  crescent: CASINO,
  // Counting/Rummy
  ginrummy: BLUE,
  conquian: BLUE,
  chinchon: GREEN,
  threethirteen: BLUE,
  tonk: BLUE,
  thirtyone: CASINO,
  yaniv: BLUE,
  gongzhu: GREEN,
  tressette: GREEN,
  canasta: BLUE,
  handandfoot: BLUE,
  burraco: GREEN,
  cribbage: BLUE,
  sevenbridge: BLUE,
  contractrummy: BLUE,
  kalooki: GREEN,
  rummy500: BLUE,
};
