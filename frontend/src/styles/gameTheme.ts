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
  | 'barbu'
  | 'macau'
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
  | 'accordion'
  | 'pokersquares'
  | 'montecarlo'
  | 'calculation'
  | 'crescent'
  // Counting/Rummy
  | 'ginrummy'
  | 'tonk'
  | 'thirtyone'
  | 'yaniv'
  | 'gongzhu'
  | 'tressette'
  | 'canasta'
  | 'burraco'
  | 'cribbage'
  | 'sevenbridge'
  | 'contractrummy'
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
  barbu: GREEN,
  macau: GREEN,
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
  accordion: CASINO,
  pokersquares: CASINO,
  montecarlo: CASINO,
  calculation: CASINO,
  crescent: CASINO,
  // Counting/Rummy
  ginrummy: BLUE,
  tonk: BLUE,
  thirtyone: CASINO,
  yaniv: BLUE,
  gongzhu: GREEN,
  tressette: GREEN,
  canasta: BLUE,
  burraco: GREEN,
  cribbage: BLUE,
  sevenbridge: BLUE,
  contractrummy: BLUE,
  rummy500: BLUE,
};
