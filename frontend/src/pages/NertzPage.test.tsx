import { fireEvent, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { nertzApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { NertzResponse } from '../types/card';
import { NertzPhase } from '../types/phases';
import { NertzPage } from './NertzPage';

vi.mock('../api/gameApi', () => ({
  nertzApi: { exec: vi.fn() },
  actionLogApi: { nertz: vi.fn() },
}));

const mockExec = vi.mocked(nertzApi.exec);

const baseConfig = {
  playerCount: 4,
  drawCount: 3,
  targetScore: 100,
  cpuDifficulty: 1,
  cpuTickMoves: 3,
};

const playingState: NertzResponse = {
  phase: NertzPhase.PLAYING,
  roundNo: 1,
  winnerIdx: -1,
  matchWinner: -1,
  moveCount: 0,
  canUndo: false,
  ...baseConfig,
  players: [
    {
      name: 'You',
      isCpu: false,
      deckIdx: 0,
      score: 0,
      nertzSize: 13,
      nertzTop: { design: 'HEART', value: 7 },
      tableau: [
        [{ card: { design: 'SPADE', value: 5 }, faceUp: true }],
        [{ card: { design: 'CLOVER', value: 6 }, faceUp: true }],
        [{ card: { design: 'DIAMOND', value: 9 }, faceUp: true }],
        [{ card: { design: 'HEART', value: 11 }, faceUp: true }],
      ],
      wasteSize: 0,
      stockSize: 35,
    },
    {
      name: 'CPU1',
      isCpu: true,
      deckIdx: 1,
      score: 0,
      nertzSize: 13,
      tableau: [[], [], [], []],
      wasteSize: 0,
      stockSize: 35,
    },
  ],
  foundations: Array.from({ length: 8 }, () => ({ suit: -1, size: 0 })),
  message: '',
};

const roundEndState: NertzResponse = {
  ...playingState,
  phase: NertzPhase.ROUND_END,
  winnerIdx: 0,
  moveCount: 17,
  canUndo: false,
};

const gameEndState: NertzResponse = {
  ...playingState,
  phase: NertzPhase.GAME_END,
  winnerIdx: 0,
  matchWinner: 0,
  canUndo: false,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
});

afterEach(() => {
  mockExec.mockReset();
});

describe('NertzPage', () => {
  it('shows loading message before state arrives', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    expect(screen.getByText(/Loading|読み込み/i)).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the human player with score, tableau, and stock counter', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/♥7/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '35' })).toBeInTheDocument();
  });

  it('clicking the stock dispatches a draw command', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: '35' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '35' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('d', { playerIdx: 0 }));
  });

  it('selecting nertz then a foundation dispatches a move', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/♥7/)).toBeInTheDocument());
    // Select nertz pile
    const nertzBtn = screen.getByText('♥7').closest('button');
    expect(nertzBtn).not.toBeNull();
    fireEvent.click(nertzBtn as HTMLElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // Click foundation 0 — aria-label uses the localized template (ja default in tests).
    fireEvent.click(screen.getByLabelText(/ファウンデーション0|Foundation 0/));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('m', {
        playerIdx: 0,
        from: { zone: 'nertz' },
        to: { zone: 'foundation', idx: 0 },
      }),
    );
  });

  it('shows the next-round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByRole('button', { name: /次ラウンド|Next round/ })).toBeInTheDocument());
  });

  it('renders game-end phase label when match is decided', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getAllByText(/ゲーム終了|Game over/).length).toBeGreaterThan(0));
  });

  it('starts CPU tick polling while round is active', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByText(/♥7/)).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('tick'), { timeout: 2000 });
  });
});
