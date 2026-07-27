// API client for frenchtarot. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FrenchTarotResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for French Tarot (フレンチタロット) game settings. */
export interface FrenchTarotConfigInput {
  cpuDifficulty?: number;
  targetDeals?: number;
}

/** Commands accepted by the French Tarot /frenchtarot/exec endpoint. */
export type FrenchTarotCommand = 'reset' | 'bid' | 'pass' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the French Tarot (フレンチタロット) /frenchtarot/exec endpoint.
 *
 * French Tarot is a 4-player trick-taker on the 78-card tarot deck. The human is
 * seat 0.
 *   - `bid` → `{ bid }` (contract string 'petite'|'garde'|'gardesans'|'gardecontre')
 *   - `pass` → carries no extra fields (pass the auction)
 *   - `discard` → `{ cardIndices }` (the 6 écart cards to bury; Petite/Garde only)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const frenchtarotApi = {
  exec: (
    command: FrenchTarotCommand,
    opts?: {
      bid?: string;
      cardIndices?: number[];
      cardIndex?: number;
      config?: FrenchTarotConfigInput;
    },
  ) =>
    gameExec<FrenchTarotResponse>('frenchtarot', {
      command,
      bid: opts?.bid,
      cardIndices: opts?.cardIndices,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
