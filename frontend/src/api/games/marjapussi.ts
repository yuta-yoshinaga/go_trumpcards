// API client for marjapussi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MarjapussiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Marjapussi game settings. */
export interface MarjapussiConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Marjapussi /marjapussi/exec endpoint. */
export type MarjapussiCommand = 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Marjapussi /marjapussi/exec endpoint.
 *
 * Marjapussi is a Finnish 4-player 2-vs-2 partnership trick-taker played with a 36-card deck.
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const marjapussiApi = {
  exec: (
    command: MarjapussiCommand,
    opts?: {
      cardIndex?: number;
      config?: MarjapussiConfigInput;
    },
  ) =>
    gameExec<MarjapussiResponse>('marjapussi', {
      command,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
