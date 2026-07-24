import { describe, expect, it } from 'vitest';
import { wizardBidAccuracy } from './wizardBidAccuracy';

describe('wizardBidAccuracy', () => {
  it('marks a player who bid 3 and took 3 as made with zero delta', () => {
    const [entry] = wizardBidAccuracy([{ name: 'You', bid: 3, trickCount: 3 }]);
    expect(entry.outcome).toBe('made');
    expect(entry.delta).toBe(0);
  });

  it('reports a +2 delta and over outcome for a player who bid 2 but took 4', () => {
    const [entry] = wizardBidAccuracy([{ name: 'CPU 1', bid: 2, trickCount: 4 }]);
    expect(entry.outcome).toBe('over');
    expect(entry.delta).toBe(2);
  });

  it('reports a -1 delta and under outcome for a player who bid 2 but took 1', () => {
    const [entry] = wizardBidAccuracy([{ name: 'CPU 2', bid: 2, trickCount: 1 }]);
    expect(entry.outcome).toBe('under');
    expect(entry.delta).toBe(-1);
  });

  it('handles a bid of 0 that is met exactly', () => {
    const [entry] = wizardBidAccuracy([{ name: 'CPU 3', bid: 0, trickCount: 0 }]);
    expect(entry.outcome).toBe('made');
    expect(entry.delta).toBe(0);
  });

  it('skips players that never placed a bid', () => {
    const entries = wizardBidAccuracy([
      { name: 'You', bid: -1, trickCount: 0 },
      { name: 'CPU 1', bid: 1, trickCount: 1 },
    ]);
    expect(entries).toHaveLength(1);
    expect(entries[0]?.name).toBe('CPU 1');
  });

  it('preserves player order and computes an entry per bidder', () => {
    const entries = wizardBidAccuracy([
      { name: 'You', bid: 3, trickCount: 3 },
      { name: 'CPU 1', bid: 2, trickCount: 4 },
    ]);
    expect(entries.map((e) => e.name)).toEqual(['You', 'CPU 1']);
    expect(entries.map((e) => e.outcome)).toEqual(['made', 'over']);
  });
});
