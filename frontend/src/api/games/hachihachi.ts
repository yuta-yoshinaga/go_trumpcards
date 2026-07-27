// API client for hachihachi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { HachiHachiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Hachi-Hachi (八八) game settings. */
export interface HachiHachiConfigInput {
  cpuDifficulty?: number;
  /** Number of rounds (deals) played before the match is settled. */
  targetRounds?: number;
}

/** Commands accepted by the Hachi-Hachi /hachihachi/exec endpoint. */
export type HachiHachiCommand = 'reset' | 'play' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Hachi-Hachi (八八) /hachihachi/exec endpoint.
 *
 * Hachi-Hachi is a 3-player hanafuda capture game with card-point scoring.
 *   - `play` → `{ cardIndex, fieldIndex? }` (fieldIndex disambiguates a 2-way
 *     field match; omit when there is at most one match)
 *   - `nextround` → deal the next round
 *   - `reset` → `{ config }`
 *   - `hint` / `log` carry no extra fields.
 */
export const hachihachiApi = {
  exec: (
    command: HachiHachiCommand,
    opts?: {
      cardIndex?: number;
      fieldIndex?: number;
      config?: HachiHachiConfigInput;
    },
  ) =>
    gameExec<HachiHachiResponse>('hachihachi', {
      command,
      cardIndex: opts?.cardIndex,
      fieldIndex: opts?.fieldIndex,
      config: opts?.config,
    }),
};
