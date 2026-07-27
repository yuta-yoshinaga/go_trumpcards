// API client for accordion. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { AccordionResponse } from '../../types/card';
import { createSolitaireMoveApi } from '../gameExec';

/** Source or target pile for an Accordion move. */
export interface AccordionMoveZone {
  zone: 'pile';
  index?: number;
}

/** API client for the Accordion /accordion/exec endpoint. */
export const accordionApi = createSolitaireMoveApi<
  AccordionResponse,
  AccordionMoveZone,
  'reset' | 'move' | 'giveup' | 'hint' | 'log' | 'undo' | 'undo_n'
>('accordion');
