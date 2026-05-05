import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { pokersquaresApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PokerSquaresResponse } from '../types/card';
import { PokerSquaresPhase } from '../types/phases';
import { PokerSquaresPage } from './PokerSquaresPage';

vi.mock('../api/gameApi', () => ({
  pokersquaresApi: { exec: vi.fn() },
  actionLogApi: { pokersquares: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(pokersquaresApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** Build an empty 5x5 board. */
function emptyBoard(): PokerSquaresResponse['board'] {
  return Array.from({ length: 5 }, () => Array.from({ length: 5 }, () => ({ card: null as Card | null })));
}

const playingState: PokerSquaresResponse = {
  board: emptyBoard(),
  currentCard: card('SPADE', 1),
  placedCount: 0,
  phase: PokerSquaresPhase.PLAYING,
  canUndo: false,
  rowScores: [0, 0, 0, 0, 0],
  colScores: [0, 0, 0, 0, 0],
  totalScore: 0,
  message: '',
};

const playingStateWithUndo: PokerSquaresResponse = {
  ...playingState,
  placedCount: 1,
  canUndo: true,
  currentCard: card('HEART', 2),
  rowScores: [0, 0, 0, 0, 0],
  colScores: [0, 0, 0, 0, 0],
};

const completeBoard = (): PokerSquaresResponse['board'] =>
  Array.from({ length: 5 }, (_, r) =>
    Array.from({ length: 5 }, (_, c) => ({ card: card('SPADE', r * 5 + c + 1) as Card | null })),
  );

const completeState: PokerSquaresResponse = {
  board: completeBoard(),
  currentCard: null,
  placedCount: 25,
  phase: PokerSquaresPhase.COMPLETE,
  canUndo: false,
  rowScores: [10, 5, 3, 2, 1],
  colScores: [4, 5, 6, 7, 8],
  totalScore: 51,
  message: '完了',
  messageCode: 'pokersquares.complete',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

afterEach(() => {
  localStorage.clear();
});

describe('PokerSquaresPage', () => {
  it('calls exec with reset on mount', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders playing phase with total score and current card info', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('ps-board')).toBeInTheDocument());
    expect(screen.getByTestId('total-score')).toHaveTextContent('合計スコア: 0');
    expect(screen.getByText(/次に置くカード/)).toBeInTheDocument();
  });

  it('calls exec with place when an empty cell is clicked', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());

    mockApi.mockResolvedValue(playingStateWithUndo);
    fireEvent.click(screen.getByTestId('cell-2-3'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('place', 2, 3));
  });

  it('does not call place when a filled cell is clicked (button is disabled)', async () => {
    const filledState: PokerSquaresResponse = {
      ...playingState,
      board: (() => {
        const b = emptyBoard();
        b[0][0] = { card: card('HEART', 13) };
        return b;
      })(),
      placedCount: 1,
    };
    mockApi.mockResolvedValue(filledState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());
    const cell = screen.getByTestId('cell-0-0') as HTMLButtonElement;
    expect(cell.disabled).toBe(true);
  });

  it('renders undo button disabled when canUndo is false and enabled when true', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect((screen.getByRole('button', { name: '元に戻す' }) as HTMLButtonElement).disabled).toBe(true);

    mockApi.mockResolvedValue(playingStateWithUndo);
    fireEvent.click(screen.getByTestId('cell-0-0'));
    await waitFor(() =>
      expect((screen.getByRole('button', { name: '元に戻す' }) as HTMLButtonElement).disabled).toBe(false),
    );
  });

  it('calls exec with undo when undo button clicked', async () => {
    mockApi.mockResolvedValue(playingStateWithUndo);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('undo'));
  });

  it('calls exec with giveup when give up button clicked', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('giveup'));
  });

  it('renders complete phase with total score and without playing buttons', async () => {
    mockApi.mockResolvedValue(completeState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('total-score')).toBeInTheDocument());
    expect(screen.getByTestId('total-score')).toHaveTextContent('合計スコア: 51');
    expect(screen.queryByRole('button', { name: '元に戻す' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('renders empty-placeholder for current card in complete phase', async () => {
    mockApi.mockResolvedValue(completeState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('current-card-empty')).toBeInTheDocument());
  });

  it('shows reset confirm dialog and dispatches reset on confirm', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    mockApi.mockClear();
    mockApi.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('dismisses the reset dialog on cancel', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('shows network error alert when the API fails', async () => {
    mockApi.mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders row and column score badges', async () => {
    mockApi.mockResolvedValue(completeState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('ps-row-scores')).toBeInTheDocument());
    expect(screen.getByTestId('row-score-0')).toHaveTextContent('10');
    expect(screen.getByTestId('col-score-4')).toHaveTextContent('8');
  });

  it('cross-highlights the row, column, and matching score badges when an empty cell is hovered', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-2-3')).toBeInTheDocument());

    fireEvent.pointerEnter(screen.getByTestId('cell-2-3'));

    // Hovered cell + every cell in row 2 and col 3 should carry the cross-hover marker.
    expect(screen.getByTestId('cell-2-3')).toHaveAttribute('data-cross-hover', 'true');
    expect(screen.getByTestId('cell-2-0')).toHaveAttribute('data-cross-hover', 'true');
    expect(screen.getByTestId('cell-0-3')).toHaveAttribute('data-cross-hover', 'true');
    expect(screen.getByTestId('cell-4-3')).toHaveAttribute('data-cross-hover', 'true');
    // A cell off the cross stays unmarked.
    expect(screen.getByTestId('cell-0-0')).not.toHaveAttribute('data-cross-hover');

    // Score badges for the hovered line should highlight; others should not.
    expect(screen.getByTestId('row-score-2')).toHaveAttribute('data-cross-hover', 'true');
    expect(screen.getByTestId('col-score-3')).toHaveAttribute('data-cross-hover', 'true');
    expect(screen.getByTestId('row-score-0')).not.toHaveAttribute('data-cross-hover');
    expect(screen.getByTestId('col-score-0')).not.toHaveAttribute('data-cross-hover');
  });

  it('clears cross-highlight when the pointer leaves the cell', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-1-1')).toBeInTheDocument());

    const cell = screen.getByTestId('cell-1-1');
    fireEvent.pointerEnter(cell);
    expect(cell).toHaveAttribute('data-cross-hover', 'true');
    fireEvent.pointerLeave(cell);
    expect(cell).not.toHaveAttribute('data-cross-hover');
    expect(screen.getByTestId('row-score-1')).not.toHaveAttribute('data-cross-hover');
  });

  it('does not cross-highlight when hovering a filled cell', async () => {
    const filledState: PokerSquaresResponse = {
      ...playingState,
      board: (() => {
        const b = emptyBoard();
        b[0][0] = { card: card('HEART', 13) };
        return b;
      })(),
      placedCount: 1,
    };
    mockApi.mockResolvedValue(filledState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());

    fireEvent.pointerEnter(screen.getByTestId('cell-0-0'));
    expect(screen.getByTestId('cell-0-0')).not.toHaveAttribute('data-cross-hover');
    expect(screen.getByTestId('row-score-0')).not.toHaveAttribute('data-cross-hover');
  });

  it('renders the CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByTestId('ps-board')).not.toBeInTheDocument();
  });
});
