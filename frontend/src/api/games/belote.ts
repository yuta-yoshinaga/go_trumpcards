// API client for belote. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BeloteResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Belote game configuration input shape. */
export interface BeloteConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
  dixDeDer?: number;
  enableBeloteRebelote?: boolean;
}

/** API client for the Belote /belote/exec endpoint. */
export const beloteApi = {
  /**
   * **The indexed argument arrives in the second slot, not the first.**
   *
   * `useTrickGameBase` dispatches `(command, arg1?, arg2?, config?)` and puts
   * the card index in `arg2`; Belote's own `calltrump` also passes its suit
   * there. Reading `arg1` as the card index left `cardIndex` undefined on
   * every play, so the server refused the move and **the game could not be
   * played from the Web GUI at all** (#6227). The slots are named per command
   * rather than positionally so the two cannot be confused again.
   */
  exec: (
    command: 'reset' | 'orderup' | 'pass' | 'calltrump' | 'play' | 'next' | 'nextround' | 'hint',
    _arg1: number | undefined,
    arg2?: number,
    config?: BeloteConfigInput,
  ) =>
    gameExec<BeloteResponse>('belote', {
      command,
      cardIndex: command === 'play' ? arg2 : undefined,
      suit: command === 'calltrump' ? arg2 : undefined,
      config,
    }),
};
