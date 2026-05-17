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
  | 'texasholdembonus'
  | 'casinoholdem'
  | 'ultimatetexasholdem'
  | 'mississippistud'
  | 'highcardflush'
  | 'paigow'
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
  | 'shortdeck'
  | 'pineapple'
  | 'crazypineapple'
  | 'sevencardstud'
  | 'razz'
  | 'badugi'
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
  | 'skat'
  | 'belote'
  | 'mighty'
  // Matching/Pass
  | 'oldmaid'
  | 'doubt'
  | 'durak'
  | 'daifugo'
  | 'president'
  | 'cassino'
  | 'sevens'
  | 'crazyeights'
  | 'pageone'
  | 'speed'
  | 'gofish'
  | 'pigtail'
  | 'war'
  | 'fiftyone'
  | 'trash'
  | 'spiteandmalice'
  | 'shithead'
  | 'nertz'
  | 'slapjack'
  | 'egyptianratscrew'
  // Solitaire
  | 'klondike'
  | 'freecell'
  | 'seahaventowers'
  | 'spider'
  | 'spiderette'
  | 'pyramid'
  | 'gaps'
  | 'tripeaks'
  | 'golf'
  | 'memory'
  | 'clocksolitaire'
  | 'fortythieves'
  | 'bakersdozen'
  | 'beleagueredcastle'
  | 'canfield'
  | 'yukon'
  | 'russiansolitaire'
  | 'cruel'
  | 'scorpion'
  | 'accordion'
  | 'pokersquares'
  | 'montecarlo'
  | 'calculation'
  | 'crescent'
  // Counting/Rummy
  | 'ginrummy'
  | 'tonk'
  | 'canasta'
  | 'cribbage'
  | 'sevenbridge'
  | 'contractrummy';

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
  texasholdembonus: POKER,
  casinoholdem: POKER,
  ultimatetexasholdem: POKER,
  mississippistud: POKER,
  highcardflush: CASINO,
  paigow: CASINO,
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
  shortdeck: POKER,
  pineapple: POKER,
  crazypineapple: POKER,
  sevencardstud: POKER,
  razz: POKER,
  badugi: POKER,
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
  callbreak: BLUE,
  tarneeb: BLUE,
  briscola: BLUE,
  // Matching/Pass
  oldmaid: GREEN,
  doubt: GREEN,
  durak: GREEN,
  daifugo: GREEN,
  president: GREEN,
  cassino: GREEN,
  sevens: GREEN,
  crazyeights: GREEN,
  pageone: GREEN,
  speed: GREEN,
  gofish: GREEN,
  pigtail: GREEN,
  war: GREEN,
  fiftyone: GREEN,
  trash: GREEN,
  spiteandmalice: GREEN,
  shithead: GREEN,
  nertz: GREEN,
  slapjack: GREEN,
  egyptianratscrew: GREEN,
  // Solitaire
  klondike: CASINO,
  freecell: CASINO,
  seahaventowers: CASINO,
  spider: CASINO,
  spiderette: CASINO,
  pyramid: CASINO,
  gaps: CASINO,
  tripeaks: CASINO,
  golf: CASINO,
  memory: CASINO,
  clocksolitaire: CASINO,
  fortythieves: CASINO,
  bakersdozen: CASINO,
  beleagueredcastle: CASINO,
  canfield: CASINO,
  yukon: CASINO,
  russiansolitaire: CASINO,
  cruel: CASINO,
  scorpion: CASINO,
  accordion: CASINO,
  pokersquares: CASINO,
  montecarlo: CASINO,
  calculation: CASINO,
  crescent: CASINO,
  // Counting/Rummy
  ginrummy: BLUE,
  tonk: BLUE,
  canasta: BLUE,
  cribbage: BLUE,
  sevenbridge: BLUE,
  contractrummy: BLUE,
};
