import type {
  ActionLogResponse,
  BaccaratResponse,
  BlackJackResponse,
  CrazyEightsResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DoubtConfig,
  DoubtResponse,
  FreeCellResponse,
  GinRummyResponse,
  HeartsResponse,
  HoldemResponse,
  IndianPokerResponse,
  KlondikeResponse,
  MemoryResponse,
  NapoleonResponse,
  OldMaidResponse,
  OmahaResponse,
  PokerResponse,
  SevensResponse,
  SpadesResponse,
  SpiderResponse,
} from '../types/card';

/** Unique session identifier for correlating API requests. */
export const sessionId: string = crypto.randomUUID();

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
  return postJson<T>(`/${game}/exec`, { ...body, sessionId });
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

const games = [
  'blackjack',
  'poker',
  'oldmaid',
  'daifugo',
  'sevens',
  'doubt',
  'holdem',
  'omaha',
  'hearts',
  'spades',
  'napoleon',
  'memory',
  'klondike',
  'freecell',
  'baccarat',
  'crazyeights',
  'ginrummy',
  'spider',
  'indianpoker',
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
