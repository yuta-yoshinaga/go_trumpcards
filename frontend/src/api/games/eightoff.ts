// API client for eightoff. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EightOffResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for an Eight Off card move. */
export interface EightOffMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Eight Off /eightoff/exec endpoint. */
export const eightoffApi = createSolitaireMoveApi<
  EightOffResponse,
  EightOffMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('eightoff');
