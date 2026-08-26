// API client for somerset. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SomersetMoveZone, SomersetResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Somerset /somerset/exec endpoint. */
export const somersetApi = createSolitaireMoveApi<
  SomersetResponse,
  SomersetMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('somerset');
