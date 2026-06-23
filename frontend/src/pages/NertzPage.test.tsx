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
  roundNumber: 1,
  winnerIdx: -1,
  matchWinner: -1,
  moveCount: 0,
  canUndo: false,
  ...baseConfig,
  players: [
    {
      name: 'You',
      isHuman: true,
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
      isHuman: false,
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
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '35' })).toBeInTheDocument();
  });

  it('renders per-player score progress bars scaled to the target score', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      players: [
        { ...playingState.players[0], score: 50 }, // 50 / 100 = 50%
        { ...playingState.players[1], score: 25 }, // 25 / 100 = 25%
      ],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    const humanBar = await screen.findByTestId('nertz-scorebar-0');
    const cpuBar = screen.getByTestId('nertz-scorebar-1');
    expect(humanBar).toHaveStyle({ width: '50%' });
    expect(cpuBar).toHaveStyle({ width: '25%' });
    // Human bar is green, CPU bar is the warning accent; both animate only when motion is allowed.
    expect(humanBar.className).toContain('bg-ds-success');
    expect(cpuBar.className).toContain('bg-ds-warning');
    expect(humanBar.className).toContain('motion-safe:transition-[width]');
  });

  it('clamps the score bar to 0% for negative scores', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      players: [{ ...playingState.players[0], score: -20 }, playingState.players[1]],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    expect(await screen.findByTestId('nertz-scorebar-0')).toHaveStyle({ width: '0%' });
  });

  it('renders a 0% bar when targetScore is not positive', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      targetScore: 0, // guard against divide-by-zero → 0% width
      players: [{ ...playingState.players[0], score: 10 }, playingState.players[1]],
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    expect(await screen.findByTestId('nertz-scorebar-0')).toHaveStyle({ width: '0%' });
  });

  it('caps the score bar at 100% when the score exceeds the target', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      players: [{ ...playingState.players[0], score: 150 }, playingState.players[1]], // > targetScore 100
    });
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    expect(await screen.findByTestId('nertz-scorebar-0')).toHaveStyle({ width: '100%' });
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
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
    // Select nertz pile
    const nertzBtn = screen.getByAltText('♥ 7').closest('button');
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
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('tick'), { timeout: 2000 });
  });

  it('pressing "d" draws stock for the human', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('d', { playerIdx: 0 }));
  });

  it('pressing "1" while a card is selected dispatches a move to foundation 0', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());

    // Pick the Nertz pile via 'n' first, then ask for foundation 0.
    fireEvent.keyDown(document, { key: 'n' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: '1' });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('m', {
        playerIdx: 0,
        from: { zone: 'nertz' },
        to: { zone: 'foundation', idx: 0 },
      }),
    );
  });

  it('flashes the target foundation when a move to it is rejected (collision)', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♥ 7').closest('button') as HTMLElement);

    // Reject the next move (simulates CPU getting there first).
    mockExec.mockRejectedValueOnce(new Error('collision'));
    fireEvent.click(screen.getByTestId('nertz-foundation-0'));
    await waitFor(() => expect(screen.getByTestId('nertz-foundation-0')).toHaveAttribute('data-collided', 'true'));
    expect(screen.getByTestId('nertz-foundation-0').className).toMatch(/animate-shake/);
  });

  it('does not blow up the page when a foundation move is rejected (game continues)', async () => {
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♥ 7').closest('button') as HTMLElement);

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(screen.getByTestId('nertz-foundation-1'));
    await waitFor(() => expect(screen.getByTestId('nertz-foundation-1')).toHaveAttribute('data-collided', 'true'));
    // Foundation grid is still rendered; the page didn't unmount into ErrorAlert.
    expect(screen.getAllByTestId(/nertz-foundation-/)).toHaveLength(8);
  });

  it('does not attribute a later unrelated error to a previously-resolved foundation move', async () => {
    // Bug 1 from the gemini/Claude review: after a foundation move succeeds, the
    // pending-foundation ref must be cleared so a later unrelated error (e.g. a
    // failing tick) does not flash that foundation.
    renderWithProviders(
      <MemoryRouter initialEntries={['/nertz']}>
        <NertzPage />
      </MemoryRouter>,
    );
    await waitFor(() => expect(screen.getByAltText('♥ 7')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♥ 7').closest('button') as HTMLElement);

    // Successful foundation move → state updates → pending ref must be cleared.
    fireEvent.click(screen.getByTestId('nertz-foundation-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('m', expect.objectContaining({})));

    // Now a different action fails — must NOT light up foundation 2.
    mockExec.mockRejectedValueOnce(new Error('tick boom'));
    fireEvent.click(screen.getByAltText('♥ 7').closest('button') as HTMLElement); // re-select
    fireEvent.click(screen.getByTestId('nertz-foundation-3')); // dispatch to a different cell
    // Only foundation 3 (the new target) should be marked, not foundation 2.
    await waitFor(() => expect(screen.getByTestId('nertz-foundation-3')).toHaveAttribute('data-collided', 'true'));
    expect(screen.getByTestId('nertz-foundation-2')).not.toHaveAttribute('data-collided');
  });
});
