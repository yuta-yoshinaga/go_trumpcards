// API client for gostop. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GoStopResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Go-Stop (Godori / ゴーストップ) game settings. */
export interface GoStopConfigInput {
  cpuDifficulty?: number;
  /** Target cumulative score that ends the match. */
  targetScore?: number;
}

/** Commands accepted by the Go-Stop /gostop/exec endpoint. */
export type GoStopCommand = 'reset' | 'play' | 'go' | 'stop' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Go-Stop (Godori / ゴーストップ) /gostop/exec endpoint.
 *
 * Go-Stop is the Korean sibling of Koi-Koi, a 2-player hanafuda capture game
 * with a Korean scoring breakdown (gwang/godori/tti/yeol/pi) plus Go/Stop.
 *   - `play` → `{ cardIndex, fieldIndex? }` (fieldIndex disambiguates a 2-way
 *     field match; omit when there is at most one match)
 *   - `go` → continue the round after reaching the target score
 *   - `stop` → bank the points and end the round
 *   - `nextround` → deal the next round
 *   - `reset` → `{ config }`
 *   - `hint` / `log` carry no extra fields.
 */
export const gostopApi = {
  exec: (
    command: GoStopCommand,
    opts?: {
      cardIndex?: number;
      fieldIndex?: number;
      config?: GoStopConfigInput;
    },
  ) =>
    gameExec<GoStopResponse>('gostop', {
      command,
      cardIndex: opts?.cardIndex,
      fieldIndex: opts?.fieldIndex,
      config: opts?.config,
    }),
};
