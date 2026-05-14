import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { piquetApi } from '../api/gameApi';
import type { PiquetResponse } from '../types/card';
import { DEFAULT_PIQUET_CONFIG, usePiquetGame } from './usePiquetGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  piquetApi: { exec: vi.fn() },
  actionLogApi: { piquet: vi.fn() },
}));

const mockExec = vi.mocked(piquetApi.exec);

const baseState: PiquetResponse = {
  players: [],
  phase: 0,
  dealNumber: 1,
  dealsPerPartie: 6,
  elderIdx: 0,
  youngerIdx: 1,
  currentPlayerIdx: 0,
  leadPlayerIdx: 0,
  trickNumber: 0,
  tricksWon: [0, 0],
  exchangeTurn: 0,
  elderExchangedCnt: 0,
  youngerExchangedCnt: 0,
  elderTalon: [],
  youngerTalon: [],
  elderRevealedTalon: [],
  youngerRevealedTalon: [],
  carteBlanche: [false, false],
  declStage: 0,
  declResults: [],
  currentTrick: [],
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, dealsPerPartie: 6 },
};

describe('usePiquetGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render with default config', async () => {
    renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
    expect(mockExec.mock.calls[0]?.[3]).toEqual(DEFAULT_PIQUET_CONFIG);
  });

  it('handleReset sends reset with current config', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
  });

  it('handleExchangeElder forwards indices', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleExchangeElder([0, 1, 2]));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('e', undefined, [0, 1, 2]));
  });

  it('handleExchangeYounger forwards indices', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleExchangeYounger([0]));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('y', undefined, [0]));
  });

  it('handleResolveDeclaration sends d', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleResolveDeclaration());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('d'));
  });

  it('handlePlay forwards cardIndex', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handlePlay(4));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', 4));
  });

  it('handleNextDeal sends nd', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleNextDeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nd'));
  });

  it('handleHint sends h', async () => {
    const { result } = renderHook(() => usePiquetGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    act(() => result.current.handleHint());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('h'));
  });
});
