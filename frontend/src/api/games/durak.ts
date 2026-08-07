// API client for durak. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DurakConfigInput, DurakResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** API client for the Durak /durak/exec endpoint. */
export const durakApi = {
  exec: (
    command: 'reset' | 'attack' | 'defend' | 'pass' | 'take' | 'sort' | 'hint',
    cardIdx?: number,
    attackIdx?: number,
    config?: DurakConfigInput,
    sortMode?: number,
  ) => gameExec<DurakResponse>('durak', { command, cardIdx, attackIdx, config, sortMode }),
};
