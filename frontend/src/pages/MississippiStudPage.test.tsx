import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { mississippiStudApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, MaskedCard, MississippiStudResponse } from '../types/card';
import { MississippiStudPhase } from '../types/phases';
import { MississippiStudPage } from './MississippiStudPage';

vi.mock('../api/gameApi', () => ({
  mississippiStudApi: { exec: vi.fn() },
  actionLogApi: { mississippistud: vi.fn() },
}));

const mockApi = vi.mocked(mississippiStudApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });
const maskedCard: MaskedCard = { design: '', value: 0 };

const antePhaseState: MississippiStudResponse = {
  playerHand: [],
  communityCards: [],
  communityRevealed: [false, false, false],
  phase: MississippiStudPhase.ANTE,
  chips: 1000,
  anteAmount: 0,
  streetMultipliers: [0, 0, 0],
  folded: false,
  totalBet: 0,
  result: 0,
  handRank: 0,
  payoutMultiplier: 0,
  antePayout: 0,
  streetPayouts: [0, 0, 0],
  totalPayout: 0,
  message: '',
};

const thirdStreetState: MississippiStudResponse = {
  ...antePhaseState,
  phase: MississippiStudPhase.THIRD_STREET,
  playerHand: [card('SPADE', 11), card('HEART', 11)],
  communityCards: [maskedCard, maskedCard, maskedCard],
  communityRevealed: [false, false, false],
  anteAmount: 100,
  totalBet: 100,
  chips: 900,
};

const fourthStreetState: MississippiStudResponse = {
  ...thirdStreetState,
  phase: MississippiStudPhase.FOURTH_STREET,
  communityCards: [card('DIAMOND', 2), maskedCard, maskedCard],
  communityRevealed: [true, false, false],
  streetMultipliers: [3, 0, 0],
  totalBet: 400,
  chips: 600,
};

const lowPairStreetState: MississippiStudResponse = {
  ...thirdStreetState,
  playerHand: [card('SPADE', 4), card('HEART', 4)],
};

const highCardStreetState: MississippiStudResponse = {
  ...thirdStreetState,
  playerHand: [card('SPADE', 8), card('HEART', 3)],
};

const endPhaseWin: MississippiStudResponse = {
  ...antePhaseState,
  phase: MississippiStudPhase.END,
  playerHand: [card('SPADE', 11), card('HEART', 11)],
  communityCards: [card('DIAMOND', 2), card('CLOVER', 3), card('SPADE', 4)],
  communityRevealed: [true, true, true],
  anteAmount: 100,
  streetMultipliers: [3, 1, 1],
  totalBet: 600,
  chips: 1600,
  result: 1,
  handRank: 1,
  payoutMultiplier: 1,
  antePayout: 200,
  streetPayouts: [600, 200, 200],
  totalPayout: 1200,
  message: 'Player wins!',
  messageCode: 'mississippistud.result.playerWins',
};

const endPhaseLoss: MississippiStudResponse = {
  ...endPhaseWin,
  result: -1,
  chips: 400,
  handRank: 0,
  payoutMultiplier: 0,
  antePayout: 0,
  streetPayouts: [0, 0, 0],
  totalPayout: 0,
  message: 'Player loses.',
  messageCode: 'mississippistud.result.playerLoses',
};

const endPhaseFold: MississippiStudResponse = {
  ...endPhaseLoss,
  communityCards: [maskedCard, maskedCard, maskedCard],
  communityRevealed: [false, false, false],
  streetMultipliers: [0, 0, 0],
  folded: true,
  totalBet: 100,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('MississippiStudPage', () => {
  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<MississippiStudPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls execApi with reset on mount', async () => {
    mockApi.mockResolvedValue(antePhaseState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders ante phase with chips and ante button', async () => {
    mockApi.mockResolvedValue(antePhaseState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'アンティ' })).toBeInTheDocument();
  });

  it('shows payout reference panel in ante phase', async () => {
    mockApi.mockResolvedValue(antePhaseState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アンティ' })).toBeInTheDocument());
    expect(screen.getByText('ペイテーブル')).toBeInTheDocument();
  });

  it('shows the collapsible payout reference panel during a betting street', async () => {
    mockApi.mockResolvedValue(thirdStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('ms-play-1x')).toBeInTheDocument());
    const paytable = screen.getByTestId('ms-paytable');
    expect(paytable).toBeInTheDocument();
    expect(paytable.tagName).toBe('DETAILS');
    // Default closed so it does not crowd the betting layout.
    expect(paytable).not.toHaveAttribute('open');
    expect(paytable).toHaveTextContent('ペイテーブル');
    expect(paytable).toHaveTextContent('ロイヤルフラッシュ: 500:1');
  });

  it('shows the payout reference panel in the end phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.getByTestId('ms-paytable')).toHaveTextContent('ペイテーブル');
  });

  it('calls execApi with bet and default amount when ante button clicked', async () => {
    mockApi.mockResolvedValue(antePhaseState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アンティ' })).toBeInTheDocument());

    mockApi.mockResolvedValue(thirdStreetState);
    fireEvent.click(screen.getByRole('button', { name: 'アンティ' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('shows street buttons with the additional bet amount per multiplier (ante 100)', async () => {
    mockApi.mockResolvedValue(thirdStreetState); // anteAmount 100
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('ms-play-1x')).toBeInTheDocument());
    expect(screen.getByTestId('ms-play-1x')).toHaveTextContent('+100');
    expect(screen.getByTestId('ms-play-2x')).toHaveTextContent('+200');
    expect(screen.getByTestId('ms-play-3x')).toHaveTextContent('+300');
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows player hand and community card sections in street phase', async () => {
    mockApi.mockResolvedValue(thirdStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.getByText('コミュニティカード')).toBeInTheDocument();
  });

  it('shows the current made hand with a pay-table badge for a paying pair', async () => {
    mockApi.mockResolvedValue(thirdStreetState); // pair of jacks (11)
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('ms-made-hand')).toBeInTheDocument());
    const madeHand = screen.getByTestId('ms-made-hand');
    expect(madeHand).toHaveTextContent('現在');
    expect(madeHand).toHaveTextContent('ワンペア');
    expect(madeHand).toHaveTextContent('配当対象');
  });

  it('shows a low pair as a made hand without the pay-table badge', async () => {
    mockApi.mockResolvedValue(lowPairStreetState); // pair of 4s
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('ms-made-hand')).toBeInTheDocument());
    const madeHand = screen.getByTestId('ms-made-hand');
    expect(madeHand).toHaveTextContent('ワンペア');
    expect(madeHand).not.toHaveTextContent('配当対象');
  });

  it('hides the made-hand readout when no made hand exists yet (high card)', async () => {
    mockApi.mockResolvedValue(highCardStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.queryByTestId('ms-made-hand')).not.toBeInTheDocument();
  });

  it('does not show the made-hand readout in the end phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.queryByTestId('ms-made-hand')).not.toBeInTheDocument();
  });

  it('shows the community reveal-count status text', async () => {
    mockApi.mockResolvedValue(fourthStreetState); // communityRevealed [true,false,false]
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('community-status')).toBeInTheDocument());
    expect(screen.getByTestId('community-status')).toHaveTextContent('開示済み: 1 / 3');
  });

  it('shows bet-status with street multipliers in street phase', async () => {
    mockApi.mockResolvedValue(fourthStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('bet-status')).toBeInTheDocument());
    const status = screen.getByTestId('bet-status');
    expect(status).toHaveTextContent('3rd: 3x');
    expect(status).toHaveTextContent('4th: 0x');
    expect(status).toHaveTextContent('5th: 0x');
    expect(status).toHaveTextContent('合計ベット: 400');
  });

  it('calls execApi with play and multiplier=3 when 3倍 clicked', async () => {
    mockApi.mockResolvedValue(thirdStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('ms-play-3x')).toBeInTheDocument());

    mockApi.mockResolvedValue(fourthStreetState);
    fireEvent.click(screen.getByTestId('ms-play-3x'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, 3));
  });

  it('calls execApi with fold when フォールド clicked', async () => {
    mockApi.mockResolvedValue(thirdStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    mockApi.mockResolvedValue(endPhaseFold);
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('fold'));
  });

  it('shows end phase with reset button and payout breakdown on win', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toHaveTextContent('合計配当: 1200');
  });

  it('shows hand rank label in end phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<MississippiStudPage />);
    // handRank=1 → 'handRank.1' → 'ワンペア'
    await waitFor(() => expect(screen.getByText(/ワンペア/)).toBeInTheDocument());
  });

  it('shows zero total payout on loss', async () => {
    mockApi.mockResolvedValue(endPhaseLoss);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toHaveTextContent('合計配当: 0');
  });

  it('reset button fires reset without confirm dialog', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    mockApi.mockResolvedValue(antePhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows network error alert', async () => {
    mockApi.mockResolvedValueOnce(antePhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アンティ' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'アンティ' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(thirdStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('ms-play-1x')).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in street phase', async () => {
    localStorage.setItem('hint_enabled_mississippistud', 'true');
    mockApi.mockResolvedValue(thirdStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows view action log button in END phase', async () => {
    mockApi.mockResolvedValue(endPhaseWin);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
  });

  it('does not render CLI toggle (CLI not yet implemented)', async () => {
    mockApi.mockResolvedValue(antePhaseState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'アンティ' })).toBeInTheDocument());
    // CLI toggle is intentionally hidden until CLI command parsing is implemented
    expect(screen.queryByRole('button', { name: /CLI/i })).not.toBeInTheDocument();
  });
});

// --- keyboard shortcut execution (#4429) ---
// 1/2/3 are the raise multipliers and differ only by their third argument, so
// asserting that argument is the only thing that distinguishes them: a test that
// checked the command name alone would pass with all three wired to the same
// multiplier.
const kbdCases: [string, unknown[], MississippiStudResponse][] = [
  ['b', ['bet', 100], antePhaseState],
  ['1', ['play', undefined, 1], thirdStreetState],
  ['2', ['play', undefined, 2], thirdStreetState],
  ['3', ['play', undefined, 3], thirdStreetState],
  ['f', ['fold'], thirdStreetState],
  ['r', ['reset'], endPhaseWin],
];

describe('MississippiStudPage keyboard shortcuts', () => {
  it.each(kbdCases)('pressing %s dispatches %j', async (key, expected, state) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(state);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith(...expected));
  });

  // The street keys stay live across all three streets, not just the third.
  it('accepts a street raise on fourth street too', async () => {
    mockApi.mockResolvedValue(fourthStreetState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(fourthStreetState);
    fireEvent.keyDown(document, { key: '3' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', undefined, 3));
  });

  it('ignores a key whose phase gate is closed', async () => {
    // Folding is a street decision; there is nothing to fold before the ante.
    mockApi.mockResolvedValue(antePhaseState);
    renderWithProviders(<MississippiStudPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'f' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalled();
  });

  // #5591: 「役はフラッシュ、配当800」とは分かっても、その配当が**どの倍率から
  // 来たのか**を説明していなかった。サーバは既に送っていたのに読んでいなかった。
  describe('payout multiplier', () => {
    it('shows the odds next to the hand name', async () => {
      mockApi.mockResolvedValue({ ...endPhaseWin, payoutMultiplier: 8 });
      renderWithProviders(<MississippiStudPage />);
      const odds = await screen.findByTestId('ms-payout-multiplier');
      expect(odds).toHaveTextContent('8');
    });

    // **プッシュ (-1) とロス (0) は倍率ではない** (受け入れ条件3)。
    it.each([
      ['push', -1],
      ['loss', 0],
    ])('says nothing on a %s', async (_name, multiplier) => {
      mockApi.mockResolvedValue({ ...endPhaseWin, payoutMultiplier: multiplier });
      renderWithProviders(<MississippiStudPage />);
      await waitFor(() => expect(mockApi).toHaveBeenCalled());
      expect(screen.queryByTestId('ms-payout-multiplier')).not.toBeInTheDocument();
    });

    it('stays hidden before the hand is settled', async () => {
      mockApi.mockResolvedValue({ ...antePhaseState, payoutMultiplier: 8 });
      renderWithProviders(<MississippiStudPage />);
      await waitFor(() => expect(mockApi).toHaveBeenCalled());
      expect(screen.queryByTestId('ms-payout-multiplier')).not.toBeInTheDocument();
    });
  });
});
