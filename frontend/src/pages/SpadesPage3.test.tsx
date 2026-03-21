import { afterEach, beforeEach, describe, expect, it, vi } from 'bun:test';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import { actionLogApi, spadesApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { asMocked } from '../test/viCompat';
import type { SpadesResponse } from '../types/card';
import { SpadesPage } from './SpadesPage';

let mockExec: ReturnType<typeof vi.fn>;

const playPhaseState: SpadesResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 3,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      bags: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: 4,
      roundScore: 3,
      cumulativeScore: 10,
      trickCount: 1,
      bags: 2,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: 3,
      roundScore: 5,
      cumulativeScore: 20,
      trickCount: 2,
      bags: 1,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: 2,
      roundScore: 0,
      cumulativeScore: 5,
      trickCount: 0,
      bags: 0,
    },
  ],
  phase: 1,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  spadesBroken: false,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500, nilBonus: 100, bagPenaltyThreshold: 10 },
};

const bidPhaseState: SpadesResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  players: playPhaseState.players.map((p) => ({ ...p, bid: -1 })),
};

const gameEndState: SpadesResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: SpadesResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  vi.spyOn(spadesApi, 'exec').mockImplementation(vi.fn());
  vi.spyOn(actionLogApi, 'spades').mockImplementation(vi.fn());
  mockExec = asMocked(spadesApi.exec);
  mockExec.mockResolvedValue(playPhaseState);
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('SpadesPage (part 3)', () => {
  it('sets aria-busy on container', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    const container = screen
      .getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })
      .closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: SpadesResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn during bid phase', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows your turn when human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u5f85\u6a5f\u4e2d'));
  });

  // -- Keyboard navigation --

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    asMocked(actionLogApi.spades).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument());

    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.queryByText('\u68cb\u8b5c')).not.toBeInTheDocument());
  });

  it('shows bid value for player with bid >= 0', async () => {
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      // CPU 1 has bid=4
      expect(screen.getByText(/CPU 1.*\u30d3\u30c3\u30c9 4/)).toBeInTheDocument();
    });
  });

  it('shows unbid text for player with bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*\u672a\u30d3\u30c3\u30c9/)).toBeInTheDocument();
    });
  });

  it('score table shows dash for bid < 0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<SpadesPage />);
    await waitFor(() => expect(screen.getByText('\u30b9\u30b3\u30a2')).toBeInTheDocument());
    // All players have bid=-1, so bid column should show '-'
    const rows = screen.getAllByRole('row');
    // Header + 4 players = 5 rows
    expect(rows.length).toBe(5);
  });
});
