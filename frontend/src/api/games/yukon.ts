// API client for yukon. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { YukonResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Yukon card move. */
export interface YukonMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

/** API client for the Yukon /yukon/exec endpoint. */
export const yukonApi = createSolitaireMoveApi<
  YukonResponse,
  YukonMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('yukon');
