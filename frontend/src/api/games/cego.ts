// API client for cego. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CegoResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Cego (チェゴ) game settings. */
export interface CegoConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the Cego /cego/exec endpoint. */
export type CegoCommand =
  | 'reset'
  | 'bid'
  | 'pass'
  | 'contract'
  | 'discard'
  | 'play'
  | 'next'
  | 'nextround'
  | 'hint'
  | 'log';

/**
 * API client for the Cego (チェゴ) /cego/exec endpoint.
 *
 * Cego is a 4-player Baden tarock trick-taker on the 54-card tarock deck. The
 * human is seat 0.
 *   - `bid` → `{ bid }` (bid string 'play')
 *   - `pass` → carries no extra fields (pass the auction)
 *   - `contract` → `{ contract }` ('cego' or 'handspiel')
 *   - `discard` → `{ cardIndices }` (the single card to KEEP in the Cego exchange)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const cegoApi = {
  exec: (
    command: CegoCommand,
    opts?: {
      bid?: string;
      contract?: string;
      cardIndices?: number[];
      cardIndex?: number;
      config?: CegoConfigInput;
    },
  ) =>
    gameExec<CegoResponse>('cego', {
      command,
      bid: opts?.bid,
      contract: opts?.contract,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
