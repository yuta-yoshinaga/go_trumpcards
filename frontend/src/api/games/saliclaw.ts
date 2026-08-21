// API client for saliclaw. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SalicLawMoveZone, SalicLawResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the SalicLaw /saliclaw/exec endpoint. */
export const salicLawApi = createSolitaireMoveApi<
  SalicLawResponse,
  SalicLawMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('saliclaw');
