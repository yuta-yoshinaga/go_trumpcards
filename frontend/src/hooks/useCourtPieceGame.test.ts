import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { courtPieceApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeCourtPieceState } from '../test/stateFactories';
import { DEFAULT_COURT_PIECE_CONFIG, useCourtPieceGame } from './useCourtPieceGame';

vi.mock('../api/gameApi', () => ({
  courtPieceApi: { exec: vi.fn() },
  actionLogApi: { courtpiece: vi.fn() },
}));

const mockExec = vi.mocked(courtPieceApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeCourtPieceState());
});

describe('useCourtPieceGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useCourtPieceGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_COURT_PIECE_CONFIG }));
  });

  it('handleDeclareTrump dispatches the given suit', async () => {
    const { result } = renderHook(() => useCourtPieceGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDeclareTrump(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 3 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useCourtPieceGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useCourtPieceGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useCourtPieceGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useCourtPieceGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('pointLimit', '9'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_COURT_PIECE_CONFIG, pointLimit: 9 } }),
    );
  });
});
