// API client for whist. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WhistConfig, WhistResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Whist /whist/exec endpoint. */
export const whistApi = {
  /**
   * **The card index arrives in the second slot, not the first.**
   *
   * `useTrickGameBase` dispatches `(command, arg1?, arg2?, config?)` — the card
   * index in `arg2`, the reset config in `config`. Reading `arg1` as the card
   * index left `cardIndex` undefined on every play *and* shifted the reset
   * config into the wrong slot, so the server refused the move and **the game
   * could not be played from the Web GUI at all** (#6227). The first slot is
   * unused here, matching `gaigelApi`.
   */
  exec: (
    command: 'reset' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
    _unused: number | undefined,
    cardIndex?: number,
    config?: Partial<WhistConfig>,
  ) => gameExec<WhistResponse>('whist', { command, cardIndex, config }),
};
