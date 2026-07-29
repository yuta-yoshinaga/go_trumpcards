// API client for americantoad. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AmericanToadMoveZone, AmericanToadResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** API client for the American Toad /americantoad/exec endpoint. */
export const americanToadApi = createSolitaireMoveApi<
  AmericanToadResponse,
  AmericanToadMoveZone,
  'reset' | 'draw' | 'move' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('americantoad');
