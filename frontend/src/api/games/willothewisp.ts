// API client for willothewisp. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { WillOTheWispResponse } from '../../types/games/willothewisp';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a WillOTheWisp card move. */
export interface WillOTheWispMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the WillOTheWisp /willothewisp/exec endpoint. */
export const willothewispApi = createSolitaireMoveApi<
  WillOTheWispResponse,
  WillOTheWispMoveZone,
  'reset' | 'deal' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('willothewisp');
