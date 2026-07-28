import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tuteApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeTuteState } from '../test/stateFactories';
import { DEFAULT_TUTE_CONFIG, useTuteGame } from './useTuteGame';

vi.mock('../api/gameApi', () => ({
  tuteApi: { exec: vi.fn() },
  actionLogApi: { tute: vi.fn() },
}));

const mockExec = vi.mocked(tuteApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeTuteState());
});

describe('useTuteGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_TUTE_CONFIG }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleDeclareMarriage dispatches marriage with the chosen suit', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDeclareMarriage(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', { suit: 3 }));
  });

  it('handleDeclareTute dispatches tute', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDeclareTute());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('tute'));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useTuteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetPoints', '151'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_TUTE_CONFIG, targetPoints: 151 } }),
    );
  });
});
