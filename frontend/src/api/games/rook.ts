// API client for rook. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RookResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Rook (ルーク) game settings. */
export interface RookConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/** Commands accepted by the Rook /rook/exec endpoint. */
export type RookCommand = 'reset' | 'bid' | 'pass' | 'exchange' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Rook (ルーク) /rook/exec endpoint.
 *
 * Rook is a 4-player, 2-team point-trick game on a special 57-card deck (four
 * colors ×1–14 plus the Rook bird). The human is seat 0.
 *   - `bid` → `{ bid }` (a numeric point bid, 70–120 in steps of 5)
 *   - `pass` → carries no extra fields
 *   - `exchange` → `{ discardIndices, trumpColor }` (discard 5 nest cards and
 *     declare a trump color: 1=red 2=gold 3=green 4=black)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const rookApi = {
  exec: (
    command: RookCommand,
    opts?: {
      bid?: number;
      discardIndices?: number[];
      trumpColor?: number;
      cardIndex?: number;
      config?: RookConfigInput;
    },
  ) =>
    gameExec<RookResponse>('rook', {
      command,
      bid: opts?.bid,
      discardIndices: opts?.discardIndices,
      trumpColor: opts?.trumpColor,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
