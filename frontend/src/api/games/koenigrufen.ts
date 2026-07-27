// API client for koenigrufen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KoenigrufenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

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
