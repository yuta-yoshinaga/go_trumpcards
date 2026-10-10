import { describe, expect, it, vi } from 'vitest';
import { tusacApi } from './tusac';

const { mockGameExec } = vi.hoisted(() => ({ mockGameExec: vi.fn() }));

vi.mock('../gameExec', () => ({ gameExec: mockGameExec }));

describe('tusacApi', () => {
  it('includes the selected rounds in reset requests', () => {
    tusacApi.exec('reset', { config: { rounds: 8 } });
    expect(mockGameExec).toHaveBeenCalledWith('tusac', { command: 'reset', config: { rounds: 8 } });
  });
});
