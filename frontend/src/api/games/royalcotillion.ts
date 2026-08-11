// API client for royalcotillion. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RoyalCotillionMoveZone, RoyalCotillionResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the RoyalCotillion /royalcotillion/exec endpoint. */
export const royalcotillionApi = createSolitaireMoveApi<
  RoyalCotillionResponse,
  RoyalCotillionMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('royalcotillion');
