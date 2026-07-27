// API client for sheepshead. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SheepsheadResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Sheepshead game settings. */
export interface SheepsheadConfigInput {
  cpuDifficulty?: number;
  baseChips?: number;
  startChips?: number;
  targetChips?: number;
}

/** Commands accepted by the Sheepshead /sheepshead/exec endpoint. */
export type SheepsheadCommand = 'reset' | 'pick' | 'bury' | 'call' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Sheepshead /sheepshead/exec endpoint.
 *
 * The multi-phase flow maps each command to its own body field:
 *   - `pick` → `{ pick: boolean }` (take or pass the blind)
 *   - `bury` → `{ buryIndices: number[] }` (picker buries 2 cards)
 *   - `call` → `{ callSuit: number }` (1=♠ 2=♣ 3=♥)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const sheepsheadApi = {
  exec: (
    command: SheepsheadCommand,
    opts?: {
      pick?: boolean;
      buryIndices?: number[];
      callSuit?: number;
      cardIndex?: number;
      config?: SheepsheadConfigInput;
    },
  ) =>
    gameExec<SheepsheadResponse>('sheepshead', {
      command,
      pick: opts?.pick,
      buryIndices: opts?.buryIndices,
      callSuit: opts?.callSuit,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
