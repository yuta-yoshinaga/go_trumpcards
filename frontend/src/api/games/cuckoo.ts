// API client for cuckoo. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CuckooResponse } from '../../types/card';
import { gameExec } from '../gameExec';

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
