// API client for grandfathersclock. Split-file layout introduced by issue
// #4434; gameApi.ts re-exports this file, so existing imports keep working.

import type { GrandfathersClockMoveZone, GrandfathersClockResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Grandfather's Clock /grandfathersclock/exec endpoint. */
export const grandfathersClockApi = createSolitaireMoveApi<
  GrandfathersClockResponse,
  GrandfathersClockMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('grandfathersclock');
