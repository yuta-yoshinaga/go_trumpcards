// API client for bisley. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { BisleyMoveZone, BisleyResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the Bisley /bisley/exec endpoint. */
export const bisleyApi = createSolitaireMoveApi<
  BisleyResponse,
  BisleyMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('bisley');
