// API client for canfield. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { CanfieldResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Canfield card move. */
export interface CanfieldMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Canfield /canfield/exec endpoint. */
export const canfieldApi = createSolitaireMoveApi<
  CanfieldResponse,
  CanfieldMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('canfield');
