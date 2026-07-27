// API client for ombre. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { OmbreResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Ombre (Hombre) game settings. */
export interface OmbreConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Ombre /ombre/exec endpoint. */
export type OmbreCommand = 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Ombre (Hombre) /ombre/exec endpoint.
 *
 * Ombre is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck. A Bid phase (pass / entrar / solo) plus a chosen trump suit decides the
 * Ombre, who then plays alone against the coalition of the other two.
 *   - `bid` → `{ bid, trumpSuit? }` (bid 0=pass, 1=entrar, 2=solo; trumpSuit
 *     1=♠ 2=♣ 3=♥ 4=♦, sent with a winning entrar/solo)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const ombreApi = {
  exec: (
    command: OmbreCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      trumpSuit?: number;
      config?: OmbreConfigInput;
    },
  ) =>
    gameExec<OmbreResponse>('ombre', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};
