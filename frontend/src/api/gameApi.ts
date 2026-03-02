import type {
  BlackJackResponse,
  DaifugoConfigInput,
  DaifugoResponse,
  DoubtConfig,
  DoubtResponse,
  HoldemResponse,
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
      | 'togglecounting',
    amount?: number,
    config?: BlackJackConfigInput,
  ) => postJson<BlackJackResponse>('/blackjack/exec', { command, amount, sessionId, ...config }),
};

export const pokerApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin',
    indices?: number[],
    amount?: number,
    cpuCount?: number,
    jokerCount?: number,
  ) => postJson<PokerResponse>('/poker/exec', { command, indices, amount, cpuCount, jokerCount, sessionId }),
};

export const oldmaidApi = {
  exec: (
    command: 'reset' | 'draw' | 'shuffle' | 'reorder',
    drawIdx?: number,
    mode?: number,
    cpuPlacementStrategy?: boolean,
    reorderIndices?: number[],
  ) =>
    postJson<OldMaidResponse>('/oldmaid/exec', {
      command,
      drawIdx,
      mode,
      cpuPlacementStrategy,
      reorderIndices,
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
    }),
};

export interface SevensConfigInput {
  tunnelEnabled?: boolean;
  jokerCount?: number;
  cpuStrategy?: boolean;
  maxPasses?: number;
  noJokerFinish?: boolean;
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
}

export const holdemApi = {
  exec: (
    command: 'reset' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin',
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
