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
  | 'texasholdembonus'
  | 'paigow'
  | 'letitride'
  | 'reddog'
  // Poker
  | 'poker'
  | 'holdem'
  | 'omaha'
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
  | 'twotenjack'
  | 'ohhell'
  | 'euchre'
  | 'bridge'
  | 'napoleon'
  | 'whist'
  | 'pinochle'
  | 'skat'
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
  | 'spider'
  | 'pyramid'
  | 'tripeaks'
  | 'golf'
  | 'memory'
  | 'clocksolitaire'
  | 'fortythieves'
  | 'bakersdozen'
  | 'canfield'
  | 'yukon'
  | 'scorpion'
  | 'accordion'
  | 'pokersquares'
  | 'calculation'
  // Counting/Rummy
  | 'ginrummy'
  | 'tonk'
  | 'canasta'
  | 'cribbage'
  | 'sevenbridge';

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
  texasholdembonus: POKER,
  paigow: CASINO,
  letitride: CASINO,
  reddog: CASINO,
  // Poker
  poker: POKER,
  holdem: POKER,
  omaha: POKER,
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
  twotenjack: BLUE,
  ohhell: BLUE,
  euchre: BLUE,
  bridge: BLUE,
  napoleon: BLUE,
  whist: BLUE,
  pinochle: BLUE,
  skat: BLUE,
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
  spider: CASINO,
  pyramid: CASINO,
  tripeaks: CASINO,
  golf: CASINO,
  memory: CASINO,
  clocksolitaire: CASINO,
  fortythieves: CASINO,
  bakersdozen: CASINO,
  canfield: CASINO,
  yukon: CASINO,
  scorpion: CASINO,
  accordion: CASINO,
  pokersquares: CASINO,
  calculation: CASINO,
  // Counting/Rummy
  ginrummy: BLUE,
  tonk: BLUE,
  canasta: BLUE,
  cribbage: BLUE,
  sevenbridge: BLUE,
};
