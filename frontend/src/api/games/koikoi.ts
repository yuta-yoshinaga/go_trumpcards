// API client for koikoi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KoiKoiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Koi-Koi (こいこい) game settings. */
export interface KoiKoiConfigInput {
  cpuDifficulty?: number;
  /** Target cumulative score that ends the match. */
  targetScore?: number;
}

/** Commands accepted by the Koi-Koi /koikoi/exec endpoint. */
export type KoiKoiCommand = 'reset' | 'play' | 'koikoi' | 'stop' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Koi-Koi (こいこい) /koikoi/exec endpoint.
 *
 * Koi-Koi is a 2-player hanafuda capture game with yaku scoring.
 *   - `play` → `{ cardIndex, fieldIndex? }` (fieldIndex disambiguates a 2-way
 *     field match; omit when there is at most one match)
 *   - `koikoi` → continue the round (double the stakes) after completing a yaku
 *   - `stop` → shobu: stop and score the completed yaku
 *   - `nextround` → deal the next round
 *   - `reset` → `{ config }`
 *   - `hint` / `log` carry no extra fields.
 */
export const koikoiApi = {
  exec: (
    command: KoiKoiCommand,
    opts?: {
      cardIndex?: number;
      fieldIndex?: number;
      config?: KoiKoiConfigInput;
    },
  ) =>
    gameExec<KoiKoiResponse>('koikoi', {
      command,
      cardIndex: opts?.cardIndex,
      fieldIndex: opts?.fieldIndex,
      config: opts?.config,
    }),
};
