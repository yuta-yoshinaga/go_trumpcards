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

  it('shows the stuck redeal banner when no legal move remains', async () => {
    // No Aces, no foundation builds, no same-suit stacks -> stuck.
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 5)], [card('HEART', 9)], [card('CLOVER', 2)]],
        foundation: [[], [], [], []],
        redealsLeft: 3,
      }),
    );
    renderWithProviders(<LaBelleLuciePage />);
    await waitFor(() => expect(screen.getByTestId('ll-stuck-banner')).toBeInTheDocument());
  });

  it('hides the stuck banner when a legal move exists', async () => {
    renderWithProviders(<LaBelleLuciePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('ll-stuck-banner')).not.toBeInTheDocument();
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

  it('highlights the hint source and destination fans after a hint', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromFan: 1, toFan: 0, toFoundation: false } }));
    renderWithProviders(<LaBelleLuciePage />);
    await screen.findByTestId('hint-button');
    // No highlight until the player asks for a hint.
    expect(screen.getByTestId('fan-1')).not.toHaveAttribute('data-hint-source');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('fan-1')).toHaveAttribute('data-hint-source', 'true'));
    expect(screen.getByTestId('fan-0')).toHaveAttribute('data-hint-dest', 'true');
    expect(screen.getByTestId('ll-foundation-row')).not.toHaveAttribute('data-hint-foundation');
  });

  it('highlights the foundation row for a to-foundation hint', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromFan: 2, toFan: -1, toFoundation: true } }));
    renderWithProviders(<LaBelleLuciePage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() =>
      expect(screen.getByTestId('ll-foundation-row')).toHaveAttribute('data-hint-foundation', 'true'),
    );
    expect(screen.getByTestId('fan-2')).toHaveAttribute('data-hint-source', 'true');
    // No fan is marked as the destination when the move targets a foundation.
    expect(screen.getByTestId('fan-0')).not.toHaveAttribute('data-hint-dest');
  });

  it('clears the hint highlight when the board changes', async () => {
    mockExec
      .mockResolvedValueOnce(makeState()) // mount reset
      .mockResolvedValueOnce(makeState({ hint: { fromFan: 1, toFan: 0, toFoundation: false } })) // hint
      .mockResolvedValue(makeState({ moveCount: 1 })); // subsequent move advances the board
    renderWithProviders(<LaBelleLuciePage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('fan-1')).toHaveAttribute('data-hint-source', 'true'));
    // Perform a move: select fan-1 then drop on fan-0.
    fireEvent.click(screen.getByTestId('fan-1'));
    fireEvent.click(screen.getByTestId('fan-0'));
    await waitFor(() => expect(screen.getByTestId('fan-1')).not.toHaveAttribute('data-hint-source'));
    expect(screen.getByTestId('fan-0')).not.toHaveAttribute('data-hint-dest');
  });
});
