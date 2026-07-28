// API client for labellelucie. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { LaBelleLucieResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the La Belle Lucie /labellelucie/exec endpoint. */
export const labellelucieApi = createSolitaireMoveApi<
  LaBelleLucieResponse,
  number,
  'reset' | 'mf' | 'ff' | 'rd' | 'ac' | 'u' | 'undo_n' | 'giveup' | 'hint' | 'log'
>('labellelucie');
