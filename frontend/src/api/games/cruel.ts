// API client for cruel. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CruelResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Cruel card move. */
export interface CruelMoveZone {
  zone: string;
  col?: number;
}

/** API client for the Cruel /cruel/exec endpoint. */
export const cruelApi = createSolitaireMoveApi<
  CruelResponse,
  CruelMoveZone,
  'reset' | 'move' | 'shift' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('cruel');
