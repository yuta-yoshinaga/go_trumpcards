// API client for congress. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CongressMoveZone, CongressResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Congress /congress/exec endpoint. */
export const congressApi = createSolitaireMoveApi<
  CongressResponse,
  CongressMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('congress');
