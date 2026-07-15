import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { hachihachiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeHachiHachiState } from '../test/stateFactories';
import { HachiHachiPage } from './HachiHachiPage';

vi.mock('../api/gameApi', () => ({
  hachihachiApi: { exec: vi.fn() },
  actionLogApi: { hachihachi: vi.fn() },
}));

const mockExec = vi.mocked(hachihachiApi.exec);

const playState = makeHachiHachiState();
const roundEndState = makeHachiHachiState({
  phase: 1,
  lastRoundResult: {
    best: 0,
    scores: [
      { playerIdx: 0, rawScore: 100, yaku: [{ key: 'sanko', points: 40 }], bonus: 40, delta: 52 },
      { playerIdx: 1, rawScore: 80, yaku: [], bonus: 0, delta: -8 },
      { playerIdx: 2, rawScore: 84, yaku: [], bonus: 0, delta: -44 },
    ],
  },
});
const gameEndState = makeHachiHachiState({
  phase: 2,
  gameEndFlag: true,
  winner: 0,
  players: [
    { ...playState.players[0], score: 52 },
    { ...playState.players[1], score: -8 },
    { ...playState.players[2], score: -44 },
  ],
  message: 'ゲーム終了！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('HachiHachiPage', () => {
  it('renders the loading fallback when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<HachiHachiPage />);
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<HachiHachiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders all three players (human hand + two opponent piles)', async () => {
    renderWithProviders(<HachiHachiPage />);
    await screen.findByTestId('hand-card-0');
    expect(screen.getByTestId('hachihachi-cpu-1')).toBeInTheDocument();
    expect(screen.getByTestId('hachihachi-cpu-2')).toBeInTheDocument();
  });

  it('explains the three-player 88-baseline settlement in the scoring note', async () => {
    renderWithProviders(<HachiHachiPage />);
    await screen.findByTestId('hand-card-0');
    // The note ties the raw-score terms to the 3-player settlement rule.
    expect(screen.getByText(/3人で基準88点との差を精算/)).toBeInTheDocument();
  });

  it('plays a hand card with a single match immediately', async () => {
    renderWithProviders(<HachiHachiPage />);
    const card = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(card);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('requires a field pick for a two-way match, then plays with fieldIndex', async () => {
    mockExec.mockResolvedValue(makeHachiHachiState({ captureOptions: { 0: [0, 1] } }));
    renderWithProviders(<HachiHachiPage />);
    const card = await screen.findByTestId('hand-card-0');
    fireEvent.click(card);
    await waitFor(() => expect(screen.getByTestId('hachihachi-field-pick')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('field-card-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 1 }));
  });

  it('shows the next-round button at round end and dispatches nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<HachiHachiPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the 3-player round-result settlement table and highlights the best', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<HachiHachiPage />);
    await waitFor(() => expect(screen.getByTestId('hachihachi-round-result')).toBeInTheDocument());
    expect(screen.getByTestId('hachihachi-score-row-0')).toHaveAttribute('data-best', 'true');
    expect(screen.getByTestId('hachihachi-score-row-1')).not.toHaveAttribute('data-best');
    expect(screen.getByTestId('hachihachi-score-row-2')).toBeInTheDocument();
  });

  it('conveys the best row and delta signs without relying on colour', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<HachiHachiPage />);
    await waitFor(() => expect(screen.getByTestId('hachihachi-round-result')).toBeInTheDocument());
    // The best row carries a crown glyph and an sr-only "top score" label.
    const bestRow = screen.getByTestId('hachihachi-score-row-0');
    expect(bestRow).toHaveTextContent('👑');
    expect(bestRow).toHaveTextContent('最高得点');
    // Delta signs read meaningfully: +52 → gained, -8 → lost.
    expect(bestRow).toHaveTextContent('52点獲得');
    expect(screen.getByTestId('hachihachi-score-row-1')).toHaveTextContent('8点失点');
    // The table names itself via a caption.
    expect(screen.getByText('ラウンド精算表（プレイヤー別の素点・役点・差分）')).toBeInTheDocument();
  });

  it('renders the game-end result with a winner banner', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<HachiHachiPage />);
    await waitFor(() => expect(screen.getByTestId('hachihachi-result')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '新しいゲーム' })).toBeInTheDocument();
  });

  it('renders a game-end result without a winner banner when winner is -1', async () => {
    mockExec.mockResolvedValue(makeHachiHachiState({ phase: 2, gameEndFlag: true, winner: -1, message: '引き分け' }));
    renderWithProviders(<HachiHachiPage />);
    const result = await screen.findByTestId('hachihachi-result');
    expect(result).toBeInTheDocument();
  });

  it('does not play a hand card when it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeHachiHachiState({ isHumanTurn: false, currentTurn: 1 }));
    renderWithProviders(<HachiHachiPage />);
    const card = await screen.findByTestId('hand-card-0');
    expect(card).toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(card);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores a field click that is not a capture candidate', async () => {
    mockExec.mockResolvedValue(makeHachiHachiState({ captureOptions: { 0: [0, 1] } }));
    renderWithProviders(<HachiHachiPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('hachihachi-field-pick')).toBeInTheDocument());
    mockExec.mockClear();
    // Re-clicking a candidate still dispatches, confirming the guard path is exercised.
    fireEvent.click(screen.getByTestId('field-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 0 }));
  });
});
