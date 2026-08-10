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
  | 'oichokabu'
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
  | 'fivecardstud'
  | 'soko'
  | 'razz'
  | 'sevencardstudhilo'
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
  | 'king'
  | 'tysiac'
  | 'calabresella'
  | 'ombre'
  | 'ulti'
  | 'scarto'
  | 'cego'
  | 'frenchtarot'
  | 'koenigrufen'
  | 'cinch'
  | 'loo'
  | 'basra'
  | 'hachihachi'
  | 'koikoi'
  | 'gostop'
  | 'tablanet'
  | 'trenteetquarante'
  | 'sedma'
  | 'knockoutwhist'
  | 'spoilfive'
  | 'solowhist'
  | 'fortyfives'
  | 'nap'
  | 'aluette'
  | 'minchiate'
  | 'tarocchini'
  | 'ganjifa'
  | 'preference'
  | 'vira'
  | 'twentynine'
  | 'courtpiece'
  | 'bezique'
  | 'ecarte'
  | 'threecardbrag'
  | 'teenpatti'
  | 'spoons'
  | 'kemps'
  | 'cuckoo'
  | 'pishti'
  | 'cuarenta'
  | 'faro'
  | 'openfacechinese'
  | 'russianbank'
  | 'pitch'
  | 'twotenjack'
  | 'ohhell'
  | 'wizard'
  | 'ninetynine'
  | 'euchre'
  | 'bridge'
  | 'napoleon'
  | 'whist'
  | 'catchten'
  | 'pinochle'
  | 'piquet'
  | 'callbreak'
  | 'tarneeb'
  | 'briscola'
  | 'schnapsen'
  | 'skat'
  | 'belote'
  | 'jass'
  | 'watten'
  | 'gaigel'
  | 'mighty'
  | 'fivehundred'
  | 'rook'
  // Matching/Pass
  | 'oldmaid'
  | 'doubt'
  | 'durak'
  | 'daifugo'
  | 'bigtwo'
  | 'tienlen'
  | 'zheng'
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
  | 'prsi'
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
  | 'streetsandalleys'
  | 'kingalbert'
  | 'flowergarden'
  | 'fortyandeight'
  | 'sultan'
  | 'agnes'
  | 'canfield'
  | 'osmosis'
  | 'bristol'
  | 'labellelucie'
  | 'simplesimon'
  | 'doubleklondike'
  | 'blackhole'
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
  | 'sirtommy'
  | 'fourseasons'
  | 'colorado'
  | 'auldlangsyne'
  | 'bisley'
  | 'napoleonssquare'
  | 'grandfathersclock'
  | 'duchess'
  | 'windmill'
  | 'americantoad'
  | 'congress'
  | 'terrace'
  | 'braid'
  | 'pontoon'
  | 'settemezzo'
  | 'niuniu'
  | 'bura'
  | 'mushi'
  | 'toepen'
  | 'chineseten'
  | 'skitgubbe'
  | 'laughandliedown'
  | 'sjavs'
  | 'trex'
  | 'loba'
  | 'desmoche'
  | 'zwicker'
  | 'poch'
  | 'popejoan'
  | 'nainjaune'
  | 'kille'
  | 'klaberjass'
  | 'kaiser'
  | 'boston'
  | 'vint'
  | 'bideuchre'
  | 'sixbidsolo'
  | 'karnoffel'
  | 'literature'
  | 'guandan'
  | 'shengji'
  | 'missmilligan'
  | 'crescent'
  // Counting/Rummy
  | 'ginrummy'
  | 'indianrummy'
  | 'machiavelli'
  | 'conquian'
  | 'chinchon'
  | 'threethirteen'
  | 'tonk'
  | 'thirtyone'
  | 'yaniv'
  | 'gongzhu'
  | 'tressette'
  | 'canasta'
  | 'samba'
  | 'handandfoot'
  | 'burraco'
  | 'cribbage'
  | 'sevenbridge'
  | 'contractrummy'
  | 'carioca'
  | 'kalooki'
  | 'rummy500'
  | 'beggarmyneighbour'
  | 'allfours'
  | 'guts'
  | 'anaconda'
  | 'bouillotte'
  | 'primero'
  | 'michigan'
  | 'pan';

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
const KING = {
  bg: 'bg-game-bg-king',
  footer: 'bg-game-bg-king-dark border-white/20',
} as const;
const TYSIAC = {
  bg: 'bg-game-bg-tysiac',
  footer: 'bg-game-bg-tysiac-dark border-white/20',
} as const;
const CALABRESELLA = {
  bg: 'bg-game-bg-calabresella',
  footer: 'bg-game-bg-calabresella-dark border-white/20',
} as const;
const OMBRE = {
  bg: 'bg-game-bg-ombre',
  footer: 'bg-game-bg-ombre-dark border-white/20',
} as const;
const ULTI = {
  bg: 'bg-game-bg-ulti',
  footer: 'bg-game-bg-ulti-dark border-white/20',
} as const;
const WATTEN = {
  bg: 'bg-game-bg-watten',
  footer: 'bg-game-bg-watten-dark border-white/20',
} as const;
const CINCH = {
  bg: 'bg-game-bg-cinch',
  footer: 'bg-game-bg-cinch-dark border-white/20',
} as const;
const LOO = {
  bg: 'bg-game-bg-loo',
  footer: 'bg-game-bg-loo-dark border-white/20',
} as const;
const BASRA = {
  bg: 'bg-game-bg-basra',
  footer: 'bg-game-bg-basra-dark border-white/20',
} as const;
const TABLANET = {
  bg: 'bg-game-bg-tablanet',
  footer: 'bg-game-bg-tablanet-dark border-white/20',
} as const;
const TRENTEETQUARANTE = {
  bg: 'bg-game-bg-trenteetquarante',
  footer: 'bg-game-bg-trenteetquarante-dark border-white/20',
} as const;
const SEDMA = {
  bg: 'bg-game-bg-sedma',
  footer: 'bg-game-bg-sedma-dark border-white/20',
} as const;
const PRSI = {
  bg: 'bg-game-bg-prsi',
  footer: 'bg-game-bg-prsi-dark border-white/20',
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
const CATCHTEN = {
  bg: 'bg-game-bg-catchten',
  footer: 'bg-game-bg-catchten-dark border-white/20',
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
const ALUETTE = {
  bg: 'bg-game-bg-aluette',
  footer: 'bg-game-bg-aluette-dark border-white/20',
} as const;
const MINCHIATE = {
  bg: 'bg-game-bg-minchiate',
  footer: 'bg-game-bg-minchiate-dark border-white/20',
} as const;
const TAROCCHINI = {
  bg: 'bg-game-bg-tarocchini',
  footer: 'bg-game-bg-tarocchini-dark border-white/20',
} as const;
const GANJIFA = {
  bg: 'bg-game-bg-ganjifa',
  footer: 'bg-game-bg-ganjifa-dark border-white/20',
} as const;
const VIRA = {
  bg: 'bg-game-bg-vira',
  footer: 'bg-game-bg-vira-dark border-white/20',
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
const SPOONS = {
  bg: 'bg-game-bg-spoons',
  footer: 'bg-game-bg-spoons-dark border-white/20',
} as const;
const KEMPS = {
  bg: 'bg-game-bg-kemps',
  footer: 'bg-game-bg-kemps-dark border-white/20',
} as const;
const CUCKOO = {
  bg: 'bg-game-bg-cuckoo',
  footer: 'bg-game-bg-cuckoo-dark border-white/20',
} as const;
const PISHTI = {
  bg: 'bg-game-bg-pishti',
  footer: 'bg-game-bg-pishti-dark border-white/20',
} as const;
const CUARENTA = {
  bg: 'bg-game-bg-cuarenta',
  footer: 'bg-game-bg-cuarenta-dark border-white/20',
} as const;
const FARO = {
  bg: 'bg-game-bg-faro',
  footer: 'bg-game-bg-faro-dark border-white/20',
} as const;
const ALLFOURS = {
  bg: 'bg-game-bg-allfours',
  footer: 'bg-game-bg-allfours-dark border-white/20',
} as const;
const NINETYNINE = {
  bg: 'bg-game-bg-ninetynine',
  footer: 'bg-game-bg-ninetynine-dark border-white/20',
} as const;
const GUTS = {
  bg: 'bg-game-bg-guts',
  footer: 'bg-game-bg-guts-dark border-white/20',
} as const;
const ANACONDA = {
  bg: 'bg-game-bg-anaconda',
  footer: 'bg-game-bg-anaconda-dark border-white/20',
} as const;
const BOUILLOTTE = {
  bg: 'bg-game-bg-bouillotte',
  footer: 'bg-game-bg-bouillotte-dark border-white/20',
} as const;
const PRIMERO = {
  bg: 'bg-game-bg-primero',
  footer: 'bg-game-bg-primero-dark border-white/20',
} as const;
const CARIOCA = {
  bg: 'bg-game-bg-carioca',
  footer: 'bg-game-bg-carioca-dark border-white/20',
} as const;
const MICHIGAN = {
  bg: 'bg-game-bg-michigan',
  footer: 'bg-game-bg-michigan-dark border-white/20',
} as const;
const SAMBA = {
  bg: 'bg-game-bg-samba',
  footer: 'bg-game-bg-samba-dark border-white/20',
} as const;
const INDIANRUMMY = {
  bg: 'bg-game-bg-indianrummy',
  footer: 'bg-game-bg-indianrummy-dark border-white/20',
} as const;
const MACHIAVELLI = {
  bg: 'bg-game-bg-machiavelli',
  footer: 'bg-game-bg-machiavelli-dark border-white/20',
} as const;
const PAN = {
  bg: 'bg-game-bg-pan',
  footer: 'bg-game-bg-pan-dark border-white/20',
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
  oichokabu: CASINO,
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
  fivecardstud: POKER,
  soko: POKER,
  razz: POKER,
  sevencardstudhilo: POKER,
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
  king: KING,
  tysiac: TYSIAC,
  calabresella: CALABRESELLA,
  ombre: OMBRE,
  ulti: ULTI,
  scarto: OMBRE,
  cego: OMBRE,
  frenchtarot: OMBRE,
  koenigrufen: OMBRE,
  cinch: CINCH,
  loo: LOO,
  basra: BASRA,
  hachihachi: GREEN,
  koikoi: GREEN,
  gostop: GREEN,
  tablanet: TABLANET,
  trenteetquarante: TRENTEETQUARANTE,
  sedma: SEDMA,
  knockoutwhist: KNOCKOUTWHIST,
  spoilfive: SPOILFIVE,
  solowhist: SOLOWHIST,
  fortyfives: FORTYFIVES,
  nap: NAP,
  aluette: ALUETTE,
  minchiate: MINCHIATE,
  tarocchini: TAROCCHINI,
  ganjifa: GANJIFA,
  preference: PREFERENCE,
  vira: VIRA,
  twentynine: TWENTYNINE,
  courtpiece: COURTPIECE,
  bezique: BEZIQUE,
  ecarte: ECARTE,
  threecardbrag: THREECARDBRAG,
  teenpatti: TEENPATTI,
  spoons: SPOONS,
  kemps: KEMPS,
  cuckoo: CUCKOO,
  pishti: PISHTI,
  cuarenta: CUARENTA,
  faro: FARO,
  openfacechinese: CASINO,
  russianbank: GREEN,
  pitch: BLUE,
  twotenjack: BLUE,
  ohhell: BLUE,
  wizard: BLUE,
  ninetynine: NINETYNINE,
  euchre: BLUE,
  bridge: BLUE,
  napoleon: BLUE,
  whist: BLUE,
  catchten: CATCHTEN,
  pinochle: BLUE,
  piquet: BLUE,
  skat: BLUE,
  belote: BLUE,
  jass: BLUE,
  watten: WATTEN,
  gaigel: BLUE,
  mighty: BLUE,
  fivehundred: BLUE,
  rook: BLUE,
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
  zheng: GREEN,
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
  prsi: PRSI,
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
  streetsandalleys: CASINO,
  kingalbert: CASINO,
  flowergarden: CASINO,
  fortyandeight: CASINO,
  sultan: CASINO,
  agnes: CASINO,
  canfield: CASINO,
  osmosis: CASINO,
  bristol: CASINO,
  labellelucie: GREEN,
  simplesimon: GREEN,
  doubleklondike: GREEN,
  blackhole: GREEN,
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
  sirtommy: CASINO,
  fourseasons: CASINO,
  colorado: CASINO,
  auldlangsyne: CASINO,
  bisley: CASINO,
  napoleonssquare: CASINO,
  grandfathersclock: CASINO,
  duchess: CASINO,
  windmill: CASINO,
  americantoad: CASINO,
  congress: CASINO,
  terrace: CASINO,
  braid: CASINO,
  pontoon: CASINO,
  settemezzo: CASINO,
  niuniu: CASINO,
  bura: CASINO,
  mushi: GREEN,
  toepen: BLUE,
  chineseten: GREEN,
  skitgubbe: GREEN,
  laughandliedown: GREEN,
  sjavs: BLUE,
  trex: BLUE,
  loba: GREEN,
  desmoche: GREEN,
  zwicker: BLUE,
  poch: CASINO,
  popejoan: CASINO,
  nainjaune: CASINO,
  kille: CASINO,
  klaberjass: CASINO,
  kaiser: CASINO,
  boston: CASINO,
  vint: CASINO,
  bideuchre: CASINO,
  sixbidsolo: CASINO,
  karnoffel: GREEN,
  literature: BLUE,
  guandan: GREEN,
  shengji: GREEN,
  missmilligan: CASINO,
  crescent: CASINO,
  // Counting/Rummy
  ginrummy: BLUE,
  indianrummy: INDIANRUMMY,
  machiavelli: MACHIAVELLI,
  conquian: BLUE,
  chinchon: GREEN,
  threethirteen: BLUE,
  tonk: BLUE,
  thirtyone: CASINO,
  yaniv: BLUE,
  gongzhu: GREEN,
  tressette: GREEN,
  canasta: BLUE,
  samba: SAMBA,
  handandfoot: BLUE,
  burraco: GREEN,
  cribbage: BLUE,
  sevenbridge: BLUE,
  contractrummy: BLUE,
  carioca: CARIOCA,
  kalooki: GREEN,
  rummy500: BLUE,
  beggarmyneighbour: GREEN,
  allfours: ALLFOURS,
  guts: GUTS,
  anaconda: ANACONDA,
  bouillotte: BOUILLOTTE,
  primero: PRIMERO,
  michigan: MICHIGAN,
  pan: PAN,
};
