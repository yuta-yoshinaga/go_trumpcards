import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { wattenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeWattenState } from '../test/stateFactories';
import { DEFAULT_WATTEN_CONFIG, useWattenGame } from './useWattenGame';

vi.mock('../api/gameApi', () => ({
  wattenApi: { exec: vi.fn() },
  actionLogApi: { watten: vi.fn() },
}));

const mockExec = vi.mocked(wattenApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeWattenState());
});

describe('useWattenGame', () => {
  it('reset dispatches with the default config in the last positional slot', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, DEFAULT_WATTEN_CONFIG),
    );
  });

  it('handleDeclare dispatches declare with rank and suit', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDeclare(13, 3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 13, 3));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play with the selected card index', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, undefined, 2));
  });

  it('handleRaise dispatches a bare raise command', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRaise());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise'));
  });

  it('handleRespond dispatches respond with the hold flag', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRespond(true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', undefined, undefined, undefined, true));
    act(() => result.current.handleRespond(false));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', undefined, undefined, undefined, false));
  });

  it('handleNextRound dispatches nextround', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleHint dispatches hint', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handleHint());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useWattenGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetScore', '21'));
    mockExec.mockClear();
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, undefined, {
        ...DEFAULT_WATTEN_CONFIG,
        targetScore: 21,
      }),
    );
  });
});
