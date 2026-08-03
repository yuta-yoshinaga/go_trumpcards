// API client for duchess. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { DuchessMoveZone, DuchessResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Duchess /duchess/exec endpoint. */
export const duchessApi = createSolitaireMoveApi<
  DuchessResponse,
  DuchessMoveZone,
  'reset' | 'base' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('duchess');
