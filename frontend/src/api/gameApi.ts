import type {
  AccordionResponse,
  ActionLogResponse,
  BaccaratResponse,
  BadugiResponse,
  BakersDozenMoveZone,
  BakersDozenResponse,
  BeloteResponse,
  BlackJackResponse,
  BlackJackSwitchResponse,
  BridgeResponse,
  CalculationMoveZone,
  CalculationResponse,
  CanastaResponse,
  CanfieldResponse,
  CaribbeanStudResponse,
  CasinoWarResponse,
  CassinoResponse,
  ClockSolitaireResponse,
  ContractRummyResponse,
  CrazyEightsResponse,
  CrescentMoveZone,
  CrescentResponse,
  CribbageResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DoubtConfig,
  DoubtResponse,
  DragonTigerResponse,
  DurakConfigInput,
  DurakResponse,
  EgyptianRatscrewResponse,
  EuchreResponse,
  FiftyOneResponse,
  FortyThievesMoveZone,
  FortyThievesResponse,
  FreeCellResponse,
  GinRummyResponse,
  GoFishResponse,
  GolfResponse,
  HeartsResponse,
  HoldemResponse,
  IndianPokerResponse,
  KlondikeResponse,
  LetItRideResponse,
  MemoryResponse,
  MississippiStudResponse,
  MonteCarloResponse,
  NapoleonResponse,
  NertzConfig as NertzConfigType,
  NertzMoveZone,
  NertzResponse,
  OhHellResponse,
  OldMaidResponse,
  OmahaResponse,
  PageOneResponse,
  PaiGowResponse,
  PigsTailResponse,
  PineappleResponse,
  PinochleResponse,
  PitchResponse,
  PokerResponse,
  PokerSquaresResponse,
  PresidentResponse,
  PyramidResponse,
  RedDogResponse,
  RussianSolitaireResponse,
  ScorpionResponse,
  SevenBridgeResponse,
  SevenCardStudResponse,
  SevensResponse,
  ShitheadConfig as ShitheadConfigType,
  ShitheadResponse,
  ShortDeckResponse,
  SkatConfig as SkatConfigType,
  SkatResponse,
  SlapjackResponse,
  SpadesResponse,
  SpeedConfig,
  SpeedResponse,
  SpiderResponse,
  SpiteAndMaliceMoveZone,
  SpiteAndMaliceResponse,
  TexasHoldemBonusResponse,
  ThreeCardResponse,
  TonkResponse,
  TrashResponse,
  TriPeaksResponse,
  TwoTenJackResponse,
  UltimateTexasHoldemResponse,
  VideoPokerResponse,
  WarResponse,
  WhistConfig,
  WhistResponse,
  YukonResponse,
} from '../types/card';

/** Unique session identifier for correlating API requests. */
export const sessionId: string = crypto.randomUUID();

/** Worker base URLs for Cloudflare deployment. Empty strings for Docker (relative URLs). */
const WORKER_CASINO = import.meta.env.VITE_WORKER_CASINO_URL || '';
const WORKER_CLASSIC = import.meta.env.VITE_WORKER_CLASSIC_URL || '';
const WORKER_SOLO = import.meta.env.VITE_WORKER_SOLO_URL || '';

/** Maps each game to its Worker base URL. */
const workerUrl: Record<string, string> = {
  blackjack: WORKER_CASINO,
  spanish21: WORKER_CASINO,
  baccarat: WORKER_CASINO,
  poker: WORKER_CASINO,
  holdem: WORKER_CASINO,
  omaha: WORKER_CASINO,
  omahahilo: WORKER_CASINO,
  shortdeck: WORKER_CASINO,
  indianpoker: WORKER_CASINO,
  videopoker: WORKER_CASINO,
  deuceswild: WORKER_CASINO,
  jokerpoker: WORKER_CASINO,
  threecard: WORKER_CASINO,
  caribbeanstud: WORKER_CASINO,
  texasholdembonus: WORKER_CASINO,
  paigow: WORKER_CASINO,
  pineapple: WORKER_CASINO,
  crazypineapple: WORKER_CASINO,
  sevencardstud: WORKER_CASINO,
  razz: WORKER_CASINO,
  badugi: WORKER_CASINO,
  calculation: WORKER_SOLO,
  hearts: WORKER_CLASSIC,
  spades: WORKER_CLASSIC,
  pitch: WORKER_CLASSIC,
  euchre: WORKER_CLASSIC,
  bridge: WORKER_CLASSIC,
  napoleon: WORKER_CLASSIC,
  ohhell: WORKER_CLASSIC,
  oldmaid: WORKER_CLASSIC,
  doubt: WORKER_CLASSIC,
  durak: WORKER_CLASSIC,
  daifugo: WORKER_CLASSIC,
  sevens: WORKER_CLASSIC,
  crazyeights: WORKER_CLASSIC,
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
  spider: WORKER_SOLO,
  pyramid: WORKER_SOLO,
  pokersquares: WORKER_SOLO,
  tripeaks: WORKER_SOLO,
  memory: WORKER_SOLO,
  ginrummy: WORKER_SOLO,
  canasta: WORKER_SOLO,
  cribbage: WORKER_SOLO,
  golf: WORKER_SOLO,
  clocksolitaire: WORKER_SOLO,
  fortythieves: WORKER_SOLO,
  canfield: WORKER_SOLO,
  yukon: WORKER_SOLO,
  russiansolitaire: WORKER_SOLO,
  scorpion: WORKER_SOLO,
  accordion: WORKER_SOLO,
  sevenbridge: WORKER_SOLO,
  trash: WORKER_CLASSIC,
  whist: WORKER_CLASSIC,
  letitride: WORKER_CASINO,
  reddog: WORKER_CASINO,
  casinowar: WORKER_CASINO,
  president: WORKER_CLASSIC,
  cassino: WORKER_CLASSIC,
  spiteandmalice: WORKER_CLASSIC,
  skat: WORKER_CLASSIC,
  shithead: WORKER_CLASSIC,
  nertz: WORKER_CLASSIC,
  slapjack: WORKER_CLASSIC,
  egyptianratscrew: WORKER_CLASSIC,
  bakersdozen: WORKER_SOLO,
  tonk: WORKER_CLASSIC,
  dragontiger: WORKER_CASINO,
  blackjackswitch: WORKER_CASINO,
  montecarlo: WORKER_SOLO,
  contractrummy: WORKER_SOLO,
  ultimatetexasholdem: WORKER_CASINO,
  crescent: WORKER_SOLO,
  mississippistud: WORKER_CASINO,
  belote: WORKER_CLASSIC,
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

/** API client for the Razz /razz/exec endpoint. */
export const razzApi = createHoldemLikeApi<SevenCardStudResponse, SevenCardStudConfigInput>('razz');

/** Configuration options for Pineapple Poker (extends Hold'em with cardIdx for discard). */
export interface PineappleConfigInput extends HoldemConfigInput {
  cardIdx?: number;
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

/** Configuration options for Cribbage game settings. */
export interface CribbageConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Cribbage /cribbage/exec endpoint. */
export const cribbageApi = {
  exec: (
    command: 'reset' | 'discard' | 'peg' | 'go' | 'shownext' | 'nextround' | 'log',
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

/** API client for the Caribbean Stud Poker /caribbeanstud/exec endpoint. */
export const caribbeanstudApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'log', amount?: number, jackpotBet?: number) =>
    gameExec<CaribbeanStudResponse>('caribbeanstud', { command, amount, jackpotBet }),
};

/** API client for the Texas Hold'em Bonus Poker /texasholdembonus/exec endpoint. */
export const texasholdembonusApi = {
  exec: (command: 'reset' | 'bet' | 'play' | 'fold' | 'check' | 'raise' | 'log', amount?: number, bonusBet?: number) =>
    gameExec<TexasHoldemBonusResponse>('texasholdembonus', { command, amount, bonusBet }),
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

/** Pig's Tail game API client. */
export const pigtailApi = {
  exec: (command: 'reset' | 'draw', cpuHesitationEnabled?: boolean) =>
    gameExec<PigsTailResponse>('pigtail', { command, cpuHesitationEnabled }),
};

/** API client for the Clock Solitaire /clocksolitaire/exec endpoint. */
export const clocksolitaireApi = {
  exec: (command: 'reset' | 'step' | 'autoplay' | 'log') =>
    gameExec<ClockSolitaireResponse>('clocksolitaire', { command }),
};

/** API client for the Forty Thieves /fortythieves/exec endpoint. */
export const fortyThievesApi = createSolitaireMoveApi<
  FortyThievesResponse,
  FortyThievesMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fortythieves');

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

export type { BakersDozenMoveZone, CrescentMoveZone, FortyThievesMoveZone };

/** API client for the Whist /whist/exec endpoint. */
export const whistApi = {
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<WhistConfig>,
  ) => gameExec<WhistResponse>('whist', { command, cardIndex, config }),
};

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

/** API client for the Dragon Tiger /dragontiger/exec endpoint. */
export const dragontigerApi = {
  exec: (command: 'reset' | 'bet' | 'clear' | 'log', amount?: number, betType?: number) =>
    gameExec<DragonTigerResponse>('dragontiger', { command, amount, betType }),
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

const games = [
  'blackjack',
  'poker',
  'oldmaid',
  'daifugo',
  'sevens',
  'doubt',
  'durak',
  'holdem',
  'omaha',
  'omahahilo',
  'shortdeck',
  'pineapple',
  'crazypineapple',
  'sevencardstud',
  'razz',
  'badugi',
  'hearts',
  'spades',
  'twotenjack',
  'napoleon',
  'ohhell',
  'memory',
  'klondike',
  'freecell',
  'baccarat',
  'crazyeights',
  'ginrummy',
  'canasta',
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
  'yukon',
  'russiansolitaire',
  'scorpion',
  'accordion',
  'sevenbridge',
  'trash',
  'whist',
  'letitride',
  'pokersquares',
  'pageone',
  'reddog',
  'president',
  'cassino',
  'spanish21',
  'spiteandmalice',
  'skat',
  'shithead',
  'nertz',
  'slapjack',
  'egyptianratscrew',
  'bakersdozen',
  'tonk',
  'casinowar',
  'pitch',
  'dragontiger',
  'blackjackswitch',
  'montecarlo',
  'contractrummy',
  'ultimatetexasholdem',
  'crescent',
  'mississippistud',
  'belote',
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
