import { describe, expect, it } from 'vitest';
import type { ActionLogEntry } from '../types/card';
import { buildWattenStakeHistory, WATTEN_BASE_STAKE } from './wattenStakeHistory';

/** Builds a minimal action-log entry for tests. */
function entry(turnNumber: number, actionType: string, playerIdx: number): ActionLogEntry {
  return { turnNumber, playerIdx, actionType, detail: '' };
}

describe('buildWattenStakeHistory', () => {
  it('returns an empty array for null, undefined, or empty input', () => {
    expect(buildWattenStakeHistory(null)).toEqual([]);
    expect(buildWattenStakeHistory(undefined)).toEqual([]);
    expect(buildWattenStakeHistory([])).toEqual([]);
  });

  it('returns an empty array when the deal has no raise/respond events', () => {
    const log = [entry(1, 'declare', 0), entry(2, 'play', 0), entry(3, 'trick_win', 2)];
    expect(buildWattenStakeHistory(log)).toEqual([]);
  });

  it('reconstructs a raise then hold escalation in order', () => {
    const log = [entry(1, 'declare', 0), entry(2, 'raise', 1), entry(3, 'hold', 0)];
    expect(buildWattenStakeHistory(log)).toEqual([
      { key: 2, type: 'raise', playerIdx: 1, stake: 3 },
      { key: 3, type: 'hold', playerIdx: 0, stake: 3 },
    ]);
  });

  it('reconstructs successive accepted raises with rising stakes', () => {
    const log = [
      entry(1, 'declare', 0),
      entry(2, 'raise', 1),
      entry(3, 'hold', 0),
      entry(4, 'raise', 0),
      entry(5, 'hold', 1),
    ];
    expect(buildWattenStakeHistory(log)).toEqual([
      { key: 2, type: 'raise', playerIdx: 1, stake: 3 },
      { key: 3, type: 'hold', playerIdx: 0, stake: 3 },
      { key: 4, type: 'raise', playerIdx: 0, stake: 4 },
      { key: 5, type: 'hold', playerIdx: 1, stake: 4 },
    ]);
  });

  it('reports a fold at the last settled stake (before the rejected raise)', () => {
    const log = [entry(1, 'declare', 0), entry(2, 'raise', 1), entry(3, 'fold', 0)];
    expect(buildWattenStakeHistory(log)).toEqual([
      { key: 2, type: 'raise', playerIdx: 1, stake: 3 },
      { key: 3, type: 'fold', playerIdx: 0, stake: WATTEN_BASE_STAKE },
    ]);
  });

  it('includes a still-pending raise that has no response yet', () => {
    const log = [entry(1, 'declare', 0), entry(2, 'raise', 1)];
    expect(buildWattenStakeHistory(log)).toEqual([{ key: 2, type: 'raise', playerIdx: 1, stake: 3 }]);
  });

  it('only reflects the current deal (entries after the last declare)', () => {
    const log = [
      entry(1, 'declare', 0),
      entry(2, 'raise', 1),
      entry(3, 'hold', 0),
      entry(4, 'deal_score', -1),
      entry(5, 'declare', 1),
      entry(6, 'raise', 0),
    ];
    expect(buildWattenStakeHistory(log)).toEqual([{ key: 6, type: 'raise', playerIdx: 0, stake: 3 }]);
  });

  it('honours a custom base stake', () => {
    const log = [entry(1, 'declare', 0), entry(2, 'raise', 1)];
    expect(buildWattenStakeHistory(log, 5)).toEqual([{ key: 2, type: 'raise', playerIdx: 1, stake: 6 }]);
  });
});
