// API client for slyfox. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SlyFoxMoveZone, SlyFoxResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Sly Fox /slyfox/exec endpoint. */
export const slyFoxApi = createSolitaireMoveApi<
  SlyFoxResponse,
  SlyFoxMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('slyfox');
