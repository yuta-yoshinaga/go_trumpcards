// API client for agnes. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AgnesResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for an Agnes Sorel card move. */
export interface AgnesMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Agnes Sorel /agnes/exec endpoint. */
export const agnesApi = createSolitaireMoveApi<
  AgnesResponse,
  AgnesMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n'
>('agnes');
