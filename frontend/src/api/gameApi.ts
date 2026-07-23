import type {
  AccordionResponse,
  AcesUpResponse,
  ActionLogResponse,
  AgnesResponse,
  AllFoursResponse,
  AnacondaResponse,
  BaccaratResponse,
  BadugiResponse,
  BakersDozenMoveZone,
  BakersDozenResponse,
  BarbuResponse,
  BasraResponse,
  BeggarMyNeighbourResponse,
  BeleagueredCastleMoveZone,
  BeleagueredCastleResponse,
  BeloteResponse,
  BeziqueResponse,
  BidWhistResponse,
  BigTwoConfigInput,
  BigTwoResponse,
  BlackHoleResponse,
  BlackJackResponse,
  BlackJackSwitchResponse,
  BouillotteResponse,
  BourreResponse,
  BridgeResponse,
  BriscolaConfig,
  BriscolaResponse,
  BristolMoveZone,
  BristolResponse,
  BurracoResponse,
  CalabresellaResponse,
  CalculationMoveZone,
  CalculationResponse,
  CallBreakResponse,
  CanastaResponse,
  CanfieldResponse,
  CaribbeanStudResponse,
  CariocaResponse,
  CasinoHoldemResponse,
  CasinoWarResponse,
  CassinoResponse,
  CatchTenConfig,
  CatchTenResponse,
  CegoResponse,
  ChinchonResponse,
  ChinesePokerResponse,
  CinchResponse,
  ClockSolitaireResponse,
  ConquianResponse,
  ContractRummyResponse,
  CourtPieceResponse,
  CrazyEightsResponse,
  CrescentMoveZone,
  CrescentResponse,
  CribbageResponse,
  CruelResponse,
  CuarentaResponse,
  CuckooResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DeuceToSevenResponse,
  DoppelkopfResponse,
  DoubleKlondikeResponse,
  DoubtConfig,
  DoubtResponse,
  DoudizhuResponse,
  DragonTigerResponse,
  DurakConfigInput,
  DurakResponse,
  EasthavenResponse,
  EcarteResponse,
  EgyptianRatscrewResponse,
  EightOffResponse,
  EscobaResponse,
  EuchreResponse,
  FaroResponse,
  FiftyOneResponse,
  FiveCardStudResponse,
  FiveHundredResponse,
  FlowerGardenMoveZone,
  FlowerGardenResponse,
  FortyAndEightMoveZone,
  FortyAndEightResponse,
  FortyFivesResponse,
  FortyThievesMoveZone,
  FortyThievesResponse,
  FourCardPokerResponse,
  FreeCellResponse,
  FrenchTarotResponse,
  GaigelResponse,
  GapsResponse,
  GinRummyResponse,
  GoFishResponse,
  GolfResponse,
  GongZhuResponse,
  GoStopResponse,
  GutsResponse,
  HachiHachiResponse,
  HandAndFootResponse,
  HeartsResponse,
  HighCardFlushResponse,
  HoldemResponse,
  IndianPokerResponse,
  IndianRummyResponse,
  JassResponse,
  KalookiResponse,
  KempsResponse,
  KingAlbertMoveZone,
  KingAlbertResponse,
  KingResponse,
  KlaverjasResponse,
  KlondikeResponse,
  KnockoutWhistResponse,
  KoenigrufenResponse,
  KoiKoiResponse,
  LaBelleLucieResponse,
  LetItRideResponse,
  LooResponse,
  MacauResponse,
  MachiavelliResponse,
  ManilleResponse,
  MaoResponse,
  MariasResponse,
  MemoryResponse,
  MichiganResponse,
  MightyResponse,
  MississippiStudResponse,
  MonteCarloResponse,
  MusResponse,
  NapoleonResponse,
  NapResponse,
  NertzConfig as NertzConfigType,
  NertzMoveZone,
  NertzResponse,
  NinetyNineResponse,
  OasisPokerResponse,
  OhHellResponse,
  OichoKabuResponse,
  OldMaidResponse,
  OmahaResponse,
  OmbreResponse,
  OpenFaceChineseResponse,
  OsmosisResponse,
  PageOneResponse,
  PaiGowResponse,
  PanResponse,
  PenguinResponse,
  PigsTailResponse,
  PineappleResponse,
  PinochleResponse,
  PiquetConfig as PiquetConfigType,
  PiquetResponse,
  PishtiResponse,
  PitchResponse,
  PokerResponse,
  PokerSquaresResponse,
  PreferenceResponse,
  PresidentResponse,
  PrimeroResponse,
  PrsiResponse,
  PyramidResponse,
  RedDogResponse,
  RookResponse,
  Rummy500Response,
  RussianBankResponse,
  RussianPokerResponse,
  RussianSolitaireResponse,
  SambaResponse,
  ScartoResponse,
  SchnapsenConfig,
  SchnapsenResponse,
  ScopaResponse,
  ScoponeResponse,
  ScorpionResponse,
  SeahavenTowersResponse,
  SedmaResponse,
  SevenBridgeResponse,
  SevenCardStudResponse,
  SevensResponse,
  SheepsheadResponse,
  ShitheadConfig as ShitheadConfigType,
  ShitheadResponse,
  ShortDeckResponse,
  SimpleSimonResponse,
  SixCardGolfResponse,
  SkatConfig as SkatConfigType,
  SkatResponse,
  SlapjackResponse,
  SoloWhistResponse,
  SpadesResponse,
  SpeedConfig,
  SpeedResponse,
  SpideretteResponse,
  SpiderResponse,
  SpiteAndMaliceMoveZone,
  SpiteAndMaliceResponse,
  SpoilFiveResponse,
  SpoonsResponse,
  StreetsAndAlleysMoveZone,
  StreetsAndAlleysResponse,
  SuecaResponse,
  SultanMoveZone,
  SultanResponse,
  TablanetResponse,
  TarneebResponse,
  TeenPattiResponse,
  TexasHoldemBonusResponse,
  ThirtyOneResponse,
  ThreeCardBragResponse,
  ThreeCardResponse,
  ThreeThirteenResponse,
  TichuResponse,
  TienLenConfigInput,
  TienLenResponse,
  TonkResponse,
  TrashResponse,
  TrenteEtQuaranteResponse,
  TressetteResponse,
  TriPeaksResponse,
  TrucoConfig,
  TrucoResponse,
  TuteResponse,
  TwentyNineResponse,
  TwoTenJackResponse,
  TysiacResponse,
  UltimateTexasHoldemResponse,
  UltiResponse,
  VideoPokerResponse,
  WarResponse,
  WaspResponse,
  WattenResponse,
  WhistConfig,
  WhistResponse,
  WizardResponse,
  YanivResponse,
  YukonResponse,
  ZhengConfigInput,
  ZhengResponse,
} from '../types/card';

/** Unique session identifier for correlating API requests. */
export const sessionId: string = crypto.randomUUID();

/** Worker base URLs for Cloudflare deployment. Empty strings for Docker (relative URLs). */
const WORKER_CASINO = import.meta.env.VITE_WORKER_CASINO_URL || '';
const WORKER_CLASSIC = import.meta.env.VITE_WORKER_CLASSIC_URL || '';
const WORKER_SOLO = import.meta.env.VITE_WORKER_SOLO_URL || '';
const WORKER_EXTRA = import.meta.env.VITE_WORKER_EXTRA_URL || '';

/** Maps each game to its Worker base URL. */
const workerUrl: Record<string, string> = {
  blackjack: WORKER_CASINO,
  spanish21: WORKER_CASINO,
  baccarat: WORKER_CASINO,
  poker: WORKER_CASINO,
  holdem: WORKER_CASINO,
  omaha: WORKER_CASINO,
  omahahilo: WORKER_CASINO,
  bigo: WORKER_CASINO,
  bigohilo: WORKER_CASINO,
  shortdeck: WORKER_CASINO,
  indianpoker: WORKER_CASINO,
  videopoker: WORKER_CASINO,
  deuceswild: WORKER_CASINO,
  jokerpoker: WORKER_CASINO,
  threecard: WORKER_CASINO,
  caribbeanstud: WORKER_CASINO,
  texasholdembonus: WORKER_CASINO,
  casinoholdem: WORKER_CASINO,
  paigow: WORKER_CASINO,
  pineapple: WORKER_CASINO,
  crazypineapple: WORKER_CASINO,
  irishpoker: WORKER_CASINO,
  sevencardstud: WORKER_CASINO,
  fivecardstud: WORKER_CASINO,
  razz: WORKER_CASINO,
  badugi: WORKER_CASINO,
  deucetoseven: WORKER_CASINO,
  ecarte: WORKER_CASINO,
  threecardbrag: WORKER_CASINO,
  teenpatti: WORKER_CASINO,
  spoons: WORKER_CLASSIC,
  kemps: WORKER_CASINO,
  cuckoo: WORKER_CLASSIC,
  pishti: WORKER_CASINO,
  cuarenta: WORKER_CASINO,
  faro: WORKER_CASINO,
  openfacechinese: WORKER_CASINO,
  calculation: WORKER_SOLO,
  hearts: WORKER_CLASSIC,
  spades: WORKER_CLASSIC,
  pitch: WORKER_CLASSIC,
  euchre: WORKER_SOLO,
  bridge: WORKER_CASINO,
  napoleon: WORKER_CASINO,
  ninetynine: WORKER_CLASSIC,
  ohhell: WORKER_CLASSIC,
  wizard: WORKER_EXTRA,
  oldmaid: WORKER_CLASSIC,
  doubt: WORKER_CLASSIC,
  durak: WORKER_CLASSIC,
  daifugo: WORKER_CLASSIC,
  bigtwo: WORKER_CLASSIC,
  tienlen: WORKER_SOLO,
  zheng: WORKER_SOLO,
  sevens: WORKER_CLASSIC,
  crazyeights: WORKER_CLASSIC,
  prsi: WORKER_CLASSIC,
  pageone: WORKER_CLASSIC,
  speed: WORKER_CLASSIC,
  war: WORKER_CLASSIC,
  fiftyone: WORKER_CLASSIC,
  gofish: WORKER_CLASSIC,
  pinochle: WORKER_CLASSIC,
  pigtail: WORKER_CLASSIC,
  twotenjack: WORKER_CLASSIC,
  klondike: WORKER_SOLO,
  freecell: WORKER_SOLO,
  bakersgame: WORKER_SOLO,
  seahaventowers: WORKER_SOLO,
  cruel: WORKER_SOLO,
  spider: WORKER_SOLO,
  pyramid: WORKER_SOLO,
  pokersquares: WORKER_SOLO,
  tripeaks: WORKER_SOLO,
  memory: WORKER_SOLO,
  ginrummy: WORKER_EXTRA,
  indianrummy: WORKER_EXTRA,
  machiavelli: WORKER_EXTRA,
  conquian: WORKER_EXTRA,
  chinchon: WORKER_EXTRA,
  threethirteen: WORKER_EXTRA,
  canasta: WORKER_EXTRA,
  samba: WORKER_EXTRA,
  handandfoot: WORKER_EXTRA,
  burraco: WORKER_EXTRA,
  cribbage: WORKER_SOLO,
  golf: WORKER_SOLO,
  acesup: WORKER_SOLO,
  clocksolitaire: WORKER_SOLO,
  fortythieves: WORKER_SOLO,
  canfield: WORKER_SOLO,
  osmosis: WORKER_SOLO,
  fivehundred: WORKER_SOLO,
  yukon: WORKER_SOLO,
  russiansolitaire: WORKER_SOLO,
  scorpion: WORKER_SOLO,
  wasp: WORKER_SOLO,
  accordion: WORKER_SOLO,
  sevenbridge: WORKER_SOLO,
  trash: WORKER_CLASSIC,
  whist: WORKER_CLASSIC,
  catchten: WORKER_CLASSIC,
  letitride: WORKER_CASINO,
  reddog: WORKER_CASINO,
  casinowar: WORKER_CASINO,
  president: WORKER_CLASSIC,
  cassino: WORKER_CLASSIC,
  spiteandmalice: WORKER_CLASSIC,
  skat: WORKER_CASINO,
  shithead: WORKER_CLASSIC,
  nertz: WORKER_CLASSIC,
  slapjack: WORKER_CLASSIC,
  egyptianratscrew: WORKER_CLASSIC,
  bakersdozen: WORKER_SOLO,
  thirtyone: WORKER_SOLO,
  yaniv: WORKER_SOLO,
  tressette: WORKER_CASINO,
  tonk: WORKER_CLASSIC,
  dragontiger: WORKER_CASINO,
  blackjackswitch: WORKER_CASINO,
  montecarlo: WORKER_SOLO,
  contractrummy: WORKER_EXTRA,
  carioca: WORKER_EXTRA,
  kalooki: WORKER_EXTRA,
  ultimatetexasholdem: WORKER_CASINO,
  crescent: WORKER_SOLO,
  mississippistud: WORKER_CASINO,
  belote: WORKER_CASINO,
  spiderette: WORKER_SOLO,
  mighty: WORKER_CASINO,
  oasispoker: WORKER_CASINO,
  russianpoker: WORKER_CASINO,
  beleagueredcastle: WORKER_SOLO,
  piquet: WORKER_SOLO,
  callbreak: WORKER_CLASSIC,
  tarneeb: WORKER_CASINO,
  highcardflush: WORKER_CASINO,
  briscola: WORKER_CLASSIC,
  schnapsen: WORKER_SOLO,
  gaps: WORKER_SOLO,
  fourcardpoker: WORKER_CASINO,
  rummy500: WORKER_EXTRA,
  streetsandalleys: WORKER_EXTRA,
  kingalbert: WORKER_EXTRA,
  flowergarden: WORKER_EXTRA,
  fortyandeight: WORKER_EXTRA,
  sultan: WORKER_EXTRA,
  agnes: WORKER_EXTRA,
  jass: WORKER_EXTRA,
  gaigel: WORKER_EXTRA,
  king: WORKER_EXTRA,
  tysiac: WORKER_EXTRA,
  calabresella: WORKER_EXTRA,
  ombre: WORKER_EXTRA,
  ulti: WORKER_EXTRA,
  scarto: WORKER_SOLO,
  cego: WORKER_SOLO,
  frenchtarot: WORKER_EXTRA,
  koenigrufen: WORKER_EXTRA,
  rook: WORKER_EXTRA,
  cinch: WORKER_EXTRA,
  loo: WORKER_EXTRA,
  basra: WORKER_EXTRA,
  hachihachi: WORKER_EXTRA,
  koikoi: WORKER_EXTRA,
  gostop: WORKER_EXTRA,
  tablanet: WORKER_EXTRA,
  trenteetquarante: WORKER_EXTRA,
  guts: WORKER_EXTRA,
  anaconda: WORKER_EXTRA,
  bouillotte: WORKER_EXTRA,
  primero: WORKER_EXTRA,
  michigan: WORKER_EXTRA,
  watten: WORKER_EXTRA,
  pan: WORKER_EXTRA,
  oichokabu: WORKER_EXTRA,
  eightoff: WORKER_SOLO,
  penguin: WORKER_SOLO,
  chinesepoker: WORKER_CASINO,
  sixcardgolf: WORKER_CLASSIC,
  doudizhu: WORKER_CLASSIC,
  truco: WORKER_CLASSIC,
  scopa: WORKER_CLASSIC,
  scopone: WORKER_CLASSIC,
  escoba: WORKER_CLASSIC,
  barbu: WORKER_SOLO,
  macau: WORKER_SOLO,
  mao: WORKER_SOLO,
  russianbank: WORKER_SOLO,
  labellelucie: WORKER_CLASSIC,
  simplesimon: WORKER_CLASSIC,
  doubleklondike: WORKER_CLASSIC,
  blackhole: WORKER_SOLO,
  gongzhu: WORKER_SOLO,
  bristol: WORKER_SOLO,
  bidwhist: WORKER_SOLO,
  easthaven: WORKER_SOLO,
  tichu: WORKER_CLASSIC,
  bourre: WORKER_CASINO,
  sheepshead: WORKER_CASINO,
  doppelkopf: WORKER_CASINO,
  mus: WORKER_CASINO,
  tute: WORKER_CASINO,
  sueca: WORKER_CASINO,
  klaverjas: WORKER_CLASSIC,
  manille: WORKER_CLASSIC,
  marias: WORKER_CLASSIC,
  sedma: WORKER_CLASSIC,
  knockoutwhist: WORKER_CLASSIC,
  spoilfive: WORKER_CLASSIC,
  solowhist: WORKER_CLASSIC,
  fortyfives: WORKER_CASINO,
  nap: WORKER_CLASSIC,
  preference: WORKER_CLASSIC,
  twentynine: WORKER_CASINO,
  courtpiece: WORKER_CASINO,
  bezique: WORKER_CLASSIC,
  beggarmyneighbour: WORKER_CASINO,
  allfours: WORKER_CLASSIC,
};

async function postJson<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`HTTP error: ${res.status}`);
  return res.json() as Promise<T>;
}

function gameExec<T>(game: string, body: Record<string, unknown>): Promise<T> {
  const base = workerUrl[game] || '';
  return postJson<T>(`${base}/${game}/exec`, { ...body, sessionId });
}

/** Configuration options for BlackJack game settings. */
export interface BlackJackConfigInput {
  dealerHitsSoft17?: boolean;
  cpuPlayerCount?: number;
  countingEnabled?: boolean;
  doubleAfterSplit?: boolean;
  countingSystem?: number;
  deckPenetration?: number;
  surrenderRule?: number;
}

/** Side bet and multi-hand options for BlackJack. */
export interface BlackJackBetOptions {
  perfectPairsBet?: number;
  twentyOnePlus3Bet?: number;
  handCount?: number;
}

/** Type alias for BlackJack/Spanish21 exec command. */
export type BlackJackCommand =
  | 'reset'
  | 'hit'
  | 'stand'
  | 'bet'
  | 'doubledown'
  | 'split'
  | 'insurance'
  | 'declineinsurance'
  | 'surrender'
  | 'togglehint'
  | 'setdeckcount'
  | 'togglesoft17'
  | 'togglecounting'
  | 'toggledas'
  | 'setcountingsystem'
  | 'setpenetration'
  | 'setcpucount'
  | 'earlysurrender'
  | 'declineearlysurrender'
  | 'setsurrenderrule';

/**
 * Factory for BlackJack-shaped APIs whose request body is
 * `{ command, amount, ...config, ...betOptions }`. Spanish 21 reuses the
 * BlackJack response and command union, so both clients are constructed
 * via the same factory rather than duplicating the eight-token shape.
 *
 * Kept narrow on purpose — only games that share the BlackJack command
 * union and bet-option payload should use it. See issue #1550.
 */
function createBlackJackLikeApi(game: string) {
  return {
    exec: (
      command: BlackJackCommand,
      amount?: number,
      config?: BlackJackConfigInput,
      betOptions?: BlackJackBetOptions,
    ) => gameExec<BlackJackResponse>(game, { command, amount, ...config, ...betOptions }),
  };
}

/** API client for the BlackJack /blackjack/exec endpoint. */
export const blackjackApi = createBlackJackLikeApi('blackjack');

/** API client for the Spanish 21 /spanish21/exec endpoint (shares BlackJack response shape). */
export const spanish21Api = createBlackJackLikeApi('spanish21');

/** Configuration options for Poker game settings. */
export interface PokerConfigInput {
  cpuCount?: number;
  jokerCount?: number;
  bettingLimit?: number;
  isLowball?: boolean;
  cpuMetaAI?: boolean;
}

/** API client for the Poker /poker/exec endpoint. */
export const pokerApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin' | 'odds',
    indices?: number[],
    amount?: number,
    config?: PokerConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) => gameExec<PokerResponse>('poker', { command, indices, amount, humanPlayMs, profile, ...config }),
};

/** Configuration options for Badugi game settings. */
export interface BadugiConfigInput {
  cpuCount?: number;
  bettingLimit?: number;
  cpuMetaAI?: boolean;
}

/** API client for the Badugi /badugi/exec endpoint. */
export const badugiApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin',
    indices?: number[],
    amount?: number,
    config?: BadugiConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) => gameExec<BadugiResponse>('badugi', { command, indices, amount, humanPlayMs, profile, ...config }),
};

/** Configuration options for 2-7 Triple Draw game settings. */
export interface DeuceToSevenConfigInput {
  cpuCount?: number;
  bettingLimit?: number;
  cpuMetaAI?: boolean;
}

/** API client for the 2-7 Triple Draw /deucetoseven/exec endpoint. */
export const deuceToSevenApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin',
    indices?: number[],
    amount?: number,
    config?: DeuceToSevenConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<DeuceToSevenResponse>('deucetoseven', {
      command,
      indices,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** API client for the Old Maid /oldmaid/exec endpoint. */
export const oldmaidApi = {
  exec: (
    command: 'reset' | 'draw' | 'shuffle' | 'reorder',
    drawIdx?: number,
    mode?: number,
    cpuPlacementStrategy?: boolean,
    reorderIndices?: number[],
    cpuMemoryAI?: boolean,
    cpuHesitationEnabled?: boolean,
    cpuMetaAI?: boolean,
    profile?: unknown,
  ) =>
    gameExec<OldMaidResponse>('oldmaid', {
      command,
      drawIdx,
      mode,
      cpuPlacementStrategy,
      reorderIndices,
      cpuMemoryAI,
      cpuHesitationEnabled,
      cpuMetaAI,
      profile,
    }),
};

/** API client for the Daifugo /daifugo/exec endpoint. */
export const daifugoApi = {
  exec: (command: 'reset' | 'play' | 'sort', indices?: number[], config?: DaifugoConfigInput, sortMode?: number) =>
    gameExec<DaifugoResponse>('daifugo', { command, indices, config, sortMode }),
};

/** API client for the Big Two /bigtwo/exec endpoint. */
export const bigtwoApi = {
  exec: (command: 'reset' | 'play', indices?: number[], config?: BigTwoConfigInput) =>
    gameExec<BigTwoResponse>('bigtwo', { command, indices, config }),
};

/** API client for the Tien Len /tienlen/exec endpoint. */
export const tienlenApi = {
  exec: (command: 'reset' | 'play', indices?: number[], config?: TienLenConfigInput) =>
    gameExec<TienLenResponse>('tienlen', { command, indices, config }),
};

/** API client for the Zheng Shangyou /zheng/exec endpoint (empty indices = pass). */
export const zhengApi = {
  exec: (command: 'reset' | 'play', indices?: number[], config?: ZhengConfigInput) =>
    gameExec<ZhengResponse>('zheng', { command, indices, config }),
};

/** API client for the Durak /durak/exec endpoint. */
export const durakApi = {
  exec: (
    command: 'reset' | 'attack' | 'defend' | 'pass' | 'take' | 'sort',
    cardIdx?: number,
    attackIdx?: number,
    config?: DurakConfigInput,
    sortMode?: number,
  ) => gameExec<DurakResponse>('durak', { command, cardIdx, attackIdx, config, sortMode }),
};

/** API client for the Doubt /doubt/exec endpoint. */
export const doubtApi = {
  exec: (
    command: 'reset' | 'play' | 'doubt' | 'skip',
    cardIndices?: number[],
    claimedValue?: number,
    doubterIndices?: number[],
    config?: DoubtConfig,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<DoubtResponse>('doubt', {
      command,
      cardIndices,
      claimedValue,
      doubterIndices,
      humanPlayMs,
      profile,
      doubtWindowSec: config?.doubtWindowSec,
      cpuMemoryLevel: config?.cpuMemoryLevel,
      penaltyDrawLimit: config?.penaltyDrawLimit,
      cpuHesitationEnabled: config?.cpuHesitationEnabled,
      cpuMetaAI: config?.cpuMetaAI,
    }),
};

/** Configuration options for Sevens game settings. */
export interface SevensConfigInput {
  tunnelEnabled?: boolean;
  tunnelSkipWidth?: number;
  jokerCount?: number;
  cpuStrategy?: number;
  maxPasses?: number;
  noJokerFinish?: boolean;
  jokerReclaim?: boolean;
  endStop?: boolean;
  jokerConsecutiveBanned?: boolean;
}

/** API client for the Sevens /sevens/exec endpoint. */
export const sevensApi = {
  exec: (
    command: 'reset' | 'play' | 'joker',
    index = -1,
    jokerTargetSuit = 0,
    jokerTargetValue = 0,
    config?: SevensConfigInput,
  ) =>
    gameExec<SevensResponse>('sevens', {
      command,
      index,
      jokerTargetSuit,
      jokerTargetValue,
      ...config,
    }),
};

/** Configuration options for Texas Hold'em game settings. */
export interface HoldemConfigInput {
  smallBlind?: number;
  bigBlind?: number;
  tournamentMode?: boolean;
  blindLevelHands?: number;
  blindMultiplier?: number;
  bettingLimit?: number;
  tableSize?: number;
  rebuyEnabled?: boolean;
  rebuyMaxCount?: number;
  rebuyChips?: number;
  rebuyPeriodHands?: number;
  addonEnabled?: boolean;
  addonChips?: number;
  addonAfterHand?: number;
  cpuMetaAI?: boolean;
}

/** Command set shared by Hold'em-family games. */
type HoldemLikeCommand =
  | 'reset'
  | 'fold'
  | 'check'
  | 'call'
  | 'bet'
  | 'raise'
  | 'allin'
  | 'rebuy'
  | 'skiprebuy'
  | 'addon'
  | 'skipaddon'
  | 'muck'
  | 'show';

/** Factory for Hold'em-family APIs that share the same exec pattern. */
function createHoldemLikeApi<T, C = HoldemConfigInput>(game: string) {
  return {
    exec: (command: HoldemLikeCommand, amount?: number, config?: C, humanPlayMs?: number, profile?: unknown) =>
      gameExec<T>(game, { command, amount, humanPlayMs, profile, ...(config as Record<string, unknown>) }),
  };
}

/** API client for the Texas Hold'em /holdem/exec endpoint. */
export const holdemApi = createHoldemLikeApi<HoldemResponse>('holdem');

/** API client for the Omaha Hold'em /omaha/exec endpoint. */
export const omahaApi = createHoldemLikeApi<OmahaResponse>('omaha');

/** API client for the Omaha Hi-Lo / 8 or Better /omahahilo/exec endpoint.
 * Shares the OmahaResponse shape — additional Hi-Lo fields (LowBestHand,
 * LowQualifies, HiWonAmount, LowWonAmount) are surfaced via omitempty
 * JSON encoding so the same TypeScript type works for both. */
export const omahaHiLoApi = createHoldemLikeApi<OmahaResponse>('omahahilo');

/** API client for the 5 Card Omaha (Big O) /bigo/exec endpoint.
 * Shares the OmahaResponse shape — Big O is Omaha dealt 5 hole cards. */
export const bigOApi = createHoldemLikeApi<OmahaResponse>('bigo');

/** API client for the 5 Card Omaha Hi-Lo (Big O) /bigohilo/exec endpoint.
 * Shares the OmahaResponse shape (Hi-Lo split fields surfaced via omitempty). */
export const bigOHiLoApi = createHoldemLikeApi<OmahaResponse>('bigohilo');

/** API client for the Short Deck Hold'em /shortdeck/exec endpoint. */
export const shortdeckApi = createHoldemLikeApi<ShortDeckResponse>('shortdeck');

/** Configuration options for Seven Card Stud game settings. */
export interface SevenCardStudConfigInput {
  ante?: number;
  bringIn?: number;
  smallBet?: number;
  bigBet?: number;
  tournamentMode?: boolean;
  anteLevelHands?: number;
  anteMultiplier?: number;
  bettingLimit?: number;
  tableSize?: number;
  rebuyEnabled?: boolean;
  rebuyMaxCount?: number;
  rebuyChips?: number;
  rebuyPeriodHands?: number;
  addonEnabled?: boolean;
  addonChips?: number;
  addonAfterHand?: number;
  cpuMetaAI?: boolean;
}

/** API client for the Seven Card Stud /sevencardstud/exec endpoint. */
export const sevenCardStudApi = createHoldemLikeApi<SevenCardStudResponse, SevenCardStudConfigInput>('sevencardstud');

/** Configuration options for Five Card Stud game settings. */
export interface FiveCardStudConfigInput {
  ante?: number;
  bringIn?: number;
  smallBet?: number;
  bigBet?: number;
  tournamentMode?: boolean;
  anteLevelHands?: number;
  anteMultiplier?: number;
  bettingLimit?: number;
  tableSize?: number;
  rebuyEnabled?: boolean;
  rebuyMaxCount?: number;
  rebuyChips?: number;
  rebuyPeriodHands?: number;
  addonEnabled?: boolean;
  addonChips?: number;
  addonAfterHand?: number;
  cpuMetaAI?: boolean;
}

/** API client for the Five Card Stud /fivecardstud/exec endpoint. */
export const fiveCardStudApi = createHoldemLikeApi<FiveCardStudResponse, FiveCardStudConfigInput>('fivecardstud');

/** API client for the Razz /razz/exec endpoint. */
export const razzApi = createHoldemLikeApi<SevenCardStudResponse, SevenCardStudConfigInput>('razz');

/** Configuration options for Pineapple Poker (extends Hold'em with cardIdx/cardIdxs for discard). */
export interface PineappleConfigInput extends HoldemConfigInput {
  cardIdx?: number;
  /** Multiple discard indices, submitted together (Irish Poker's 2-card discard). */
  cardIdxs?: number[];
}

/** API client for the Pineapple Poker /pineapple/exec endpoint. */
export const pineappleApi = {
  exec: (
    command: HoldemLikeCommand | 'discard',
    amount?: number,
    config?: PineappleConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<PineappleResponse>('pineapple', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** API client for the Crazy Pineapple Poker /crazypineapple/exec endpoint. */
export const crazyPineappleApi = {
  exec: (
    command: HoldemLikeCommand | 'discard',
    amount?: number,
    config?: PineappleConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<PineappleResponse>('crazypineapple', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** API client for the Irish Poker /irishpoker/exec endpoint. */
export const irishPokerApi = {
  exec: (
    command: HoldemLikeCommand | 'discard',
    amount?: number,
    config?: PineappleConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<PineappleResponse>('irishpoker', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** Configuration options for Hearts game settings. */
export interface HeartsConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
  omnibusJD?: boolean;
}

/** API client for the Hearts /hearts/exec endpoint. */
export const heartsApi = {
  exec: (
    command: 'reset' | 'pass' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndices?: number[],
    cardIndex?: number,
    config?: HeartsConfigInput,
  ) =>
    gameExec<HeartsResponse>('hearts', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};

/** Configuration options for Gong Zhu game settings. */
export interface GongZhuConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Gong Zhu /gongzhu/exec endpoint. */
export const gongzhuApi = {
  exec: (
    command: 'reset' | 'expose' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndices?: number[],
    cardIndex?: number,
    config?: GongZhuConfigInput,
  ) =>
    gameExec<GongZhuResponse>('gongzhu', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};

/** Configuration options for Tressette game settings. */
export interface TressetteConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** API client for the Tressette /tressette/exec endpoint. */
export const tressetteApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndices?: number[],
    cardIndex?: number,
    config?: TressetteConfigInput,
  ) =>
    gameExec<TressetteResponse>('tressette', {
      command,
      cardIndices,
      cardIndex,
      config,
    }),
};

/** Configuration options for Spades game settings. */
export interface SpadesConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
  nilBonus?: number;
  bagPenaltyThreshold?: number;
}

/** Factory for bid-play trick-taking APIs that share the same exec pattern. */
function createBidPlayApi<T, C>(game: string) {
  return {
    exec: (
      command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
      bid?: number,
      cardIndex?: number,
      config?: C,
    ) => gameExec<T>(game, { command, bid, cardIndex, config }),
  };
}

/** API client for the Spades /spades/exec endpoint. */
export const spadesApi = createBidPlayApi<SpadesResponse, SpadesConfigInput>('spades');

/** Configuration options for Call Break game settings. */
export interface CallBreakConfigInput {
  cpuDifficulty?: number;
  maxRounds?: number;
}

/** API client for the Call Break /callbreak/exec endpoint. */
export const callBreakApi = createBidPlayApi<CallBreakResponse, CallBreakConfigInput>('callbreak');

/** Configuration options for Tarneeb game settings. */
export interface TarneebConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
  minBid?: number;
}

/**
 * API client for the Tarneeb /tarneeb/exec endpoint.
 *
 * Signature mirrors {@link twoTenJackApi}: `(command, arg1, cardIndex, config)`.
 * `arg1` is overloaded based on the command:
 *   - `command === 'bid'` → `arg1` is the bid value (0=pass, 7-13=bid).
 *   - `command === 'trump'` → `arg1` is the trump suit (1=♠ 2=♣ 3=♥ 4=♦).
 *   - otherwise `arg1` is ignored.
 */
export const tarneebApi = {
  exec: (
    command: 'reset' | 'bid' | 'trump' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    arg1?: number,
    cardIndex?: number,
    config?: TarneebConfigInput,
  ) => {
    const body: {
      command: string;
      bid?: number;
      trumpSuit?: number;
      cardIndex?: number;
      config?: TarneebConfigInput;
    } = { command };
    if (command === 'bid') body.bid = arg1;
    else if (command === 'trump') body.trumpSuit = arg1;
    if (cardIndex !== undefined) body.cardIndex = cardIndex;
    if (config) body.config = config;
    return gameExec<TarneebResponse>('tarneeb', body);
  },
};

/** Configuration options for Pitch (Setback) game settings. */
export interface PitchConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Pitch /pitch/exec endpoint. */
export const pitchApi = createBidPlayApi<PitchResponse, PitchConfigInput>('pitch');

/** Configuration options for Two Ten Jack game settings. */
export interface TwoTenJackConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Two Ten Jack /twotenjack/exec endpoint.
 *
 * Argument order mirrors {@link spadesApi}: command, trumpSuit, cardIndex, config.
 * This keeps compatibility with {@link useTrickGameBase} which invokes play as
 * `(command, undefined, cardIndex)`.
 */
export const twoTenJackApi = {
  exec: (
    command: 'reset' | 'declare' | 'play' | 'next' | 'nextround' | 'hint',
    trumpSuit?: number,
    cardIndex?: number,
    config?: TwoTenJackConfigInput,
  ) =>
    gameExec<TwoTenJackResponse>('twotenjack', {
      command,
      trumpSuit,
      cardIndex,
      config,
    }),
};

/** Configuration options for Oh Hell game settings. */
export interface OhHellConfigInput {
  cpuDifficulty?: number;
  maxHandSize?: number;
  scoringVariant?: number;
  roundDirection?: number;
}

/** API client for the Oh Hell /ohhell/exec endpoint. */
export const ohHellApi = createBidPlayApi<OhHellResponse, OhHellConfigInput>('ohhell');

/** Configuration options for Wizard game settings. */
export interface WizardConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Wizard /wizard/exec endpoint. */
export const wizardApi = createBidPlayApi<WizardResponse, WizardConfigInput>('wizard');

/** Configuration options for Ninety-Nine game settings. */
export interface NinetyNineConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** API client for the Ninety-Nine /ninetynine/exec endpoint. Bidding submits 3 bury-card indices. */
export const ninetyNineApi = {
  exec: (
    command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    buryIndices?: number[],
    cardIndex?: number,
    config?: NinetyNineConfigInput,
  ) => gameExec<NinetyNineResponse>('ninetynine', { command, buryIndices, cardIndex, config }),
};

/** Configuration options for Memory game settings. */
export interface MemoryConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Memory /memory/exec endpoint. */
export const memoryApi = {
  exec: (command: 'reset' | 'flip' | 'next' | 'log', position?: number, config?: MemoryConfigInput) =>
    gameExec<MemoryResponse>('memory', {
      command,
      position,
      config,
    }),
};

/**
 * Factory for solitaire-style move APIs whose request body is `{ command, from, to, n }`.
 *
 * Used by Canfield, FreeCell, Yukon, Scorpion, Accordion, FortyThieves, and
 * Calculation — every solitaire variant whose move endpoint takes only
 * source/target zones and an optional batch-undo count.
 *
 * `Cmd` is intentionally not defaulted: each call site declares the exact
 * command union its game accepts so invalid commands are rejected at compile
 * time instead of being silently widened to a broader shared union.
 */
function createSolitaireMoveApi<T, Zone, Cmd extends string>(game: string) {
  return {
    exec: (command: Cmd, from?: Zone, to?: Zone, n?: number) => gameExec<T>(game, { command, from, to, n }),
  };
}

/**
 * Factory for solitaire-style move APIs that also carry an optional `config`
 * object (Klondike, Spider). Body shape: `{ command, from, to, config, n }`.
 *
 * Like {@link createSolitaireMoveApi}, the `Cmd` generic is not defaulted —
 * each call site declares its exact command union.
 */
function createSolitaireMoveApiWithConfig<T, Zone, C, Cmd extends string>(game: string) {
  return {
    exec: (command: Cmd, from?: Zone, to?: Zone, config?: C, n?: number) =>
      gameExec<T>(game, { command, from, to, config, n }),
  };
}

/** Source or target zone for a Klondike card move. */
export interface KlondikeMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** Configuration options for Klondike game settings. */
export interface KlondikeConfigInput {
  drawCount?: number;
  scoringMode?: number;
}

/** API client for the Klondike /klondike/exec endpoint. */
export const klondikeApi = createSolitaireMoveApiWithConfig<
  KlondikeResponse,
  KlondikeMoveZone,
  KlondikeConfigInput,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('klondike');

/** Source or target zone for a Canfield card move. */
export interface CanfieldMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Canfield /canfield/exec endpoint. */
export const canfieldApi = createSolitaireMoveApi<
  CanfieldResponse,
  CanfieldMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('canfield');

/** Source or target zone for an Agnes Sorel card move. */
export interface AgnesMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Agnes Sorel /agnes/exec endpoint. */
export const agnesApi = createSolitaireMoveApi<
  AgnesResponse,
  AgnesMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n'
>('agnes');

/** Source or target zone for an Osmosis card move. */
export interface OsmosisMoveZone {
  zone: string;
  col?: number;
}

/** API client for the Osmosis /osmosis/exec endpoint. */
export const osmosisApi = createSolitaireMoveApi<
  OsmosisResponse,
  OsmosisMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('osmosis');

/** API client for the Bristol /bristol/exec endpoint. */
export const bristolApi = createSolitaireMoveApi<
  BristolResponse,
  BristolMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bristol');

/** API client for the La Belle Lucie /labellelucie/exec endpoint. */
export const labellelucieApi = createSolitaireMoveApi<
  LaBelleLucieResponse,
  number,
  'reset' | 'mf' | 'ff' | 'rd' | 'ac' | 'u' | 'undo_n' | 'giveup' | 'hint' | 'log'
>('labellelucie');

/** Commands accepted by the Simple Simon /simplesimon/exec endpoint. */
export type SimpleSimonCommand = 'reset' | 'm' | 'g' | 'u' | 'undo_n' | 'hint' | 'log';

/** API client for the Simple Simon /simplesimon/exec endpoint. */
export const simplesimonApi = {
  exec: (command: SimpleSimonCommand, opts?: { fromCol?: number; cardIndex?: number; toCol?: number; n?: number }) =>
    gameExec<SimpleSimonResponse>('simplesimon', {
      command,
      fromCol: opts?.fromCol,
      cardIndex: opts?.cardIndex,
      toCol: opts?.toCol,
      n: opts?.n,
    }),
};

/** Commands accepted by the Double Klondike /doubleklondike/exec endpoint. */
export type DoubleKlondikeCommand =
  | 'reset'
  | 'd'
  | 'mwt'
  | 'mwf'
  | 'mtt'
  | 'mtf'
  | 'g'
  | 'ac'
  | 'u'
  | 'undo_n'
  | 'hint'
  | 'log';

/** API client for the Double Klondike /doubleklondike/exec endpoint. */
export const doubleklondikeApi = {
  exec: (
    command: DoubleKlondikeCommand,
    opts?: { col?: number; fromCol?: number; cardIndex?: number; toCol?: number; n?: number },
  ) =>
    gameExec<DoubleKlondikeResponse>('doubleklondike', {
      command,
      col: opts?.col,
      fromCol: opts?.fromCol,
      cardIndex: opts?.cardIndex,
      toCol: opts?.toCol,
      n: opts?.n,
    }),
};

/** Commands accepted by the Black Hole /blackhole/exec endpoint. */
export type BlackHoleCommand = 'reset' | 'mb' | 'g' | 'u' | 'undo_n' | 'hint' | 'log';

/** API client for the Black Hole /blackhole/exec endpoint. */
export const blackholeApi = {
  exec: (command: BlackHoleCommand, opts?: { fan?: number; n?: number }) =>
    gameExec<BlackHoleResponse>('blackhole', {
      command,
      fan: opts?.fan,
      n: opts?.n,
    }),
};

/** Source or target zone for a FreeCell card move. */
export interface FreeCellMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the FreeCell /freecell/exec endpoint. */
export const freecellApi = createSolitaireMoveApi<
  FreeCellResponse,
  FreeCellMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('freecell');

/**
 * API client for the Baker's Game /bakersgame/exec endpoint. Baker's Game is the
 * same-suit FreeCell variant; it reuses the FreeCell wire shape.
 */
export const bakersgameApi = createSolitaireMoveApi<
  FreeCellResponse,
  FreeCellMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bakersgame');

/** Source or target zone for an Eight Off card move. */
export interface EightOffMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Eight Off /eightoff/exec endpoint. */
export const eightoffApi = createSolitaireMoveApi<
  EightOffResponse,
  EightOffMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('eightoff');

/** Source or target zone for a Penguin card move. */
export interface PenguinMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Penguin /penguin/exec endpoint. */
export const penguinApi = createSolitaireMoveApi<
  PenguinResponse,
  PenguinMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('penguin');

/** Source or target zone for a Seahaven Towers card move. */
export interface SeahavenTowersMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Seahaven Towers /seahaventowers/exec endpoint. */
export const seahaventowersApi = createSolitaireMoveApi<
  SeahavenTowersResponse,
  SeahavenTowersMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('seahaventowers');

/** Configuration options for Crazy Eights game settings. */
export interface CrazyEightsConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Crazy Eights /crazyeights/exec endpoint. */
export const crazyeightsApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'suit' | 'nextround',
    cardIndex?: number,
    suit?: number,
    config?: CrazyEightsConfigInput,
  ) =>
    gameExec<CrazyEightsResponse>('crazyeights', {
      command,
      cardIndex,
      suit,
      config,
    }),
};

/** Configuration options for Prší game settings. */
export interface PrsiConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Prší /prsi/exec endpoint. */
export const prsiApi = {
  exec: (command: 'reset' | 'play' | 'draw' | 'log', cardIndex?: number, config?: PrsiConfigInput) =>
    gameExec<PrsiResponse>('prsi', {
      command,
      cardIndex,
      config,
    }),
};

/** Configuration options for Page One game settings. */
export interface PageOneConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Page One /pageone/exec endpoint. */
export const pageoneApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'declare' | 'skip' | 'nextround',
    cardIndex?: number,
    config?: PageOneConfigInput,
  ) =>
    gameExec<PageOneResponse>('pageone', {
      command,
      cardIndex,
      config,
    }),
};

/** Configuration options for Gin Rummy game settings. */
export interface GinRummyConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Gin Rummy /ginrummy/exec endpoint. */
export const ginrummyApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'layoff' | 'nextround' | 'log',
    cardIndex?: number,
    config?: GinRummyConfigInput,
    cardIndices?: number[],
  ) =>
    gameExec<GinRummyResponse>('ginrummy', {
      command,
      cardIndex,
      cardIndices,
      config,
    }),
};

/** Configuration options for Indian Rummy game settings. */
export interface IndianRummyConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** API client for the Indian Rummy /indianrummy/exec endpoint. */
export const indianRummyApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'declare' | 'nextround' | 'log',
    cardIndex?: number,
    config?: IndianRummyConfigInput,
  ) =>
    gameExec<IndianRummyResponse>('indianrummy', {
      command,
      cardIndex,
      config,
    }),
};

/** Move parameters for a Machiavelli turn action (newmeld / layoff / play). */
export interface MachiavelliMoveParams {
  /** Full proposed table for the `play` power move (card refs by design + value). */
  tableMelds?: { design: number; value: number }[][];
  /** Hand-card indices for `newmeld` (or the cards added by `play`). */
  handIndices?: number[];
  /** Target table meld index for `layoff`. */
  meldIdx?: number;
  /** Hand-card index added to an existing meld for `layoff`. */
  handIndex?: number;
}

/** Configuration options for Machiavelli game settings. */
export interface MachiavelliConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** API client for the Machiavelli /machiavelli/exec endpoint. */
export const machiavelliApi = {
  exec: (
    command: 'reset' | 'draw' | 'play' | 'newmeld' | 'layoff' | 'nextround' | 'log',
    params?: MachiavelliMoveParams,
    config?: MachiavelliConfigInput,
  ) =>
    gameExec<MachiavelliResponse>('machiavelli', {
      command,
      ...(params ?? {}),
      config,
    }),
};

/** Configuration options for Panguingue (Pan) game settings. */
export interface PanConfigInput {
  playerCount?: number;
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Action parameters for a Panguingue (Pan) turn. */
export interface PanActionParams {
  /** Hand-card indices forming a new meld (set or rope). */
  cardIndices?: number[];
  /** Hand-card index to discard or lay off. */
  cardIndex?: number;
  /** Owning player id of the target meld for a layoff. */
  meldOwner?: number;
  /** Index of the target meld within the owner's laid melds for a layoff. */
  meldIdx?: number;
}

/** API client for the Panguingue (Pan) /pan/exec endpoint. */
export const panApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'layoff' | 'discard' | 'nextround' | 'log',
    params?: PanActionParams,
    config?: PanConfigInput,
  ) =>
    gameExec<PanResponse>('pan', {
      command,
      cardIndices: params?.cardIndices,
      cardIndex: params?.cardIndex,
      meldOwner: params?.meldOwner,
      meldIdx: params?.meldIdx,
      config,
    }),
};

/** Configuration options for Chinchón game settings. */
export interface ChinchonConfigInput {
  cpuDifficulty?: number;
  playerCount?: number;
  knockThreshold?: number;
  eliminationLimit?: number;
}

/** API client for the Chinchón /chinchon/exec endpoint. */
export const chinchonApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'layoff' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ChinchonConfigInput,
    cardIndices?: number[],
  ) =>
    gameExec<ChinchonResponse>('chinchon', {
      command,
      cardIndex,
      cardIndices,
      config,
    }),
};

/** Configuration options for Three Thirteen game settings. */
export interface ThreeThirteenConfigInput {
  cpuDifficulty?: number;
  playerCount?: number;
}

/** API client for the Three Thirteen /threethirteen/exec endpoint. */
export const threethirteenApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ThreeThirteenConfigInput,
  ) =>
    gameExec<ThreeThirteenResponse>('threethirteen', {
      command,
      cardIndex,
      config,
    }),
};

/** Configuration options for Conquian game settings. */
export interface ConquianConfigInput {
  cpuDifficulty?: number;
  targetWins?: number;
}

/** API client for the Conquian /conquian/exec endpoint. */
export const conquianApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'discard' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ConquianConfigInput,
    meldGroups?: number[][],
  ) =>
    gameExec<ConquianResponse>('conquian', {
      command,
      cardIndex,
      config,
      meldGroups,
    }),
};

/** Configuration options for Rummy 500 game settings. */
export interface Rummy500ConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Layoff parameters: meld owner, meld index, and card index in hand. */
export interface Rummy500LayoffInput {
  meldOwner: number;
  meldIdx: number;
  cardIndex: number;
}

/** API client for the Rummy 500 /rummy500/exec endpoint. */
export const rummy500Api = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'layoff' | 'discard' | 'nextround' | 'log',
    cardIndex?: number,
    config?: Rummy500ConfigInput,
    cardIndices?: number[],
    discardIdx?: number,
    layoff?: Rummy500LayoffInput,
  ) =>
    gameExec<Rummy500Response>('rummy500', {
      command,
      cardIndex: layoff?.cardIndex ?? cardIndex,
      cardIndices,
      discardIdx,
      meldOwner: layoff?.meldOwner,
      meldIdx: layoff?.meldIdx,
      config,
    }),
};

/** Configuration options for Tonk game settings. */
export interface TonkConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Tonk /tonk/exec endpoint. */
export const tonkApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: TonkConfigInput,
  ) =>
    gameExec<TonkResponse>('tonk', {
      command,
      cardIndex,
      config,
    }),
};

/** Configuration options for Thirty-One game settings. */
export interface ThirtyOneConfigInput {
  cpuDifficulty?: number;
  initialLives?: number;
}

/** API client for the Thirty-One /thirtyone/exec endpoint. */
export const thirtyoneApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'discard' | 'knock' | 'nextround' | 'log',
    cardIndex?: number,
    config?: ThirtyOneConfigInput,
  ) =>
    gameExec<ThirtyOneResponse>('thirtyone', {
      command,
      cardIndex,
      config,
    }),
};

/** Configuration options for Yaniv game settings. */
export interface YanivConfigInput {
  cpuDifficulty?: number;
  scoreLimit?: number;
}

/** API client for the Yaniv /yaniv/exec endpoint. */
export const yanivApi = {
  exec: (
    command: 'reset' | 'discard' | 'yaniv' | 'drawstock' | 'drawpickup' | 'nextround' | 'log',
    opts?: { cardIndices?: number[]; end?: number; config?: YanivConfigInput },
  ) =>
    gameExec<YanivResponse>('yaniv', {
      command,
      cardIndices: opts?.cardIndices,
      end: opts?.end,
      config: opts?.config,
    }),
};

/** Configuration options for Canasta game settings. */
export interface CanastaConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Canasta /canasta/exec endpoint. */
export const canastaApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: CanastaConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<CanastaResponse>('canasta', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};

/** Configuration options for Samba game settings. */
export interface SambaConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Samba /samba/exec endpoint. */
export const sambaApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: SambaConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<SambaResponse>('samba', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};

/** Configuration options for Hand and Foot game settings. */
export interface HandAndFootConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Hand and Foot /handandfoot/exec endpoint. */
export const handandfootApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: HandAndFootConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<HandAndFootResponse>('handandfoot', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};

/** Configuration options for Burraco game settings. */
export interface BurracoConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Burraco /burraco/exec endpoint. */
export const burracoApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'skipmeld' | 'discard' | 'goout' | 'nextround' | 'log',
    cardIndex?: number,
    config?: BurracoConfigInput,
    naturalPairIndices?: number[],
    meldGroups?: number[][],
  ) =>
    gameExec<BurracoResponse>('burraco', {
      command,
      cardIndex,
      config,
      naturalPairIndices,
      meldGroups,
    }),
};

/** Configuration options for Pinochle game settings. */
export interface PinochleConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Pinochle /pinochle/exec endpoint. */
export const pinochleApi = {
  exec: (
    command: 'reset' | 'bid' | 'pass' | 'trump' | 'meld' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: PinochleConfigInput,
    bidAmount?: number,
    suit?: number,
  ) =>
    gameExec<PinochleResponse>('pinochle', {
      command,
      cardIndex,
      config,
      bidAmount,
      suit,
    }),
};

/** Configuration options for Piquet game settings. */
export interface PiquetConfigInput {
  cpuDifficulty?: number;
  dealsPerPartie?: number;
}

/** API client for the Piquet /piquet/exec endpoint. */
export const piquetApi = {
  exec: (
    command: 'reset' | 'e' | 'y' | 'd' | 'p' | 'nd' | 'h' | 'log',
    cardIndex?: number,
    discardIndices?: number[],
    config?: PiquetConfigType,
  ) =>
    gameExec<PiquetResponse>('piquet', {
      command,
      cardIndex,
      discardIndices,
      config,
    }),
};

/** Configuration options for Cribbage game settings. */
export interface CribbageConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Cribbage /cribbage/exec endpoint. */
export const cribbageApi = {
  exec: (
    command: 'reset' | 'discard' | 'cut' | 'peg' | 'go' | 'shownext' | 'nextround' | 'log',
    cardIndex?: number,
    cardIndices?: number[],
    config?: CribbageConfigInput,
  ) =>
    gameExec<CribbageResponse>('cribbage', {
      command,
      cardIndex,
      cardIndices,
      config,
    }),
};

/** API client for the Baccarat /baccarat/exec endpoint. */
export const baccaratApi = {
  exec: (
    command: 'reset' | 'bet' | 'log' | 'clearhistory',
    amount?: number,
    betType?: number,
    playerPairBet?: number,
    bankerPairBet?: number,
  ) => gameExec<BaccaratResponse>('baccarat', { command, amount, betType, playerPairBet, bankerPairBet }),
};

/** API client for the Three Card Poker /threecard/exec endpoint. */
export const threecardApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, pairPlusBet?: number) =>
    gameExec<ThreeCardResponse>('threecard', { command, amount, pairPlusBet }),
};

/** API client for the Four Card Poker /fourcardpoker/exec endpoint. */
export const fourcardpokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'fold' | 'log',
    amount?: number,
    acesUpBet?: number,
    playMultiplier?: number,
  ) =>
    gameExec<FourCardPokerResponse>('fourcardpoker', {
      command,
      amount,
      acesUpBet,
      playMultiplier,
    }),
};

/** API client for the High Card Flush /highcardflush/exec endpoint. */
export const highcardflushApi = {
  exec: (
    command: 'reset' | 'bet' | 'raise' | 'fold' | 'log',
    amount?: number,
    flushBonusBet?: number,
    straightFlushBet?: number,
    multiplier?: number,
  ) =>
    gameExec<HighCardFlushResponse>('highcardflush', {
      command,
      amount,
      flushBonusBet,
      straightFlushBet,
      multiplier,
    }),
};

/** API client for the Caribbean Stud Poker /caribbeanstud/exec endpoint. */
export const caribbeanstudApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, jackpotBet?: number) =>
    gameExec<CaribbeanStudResponse>('caribbeanstud', { command, amount, jackpotBet }),
};

/** API client for the Oasis Poker /oasispoker/exec endpoint. */
export const oasispokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'exchange' | 'stand' | 'play' | 'fold' | 'log',
    amount?: number,
    jackpotBet?: number,
    indices?: number[],
  ) => gameExec<OasisPokerResponse>('oasispoker', { command, amount, jackpotBet, indices }),
};

/** API client for the Russian Poker /russianpoker/exec endpoint. */
export const russianpokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'exchange' | 'buy6th' | 'select' | 'play' | 'fold' | 'force' | 'decline' | 'log',
    amount?: number,
    indices?: number[],
    discardIndex?: number,
  ) => gameExec<RussianPokerResponse>('russianpoker', { command, amount, indices, discardIndex }),
};

/** API client for the Texas Hold'em Bonus Poker /texasholdembonus/exec endpoint. */
export const texasholdembonusApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'check' | 'raise' | 'log', amount?: number, bonusBet?: number) =>
    gameExec<TexasHoldemBonusResponse>('texasholdembonus', { command, amount, bonusBet }),
};

/** API client for the Casino Hold'em /casinoholdem/exec endpoint. */
export const casinoholdemApi = {
  exec: (command: 'reset' | 'bet' | 'call' | 'fold' | 'log', amount?: number, bonusBet?: number) =>
    gameExec<CasinoHoldemResponse>('casinoholdem', { command, amount, bonusBet }),
};

/** API client for the Ultimate Texas Hold'em /ultimatetexasholdem/exec endpoint. */
export const ultimatetexasholdemApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'check' | 'fold' | 'log',
    amount?: number,
    tripsBet?: number,
    multiplier?: number,
  ) => gameExec<UltimateTexasHoldemResponse>('ultimatetexasholdem', { command, amount, tripsBet, multiplier }),
};

/** API client for the Pai Gow Poker /paigow/exec endpoint. */
export const paigowApi = {
  exec: (command: 'reset' | 'bet' | 'set' | 'log', amount?: number, low0?: number, low1?: number) =>
    gameExec<PaiGowResponse>('paigow', { command, amount, low0, low1 }),
};

/** API client for the Chinese Poker /chinesepoker/exec endpoint. */
export const chinesepokerApi = {
  exec: (
    command: 'reset' | 'bet' | 'set' | 'log',
    amount?: number,
    frontIndices?: number[],
    middleIndices?: number[],
  ) => gameExec<ChinesePokerResponse>('chinesepoker', { command, amount, frontIndices, middleIndices }),
};

/** API client for the Six Card Golf /sixcardgolf/exec endpoint. */
export const sixcardgolfApi = {
  exec: (params: {
    command: string;
    position?: number;
    config?: { playerCount?: number; cpuDifficulty?: number; rounds?: number };
  }) => gameExec<SixCardGolfResponse>('sixcardgolf', params),
};

/** API client for the Dou Dizhu /doudizhu/exec endpoint. */
export const doudizhuApi = {
  exec: (params: { command: string; indices?: number[]; bidValue?: number; config?: { cpuDifficulty?: number } }) =>
    gameExec<DoudizhuResponse>('doudizhu', params),
};

/** API client for the Tichu /tichu/exec endpoint. */
export const tichuApi = {
  exec: (params: { command: string; indices?: number[]; declType?: number; config?: { cpuDifficulty?: number } }) =>
    gameExec<TichuResponse>('tichu', params),
};

/** API client for the Bourré /bourre/exec endpoint. */
export const bourreApi = {
  exec: (params: {
    command: string;
    decide?: boolean;
    indices?: number[];
    cardIndex?: number;
    config?: { cpuDifficulty?: number };
  }) => gameExec<BourreResponse>('bourre', params),
};

/** Configuration options for Sheepshead game settings. */
export interface SheepsheadConfigInput {
  cpuDifficulty?: number;
  baseChips?: number;
  startChips?: number;
  targetChips?: number;
}

/** Commands accepted by the Sheepshead /sheepshead/exec endpoint. */
export type SheepsheadCommand = 'reset' | 'pick' | 'bury' | 'call' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Sheepshead /sheepshead/exec endpoint.
 *
 * The multi-phase flow maps each command to its own body field:
 *   - `pick` → `{ pick: boolean }` (take or pass the blind)
 *   - `bury` → `{ buryIndices: number[] }` (picker buries 2 cards)
 *   - `call` → `{ callSuit: number }` (1=♠ 2=♣ 3=♥)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const sheepsheadApi = {
  exec: (
    command: SheepsheadCommand,
    opts?: {
      pick?: boolean;
      buryIndices?: number[];
      callSuit?: number;
      cardIndex?: number;
      config?: SheepsheadConfigInput;
    },
  ) =>
    gameExec<SheepsheadResponse>('sheepshead', {
      command,
      pick: opts?.pick,
      buryIndices: opts?.buryIndices,
      callSuit: opts?.callSuit,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Mus game settings. */
export interface MusConfigInput {
  cpuDifficulty?: number;
  targetAmarrakos?: number;
}

/** Commands accepted by the Mus /mus/exec endpoint. */
export type MusCommand = 'reset' | 'mus' | 'discard' | 'bet' | 'next' | 'hint' | 'log';

/**
 * API client for the Mus /mus/exec endpoint.
 *
 * Mus is a Basque vying (betting) game, so each command maps to its own body
 * field rather than a card-play action:
 *   - `mus` → `{ mus: boolean }` (true = call Mus / exchange, false = cut and bet)
 *   - `discard` → `{ discardIndices: number[] }` (cards to exchange; empty keeps all)
 *   - `bet` → `{ betAction: number, betAmount?: number }`
 *     (betAction: 0=paso 1=envido 2=ordago 3=quiero 4=noquiero)
 *   - `reset` → `{ config }`
 *   - `next` / `hint` / `log` carry no extra fields.
 */
export const musApi = {
  exec: (
    command: MusCommand,
    opts?: {
      mus?: boolean;
      discardIndices?: number[];
      betAction?: number;
      betAmount?: number;
      config?: MusConfigInput;
    },
  ) =>
    gameExec<MusResponse>('mus', {
      command,
      mus: opts?.mus,
      discardIndices: opts?.discardIndices,
      betAction: opts?.betAction,
      betAmount: opts?.betAmount,
      config: opts?.config,
    }),
};

/** Configuration options for Doppelkopf game settings. */
export interface DoppelkopfConfigInput {
  cpuDifficulty?: number;
  baseChips?: number;
  startChips?: number;
  targetChips?: number;
}

/** Commands accepted by the Doppelkopf /doppelkopf/exec endpoint. */
export type DoppelkopfCommand = 'reset' | 'play' | 'announce' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Doppelkopf /doppelkopf/exec endpoint.
 *
 * Doppelkopf is a plain trick-taking flow (no pick/bury/call). The only extra
 * action beyond playing a card is `announce` (Re/Kontra, first trick only):
 *   - `play` → `{ cardIndex: number }`
 *   - `announce` → no extra fields (declares Re or Kontra based on the human's team)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const doppelkopfApi = {
  exec: (
    command: DoppelkopfCommand,
    opts?: {
      cardIndex?: number;
      config?: DoppelkopfConfigInput;
    },
  ) =>
    gameExec<DoppelkopfResponse>('doppelkopf', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Sueca game settings. */
export interface SuecaConfigInput {
  cpuDifficulty?: number;
  targetGamePoints?: number;
}

/** Commands accepted by the Sueca /sueca/exec endpoint. */
export type SuecaCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Sueca /sueca/exec endpoint.
 *
 * Sueca is a Portuguese/Brazilian 4-player (2 vs 2) trump trick-taker. The only
 * play action is playing a card; there are no declarations.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const suecaApi = {
  exec: (
    command: SuecaCommand,
    opts?: {
      cardIndex?: number;
      config?: SuecaConfigInput;
    },
  ) =>
    gameExec<SuecaResponse>('sueca', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Klaverjas game settings. */
export interface KlaverjasConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Klaverjas /klaverjas/exec endpoint. */
export type KlaverjasCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Klaverjas /klaverjas/exec endpoint.
 *
 * Klaverjas is a Dutch 4-player (2 vs 2) trump trick-taker with Roem (run/marriage)
 * bonuses. The only play action is playing a card; there are no declarations.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const klaverjasApi = {
  exec: (
    command: KlaverjasCommand,
    opts?: {
      cardIndex?: number;
      config?: KlaverjasConfigInput;
    },
  ) =>
    gameExec<KlaverjasResponse>('klaverjas', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Manille game settings. */
export interface ManilleConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Manille /manille/exec endpoint. */
export type ManilleCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Manille /manille/exec endpoint.
 *
 * Manille is a French/Belgian 4-player (2 vs 2) trump trick-taker. The only
 * play action is playing a card (must follow suit / overtrump unless the
 * partner already holds the trick); there are no declarations and no Roem.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const manilleApi = {
  exec: (
    command: ManilleCommand,
    opts?: {
      cardIndex?: number;
      config?: ManilleConfigInput;
    },
  ) =>
    gameExec<ManilleResponse>('manille', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Sedma game settings. */
export interface SedmaConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Sedma /sedma/exec endpoint. */
export type SedmaCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Sedma /sedma/exec endpoint.
 *
 * Sedma is a Czech/Slovak 32-card no-trump capture trick-taker, 4 players in 2
 * teams. There is no trump suit and no follow obligation — any card is legal.
 * A card captures the trick if its rank equals the lead card's rank or it is a
 * 7 (wild); the last player to capture wins the trick.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const sedmaApi = {
  exec: (
    command: SedmaCommand,
    opts?: {
      cardIndex?: number;
      config?: SedmaConfigInput;
    },
  ) =>
    gameExec<SedmaResponse>('sedma', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Mariáš game settings. */
export interface MariasConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Mariáš /marias/exec endpoint. */
export type MariasCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Mariáš /marias/exec endpoint.
 *
 * Mariáš is a Czech/Slovak 3-player 32-card trump trick-taker. A rotating
 * Soloist plays alone against the 2 Defenders; trump is the Soloist's longest
 * suit (auto). The only play action is playing a card (must follow, trump when
 * void); there are no declarations.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const mariasApi = {
  exec: (
    command: MariasCommand,
    opts?: {
      cardIndex?: number;
      config?: MariasConfigInput;
    },
  ) =>
    gameExec<MariasResponse>('marias', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for King game settings. */
export interface KingConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the King /king/exec endpoint. */
export type KingCommand = 'reset' | 'contract' | 'play' | 'next' | 'hint' | 'log';

/**
 * API client for the King /king/exec endpoint.
 *
 * King is a 4-player 52-card compendium trick-avoidance game. Each match runs
 * exactly seven deals; the dealer of each deal selects one of seven unused
 * contracts and all four seats play thirteen must-follow tricks.
 *   - `contract` → `{ contract, trumpSuit }` (the dealer picks the deal's
 *     contract 0..6; `trumpSuit` is 1..4 for contract 6 "King (Trump)", else -1)
 *   - `play` → `{ handIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `hint` / `log` carry no extra fields.
 */
export const kingApi = {
  exec: (
    command: KingCommand,
    opts?: {
      contract?: number;
      trumpSuit?: number;
      handIndex?: number;
      config?: KingConfigInput;
    },
  ) =>
    gameExec<KingResponse>('king', {
      command,
      contract: opts?.contract,
      trumpSuit: opts?.trumpSuit,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Tysiąc (Thousand) game settings. */
export interface TysiacConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Tysiąc /tysiac/exec endpoint. */
export type TysiacCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tysiąc (Thousand) /tysiac/exec endpoint.
 *
 * Tysiąc is a Polish 3-player 24-card trump trick-taker with a Bid phase, a
 * Talon exchange phase, and marriage (K+Q) declarations during play.
 *   - `bid` → `{ raise: boolean }` (raise=true means +10, false means pass)
 *   - `discard` → `{ cardIndex }` (talon exchange: the human Declarer gives one
 *     card to an opponent; called once per opponent — twice total)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tysiacApi = {
  exec: (
    command: TysiacCommand,
    opts?: {
      cardIndex?: number;
      raise?: boolean;
      config?: TysiacConfigInput;
    },
  ) =>
    gameExec<TysiacResponse>('tysiac', {
      command,
      cardIndex: opts?.cardIndex,
      raise: opts?.raise,
      config: opts?.config,
    }),
};

/** Configuration options for Calabresella (Terziglio) game settings. */
export interface CalabresellaConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Calabresella /calabresella/exec endpoint. */
export type CalabresellaCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Calabresella (Terziglio) /calabresella/exec endpoint.
 *
 * Calabresella is a Calabrian/Italian 3-player 40-card (Tressette-family)
 * trick-taker with a Bid phase, a monte exchange (discard four) phase, and no
 * trump. One Soloist plays alone against the coalition of the other two.
 *   - `bid` → `{ bid: number }` (0=pass, 1=chiamo, 2=solo)
 *   - `discard` → `{ cardIndex }` (monte exchange: the Soloist discards one card
 *     per call, four times, from 16 down to 12)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const calabresellaApi = {
  exec: (
    command: CalabresellaCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      config?: CalabresellaConfigInput;
    },
  ) =>
    gameExec<CalabresellaResponse>('calabresella', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      config: opts?.config,
    }),
};

/** Configuration options for Ombre (Hombre) game settings. */
export interface OmbreConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Ombre /ombre/exec endpoint. */
export type OmbreCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Ombre (Hombre) /ombre/exec endpoint.
 *
 * Ombre is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck. A Bid phase (pass / entrar / solo) plus a chosen trump suit decides the
 * Ombre, who then plays alone against the coalition of the other two.
 *   - `bid` → `{ bid, trumpSuit? }` (bid 0=pass, 1=entrar, 2=solo; trumpSuit
 *     1=♠ 2=♣ 3=♥ 4=♦, sent with a winning entrar/solo)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const ombreApi = {
  exec: (
    command: OmbreCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      trumpSuit?: number;
      config?: OmbreConfigInput;
    },
  ) =>
    gameExec<OmbreResponse>('ombre', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};

/** Configuration options for Ulti (Ultimo) game settings. */
export interface UltiConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Ulti /ulti/exec endpoint. */
export type UltiCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Ulti (Ultimo) /ulti/exec endpoint.
 *
 * Ulti is a 3-player Hungarian contract trick-taker on a 32-card deck. The human
 * (seat 0) is always the declarer versus a 2-CPU defending coalition.
 *   - `bid` → `{ contract, trumpSuit? }` (contract 'party'|'betli'|'durchmarsch';
 *     trumpSuit 1=♠ 2=♣ 3=♥ 4=♦, meaningful only for a Party contract)
 *   - `discard` → `{ cardIndices }` (the 2 talon cards to discard)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const ultiApi = {
  exec: (
    command: UltiCommand,
    opts?: {
      contract?: string;
      trumpSuit?: number;
      cardIndices?: number[];
      cardIndex?: number;
      config?: UltiConfigInput;
    },
  ) =>
    gameExec<UltiResponse>('ulti', {
      command,
      contract: opts?.contract,
      trumpSuit: opts?.trumpSuit,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Scarto (スカルト) game settings. */
export interface ScartoConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the Scarto /scarto/exec endpoint. */
export type ScartoCommand = 'reset' | 'scarto' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Scarto (スカルト) /scarto/exec endpoint.
 *
 * Scarto is a 3-player Italian tarocchi trick-taker on the 78-card tarot deck.
 * The human is seat 0. There is no bidding, chien, or partnership.
 *   - `scarto` → `{ cardIndices }` (the 3 low pip cards the dealer buries)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const scartoApi = {
  exec: (
    command: ScartoCommand,
    opts?: {
      cardIndices?: number[];
      cardIndex?: number;
      config?: ScartoConfigInput;
    },
  ) =>
    gameExec<ScartoResponse>('scarto', {
      command,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for French Tarot (フレンチタロット) game settings. */
export interface FrenchTarotConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the French Tarot /frenchtarot/exec endpoint. */
export type FrenchTarotCommand = 'reset' | 'bid' | 'pass' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the French Tarot (フレンチタロット) /frenchtarot/exec endpoint.
 *
 * French Tarot is a 4-player trick-taker on the 78-card tarot deck. The human is
 * seat 0.
 *   - `bid` → `{ bid }` (contract string 'petite'|'garde'|'gardesans'|'gardecontre')
 *   - `pass` → carries no extra fields (pass the auction)
 *   - `discard` → `{ cardIndices }` (the 6 écart cards to bury; Petite/Garde only)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const frenchtarotApi = {
  exec: (
    command: FrenchTarotCommand,
    opts?: {
      bid?: string;
      cardIndices?: number[];
      cardIndex?: number;
      config?: FrenchTarotConfigInput;
    },
  ) =>
    gameExec<FrenchTarotResponse>('frenchtarot', {
      command,
      bid: opts?.bid,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Königrufen (ケーニッヒルーフェン) game settings. */
export interface KoenigrufenConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the Königrufen /koenigrufen/exec endpoint. */
export type KoenigrufenCommand =
  | 'reset'
  | 'bid'
  | 'pass'
  | 'callking'
  | 'discard'
  | 'play'
  | 'next'
  | 'nextround'
  | 'hint'
  | 'log';

/**
 * API client for the Königrufen (ケーニッヒルーフェン) /koenigrufen/exec endpoint.
 *
 * Königrufen is a 4-player tarock trick-taker on the 54-card tarock deck. The
 * human is seat 0.
 *   - `bid` → `{ bid }` (contract string 'rufer')
 *   - `pass` → carries no extra fields (pass the auction)
 *   - `callking` → `{ callSuit }` (1-4: the King suit the declarer calls)
 *   - `discard` → `{ cardIndices }` (the 6 talon cards to bury)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const koenigrufenApi = {
  exec: (
    command: KoenigrufenCommand,
    opts?: {
      bid?: string;
      callSuit?: number;
      cardIndices?: number[];
      cardIndex?: number;
      config?: KoenigrufenConfigInput;
    },
  ) =>
    gameExec<KoenigrufenResponse>('koenigrufen', {
      command,
      bid: opts?.bid,
      callSuit: opts?.callSuit,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Cego (チェゴ) game settings. */
export interface CegoConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the Cego /cego/exec endpoint. */
export type CegoCommand =
  | 'reset'
  | 'bid'
  | 'pass'
  | 'contract'
  | 'discard'
  | 'play'
  | 'next'
  | 'nextround'
  | 'hint'
  | 'log';

/**
 * API client for the Cego (チェゴ) /cego/exec endpoint.
 *
 * Cego is a 4-player Baden tarock trick-taker on the 54-card tarock deck. The
 * human is seat 0.
 *   - `bid` → `{ bid }` (bid string 'play')
 *   - `pass` → carries no extra fields (pass the auction)
 *   - `contract` → `{ contract }` ('cego' or 'handspiel')
 *   - `discard` → `{ cardIndices }` (the single card to KEEP in the Cego exchange)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const cegoApi = {
  exec: (
    command: CegoCommand,
    opts?: {
      bid?: string;
      contract?: string;
      cardIndices?: number[];
      cardIndex?: number;
      config?: CegoConfigInput;
    },
  ) =>
    gameExec<CegoResponse>('cego', {
      command,
      bid: opts?.bid,
      contract: opts?.contract,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Rook (ルーク) game settings. */
export interface RookConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Rook /rook/exec endpoint. */
export type RookCommand = 'reset' | 'bid' | 'pass' | 'exchange' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Rook (ルーク) /rook/exec endpoint.
 *
 * Rook is a 4-player, 2-team point-trick game on a special 57-card deck (four
 * colors ×1–14 plus the Rook bird). The human is seat 0.
 *   - `bid` → `{ bid }` (a numeric point bid, 70–120 in steps of 5)
 *   - `pass` → carries no extra fields
 *   - `exchange` → `{ discardIndices, trumpColor }` (discard 5 nest cards and
 *     declare a trump color: 1=red 2=gold 3=green 4=black)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const rookApi = {
  exec: (
    command: RookCommand,
    opts?: {
      bid?: number;
      discardIndices?: number[];
      trumpColor?: number;
      cardIndex?: number;
      config?: RookConfigInput;
    },
  ) =>
    gameExec<RookResponse>('rook', {
      command,
      bid: opts?.bid,
      discardIndices: opts?.discardIndices,
      trumpColor: opts?.trumpColor,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Cinch (Double Pedro) game settings. */
export interface CinchConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Commands accepted by the Cinch /cinch/exec endpoint. */
export type CinchCommand = 'reset' | 'bid' | 'trump' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Cinch (Double Pedro / High Five) /cinch/exec endpoint.
 *
 * Cinch is a 4-player 52-card All-Fours/Pitch-family bidding trick-taker. A Bid
 * phase (0=pass, 1-14) decides the bidder, who then names trump and leads.
 *   - `bid` → `{ bid: number }` (0=pass, 1-14)
 *   - `trump` → `{ trumpSuit: number }` (1=♠ 2=♣ 3=♥ 4=♦)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const cinchApi = {
  exec: (
    command: CinchCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      trumpSuit?: number;
      config?: CinchConfigInput;
    },
  ) =>
    gameExec<CinchResponse>('cinch', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};

/** Configuration options for Loo (Lanterloo) game settings. */
export interface LooConfigInput {
  cpuDifficulty?: number;
  ante?: number;
}

/** Commands accepted by the Loo /loo/exec endpoint. */
export type LooCommand = 'reset' | 'decide' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Loo (Lanterloo) /loo/exec endpoint.
 *
 * Loo is a 4-player 52-card pot-based gambling trick-taker. Trump is set from the
 * turn-up card (no bidding, no trump naming). Each player decides play or pass.
 *   - `decide` → `{ play: boolean }` (true=play, false=pass)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const looApi = {
  exec: (
    command: LooCommand,
    opts?: {
      cardIndex?: number;
      play?: boolean;
      config?: LooConfigInput;
    },
  ) =>
    gameExec<LooResponse>('loo', {
      command,
      cardIndex: opts?.cardIndex,
      play: opts?.play,
      config: opts?.config,
    }),
};

/** Configuration options for Basra (Bastra) game settings (CPU difficulty only). */
export interface BasraConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Basra /basra/exec endpoint. */
export type BasraCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Basra (Bastra) /basra/exec endpoint.
 *
 * Basra is a 4-player 52-card fishing/capture game.
 *   - `play` → `{ cardIndex, tableIndices? }` (tableIndices = table cards to
 *     capture; omit to trail, a Jack always sweeps)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const basraApi = {
  exec: (
    command: BasraCommand,
    opts?: {
      cardIndex?: number;
      tableIndices?: number[];
      config?: BasraConfigInput;
    },
  ) =>
    gameExec<BasraResponse>('basra', {
      command,
      cardIndex: opts?.cardIndex,
      tableIndices: opts?.tableIndices,
      config: opts?.config,
    }),
};

/** Configuration options for Hachi-Hachi (八八) game settings. */
export interface HachiHachiConfigInput {
  cpuDifficulty?: number;
  /** Number of rounds (deals) played before the match is settled. */
  targetRounds?: number;
}

/** Commands accepted by the Hachi-Hachi /hachihachi/exec endpoint. */
export type HachiHachiCommand = 'reset' | 'play' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Hachi-Hachi (八八) /hachihachi/exec endpoint.
 *
 * Hachi-Hachi is a 3-player hanafuda capture game with card-point scoring.
 *   - `play` → `{ cardIndex, fieldIndex? }` (fieldIndex disambiguates a 2-way
 *     field match; omit when there is at most one match)
 *   - `nextround` → deal the next round
 *   - `reset` → `{ config }`
 *   - `hint` / `log` carry no extra fields.
 */
export const hachihachiApi = {
  exec: (
    command: HachiHachiCommand,
    opts?: {
      cardIndex?: number;
      fieldIndex?: number;
      config?: HachiHachiConfigInput;
    },
  ) =>
    gameExec<HachiHachiResponse>('hachihachi', {
      command,
      cardIndex: opts?.cardIndex,
      fieldIndex: opts?.fieldIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Koi-Koi (こいこい) game settings. */
export interface KoiKoiConfigInput {
  cpuDifficulty?: number;
  /** Target cumulative score that ends the match. */
  targetScore?: number;
}

/** Commands accepted by the Koi-Koi /koikoi/exec endpoint. */
export type KoiKoiCommand = 'reset' | 'play' | 'koikoi' | 'stop' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Koi-Koi (こいこい) /koikoi/exec endpoint.
 *
 * Koi-Koi is a 2-player hanafuda capture game with yaku scoring.
 *   - `play` → `{ cardIndex, fieldIndex? }` (fieldIndex disambiguates a 2-way
 *     field match; omit when there is at most one match)
 *   - `koikoi` → continue the round (double the stakes) after completing a yaku
 *   - `stop` → shobu: stop and score the completed yaku
 *   - `nextround` → deal the next round
 *   - `reset` → `{ config }`
 *   - `hint` / `log` carry no extra fields.
 */
export const koikoiApi = {
  exec: (
    command: KoiKoiCommand,
    opts?: {
      cardIndex?: number;
      fieldIndex?: number;
      config?: KoiKoiConfigInput;
    },
  ) =>
    gameExec<KoiKoiResponse>('koikoi', {
      command,
      cardIndex: opts?.cardIndex,
      fieldIndex: opts?.fieldIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Go-Stop (Godori / ゴーストップ) game settings. */
export interface GoStopConfigInput {
  cpuDifficulty?: number;
  /** Target cumulative score that ends the match. */
  targetScore?: number;
}

/** Commands accepted by the Go-Stop /gostop/exec endpoint. */
export type GoStopCommand = 'reset' | 'play' | 'go' | 'stop' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Go-Stop (Godori / ゴーストップ) /gostop/exec endpoint.
 *
 * Go-Stop is the Korean sibling of Koi-Koi, a 2-player hanafuda capture game
 * with a Korean scoring breakdown (gwang/godori/tti/yeol/pi) plus Go/Stop.
 *   - `play` → `{ cardIndex, fieldIndex? }` (fieldIndex disambiguates a 2-way
 *     field match; omit when there is at most one match)
 *   - `go` → continue the round after reaching the target score
 *   - `stop` → bank the points and end the round
 *   - `nextround` → deal the next round
 *   - `reset` → `{ config }`
 *   - `hint` / `log` carry no extra fields.
 */
export const gostopApi = {
  exec: (
    command: GoStopCommand,
    opts?: {
      cardIndex?: number;
      fieldIndex?: number;
      config?: GoStopConfigInput;
    },
  ) =>
    gameExec<GoStopResponse>('gostop', {
      command,
      cardIndex: opts?.cardIndex,
      fieldIndex: opts?.fieldIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Tablanet (Tablić) game settings (CPU difficulty only). */
export interface TablanetConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Tablanet /tablanet/exec endpoint. */
export type TablanetCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tablanet (Tablić) /tablanet/exec endpoint.
 *
 * Tablanet is a 4-player 52-card fishing/capture game.
 *   - `play` → `{ cardIndex, tableIndices? }` (tableIndices = table cards to
 *     capture; omit to trail, a Jack always sweeps)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tablanetApi = {
  exec: (
    command: TablanetCommand,
    opts?: {
      cardIndex?: number;
      tableIndices?: number[];
      config?: TablanetConfigInput;
    },
  ) =>
    gameExec<TablanetResponse>('tablanet', {
      command,
      cardIndex: opts?.cardIndex,
      tableIndices: opts?.tableIndices,
      config: opts?.config,
    }),
};

/** Configuration options for Knockout Whist game settings (CPU difficulty only — no target points). */
export interface KnockoutWhistConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Knockout Whist /knockoutwhist/exec endpoint. */
export type KnockoutWhistCommand = 'reset' | 'play' | 'selecttrump' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Knockout Whist /knockoutwhist/exec endpoint.
 *
 * Knockout Whist is a British play-only survival trick-taker for 4 players on a
 * 52-card deck. Each round deals one fewer card; the previous round's winner's
 * longest suit becomes trump (auto). Must-follow, Ace-high. A player who wins
 * zero tricks in a round must spend a Dogbone token to survive, or is
 * eliminated. Last player standing wins.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const knockoutWhistApi = {
  exec: (
    command: KnockoutWhistCommand,
    opts?: {
      cardIndex?: number;
      trumpSuit?: number;
      config?: KnockoutWhistConfigInput;
    },
  ) =>
    gameExec<KnockoutWhistResponse>('knockoutwhist', {
      command,
      cardIndex: opts?.cardIndex,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};

/** Configuration options for Spoil Five game settings (CPU difficulty only — target points are fixed server-side). */
export interface SpoilFiveConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Spoil Five /spoilfive/exec endpoint. */
export type SpoilFiveCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Spoil Five /spoilfive/exec endpoint.
 *
 * Spoil Five (Maw) is an Irish play-only trick-taker for 5 players on a 52-card
 * deck (5 cards each). Trump is a turned-up card; the trump 5, trump J, and ♥A
 * are the fixed top trumps and may be held back (Reneging). The first player to
 * win 3 of the 5 tricks takes the pot; otherwise the round is a Spoil (流局) and
 * the pot carries over. First player to targetPoints wins the match.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const spoilFiveApi = {
  exec: (
    command: SpoilFiveCommand,
    opts?: {
      cardIndex?: number;
      config?: SpoilFiveConfigInput;
    },
  ) =>
    gameExec<SpoilFiveResponse>('spoilfive', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Solo Whist game settings. */
export interface SoloWhistConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Solo Whist /solowhist/exec endpoint. */
export type SoloWhistCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Solo Whist /solowhist/exec endpoint.
 *
 * Solo Whist is a British 4-player 52-card trick-taker with a bidding phase.
 * Each player bids once (Pass/Solo/Misère/Abundance); the highest bidder is the
 * declarer who plays alone against the other 3 defenders.
 *   - `bid` → `{ bid: number }` (0=Pass 1=Solo 2=Misère 3=Abundance)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const soloWhistApi = {
  exec: (
    command: SoloWhistCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: SoloWhistConfigInput;
    },
  ) =>
    gameExec<SoloWhistResponse>('solowhist', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Auction Forty-Fives game settings. */
export interface FortyFivesConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Auction Forty-Fives /fortyfives/exec endpoint. */
export type FortyFivesCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Auction Forty-Fives /fortyfives/exec endpoint.
 *
 * Auction Forty-Fives is an Irish/Canadian 4-player, 2-team trick-taker with a
 * bidding phase. Players bid Pass/15/20/25 (Jink); the highest bidder's team
 * declares trump and plays five tricks.
 *   - `bid` → `{ bid: number }` (0=Pass 15 20 25)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const fortyFivesApi = {
  exec: (
    command: FortyFivesCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: FortyFivesConfigInput;
    },
  ) =>
    gameExec<FortyFivesResponse>('fortyfives', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Twenty-Nine (29) game settings. */
export interface TwentyNineConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Twenty-Nine (29) /twentynine/exec endpoint. */
export type TwentyNineCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Twenty-Nine (29) /twentynine/exec endpoint.
 *
 * Twenty-Nine is an Indian/Bangladeshi 4-player, 2-team trick-taker with a
 * bidding phase and a hidden trump. Players bid Pass/16/20/24/28; the highest
 * bidder's team picks a hidden trump suit (revealed only mid-play) and plays
 * eight tricks.
 *   - `bid` → `{ bid: number }` (0=Pass 16 20 24 28)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const twentyNineApi = {
  exec: (
    command: TwentyNineCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: TwentyNineConfigInput;
    },
  ) =>
    gameExec<TwentyNineResponse>('twentynine', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Court Piece (Rang) game settings. */
export interface CourtPieceConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Commands accepted by the Court Piece (Rang) /courtpiece/exec endpoint. */
export type CourtPieceCommand = 'reset' | 'trump' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Court Piece (Rang) /courtpiece/exec endpoint.
 *
 * Court Piece is a 4-player, 2-team (seats 0&2 vs 1&3) trick-taker with no
 * numeric bidding. The caller (Hakim) peeks at 5 cards and declares a trump
 * suit; the teams then play 13 tricks. A team taking 7+ tricks wins the round
 * (Sar = +1 point); sweeps and consecutive wins add a Court bonus (+2). The
 * first team to reach the point limit (default 7) wins.
 *   - `trump` → `{ trumpSuit: number }` (1=♠ 2=♣ 3=♥ 4=♦)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const courtPieceApi = {
  exec: (
    command: CourtPieceCommand,
    opts?: {
      trumpSuit?: number;
      cardIndex?: number;
      config?: CourtPieceConfigInput;
    },
  ) =>
    gameExec<CourtPieceResponse>('courtpiece', {
      command,
      trumpSuit: opts?.trumpSuit,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Bezique game settings. */
export interface BeziqueConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Bezique /bezique/exec endpoint. */
export type BeziqueCommand = 'reset' | 'play' | 'meld' | 'skip' | 'next' | 'hint' | 'log' | 'config';

/**
 * API client for the Bezique /bezique/exec endpoint.
 *
 * Bezique is a 2-player ancestor of Pinochle played with a 64-card deck. In
 * phase 1 (while stock remains) the trick winner may declare ONE meld for
 * points (marriage 20 / royal marriage 40 / Bezique 40 / four aces 100 / kings
 * 80 / queens 60 / jacks 40) or skip; then both draw. When the stock empties,
 * phase 2 is strict must-follow with no melds and a +10 last-trick bonus.
 * Scores accumulate across deals to a target (default 1000).
 *   - `play` → `{ cardIndex: number }`
 *   - `meld` → `{ meldIndex: number }`
 *   - `reset` / `config` → `{ config }`
 *   - `skip` / `next` / `hint` / `log` carry no extra fields.
 */
export const beziqueApi = {
  exec: (
    command: BeziqueCommand,
    opts?: {
      cardIndex?: number;
      meldIndex?: number;
      config?: BeziqueConfigInput;
    },
  ) =>
    gameExec<BeziqueResponse>('bezique', {
      command,
      cardIndex: opts?.cardIndex,
      meldIndex: opts?.meldIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Écarté game settings. */
export interface EcarteConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Écarté /ecarte/exec endpoint. */
export type EcarteCommand =
  | 'reset'
  | 'propose'
  | 'stand'
  | 'respond'
  | 'discard'
  | 'play'
  | 'next'
  | 'hint'
  | 'log'
  | 'config';

/**
 * API client for the Écarté /ecarte/exec endpoint.
 *
 * Écarté is a 2-player French 32-card trick game with an Exchange phase. The
 * elder (non-dealer) chooses Propose or Stand; if proposed, the dealer Accepts
 * or Refuses; on accept, each player discards any number of cards and draws
 * replacements, then the elder decides again (until the stock empties). Play is
 * 5 strict must-follow tricks (rank K>Q>J>A>10>9>8>7). Scores accumulate to a
 * target (default 5).
 *   - `respond` → `{ accept: boolean }`
 *   - `discard` → `{ discardIndices: number[] }`
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` / `config` → `{ config }`
 *   - `propose` / `stand` / `next` / `hint` / `log` carry no extra fields.
 */
export const ecarteApi = {
  exec: (
    command: EcarteCommand,
    opts?: {
      accept?: boolean;
      cardIndex?: number;
      discardIndices?: number[];
      config?: EcarteConfigInput;
    },
  ) =>
    gameExec<EcarteResponse>('ecarte', {
      command,
      accept: opts?.accept,
      cardIndex: opts?.cardIndex,
      discardIndices: opts?.discardIndices,
      config: opts?.config,
    }),
};

/** Configuration options for Three Card Brag game settings. */
export interface ThreeCardBragConfigInput {
  cpuDifficulty?: number;
  ante?: number;
  startingChips?: number;
}

/** Commands accepted by the Three Card Brag /threecardbrag/exec endpoint. */
export type ThreeCardBragCommand =
  | 'reset'
  | 'see'
  | 'bet'
  | 'raise'
  | 'fold'
  | 'show'
  | 'next'
  | 'hint'
  | 'log'
  | 'config';

/**
 * API client for the Three Card Brag /threecardbrag/exec endpoint.
 *
 * Three Card Brag is a 4-player British vying game (poker ancestor) with chips
 * and a pot. On the human's turn: `see` (reveal, Blind→Seen), `bet` (call the
 * stake), `raise` (with `raiseStake`), `fold`, or `show` (when allowed). `next`
 * advances to the following deal; `reset` / `config` apply the config.
 *   - `raise` → `{ raiseStake: number }`
 *   - `reset` / `config` → `{ config }`
 *   - `see` / `bet` / `fold` / `show` / `next` / `hint` / `log` carry no extra fields.
 */
export const threeCardBragApi = {
  exec: (
    command: ThreeCardBragCommand,
    opts?: {
      raiseStake?: number;
      config?: ThreeCardBragConfigInput;
    },
  ) =>
    gameExec<ThreeCardBragResponse>('threecardbrag', {
      command,
      raiseStake: opts?.raiseStake,
      config: opts?.config,
    }),
};

/** Configuration options for Teen Patti game settings. */
export interface TeenPattiConfigInput {
  cpuDifficulty?: number;
  ante?: number;
  startingChips?: number;
}

/** Commands accepted by the Teen Patti /teenpatti/exec endpoint. */
export type TeenPattiCommand =
  | 'reset'
  | 'see'
  | 'bet'
  | 'raise'
  | 'fold'
  | 'show'
  | 'sideshow'
  | 'respond'
  | 'next'
  | 'hint'
  | 'log'
  | 'config';

/**
 * API client for the Teen Patti /teenpatti/exec endpoint.
 *
 * Teen Patti is the Indian variant of Three Card Brag — a 4-player vying game
 * with chips and a pot. On the human's turn: `see` (reveal, Blind→Seen), `bet`
 * (call the stake), `raise` (with `raiseStake`), `fold`, `show` (when allowed),
 * or `sideshow` (request a private hand comparison with the previous Seen
 * player). When a Side Show is requested of the human, `respond` (with
 * `accept`) accepts or declines it. `next` advances to the following deal;
 * `reset` / `config` apply the config.
 *   - `raise` → `{ raiseStake: number }`
 *   - `respond` → `{ accept: boolean }`
 *   - `reset` / `config` → `{ config }`
 *   - `see` / `bet` / `fold` / `show` / `sideshow` / `next` / `hint` / `log` carry no extra fields.
 */
export const teenPattiApi = {
  exec: (
    command: TeenPattiCommand,
    opts?: {
      raiseStake?: number;
      accept?: boolean;
      config?: TeenPattiConfigInput;
    },
  ) =>
    gameExec<TeenPattiResponse>('teenpatti', {
      command,
      raiseStake: opts?.raiseStake,
      accept: opts?.accept,
      config: opts?.config,
    }),
};

/** Configuration options for Spoons game settings. */
export interface SpoonsConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Spoons /spoons/exec endpoint. */
export type SpoonsCommand = 'reset' | 'pass' | 'grab' | 'next' | 'log';

/**
 * API client for the Spoons /spoons/exec endpoint.
 *
 * Spoons is a 4-player pass-and-grab speed game. On the Pass phase the human
 * picks one of their four cards to pass to the next player (`pass` →
 * `{ cardIndex }`). When someone collects four of a kind the Grab window opens;
 * everyone races to `grab` a spoon — the one who misses out gains a letter
 * (S-P-O-O-N-S). `next` advances to the following round; `reset` applies the
 * config (CPU difficulty); `log` fetches the action log.
 *   - `pass` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `grab` / `next` / `log` carry no extra fields.
 */
export const spoonsApi = {
  exec: (
    command: SpoonsCommand,
    opts?: {
      cardIndex?: number;
      config?: SpoonsConfigInput;
    },
  ) =>
    gameExec<SpoonsResponse>('spoons', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Kemps game settings. */
export interface KempsConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Kemps /kemps/exec endpoint. */
export type KempsCommand = 'reset' | 'swap' | 'pass' | 'signal' | 'kemps' | 'counter' | 'next' | 'log';

/**
 * API client for the Kemps /kemps/exec endpoint.
 *
 * Kemps is a 4-player, 2-team matching game. On the Exchange phase the human
 * swaps one hand card for a field card (`swap` → `{ handIndex, fieldIndex }`)
 * or skips with `pass`. The human sets a secret signal type with `signal` →
 * `{ signalType }` (0=Sound, 1=Blink). When a team completes four of a kind the
 * Declare window opens: `kemps` declares "Kemps!" and `counter` →
 * `{ targetSeat }` declares "Counter-Kemps!" against an opponent seat. `next`
 * advances to the following round; `reset` applies the config; `log` fetches
 * the action log.
 *   - `swap` → `{ handIndex: number, fieldIndex: number }`
 *   - `signal` → `{ signalType: number }`
 *   - `counter` → `{ targetSeat: number }`
 *   - `reset` → `{ config }`
 *   - `pass` / `kemps` / `next` / `log` carry no extra fields.
 */
export const kempsApi = {
  exec: (
    command: KempsCommand,
    opts?: {
      handIndex?: number;
      fieldIndex?: number;
      signalType?: number;
      targetSeat?: number;
      config?: KempsConfigInput;
    },
  ) =>
    gameExec<KempsResponse>('kemps', {
      command,
      handIndex: opts?.handIndex,
      fieldIndex: opts?.fieldIndex,
      signalType: opts?.signalType,
      targetSeat: opts?.targetSeat,
      config: opts?.config,
    }),
};

/** Configuration options for Cuckoo game settings. */
export interface CuckooConfigInput {
  cpuDifficulty?: number;
  initialLives?: number;
}

/** Commands accepted by the Cuckoo /cuckoo/exec endpoint. */
export type CuckooCommand = 'reset' | 'keep' | 'swap' | 'refuse' | 'accept' | 'nextround' | 'log';

/**
 * API client for the Cuckoo /cuckoo/exec endpoint.
 *
 * Cuckoo (a.k.a. Chase the Ace / Ranter-Go-Round) is a 4-player life-survival
 * game. On your turn you `keep` your card or `swap` it with your neighbour (the
 * dealer swaps with the stock). When you hold a King and someone tries to swap
 * into you, `refuse` reveals the King to block it or `accept` allows the swap.
 * `nextround` advances after the lowest card loses a life; `reset` applies the
 * config (CPU difficulty, initial lives); `log` fetches the action log. None of
 * the play commands carry extra fields — only `reset` takes a `config`.
 */
export const cuckooApi = {
  exec: (command: CuckooCommand, opts?: { config?: CuckooConfigInput }) =>
    gameExec<CuckooResponse>('cuckoo', {
      command,
      config: opts?.config,
    }),
};

/** Configuration options for Pişti game settings. */
export interface PishtiConfigInput {
  playerCnt?: number;
  cpuDifficulty?: number;
}

/** Commands accepted by the Pişti /pishti/exec endpoint. */
export type PishtiCommand = 'reset' | 'play' | 'next' | 'log';

/**
 * API client for the Pişti /pishti/exec endpoint.
 *
 * Pişti is a Turkish 2–4 player capture (fishing) game. On your turn you `play`
 * a hand card onto the central pile (`play` → `{ handIndex }`); matching the
 * pile's top rank, or playing a Jack, captures the whole pile, and capturing a
 * lone card scores a Pişti bonus. `next` starts the next game after one ends;
 * `reset` applies the config (player count, CPU difficulty); `log` fetches the
 * action log.
 *   - `play` → `{ handIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `log` carry no extra fields.
 */
export const pishtiApi = {
  exec: (command: PishtiCommand, opts?: { handIndex?: number; config?: PishtiConfigInput }) =>
    gameExec<PishtiResponse>('pishti', {
      command,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Cuarenta game settings. */
export interface CuarentaConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Cuarenta /cuarenta/exec endpoint. */
export type CuarentaCommand = 'reset' | 'play' | 'next' | 'log';

/**
 * API client for the Cuarenta /cuarenta/exec endpoint.
 *
 * Cuarenta is an Ecuadorian 4-player, 2-team capture game played with a 40-card
 * deck (no 8/9/10). On your turn you `play` a hand card (`play` → `{ handIndex }`):
 * it captures all same-rank table cards (with caída / ronda / limpia bonuses) or
 * is laid on the table. `next` starts the next round, `reset` applies the config
 * (CPU difficulty, target score), and `log` fetches the action log.
 *   - `play` → `{ handIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `log` carry no extra fields.
 */
export const cuarentaApi = {
  exec: (command: CuarentaCommand, opts?: { handIndex?: number; config?: CuarentaConfigInput }) =>
    gameExec<CuarentaResponse>('cuarenta', {
      command,
      handIndex: opts?.handIndex,
      config: opts?.config,
    }),
};

/** Commands accepted by the Faro /faro/exec endpoint. */
export type FaroCommand = 'reset' | 'bet' | 'clearBet' | 'clearAll' | 'deal' | 'call' | 'next' | 'log';

/**
 * API client for the Faro /faro/exec endpoint.
 *
 * Faro is a 19th-century single-player-vs-bank banking game. The player places
 * chips on a 13-rank layout (A=1 .. K=13) during the Betting phase, then the
 * bank deals turns of two cards (loser then winner). Commands:
 *   - `bet` → `{ rank, amount, copper }` (copper = bet the rank to lose)
 *   - `clearBet` → `{ rank }`
 *   - `clearAll` / `deal` / `next` / `log` carry no extra fields
 *   - `call` → `{ order }` predicting the order of the final three cards (4:1)
 */
export const faroApi = {
  exec: (command: FaroCommand, opts?: { rank?: number; amount?: number; copper?: boolean; order?: number[] }) =>
    gameExec<FaroResponse>('faro', {
      command,
      rank: opts?.rank,
      amount: opts?.amount,
      copper: opts?.copper,
      order: opts?.order,
    }),
};

/** Commands accepted by the Open Face Chinese Poker (OFC) /openfacechinese/exec endpoint. */
export type OpenFaceChineseCommand = 'reset' | 'place' | 'nextround' | 'hint' | 'log';

/** Commands accepted by the Russian Bank /russianbank/exec endpoint. */
export type RussianBankCommand = 'reset' | 'pf' | 'mt' | 'd' | 's' | 'u' | 'hint' | 'log';

/**
 * API client for the Open Face Chinese Poker (OFC) /openfacechinese/exec endpoint.
 *
 * OFC is a solo-vs-CPU game where each dealt card must be committed to one of
 * three rows — Top (3 cards), Middle (5 cards) or Bottom (5 cards) — and once
 * placed a card cannot be moved. Commands:
 *   - `place` -> `{ row }` where row is 0=Top, 1=Middle, 2=Bottom
 *   - `reset` / `nextround` / `hint` / `log` carry no extra fields
 */
export const openfacechineseApi = {
  exec: (command: OpenFaceChineseCommand, opts?: { row?: number }) =>
    gameExec<OpenFaceChineseResponse>('openfacechinese', {
      command,
      row: opts?.row,
    }),
};

/** Options accepted by a Russian Bank move command. */
export interface RussianBankMoveOpts {
  zone?: number;
  fromOpp?: boolean;
  col?: number;
  toCol?: number;
  config?: { cpuDifficulty?: number };
}

/** API client for the Russian Bank /russianbank/exec endpoint. */
export const russianbankApi = {
  exec: (command: RussianBankCommand, opts?: RussianBankMoveOpts) =>
    gameExec<RussianBankResponse>('russianbank', {
      command,
      zone: opts?.zone,
      fromOpp: opts?.fromOpp,
      col: opts?.col,
      toCol: opts?.toCol,
      config: opts?.config,
    }),
};

/** Configuration options for Préférence game settings. */
export interface PreferenceConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Préférence /preference/exec endpoint. */
export type PreferenceCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Préférence /preference/exec endpoint.
 *
 * Préférence is a Russian/Austrian 3-player 32-card trick-taker with a bidding
 * phase. Each player bids once (Pass/Six/Misère/Seven/Eight); the highest bidder
 * is the declarer who plays alone against the other 2 defenders.
 *   - `bid` → `{ bid: number }` (0=Pass 1=Six 2=Misère 3=Seven 4=Eight)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const preferenceApi = {
  exec: (
    command: PreferenceCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: PreferenceConfigInput;
    },
  ) =>
    gameExec<PreferenceResponse>('preference', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Nap (Napoleon) game settings. */
export interface NapConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Nap /nap/exec endpoint. */
export type NapCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Nap (Napoleon) /nap/exec endpoint.
 *
 * Nap is a British 4-player 5-card gambling trick-taker with a bidding phase.
 * Each player bids once (Pass/Two/Three/Four/Nap = how many of the 5 tricks they
 * will take); the highest bidder becomes the declarer who picks trump and leads.
 *   - `bid` → `{ bid: number }` (0=Pass 2=Two 3=Three 4=Four 5=Nap)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const napApi = {
  exec: (
    command: NapCommand,
    opts?: {
      bid?: number;
      cardIndex?: number;
      config?: NapConfigInput;
    },
  ) =>
    gameExec<NapResponse>('nap', {
      command,
      bid: opts?.bid,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};

/** Configuration options for Tute game settings. */
export interface TuteConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Tute /tute/exec endpoint. */
export type TuteCommand = 'reset' | 'play' | 'marriage' | 'tute' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tute /tute/exec endpoint.
 *
 * Tute is a Spanish 4-player (2 vs 2) trump trick-taker. The play actions are:
 *   - `play` → `{ cardIndex: number }`
 *   - `marriage` → `{ suit: number }` (declare a King+Queen marriage; 1=♠ 2=♣ 3=♥ 4=♦)
 *   - `tute` → no extra fields (declare four Kings or four Queens for an instant win)
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tuteApi = {
  exec: (
    command: TuteCommand,
    opts?: {
      cardIndex?: number;
      suit?: number;
      config?: TuteConfigInput;
    },
  ) =>
    gameExec<TuteResponse>('tute', {
      command,
      cardIndex: opts?.cardIndex,
      suit: opts?.suit,
      config: opts?.config,
    }),
};

/** Source or target zone for a Spider card move. */
export interface SpiderMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** Configuration options for Spider game settings. */
export interface SpiderConfigInput {
  difficulty?: number;
}

/** API client for the Spider /spider/exec endpoint. */
export const spiderApi = createSolitaireMoveApiWithConfig<
  SpiderResponse,
  SpiderMoveZone,
  SpiderConfigInput,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('spider');

/** Source or target zone for a Spiderette card move. */
export interface SpideretteMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Spiderette /spiderette/exec endpoint. */
export const spideretteApi = createSolitaireMoveApi<
  SpideretteResponse,
  SpideretteMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('spiderette');

/** Configuration options for Napoleon game settings. */
export interface NapoleonConfigInput {
  cpuDifficulty?: number;
  minBid?: number;
  pointLimit?: number;
}

/** API client for the Napoleon /napoleon/exec endpoint. */
export const napoleonApi = {
  exec: (
    command:
      | 'reset'
      | 'bid'
      | 'trump'
      | 'exchange'
      | 'play'
      | 'next'
      | 'nextround'
      | 'hint'
      | 'log'
      | 'setdifficulty'
      | 'setlimit'
      | 'setminbid',
    bid?: number,
    trumpSuit?: number,
    adjutantSuit?: number,
    adjutantValue?: number,
    discardIndex?: number,
    cardIndex?: number,
    config?: NapoleonConfigInput,
  ) =>
    gameExec<NapoleonResponse>('napoleon', {
      command,
      bid,
      trumpSuit,
      adjutantSuit,
      adjutantValue,
      discardIndex,
      cardIndex,
      config,
    }),
};

/** Configuration options for Mighty game settings. */
export interface MightyConfigInput {
  cpuDifficulty?: number;
  minBid?: number;
  noTrumpExtra?: number;
  pointLimit?: number;
}

/** API client for the Mighty /mighty/exec endpoint. */
export const mightyApi = {
  exec: (
    command:
      | 'reset'
      | 'b'
      | 'bid'
      | 't'
      | 'trump'
      | 'e'
      | 'exchange'
      | 'p'
      | 'play'
      | 'jl'
      | 'jokerlead'
      | 'n'
      | 'next'
      | 'nr'
      | 'nextround'
      | 'hint'
      | 'log',
    bid?: number,
    noTrump?: boolean,
    cardIndex?: number,
    trumpSuit?: number,
    partnerSuit?: number,
    partnerValue?: number,
    discardIndices?: number[],
    jokerLeadSuit?: number,
    config?: MightyConfigInput,
  ) =>
    gameExec<MightyResponse>('mighty', {
      command,
      bid,
      noTrump,
      cardIndex,
      trumpSuit,
      partnerSuit,
      partnerValue,
      discardIndices,
      jokerLeadSuit,
      config,
    }),
};

/** Configuration options for 500 (Five Hundred) game settings. */
export interface FiveHundredConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Optional parameters for a 500 (Five Hundred) action. */
export interface FiveHundredParams {
  bidKind?: number;
  bidTricks?: number;
  bidSuit?: number;
  discardIndices?: number[];
  cardIndex?: number;
  jokerSuit?: number;
  config?: FiveHundredConfigInput;
}

/** API client for the 500 (Five Hundred) game. Calls POST /fivehundred/exec. */
export const fiveHundredApi = {
  exec: (
    command:
      | 'reset'
      | 'b'
      | 'bid'
      | 'pa'
      | 'pass'
      | 'e'
      | 'exchange'
      | 'p'
      | 'play'
      | 'n'
      | 'next'
      | 'nr'
      | 'nextround'
      | 'hint'
      | 'log',
    params: FiveHundredParams = {},
  ) => gameExec<FiveHundredResponse>('fivehundred', { command, ...params }),
};

/** Configuration options for Bid Whist game settings. */
export interface BidWhistConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Optional parameters for a Bid Whist action. */
export interface BidWhistParams {
  bidTricks?: number;
  bidDirection?: number;
  trumpSuit?: number;
  discardIndices?: number[];
  cardIndex?: number;
  config?: BidWhistConfigInput;
}

/** API client for the Bid Whist game. Calls POST /bidwhist/exec. */
export const bidWhistApi = {
  exec: (
    command:
      | 'reset'
      | 'b'
      | 'bid'
      | 'pa'
      | 'pass'
      | 't'
      | 'trump'
      | 'e'
      | 'exchange'
      | 'p'
      | 'play'
      | 'n'
      | 'next'
      | 'nr'
      | 'nextround'
      | 'hint'
      | 'log',
    params: BidWhistParams = {},
  ) => gameExec<BidWhistResponse>('bidwhist', { command, ...params }),
};

/** Configuration options for Skat game settings. */
export interface SkatConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** API client for the Skat /skat/exec endpoint. */
export const skatApi = {
  exec: (
    command: 'reset' | 'bid' | 'pickskat' | 'discard' | 'game' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    args?: {
      accept?: boolean;
      pickup?: boolean;
      discardA?: number;
      discardB?: number;
      gameType?: number;
      trumpSuit?: number;
      cardIndex?: number;
      config?: SkatConfigInput;
    },
  ) =>
    gameExec<SkatResponse>('skat', {
      command,
      ...(args || {}),
    }),
};

// SkatConfigType import is used only for type re-export; ensure it's referenced.
export type { SkatConfigType };

/** Configuration options for Shithead game settings. */
export interface ShitheadConfigInput {
  magicTwo?: boolean;
  magicSeven?: boolean;
  magicEight?: boolean;
  magicTen?: boolean;
  fourOfAKindBurn?: boolean;
  cpuDifficulty?: number;
}

/** API client for the Shithead /shithead/exec endpoint. */
export const shitheadApi = {
  exec: (
    command: 'reset' | 'play' | 'log',
    args?: {
      indices?: number[];
      config?: ShitheadConfigInput;
    },
  ) =>
    gameExec<ShitheadResponse>('shithead', {
      command,
      ...(args || {}),
    }),
};

// ShitheadConfigType import is used only for type re-export.
export type { ShitheadConfigType };

/** Configuration options for Nertz / Pounce game settings. */
export interface NertzConfigInput {
  playerCount?: number;
  drawCount?: number;
  targetScore?: number;
  cpuDifficulty?: number;
  cpuTickMoves?: number;
}

/** Source/target zone identifier for a Nertz move. */
export type { NertzMoveZone };

/** API client for the Nertz / Pounce /nertz/exec endpoint. */
export const nertzApi = {
  exec: (
    command: 'reset' | 'nr' | 'tick' | 'd' | 'm' | 'u' | 'h' | 'log',
    args?: {
      playerIdx?: number;
      from?: NertzMoveZone;
      to?: NertzMoveZone;
      config?: NertzConfigInput;
    },
  ) =>
    gameExec<NertzResponse>('nertz', {
      command,
      ...(args || {}),
    }),
};

// NertzConfigType import is used only for type re-export.
export type { NertzConfigType };

/** Configuration options for Indian Poker game settings. */
export interface IndianPokerConfigInput {
  ante?: number;
  bettingLimit?: number;
  cpuMetaAI?: boolean;
}

/** API client for the Indian Poker /indianpoker/exec endpoint. */
export const indianpokerApi = {
  exec: (
    command: 'reset' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin' | 'log',
    amount?: number,
    config?: IndianPokerConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<IndianPokerResponse>('indianpoker', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** Configuration options for Bridge game settings. */
export interface BridgeConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Bridge /bridge/exec endpoint. */
export const bridgeApi = {
  exec: (
    command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    bidType?: number,
    bidLevel?: number,
    bidSuit?: number,
    config?: BridgeConfigInput,
  ) =>
    gameExec<BridgeResponse>('bridge', {
      command,
      cardIndex,
      bidType,
      bidLevel,
      bidSuit,
      config,
    }),
};

/** Configuration options for Euchre game settings. */
export interface EuchreConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Belote game configuration input shape. */
export interface BeloteConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  dixDeDer?: number;
  enableBeloteRebelote?: boolean;
}

/** API client for the Belote /belote/exec endpoint. */
export const beloteApi = {
  exec: (
    command: 'reset' | 'orderup' | 'pass' | 'calltrump' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndex?: number,
    suit?: number,
    config?: BeloteConfigInput,
  ) =>
    gameExec<BeloteResponse>('belote', {
      command,
      cardIndex,
      suit,
      config,
    }),
};

/** Jass (Schieber) game configuration input shape. */
export interface JassConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  lastTrickBonus?: number;
  enableWeis?: boolean;
}

/** API client for the Jass /jass/exec endpoint. */
export const jassApi = {
  exec: (
    command: 'reset' | 'calltrump' | 'schieben' | 'play' | 'next' | 'nextround' | 'hint',
    suit?: number,
    cardIndex?: number,
    config?: JassConfigInput,
  ) =>
    gameExec<JassResponse>('jass', {
      command,
      suit,
      cardIndex,
      config,
    }),
};

/** Watten (ヴァッテン) game configuration input shape. */
export interface WattenConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  maxRaises?: number;
}

/**
 * API client for the Watten /watten/exec endpoint.
 *
 * Watten is a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff
 * stake mechanic. `declare` carries the Schlag `rank` and critical `suit`, `play`
 * carries a `cardIndex`, `raise` takes no args, and `respond` carries `hold`
 * (true = hold/accept, false = fold/concede).
 */
export const wattenApi = {
  exec: (
    command: 'reset' | 'declare' | 'play' | 'raise' | 'respond' | 'nextround' | 'hint',
    rank?: number,
    suit?: number,
    cardIndex?: number,
    hold?: boolean,
    config?: WattenConfigInput,
  ) =>
    gameExec<WattenResponse>('watten', {
      command,
      rank,
      suit,
      cardIndex,
      hold,
      config,
    }),
};

/** Gaigel game configuration input. */
export interface GaigelConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/**
 * API client for the Gaigel /gaigel/exec endpoint.
 *
 * The second positional slot is unused (Gaigel has no suit/bid argument); it
 * exists so the exec signature matches the `(command, arg1?, cardIndex?, config?)`
 * shape that `useTrickGameBase` dispatches for reset/play.
 */
export const gaigelApi = {
  exec: (
    command: 'reset' | 'play' | 'marriage' | 'next' | 'nextround' | 'hint',
    _unused?: number,
    cardIndex?: number,
    config?: GaigelConfigInput,
  ) =>
    gameExec<GaigelResponse>('gaigel', {
      command,
      cardIndex,
      config,
    }),
};

/** API client for the Euchre /euchre/exec endpoint. */
export const euchreApi = {
  exec: (
    command: 'reset' | 'orderup' | 'pass' | 'calltrump' | 'discard' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndex?: number,
    suit?: number,
    goAlone?: boolean,
    config?: EuchreConfigInput,
  ) =>
    gameExec<EuchreResponse>('euchre', {
      command,
      cardIndex,
      suit,
      goAlone,
      config,
    }),
};

/** Source card for a Pyramid remove action. */
export interface PyramidRemoveCard {
  zone: string;
  row?: number;
  col?: number;
}

/** API client for the Pyramid /pyramid/exec endpoint. */
export const pyramidApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    card1?: PyramidRemoveCard,
    card2?: PyramidRemoveCard,
    n?: number,
  ) => gameExec<PyramidResponse>('pyramid', { command, card1, card2, n }),
};

/** API client for the TriPeaks /tripeaks/exec endpoint. */
export const tripeaksApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    row?: number,
    col?: number,
    n?: number,
  ) => gameExec<TriPeaksResponse>('tripeaks', { command, row, col, n }),
};

/** Factory for video poker variant APIs that share the same exec pattern. */
function createVideoPokerApi(game: string) {
  return {
    exec: (command: 'reset' | 'bet' | 'hold' | 'log', amount?: number, indices?: number[]) =>
      gameExec<VideoPokerResponse>(game, { command, amount, indices }),
  };
}

/** API client for the Video Poker /videopoker/exec endpoint. */
export const videopokerApi = createVideoPokerApi('videopoker');

/** API client for the Deuces Wild /deuceswild/exec endpoint. */
export const deuceswildApi = createVideoPokerApi('deuceswild');

/** API client for the Joker Poker /jokerpoker/exec endpoint. */
export const jokerpokerApi = createVideoPokerApi('jokerpoker');

/** API client for the War /war/exec endpoint. */
export const warApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'log', config?: { maxRounds?: number }) =>
    gameExec<WarResponse>('war', { command, ...config }),
};

/** API client for the Beggar-My-Neighbour /beggarmyneighbour/exec endpoint. */
export const beggarmyneighbourApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'log', config?: { maxRounds?: number }) =>
    gameExec<BeggarMyNeighbourResponse>('beggarmyneighbour', { command, ...config }),
};

/** Configuration options for All Fours (Seven Up) game settings. */
export interface AllFoursConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the All Fours /allfours/exec endpoint. */
export const allfoursApi = {
  exec: (
    command: 'reset' | 'beg' | 'respond' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    beg?: boolean,
    run?: boolean,
    cardIndex?: number,
    config?: AllFoursConfigInput,
  ) => gameExec<AllFoursResponse>('allfours', { command, beg, run, cardIndex, config }),
};

/** Configuration options for Slapjack game settings. */
export interface SlapjackConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Slapjack /slapjack/exec endpoint. */
export const slapjackApi = {
  exec: (command: 'reset' | 'step' | 'slap' | 'tick' | 'log', args?: { config?: SlapjackConfigInput }) =>
    gameExec<SlapjackResponse>('slapjack', {
      command,
      ...(args || {}),
    }),
};

/** Configuration options for Egyptian Ratscrew game settings. */
export interface EgyptianRatscrewConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Egyptian Ratscrew /egyptianratscrew/exec endpoint. */
export const egyptianRatscrewApi = {
  exec: (command: 'reset' | 'step' | 'slap' | 'tick' | 'log', args?: { config?: EgyptianRatscrewConfigInput }) =>
    gameExec<EgyptianRatscrewResponse>('egyptianratscrew', {
      command,
      ...(args || {}),
    }),
};

/** API client for the Fifty-one /fiftyone/exec endpoint. */
export const fiftyoneApi = {
  exec: (
    command: 'reset' | 'play' | 'exchangeall' | 'stop' | 'log',
    opts?: { handIdx?: number; tableIdx?: number; config?: { cpuDifficulty?: number } },
  ) => gameExec<FiftyOneResponse>('fiftyone', { command, ...opts }),
};

/** Source or target zone for a Yukon card move. */
export interface YukonMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Yukon /yukon/exec endpoint. */
export const yukonApi = createSolitaireMoveApi<
  YukonResponse,
  YukonMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('yukon');

/** Source or target zone for a Cruel card move. */
export interface CruelMoveZone {
  zone: string;
  col?: number;
}

/** API client for the Cruel /cruel/exec endpoint. */
export const cruelApi = createSolitaireMoveApi<
  CruelResponse,
  CruelMoveZone,
  'reset' | 'move' | 'shift' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('cruel');

/** Source or target zone for a Russian Solitaire card move. */
export interface RussianSolitaireMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Russian Solitaire /russiansolitaire/exec endpoint. */
export const russianSolitaireApi = createSolitaireMoveApi<
  RussianSolitaireResponse,
  RussianSolitaireMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('russiansolitaire');

/** Source or target zone for a Scorpion card move. */
export interface ScorpionMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Scorpion /scorpion/exec endpoint. */
export const scorpionApi = createSolitaireMoveApi<
  ScorpionResponse,
  ScorpionMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('scorpion');

/** Source or target zone for a Wasp card move. */
export interface WaspMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Wasp /wasp/exec endpoint. */
export const waspApi = createSolitaireMoveApi<
  WaspResponse,
  WaspMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('wasp');

/** Source or target zone for an Easthaven card move. */
export interface EasthavenMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Easthaven /easthaven/exec endpoint. */
export const easthavenApi = createSolitaireMoveApi<
  EasthavenResponse,
  EasthavenMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('easthaven');

/** Source or target pile for an Accordion move. */
export interface AccordionMoveZone {
  zone: 'pile';
  index?: number;
}

/** API client for the Accordion /accordion/exec endpoint. */
export const accordionApi = createSolitaireMoveApi<
  AccordionResponse,
  AccordionMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n'
>('accordion');

/** Configuration options for Seven Bridge game settings. */
export interface SevenBridgeConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** Command verbs accepted by the Seven Bridge /sevenbridge/exec endpoint. */
export type SevenBridgeCommand =
  | 'reset'
  | 'drawstock'
  | 'pon'
  | 'chi'
  | 'meld'
  | 'layoff'
  | 'discard'
  | 'nextround'
  | 'log';

/** API client for the Seven Bridge /sevenbridge/exec endpoint. */
export const sevenBridgeApi = {
  exec: (
    command: SevenBridgeCommand,
    cardIndex?: number,
    config?: SevenBridgeConfigInput,
    cardIndices?: number[],
    targetPlayerIdx?: number,
    meldIdx?: number,
  ) =>
    gameExec<SevenBridgeResponse>('sevenbridge', {
      command,
      cardIndex,
      cardIndices,
      targetPlayerIdx,
      meldIdx,
      config,
    }),
};

/** Command verbs accepted by the Trash /trash/exec endpoint. */
export type TrashCommand = 'reset' | 'draw' | 'place' | 'cpu' | 'log';

/** API client for the Trash /trash/exec endpoint. */
export const trashApi = {
  exec: (command: TrashCommand, position?: number) => gameExec<TrashResponse>('trash', { command, position }),
};

/** API client for the Speed /speed/exec endpoint. */
export const speedApi = {
  exec: (
    command: 'reset' | 'play' | 'flip' | 'hint' | 'log',
    cardIndex?: number,
    pileIndex?: number,
    config?: SpeedConfig,
  ) => gameExec<SpeedResponse>('speed', { command, cardIndex, pileIndex, ...config }),
};

/** Configuration options for Go Fish game settings. */
export interface GoFishConfigInput {
  cpuDifficulty?: number;
}

/** API client for the Go Fish /gofish/exec endpoint. */
export const goFishApi = {
  exec: (command: 'reset' | 'ask' | 'log', targetIdx?: number, rank?: number, config?: GoFishConfigInput) =>
    gameExec<GoFishResponse>('gofish', { command, targetIdx, rank, config }),
};

/** API client for the Golf Solitaire /golf/exec endpoint. */
export const golfApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    col?: number,
    n?: number,
  ) => gameExec<GolfResponse>('golf', { command, col, n }),
};

/** API client for the Aces Up /acesup/exec endpoint. */
export const acesupApi = {
  exec: (
    command: 'reset' | 'draw' | 'remove' | 'move' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n',
    col?: number,
    n?: number,
  ) => gameExec<AcesUpResponse>('acesup', { command, col, n }),
};

/** Pig's Tail game API client. */
export const pigtailApi = {
  exec: (command: 'reset' | 'draw', cpuHesitationEnabled?: boolean, playerCount?: number) =>
    gameExec<PigsTailResponse>('pigtail', { command, cpuHesitationEnabled, playerCount }),
};

/** API client for the Clock Solitaire /clocksolitaire/exec endpoint. */
export const clocksolitaireApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'undo' | 'log') =>
    gameExec<ClockSolitaireResponse>('clocksolitaire', { command }),
};

/** API client for the Forty Thieves /fortythieves/exec endpoint. */
export const fortyThievesApi = createSolitaireMoveApi<
  FortyThievesResponse,
  FortyThievesMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fortythieves');

/** API client for the Forty and Eight /fortyandeight/exec endpoint. */
export const fortyAndEightApi = createSolitaireMoveApi<
  FortyAndEightResponse,
  FortyAndEightMoveZone,
  'reset' | 'draw' | 'redeal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fortyandeight');

/** API client for the Sultan of Turkey /sultan/exec endpoint. */
export const sultanApi = createSolitaireMoveApi<
  SultanResponse,
  SultanMoveZone,
  'reset' | 'draw' | 'redeal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('sultan');

/** API client for the Crescent Solitaire /crescent/exec endpoint. */
export const crescentApi = createSolitaireMoveApi<
  CrescentResponse,
  CrescentMoveZone,
  'reset' | 'move' | 'redeal' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('crescent');

/** API client for the Baker's Dozen /bakersdozen/exec endpoint. */
export const bakersDozenApi = createSolitaireMoveApi<
  BakersDozenResponse,
  BakersDozenMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bakersdozen');

/** API client for the Beleaguered Castle /beleagueredcastle/exec endpoint. */
export const beleagueredCastleApi = createSolitaireMoveApi<
  BeleagueredCastleResponse,
  BeleagueredCastleMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('beleagueredcastle');

/** API client for the Streets and Alleys /streetsandalleys/exec endpoint. */
export const streetsAndAlleysApi = createSolitaireMoveApi<
  StreetsAndAlleysResponse,
  StreetsAndAlleysMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('streetsandalleys');

/** API client for the King Albert /kingalbert/exec endpoint. */
export const kingAlbertApi = createSolitaireMoveApi<
  KingAlbertResponse,
  KingAlbertMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('kingalbert');

/** API client for the Flower Garden /flowergarden/exec endpoint. */
export const flowerGardenApi = createSolitaireMoveApi<
  FlowerGardenResponse,
  FlowerGardenMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('flowergarden');

/** API client for the Calculation /calculation/exec endpoint. */
export const calculationApi = createSolitaireMoveApi<
  CalculationResponse,
  CalculationMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('calculation');

/** Command verbs accepted by the Spite & Malice /spiteandmalice/exec endpoint. */
export type SpiteAndMaliceCommand = 'reset' | 'move' | 'discard' | 'cpu' | 'autocomplete' | 'hint' | 'log';

/** API client for the Spite & Malice /spiteandmalice/exec endpoint. */
export const spiteAndMaliceApi = {
  exec: (command: SpiteAndMaliceCommand, from?: SpiteAndMaliceMoveZone, to?: SpiteAndMaliceMoveZone) =>
    gameExec<SpiteAndMaliceResponse>('spiteandmalice', { command, from, to }),
};

/** Configuration options for President game settings. */
export interface PresidentConfigInput {
  revolutionEnabled?: boolean;
  cardExchangeEnabled?: boolean;
  passFieldFlushEnabled?: boolean;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the President /president/exec endpoint. */
export type PresidentCommand = 'reset' | 'play' | 'log';

/** API client for the President /president/exec endpoint. */
export const presidentApi = {
  exec: (command: PresidentCommand, indices?: number[], config?: PresidentConfigInput) =>
    gameExec<PresidentResponse>('president', { command, indices, config }),
};

/** Configuration options for Cassino game settings. */
export interface CassinoConfigInput {
  targetScore?: number;
  multiBuildEnabled?: boolean;
  sweepBonusEnabled?: boolean;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Cassino /cassino/exec endpoint. */
export type CassinoCommand = 'reset' | 'take' | 'build' | 'trail' | 'next' | 'log';

/** Extra payload fields for the Cassino /cassino/exec endpoint. */
export interface CassinoExecParams {
  handIndex?: number;
  tableIndices?: number[];
  buildIndices?: number[];
  declaredValue?: number;
  config?: CassinoConfigInput;
}

/** API client for the Cassino /cassino/exec endpoint. */
export const cassinoApi = {
  exec: (command: CassinoCommand, params?: CassinoExecParams) =>
    gameExec<CassinoResponse>('cassino', { command, ...(params ?? {}) }),
};

/** Configuration options for Scopa game settings. */
export interface ScopaConfigInput {
  targetScore?: number;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Scopa /scopa/exec endpoint (short forms). */
export type ScopaCommand = 'r' | 'n' | 'p' | 'log';

/** Extra payload fields for the Scopa /scopa/exec endpoint. */
export interface ScopaExecParams {
  handIndex?: number;
  tableIndices?: number[];
  config?: ScopaConfigInput;
}

/** API client for the Scopa /scopa/exec endpoint. */
export const scopaApi = {
  exec: (command: ScopaCommand, params?: ScopaExecParams) =>
    gameExec<ScopaResponse>('scopa', { command, ...(params ?? {}) }),
};

/** Configuration options for Scopone game settings. */
export interface ScoponeConfigInput {
  targetScore?: number;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Scopone /scopone/exec endpoint (short forms). */
export type ScoponeCommand = 'r' | 'n' | 'p' | 'log';

/** Extra payload fields for the Scopone /scopone/exec endpoint. */
export interface ScoponeExecParams {
  handIndex?: number;
  tableIndices?: number[];
  config?: ScoponeConfigInput;
}

/** API client for the Scopone /scopone/exec endpoint. */
export const scoponeApi = {
  exec: (command: ScoponeCommand, params?: ScoponeExecParams) =>
    gameExec<ScoponeResponse>('scopone', { command, ...(params ?? {}) }),
};

/** Configuration options for Escoba game settings. */
export interface EscobaConfigInput {
  targetScore?: number;
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Escoba /escoba/exec endpoint (short forms). */
export type EscobaCommand = 'r' | 'n' | 'p' | 'log';

/** Extra payload fields for the Escoba /escoba/exec endpoint. */
export interface EscobaExecParams {
  handIndex?: number;
  tableIndices?: number[];
  config?: EscobaConfigInput;
}

/** API client for the Escoba /escoba/exec endpoint. */
export const escobaApi = {
  exec: (command: EscobaCommand, params?: EscobaExecParams) =>
    gameExec<EscobaResponse>('escoba', { command, ...(params ?? {}) }),
};

/** Configuration options for Barbu game settings. */
export interface BarbuConfigInput {
  cpuDifficulty?: number;
}

/** Command verbs accepted by the Barbu /barbu/exec endpoint (short forms). */
export type BarbuCommand = 'r' | 'n' | 'c' | 'p' | 'log';

/** Extra payload fields for the Barbu /barbu/exec endpoint. */
export interface BarbuExecParams {
  contract?: number;
  trumpSuit?: number;
  handIndex?: number;
  tableIndices?: number[];
  config?: BarbuConfigInput;
}

/** API client for the Barbu /barbu/exec endpoint. */
export const barbuApi = {
  exec: (command: BarbuCommand, params?: BarbuExecParams) =>
    gameExec<BarbuResponse>('barbu', { command, ...(params ?? {}) }),
};

/** Configuration options for Macau game settings. */
export interface MacauConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Macau /macau/exec endpoint. */
export const macauApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'suit' | 'declare' | 'skipdeclare' | 'nextround',
    cardIndex?: number,
    suit?: number,
    config?: MacauConfigInput,
  ) =>
    gameExec<MacauResponse>('macau', {
      command,
      cardIndex,
      suit,
      config,
    }),
};

/** Configuration options for Mao game settings. */
export interface MaoConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/**
 * API client for the Mao /mao/exec endpoint.
 *
 * Mao is a Crazy Eights-style shedding game with a secret hidden rule. The
 * `declareword` command carries the player's compliance utterance (`word`);
 * the server never reveals the rule itself.
 */
export const maoApi = {
  exec: (
    command: 'reset' | 'play' | 'draw' | 'suit' | 'declare' | 'skipdeclare' | 'declareword' | 'nextround',
    cardIndex?: number,
    suit?: number,
    config?: MaoConfigInput,
    word?: string,
  ) =>
    gameExec<MaoResponse>('mao', {
      command,
      cardIndex,
      suit,
      config,
      word,
    }),
};

export type {
  BakersDozenMoveZone,
  BeleagueredCastleMoveZone,
  CrescentMoveZone,
  FlowerGardenMoveZone,
  FortyAndEightMoveZone,
  FortyThievesMoveZone,
  KingAlbertMoveZone,
  StreetsAndAlleysMoveZone,
  SultanMoveZone,
};

/** API client for the Whist /whist/exec endpoint. */
export const whistApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<WhistConfig>,
  ) => gameExec<WhistResponse>('whist', { command, cardIndex, config }),
};

/** API client for the Catch the Ten /catchten/exec endpoint. */
export const catchtenApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<CatchTenConfig>,
  ) => gameExec<CatchTenResponse>('catchten', { command, cardIndex, config }),
};

/** API client for the Briscola /briscola/exec endpoint. */
export const briscolaApi = {
  exec: (command: 'reset' | 'play' | 'next' | 'hint' | 'log', cardIndex?: number, config?: Partial<BriscolaConfig>) =>
    gameExec<BriscolaResponse>('briscola', { command, cardIndex, config }),
};

/** API client for the Schnapsen /schnapsen/exec endpoint. */
export const schnapsenApi = {
  exec: (
    command: 'reset' | 'play' | 'marriage' | 'next' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<SchnapsenConfig>,
  ) => gameExec<SchnapsenResponse>('schnapsen', { command, cardIndex, config }),
};

/** API client for the Truco /truco/exec endpoint. */
export const trucoApi = {
  exec: (
    command: 'reset' | 'play' | 'truco' | 'accept' | 'decline' | 'next' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<TrucoConfig>,
  ) => gameExec<TrucoResponse>('truco', { command, cardIndex, config }),
};

/** Source or target zone for a Gaps card move. */
export interface GapsMoveZone {
  zone: 'grid';
  row: number;
  col: number;
}

/** API client for the Gaps /gaps/exec endpoint. */
export const gapsApi = createSolitaireMoveApi<
  GapsResponse,
  GapsMoveZone,
  'reset' | 'move' | 'redeal' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n'
>('gaps');

/** API client for the Poker Squares /pokersquares/exec endpoint. */
export const pokersquaresApi = {
  exec: (command: 'reset' | 'place' | 'undo' | 'giveup' | 'log', row?: number, col?: number) =>
    gameExec<PokerSquaresResponse>('pokersquares', { command, row, col }),
};

/**
 * Factory for casino bet APIs whose request body is `{ command, amount }`.
 * Used by Let It Ride and Red Dog — table games whose only per-action input
 * is the wager amount.
 */
function createBetAmountApi<T, Cmd extends string>(game: string) {
  return {
    exec: (command: Cmd, amount?: number) => gameExec<T>(game, { command, amount }),
  };
}

/** API client for the Let It Ride /letitride/exec endpoint. */
export const letitrideApi = createBetAmountApi<LetItRideResponse, 'reset' | 'bet' | 'pull' | 'letitride' | 'log'>(
  'letitride',
);

/** API client for the Red Dog /reddog/exec endpoint. */
export const reddogApi = createBetAmountApi<RedDogResponse, 'reset' | 'bet' | 'raise' | 'stay' | 'log'>('reddog');

/** API client for the Casino War /casinowar/exec endpoint. */
export const casinowarApi = createBetAmountApi<CasinoWarResponse, 'reset' | 'bet' | 'surrender' | 'war' | 'log'>(
  'casinowar',
);

/** API client for the Oicho-Kabu /oichokabu/exec endpoint. */
export const oichokabuApi = createBetAmountApi<OichoKabuResponse, 'reset' | 'bet' | 'draw' | 'stand' | 'log'>(
  'oichokabu',
);

/** API client for the Dragon Tiger /dragontiger/exec endpoint. */
export const dragontigerApi = {
  exec: (command: 'reset' | 'bet' | 'clear' | 'log', amount?: number, betType?: number) =>
    gameExec<DragonTigerResponse>('dragontiger', { command, amount, betType }),
};

/**
 * API client for the Trente et Quarante (Rouge et Noir) /trenteetquarante/exec endpoint.
 *
 * Trente et Quarante is a pure banking game (no player card decisions).
 *   - `bet` → `(bet, stake)` places the stake on one of the four bets
 *     (0=Noir, 1=Rouge, 2=Couleur, 3=Inverse) and immediately deals both rows
 *     and resolves the round.
 *   - `nextround` starts the next round (chips persist server-side).
 *   - `reset` starts a fresh game (chips persist).
 *   - `log` and `hint` carry no extra fields.
 */
export const trenteetquaranteApi = {
  exec: (command: 'reset' | 'bet' | 'nextround' | 'log' | 'hint', bet?: number, stake?: number) =>
    gameExec<TrenteEtQuaranteResponse>('trenteetquarante', { command, bet, stake }),
};

/** Configuration options accepted by {@link gutsApi}.exec on `reset`. */
export interface GutsConfigInput {
  /** Number of players at the table (2–7, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (1–1000, default 10). */
  ante?: number;
  /** Chips each player starts the match with (10–100000, default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (1–100, default 10). */
  targetRounds?: number;
}

/**
 * API client for the Guts /guts/exec endpoint.
 *
 * Guts is a fast multi-player pot-vying gambling game.
 *   - `reset` starts a fresh game, optionally applying a {@link GutsConfigInput}.
 *   - `declare` → `(declaration)` submits the human's call (0=out/fold,
 *     1=in/stay) and resolves the round.
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const gutsApi = {
  exec: (command: 'reset' | 'declare' | 'nextround' | 'log' | 'hint', declaration?: number, config?: GutsConfigInput) =>
    gameExec<GutsResponse>('guts', { command, declaration, config }),
};

/** Configuration options accepted by {@link anacondaApi}.exec on `reset`. */
export interface AnacondaConfigInput {
  /** Number of players at the table (3–7, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (default 10). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/** Bet action accepted by {@link anacondaApi}.exec on the `bet` command. */
export type AnacondaBetAction = 'call' | 'raise' | 'fold';

/**
 * API client for the Anaconda (Pass the Trash) /anaconda/exec endpoint.
 *
 * Anaconda is a poker pot game.
 *   - `reset` starts a fresh game, optionally applying an {@link AnacondaConfigInput}.
 *   - `pass` → `(cardIndices)` passes the selected cards left (3→2→1).
 *   - `keep` → `(cardIndices)` keeps exactly 5 cards (discarding the other 2).
 *   - `bet` → `(action)` calls (also checks), raises, or folds during Roll.
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const anacondaApi = {
  exec: (
    command: 'reset' | 'pass' | 'keep' | 'bet' | 'nextround' | 'log' | 'hint',
    cardIndices?: number[],
    action?: AnacondaBetAction,
    config?: AnacondaConfigInput,
  ) => gameExec<AnacondaResponse>('anaconda', { command, cardIndices, action, config }),
};

/** Configuration options accepted by {@link bouillotteApi}.exec on `reset`. */
export interface BouillotteConfigInput {
  /** Number of players at the table (3–4, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (default 10). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/**
 * API client for the Bouillotte /bouillotte/exec endpoint.
 *
 * Bouillotte is an 18th-century French 3-card poker-vying pot game.
 *   - `reset` starts a fresh game, optionally applying a {@link BouillotteConfigInput}.
 *   - `bet` → `(action)` submits the human's betting action (`"call"` /
 *     `"raise"` / `"fold"`). Raise uses a fixed increment (no amount field).
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const bouillotteApi = {
  exec: (
    command: 'reset' | 'bet' | 'nextround' | 'log' | 'hint',
    action?: 'call' | 'raise' | 'fold',
    config?: BouillotteConfigInput,
  ) => gameExec<BouillotteResponse>('bouillotte', { command, action, config }),
};

/** Configuration options accepted by {@link primeroApi}.exec on `reset`. */
export interface PrimeroConfigInput {
  /** Number of players at the table (2–6, default 4). */
  playerCount?: number;
  /** Chips each player antes into the pot per round (default 10). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/**
 * API client for the Primero /primero/exec endpoint.
 *
 * Primero is a Renaissance (16th-century) 4-card poker-vying pot game.
 *   - `reset` starts a fresh game, optionally applying a {@link PrimeroConfigInput}.
 *   - `bet` → `(action)` submits the human's betting action (`"call"` /
 *     `"raise"` / `"fold"`). Raise uses a fixed increment (no amount field).
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const primeroApi = {
  exec: (
    command: 'reset' | 'bet' | 'nextround' | 'log' | 'hint',
    action?: 'call' | 'raise' | 'fold',
    config?: PrimeroConfigInput,
  ) => gameExec<PrimeroResponse>('primero', { command, action, config }),
};

/** Configuration options accepted by {@link michiganApi}.exec on `reset`. */
export interface MichiganConfigInput {
  /** Number of players at the table (3–8, default 4). */
  playerCount?: number;
  /** Total chips each player distributes across the four boodles per round (default 8). */
  ante?: number;
  /** Chips each player starts the match with (default 200). */
  startingChips?: number;
  /** Rounds played before the richest player wins the match (default 10). */
  targetRounds?: number;
}

/**
 * API client for the Michigan (Newmarket) /michigan/exec endpoint.
 *
 * Michigan is a "stops" chip-betting game.
 *   - `reset` starts a fresh game, optionally applying a {@link MichiganConfigInput}.
 *   - `bet` → `(boodleBets)` distributes the human's chips across the four
 *     boodles (order A♥, K♣, Q♦, J♠); the array must sum to `betBudget`.
 *   - `play` → `(cardIndex)` plays the hand card at that index (must be in
 *     `playableIndices`).
 *   - `nextround` deals the next round (chips persist server-side).
 *   - `log` and `hint` carry no extra fields.
 */
export const michiganApi = {
  exec: (
    command: 'reset' | 'bet' | 'play' | 'nextround' | 'log' | 'hint',
    boodleBets?: number[],
    cardIndex?: number,
    config?: MichiganConfigInput,
  ) => gameExec<MichiganResponse>('michigan', { command, boodleBets, cardIndex, config }),
};

/** API client for the Blackjack Switch /blackjackswitch/exec endpoint. */
export const blackjackswitchApi = {
  exec: (command: 'reset' | 'bet' | 'switch' | 'keep' | 'hit' | 'stand' | 'doubledown' | 'log', amount?: number) =>
    gameExec<BlackJackSwitchResponse>('blackjackswitch', { command, amount }),
};

/** API client for the Monte Carlo Solitaire /montecarlo/exec endpoint. */
export const montecarloApi = {
  exec: (
    command: 'reset' | 'remove' | 'deal' | 'undo' | 'giveup' | 'hint' | 'log',
    fromR?: number,
    fromC?: number,
    toR?: number,
    toC?: number,
  ) => gameExec<MonteCarloResponse>('montecarlo', { command, fromR, fromC, toR, toC }),
};

/** API client for the Mississippi Stud /mississippistud/exec endpoint. */
export const mississippiStudApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, multiplier?: number) =>
    gameExec<MississippiStudResponse>('mississippistud', { command, amount, multiplier }),
};

/** API client for the Contract Rummy /contractrummy/exec endpoint. */
export const contractrummyApi = {
  exec: (
    command:
      | 'reset'
      | 'drawstock'
      | 'drawdiscard'
      | 'meldcontract'
      | 'meldextra'
      | 'layoff'
      | 'discard'
      | 'nextround'
      | 'log',
    params?: {
      cardIndex?: number;
      cardIndices?: number[];
      indicesPerSlot?: number[][];
      targetPlayerIdx?: number;
      meldIdx?: number;
      config?: { cpuDifficulty?: number; failContractPenalty?: number };
    },
  ) => gameExec<ContractRummyResponse>('contractrummy', { command, ...(params ?? {}) }),
};

/** API client for the Carioca /carioca/exec endpoint. */
export const cariocaApi = {
  exec: (
    command:
      | 'reset'
      | 'drawstock'
      | 'drawdiscard'
      | 'meldcontract'
      | 'meldextra'
      | 'layoff'
      | 'discard'
      | 'nextround'
      | 'log',
    params?: {
      cardIndex?: number;
      cardIndices?: number[];
      indicesPerSlot?: number[][];
      targetPlayerIdx?: number;
      meldIdx?: number;
      config?: { playerCount?: number; cpuDifficulty?: number; failContractPenalty?: number };
    },
  ) => gameExec<CariocaResponse>('carioca', { command, ...(params ?? {}) }),
};

/** API client for the Kalooki /kalooki/exec endpoint. */
export const kalookiApi = {
  exec: (
    command: 'reset' | 'drawstock' | 'drawdiscard' | 'meld' | 'layoff' | 'discard' | 'nextround' | 'log',
    params?: {
      cardIndex?: number;
      meldGroups?: number[][];
      targetPlayerIdx?: number;
      meldIdx?: number;
      config?: { cpuDifficulty?: number; playerCount?: number; openingThreshold?: number };
    },
  ) => gameExec<KalookiResponse>('kalooki', { command, ...(params ?? {}) }),
};

const games = [
  'blackjack',
  'poker',
  'oldmaid',
  'daifugo',
  'bigtwo',
  'tienlen',
  'sevens',
  'doubt',
  'durak',
  'holdem',
  'omaha',
  'omahahilo',
  'bigo',
  'bigohilo',
  'shortdeck',
  'pineapple',
  'crazypineapple',
  'irishpoker',
  'sevencardstud',
  'fivecardstud',
  'razz',
  'badugi',
  'deucetoseven',
  'hearts',
  'spades',
  'twotenjack',
  'napoleon',
  'ohhell',
  'wizard',
  'ninetynine',
  'memory',
  'klondike',
  'freecell',
  'bakersgame',
  'seahaventowers',
  'cruel',
  'baccarat',
  'crazyeights',
  'prsi',
  'ginrummy',
  'indianrummy',
  'machiavelli',
  'conquian',
  'chinchon',
  'threethirteen',
  'canasta',
  'samba',
  'handandfoot',
  'burraco',
  'spider',
  'indianpoker',
  'videopoker',
  'deuceswild',
  'jokerpoker',
  'euchre',
  'bridge',
  'pyramid',
  'tripeaks',
  'cribbage',
  'threecard',
  'caribbeanstud',
  'texasholdembonus',
  'paigow',
  'speed',
  'war',
  'fiftyone',
  'gofish',
  'pinochle',
  'golf',
  'pigtail',
  'clocksolitaire',
  'fortythieves',
  'calculation',
  'canfield',
  'osmosis',
  'fivehundred',
  'yukon',
  'russiansolitaire',
  'scorpion',
  'wasp',
  'accordion',
  'sevenbridge',
  'trash',
  'whist',
  'catchten',
  'letitride',
  'pokersquares',
  'pageone',
  'reddog',
  'president',
  'cassino',
  'scopa',
  'scopone',
  'escoba',
  'barbu',
  'macau',
  'mao',
  'bristol',
  'bidwhist',
  'spanish21',
  'spiteandmalice',
  'skat',
  'shithead',
  'nertz',
  'slapjack',
  'egyptianratscrew',
  'bakersdozen',
  'thirtyone',
  'yaniv',
  'gongzhu',
  'tonk',
  'casinowar',
  'pitch',
  'dragontiger',
  'blackjackswitch',
  'montecarlo',
  'contractrummy',
  'carioca',
  'kalooki',
  'ultimatetexasholdem',
  'crescent',
  'mississippistud',
  'belote',
  'jass',
  'watten',
  'spiderette',
  'mighty',
  'oasispoker',
  'beleagueredcastle',
  'streetsandalleys',
  'kingalbert',
  'flowergarden',
  'fortyandeight',
  'sultan',
  'agnes',
  'piquet',
  'casinoholdem',
  'callbreak',
  'tarneeb',
  'highcardflush',
  'briscola',
  'gaps',
  'fourcardpoker',
  'rummy500',
  'eightoff',
  'penguin',
  'russianpoker',
  'chinesepoker',
  'sixcardgolf',
  'doudizhu',
  'truco',
  'acesup',
  'schnapsen',
  'tressette',
  'easthaven',
  'tichu',
  'bourre',
  'sheepshead',
  'doppelkopf',
  'mus',
  'tute',
  'sueca',
  'klaverjas',
  'manille',
  'marias',
  'sedma',
  'knockoutwhist',
  'spoilfive',
  'solowhist',
  'fortyfives',
  'nap',
  'preference',
  'twentynine',
  'courtpiece',
  'bezique',
  'ecarte',
  'threecardbrag',
  'teenpatti',
  'spoons',
  'kemps',
  'cuckoo',
  'pishti',
  'cuarenta',
  'faro',
  'openfacechinese',
  'russianbank',
  'labellelucie',
  'simplesimon',
  'doubleklondike',
  'blackhole',
  'beggarmyneighbour',
  'allfours',
  'gaigel',
  'king',
  'tysiac',
  'calabresella',
  'ombre',
  'ulti',
  'rook',
  'cinch',
  'loo',
  'basra',
  'hachihachi',
  'koikoi',
  'gostop',
  'tablanet',
  'trenteetquarante',
  'guts',
  'anaconda',
  'bouillotte',
  'primero',
  'michigan',
  'pan',
  'oichokabu',
  'scarto',
  'cego',
  'frenchtarot',
  'koenigrufen',
  'zheng',
] as const;
type Game = (typeof games)[number];

/** API clients for fetching action logs from each game's /log endpoint. */
export const actionLogApi: { [K in Game]: () => Promise<ActionLogResponse> } = games.reduce(
  (acc, game) => {
    acc[game] = () => gameExec<ActionLogResponse>(game, { command: 'log' });
    return acc;
  },
  {} as { [K in Game]: () => Promise<ActionLogResponse> },
);
