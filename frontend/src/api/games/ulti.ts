// API client for ulti. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { UltiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Ulti (Ultimo) game settings. */
export interface UltiConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Ulti /ulti/exec endpoint. */
export type UltiCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Ulti (Ultimo) /ulti/exec endpoint.
 *
 * Ulti is a 3-player Hungarian contract trick-taker on a 32-card deck. The human
 * (seat 0) is always the declarer versus a 2-CPU defending coalition.
 *   - `bid` → `{ contract, trumpSuit? }` (contract 'party'|'betli'|'durchmarsch';
 *     trumpSuit 1=♠ 2=♣ 3=♥ 4=♦, meaningful only for a Party contract)
 *   - `discard` → `{ cardIndices }` (the 2 talon cards to discard)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const ultiApi = {
  exec: (
    command: UltiCommand,
    opts?: {
      contract?: string;
      trumpSuit?: number;
      cardIndices?: number[];
      cardIndex?: number;
      config?: UltiConfigInput;
    },
  ) =>
    gameExec<UltiResponse>('ulti', {
      command,
      contract: opts?.contract,
      trumpSuit: opts?.trumpSuit,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
