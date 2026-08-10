// API client for auldlangsyne. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AuldLangSyneMoveZone, AuldLangSyneResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/**
 * API client for the AuldLangSyne /auldlangsyne/exec endpoint.
 *
 * `deal` is in the command union where Sir Tommy has a stock->waste `move`:
 * dealing here is a single forced action across all four wastes rather than a
 * placement the player directs.
 */
export const auldlangsyneApi = createSolitaireMoveApi<
  AuldLangSyneResponse,
  AuldLangSyneMoveZone,
  'reset' | 'move' | 'deal' | 'giveup' | 'hint' | 'autocomplete' | 'log' | 'undo' | 'undo_n'
>('auldlangsyne');
