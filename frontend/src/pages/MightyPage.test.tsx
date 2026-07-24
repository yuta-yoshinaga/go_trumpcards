import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, mightyApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MightyResponse } from '../types/card';
import { MightyPage } from './MightyPage';

vi.mock('../api/gameApi', () => ({
  mightyApi: { exec: vi.fn() },
  actionLogApi: { mighty: vi.fn() },
}));

const mockCall = vi.mocked(mightyApi.exec);

const playPhaseState: MightyResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 14,
      bidNoTrump: false,
      isDeclarer: true,
      isPartner: false,
      partnerRevealed: false,
      pointCards: 2,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      bidNoTrump: false,
      isDeclarer: false,
      isPartner: true,
      partnerRevealed: false,
      pointCards: 1,
      roundScore: 3,
      cumulativeScore: 10,
      trickCount: 1,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      bidNoTrump: false,
      isDeclarer: false,
      isPartner: false,
      partnerRevealed: false,
      pointCards: 0,
      roundScore: 5,
      cumulativeScore: 20,
      trickCount: 2,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      bidNoTrump: false,
      isDeclarer: false,
      isPartner: false,
      partnerRevealed: false,
      pointCards: 0,
      roundScore: 0,
      cumulativeScore: 5,
      trickCount: 0,
    },
    {
      id: 4,
      isHuman: false,
      cardCount: 10,
      cards: [],
      bid: 0,
      bidNoTrump: false,
      isDeclarer: false,
      isPartner: false,
      partnerRevealed: false,
      pointCards: 0,
      roundScore: 1,
      cumulativeScore: 8,
      trickCount: 0,
    },
  ],
  phase: 3,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  trumpSuit: 1,
  partnerCard: { design: 'HEART', value: 1 },
  declarerIdx: 0,
  partnerIdx: 1,
  partnerRevealed: false,
  highestBid: 14,
  highestBidder: 0,
  winningBidNoTrump: false,
  kitty: [],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, minBid: 13, noTrumpExtra: 2, pointLimit: 100 },
};

const bidPhaseState: MightyResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  trumpSuit: 0,
  partnerCard: null,
  highestBid: 0,
  declarerIdx: -1,
  partnerIdx: -1,
  players: playPhaseState.players.map((p) => ({
    ...p,
    bid: -1,
    isDeclarer: false,
    isPartner: false,
  })),
};

const bidPhaseCpuTurnState: MightyResponse = {
  ...bidPhaseState,
  bidPlayerIdx: 1,
};

const trumpAndFriendState: MightyResponse = {
  ...playPhaseState,
  phase: 1,
  declarerIdx: 0,
};

const trumpAndFriendCpuState: MightyResponse = {
  ...playPhaseState,
  phase: 1,
  declarerIdx: 1,
  players: playPhaseState.players.map((p, i) => ({
    ...p,
    isDeclarer: i === 1,
  })),
};

const kittyExchangeState: MightyResponse = {
  ...playPhaseState,
  phase: 2,
  declarerIdx: 0,
  kitty: [
    { design: 'DIAMOND', value: 7 },
    { design: 'CLOVER', value: 3 },
    { design: 'HEART', value: 9 },
  ],
};

const kittyExchangeCpuState: MightyResponse = {
  ...playPhaseState,
  phase: 2,
  declarerIdx: 1,
  kitty: [],
  players: playPhaseState.players.map((p, i) => ({
    ...p,
    isDeclarer: i === 1,
  })),
};

const trickEndState: MightyResponse = {
  ...playPhaseState,
  phase: 4,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: MightyResponse = {
  ...playPhaseState,
  phase: 5,
};

const gameEndState: MightyResponse = {
  ...playPhaseState,
  phase: 6,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const cpuTurnState: MightyResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const partnerRevealedState: MightyResponse = {
  ...playPhaseState,
  partnerRevealed: true,
};

const noTrumpWinState: MightyResponse = {
  ...playPhaseState,
  trumpSuit: 0,
  winningBidNoTrump: true,
};

const humanWithJokerState: MightyResponse = {
  ...playPhaseState,
  players: playPhaseState.players.map((p, i) =>
    i === 0
      ? {
          ...p,
          cards: [
            { design: 'JOKER' as const, value: 0 },
            { design: 'HEART' as const, value: 5 },
          ],
        }
      : p,
  ),
};

beforeEach(() => {
  mockCall.mockResolvedValue(playPhaseState);
});

describe('MightyPage', () => {
  it('renders skeleton when no state', () => {
    mockCall.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MightyPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() =>
      expect(mockCall).toHaveBeenCalledWith(
        'reset',
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        { cpuDifficulty: 1, minBid: 13, noTrumpExtra: 2, pointLimit: 100 },
      ),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('badges Mighty and partner cards in the human hand', async () => {
    // With trumpSuit=3 (HEART) the Mighty card is ♠A.
    const heartTrump: MightyResponse = {
      ...playPhaseState,
      trumpSuit: 3,
      partnerCard: { design: 'HEART', value: 11 },
      players: playPhaseState.players.map((p, i) =>
        i === 0
          ? {
              ...p,
              cards: [
                { design: 'SPADE', value: 1 },
                { design: 'HEART', value: 11 },
                { design: 'DIAMOND', value: 8 },
              ],
            }
          : p,
      ),
    };
    mockCall.mockResolvedValue(heartTrump);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('card-role-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('card-role-badge-1')).toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument();
  });

  it('renders bid phase with bid button and a discrete bid grid', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument();
      expect(screen.getByTestId('mighty-bid-grid')).toBeInTheDocument();
    });
    // minBid (13) through MightyMaxPoints (20) are rendered as buttons.
    for (let n = 13; n <= 20; n++) {
      expect(screen.getByTestId(`bid-option-${n}`)).toBeInTheDocument();
    }
    expect(screen.queryByTestId('bid-option-12')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-21')).not.toBeInTheDocument();
  });

  it('marks the selected bid button with aria-pressed', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('bid-option-15')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('bid-option-15'));
    expect(screen.getByTestId('bid-option-15')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('bid-option-16')).toHaveAttribute('aria-pressed', 'false');
  });

  it('disables bids at or below the current highest bid', async () => {
    mockCall.mockResolvedValue({ ...bidPhaseState, highestBid: 15 });
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('bid-option-16')).toBeInTheDocument());

    expect(screen.getByTestId('bid-option-15')).toBeDisabled();
    expect(screen.getByTestId('bid-option-16')).toBeEnabled();
  });

  it('disables bids below the raised minimum when no-trump is toggled on', async () => {
    // minBid 13 + noTrumpExtra 2 = 15 effective minimum under no-trump.
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByLabelText('bid-no-trump')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('bid-no-trump'));
    expect(screen.getByTestId('bid-option-14')).toBeDisabled();
    expect(screen.getByTestId('bid-option-15')).toBeEnabled();
  });

  it('shows bid instruction during human bid turn', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText(/ビッドを宣言してください/)).toBeInTheDocument();
    });
  });

  it('shows no-trump extra explanation during human bid turn', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByTestId('mighty-notrump-explain')).toHaveTextContent(/ノートランプ加算/);
    });
  });

  it('does not show no-trump extra explanation when cpu bid turn', async () => {
    mockCall.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('mighty-notrump-explain')).not.toBeInTheDocument();
  });

  it('does not show bid instruction when cpu bid turn', async () => {
    mockCall.mockResolvedValue(bidPhaseCpuTurnState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText(/ビッドを宣言/)).not.toBeInTheDocument();
  });

  it('shows pass button during human bid turn', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
    });
  });

  it('calls bid command with the selected grid value when bid button clicked', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('bid-option-15')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('bid-option-15'));

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('bid', 15, false));
  });

  it('calls bid with noTrump=true when toggle is on', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByLabelText('bid-no-trump')).toBeInTheDocument());

    fireEvent.click(screen.getByLabelText('bid-no-trump'));
    fireEvent.click(screen.getByTestId('bid-option-15'));

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('bid', 15, true));
  });

  it('calls pass (bid 0 noTrump=false) when pass button clicked', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('bid', 0, false));
  });

  it('shows trump-and-friend declaration controls when human is declarer', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByTestId('trump-suit--1')).toBeInTheDocument();
      expect(screen.getByTestId('trump-suit-1')).toBeInTheDocument();
      expect(screen.getByTestId('partner-suit-0')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument();
    });
  });

  it('highlights the selected trump suit button', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    const heart = await screen.findByTestId('trump-suit-3');
    fireEvent.click(heart);
    expect(screen.getByTestId('trump-suit-3')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('trump-suit-1')).toHaveAttribute('aria-pressed', 'false');
  });

  it('highlights the selected partner suit button', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    const club = await screen.findByTestId('partner-suit-2');
    fireEvent.click(club);
    expect(screen.getByTestId('partner-suit-2')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('partner-suit-0')).toHaveAttribute('aria-pressed', 'false');
  });

  it('does not show declaration controls when cpu is declarer', async () => {
    mockCall.mockResolvedValue(trumpAndFriendCpuState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByTestId('trump-suit--1')).not.toBeInTheDocument();
  });

  it('calls trump command when declare button clicked', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言' })).toBeInTheDocument());

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('trump', undefined, undefined, undefined, 1, 1, 1));
  });

  it('hides partner-value when partner-suit is joker (0)', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('partner-suit-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('partner-suit-0'));
    expect(screen.queryByLabelText('partner-value')).not.toBeInTheDocument();
  });

  it('sends partner-value 0 when joker partner is chosen', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('partner-suit-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('partner-suit-0'));

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('trump', undefined, undefined, undefined, 1, 0, 0));
  });

  it('supports No-Trump trump declaration (-1)', async () => {
    mockCall.mockResolvedValue(trumpAndFriendState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('trump-suit--1')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('trump-suit--1'));

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('trump', undefined, undefined, undefined, -1, 1, 1));
  });

  it('shows kitty exchange controls when human declarer', async () => {
    mockCall.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText('場札')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument();
    });
  });

  it('does not show kitty exchange controls when cpu declarer', async () => {
    mockCall.mockResolvedValue(kittyExchangeCpuState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '交換' })).not.toBeInTheDocument();
  });

  it('exchange button disabled when not exactly 3 cards selected', async () => {
    mockCall.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeDisabled());
  });

  it('play button disabled when no card selected', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('calls play command when play button clicked', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('play', undefined, undefined, 0));
  });

  it('does not show play button when not human turn', async () => {
    mockCall.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows joker-lead button when joker selected while leading', async () => {
    mockCall.mockResolvedValue(humanWithJokerState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('ジョーカー').closest('button') as HTMLButtonElement);
    expect(screen.getByLabelText('joker-lead-button')).toBeInTheDocument();
  });

  it('opens joker suit picker on click and dispatches jokerlead', async () => {
    mockCall.mockResolvedValue(humanWithJokerState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByAltText('ジョーカー')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('ジョーカー').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByLabelText('joker-lead-button'));
    expect(screen.getByRole('dialog', { name: 'joker-suit-picker' })).toBeInTheDocument();

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));

    await waitFor(() =>
      expect(mockCall).toHaveBeenCalledWith(
        'jokerlead',
        undefined,
        undefined,
        0,
        undefined,
        undefined,
        undefined,
        undefined,
        3,
      ),
    );
  });

  it('shows next trick button on trick end', async () => {
    mockCall.mockResolvedValue(trickEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('calls next when next trick button clicked', async () => {
    mockCall.mockResolvedValue(trickEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('next'));
  });

  it('shows next round button on round end', async () => {
    mockCall.mockResolvedValue(roundEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('calls nextround when next round button clicked', async () => {
    mockCall.mockResolvedValue(roundEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));

    await waitFor(() => expect(mockCall).toHaveBeenCalledWith('nextround'));
  });

  it('shows game end with action log button', async () => {
    mockCall.mockResolvedValue(gameEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('shows error alert on retry', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockCall.mockReset();
    mockCall.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockCall).toHaveBeenCalledWith(
        'reset',
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        { cpuDifficulty: 2, pointLimit: 100, minBid: 13, noTrumpExtra: 2 },
      ),
    );
  });

  it('reset button calls api', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockCall.mockClear();
    mockCall.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockCall).toHaveBeenCalledWith(
        'reset',
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        { cpuDifficulty: 1, pointLimit: 100, minBid: 13, noTrumpExtra: 2 },
      ),
    );
  });

  it('card aria-pressed toggles on click', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('shows trump suit info during play phase', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText(/スペード/)).toBeInTheDocument());
  });

  it('shows No-Trump in trump label when winning bid is no-trump', async () => {
    mockCall.mockResolvedValue(noTrumpWinState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getAllByText(/ノートランプ/).length).toBeGreaterThan(0));
  });

  it('does not show trump info when trumpSuit is 0 and not no-trump', async () => {
    mockCall.mockResolvedValue(bidPhaseState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText('切り札:')).not.toBeInTheDocument();
  });

  it('shows highest bid info', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('最高ビッド: 14')).toBeInTheDocument());
  });

  it('shows current trick cards in trick end state', async () => {
    mockCall.mockResolvedValue(trickEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText('現在のトリック')).toBeInTheDocument();
      expect(screen.getByAltText('♦ 3')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 5')).toBeInTheDocument();
    });
  });

  it('shows CPU player areas with 4 opponents', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*10枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*10枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*10枚/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 4.*10枚/)).toBeInTheDocument();
    });
  });

  it('shows declarer role badge', async () => {
    renderWithProviders(<MightyPage />);
    // human is declarer
    await waitFor(() => expect(screen.getAllByText(/宣言者/).length).toBeGreaterThan(0));
  });

  it('shows partner role badge when revealed', async () => {
    mockCall.mockResolvedValue(partnerRevealedState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*パートナー/)).toBeInTheDocument();
    });
  });

  it('does not show partner role badge when not revealed', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByText(/CPU 1.*パートナー/)).not.toBeInTheDocument();
  });

  it('shows kitty cards during exchange phase', async () => {
    mockCall.mockResolvedValue(kittyExchangeState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText('場札')).toBeInTheDocument();
      expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
      expect(screen.getByAltText('♣ 3')).toBeInTheDocument();
    });
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => {
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
      expect(screen.getByText('トリック 1')).toBeInTheDocument();
    });
  });

  it('shows confirm dialog when reset clicked', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('handles action log open/close', async () => {
    mockCall.mockResolvedValue(gameEndState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.mighty).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));

    await waitFor(() => expect(actionLogApi.mighty).toHaveBeenCalledTimes(1));
    expect(screen.getByText('棋譜')).toBeInTheDocument();
  });

  it('phase indicator shows your turn during human play', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockCall.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<MightyPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
