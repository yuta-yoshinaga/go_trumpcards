import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi } from '../api/gameApi';
import { useActionLog } from './useActionLog';

vi.mock('../api/gameApi', () => ({
  actionLogApi: {
    blackjack: vi.fn(),
  },
}));

describe('useActionLog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('initializes with null actionLog', () => {
    const { result } = renderHook(() => useActionLog('blackjack'));
    expect(result.current.actionLog).toBeNull();
  });

  it('fetches and sets actionLog on showActionLog', async () => {
    const mockEntries = [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test', cards: [] }];
    vi.mocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: mockEntries });

    const { result } = renderHook(() => useActionLog('blackjack'));

    await act(async () => {
      await result.current.showActionLog();
    });

    expect(actionLogApi.blackjack).toHaveBeenCalledTimes(1);
    expect(result.current.actionLog).toEqual(mockEntries);
  });

  it('clears actionLog on hideActionLog', async () => {
    const mockEntries = [{ turnNumber: 1, playerIdx: 0, actionType: 'hit', detail: 'test', cards: [] }];
    vi.mocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: mockEntries });

    const { result } = renderHook(() => useActionLog('blackjack'));

    await act(async () => {
      await result.current.showActionLog();
    });
    expect(result.current.actionLog).toEqual(mockEntries);

    act(() => {
      result.current.hideActionLog();
    });
    expect(result.current.actionLog).toBeNull();
  });
});
