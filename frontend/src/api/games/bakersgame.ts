// API client for bakersgame. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FreeCellResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';
import type { FreeCellMoveZone } from './freecell';

/**
 * API client for the Baker's Game /bakersgame/exec endpoint. Baker's Game is the
 * same-suit FreeCell variant; it reuses the FreeCell wire shape.
 */
export const bakersgameApi = createSolitaireMoveApi<
  FreeCellResponse,
  FreeCellMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bakersgame');
