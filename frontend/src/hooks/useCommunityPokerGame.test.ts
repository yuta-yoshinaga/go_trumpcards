import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { omahaApi } from '../api/gameApi';
import { SoundProvider } from '../providers/SoundProvider';
import type { OmahaResponse } from '../types/card';
import { OmahaPhase } from '../types/phases';
import { OMAHA_HELP, parseOmahaCommand } from '../utils/cli/commands/omahaCommands';
import { formatOmahaState } from '../utils/cli/formatters/omahaFormatter';
import { useCommunityPokerGame } from './useCommunityPokerGame';

vi.mock('../api/gameApi', () => ({
  omahaApi: { exec: vi.fn() },
  actionLogApi: { omaha: vi.fn() },
}));

const mockExec = vi.mocked(omahaApi.exec);

// Minimal state exercising the hook's derived-value logic (it reads only a
// handful of fields); cast because building the full OmahaResponse is noise.
function flopState(currentTurn: number): OmahaResponse {
  return {
    players: [
      { id: 0, isHuman: true, folded: false, allIn: false, currentBet: 0 },
      { id: 1, isHuman: false, folded: false, allIn: false, currentBet: 0 },
    ],
    communityCards: [],
    phase: OmahaPhase.FLOP,
    currentTurn,
    lastBet: 0,
    minRaise: 20,
    muckAvailable: false,
    rebuyPhaseType: 0,
    rebuyCounts: [],
  } as unknown as OmahaResponse;
}

const config = {
  game: 'omaha',
  exec: omahaApi.exec,
  phaseKeys: { [OmahaPhase.FLOP]: 'flop' },
  cli: { parseCommand: parseOmahaCommand, formatResponse: formatOmahaState, helpText: OMAHA_HELP },
} as const;

function wrapper({ children }: { children: ReactNode }) {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return createElement(QueryClientProvider, { client: queryClient }, createElement(SoundProvider, null, children));
}

describe('useCommunityPokerGame', () => {
  beforeEach(() => {
    mockExec.mockReset();
    mockExec.mockResolvedValue(flopState(0));
  });

  it('dispatches a reset on mount and exposes the returned state', async () => {
    const { result } = renderHook(() => useCommunityPokerGame(config), { wrapper });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
    await waitFor(() => expect(result.current.state).not.toBeNull());
  });

  it('handleManualReset dispatches reset with the meta-AI flag', async () => {
    const { result } = renderHook(() => useCommunityPokerGame(config), { wrapper });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    result.current.handleManualReset();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
  });

  it('derives canAct true on the human turn during an active phase', async () => {
    const { result } = renderHook(() => useCommunityPokerGame(config), { wrapper });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    expect(result.current.isActive).toBe(true);
    expect(result.current.isShowdown).toBe(false);
    expect(result.current.canAct).toBe(true);
  });

  it('derives canAct false when it is not the human turn', async () => {
    mockExec.mockResolvedValue(flopState(1));
    const { result } = renderHook(() => useCommunityPokerGame(config), { wrapper });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    expect(result.current.canAct).toBe(false);
  });
});
