// API client for fortyandeight. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FortyAndEightMoveZone, FortyAndEightResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Forty and Eight /fortyandeight/exec endpoint. */
export const fortyAndEightApi = createSolitaireMoveApi<
  FortyAndEightResponse,
  FortyAndEightMoveZone,
  'reset' | 'draw' | 'redeal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fortyandeight');
