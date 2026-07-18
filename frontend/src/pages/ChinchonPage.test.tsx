import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, chinchonApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { ChinchonResponse } from '../types/card';
import { ChinchonPage } from './ChinchonPage';

vi.mock('../api/gameApi', () => ({
  chinchonApi: { exec: vi.fn() },
  actionLogApi: { chinchon: vi.fn() },
}));

const mockExec = vi.mocked(chinchonApi.exec);

const RESET_CONFIG = { cpuDifficulty: 1, playerCount: 2, knockThreshold: 5, eliminationLimit: 100 };

const drawPhaseState: ChinchonResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 7,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      eliminated: false,
    },
    { id: 1, isHuman: false, cardCount: 7, cards: [], roundScore: 3, cumulativeScore: 10, eliminated: false },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  knockerIdx: -1,
  knockerMelds: [],
  message: '',
  config: { cpuDifficulty: 1, playerCount: 2, knockThreshold: 5, eliminationLimit: 100 },
};

const discardPhaseState: ChinchonResponse = { ...drawPhaseState, phase: 1 };

const layoffPhaseState: ChinchonResponse = {
  ...drawPhaseState,
  phase: 2,
  knockerIdx: 1,
  knockerMelds: [
    {
      cards: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 3 },
        { design: 'DIAMOND', value: 3 },
      ],
    },
  ],
};

const roundEndState: ChinchonResponse = { ...drawPhaseState, phase: 3 };

const gameEndState: ChinchonResponse = {
  ...drawPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const gameEndByFlagState: ChinchonResponse = {
  ...drawPhaseState,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: ChinchonResponse = { ...drawPhaseState, currentPlayerIdx: 1 };

beforeEach(() => {
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('ChinchonPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ChinchonPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, RESET_CONFIG));
  });

  it('shows live deadwood indicator during discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByTestId('chinchon-deadwood-indicator')).toBeInTheDocument());
  });

  it('pulses the knock button when deadwood is at or below the threshold', async () => {
    const lowDeadwoodHand: ChinchonResponse = {
      ...discardPhaseState,
      players: [
        {
          ...discardPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'SPADE', value: 6 },
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 1 },
          ],
        },
        ...discardPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(lowDeadwoodHand);
    renderWithProviders(<ChinchonPage />);
    const knockBtn = await screen.findByTestId('chinchon-knock-button');
    expect(knockBtn.className).toContain('animate-pulse');
  });

  it('shows a per-card deadwood breakdown during the discard phase', async () => {
    // No melds possible; best discard drops ♥K, leaving ♠5 + ♣3 = 8 deadwood.
    const handWithDeadwood: ChinchonResponse = {
      ...discardPhaseState,
      players: [
        {
          ...discardPhaseState.players[0],
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 13 },
            { design: 'CLOVER', value: 3 },
          ],
        },
        ...discardPhaseState.players.slice(1),
      ],
    };
    mockExec.mockResolvedValue(handWithDeadwood);
    renderWithProviders(<ChinchonPage />);
    const bd = await screen.findByTestId('chinchon-deadwood-breakdown');
    expect(bd).toHaveTextContent('5 + 3 = 8');
  });

  it('renders draw phase with human cards', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('renders draw stock and draw discard buttons when human draw turn', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument();
    });
  });

  it('calls drawstock command when draw stock button is clicked', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard command when draw discard button is clicked', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('renders discard and knock buttons when human discard turn', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'ノック' })).toBeInTheDocument();
    });
  });

  it('calls discard command when discard button is clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('calls knock command (which discards a card) when knock button is clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(layoffPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ノック' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock', 0));
  });

  it('renders layoff and skip buttons when human layoff turn', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'レイオフ' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
    });
  });

  it('calls layoff command when layoff button is clicked', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: 'レイオフ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, [0]));
  });

  it('calls layoff with empty array when skip button is clicked', async () => {
    mockExec.mockResolvedValue(layoffPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, []));
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-4 phase (instant Chinchón win)', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('renders three opponents for a 4-player game', async () => {
    const fourPlayerState: ChinchonResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        { id: 1, isHuman: false, cardCount: 7, cards: [], roundScore: 0, cumulativeScore: 0, eliminated: false },
        { id: 2, isHuman: false, cardCount: 7, cards: [], roundScore: 0, cumulativeScore: 0, eliminated: false },
        { id: 3, isHuman: false, cardCount: 7, cards: [], roundScore: 0, cumulativeScore: 0, eliminated: false },
      ],
      config: { ...drawPhaseState.config, playerCount: 4 },
    };
    mockExec.mockResolvedValue(fourPlayerState);
    renderWithProviders(<ChinchonPage />);
    // Each CPU appears in the sidebar (with a card count) and the score table.
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*7枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*7枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*7枚/)).toBeInTheDocument();
    });
  });

  it('marks an eliminated player in the score table', async () => {
    const eliminatedState: ChinchonResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 120, eliminated: true },
      ],
    };
    mockExec.mockResolvedValue(eliminatedState);
    const { container } = renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(container.querySelector('.line-through')).toBeInTheDocument();
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, RESET_CONFIG));
  });

  it('settings panel changes playerCount', async () => {
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '4' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ...RESET_CONFIG, playerCount: 4 }));
  });

  it('does not show draw buttons when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });

  it('Enter key triggers discard in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.chinchon).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.chinchon).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
  });

  // Hand with a ♠5-6-7 run (idx 0-2) plus ♥5, ♥6 deadwood (idx 3-4).
  const meldHandState: ChinchonResponse = {
    ...discardPhaseState,
    players: [
      {
        ...discardPhaseState.players[0],
        cards: [
          { design: 'SPADE', value: 5 },
          { design: 'SPADE', value: 6 },
          { design: 'SPADE', value: 7 },
          { design: 'HEART', value: 5 },
          { design: 'HEART', value: 6 },
        ],
      },
      ...discardPhaseState.players.slice(1),
    ],
  };

  it('color-codes meld vs deadwood hand cards during the discard phase', async () => {
    mockExec.mockResolvedValue(meldHandState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByTestId('chinchon-hand-card-0')).toBeInTheDocument());

    // Run members are melds; the two loose hearts are deadwood.
    expect(screen.getByTestId('chinchon-hand-card-0')).toHaveAttribute('data-meld', 'meld');
    expect(screen.getByTestId('chinchon-hand-card-1')).toHaveAttribute('data-meld', 'meld');
    expect(screen.getByTestId('chinchon-hand-card-2')).toHaveAttribute('data-meld', 'meld');
    expect(screen.getByTestId('chinchon-hand-card-3')).toHaveAttribute('data-meld', 'deadwood');
    expect(screen.getByTestId('chinchon-hand-card-4')).toHaveAttribute('data-meld', 'deadwood');

    // Legend distinguishes the two categories with a non-color cue (label text).
    expect(screen.getByTestId('chinchon-meld-legend')).toHaveTextContent('メルド');
    expect(screen.getByTestId('chinchon-meld-legend')).toHaveTextContent('デッドウッド');
  });

  it('re-splits the meld coloring when a discard candidate is selected', async () => {
    mockExec.mockResolvedValue(meldHandState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByTestId('chinchon-hand-card-1')).toBeInTheDocument());

    // Projecting the discard of ♠5 breaks the run, so ♠6 is no longer a meld.
    fireEvent.click(screen.getByTestId('chinchon-hand-card-0'));
    expect(screen.getByTestId('chinchon-hand-card-1')).toHaveAttribute('data-meld', 'deadwood');
    expect(screen.getByTestId('chinchon-hand-card-2')).toHaveAttribute('data-meld', 'deadwood');
  });

  it('does not color-code hand cards outside the discard phase', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<ChinchonPage />);
    await waitFor(() => expect(screen.getByTestId('chinchon-hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('chinchon-hand-card-0')).not.toHaveAttribute('data-meld');
    expect(screen.queryByTestId('chinchon-meld-legend')).not.toBeInTheDocument();
  });
});
