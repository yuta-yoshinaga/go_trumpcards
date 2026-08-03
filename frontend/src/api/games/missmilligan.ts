// API client for missmilligan. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MissMilliganMoveZone, MissMilliganResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Miss Milligan /missmilligan/exec endpoint. */
export const missMilliganApi = createSolitaireMoveApi<
  MissMilliganResponse,
  MissMilliganMoveZone,
  'reset' | 'deal' | 'move' | 'waive' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('missmilligan');
