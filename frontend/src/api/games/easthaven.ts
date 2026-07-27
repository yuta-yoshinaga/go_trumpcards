// API client for easthaven. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { EasthavenResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for an Easthaven card move. */
export interface EasthavenMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Easthaven /easthaven/exec endpoint. */
export const easthavenApi = createSolitaireMoveApi<
  EasthavenResponse,
  EasthavenMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('easthaven');
