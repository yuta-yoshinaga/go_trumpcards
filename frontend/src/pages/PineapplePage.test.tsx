import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyPineappleApi, irishPokerApi, pineappleApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PineappleResponse } from '../types/card';
import { PineapplePage } from './PineapplePage';

vi.mock('../api/gameApi', () => ({
  pineappleApi: { exec: vi.fn() },
  irishPokerApi: { exec: vi.fn() },
  crazyPineappleApi: { exec: vi.fn() },
  actionLogApi: { pineapple: vi.fn(), irishpoker: vi.fn(), crazypineapple: vi.fn() },
}));

const mockExec = vi.mocked(pineappleApi.exec);
const mockIrishExec = vi.mocked(irishPokerApi.exec);
const mockCrazyExec = vi.mocked(crazyPineappleApi.exec);

/** Helper: base human player */
const humanPlayer = (overrides: Partial<import('../types/card').HoldemPlayerData> = {}) => ({
  id: 0,
  isHuman: true,
  cards: [
    { design: 'SPADE' as const, value: 1 },
    { design: 'HEART' as const, value: 13 },
    { design: 'DIAMOND' as const, value: 7 },
  ],
  chips: 980,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

/** Helper: base CPU player */
const cpuPlayer = (id: number, overrides: Partial<import('../types/card').HoldemPlayerData> = {}) => ({
  id,
  isHuman: false,
  cards: [
    { design: 'DIAMOND' as const, value: 2 },
    { design: 'CLOVER' as const, value: 7 },
  ],
  chips: 1000,
  currentBet: 0,
  folded: false,
  allIn: false,
  handRank: 0,
  handName: '',
  bestHand: [],
  playStyleName: '',
  totalHands: 0,
  vpip: 0,
  pfr: 0,
  threeBet: 0,
  af: '-',
  ...overrides,
});

/** INIT state (phase 0): no players yet */
const initState: PineappleResponse = {
  players: [],
  communityCards: [],
  pot: 0,
  sidePots: [],
  dealerIdx: 0,
  currentTurn: 0,
  phase: 0,
  gameEndFlag: false,
  lastBet: 0,
  minRaise: 20,
  roundResults: [],
  cpuActions: [],
  message: '',
  handCount: 0,
  smallBlind: 5,
  bigBlind: 10,
  tournamentMode: false,
  blindLevelHands: 10,
  blindMultiplier: 200,
  bettingLimit: 0,
  raiseCount: 0,
  maxBetAmount: 0,
  tableSize: 4,
  rebuyPhaseType: 0,
  rebuyChips: 0,
  rebuyMaxCount: 0,
  rebuyCounts: [],
  addonChips: 0,
  rebuyAvailable: false,
  addonAvailable: false,
  rebuyEnabled: false,
  addonEnabled: false,
  rebuyPeriodHands: 0,
  addonAfterHand: 0,
  addonUsed: [],
  muckAvailable: false,
  isDiscardPhase: false,
  discardDone: [],
  initialDealCount: 3,
};

/** PRE_FLOP (phase 1): human's turn, no outstanding bet */
const preFlopState: PineappleResponse = {
  ...initState,
  players: [humanPlayer(), cpuPlayer(1), cpuPlayer(2), cpuPlayer(3)],
  pot: 30,
  dealerIdx: 3,
  currentTurn: 0,
  phase: 1,
  message: 'あなたの番です',
  handCount: 1,
  discardDone: [false, false, false, false],
};

/** Discard phase state */
const discardState: PineappleResponse = {
  ...preFlopState,
  phase: 8,
  isDiscardPhase: true,
  discardDone: [false, true, true, true],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
  ],
};

/** SHOWDOWN (phase 5) */
const showdownState: PineappleResponse = {
  ...initState,
  players: [
    humanPlayer({
      handName: 'ワンペア',
      currentBet: 0,
      chips: 950,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 13 },
      ],
    }),
    cpuPlayer(1, {
      handName: 'ツーペア',
      folded: false,
      cards: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
      ],
    }),
    cpuPlayer(2, { folded: true }),
  ],
  communityCards: [
    { design: 'SPADE', value: 10 },
    { design: 'HEART', value: 5 },
    { design: 'DIAMOND', value: 8 },
    { design: 'CLOVER', value: 2 },
    { design: 'HEART', value: 9 },
  ],
  pot: 0,
  dealerIdx: 2,
  currentTurn: -1,
  phase: 5,
  roundResults: [
    { playerIdx: 0, handRank: 1, handName: 'ワンペア', kickers: 'A, Q, 10', bestHand: [], wonAmount: 0, mucked: false },
    { playerIdx: 1, handRank: 2, handName: 'ツーペア', kickers: '8', bestHand: [], wonAmount: 200, mucked: false },
  ],
  message: 'CPU 1 の勝ち',
  handCount: 1,
  discardDone: [true, true, true],
};

beforeEach(() => {
  mockExec.mockResolvedValue(initState);
});

describe('PineapplePage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PineapplePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows phase name "初期化中" for INIT phase', async () => {
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('初期化中')).toBeInTheDocument());
  });

  it('shows pot and the dealer name via playerName (CPU dealer)', async () => {
    mockExec.mockResolvedValue(preFlopState); // dealerIdx 3 → CPU
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText(/ポット:/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー:/)).toBeInTheDocument();
    // Dealer renders via playerName (CPU 3), not the raw index.
    expect(screen.getAllByText('CPU 3').length).toBeGreaterThan(0);
    expect(screen.queryByText(/Player 3|プレイヤー 3/)).not.toBeInTheDocument();
  });

  it('shows betting controls during pre-flop when it is human turn', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('renders discard controls during discard phase', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    expect(screen.getByText('捨てるカードを選択してください')).toBeInTheDocument();
  });

  it('allows selecting a card and clicking discard', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());

    // Discard button should be disabled until a card is selected
    const discardBtn = screen.getByRole('button', { name: '1枚捨ててください。' });
    expect(discardBtn).toBeDisabled();

    // Click the first card
    const cardButtons = screen.getAllByRole('button').filter((btn) => btn.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);

    // Now discard button should be enabled
    await waitFor(() => expect(discardBtn).not.toBeDisabled());

    // Pressing discard now opens an inline confirm step rather than committing immediately.
    fireEvent.click(discardBtn);
    expect(mockExec).not.toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] });
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '確定' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] }));
  });

  it('previews the kept cards and tentative hand for Irish Poker discard', async () => {
    const irishDiscardState: PineappleResponse = {
      ...discardState,
      initialDealCount: 4,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 1 },
            { design: 'DIAMOND', value: 5 },
            { design: 'CLOVER', value: 8 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
        cpuPlayer(3),
      ],
      communityCards: [
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 8 },
      ],
      discardDone: [false, true, true, true],
    };
    mockIrishExec.mockResolvedValue(irishDiscardState);
    renderWithProviders(<PineapplePage variant="irishpoker" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());

    // Discard ♦5 and ♣8, keeping the pair of Aces.
    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 8').closest('button') as HTMLButtonElement);

    const preview = await screen.findByTestId('irishpoker-discard-preview');
    // Kept ♠A ♥A + board makes one pair.
    expect(preview).toHaveTextContent('ワンペア');
  });

  it('evaluates the best five when the board has more than three cards', async () => {
    const irishRiverState: PineappleResponse = {
      ...discardState,
      initialDealCount: 4,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 1 },
            { design: 'DIAMOND', value: 5 },
            { design: 'CLOVER', value: 8 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
        cpuPlayer(3),
      ],
      // Five board cards → kept(2) + board(5) = 7, so holdemBestFive picks the best 5.
      communityCards: [
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 8 },
        { design: 'CLOVER', value: 2 },
        { design: 'HEART', value: 9 },
      ],
      discardDone: [false, true, true, true],
    };
    mockIrishExec.mockResolvedValue(irishRiverState);
    renderWithProviders(<PineapplePage variant="irishpoker" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 8').closest('button') as HTMLButtonElement);
    const preview = await screen.findByTestId('irishpoker-discard-preview');
    expect(preview).toHaveTextContent('ワンペア');
  });

  it('shows kept cards without a hand badge when the board is too small to evaluate', async () => {
    const irishNoBoardState: PineappleResponse = {
      ...discardState,
      initialDealCount: 4,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 1 },
            { design: 'DIAMOND', value: 5 },
            { design: 'CLOVER', value: 8 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
        cpuPlayer(3),
      ],
      communityCards: [], // fewer than 3 board cards → no hand can be formed
      discardDone: [false, true, true, true],
    };
    mockIrishExec.mockResolvedValue(irishNoBoardState);
    renderWithProviders(<PineapplePage variant="irishpoker" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 8').closest('button') as HTMLButtonElement);
    const preview = await screen.findByTestId('irishpoker-discard-preview');
    // Kept cards are shown, but no tentative-hand badge.
    expect(preview).toHaveTextContent('残す札');
    expect(preview).not.toHaveTextContent('暫定役');
  });

  it('does not show the Irish Poker discard preview outside the discard phase', async () => {
    mockIrishExec.mockResolvedValue({ ...preFlopState, initialDealCount: 4 });
    renderWithProviders(<PineapplePage variant="irishpoker" />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.queryByTestId('irishpoker-discard-preview')).not.toBeInTheDocument();
  });

  it('annotates each Crazy Pineapple hole card with the keep-hand if discarded', async () => {
    const crazyDiscardState: PineappleResponse = {
      ...discardState,
      initialDealCount: 3,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 1 },
            { design: 'DIAMOND', value: 5 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
        cpuPlayer(3),
      ],
      communityCards: [
        { design: 'SPADE', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 8 },
      ],
      discardDone: [false, true, true, true],
    };
    mockCrazyExec.mockResolvedValue(crazyDiscardState);
    renderWithProviders(<PineapplePage variant="crazypineapple" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const labels = screen.getAllByTestId('cp-discard-candidate');
    expect(labels).toHaveLength(3); // one per hole card
    // Each keep makes at least a pair (Aces or fives), so a hand name is shown.
    expect(labels[0]).toHaveTextContent('ワンペア');
  });

  it('does not show Crazy Pineapple candidate labels outside the discard phase', async () => {
    mockCrazyExec.mockResolvedValue({ ...preFlopState, initialDealCount: 3 });
    renderWithProviders(<PineapplePage variant="crazypineapple" />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.queryByTestId('cp-discard-candidate')).not.toBeInTheDocument();
  });

  it('does not show the Irish Poker discard preview for the Pineapple variant', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((btn) => btn.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    expect(screen.queryByTestId('irishpoker-discard-preview')).not.toBeInTheDocument();
  });

  it('cancels the discard confirm step and returns to selection', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    mockExec.mockClear(); // isolate from prior tests' discard calls (beforeEach doesn't clear history)

    const cardButtons = screen.getAllByRole('button').filter((btn) => btn.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(screen.getByRole('button', { name: '1枚捨ててください。' }));
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    // Back to selection; nothing discarded.
    expect(screen.queryByTestId('discard-confirm')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '1枚捨ててください。' })).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] });
  });

  it('shows round results during showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
  });

  it('highlights the winning best-5 cards at showdown', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('ワンペア')).toBeInTheDocument());
    // 2 hole + 3 board cards form the best 5.
    expect(screen.getAllByTestId('pn-best5-card').length).toBe(5);
  });

  it('does not highlight best-5 when the human has folded', async () => {
    mockExec.mockResolvedValue({
      ...showdownState,
      players: [
        humanPlayer({ handName: 'ワンペア', folded: true, cards: showdownState.players[0].cards }),
        showdownState.players[1],
        showdownState.players[2],
      ],
    });
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('ツーペア')).toBeInTheDocument());
    expect(screen.queryByTestId('pn-best5-card')).not.toBeInTheDocument();
  });

  it('shows reset button and triggers reset flow', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());

    const resetBtn = screen.getByRole('button', { name: 'リセット' });
    expect(resetBtn).toBeInTheDocument();
    fireEvent.click(resetBtn);
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
  });

  it('wraps settings in collapsible details element', async () => {
    mockExec.mockResolvedValue(preFlopState);
    const { container } = renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    const allSummaries = container.querySelectorAll('details summary');
    const settingsSummary = Array.from(allSummaries).find((s) => s.textContent?.includes('設定'));
    expect(settingsSummary).toBeTruthy();
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(showdownState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuMetaAI: false }));
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });
});
