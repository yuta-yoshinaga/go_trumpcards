import { act, renderHook, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cassinoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CassinoResponse } from '../types/card';
import { useCassinoGame } from './useCassinoGame';

vi.mock('../api/gameApi', () => ({
  cassinoApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(cassinoApi.exec);

function makeState(overrides: Partial<CassinoResponse> = {}): CassinoResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 2, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
      { id: 1, isHuman: false, cardCount: 2, cards: [], capturedCount: 0, sweepCount: 0, totalScore: 0 },
    ],
    currentTurn: 0,
    tableCards: [],
    builds: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 21, multiBuildEnabled: true, sweepBonusEnabled: true, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 0,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

const wrapper = (children: React.ReactNode) =>
  renderWithProviders(children as React.ReactElement).container ? null : null;

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
// Use renderWithProviders' Provider tree by wrapping a placeholder.
// renderHook needs a wrapper element, so we forward the providers manually.
import type { ReactNode } from 'react';
import { MemoryRouter } from 'react-router';
import { flushPendingDispatch } from '../test/flushPendingDispatch';

function Hookwrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return (
    <QueryClientProvider client={qc}>
      <MemoryRouter>{children}</MemoryRouter>
    </QueryClientProvider>
  );
}

describe('useCassinoGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('exposes initial selection state', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    expect(result.current.handIndex).toBeNull();
    expect(result.current.tableIndices).toEqual([]);
    expect(result.current.buildIndices).toEqual([]);
    expect(result.current.declaredValue).toBe(8);
  });

  it('toggles table indices', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    act(() => {
      result.current.toggleTable(0);
    });
    expect(result.current.tableIndices).toEqual([0]);
    act(() => {
      result.current.toggleTable(0);
    });
    expect(result.current.tableIndices).toEqual([]);
  });

  it('toggles build indices', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    act(() => {
      result.current.toggleBuild(2);
    });
    expect(result.current.buildIndices).toEqual([2]);
    act(() => {
      result.current.toggleBuild(2);
    });
    expect(result.current.buildIndices).toEqual([]);
  });

  it('clearSelection resets all selections', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    act(() => {
      result.current.setHandIndex(1);
      result.current.toggleTable(0);
      result.current.toggleBuild(0);
    });
    act(() => {
      result.current.clearSelection();
    });
    expect(result.current.handIndex).toBeNull();
    expect(result.current.tableIndices).toEqual([]);
    expect(result.current.buildIndices).toEqual([]);
  });

  it('playTake is a no-op when handIndex is null', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.playTake();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('playBuild is a no-op when handIndex is null', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.playBuild();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('playTrail is a no-op when handIndex is null', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.playTrail();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('playTake with hand selected calls take with sorted indices', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.setHandIndex(0);
      result.current.toggleTable(2);
      result.current.toggleTable(0);
      result.current.toggleBuild(1);
    });
    act(() => {
      result.current.playTake();
    });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('take', {
        handIndex: 0,
        tableIndices: [0, 2],
        buildIndices: [1],
      }),
    );
  });

  it('playBuild with hand selected calls build with declared value', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.setHandIndex(0);
      result.current.toggleTable(0);
      result.current.setDeclaredValue(7);
    });
    act(() => {
      result.current.playBuild();
    });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('build', {
        handIndex: 0,
        tableIndices: [0],
        declaredValue: 7,
      }),
    );
  });

  it('playTrail with hand selected calls trail', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.setHandIndex(2);
    });
    act(() => {
      result.current.playTrail();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trail', { handIndex: 2 }));
  });

  it('handleResetWithConfig forwards configInput', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    act(() => {
      result.current.handleConfigChange('cpuDifficulty', 2);
    });
    act(() => {
      result.current.handleResetWithConfig();
    });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: expect.objectContaining({ cpuDifficulty: 2 }),
      }),
    );
  });

  it('retry replays the last call', async () => {
    const { result } = renderHook(() => useCassinoGame(), { wrapper: Hookwrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    await act(async () => {
      await result.current.retry();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });
});

// silence unused-helper lint
void wrapper;
