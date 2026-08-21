// API client for fourteenout. Split out of gameApi.ts (issue #4434);
// gameApi.ts re-exports this file, so existing imports keep working.

import type { FourteenOutResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/**
 * API client for the Fourteen Out /fourteenout/exec endpoint.
 *
 * **列番号 2 つで足りる。**動かせるのは各列の末尾だけなので、クローン元の
 * Monte Carlo が要る (行,列) x2 は要らない。`deal` も無い ── 山札が無いので、
 * 受け付けても盤が変わらない無言の no-op になる。
 */
export const fourteenoutApi = {
  exec: (command: 'reset' | 'remove' | 'undo' | 'giveup' | 'hint' | 'log', fromCol?: number, toCol?: number) =>
    gameExec<FourteenOutResponse>('fourteenout', { command, fromCol, toCol }),
};
