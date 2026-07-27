// API client for bristol. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BristolMoveZone, BristolResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Bristol /bristol/exec endpoint. */
export const bristolApi = createSolitaireMoveApi<
  BristolResponse,
  BristolMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bristol');
