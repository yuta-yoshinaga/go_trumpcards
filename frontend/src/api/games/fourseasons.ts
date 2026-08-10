// API client for fourseasons. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FourSeasonsMoveZone, FourSeasonsResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Four Seasons /fourseasons/exec endpoint. */
export const fourseasonsApi = createSolitaireMoveApi<
  FourSeasonsResponse,
  FourSeasonsMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('fourseasons');
