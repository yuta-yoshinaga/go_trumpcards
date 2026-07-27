// API client for wasp. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WaspResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Wasp card move. */
export interface WaspMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Wasp /wasp/exec endpoint. */
export const waspApi = createSolitaireMoveApi<
  WaspResponse,
  WaspMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('wasp');
