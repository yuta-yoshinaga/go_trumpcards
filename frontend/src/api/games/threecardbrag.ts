// API client for threecardbrag. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { ThreeCardBragResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Three Card Brag game settings. */
export interface ThreeCardBragConfigInput {
  cpuDifficulty?: number;
  ante?: number;
  startingChips?: number;
}

/** Commands accepted by the Three Card Brag /threecardbrag/exec endpoint. */
export type ThreeCardBragCommand =
  | 'reset'
  | 'see'
  | 'bet'
  | 'raise'
  | 'fold'
  | 'show'
  | 'next'
  | 'hint'
  | 'log'
  | 'config';

/**
 * API client for the Three Card Brag /threecardbrag/exec endpoint.
 *
 * Three Card Brag is a 4-player British vying game (poker ancestor) with chips
 * and a pot. On the human's turn: `see` (reveal, Blind→Seen), `bet` (call the
 * stake), `raise` (with `raiseStake`), `fold`, or `show` (when allowed). `next`
 * advances to the following deal; `reset` / `config` apply the config.
 *   - `raise` → `{ raiseStake: number }`
 *   - `reset` / `config` → `{ config }`
 *   - `see` / `bet` / `fold` / `show` / `next` / `hint` / `log` carry no extra fields.
 */
export const threeCardBragApi = {
  exec: (
    command: ThreeCardBragCommand,
    opts?: {
      raiseStake?: number;
      config?: ThreeCardBragConfigInput;
    },
  ) =>
    gameExec<ThreeCardBragResponse>('threecardbrag', {
      command,
      raiseStake: opts?.raiseStake,
      config: opts?.config,
    }),
};
