import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { macauApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { MacauResponse } from '../types/card';
import { useMacauGame } from './useMacauGame';

vi.mock('../api/gameApi', () => ({
  macauApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(macauApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: MacauResponse = {
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
  chosenSuit: 0,
  penaltyDrawCount: 0,
  playableIndices: [],
  direction: 1,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useMacauGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('handlePlay dispatches play with single selected card', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => result.current.toggleCard(2));
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  it('handlePlay does nothing when no card selected', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleDraw());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleChooseSuit dispatches suit command', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleChooseSuit(3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 3));
  });

  it('handleDeclare dispatches declare command', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleDeclare());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('handleSkipDeclare dispatches skipdeclare command', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleSkipDeclare());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipdeclare'));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config', async () => {
    const { result } = renderHook(() => useMacauGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => result.current.handleConfigChange('pointLimit', '300'));
    expect(result.current.macauConfig.pointLimit).toBe(300);
  });
});
