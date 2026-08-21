// API client for stalactites. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { StalactitesResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Stalactites card move. */
export interface StalactitesMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Stalactites /stalactites/exec endpoint. */
export const stalactitesApi = createSolitaireMoveApi<
  StalactitesResponse,
  StalactitesMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('stalactites');
