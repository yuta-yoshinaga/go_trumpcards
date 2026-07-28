// API client for bourre. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BourreResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Bourré /bourre/exec endpoint. */
export const bourreApi = {
  exec: (params: {
    command: string;
    decide?: boolean;
    indices?: number[];
    cardIndex?: number;
    config?: { cpuDifficulty?: number };
  }) => gameExec<BourreResponse>('bourre', params),
};
