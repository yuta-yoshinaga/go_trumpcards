// API client for faro. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FaroResponse } from '../../types/card';
import { gameExec } from '../gameExec';

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
