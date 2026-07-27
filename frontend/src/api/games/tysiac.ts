// API client for tysiac. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { TysiacResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Tysiąc (Thousand) game settings. */
export interface TysiacConfigInput {
  cpuDifficulty?: number;
  targetPoints?: number;
}

/** Commands accepted by the Tysiąc /tysiac/exec endpoint. */
export type TysiacCommand = 'reset' | 'bid' | 'discard' | 'play' | 'next' | 'nextround' | 'hint' | 'log';

/**
 * API client for the Tysiąc (Thousand) /tysiac/exec endpoint.
 *
 * Tysiąc is a Polish 3-player 24-card trump trick-taker with a Bid phase, a
 * Talon exchange phase, and marriage (K+Q) declarations during play.
 *   - `bid` → `{ raise: boolean }` (raise=true means +10, false means pass)
 *   - `discard` → `{ cardIndex }` (talon exchange: the human Declarer gives one
 *     card to an opponent; called once per opponent — twice total)
 *   - `play` → `{ cardIndex }`
 *   - `reset` → `{ config }`
 *   - `next` / `nextround` / `hint` / `log` carry no extra fields.
 */
export const tysiacApi = {
  exec: (
    command: TysiacCommand,
    opts?: {
      cardIndex?: number;
      raise?: boolean;
      config?: TysiacConfigInput;
    },
  ) =>
    gameExec<TysiacResponse>('tysiac', {
      command,
      cardIndex: opts?.cardIndex,
      raise: opts?.raise,
      config: opts?.config,
    }),
};
