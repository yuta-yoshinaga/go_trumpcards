import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pageoneApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { PageOneResponse } from '../types/card';
import { usePageOneGame } from './usePageOneGame';

vi.mock('../api/gameApi', () => ({
  pageoneApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(pageoneApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: PageOneResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [{ design: 'SPADE', value: 8 }],
      roundScore: 0,
      cumulativeScore: 0,
      hasDeclared: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('usePageOneGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('handlePlay dispatches play with selected card', async () => {
    const { result } = renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  it('handlePlay does nothing when no card selected', async () => {
    const { result } = renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDraw dispatches draw', async () => {
    const { result } = renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleDraw();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleDeclare dispatches declare', async () => {
    const { result } = renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleDeclare();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('handleSkipDeclare dispatches skip', async () => {
    const { result } = renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleSkipDeclare();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip'));
  });

  it('handleNextRound dispatches nextround', async () => {
    const { result } = renderHook(() => usePageOneGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleNextRound();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
