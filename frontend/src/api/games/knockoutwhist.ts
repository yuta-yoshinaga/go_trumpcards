// API client for knockoutwhist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KnockoutWhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Knockout Whist game settings (CPU difficulty only — no target points). */
export interface KnockoutWhistConfigInput {
  cpuDifficulty?: number;
}

/** Commands accepted by the Knockout Whist /knockoutwhist/exec endpoint. */
export type KnockoutWhistCommand = 'reset' | 'play' | 'selecttrump' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Knockout Whist /knockoutwhist/exec endpoint.
 *
 * Knockout Whist is a British play-only survival trick-taker for 4 players on a
 * 52-card deck. Each round deals one fewer card; the previous round's winner's
 * longest suit becomes trump (auto). Must-follow, Ace-high. A player who wins
 * zero tricks in a round must spend a Dogbone token to survive, or is
 * eliminated. Last player standing wins.
 *   - `play` → `{ cardIndex: number }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const knockoutWhistApi = {
  exec: (
    command: KnockoutWhistCommand,
    opts?: {
      cardIndex?: number;
      trumpSuit?: number;
      config?: KnockoutWhistConfigInput;
    },
  ) =>
    gameExec<KnockoutWhistResponse>('knockoutwhist', {
      command,
      cardIndex: opts?.cardIndex,
      trumpSuit: opts?.trumpSuit,
      config: opts?.config,
    }),
};
