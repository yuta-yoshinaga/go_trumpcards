// API client for fortythieves. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FortyThievesMoveZone, FortyThievesResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Forty Thieves /fortythieves/exec endpoint. */
export const fortyThievesApi = createSolitaireMoveApi<
  FortyThievesResponse,
  FortyThievesMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fortythieves');
