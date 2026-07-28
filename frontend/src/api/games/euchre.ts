// API client for euchre. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EuchreResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Euchre game settings. */
export interface EuchreConfigInput {
  cpuDifficulty?: number;
  pointLimit?: number;
}

/** API client for the Euchre /euchre/exec endpoint. */
export const euchreApi = {
  exec: (
    command: 'reset' | 'orderup' | 'pass' | 'calltrump' | 'discard' | 'play' | 'next' | 'nextround' | 'hint',
    cardIndex?: number,
    suit?: number,
    goAlone?: boolean,
    config?: EuchreConfigInput,
  ) =>
    gameExec<EuchreResponse>('euchre', {
      command,
      cardIndex,
      suit,
      goAlone,
      config,
    }),
};
