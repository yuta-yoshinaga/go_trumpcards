// API client for badugi. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BadugiResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Badugi game settings. */
export interface BadugiConfigInput {
  cpuCount?: number;
  bettingLimit?: number;
  cpuMetaAI?: boolean;
}

/** API client for the Badugi /badugi/exec endpoint. */
export const badugiApi = {
  exec: (
    command: 'reset' | 'exchange' | 'stand' | 'fold' | 'check' | 'call' | 'bet' | 'raise' | 'allin',
    indices?: number[],
    amount?: number,
    config?: BadugiConfigInput,
    humanPlayMs?: number,
    profile?: unknown,
  ) => gameExec<BadugiResponse>('badugi', { command, indices, amount, humanPlayMs, profile, ...config }),
};
