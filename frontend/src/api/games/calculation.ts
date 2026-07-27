// API client for calculation. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CalculationMoveZone, CalculationResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Calculation /calculation/exec endpoint. */
export const calculationApi = createSolitaireMoveApi<
  CalculationResponse,
  CalculationMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('calculation');
