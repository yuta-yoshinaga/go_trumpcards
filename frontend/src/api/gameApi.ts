import type {
  ActionLogResponse,
  BaccaratResponse,
  BlackJackResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DoubtConfig,
  DoubtResponse,
  HeartsResponse,
  HoldemResponse,
  KlondikeResponse,
  MemoryResponse,
  OldMaidResponse,
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
  ) => postJson<BlackJackResponse>('/blackjack/exec', { command, amount, sessionId, ...config, ...betOptions }),
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
  ) => postJson<PokerResponse>('/poker/exec', { command, indices, amount, ...config, sessionId }),
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
    postJson<OldMaidResponse>('/oldmaid/exec', {
      command,
      drawIdx,
      mode,
      cpuPlacementStrategy,
      reorderIndices,
      cpuMemoryAI,
      cpuHesitationEnabled,
      cpuMetaAI,
      sessionId,
    }),
};

export const daifugoApi = {
  exec: (command: 'reset' | 'play' | 'sort', indices?: number[], config?: DaifugoConfigInput, sortMode?: number) =>
    postJson<DaifugoResponse>('/daifugo/exec', { command, indices, config, sortMode, sessionId }),
};

export const doubtApi = {
  exec: (
    command: 'reset' | 'play' | 'doubt' | 'skip',
    cardIndices?: number[],
    claimedValue?: number,
    doubterIndices?: number[],
    config?: DoubtConfig,
  ) =>
    postJson<DoubtResponse>('/doubt/exec', {
      command,
      cardIndices,
      claimedValue,
      doubterIndices,
      sessionId,
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
    postJson<SevensResponse>('/sevens/exec', {
      command,
      index,
      jokerTargetSuit,
      jokerTargetValue,
      sessionId,
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
    postJson<HoldemResponse>('/holdem/exec', {
      command,
      amount,
      sessionId,
      ...config,
    }),
};

export interface HeartsConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

export const heartsApi = {
  exec: (
    command: 'reset' | 'pass' | 'play' | 'next' | 'nextround',
    cardIndices?: number[],
    cardIndex?: number,
    config?: HeartsConfigInput,
  ) =>
    postJson<HeartsResponse>('/hearts/exec', {
      command,
      cardIndices,
      cardIndex,
      sessionId,
      config,
    }),
};

export interface MemoryConfigInput {
  cpuDifficulty?: number;
}

export const memoryApi = {
  exec: (command: 'reset' | 'flip' | 'next' | 'log', position?: number, config?: MemoryConfigInput) =>
    postJson<MemoryResponse>('/memory/exec', {
      command,
      position,
      sessionId,
      config,
    }),
};

export interface KlondikeMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

export const klondikeApi = {
  exec: (
    command: 'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log',
    from?: KlondikeMoveZone,
    to?: KlondikeMoveZone,
  ) =>
    postJson<KlondikeResponse>('/klondike/exec', {
      command,
      from,
      to,
      sessionId,
    }),
};

export const baccaratApi = {
  exec: (command: 'reset' | 'bet' | 'log', amount?: number, betType?: number) =>
    postJson<BaccaratResponse>('/baccarat/exec', { command, amount, betType, sessionId }),
};

function fetchLog(url: string): Promise<ActionLogResponse> {
  return postJson<ActionLogResponse>(url, { command: 'log', sessionId });
}

const games = [
  'blackjack',
  'poker',
  'oldmaid',
  'daifugo',
  'sevens',
  'doubt',
  'holdem',
  'hearts',
  'memory',
  'klondike',
  'baccarat',
] as const;
type Game = (typeof games)[number];

export const actionLogApi: { [K in Game]: () => Promise<ActionLogResponse> } = games.reduce(
  (acc, game) => {
    acc[game] = () => fetchLog(`/${game}/exec`);
    return acc;
  },
  {} as { [K in Game]: () => Promise<ActionLogResponse> },
);
