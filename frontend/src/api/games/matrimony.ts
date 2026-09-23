// API client for matrimony. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { MatrimonyMoveZone, MatrimonyResponse } from '../../types/games/matrimony';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Matrimony /matrimony/exec endpoint. */
export const matrimonyApi = createSolitaireMoveApi<
  MatrimonyResponse,
  MatrimonyMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('matrimony');
