import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { labellelucieApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, LaBelleLucieResponse } from '../types/card';
import { LaBelleLuciePage } from './LaBelleLuciePage';

vi.mock('../api/gameApi', () => ({
  labellelucieApi: { exec: vi.fn() },
  actionLogApi: { labellelucie: vi.fn() },
}));

const mockExec = vi.mocked(labellelucieApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<LaBelleLucieResponse> = {}): LaBelleLucieResponse {
  return {
    fans: [[card('SPADE', 9)], [card('SPADE', 8)], [card('DIAMOND', 1)]],
    foundation: [[], [], [], []],
    redealsLeft: 3,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('LaBelleLuciePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<LaBelleLuciePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<LaBelleLuciePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders fans and foundations', async () => {
    renderWithProviders(<LaBelleLuciePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.getByTestId('foundation-0')).toBeInTheDocument();
  });

  it('selects a source fan then moves to another fan', async () => {
    renderWithProviders(<LaBelleLuciePage />);
    const src = await screen.findByTestId('fan-1');
    fireEvent.click(src);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('fan-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mf', 1, 0));
  });

  it('selects a source fan then sends it to a foundation', async () => {
    renderWithProviders(<LaBelleLuciePage />);
    const src = await screen.findByTestId('fan-2');
    fireEvent.click(src);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ff', 2));
  });

  it('redeals, auto-completes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<LaBelleLuciePage />);
    await screen.findByTestId('redeal-button');
    for (const [testid, cmd] of [
      ['redeal-button', 'rd'],
      ['autocomplete-button', 'ac'],
      ['undo-button', 'u'],
      ['hint-button', 'hint'],
      ['giveup-button', 'giveup'],
    ] as const) {
      mockExec.mockClear();
      fireEvent.click(screen.getByTestId(testid));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(cmd));
    }
  });

  it('disables redeal when none are left', async () => {
    mockExec.mockResolvedValue(makeState({ redealsLeft: 0 }));
    renderWithProviders(<LaBelleLuciePage />);
    await waitFor(() => expect(screen.getByTestId('redeal-button')).toBeDisabled());
  });

  it('hides action buttons at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<LaBelleLuciePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('redeal-button')).not.toBeInTheDocument();
  });
});
