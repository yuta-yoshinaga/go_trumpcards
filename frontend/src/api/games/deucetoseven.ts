// API client for deucetoseven. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DeuceToSevenResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for 2-7 Triple Draw game settings. */
export interface DeuceToSevenConfigInput {
  cpuCount?: number;
  bettingLimit?: number;
  cpuMetaAI?: boolean;
}

/** API client for the 2-7 Triple Draw /deucetoseven/exec endpoint. */
export const deuceToSevenApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin',
    indices?: number[],
    amount?: number,
    config?: DeuceToSevenConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) =>
    gameExec<DeuceToSevenResponse>('deucetoseven', {
      command,
      indices,
      amount,
      humanPlayMs,
      profile,
      ...config,
    }),
};
