// API client for gleek. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GleekResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Gleek game settings. */
export interface GleekConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Gleek /gleek/exec endpoint. */
export type GleekCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Gleek /gleek/exec endpoint.
 *
 * Gleek runs four scoring stages in one deal; only two of them take input.
 *   - `bid` → `{ bid }` (0 drops out; anything else must equal the response's
 *     `nextBidAmount`, since raises go up in fixed steps)
 *   - `discard` → `{ discardIndices }` (exactly `discardCount` cards, thrown by
 *     the buyer)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 *
 * The ruff and the melds are readings of the settled hands rather than
 * decisions, so the server scores them on entering the play phase and reports
 * them on the response.
 */
export const gleekApi = {
  exec: (
    command: GleekCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      discardIndices?: number[];
      config?: GleekConfigInput;
    },
  ) =>
    gameExec<GleekResponse>('gleek', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      discardIndices: opts?.discardIndices,
      config: opts?.config,
    }),
};
