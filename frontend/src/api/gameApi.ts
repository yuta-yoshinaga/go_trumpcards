import type {
  ActionLogResponse,
  BaccaratResponse,
  BlackJackResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DoubtConfig,
  DoubtResponse,
  FreeCellResponse,
  HeartsResponse,
  HoldemResponse,
  KlondikeResponse,
  MemoryResponse,
  OldMaidResponse,
  OmahaResponse,
  PokerResponse,
  SevensResponse,
} from '../types/card';

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

export interface BlackJackConfigInput {
  dealerHitsSoft17?: boolean;
  cpuPlayerCount?: number;
  countingEnabled?: boolean;
  doubleAfterSplit?: boolean;
  countingSystem?: number;
  deckPenetration?: number;
  surrenderRule?: number;
}

export interface BlackJackBetOptions {
  perfectPairsBet?: number;
  twentyOnePlus3Bet?: number;
  handCount?: number;
}

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

export interface PokerConfigInput {
  cpuCount?: number;
  jokerCount?: number;
  bettingLimit?: number;
  isLowball?: boolean;
}

export const pokerApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin' | 'odds',
    indices?: number[],
    amount?: number,
    config?: PokerConfigInput,
  ) => gameExec<PokerResponse>('poker', { command, indices, amount, ...config }),
};

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
    }),
};

export const daifugoApi = {
  exec: (command: 'reset' | 'play' | 'sort', indices?: number[], config?: DaifugoConfigInput, sortMode?: number) =>
    gameExec<DaifugoResponse>('daifugo', { command, indices, config, sortMode }),
};

export const doubtApi = {
  exec: (
    command: 'reset' | 'play' | 'doubt' | 'skip',
    cardIndices?: number[],
    claimedValue?: number,
    doubterIndices?: number[],
    config?: DoubtConfig,
    humanPlayMs?: number,
  ) =>
    gameExec<DoubtResponse>('doubt', {
      command,
      cardIndices,
      claimedValue,
      doubterIndices,
      humanPlayMs,
      doubtWindowSec: config?.doubtWindowSec,
      cpuMemoryLevel: config?.cpuMemoryLevel,
      penaltyDrawLimit: config?.penaltyDrawLimit,
      cpuHesitationEnabled: config?.cpuHesitationEnabled,
      cpuMetaAI: config?.cpuMetaAI,
    }),
};

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
}

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
  ) =>
    gameExec<HoldemResponse>('holdem', {
      command,
      amount,
      ...config,
    }),
};

export type OmahaConfigInput = HoldemConfigInput;

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
  ) =>
    gameExec<OmahaResponse>('omaha', {
      command,
      amount,
      ...config,
    }),
};

export interface HeartsConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
  omnibusJD?: boolean;
}

export const heartsApi = {
  exec: (
    command: 'reset' | 'pass' | 'play' | 'next' | 'nextround',
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

export interface MemoryConfigInput {
  cpuDifficulty?: number;
}

export const memoryApi = {
  exec: (command: 'reset' | 'flip' | 'next' | 'log', position?: number, config?: MemoryConfigInput) =>
    gameExec<MemoryResponse>('memory', {
      command,
      position,
      config,
    }),
};

export interface KlondikeMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

export interface KlondikeConfigInput {
  drawCount?: number;
  scoringMode?: number;
}

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

export interface FreeCellMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

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

export const baccaratApi = {
  exec: (
    command: 'reset' | 'bet' | 'log' | 'clearhistory',
    amount?: number,
    betType?: number,
    playerPairBet?: number,
    bankerPairBet?: number,
  ) => gameExec<BaccaratResponse>('baccarat', { command, amount, betType, playerPairBet, bankerPairBet }),
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
  'memory',
  'klondike',
  'freecell',
  'baccarat',
] as const;
type Game = (typeof games)[number];

export const actionLogApi: { [K in Game]: () => Promise<ActionLogResponse> } = games.reduce(
  (acc, game) => {
    acc[game] = () => gameExec<ActionLogResponse>(game, { command: 'log' });
    return acc;
  },
  {} as { [K in Game]: () => Promise<ActionLogResponse> },
);
