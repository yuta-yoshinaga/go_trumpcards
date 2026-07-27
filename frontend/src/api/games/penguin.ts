// API client for penguin. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { PenguinResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target zone for a Penguin card move. */
export interface PenguinMoveZone {
  zone: string;
  col?: number;
  cell?: number;
  cardIndex?: number;
}

/** API client for the Penguin /penguin/exec endpoint. */
export const penguinApi = createSolitaireMoveApi<
  PenguinResponse,
  PenguinMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('penguin');
