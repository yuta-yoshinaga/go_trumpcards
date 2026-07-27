import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fortyFivesApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeFortyFivesState } from '../test/stateFactories';
import { DEFAULT_FORTY_FIVES_CONFIG, useFortyFivesGame } from './useFortyFivesGame';

vi.mock('../api/gameApi', () => ({
  fortyFivesApi: { exec: vi.fn() },
  actionLogApi: { fortyfives: vi.fn() },
}));

const mockExec = vi.mocked(fortyFivesApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeFortyFivesState());
});

describe('useFortyFivesGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useFortyFivesGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_FORTY_FIVES_CONFIG }));
  });

  it('handleBid dispatches the given bid', async () => {
    const { result } = renderHook(() => useFortyFivesGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(15));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 15 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useFortyFivesGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useFortyFivesGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useFortyFivesGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useFortyFivesGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetPoints', '60'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_FORTY_FIVES_CONFIG, targetPoints: 60 } }),
    );
  });
});
