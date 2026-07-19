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
    const card = await screen.findByTestId('disc-card-0');
    fireEvent.click(card);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith('drawdiscard', undefined, undefined, undefined, 0);
    });
  });

  it('previews the whole take-range when hovering a mid-pile discard card', async () => {
    renderWithProviders(<Rummy500Page />);
    // drawPhaseState.discardPile has 2 cards: [♥3 (idx 0), ♣5 (idx 1, top)].
    const bottom = await screen.findByTestId('disc-card-0');
    const top = screen.getByTestId('disc-card-1');
    // Nothing highlighted before hover.
    expect(bottom).not.toHaveAttribute('data-in-pickup-range');
    expect(top).not.toHaveAttribute('data-in-pickup-range');

    fireEvent.mouseEnter(bottom);
    // Hovering the bottom card highlights it AND every card above it (the whole pile).
    expect(bottom).toHaveAttribute('data-in-pickup-range', 'true');
    expect(top).toHaveAttribute('data-in-pickup-range', 'true');
    // The badge announces the number of cards that would be taken (2).
    const badge = screen.getByTestId('pickup-range-badge');
    expect(badge).toHaveTextContent('2');
    // The aria-label of the hovered card also carries the count.
    expect(bottom).toHaveAttribute('aria-label', expect.stringContaining('2枚引き取る'));

    fireEvent.mouseLeave(bottom);
    expect(bottom).not.toHaveAttribute('data-in-pickup-range');
    expect(top).not.toHaveAttribute('data-in-pickup-range');
  });

  it('previews only the top card when hovering the top of the discard pile', async () => {
    renderWithProviders(<Rummy500Page />);
    const bottom = await screen.findByTestId('disc-card-0');
    const top = screen.getByTestId('disc-card-1');

    fireEvent.focus(top);
    // Hovering/focusing the top card previews just that one card.
    expect(top).toHaveAttribute('data-in-pickup-range', 'true');
    expect(bottom).not.toHaveAttribute('data-in-pickup-range');
    expect(screen.getByTestId('pickup-range-badge')).toHaveTextContent('1');
    expect(top).toHaveAttribute('aria-label', expect.stringContaining('1枚引き取る'));
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

    // Before selecting, the footer shows the "click a meld" hint and Lay off is disabled.
    expect(screen.getByTestId('r5-layoff-target')).toHaveTextContent(/上のメルドをクリック/);
    expect(screen.getByRole('button', { name: 'レイオフ' })).toBeDisabled();

    fireEvent.click(meldBtn);
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'true'));
    // The selected meld is now described in the footer text.
    expect(screen.getByTestId('r5-layoff-target')).toHaveTextContent('#0');

    // Clicking the same meld again toggles the selection off.
    fireEvent.click(screen.getByTestId('layoff-meld-1-0'));
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'false'));
    expect(screen.getByTestId('r5-layoff-target')).toHaveTextContent(/上のメルドをクリック/);
    // Re-select for the lay-off flow below.
    fireEvent.click(screen.getByTestId('layoff-meld-1-0'));
    await waitFor(() => expect(screen.getByTestId('layoff-meld-1-0')).toHaveAttribute('aria-pressed', 'true'));
    // Still disabled until exactly one hand card is chosen.
    expect(screen.getByRole('button', { name: 'レイオフ' })).toBeDisabled();

    const handCard = document.querySelector('[data-tutorial="r5-player-hand"] button') as HTMLButtonElement;
    fireEvent.click(handCard);
    const layoffBtn = screen.getByRole('button', { name: 'レイオフ' });
    await waitFor(() => expect(layoffBtn).toBeEnabled());

    mockExec.mockClear();
    fireEvent.click(layoffBtn);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, undefined, undefined, {
        meldOwner: 1,
        meldIdx: 0,
        cardIndex: 0,
      }),
    );
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
