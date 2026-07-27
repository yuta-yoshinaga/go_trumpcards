import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ombreApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeOmbreState } from '../test/stateFactories';
import { DEFAULT_OMBRE_CONFIG, useOmbreGame } from './useOmbreGame';

vi.mock('../api/gameApi', () => ({
  ombreApi: { exec: vi.fn() },
  actionLogApi: { ombre: vi.fn() },
}));

const mockExec = vi.mocked(ombreApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeOmbreState());
});

describe('useOmbreGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_OMBRE_CONFIG }));
  });

  it('handleBid dispatches a pass with no trump', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0, trumpSuit: undefined }));
  });

  it('handleBid dispatches an entrar with the chosen trump suit', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(1, 3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1, trumpSuit: 3 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useOmbreGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetRounds', '7'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_OMBRE_CONFIG, targetRounds: 7 } }),
    );
  });
});
