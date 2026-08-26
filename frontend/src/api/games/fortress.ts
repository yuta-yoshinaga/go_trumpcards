// API client for fortress. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FortressMoveZone, FortressResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Fortress /fortress/exec endpoint. */
export const fortressApi = createSolitaireMoveApi<
  FortressResponse,
  FortressMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fortress');
