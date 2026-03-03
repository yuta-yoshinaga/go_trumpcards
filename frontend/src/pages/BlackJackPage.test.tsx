import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackjackApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import type { BlackJackCpuSeat, BlackJackHand, BlackJackResponse } from '../types/card';
import { BlackJackPage } from './BlackJackPage';

vi.mock('../api/gameApi', () => ({
  blackjackApi: { exec: vi.fn() },
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
};

beforeEach(() => {
  mockExec.mockResolvedValue(betPhaseState);
});

describe('BlackJackPage', () => {
  it('calls reset command on mount', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows chip info bar', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー: 1000 chips/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー: 1000 chips/)).toBeInTheDocument();
  });

  it('shows bet button in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
  });

  it('shows bet amount input in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());
  });

  it('calls bet command when bet button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 10, undefined, {}));
  });

  it('shows hit and stand buttons in action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('shows double down button when 2 cards and sufficient chips', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeInTheDocument());
  });

  it('shows split button when canSplit and sufficient chips', async () => {
    const splitState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, canSplit: true }],
    };
    mockExec.mockResolvedValue(splitState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スプリット' })).toBeInTheDocument());
  });

  it('does not show split button when canSplit is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  it('shows insurance buttons in insurance phase', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'インシュランス' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '辞退' })).toBeInTheDocument();
  });

  it('calls insurance command when insurance button is clicked', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'インシュランス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('insurance'));
  });

  it('calls declineinsurance command when decline button is clicked', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '辞退' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declineinsurance'));
  });

  it('shows reset button in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
  });

  it('shows message overlay when message is non-empty', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('You are the winner.')).toBeInTheDocument());
  });

  it('does not show message overlay when message is empty', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('ヒット')).toBeInTheDocument());
    expect(screen.queryByText('You are the winner.')).not.toBeInTheDocument();
  });

  it('calls hit command when Hit button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hit'));
  });

  it('calls stand command when Stand button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('displays player score and bet', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 15/)).toBeInTheDocument());
    expect(screen.getByText(/ベット 100/)).toBeInTheDocument();
  });

  it('displays dealer score when non-zero in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 19/)).toBeInTheDocument());
  });

  it('shows card back when dealer score is zero', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => {
      const imgs = screen.getAllByRole('img');
      const cardBackImg = imgs.find((img) => img.getAttribute('src') === '/images/z01.png');
      expect(cardBackImg).toBeInTheDocument();
    });
  });

  it('shows BJ flag for blackjack hand', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\[BJ\]/)).toBeInTheDocument());
  });

  it('does not show dealer area in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.queryByText('ディーラー手札')).not.toBeInTheDocument();
  });

  it('disables bet button while loading', async () => {
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE)).toBeInTheDocument());
  });

  it('clears error message on successful API call after failure', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE)).toBeInTheDocument());

    mockExec.mockResolvedValueOnce(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.queryByText(NETWORK_ERROR_MESSAGE)).not.toBeInTheDocument());
  });

  it('shows [BUST] flag for busted hand', async () => {
    const bustState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, busted: true }],
    };
    mockExec.mockResolvedValue(bustState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\[BUST\]/)).toBeInTheDocument());
  });

  it('shows [DD] flag for doubled hand', async () => {
    const ddState: BlackJackResponse = {
      ...actionPhaseState,
      hands: [{ ...baseHand, doubled: true }],
    };
    mockExec.mockResolvedValue(ddState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\[DD\]/)).toBeInTheDocument());
  });

  it('shows insurance bet info when insuranceBet > 0', async () => {
    const insState: BlackJackResponse = { ...actionPhaseState, insuranceBet: 50 };
    mockExec.mockResolvedValue(insState);
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  // --- New feature tests ---

  it('shows deck count selector in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('デッキ数:')).toBeInTheDocument());
  });

  it('calls setdeckcount when deck count is changed', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.change(screen.getByLabelText('デッキ数:'), { target: { value: '6' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('setdeckcount', 6));
  });

  it('shows hint toggle button with OFF state in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント OFF' })).toBeInTheDocument());
  });

  it('calls togglehint when hint toggle button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント OFF' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('togglehint'));
  });

  it('shows hint button as ON when hintEnabled is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, hintEnabled: true });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント ON' })).toBeInTheDocument());
  });

  it('shows surrender button when canSurrender is true', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hands: [{ ...baseHand, canSurrender: true }],
    });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'サレンダー' })).toBeInTheDocument());
  });

  it('does not show surrender button when canSurrender is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'サレンダー' })).not.toBeInTheDocument();
  });

  it('calls surrender when surrender button is clicked', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hands: [{ ...baseHand, canSurrender: true }],
    });
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('推奨: ヒット')).toBeInTheDocument());
  });

  it('does not show hint banner when hintEnabled is false', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: false, suggestedAction: 1 });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByText(/推奨:/)).not.toBeInTheDocument();
  });

  it('does not show hint banner when suggestedAction is none', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 0 });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByText(/推奨:/)).not.toBeInTheDocument();
  });

  it('highlights hit button when hint suggests hit', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 1 });
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ヒット' })).toHaveClass('ring-2');
    });
  });

  it('highlights stand button when hint suggests stand', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 2 });
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'スタンド' })).toHaveClass('ring-2');
    });
  });

  it('highlights double down button when hint suggests double', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 3 });
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ダブルダウン' })).toHaveClass('ring-2');
    });
  });

  it('highlights double down button when hint suggests doubleStand', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 7 });
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ダブルダウン' })).toHaveClass('ring-2');
    });
  });

  it('shows hint banner for doubleStand suggestion', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, hintEnabled: true, suggestedAction: 7 });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('推奨: ダブルダウン')).toBeInTheDocument());
  });

  it('highlights split button when hint suggests split', async () => {
    mockExec.mockResolvedValue({
      ...actionPhaseState,
      hintEnabled: true,
      suggestedAction: 4,
      hands: [{ ...baseHand, canSplit: true }],
    });
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'サレンダー' })).toHaveClass('ring-2');
    });
  });

  it('highlights decline button when hint suggests decline insurance', async () => {
    mockExec.mockResolvedValue({ ...insurancePhaseState, hintEnabled: true, suggestedAction: 6 });
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '辞退' })).toHaveClass('ring-2');
    });
  });

  it('shows SURRENDER badge on surrendered hand', async () => {
    const surrenderedEndState: BlackJackResponse = {
      ...endPhaseState,
      hands: [{ ...(endPhaseState.hands?.[0] as BlackJackHand), surrendered: true }],
    };
    mockExec.mockResolvedValue(surrenderedEndState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('SURRENDER')).toBeInTheDocument());
  });

  it('shows deck count in chip bar', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, deckCount: 6 });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('デッキ: 6デッキ')).toBeInTheDocument());
  });

  it('sets aria-busy and sr-only loading text while loading', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-live]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
    expect(screen.queryByText('処理中...')).not.toBeInTheDocument();

    let resolve!: (value: BlackJackResponse) => void;
    const slowPromise = new Promise<BlackJackResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('処理中...')).toBeInTheDocument();

    resolve(betPhaseState);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
      expect(screen.queryByText('処理中...')).not.toBeInTheDocument();
    });
  });

  // --- S17/H17 and counting toggle tests ---

  it('shows S17 button in bet phase by default', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'S17' })).toBeInTheDocument());
  });

  it('shows H17 button when dealerHitsSoft17 is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, dealerHitsSoft17: true });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'H17' })).toBeInTheDocument());
  });

  it('calls togglesoft17 when S17 button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, dealerHitsSoft17: true });
    fireEvent.click(screen.getByRole('button', { name: 'S17' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('togglesoft17'));
  });

  it('shows counting OFF button in bet phase by default', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'カウント OFF' })).toBeInTheDocument());
  });

  it('shows counting ON button when countingEnabled is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'カウント ON' })).toBeInTheDocument());
  });

  it('calls togglecounting when counting button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true });
    fireEvent.click(screen.getByRole('button', { name: 'カウント OFF' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('togglecounting'));
  });

  it('shows RC and TC when countingEnabled is true', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: true, runningCount: 5, trueCount: 2.3 });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('RC=5 TC=2.3')).toBeInTheDocument());
  });

  it('does not show RC and TC when countingEnabled is false', async () => {
    mockExec.mockResolvedValue({ ...betPhaseState, countingEnabled: false, runningCount: 5, trueCount: 2.3 });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー: 1000 chips/)).toBeInTheDocument());
    expect(screen.queryByText(/RC=/)).not.toBeInTheDocument();
  });

  it('shows CPU count selector in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('CPU人数:')).toBeInTheDocument());
  });

  it('shows CPU player hands in action phase when cpuPlayers exist', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
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
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/CPU 1 \(800 chips\)/)).toBeInTheDocument());
    expect(screen.getByText(/スコア 18/)).toBeInTheDocument();
  });

  it('shows CPU hand flags (busted, doubled, blackjack, surrendered)', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
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
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByText(/\[BUST\]/)).toBeInTheDocument();
      expect(screen.getByText(/\[DD\]/)).toBeInTheDocument();
    });
  });

  it('shows CPU BJ and SUR flags', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
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
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByText(/\[BJ\]/)).toBeInTheDocument();
      expect(screen.getByText(/\[SUR\]/)).toBeInTheDocument();
    });
  });

  it('shows hand labels when CPU has multiple hands', async () => {
    const cpuSeat: BlackJackCpuSeat = {
      chips: 800,
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
    render(<BlackJackPage />);
    await waitFor(() => {
      expect(screen.getByText(/ハンド 1 スコア 15/)).toBeInTheDocument();
      expect(screen.getByText(/ハンド 2 スコア 18/)).toBeInTheDocument();
    });
  });

  it('shows (H17) in dealer heading when dealerHitsSoft17 is true', async () => {
    mockExec.mockResolvedValue({ ...actionPhaseState, dealerHitsSoft17: true });
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\(H17\)/)).toBeInTheDocument());
  });

  it('shows (S17) in dealer heading when dealerHitsSoft17 is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\(S17\)/)).toBeInTheDocument());
  });

  it('updates bet amount when input value is changed', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('ベット額:'), { target: { value: '50' } });
    expect(screen.getByLabelText('ベット額:')).toHaveValue(50);
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 50, undefined, {}));
  });

  it('updates CPU count when selector value is changed', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('CPU人数:')).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('CPU人数:'), { target: { value: '2' } });
    expect(screen.getByLabelText('CPU人数:')).toHaveValue('2');
  });

  it('calls doubledown command when double down button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スプリット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('split'));
  });

  it('sends config params when reset button is clicked in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        dealerHitsSoft17: false,
        cpuPlayerCount: 0,
        countingEnabled: false,
      }),
    );
  });

  // --- Side bet tests ---

  it('shows PP and 21+3 inputs in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('PP:')).toBeInTheDocument());
    expect(screen.getByLabelText('21+3:')).toBeInTheDocument();
  });

  it('sends side bets when PP and T3 are set', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.change(screen.getByLabelText('PP:'), { target: { value: '10' } });
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
    render(<BlackJackPage />);
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
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/Perfect Pairs:.*Perfect Pair WIN \+250/)).toBeInTheDocument());
  });

  it('shows side bet results banner when sideBetResults exist (lose)', async () => {
    const stateWithSideBets: BlackJackResponse = {
      ...actionPhaseState,
      sideBetResults: [{ betType: 2, resultType: 0, resultName: '', betAmount: 20, payout: 0 }],
    };
    mockExec.mockResolvedValue(stateWithSideBets);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/21\+3:.*LOSE -20/)).toBeInTheDocument());
  });

  it('does not show side bet results when sideBetResults is empty', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByText(/Perfect Pairs:/)).not.toBeInTheDocument();
    expect(screen.queryByText(/21\+3:/)).not.toBeInTheDocument();
  });

  // --- Auto-advance tests ---

  it('shows auto-advance selector in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('自動進行:')).toBeInTheDocument());
  });

  it('auto-advance selector defaults to OFF', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('自動進行:')).toHaveValue('0'));
  });

  it('does not show countdown on reset button when auto-advance is OFF', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
  });
});
