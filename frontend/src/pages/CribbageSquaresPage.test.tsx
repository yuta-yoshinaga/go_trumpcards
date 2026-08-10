import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cribbagesquaresApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CribbageSquaresResponse, CribbageSquaresScore } from '../types/card';
import { CribbageSquaresPage, cribbageBreakdownParts } from './CribbageSquaresPage';

vi.mock('../api/gameApi', () => ({
  cribbagesquaresApi: { exec: vi.fn() },
  actionLogApi: { cribbagesquares: vi.fn() },
}));

const mockExec = vi.mocked(cribbagesquaresApi.exec);
const card = (design: CardDesign, value: number): Card => ({ design, value });
const zero = (): CribbageSquaresScore => ({ fifteens: 0, pairs: 0, runs: 0, flush: 0, nobs: 0, total: 0 });

function makeState(overrides: Partial<CribbageSquaresResponse> = {}): CribbageSquaresResponse {
  const board = Array.from({ length: 4 }, () => Array.from({ length: 4 }, () => ({ card: null as Card | null })));
  board[0][0] = { card: card('SPADE', 5) };
  return {
    board,
    currentCard: card('HEART', 10),
    starter: null,
    placedCount: 1,
    phase: 0,
    canUndo: true,
    rowScores: [0, 0, 0, 0],
    colScores: [0, 0, 0, 0],
    rowDetails: [zero(), zero(), zero(), zero()],
    colDetails: [zero(), zero(), zero(), zero()],
    totalScore: 0,
    winScore: 61,
    isWin: false,
    message: '',
    messageCode: 'cribbagesquares.playing',
    ...overrides,
  };
}

describe('cribbageBreakdownParts', () => {
  const label = (key: string, n: number) => `${key}:${n}`;

  it('lists only the components that scored', () => {
    expect(cribbageBreakdownParts({ fifteens: 4, pairs: 2, runs: 0, flush: 0, nobs: 1, total: 7 }, label)).toEqual([
      'fifteens:4',
      'pairs:2',
      'nobs:1',
    ]);
  });

  it('returns nothing for a hand that scored nothing', () => {
    expect(cribbageBreakdownParts(zero(), label)).toEqual([]);
    expect(cribbageBreakdownParts(undefined, label)).toEqual([]);
  });
});

describe('CribbageSquaresPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // The grid is 4x4, so (3,3) exists and (4,0) must not.
  it('renders a 4x4 grid', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeInTheDocument());
    expect(screen.getByTestId('cell-3-3')).toBeInTheDocument();
    expect(screen.queryByTestId('cell-4-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('cell-0-4')).not.toBeInTheDocument();
  });

  it('places the card in hand into an empty cell', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-1-2')).toBeEnabled());
    fireEvent.click(screen.getByTestId('cell-1-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('place', 1, 2));
  });

  it('refuses to place into an occupied cell', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-0-0')).toBeDisabled());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('cell-0-0'));
    // Without this await the assertion passes whether or not a call fired (#4439).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('place', 0, 0);

    // Negative control: an empty cell does place.
    fireEvent.click(screen.getByTestId('cell-1-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('place', 1, 1));
  });

  // The starter existing but being unknown is the rule the game turns on, so
  // it renders face down rather than being left out.
  it('shows the starter face down during play', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cs-starter-facedown')).toBeInTheDocument());
  });

  it('reveals the starter once the board is full', async () => {
    mockExec.mockResolvedValue(makeState({ starter: card('CLOVER', 7), placedCount: 16, phase: 1 }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.queryByTestId('cs-starter-facedown')).not.toBeInTheDocument());
  });

  it('shows the total against the target', async () => {
    mockExec.mockResolvedValue(makeState({ totalScore: 44 }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('total-score')).toHaveTextContent('44'));
    expect(screen.getByTestId('total-score')).toHaveTextContent('61');
  });

  it('shows a per-hand breakdown once the hands score', async () => {
    const details = [{ fifteens: 4, pairs: 2, runs: 0, flush: 0, nobs: 0, total: 6 }, zero(), zero(), zero()];
    mockExec.mockResolvedValue(makeState({ rowDetails: details, rowScores: [6, 0, 0, 0], phase: 1 }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('row-breakdown-0')).toBeInTheDocument());
    expect(screen.getByTestId('row-breakdown-0')).toHaveTextContent('15が4');
    // A hand that scored nothing shows no breakdown at all.
    expect(screen.queryByTestId('row-breakdown-1')).not.toBeInTheDocument();
  });

  // Finishing at 40 is a completed game, not a win; the two must read differently.
  it('distinguishes clearing the target from falling short', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, totalScore: 40, isWin: false }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cs-verdict')).toHaveTextContent('未達成'));

    mockExec.mockResolvedValue(makeState({ phase: 1, totalScore: 64, isWin: true }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getAllByTestId('cs-verdict')[1]).toHaveTextContent('クリア'));
  });

  it('hides the verdict while the game is still running', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('total-score')).toBeInTheDocument());
    expect(screen.queryByTestId('cs-verdict')).not.toBeInTheDocument();
  });

  it('requests a hint', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cs-hint-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('cs-hint-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('shows the server hint only once it was requested', async () => {
    mockExec.mockResolvedValue(
      makeState({ hint: { row: 1, col: 2, score: 6, synergy: true }, messageCode: 'cribbagesquares.hintAvailable' }),
    );
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cs-server-hint')).toBeInTheDocument());
  });

  // The other half of the gate: a passive hint must not surface the banner.
  it('hides the server hint when it was not requested', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { row: 1, col: 2, score: 6, synergy: true } }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cs-board')).toBeInTheDocument());
    expect(screen.queryByTestId('cs-server-hint')).not.toBeInTheDocument();
  });

  it('undoes only when there is history', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: false }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());

    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: '元に戻す' })[1]).toBeEnabled());
    fireEvent.click(screen.getAllByRole('button', { name: '元に戻す' })[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('gives up through the confirm dialog', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('hides the playing controls once the game completes', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.queryByTestId('cs-hint-button')).not.toBeInTheDocument());
  });

  it('highlights the row and column of a focused empty cell', async () => {
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByTestId('cell-1-1')).toBeEnabled());

    fireEvent.focus(screen.getByTestId('cell-1-1'));
    await waitFor(() => expect(screen.getByTestId('row-score-1')).toHaveAttribute('data-cross-hover', 'true'));
    expect(screen.getByTestId('col-score-1')).toHaveAttribute('data-cross-hover', 'true');
    expect(screen.getByTestId('row-score-0')).not.toHaveAttribute('data-cross-hover');

    fireEvent.blur(screen.getByTestId('cell-1-1'));
    await waitFor(() => expect(screen.getByTestId('row-score-1')).not.toHaveAttribute('data-cross-hover'));
  });

  it('shows an error with a retry', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<CribbageSquaresPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /再試行|retry/i })).toBeInTheDocument());
  });
});
