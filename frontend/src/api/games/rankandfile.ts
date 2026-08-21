// API client for rankandfile. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { RankAndFileMoveZone, RankAndFileResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Rank and File /rankandfile/exec endpoint. */
export const rankAndFileApi = createSolitaireMoveApi<
  RankAndFileResponse,
  RankAndFileMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('rankandfile');
