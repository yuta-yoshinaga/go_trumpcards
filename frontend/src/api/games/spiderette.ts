// API client for spiderette. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SpideretteResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Spiderette card move. */
export interface SpideretteMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Spiderette /spiderette/exec endpoint. */
export const spideretteApi = createSolitaireMoveApi<
  SpideretteResponse,
  SpideretteMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('spiderette');
