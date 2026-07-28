// API client for kingalbert. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { KingAlbertMoveZone, KingAlbertResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the King Albert /kingalbert/exec endpoint. */
export const kingAlbertApi = createSolitaireMoveApi<
  KingAlbertResponse,
  KingAlbertMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('kingalbert');
