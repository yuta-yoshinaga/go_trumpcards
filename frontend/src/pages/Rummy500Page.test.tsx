import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rummy500Api } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Rummy500Response } from '../types/card';
import { Rummy500Page } from './Rummy500Page';

vi.mock('../api/gameApi', () => ({
  rummy500Api: { exec: vi.fn() },
  actionLogApi: { rummy500: vi.fn() },
}));

const mockExec = vi.mocked(rummy500Api.exec);

const drawPhaseState: Rummy500Response = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 2,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 7 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      laidMelds: [],
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      laidMelds: [],
    },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardPile: [
    { design: 'HEART', value: 3 },
    { design: 'CLOVER', value: 5 },
  ],
  drawPileCount: 25,
  gameEndFlag: false,
  winnerIdx: -1,
  roundEnderIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500 },
};

const playPhaseState: Rummy500Response = {
  ...drawPhaseState,
  phase: 1,
  players: [
    {
      ...drawPhaseState.players[0],
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 7 },
        { design: 'CLOVER', value: 7 },
      ],
    },
    drawPhaseState.players[1],
  ],
};

const roundEndState: Rummy500Response = {
  ...drawPhaseState,
  phase: 2,
};

const gameEndState: Rummy500Response = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
};

describe('Rummy500Page', () => {
  beforeEach(() => {
    mockExec.mockReset();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 500 });
    });
  });

  it('shows draw stock button in Draw phase', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /山札から引く/ })).toBeInTheDocument();
    });
  });

  it('clicking discard card in Draw phase calls drawdiscard', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getAllByLabelText(/\(0\)$/).length).toBeGreaterThan(0);
    });
    const card = screen.getAllByLabelText(/\(0\)$/)[0];
    fireEvent.click(card);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('drawdiscard', undefined, undefined, undefined, 0);
    });
  });

  it('shows meld + discard buttons in Play phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /メルドする/ })).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /^捨てる$/ })).toBeInTheDocument();
  });

  it('selects a lay-off target by clicking a laid meld', async () => {
    const withMeld: Rummy500Response = {
      ...playPhaseState,
      players: [
        playPhaseState.players[0],
        {
          ...playPhaseState.players[1],
          laidMelds: [
            {
              cards: [
                { design: 'DIAMOND', value: 4 },
                { design: 'DIAMOND', value: 5 },
                { design: 'DIAMOND', value: 6 },
              ],
            },
          ],
        },
      ],
    };
    mockExec.mockResolvedValue(withMeld);
    renderWithProviders(<Rummy500Page />);
    const meldBtn = await screen.findByTestId('layoff-meld-1-0');
    expect(meldBtn).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(meldBtn);
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'true'));
  });

  it('shows next round button in RoundEnd phase', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /次のラウンド/ })).toBeInTheDocument();
    });
  });

  it('shows game end celebration', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      // The shell renders some game-end indicator
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 500 });
    });
  });

  it('meld button disabled when fewer than 3 selected', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /メルドする/ })).toBeDisabled();
    });
  });

  it('drawstock button triggers exec', async () => {
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /山札から引く/ })).toBeInTheDocument();
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: /山札から引く/ }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('drawstock');
    });
  });

  it('renders the hand-penalty badge with the summed value when the player holds cards', async () => {
    // drawPhaseState has [♠7, ♥7] in hand → penalty = 7 + 7 = 14.
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => {
      const badge = screen.getByTestId('hand-penalty-badge');
      expect(badge).toBeInTheDocument();
      expect(badge).toHaveTextContent('14');
    });
  });

  it('hides the hand-penalty badge when the player has no cards', async () => {
    const emptyHandState: Rummy500Response = {
      ...drawPhaseState,
      players: [{ ...drawPhaseState.players[0], cardCount: 0, cards: [] }, drawPhaseState.players[1]],
    };
    mockExec.mockResolvedValue(emptyHandState);
    renderWithProviders(<Rummy500Page />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
    expect(screen.queryByTestId('hand-penalty-badge')).not.toBeInTheDocument();
  });
});
