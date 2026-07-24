import { describe, expect, it } from 'vitest';
import { spadesBagWarning, spadesBidProgress } from './spadesBid';

describe('spadesBidProgress', () => {
  it('reports tricks remaining when the bid is not yet met', () => {
    expect(spadesBidProgress(4, 1)).toEqual({ kind: 'remaining', remaining: 3 });
  });

  it('reports bags (overtricks) once the bid is met or exceeded', () => {
    expect(spadesBidProgress(3, 3)).toEqual({ kind: 'made', bags: 0 });
    expect(spadesBidProgress(3, 5)).toEqual({ kind: 'made', bags: 2 });
  });

  it('treats a Nil bid (0) as ok while no tricks are taken', () => {
    expect(spadesBidProgress(0, 0)).toEqual({ kind: 'nilOk' });
  });

  it('treats a Nil bid as failed once a trick is taken', () => {
    expect(spadesBidProgress(0, 1)).toEqual({ kind: 'nilFail' });
  });
});

describe('spadesBagWarning', () => {
  it('returns null when the player is more than two bags away', () => {
    expect(spadesBagWarning(7, 10)).toBeNull();
    expect(spadesBagWarning(0, 10)).toBeNull();
  });

  it('warns at exactly two bags away from the threshold', () => {
    expect(spadesBagWarning(8, 10)).toEqual({ level: 'warn', bags: 8, threshold: 10 });
  });

  it('escalates to danger at one bag away or once the threshold is reached', () => {
    expect(spadesBagWarning(9, 10)).toEqual({ level: 'danger', bags: 9, threshold: 10 });
    expect(spadesBagWarning(10, 10)).toEqual({ level: 'danger', bags: 10, threshold: 10 });
  });

  it('returns null when the threshold is disabled', () => {
    expect(spadesBagWarning(5, 0)).toBeNull();
  });
});
