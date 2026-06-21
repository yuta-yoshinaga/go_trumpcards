import { describe, expect, it } from 'vitest';
import { spadesBidProgress } from './spadesBid';

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
