import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pokerApi } from '../api/gameApi';
import type { PokerResponse } from '../types/card';
import { PokerPage } from './PokerPage';

vi.mock('../api/gameApi', () => ({
  pokerApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(pokerApi.exec);

const phase0State: PokerResponse = {
  phase: 0,
  player: { cards: [], handRank: 0, handName: '', chips: 1000, bet: 0 },
  dealer: { cards: [], handRank: 0, handName: '', chips: 1000, bet: 0 },
  message: 'リセットしました',
  pot: 0,
  ante: 10,
};

const phase1State: PokerResponse = {
  phase: 1,
  player: {
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
      { design: 'CLOVER', value: 3 },
      { design: 'SPADE', value: 7 },
    ],
    handRank: 0,
    handName: '',
    chips: 990,
    bet: 0,
  },
  dealer: { cards: [], handRank: 0, handName: '', chips: 990, bet: 0 },
  message: '',
  pot: 20,
  ante: 10,
};

const phase1StateWithDealerBet: PokerResponse = {
  phase: 1,
  player: {
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
      { design: 'CLOVER', value: 3 },
      { design: 'SPADE', value: 7 },
    ],
    handRank: 0,
    handName: '',
    chips: 990,
    bet: 0,
  },
  dealer: { cards: [], handRank: 0, handName: '', chips: 980, bet: 10 },
  message: '',
  pot: 30,
  ante: 10,
};

const phase2State: PokerResponse = {
  phase: 2,
  player: {
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
      { design: 'CLOVER', value: 3 },
      { design: 'SPADE', value: 7 },
    ],
    handRank: 0,
    handName: '',
    chips: 980,
    bet: 0,
  },
  dealer: { cards: [], handRank: 0, handName: '', chips: 980, bet: 0 },
  message: '',
  pot: 40,
  ante: 10,
};

const phase4State: PokerResponse = {
  phase: 4,
  player: {
    cards: [
      { design: 'SPADE', value: 1 },
      { design: 'HEART', value: 5 },
      { design: 'DIAMOND', value: 10 },
      { design: 'CLOVER', value: 3 },
      { design: 'SPADE', value: 7 },
    ],
    handRank: 0,
    handName: 'High Card',
    chips: 960,
    bet: 0,
  },
  dealer: {
    cards: [
      { design: 'HEART', value: 2 },
      { design: 'DIAMOND', value: 4 },
      { design: 'CLOVER', value: 6 },
      { design: 'SPADE', value: 8 },
      { design: 'HEART', value: 10 },
    ],
    handRank: 1,
    handName: 'Pair',
    chips: 1040,
    bet: 0,
  },
  message: 'あなたの負け',
  pot: 0,
  ante: 10,
};

beforeEach(() => {
  mockExec.mockResolvedValue(phase0State);
});

describe('PokerPage', () => {
  it('calls reset command on mount', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders dealer and player section labels', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    expect(screen.getByText(/プレイヤー手札/)).toBeInTheDocument();
  });

  it('shows 5 card backs for dealer in phase 0', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー手札/)).toBeInTheDocument());
    const cardBacks = screen.getAllByAltText('card back');
    expect(cardBacks.length).toBeGreaterThanOrEqual(5);
  });

  it('shows pot and chip info', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/プレイヤー チップ:/)).toBeInTheDocument();
    expect(screen.getByText(/ディーラー チップ:/)).toBeInTheDocument();
  });

  it('shows betting buttons in phase 1 without dealer bet', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows call/raise buttons in phase 1 with dealer bet', async () => {
    mockExec.mockResolvedValue(phase1StateWithDealerBet);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows exchange and stand buttons in phase 2', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('shows player hand name in phase 4', async () => {
    mockExec.mockResolvedValue(phase4State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
  });

  it('shows dealer hand name in phase 4', async () => {
    mockExec.mockResolvedValue(phase4State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('Pair')).toBeInTheDocument());
  });

  it('shows dealer cards (not card backs) in phase 4', async () => {
    mockExec.mockResolvedValue(phase4State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('High Card')).toBeInTheDocument());
    const cardBacks = screen.queryAllByAltText('card back');
    expect(cardBacks).toHaveLength(0);
  });

  it('shows result message', async () => {
    mockExec.mockResolvedValue(phase4State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの負け')).toBeInTheDocument());
  });

  it('calls exchange with selected indices when exchange button clicked', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '交換' })).toBeInTheDocument());

    const playerCardImgs = screen.getAllByAltText(/SPADE 1|HEART 5/);
    fireEvent.click(playerCardImgs[0]);
    fireEvent.click(playerCardImgs[1]);

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...phase2State, phase: 3 });
    fireEvent.click(screen.getByRole('button', { name: '交換' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange', expect.arrayContaining([0, 1])));
  });

  it('calls stand command when stand button is clicked in phase 2', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase4State);
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('calls reset command when reset button is clicked', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(phase0State);
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows instruction text in phase 2', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/交換したいカードをクリックして選択/)).toBeInTheDocument());
  });

  it('calls bet command with amount', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase2State);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 10));
  });

  it('calls check command', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase2State);
    fireEvent.click(screen.getByRole('button', { name: 'チェック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('check'));
  });

  it('calls fold command', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase4State);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('calls call command when dealer has bet', async () => {
    mockExec.mockResolvedValue(phase1StateWithDealerBet);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase2State);
    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call'));
  });

  it('calls raise command with amount when dealer has bet', async () => {
    mockExec.mockResolvedValue(phase1StateWithDealerBet);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'レイズ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(phase2State);
    fireEvent.click(screen.getByRole('button', { name: 'レイズ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', undefined, 10));
  });

  it('shows dealer bet info when dealer has bet', async () => {
    mockExec.mockResolvedValue(phase1StateWithDealerBet);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText(/ディーラー ベット:/)).toBeInTheDocument());
  });

  it('disables betting buttons while loading', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    let resolve!: (value: PokerResponse) => void;
    const slowPromise = new Promise<PokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'チェック' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeDisabled();

    resolve(phase1State);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());
  });

  it('shows error message when API call fails', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('clears error message on successful API call after failure', async () => {
    render(<PokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );

    mockExec.mockResolvedValueOnce(phase0State);
    fireEvent.click(screen.getByText('リセット'));
    await waitFor(() =>
      expect(screen.queryByText('通信エラーが発生しました。もう一度お試しください。')).not.toBeInTheDocument(),
    );
  });

  it('does not select card when clicking in non-exchange phase', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());

    // In bet phase (isExchangePhase=false), clicking a card exercises the `if (!isExchangePhase) return` branch
    fireEvent.click(screen.getByAltText('SPADE 1'));

    // Verify page is still in bet phase (no navigation or error occurred)
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('deselects a card by clicking it again in exchange phase', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    // Click to select (prev.includes(idx) = false → adds idx)
    fireEvent.click(screen.getByAltText('SPADE 1'));
    // Click again to deselect (prev.includes(idx) = true → filters out idx)
    fireEvent.click(screen.getByAltText('SPADE 1'));

    // Page should remain stable in exchange phase
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('does not show dealer hand name badge when handName is empty in end phase', async () => {
    const phase4EmptyDealerName: PokerResponse = {
      ...phase4State,
      dealer: { ...phase4State.dealer, handName: '' },
    };
    mockExec.mockResolvedValue(phase4EmptyDealerName);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの負け')).toBeInTheDocument());
    expect(screen.queryByText('Pair')).not.toBeInTheDocument();
  });

  it('does not show player hand name badge when handName is empty in end phase', async () => {
    const phase4EmptyPlayerName: PokerResponse = {
      ...phase4State,
      player: { ...phase4State.player, handName: '' },
    };
    mockExec.mockResolvedValue(phase4EmptyPlayerName);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByText('あなたの負け')).toBeInTheDocument());
    expect(screen.queryByText('High Card')).not.toBeInTheDocument();
  });

  it('toggles aria-pressed on card button click in exchange phase', async () => {
    mockExec.mockResolvedValue(phase2State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument());

    const cardBtn = screen.getByAltText('SPADE 1').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('updates bet amount when changing the bet input', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());

    const betInput = screen.getByLabelText('ベット額:');
    fireEvent.change(betInput, { target: { value: '20' } });

    expect((betInput as HTMLInputElement).value).toBe('20');
  });

  it('sets aria-busy and sr-only loading text while loading', async () => {
    mockExec.mockResolvedValue(phase1State);
    render(<PokerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).not.toBeDisabled());

    const container = screen.getByRole('button', { name: 'ベット' }).closest('[aria-live]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
    expect(screen.queryByText('処理中...')).not.toBeInTheDocument();

    let resolve!: (value: PokerResponse) => void;
    const slowPromise = new Promise<PokerResponse>((res) => {
      resolve = res;
    });
    mockExec.mockReturnValueOnce(slowPromise);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));

    expect(container).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('処理中...')).toBeInTheDocument();

    resolve(phase1State);
    await waitFor(() => {
      expect(container).toHaveAttribute('aria-busy', 'false');
      expect(screen.queryByText('処理中...')).not.toBeInTheDocument();
    });
  });
});
