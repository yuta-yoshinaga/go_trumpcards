// API client for bakersdozen. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BakersDozenMoveZone, BakersDozenResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Baker's Dozen /bakersdozen/exec endpoint. */
export const bakersDozenApi = createSolitaireMoveApi<
  BakersDozenResponse,
  BakersDozenMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bakersdozen');
