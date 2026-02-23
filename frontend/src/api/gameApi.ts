import type { BlackJackResponse, DaifugoResponse, OldMaidResponse, PokerResponse, SevensResponse } from '../types/card';

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

export const blackjackApi = {
  exec: (
    command: 'reset' | 'hit' | 'stand' | 'bet' | 'doubledown' | 'split' | 'insurance' | 'declineinsurance',
    amount?: number,
  ) => postJson<BlackJackResponse>('/blackjack/exec', { command, amount, sessionId }),
};

export const pokerApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'bet' | 'call' | 'raise' | 'fold' | 'check',
    indices?: number[],
    amount?: number,
  ) => postJson<PokerResponse>('/poker/exec', { command, indices, amount, sessionId }),
};

export const oldmaidApi = {
  exec: (command: 'reset' | 'draw', drawIdx?: number) =>
    postJson<OldMaidResponse>('/oldmaid/exec', { command, drawIdx, sessionId }),
};

export const daifugoApi = {
  exec: (command: 'reset' | 'play', indices?: number[]) =>
    postJson<DaifugoResponse>('/daifugo/exec', { command, indices, sessionId }),
};

export interface SevensConfigInput {
  tunnelEnabled?: boolean;
  jokerCount?: number;
  cpuStrategy?: boolean;
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
