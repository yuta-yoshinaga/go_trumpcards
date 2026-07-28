// API client for kemps. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KempsResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Kemps game settings. */
export interface KempsConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Kemps /kemps/exec endpoint. */
export type KempsCommand = 'reset' | 'swap' | 'pass' | 'signal' | 'kemps' | 'counter' | 'next' | 'log';

/**
 * API client for the Kemps /kemps/exec endpoint.
 *
 * Kemps is a 4-player, 2-team matching game. On the Exchange phase the human
 * swaps one hand card for a field card (`swap` → `{ handIndex, fieldIndex }`)
 * or skips with `pass`. The human sets a secret signal type with `signal` →
 * `{ signalType }` (0=Sound, 1=Blink). When a team completes four of a kind the
 * Declare window opens: `kemps` declares "Kemps!" and `counter` →
 * `{ targetSeat }` declares "Counter-Kemps!" against an opponent seat. `next`
 * advances to the following round; `reset` applies the config; `log` fetches
 * the action log.
 *   - `swap` → `{ handIndex: number, fieldIndex: number }`
 *   - `signal` → `{ signalType: number }`
 *   - `counter` → `{ targetSeat: number }`
 *   - `reset` → `{ config }`
 *   - `pass` / `kemps` / `next` / `log` carry no extra fields.
 */
export const kempsApi = {
  exec: (
    command: KempsCommand,
    opts?: {
      handIndex?: number;
      fieldIndex?: number;
      signalType?: number;
      targetSeat?: number;
      config?: KempsConfigInput;
    },
  ) =>
    gameExec<KempsResponse>('kemps', {
      command,
      handIndex: opts?.handIndex,
      fieldIndex: opts?.fieldIndex,
      signalType: opts?.signalType,
      targetSeat: opts?.targetSeat,
      config: opts?.config,
    }),
};
