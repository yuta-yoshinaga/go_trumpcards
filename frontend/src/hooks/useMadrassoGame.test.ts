import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { madrassoApi } from '../api/gameApi';
import { makeMadrassoState } from '../test/stateFactories';
import { useMadrassoGame } from './useMadrassoGame';

vi.mock('../api/gameApi', () => ({
  madrassoApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(madrassoApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState = makeMadrassoState();

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useMadrassoGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useMadrassoGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetPoints: 21,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useMadrassoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handlePlay dispatches play with the single selected card', async () => {
    const { result } = renderHook(() => useMadrassoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(1);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 1));
  });

  it('handleNextTrick dispatches next command', async () => {
    const { result } = renderHook(() => useMadrassoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextTrick();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useMadrassoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
