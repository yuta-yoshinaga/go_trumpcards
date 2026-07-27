// API client for streetsandalleys. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { StreetsAndAlleysMoveZone, StreetsAndAlleysResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Streets and Alleys /streetsandalleys/exec endpoint. */
export const streetsAndAlleysApi = createSolitaireMoveApi<
  StreetsAndAlleysResponse,
  StreetsAndAlleysMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('streetsandalleys');
