import { describe, expect, it, vi } from 'vitest';
import { gameExec } from '../gameExec';
import { baseballpokerApi } from './baseballpoker';

vi.mock('../gameExec', () => ({ gameExec: vi.fn() }));

describe('baseballpokerApi', () => {
  it('sends table settings with reset', () => {
    baseballpokerApi.exec('reset', { config: { seats: 3, initialChips: 500, ante: 15 } });
    expect(gameExec).toHaveBeenCalledWith('baseballpoker', {
      command: 'reset',
      config: { seats: 3, initialChips: 500, ante: 15 },
    });
  });
});
