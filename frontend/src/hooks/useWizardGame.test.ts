import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { wizardApi } from '../api/gameApi';
import type { WizardResponse } from '../types/card';
import { DEFAULT_WIZARD_CONFIG, useWizardGame } from './useWizardGame';

vi.mock('../api/gameApi', () => ({
  wizardApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(wizardApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

/** Minimal response — only the fields these tests read. */
const baseState = { players: [], phase: 0, message: '' } as unknown as WizardResponse;

describe('useWizardGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  it('resets with the default config on mount', async () => {
    renderHook(() => useWizardGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_WIZARD_CONFIG));
  });

  it('stores the hint the server returns', async () => {
    const hint = { cardIndices: [2], reason: 'lead trump' };
    const { result } = renderHook(() => useWizardGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockResolvedValue({ ...baseState, hint } as unknown as WizardResponse);
    await act(async () => {
      await result.current.handleHint();
    });

    expect(mockExec).toHaveBeenCalledWith('hint');
    expect(result.current.hint).toEqual(hint);
    expect(result.current.hintError).toBeNull();
    expect(result.current.hintLoading).toBe(false);
  });

  it('normalises a response with no hint to null', async () => {
    const { result } = renderHook(() => useWizardGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hint).toBeNull();
  });

  it('surfaces a hint request failure without clearing hintLoading forever', async () => {
    const { result } = renderHook(() => useWizardGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockRejectedValue(new Error('network'));
    await act(async () => {
      await result.current.handleHint();
    });

    expect(result.current.hintError).toEqual(expect.any(String));
    expect(result.current.hintLoading).toBe(false);
  });

  it('clears the hint when a command succeeds, so a stale suggestion is not shown', async () => {
    const { result } = renderHook(() => useWizardGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockResolvedValue({
      ...baseState,
      hint: { cardIndices: [0], reason: 'x' },
    } as unknown as WizardResponse);
    await act(async () => {
      await result.current.handleHint();
    });
    expect(result.current.hint).not.toBeNull();

    mockExec.mockResolvedValue(baseState);
    await act(async () => {
      await result.current.exec('play', 0);
    });
    expect(result.current.hint).toBeNull();
  });
});
