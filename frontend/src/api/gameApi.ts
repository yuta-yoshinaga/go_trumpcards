import type {
  ActionLogResponse,
  BaccaratResponse,
  BlackJackResponse,
  BridgeResponse,
  CrazyEightsResponse,
  CribbageResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DoubtConfig,
  DoubtResponse,
  EuchreResponse,
  FreeCellResponse,
  GinRummyResponse,
  HeartsResponse,
  HoldemResponse,
  IndianPokerResponse,
  KlondikeResponse,
  MemoryResponse,
  NapoleonResponse,
  OhHellResponse,
  OldMaidResponse,
  OmahaResponse,
  PineappleResponse,
  PokerResponse,
  PyramidResponse,
  SevensResponse,
  ShortDeckResponse,
  SpadesResponse,
  SpiderResponse,
  ThreeCardResponse,
  TriPeaksResponse,
  VideoPokerResponse,
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
  baccarat: WORKER_CASINO,
  poker: WORKER_CASINO,
  holdem: WORKER_CASINO,
  omaha: WORKER_CASINO,
  shortdeck: WORKER_CASINO,
  indianpoker: WORKER_CASINO,
  videopoker: WORKER_CASINO,
  deuceswild: WORKER_CASINO,
  jokerpoker: WORKER_CASINO,
  threecard: WORKER_CASINO,
  pineapple: WORKER_CASINO,
  hearts: WORKER_CLASSIC,
  spades: WORKER_CLASSIC,
  euchre: WORKER_CLASSIC,
  bridge: WORKER_CLASSIC,
  napoleon: WORKER_CLASSIC,
  ohhell: WORKER_CLASSIC,
  oldmaid: WORKER_CLASSIC,
  doubt: WORKER_CLASSIC,
  daifugo: WORKER_CLASSIC,
  sevens: WORKER_CLASSIC,
  crazyeights: WORKER_CLASSIC,
  klondike: WORKER_SOLO,
  freecell: WORKER_SOLO,
  spider: WORKER_SOLO,
  pyramid: WORKER_SOLO,
  tripeaks: WORKER_SOLO,
  memory: WORKER_SOLO,
  ginrummy: WORKER_SOLO,
  cribbage: WORKER_SOLO,
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

/** API client for the BlackJack /blackjack/exec endpoint. */
export const blackjackApi = {
  exec: (
    command:
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
      | 'setsurrenderrule',
    amount?: number,
    config?: BlackJackConfigInput,
    betOptions?: BlackJackBetOptions,
  ) => gameExec<BlackJackResponse>('blackjack', { command, amount, ...config, ...betOptions }),
};

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

/** API client for the Texas Hold'em /holdem/exec endpoint. */
export const holdemApi = {
  exec: (
    command:
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
      | 'show',
    amount?: number,
    config?: HoldemConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<HoldemResponse>('holdem', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** Configuration options for Pineapple Poker (extends Hold'em with cardIdx for discard). */
export interface PineappleConfigInput extends HoldemConfigInput {
  cardIdx?: number;
}

/** API client for the Pineapple Poker /pineapple/exec endpoint. */
export const pineappleApi = {
  exec: (
    command:
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
      | 'show'
      | 'discard',
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
      cardIdx: config?.cardIdx,
      smallBlind: config?.smallBlind,
      bigBlind: config?.bigBlind,
      tournamentMode: config?.tournamentMode,
      blindLevelHands: config?.blindLevelHands,
      blindMultiplier: config?.blindMultiplier,
      bettingLimit: config?.bettingLimit,
      tableSize: config?.tableSize,
      rebuyEnabled: config?.rebuyEnabled,
      rebuyMaxCount: config?.rebuyMaxCount,
      rebuyChips: config?.rebuyChips,
      rebuyPeriodHands: config?.rebuyPeriodHands,
      addonEnabled: config?.addonEnabled,
      addonChips: config?.addonChips,
      addonAfterHand: config?.addonAfterHand,
      cpuMetaAI: config?.cpuMetaAI,
    }),
};

/** Configuration options for Omaha Hold'em (same as Hold'em). */
export type OmahaConfigInput = HoldemConfigInput;

/** API client for the Omaha Hold'em /omaha/exec endpoint. */
export const omahaApi = {
  exec: (
    command:
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
      | 'show',
    amount?: number,
    config?: OmahaConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<OmahaResponse>('omaha', {
      command,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};

/** Configuration options for Short Deck Hold'em (same as Hold'em). */
export type ShortDeckConfigInput = HoldemConfigInput;

/** API client for the Short Deck Hold'em /shortdeck/exec endpoint. */
export const shortdeckApi = {
  exec: (
    command:
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
      | 'show',
    amount?: number,
    config?: ShortDeckConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<ShortDeckResponse>('shortdeck', {
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

/** API client for the Spades /spades/exec endpoint. */
export const spadesApi = {
  exec: (
    command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint',
    bid?: number,
    cardIndex?: number,
    config?: SpadesConfigInput,
  ) =>
    gameExec<SpadesResponse>('spades', {
      command,
      bid,
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
export const ohHellApi = {
  exec: (
    command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint',
    bid?: number,
    cardIndex?: number,
    config?: OhHellConfigInput,
  ) =>
    gameExec<OhHellResponse>('ohhell', {
      command,
      bid,
      cardIndex,
      config,
    }),
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
export const klondikeApi = {
  exec: (
    command: 'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo',
    from?: KlondikeMoveZone,
    to?: KlondikeMoveZone,
    config?: KlondikeConfigInput,
  ) =>
    gameExec<KlondikeResponse>('klondike', {
      command,
      from,
      to,
      config,
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
export const freecellApi = {
  exec: (
    command: 'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo',
    from?: FreeCellMoveZone,
    to?: FreeCellMoveZone,
  ) =>
    gameExec<FreeCellResponse>('freecell', {
      command,
      from,
      to,
    }),
};

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
export const spiderApi = {
  exec: (
    command: 'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo',
    from?: SpiderMoveZone,
    to?: SpiderMoveZone,
    config?: SpiderConfigInput,
  ) =>
    gameExec<SpiderResponse>('spider', {
      command,
      from,
      to,
      config,
    }),
};

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
    command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo',
    card1?: PyramidRemoveCard,
    card2?: PyramidRemoveCard,
  ) => gameExec<PyramidResponse>('pyramid', { command, card1, card2 }),
};

/** API client for the TriPeaks /tripeaks/exec endpoint. */
export const tripeaksApi = {
  exec: (command: 'reset' | 'draw' | 'remove' | 'giveup' | 'hint' | 'log' | 'undo', row?: number, col?: number) =>
    gameExec<TriPeaksResponse>('tripeaks', { command, row, col }),
};

/** API client for the Video Poker /videopoker/exec endpoint. */
export const videopokerApi = {
  exec: (command: 'reset' | 'bet' | 'hold' | 'log', amount?: number, indices?: number[]) =>
    gameExec<VideoPokerResponse>('videopoker', { command, amount, indices }),
};

/** API client for the Deuces Wild /deuceswild/exec endpoint. */
export const deuceswildApi = {
  exec: (command: 'reset' | 'bet' | 'hold' | 'log', amount?: number, indices?: number[]) =>
    gameExec<VideoPokerResponse>('deuceswild', { command, amount, indices }),
};

/** API client for the Joker Poker /jokerpoker/exec endpoint. */
export const jokerpokerApi = {
  exec: (command: 'reset' | 'bet' | 'hold' | 'log', amount?: number, indices?: number[]) =>
    gameExec<VideoPokerResponse>('jokerpoker', { command, amount, indices }),
};

const games = [
  'blackjack',
  'poker',
  'oldmaid',
  'daifugo',
  'sevens',
  'doubt',
  'holdem',
  'omaha',
  'shortdeck',
  'pineapple',
  'hearts',
  'spades',
  'napoleon',
  'ohhell',
  'memory',
  'klondike',
  'freecell',
  'baccarat',
  'crazyeights',
  'ginrummy',
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
