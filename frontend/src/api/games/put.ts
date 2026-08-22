// API client for put. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PutConfig, PutResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Put /put/exec endpoint. */
export const putApi = {
  exec: (
    command: 'reset' | 'play' | 'put' | 'accept' | 'decline' | 'next' | 'hint' | 'log',
    cardIndex?: number,
    config?: Partial<PutConfig>,
  ) => gameExec<PutResponse>('put', { command, cardIndex, config }),
};
