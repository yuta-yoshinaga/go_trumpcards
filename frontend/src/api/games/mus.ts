// API client for mus. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MusResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Mus game settings. */
export interface MusConfigInput {
  cpuDifficulty?: number;
  targetAmarrakos?: number;
}

/** Commands accepted by the Mus /mus/exec endpoint. */
export type MusCommand = 'reset' | 'mus' | 'discard' | 'bet' | 'next' | 'hint' | 'log';

/**
 * API client for the Mus /mus/exec endpoint.
 *
 * Mus is a Basque vying (betting) game, so each command maps to its own body
 * field rather than a card-play action:
 *   - `mus` → `{ mus: boolean }` (true = call Mus / exchange, false = cut and bet)
 *   - `discard` → `{ discardIndices: number[] }` (cards to exchange; empty keeps all)
 *   - `bet` → `{ betAction: number, betAmount?: number }`
 *     (betAction: 0=paso 1=envido 2=ordago 3=quiero 4=noquiero)
 *   - `reset` → `{ config }`
 *   - `next` / `hint` / `log` carry no extra fields.
 */
export const musApi = {
  exec: (
    command: MusCommand,
    opts?: {
      mus?: boolean;
      discardIndices?: number[];
      betAction?: number;
      betAmount?: number;
      config?: MusConfigInput;
    },
  ) =>
    gameExec<MusResponse>('mus', {
      command,
      mus: opts?.mus,
      discardIndices: opts?.discardIndices,
      betAction: opts?.betAction,
      betAmount: opts?.betAmount,
      config: opts?.config,
    }),
};
