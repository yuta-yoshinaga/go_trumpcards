import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pochApi } from './poch';

const { mockGameExec } = vi.hoisted(() => ({ mockGameExec: vi.fn() }));

vi.mock('../gameExec', () => ({ gameExec: mockGameExec }));

describe('pochApi', () => {
  beforeEach(() => mockGameExec.mockReset());

  it('sends the CPU difficulty config in the request payload', () => {
    pochApi.exec('reset', undefined, { cpuDifficulty: 2 });
    expect(mockGameExec).toHaveBeenCalledWith('poch', {
      command: 'reset',
      cardIndex: undefined,
      config: { cpuDifficulty: 2 },
    });
  });
});
