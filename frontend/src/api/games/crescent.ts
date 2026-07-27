// API client for crescent. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CrescentMoveZone, CrescentResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Crescent Solitaire /crescent/exec endpoint. */
export const crescentApi = createSolitaireMoveApi<
  CrescentResponse,
  CrescentMoveZone,
  'reset' | 'move' | 'redeal' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('crescent');
