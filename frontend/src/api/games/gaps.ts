// API client for gaps. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { GapsResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Gaps card move. */
export interface GapsMoveZone {
  zone: 'grid';
  row: number;
  col: number;
}

/** API client for the Gaps /gaps/exec endpoint. */
export const gapsApi = createSolitaireMoveApi<
  GapsResponse,
  GapsMoveZone,
  'reset' | 'move' | 'redeal' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n'
>('gaps');
