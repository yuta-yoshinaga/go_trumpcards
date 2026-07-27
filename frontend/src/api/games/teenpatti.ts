// API client for teenpatti. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TeenPattiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Teen Patti game settings. */
export interface TeenPattiConfigInput {
  cpuDifficulty?: number;
  ante?: number;
  startingChips?: number;
}

/** Commands accepted by the Teen Patti /teenpatti/exec endpoint. */
export type TeenPattiCommand =
  | 'reset'
  | 'see'
  | 'bet'
  | 'raise'
  | 'fold'
  | 'show'
  | 'sideshow'
  | 'respond'
  | 'next'
  | 'hint'
  | 'log'
  | 'config';

/**
 * API client for the Teen Patti /teenpatti/exec endpoint.
 *
 * Teen Patti is the Indian variant of Three Card Brag — a 4-player vying game
 * with chips and a pot. On the human's turn: `see` (reveal, Blind→Seen), `bet`
 * (call the stake), `raise` (with `raiseStake`), `fold`, `show` (when allowed),
 * or `sideshow` (request a private hand comparison with the previous Seen
 * player). When a Side Show is requested of the human, `respond` (with
 * `accept`) accepts or declines it. `next` advances to the following deal;
 * `reset` / `config` apply the config.
 *   - `raise` → `{ raiseStake: number }`
 *   - `respond` → `{ accept: boolean }`
 *   - `reset` / `config` → `{ config }`
 *   - `see` / `bet` / `fold` / `show` / `sideshow` / `next` / `hint` / `log` carry no extra fields.
 */
export const teenPattiApi = {
  exec: (
    command: TeenPattiCommand,
    opts?: {
      raiseStake?: number;
      accept?: boolean;
      config?: TeenPattiConfigInput;
    },
  ) =>
    gameExec<TeenPattiResponse>('teenpatti', {
      command,
      raiseStake: opts?.raiseStake,
      accept: opts?.accept,
      config: opts?.config,
    }),
};
