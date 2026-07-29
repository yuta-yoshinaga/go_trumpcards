// API client for windmill. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WindmillMoveZone, WindmillResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Windmill /windmill/exec endpoint. */
export const windmillApi = createSolitaireMoveApi<
  WindmillResponse,
  WindmillMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('windmill');
