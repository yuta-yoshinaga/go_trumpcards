// API client for schafkopf. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SchafkopfResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Schafkopf game settings. */
export interface SchafkopfConfigInput {
  cpuDifficulty?: number;
  baseChips?: number;
  startChips?: number;
  targetChips?: number;
}

/** Commands accepted by the Schafkopf /schafkopf/exec endpoint. */
export type SchafkopfCommand = 'reset' | 'pick' | 'call' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/** Contract the picker declares (0=Rufspiel 1=Wenz 2=Solo). */
export type SchafkopfContract = 0 | 1 | 2;

/**
 * API client for the Schafkopf /schafkopf/exec endpoint.
 *
 * The multi-phase flow maps each command to its own body field:
 *   - `pick` → `{ pick: boolean, contract?, soloSuit? }` (declare or pass)
 *     `contract` defaults to Rufspiel; `soloSuit` is required for Solo.
 *   - `call` → `{ callSuit: number }` (1=♠ 2=♣ 4=♦; ♥ is trump, so never callable)
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const schafkopfApi = {
  exec: (
    command: SchafkopfCommand,
    opts?: {
      pick?: boolean;
      contract?: SchafkopfContract;
      soloSuit?: number;
      callSuit?: number;
      cardIndex?: number;
      config?: SchafkopfConfigInput;
    },
  ) =>
    gameExec<SchafkopfResponse>('schafkopf', {
      command,
      pick: opts?.pick,
      contract: opts?.contract,
      soloSuit: opts?.soloSuit,
      callSuit: opts?.callSuit,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
