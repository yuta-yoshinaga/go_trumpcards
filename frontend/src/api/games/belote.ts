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
  exec: (
    command: 'reset' | 'orderup' | 'pass' | 'calltrump' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndex?: number,
    suit?: number,
    config?: BeloteConfigInput,
  ) =>
    gameExec<BeloteResponse>('belote', {
      command,
      cardIndex,
      suit,
      config,
    }),
};
