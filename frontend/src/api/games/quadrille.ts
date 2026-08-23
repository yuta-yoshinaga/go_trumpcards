// API client for quadrille. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { QuadrilleResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Quadrille game settings. */
export interface QuadrilleConfigInput {
  cpuDifficulty?: number;
  targetRounds?: number;
}

/** Commands accepted by the Quadrille /quadrille/exec endpoint. */
export type QuadrilleCommand = 'reset' | 'bid' | 'king' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Quadrille /quadrille/exec endpoint.
 *
 * Quadrille is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck. A Bid phase (pass / entrar / solo) plus a chosen trump suit decides the
 * Quadrille, who then calls a king to find a hidden partner.
 *   - `bid` → `{ bid, trumpSuit? }` (bid 0=pass, 1=entrar, 2=solo; trumpSuit
 *     1=♠ 2=♣ 3=♥ 4=♦, sent with a winning entrar/solo)
 *   - `king` → `{ kingSuit }` (1=♠ 2=♣ 3=♥ 4=♦; must be a king the bidder does
 *     not hold)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const quadrilleApi = {
  exec: (
    command: QuadrilleCommand,
    opts?: {
      cardIndex?: number;
      bid?: number;
      trumpSuit?: number;
      kingSuit?: number;
      config?: QuadrilleConfigInput;
    },
  ) =>
    gameExec<QuadrilleResponse>('quadrille', {
      command,
      cardIndex: opts?.cardIndex,
      bid: opts?.bid,
      trumpSuit: opts?.trumpSuit,
      kingSuit: opts?.kingSuit,
      config: opts?.config,
    }),
};
