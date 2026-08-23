// API client for bauernschnapsen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BauernschnapsenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Bauernschnapsen game configuration input. */
export interface BauernschnapsenConfigInput {
  cpuDifficulty?: number;
  targetScore?: number;
}

/**
 * API client for the Bauernschnapsen /bauernschnapsen/exec endpoint.
 *
 * The second positional slot carries the contract for the `contract` command
 * (0=pass, 1=Rufer, 2=Farbenzwang, 3=Bettel); it is unused by every other
 * command, which keeps the `(command, arg1?, cardIndex?, config?)` shape that
 * `useTrickGameBase` dispatches for reset/play.
 */
export const bauernschnapsenApi = {
  exec: (
    command: 'reset' | 'contract' | 'play' | 'marriage' | 'next' | 'nextround' | 'hint',
    contract?: number,
    cardIndex?: number,
    config?: BauernschnapsenConfigInput,
    trumpSuit?: number,
  ) =>
    gameExec<BauernschnapsenResponse>('bauernschnapsen', {
      command,
      cardIndex,
      config,
      contract,
      trumpSuit,
    }),
};
