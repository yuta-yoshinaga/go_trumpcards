import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, blackjackApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackJackCpuSeat, BlackJackHand, BlackJackResponse } from '../types/card';
import { BlackJackPage } from './BlackJackPage';

vi.mock('../api/gameApi', () => ({
  blackjackApi: { exec: vi.fn() },
  actionLogApi: { blackjack: vi.fn() },
}));

const mockExec = vi.mocked(blackjackApi.exec);

const betPhaseState: BlackJackResponse = {
  dealer: { chips: 1000 },
  player: { chips: 1000 },
  phase: 1,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

const baseHand: BlackJackHand = {
  score: 15,
  cards: [
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 10 },
  ],
  bet: 100,
  stood: false,
  doubled: false,
  busted: false,
  isBlackJack: false,
  canSplit: false,
  surrendered: false,
  canSurrender: false,
};

const actionPhaseState: BlackJackResponse = {
  dealer: { cards: [{ design: 'SPADE', value: 1 }], chips: 1000 },
  player: { chips: 900 },
  hands: [{ ...baseHand }],
  phase: 4,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

const insurancePhaseState: BlackJackResponse = {
  dealer: { cards: [{ design: 'SPADE', value: 1 }], chips: 1000 },
  player: { chips: 900 },
  hands: [{ ...baseHand }],
  phase: 3,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: true,
  message: '',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

const endPhaseState: BlackJackResponse = {
  dealer: {
    score: 19,
    cards: [
      { design: 'CLOVER', value: 9 },
      { design: 'DIAMOND', value: 10 },
    ],
    chips: 1000,
  },
  player: { chips: 1150 },
  hands: [
    {
      score: 21,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 10 },
      ],
      bet: 100,
      stood: true,
      doubled: false,
      busted: false,
      isBlackJack: true,
      canSplit: false,
      surrendered: false,
      canSurrender: false,
    },
  ],
  phase: 5,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: 'You are the winner.',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(betPhaseState);
});

describe('BlackJackPage', () => {
  it('shows the out-of-chips restart button instead of bet controls at zero chips', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, player: { chips: 0 } });
    renderWithProviders(<BlackJackPage />);
    const restart = await screen.findByRole('button', { name: /チップ不足/ });
    expect(screen.queryByRole('button', { name: 'ベット' })).not.toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(restart);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.anything()));
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BlackJackPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows chip info bar', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー: 1000 chips/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー: 1000 chips/)).toBeInTheDocument();
  });

  it('shows bet button in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
  });

  it('shows bet amount input in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());
  });

  it('calls bet command when bet button is clicked', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 10, undefined, {}));
  });

  it('shows hit and stand buttons in action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('shows double down button when 2 cards and sufficient chips', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeInTheDocument());
  });

  it('shows split button when canSplit and sufficient chips', async () => {
    const splitState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, canSplit: true }],
    };
    mockExec.mockResolvedValue(splitState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スプリット' })).toBeInTheDocument());
  });

  it('does not show split button when canSplit is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  it('shows insurance buttons in insurance phase', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'インシュランス' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '辞退' })).toBeInTheDocument();
  });

  it('calls insurance command when insurance button is clicked', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'インシュランス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('insurance'));
  });

  it('calls declineinsurance command when decline button is clicked', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '辞退' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declineinsurance'));
  });

  it('shows reset button in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
  });

  it('shows variant bonus badges at end phase when bonuses are present', async () => {
    mockExec.mockResolvedValue({ ...endPhaseState, bonuses: ['spanish21.bonus.777.spade'] });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByTestId('bj-bonus-badges')).toBeInTheDocument());
    expect(screen.getAllByTestId('bj-bonus-badge')).toHaveLength(1);
  });

  it('shows no bonus badges at end phase without bonuses', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByTestId('bj-bonus-badges')).not.toBeInTheDocument();
  });

  it('shows message overlay when message is non-empty', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('You are the winner.')).toBeInTheDocument());
  });

  it('does not show message overlay when message is empty', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('ヒット')).toBeInTheDocument());
    expect(screen.queryByText('You are the winner.')).not.toBeInTheDocument();
  });

  it('calls hit command when Hit button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hit'));
  });

  it('calls stand command when Stand button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('displays player score and bet', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 15/)).toBeInTheDocument());
    expect(screen.getByText(/ベット:? ?100/)).toBeInTheDocument();
  });

  it('displays dealer score when non-zero in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 19/)).toBeInTheDocument());
  });

  it('shows card back when dealer score is zero', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const imgs = screen.getAllByRole('img');
      const cardBackImg = imgs.find((img) => img.getAttribute('src') === '/images/z01.png');
      expect(cardBackImg).toBeInTheDocument();
    });
  });

  it('shows BJ flag with tooltip for blackjack hand', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const elem = screen.getByTitle(i18n.t('blackjack:status.bjTooltip'));
      expect(elem).toBeInTheDocument();
      expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.bj')}]`);
    });
  });

  it('does not show dealer area in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.queryByText('ディーラー手札')).not.toBeInTheDocument();
  });

  it('does not expand card area with flex-1 during bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).not.toHaveClass('flex-1');
  });

  it('expands card area with flex-1 during action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    expect(screen.getByTestId('card-area')).toHaveClass('flex-1');
  });

  it('disables bet button while loading', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: BlackJackResponse) => void;
    const slowPromise = new Promise<BlackJackResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();

    resolve(betPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());
  });

  it('disables action buttons while loading', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).not.toBeDisabled());

    let resolve!: (value: BlackJackResponse) => void;
    const slowPromise = new Promise<BlackJackResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));

    expect(screen.getByRole('button', { name: 'ヒット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeDisabled();

    resolve(actionPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('clears error message on successful API call after failure', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE())).not.toBeInTheDocument());
  });

  it('shows [BUST] flag with tooltip for busted hand', async () => {
    const bustState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, busted: true }],
    };
    mockExec.mockResolvedValue(bustState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const elem = screen.getByTitle(i18n.t('blackjack:status.bustTooltip'));
      expect(elem).toBeInTheDocument();
      expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.bust')}]`);
    });
  });

  it('shows [DD] flag with tooltip for doubled hand', async () => {
    const ddState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, doubled: true }],
    };
    mockExec.mockResolvedValue(ddState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const elem = screen.getByTitle(i18n.t('blackjack:status.ddTooltip'));
      expect(elem).toBeInTheDocument();
      expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.dd')}]`);
    });
  });

  it('shows [SUR] flag with tooltip for surrendered player hand', async () => {
    const surState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, surrendered: true }],
    };
    mockExec.mockResolvedValue(surState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const elem = screen.getByTitle(i18n.t('blackjack:status.surTooltip'));
      expect(elem).toBeInTheDocument();
      expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.sur')}]`);
    });
  });

  it('shows insurance bet info when insuranceBet > 0', async () => {
    const insState: BlackJackResponse = { ...actionPhaseState, insuranceBet: 50 };
    mockExec.mockResolvedValue(insState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('インシュランス: 50')).toBeInTheDocument());
  });

  it('shows multiple hand labels when there are 2 hands', async () => {
    const multiHandState: BlackJackResponse = {
      ...actionPhaseState,
      currentHandIdx: 0,
      hands: [
        { ...baseHand },
        {
          ...baseHand,
          cards: [
            { design: 'CLOVER', value: 5 },
            { design: 'SPADE', value: 6 },
          ],
        },
      ],
    };
    mockExec.mockResolvedValue(multiHandState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByText(/ハンド 1/)).toBeInTheDocument();
      expect(screen.getByText(/ハンド 2/)).toBeInTheDocument();
      expect(screen.getByText(/\(\*\)/)).toBeInTheDocument();
    });
  });

  it('does not show double down button when hand has more than 2 cards', async () => {
    const threeCardState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [
        {
          ...baseHand,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'DIAMOND', value: 10 },
            { design: 'SPADE', value: 2 },
          ],
        },
      ],
    };
    mockExec.mockResolvedValue(threeCardState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ダブルダウン' })).not.toBeInTheDocument();
  });

  it('does not show double down button when player has insufficient chips', async () => {
    const lowChipState: BlackJackResponse = {
      ...actionPhaseState,
      player: { chips: 50 },
      hands: [{ ...baseHand, bet: 100 }],
    };
    mockExec.mockResolvedValue(lowChipState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ダブルダウン' })).not.toBeInTheDocument();
  });

  it('does not show split button when chips are insufficient even if canSplit is true', async () => {
    const lowChipSplitState: BlackJackResponse = {
      ...actionPhaseState,
      player: { chips: 50 },
      hands: [{ ...baseHand, bet: 100, canSplit: true }],
    };
    mockExec.mockResolvedValue(lowChipSplitState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  // --- New feature tests ---

  it('shows deck count selector in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('デッキ数:')).toBeInTheDocument());
  });

  it('calls setdeckcount when deck count is changed', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.change(screen.getByLabelText('デッキ数:'), { target: { value: '6' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('setdeckcount', 6));
  });

  it('shows hint toggle button with OFF state in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント OFF' })).toBeInTheDocument());
  });

  it('calls togglehint when hint toggle button is clicked', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント OFF' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('togglehint'));
  });

  it('shows hint button as ON when hintEnabled is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, hintEnabled: true });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント ON' })).toBeInTheDocument());
  });

  it('shows surrender button when canSurrender is true', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hands: [{ ...baseHand, canSurrender: true }],
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'サレンダー' })).toBeInTheDocument());
  });

  it('does not show surrender button when canSurrender is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'サレンダー' })).not.toBeInTheDocument();
  });

  it('calls surrender when surrender button is clicked', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hands: [{ ...baseHand, canSurrender: true }],
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'サレンダー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('surrender'));
  });

  it('shows hint banner when hint is enabled and action is suggested', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hintEnabled: true,
      suggestedAction: 1, // BJ_SUGGEST_HIT
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('推奨: ヒット')).toBeInTheDocument());
  });

  it('does not show hint banner when hintEnabled is false', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: false, suggestedAction: 1 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByText(/推奨:/)).not.toBeInTheDocument();
  });

  it('does not show hint banner when suggestedAction is none', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 0 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByText(/推奨:/)).not.toBeInTheDocument();
  });

  it('highlights hit button when hint suggests hit', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 1 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ヒット' })).toHaveClass('ring-2');
    });
  });

  it('highlights stand button when hint suggests stand', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 2 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'スタンド' })).toHaveClass('ring-2');
    });
  });

  it('highlights double down button when hint suggests double', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 3 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ダブルダウン' })).toHaveClass('ring-2');
    });
  });

  it('highlights double down button when hint suggests doubleStand', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 7 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ダブルダウン' })).toHaveClass('ring-2');
    });
  });

  it('shows hint banner for doubleStand suggestion', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 7 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('推奨: ダブルダウン')).toBeInTheDocument());
  });

  it('highlights split button when hint suggests split', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hintEnabled: true,
      suggestedAction: 4,
      hands: [{ ...baseHand, canSplit: true }],
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'スプリット' })).toHaveClass('ring-2');
    });
  });

  it('highlights surrender button when hint suggests surrender', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hintEnabled: true,
      suggestedAction: 5,
      hands: [{ ...baseHand, canSurrender: true }],
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'サレンダー' })).toHaveClass('ring-2');
    });
  });

  it('highlights decline button when hint suggests decline insurance', async () => {
    mockExec.mockResolvedValue({ ...insurancePhaseState, hintEnabled: true, suggestedAction: 6 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '辞退' })).toHaveClass('ring-2');
    });
  });

  it('shows [SUR] badge with tooltip on surrendered hand', async () => {
    const surrenderedEndState: BlackJackResponse = {
      ...endPhaseState,
      hands: [{ ...(endPhaseState.hands?.[0] as BlackJackHand), surrendered: true }],
    };
    mockExec.mockResolvedValue(surrenderedEndState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const elem = screen.getByTitle(i18n.t('blackjack:status.surTooltip'));
      expect(elem).toBeInTheDocument();
      expect(elem).toHaveTextContent(`[${i18n.t('blackjack:status.sur')}]`);
    });
  });

  it('shows deck count in chip bar', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, deckCount: 6 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('デッキ: 6デッキ')).toBeInTheDocument());
  });

  it('sets aria-busy while loading', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');

    let resolve!: (value: BlackJackResponse) => void;
    const slowPromise = new Promise<BlackJackResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');

    resolve(betPhaseState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
    });
  });

  // --- S17/H17 and counting toggle tests ---

  it('shows S17 button in bet phase by default', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'S17' })).toBeInTheDocument());
  });

  it('shows H17 button when dealerHitsSoft17 is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, dealerHitsSoft17: true });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'H17' })).toBeInTheDocument());
  });

  it('calls togglesoft17 when S17 button is clicked', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, dealerHitsSoft17: true });
    fireEvent.click(screen.getByRole('button', { name: 'S17' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('togglesoft17'));
  });

  it('shows counting OFF button in bet phase by default', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'カウント OFF' })).toBeInTheDocument());
  });

  it('shows counting ON button when countingEnabled is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'カウント ON' })).toBeInTheDocument());
  });

  it('calls togglecounting when counting button is clicked', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true });
    fireEvent.click(screen.getByRole('button', { name: 'カウント OFF' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('togglecounting'));
  });

  it('shows RC and TC with system name when countingEnabled is true (Hi-Lo)', async () => {
    mockExec.mockResolvedValue({
      ...betPhaseState,
      countingEnabled: true,
      runningCount: 5,
      trueCount: 2.3,
      countingSystem: 0,
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/Hi-Lo RC=5 TC=2\.3/)).toBeInTheDocument());
  });

  it('shows TC=N/A for KO system', async () => {
    mockExec.mockResolvedValue({
      ...betPhaseState,
      countingEnabled: true,
      runningCount: 3,
      trueCount: 0,
      countingSystem: 1,
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/KO RC=3 TC=N\/A/)).toBeInTheDocument());
  });

  it('does not show RC and TC when countingEnabled is false', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: false, runningCount: 5, trueCount: 2.3 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー: 1000 chips/)).toBeInTheDocument());
    expect(screen.queryByText(/RC=/)).not.toBeInTheDocument();
  });

  it('shows CPU count selector in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('CPU人数:')).toBeInTheDocument());
  });

  it('shows CPU player hands in action phase when cpuPlayers exist', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
      insuranceBet: 0,
      hands: [
        {
          score: 18,
          cards: [
            { design: 'HEART', value: 8 },
            { design: 'DIAMOND', value: 10 },
          ],
          bet: 100,
          stood: false,
          doubled: false,
          busted: false,
          isBlackJack: false,
          canSplit: false,
          surrendered: false,
          canSurrender: false,
        },
      ],
    };
    const stateWithCpu: BlackJackResponse = {
      ...actionPhaseState,
      cpuPlayerCount: 1,
      cpuPlayers: [cpuSeat],
    };
    mockExec.mockResolvedValue(stateWithCpu);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1 \(800 chips\)/)).toBeInTheDocument());
    expect(screen.getByText(/スコア 18/)).toBeInTheDocument();
  });

  it('shows CPU hand flags (busted, doubled, blackjack, surrendered)', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
      insuranceBet: 0,
      hands: [
        {
          score: 25,
          cards: [
            { design: 'HEART', value: 10 },
            { design: 'DIAMOND', value: 10 },
            { design: 'SPADE', value: 5 },
          ],
          bet: 100,
          stood: false,
          doubled: true,
          busted: true,
          isBlackJack: false,
          canSplit: false,
          surrendered: false,
          canSurrender: false,
        },
      ],
    };
    const stateWithCpu: BlackJackResponse = {
      ...actionPhaseState,
      cpuPlayerCount: 1,
      cpuPlayers: [cpuSeat],
    };
    mockExec.mockResolvedValue(stateWithCpu);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const bustElem = screen.getByTitle(i18n.t('blackjack:status.bustTooltip'));
      expect(bustElem).toBeInTheDocument();
      expect(bustElem).toHaveTextContent(`[${i18n.t('blackjack:status.bust')}]`);
      const ddElem = screen.getByTitle(i18n.t('blackjack:status.ddTooltip'));
      expect(ddElem).toBeInTheDocument();
      expect(ddElem).toHaveTextContent(`[${i18n.t('blackjack:status.dd')}]`);
    });
  });

  it('shows CPU BJ and SUR flags', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
      insuranceBet: 0,
      hands: [
        {
          score: 21,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 10 },
          ],
          bet: 100,
          stood: true,
          doubled: false,
          busted: false,
          isBlackJack: true,
          canSplit: false,
          surrendered: true,
          canSurrender: false,
        },
      ],
    };
    const stateWithCpu: BlackJackResponse = {
      ...actionPhaseState,
      cpuPlayerCount: 1,
      cpuPlayers: [cpuSeat],
    };
    mockExec.mockResolvedValue(stateWithCpu);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      const bjElem = screen.getByTitle(i18n.t('blackjack:status.bjTooltip'));
      expect(bjElem).toBeInTheDocument();
      expect(bjElem).toHaveTextContent(`[${i18n.t('blackjack:status.bj')}]`);
      const surElem = screen.getByTitle(i18n.t('blackjack:status.surTooltip'));
      expect(surElem).toBeInTheDocument();
      expect(surElem).toHaveTextContent(`[${i18n.t('blackjack:status.sur')}]`);
    });
  });

  it('shows hand labels when CPU has multiple hands', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
      insuranceBet: 0,
      hands: [
        {
          score: 15,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'DIAMOND', value: 10 },
          ],
          bet: 100,
          stood: false,
          doubled: false,
          busted: false,
          isBlackJack: false,
          canSplit: false,
          surrendered: false,
          canSurrender: false,
        },
        {
          score: 18,
          cards: [
            { design: 'CLOVER', value: 8 },
            { design: 'SPADE', value: 10 },
          ],
          bet: 100,
          stood: false,
          doubled: false,
          busted: false,
          isBlackJack: false,
          canSplit: false,
          surrendered: false,
          canSurrender: false,
        },
      ],
    };
    const stateWithCpu: BlackJackResponse = {
      ...actionPhaseState,
      cpuPlayerCount: 1,
      cpuPlayers: [cpuSeat],
    };
    mockExec.mockResolvedValue(stateWithCpu);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByText(/ハンド 1 スコア 15/)).toBeInTheDocument();
      expect(screen.getByText(/ハンド 2 スコア 18/)).toBeInTheDocument();
    });
  });

  it('shows (H17) in dealer heading when dealerHitsSoft17 is true', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, dealerHitsSoft17: true });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\(H17\)/)).toBeInTheDocument());
  });

  it('shows (S17) in dealer heading when dealerHitsSoft17 is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\(S17\)/)).toBeInTheDocument());
  });

  it('updates bet amount when input value is changed', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('ベット額:'), { target: { value: '50' } });
    // ChipBetInput now uses type=text + inputMode=numeric (#1615), so the
    // displayed value is a string rather than a number.
    expect(screen.getByLabelText('ベット額:')).toHaveValue('50');
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 50, undefined, {}));
  });

  it('calls setcpucount when CPU count selector value is changed', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('CPU人数:')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, cpuPlayerCount: 2 });
    fireEvent.change(screen.getByLabelText('CPU人数:'), { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('setcpucount', 2));
  });

  it('calls doubledown command when double down button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ダブルダウン' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubledown'));
  });

  it('calls split command when split button is clicked', async () => {
    const splitState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, canSplit: true }],
    };
    mockExec.mockResolvedValue(splitState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スプリット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('split'));
  });

  it('sends config params when reset button is confirmed in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        dealerHitsSoft17: false,
        cpuPlayerCount: 0,
        countingEnabled: false,
        doubleAfterSplit: true,
        countingSystem: 0,
        deckPenetration: 75,
        surrenderRule: 0,
      }),
    );
  });

  // --- DAS toggle tests ---

  it('shows DAS ON button in bet phase by default', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'DAS ON' })).toBeInTheDocument());
  });

  it('shows DAS OFF button when doubleAfterSplit is false', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, doubleAfterSplit: false });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'DAS OFF' })).toBeInTheDocument());
  });

  it('calls toggledas when DAS button is clicked', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, doubleAfterSplit: false });
    fireEvent.click(screen.getByRole('button', { name: 'DAS ON' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('toggledas'));
  });

  it('hides double down button after split when DAS is disabled', async () => {
    const splitHands: BlackJackHand[] = [
      {
        ...baseHand,
        score: 10,
        cards: [
          { design: 'SPADE', value: 8 },
          { design: 'HEART', value: 2 },
        ],
      },
      {
        ...baseHand,
        score: 12,
        cards: [
          { design: 'DIAMOND', value: 8 },
          { design: 'CLOVER', value: 4 },
        ],
      },
    ];
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hands: splitHands,
      doubleAfterSplit: false,
      player: { chips: 900 },
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ダブルダウン' })).not.toBeInTheDocument();
  });

  it('shows double down button after split when DAS is enabled', async () => {
    const splitHands: BlackJackHand[] = [
      {
        ...baseHand,
        score: 10,
        cards: [
          { design: 'SPADE', value: 8 },
          { design: 'HEART', value: 2 },
        ],
      },
      {
        ...baseHand,
        score: 12,
        cards: [
          { design: 'DIAMOND', value: 8 },
          { design: 'CLOVER', value: 4 },
        ],
      },
    ];
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hands: splitHands,
      doubleAfterSplit: true,
      player: { chips: 900 },
    });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeInTheDocument());
  });

  // --- Counting system tests ---

  it('shows counting system selector in bet phase', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('カウンティング方式')).toBeInTheDocument());
  });

  it('counting system selector is disabled when counting is off', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('カウンティング方式')).toBeDisabled());
  });

  it('calls setcountingsystem when counting system is changed', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true, countingSystem: 1 });
    fireEvent.change(screen.getByLabelText('カウンティング方式'), { target: { value: '1' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('setcountingsystem', 1));
  });

  // --- Deck penetration tests ---

  it('shows penetration selector in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ペネトレーション:')).toBeInTheDocument());
  });

  it('syncs deckPenetration from response', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, deckPenetration: 50 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ペネトレーション:')).toHaveValue('50'));
  });

  it('calls setpenetration when penetration selector is changed', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, deckPenetration: 50 });
    fireEvent.change(screen.getByLabelText('ペネトレーション:'), { target: { value: '50' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('setpenetration', 50));
  });

  // --- Side bet tests ---

  it('shows PP and 21+3 inputs in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('PP (ペアベット):')).toBeInTheDocument());
    expect(screen.getByLabelText('21+3:')).toBeInTheDocument();
  });

  it('sends side bets when PP and T3 are set', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.change(screen.getByLabelText('PP (ペアベット):'), { target: { value: '10' } });
    fireEvent.change(screen.getByLabelText('21+3:'), { target: { value: '20' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('bet', 10, undefined, {
        perfectPairsBet: 10,
        twentyOnePlus3Bet: 20,
      }),
    );
  });

  it('does not include side bets in body when they are zero', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 10, undefined, {}));
  });

  it('shows side bet results banner when sideBetResults exist (win)', async () => {
    const stateWithSideBets: BlackJackResponse = {
      ...actionPhaseState,
      sideBetResults: [{ betType: 1, resultType: 1, resultName: 'Perfect Pair', betAmount: 10, payout: 250 }],
    };
    mockExec.mockResolvedValue(stateWithSideBets);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/Perfect Pairs:.*Perfect Pair WIN \+250/)).toBeInTheDocument());
  });

  it('shows side bet results banner when sideBetResults exist (lose)', async () => {
    const stateWithSideBets: BlackJackResponse = {
      ...actionPhaseState,
      sideBetResults: [{ betType: 2, resultType: 0, resultName: '', betAmount: 20, payout: 0 }],
    };
    mockExec.mockResolvedValue(stateWithSideBets);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/21\+3:.*LOSE -20/)).toBeInTheDocument());
  });

  it('does not show side bet results when sideBetResults is empty', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByText(/Perfect Pairs:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/21\+3:/)).not.toBeInTheDocument();
  });

  // --- Multi-hand tests ---

  it('shows hand count selector in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ハンド数:')).toBeInTheDocument());
  });

  it('hand count selector defaults to 1', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ハンド数:')).toHaveValue('1'));
  });

  it('sends handCount when hand count is greater than 1', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.change(screen.getByLabelText('ハンド数:'), { target: { value: '2' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('bet', 10, undefined, {
        handCount: 2,
      }),
    );
  });

  it('does not include handCount when hand count is 1', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 10, undefined, {}));
  });

  // --- Auto-advance tests ---

  it('shows auto-advance selector in bet phase', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('自動進行:')).toBeInTheDocument());
  });

  it('auto-advance selector defaults to OFF', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('自動進行:')).toHaveValue('0'));
  });

  it('does not show countdown on reset button when auto-advance is OFF', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
  });

  it('shows CPU insurance bet when insuranceBet > 0', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
      insuranceBet: 50,
      hands: [
        {
          score: 18,
          cards: [
            { design: 'HEART', value: 8 },
            { design: 'DIAMOND', value: 10 },
          ],
          bet: 100,
          stood: false,
          doubled: false,
          busted: false,
          isBlackJack: false,
          canSplit: false,
          surrendered: false,
          canSurrender: false,
        },
      ],
    };
    const stateWithCpu: BlackJackResponse = {
      ...actionPhaseState,
      cpuPlayerCount: 1,
      cpuPlayers: [cpuSeat],
    };
    mockExec.mockResolvedValue(stateWithCpu);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1 \(800 chips\)/)).toBeInTheDocument());
    expect(screen.getByText('[インシュランス: 50]')).toBeInTheDocument();
  });

  it('does not show CPU insurance info when insuranceBet is 0', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
      insuranceBet: 0,
      hands: [
        {
          score: 18,
          cards: [
            { design: 'HEART', value: 8 },
            { design: 'DIAMOND', value: 10 },
          ],
          bet: 100,
          stood: false,
          doubled: false,
          busted: false,
          isBlackJack: false,
          canSplit: false,
          surrendered: false,
          canSurrender: false,
        },
      ],
    };
    const stateWithCpu: BlackJackResponse = {
      ...actionPhaseState,
      cpuPlayerCount: 1,
      cpuPlayers: [cpuSeat],
    };
    mockExec.mockResolvedValue(stateWithCpu);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1 \(800 chips\)/)).toBeInTheDocument());
    expect(screen.queryByText(/インシュランス/)).not.toBeInTheDocument();
  });

  // --- Early surrender tests ---

  it('renders early surrender phase controls when phase is 6', async () => {
    const earlySurrenderState: BlackJackResponse = {
      ...actionPhaseState,
      phase: 6,
    };
    mockExec.mockResolvedValue(earlySurrenderState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アーリーサレンダー' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '続行' })).toBeInTheDocument();
  });

  it('calls earlysurrender when surrender button clicked in early surrender phase', async () => {
    const earlySurrenderState: BlackJackResponse = {
      ...actionPhaseState,
      phase: 6,
    };
    mockExec.mockResolvedValue(earlySurrenderState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'アーリーサレンダー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('earlysurrender'));
  });

  it('calls declineearlysurrender when continue button clicked', async () => {
    const earlySurrenderState: BlackJackResponse = {
      ...actionPhaseState,
      phase: 6,
    };
    mockExec.mockResolvedValue(earlySurrenderState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '続行' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declineearlysurrender'));
  });

  it('shows active hand marker (*) during early surrender phase', async () => {
    const earlySurrenderState: BlackJackResponse = {
      ...actionPhaseState,
      phase: 6,
    };
    mockExec.mockResolvedValue(earlySurrenderState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー手札 \(\*\)/)).toBeInTheDocument());
  });

  it('syncs surrenderRule from response', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, surrenderRule: 1 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('サレンダー:')).toHaveValue('1'));
  });

  it('calls setsurrenderrule when surrender rule selector changes', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('サレンダー:')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('サレンダー:'), { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('setsurrenderrule', 2));
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue({
      gameEndFlag: true,
      phase: 5, // BjPhase.END
      players: [],
      playerIdx: 0,
      player: { chips: 100 },
      dealer: { chips: 100 },
      currentHandIdx: 0,
    } as unknown as BlackJackResponse);

    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.blackjack).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.blackjack).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();

    fireEvent.click(screen.getByText('閉じる'));
    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  // --- Next Game button / reset confirmation tests ---

  it('shows the reset confirmation dialog when the next game button is clicked', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    // Confirmation dialog appears and no reset fires until it is confirmed.
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('executes reset after confirming the dialog in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
  });

  it('does not reset when the confirmation dialog is cancelled', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    // Dialog closes and no reset command is sent.
    await waitFor(() => expect(screen.queryByText('本当にゲームをリセットしますか？')).not.toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });

  // --- Keyboard navigation tests ---

  it('pressing h triggers hit in ACTION phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hit'));
  });

  it('pressing s triggers stand in ACTION phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.keyDown(document, { key: 's' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('pressing d triggers doubledown when double is available', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('doubledown'));
  });

  it('pressing d does not trigger doubledown when double is unavailable', async () => {
    const noDoubleState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [
        {
          ...baseHand,
          cards: [
            { design: 'HEART', value: 5 },
            { design: 'DIAMOND', value: 10 },
            { design: 'SPADE', value: 3 },
          ],
        },
      ],
    };
    mockExec.mockResolvedValue(noDoubleState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    expect(mockExec).not.toHaveBeenCalledWith('doubledown');
  });

  it('keyboard shortcuts are disabled when not in ACTION phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'h' });
    fireEvent.keyDown(document, { key: 's' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn during action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows your turn during early surrender phase', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, phase: 6 });
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting during end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('phase indicator shows your turn during insurance phase', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByText('あなたのターン')).toBeInTheDocument();
    expect(screen.queryByText('待機中')).not.toBeInTheDocument();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  // --- Payout table collapsible tests ---

  it('payout table is rendered as a collapsible details element in bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    const { container } = renderWithProviders(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('配当表')).toBeInTheDocument());
    const details = container.querySelector('details');
    expect(details).toBeInTheDocument();
    const summary = details?.querySelector('summary');
    expect(summary).toHaveTextContent('配当表');
  });
});
