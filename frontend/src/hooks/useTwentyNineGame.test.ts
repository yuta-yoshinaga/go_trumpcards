import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { twentyNineApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeTwentyNineState } from '../test/stateFactories';
import { DEFAULT_TWENTY_NINE_CONFIG, useTwentyNineGame } from './useTwentyNineGame';

vi.mock('../api/gameApi', () => ({
  twentyNineApi: { exec: vi.fn() },
  actionLogApi: { twentynine: vi.fn() },
}));

const mockExec = vi.mocked(twentyNineApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeTwentyNineState());
});

describe('useTwentyNineGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useTwentyNineGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_TWENTY_NINE_CONFIG }));
  });

  it('handleBid dispatches the given bid', async () => {
    const { result } = renderHook(() => useTwentyNineGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(20));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 20 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useTwentyNineGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useTwentyNineGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useTwentyNineGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useTwentyNineGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetPoints', '8'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_TWENTY_NINE_CONFIG, targetPoints: 8 } }),
    );
  });
});
