import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useLobaGame } from './useLobaGame';

const { mockOptions, mockRunApi } = vi.hoisted(() => ({
  mockOptions: { current: null as { onSuccess?: (response: unknown, args: unknown[]) => void } | null },
  mockRunApi: vi.fn<(...args: unknown[]) => Promise<void>>(),
}));

vi.mock('../api/gameApi', () => ({ lobaApi: { exec: vi.fn() } }));
vi.mock('./useGameApi', () => ({
  isRejectedAction: (response: unknown) => {
    const result = response as { message?: string; messageCode?: string };
    return Boolean(result.message && !result.messageCode);
  },
  useGameApi: (_api: unknown, options: typeof mockOptions.current) => {
    mockOptions.current = options;
    return { state: null, loading: false, error: null, exec: mockRunApi, retry: vi.fn() };
  },
}));

describe('useLobaGame action acceptance', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRunApi.mockImplementation(async (...args) => {
      mockOptions.current?.onSuccess?.({ message: '', messageCode: '' }, args);
    });
  });

  it('reports meld and layoff acceptance from the server response', async () => {
    const { result } = renderHook(() => useLobaGame());
    let meldAccepted = false;
    await act(async () => {
      meldAccepted = await result.current.handleMeld([0, 1, 2]);
    });
    expect(meldAccepted).toBe(true);

    mockRunApi.mockImplementationOnce(async (...args) => {
      mockOptions.current?.onSuccess?.({ message: 'illegal layoff', messageCode: '' }, args);
    });
    let layoffAccepted = true;
    await act(async () => {
      layoffAccepted = await result.current.handleLayOff(3, 0);
    });
    expect(layoffAccepted).toBe(false);
  });
});
