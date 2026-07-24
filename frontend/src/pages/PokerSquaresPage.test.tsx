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

  it('grows the card width to fill the viewport on a 375px mobile screen', async () => {
    const original = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockApi.mockResolvedValue(playingState);
      renderWithProviders(<PokerSquaresPage />);
      const cell = await screen.findByTestId('cell-0-0');
      // floor((375 - 112) / 5) = 52, clamped to [40, 60] → 52px (> the fixed 40px mobile preset).
      expect(cell.querySelector('div')).toHaveStyle({ width: '52px' });
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: original });
    }
  });

  it('calls exec with place when an empty cell is clicked', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());

    mockApi.mockResolvedValue(playingStateWithUndo);
    fireEvent.click(screen.getByTestId('cell-2-3'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('place', 2, 3));
  });

  it('announces the row score preview in a live region on focus and clears it on blur', async () => {
    const b = emptyBoard();
    b[0][1] = { card: card('HEART', 2) };
    b[0][2] = { card: card('HEART', 3) };
    b[0][3] = { card: card('HEART', 4) };
    b[0][4] = { card: card('HEART', 5) };
    // Placing HEART 6 at (0,0) completes row 0 as a straight flush; column 0 is
    // still empty so only the row preview is announced.
    mockApi.mockResolvedValue({
      ...playingState,
      board: b,
      currentCard: card('HEART', 6),
      placedCount: 4,
    });
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());
    const live = screen.getByTestId('ps-preview-live');
    expect(live).toHaveTextContent('');
    fireEvent.focus(screen.getByTestId('cell-0-0'));
    await waitFor(() => expect(live).toHaveTextContent(/1行目.*完成.*点/));
    fireEvent.blur(screen.getByTestId('cell-0-0'));
    await waitFor(() => expect(live).toHaveTextContent(''));
  });

  it('announces the column score preview when focusing a cell that completes a column', async () => {
    const b = emptyBoard();
    b[1][0] = { card: card('HEART', 2) };
    b[2][0] = { card: card('HEART', 3) };
    b[3][0] = { card: card('HEART', 4) };
    b[4][0] = { card: card('HEART', 5) };
    // Placing HEART 6 at (0,0) completes column 0; row 0 is otherwise empty so
    // only the column preview is announced.
    mockApi.mockResolvedValue({
      ...playingState,
      board: b,
      currentCard: card('HEART', 6),
      placedCount: 4,
    });
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());
    const live = screen.getByTestId('ps-preview-live');
    fireEvent.focus(screen.getByTestId('cell-0-0'));
    await waitFor(() => expect(live).toHaveTextContent(/1列目.*完成.*点/));
  });

  it('does not announce a preview when the focused cell completes no line', async () => {
    // Only two cards in row 0 — focusing (0,0) completes nothing.
    const b = emptyBoard();
    b[0][1] = { card: card('HEART', 2) };
    b[0][2] = { card: card('HEART', 3) };
    mockApi.mockResolvedValue({ ...playingState, board: b, currentCard: card('SPADE', 9), placedCount: 2 });
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());
    fireEvent.focus(screen.getByTestId('cell-0-0'));
    // No completed line -> live region stays empty.
    expect(screen.getByTestId('ps-preview-live')).toHaveTextContent('');
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

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockApi.mockResolvedValue(playingState);
    renderWithProviders(<PokerSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockApi).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
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

  it('shows a partial made-hand hint for an incomplete row when hovering an empty cell', async () => {
    // Row 0 holds two 9's; placing a third 9 at (0,2) makes trips but the row is
    // still incomplete, so the muted partial hint (not a locked +N) should show.
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 9) };
    board[0][1] = { card: card('CLOVER', 9) };
    mockApi.mockResolvedValue({
      ...playingState,
      board,
      currentCard: card('HEART', 9),
      placedCount: 2,
    });
    renderWithProviders(<PokerSquaresPage />);
    const cell = await screen.findByTestId('cell-0-2');
    fireEvent.pointerEnter(cell);
    const partial = await screen.findByTestId('row-partial-preview-0');
    expect(partial).toHaveTextContent('スリーカード');
    // No completed-line +N preview should be present for the incomplete row.
    expect(screen.queryByTestId('row-score-preview-0')).not.toBeInTheDocument();
  });

  it('shows a partial made-hand hint for an incomplete column when hovering an empty cell', async () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 4) };
    board[1][0] = { card: card('CLOVER', 4) };
    mockApi.mockResolvedValue({
      ...playingState,
      board,
      currentCard: card('HEART', 7),
      placedCount: 2,
    });
    renderWithProviders(<PokerSquaresPage />);
    const cell = await screen.findByTestId('cell-2-0');
    fireEvent.pointerEnter(cell);
    // Column 0 already has a pair of 4's; placing an unrelated card keeps the pair.
    const partial = await screen.findByTestId('col-partial-preview-0');
    expect(partial).toHaveTextContent('ワンペア');
  });

  it('does not show a partial hint when the placement forms no made hand', async () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 2) };
    board[0][1] = { card: card('CLOVER', 5) };
    mockApi.mockResolvedValue({
      ...playingState,
      board,
      currentCard: card('HEART', 9),
      placedCount: 2,
    });
    renderWithProviders(<PokerSquaresPage />);
    const cell = await screen.findByTestId('cell-0-2');
    fireEvent.pointerEnter(cell);
    await waitFor(() => expect(screen.getByTestId('cell-0-2')).toHaveAttribute('data-cross-hover', 'true'));
    expect(screen.queryByTestId('row-partial-preview-0')).not.toBeInTheDocument();
  });

  it('announces the partial made hand in the live region for keyboard users', async () => {
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 9) };
    board[0][1] = { card: card('CLOVER', 9) };
    mockApi.mockResolvedValue({
      ...playingState,
      board,
      currentCard: card('HEART', 9),
      placedCount: 2,
    });
    renderWithProviders(<PokerSquaresPage />);
    const cell = await screen.findByTestId('cell-0-2');
    fireEvent.focus(cell);
    const live = screen.getByTestId('ps-preview-live');
    await waitFor(() => expect(live).toHaveTextContent(/1行目.*見込み/));
  });

  it('shows the projected row score when hovering a cell that would complete a row', async () => {
    // Row 0 has four 9's of different suits — placing a 5 at (0,4) completes a four-of-a-kind (50 pts).
    const board = emptyBoard();
    board[0][0] = { card: card('SPADE', 9) };
    board[0][1] = { card: card('CLOVER', 9) };
    board[0][2] = { card: card('HEART', 9) };
    board[0][3] = { card: card('DIAMOND', 9) };
    mockApi.mockResolvedValue({
      ...playingState,
      board,
      currentCard: card('SPADE', 5),
      placedCount: 4,
    });
    renderWithProviders(<PokerSquaresPage />);
    const cell = await screen.findByTestId('cell-0-4');
    fireEvent.pointerEnter(cell);
    await waitFor(() => expect(screen.getByTestId('cell-0-4')).toHaveAttribute('data-cross-hover', 'true'));
    const preview = await screen.findByTestId('row-score-preview-0');
    expect(preview.textContent).toContain('+50');
  });
});
