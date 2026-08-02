import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, threethirteenApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { ThreeThirteenResponse } from '../types/card';
import { ThreeThirteenPage } from './ThreeThirteenPage';

vi.mock('../api/gameApi', () => ({
  threethirteenApi: { exec: vi.fn() },
  actionLogApi: { threethirteen: vi.fn() },
}));

const mockExec = vi.mocked(threethirteenApi.exec);

const RESET_CONFIG = { cpuDifficulty: 1, playerCount: 2 };

const drawPhaseState: ThreeThirteenResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      deadwood: 11,
      roundScore: 0,
      cumulativeScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 4, cards: [], deadwood: 0, roundScore: 3, cumulativeScore: 10 },
  ],
  phase: 0,
  round: 2,
  wildRank: 4,
  dealCount: 4,
  currentPlayerIdx: 0,
  knockerIdx: -1,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, playerCount: 2 },
};

const discardPhaseState: ThreeThirteenResponse = { ...drawPhaseState, phase: 1 };
const roundEndState: ThreeThirteenResponse = { ...drawPhaseState, phase: 2 };
const gameEndState: ThreeThirteenResponse = {
  ...drawPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const cpuTurnState: ThreeThirteenResponse = { ...drawPhaseState, currentPlayerIdx: 1 };

beforeEach(() => {
  mockExec.mockResolvedValue(drawPhaseState);
});

describe('ThreeThirteenPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ThreeThirteenPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, RESET_CONFIG));
  });

  it('shows round, wild rank, and deal count banner', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByTestId('threethirteen-round-banner')).toBeInTheDocument());
    const banner = screen.getByTestId('threethirteen-round-banner');
    expect(banner).toHaveTextContent('ワイルド: 4');
    expect(banner).toHaveTextContent('ラウンド 2/11');
  });

  it('shows deadwood indicator during discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByTestId('threethirteen-deadwood-indicator')).toBeInTheDocument());
  });

  it('shows predicted post-discard deadwood that changes with card selection', async () => {
    // ♠5-6-7 form a run; ♥K is the odd card. Wild rank 2 does not match any card.
    const meldHandState: ThreeThirteenResponse = {
      ...discardPhaseState,
      wildRank: 2,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'SPADE', value: 6 },
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 13 },
          ],
          deadwood: 10,
          roundScore: 0,
          cumulativeScore: 0,
        },
        discardPhaseState.players[1],
      ],
    };
    mockExec.mockResolvedValue(meldHandState);
    renderWithProviders(<ThreeThirteenPage />);

    // No selection → best single discard drops ♥K, leaving the run → 0.
    await waitFor(() =>
      expect(screen.getByTestId('threethirteen-deadwood-indicator')).toHaveTextContent('予測デッドウッド: 0点'),
    );

    // Selecting a run card (♠5) breaks the meld → ♠6 + ♠7 + ♥K = 23.
    fireEvent.click(screen.getByAltText('♠ 5').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('threethirteen-deadwood-indicator')).toHaveTextContent('予測デッドウッド: 23点');
  });

  it('badges wild-rank cards in the hand and on the discard top', async () => {
    // Aces wild: the ♠A in hand and the ♣A discard top both match wildRank 1.
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      wildRank: 1,
      discardTop: { design: 'CLOVER', value: 1 },
    });
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // One hand card (♠A) + the discard top (♣A) → two wild badges. The ♥J is not wild.
    expect(screen.getAllByTestId('tt-wild-badge')).toHaveLength(2);
  });

  it('shows no wild badges when no card matches the wild rank', async () => {
    // Default wildRank 4: neither the ♠A/♥J hand nor the ♥7 discard top matches.
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.queryByTestId('tt-wild-badge')).not.toBeInTheDocument();
  });

  it('renders draw phase with human cards', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('calls drawstock command when draw stock button is clicked', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('calls drawdiscard command when draw discard button is clicked', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨て札から引く' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('renders discard and knock buttons when human discard turn', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'ノック' })).toBeInTheDocument();
    });
  });

  it('calls discard command when discard button is clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('calls knock command (which discards a card) when knock button is clicked', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(screen.getByRole('button', { name: 'ノック' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock', 0));
  });

  it('shows next round button on round end and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('renders three opponents for a 4-player game', async () => {
    const fourPlayerState: ThreeThirteenResponse = {
      ...drawPhaseState,
      players: [
        drawPhaseState.players[0],
        { id: 1, isHuman: false, cardCount: 4, cards: [], deadwood: 0, roundScore: 0, cumulativeScore: 0 },
        { id: 2, isHuman: false, cardCount: 4, cards: [], deadwood: 0, roundScore: 0, cumulativeScore: 0 },
        { id: 3, isHuman: false, cardCount: 4, cards: [], deadwood: 0, roundScore: 0, cumulativeScore: 0 },
      ],
      config: { ...drawPhaseState.config, playerCount: 4 },
    };
    mockExec.mockResolvedValue(fourPlayerState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*4枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*4枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*4枚/)).toBeInTheDocument();
    });
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('settings panel changes playerCount', async () => {
    renderWithProviders(<ThreeThirteenPage />);
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
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });

  it('Enter key triggers discard in discard phase', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(drawPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<ThreeThirteenPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.threethirteen).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.threethirteen).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<ThreeThirteenPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
