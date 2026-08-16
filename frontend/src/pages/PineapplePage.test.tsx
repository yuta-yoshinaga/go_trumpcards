import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyPineappleApi, irishPokerApi, pineappleApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
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
  liveBestHand: '',
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

  // **3 枚配って 1 枚捨てる game なので、どの 2 枚を残すかの判断材料が要る。**
  // Omaha は暫定ベストを常時出しているのに Pineapple には無かった (#5488)。
  it('shows the live best hand the server sent', async () => {
    mockExec.mockResolvedValue({
      ...initState,
      players: [humanPlayer(), cpuPlayer(1)],
      liveBestHand: 'fullHouse',
    });
    renderWithProviders(<PineapplePage />);
    const badge = await screen.findByTestId('pineapple-live-besthand-name');
    // キーではなく訳文が出ること。キーのまま出ていたら i18n を素通りしている。
    expect(badge).toHaveTextContent('フルハウス');
    expect(badge).not.toHaveTextContent('fullHouse');
  });

  it('shows no live best hand when the server sent none', async () => {
    mockExec.mockResolvedValue({ ...initState, players: [humanPlayer(), cpuPlayer(1)], liveBestHand: '' });
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('pineapple-live-besthand')).not.toBeInTheDocument();
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
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] });
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '確定' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] }));
  });

  it('warns via a live region when clicking past the discard cap', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);

    // Cap is 1 (3 dealt → discard 1). Selecting card 0 reaches the cap; no warning yet.
    fireEvent.click(cardButtons[0]);
    expect(screen.queryByTestId('pn-discard-limit-announce')).not.toBeInTheDocument();

    // Clicking a second, unselected card is over the cap → the warning appears.
    fireEvent.click(cardButtons[1]);
    const announce = await screen.findByTestId('pn-discard-limit-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toHaveTextContent('選択できるのは1枚まで');
  });

  it('describes the discard cap on the hand group and does not warn on normal deselect', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);

    const group = cardButtons[0].closest('[aria-describedby]');
    expect(group).toHaveAttribute('aria-describedby', 'pn-discard-limit-desc');
    expect(document.getElementById('pn-discard-limit-desc')).toHaveTextContent('最大1枚');

    // Selecting then deselecting the same card is normal — no warning.
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[0]);
    expect(screen.queryByTestId('pn-discard-limit-announce')).not.toBeInTheDocument();
  });

  it('a card click is inert once the confirm step has started', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    // Select the first card and open the confirm step.
    fireEvent.click(cardButtons[0]);
    const discardBtn = screen.getByRole('button', { name: '1枚捨ててください。' });
    await waitFor(() => expect(discardBtn).not.toBeDisabled());
    fireEvent.click(discardBtn);
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();
    // Clicking another card during confirm must not change the selection or
    // dismiss the dialog (consistent with the keyboard behavior).
    fireEvent.click(cardButtons[1]);
    expect(cardButtons[1]).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();
  });

  it('keyboard: a number key selects a card and Enter steps through confirm to commit', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    mockExec.mockClear();
    // "1" selects the first hole card.
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true'));
    // First Enter opens the confirm step without committing.
    fireEvent.keyDown(document.body, { key: 'Enter' });
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] });
    // Second Enter commits the discard.
    fireEvent.keyDown(document.body, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', undefined, { cardIdxs: [0] }));
  });

  it('keyboard: Escape deselects a discard candidate', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.keyDown(document.body, { key: 'Escape' });
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false'));
  });

  it('keyboard: re-pressing deselects, the 1-card cap blocks a second pick, and empty Enter is a no-op', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    mockExec.mockClear();
    // Select card 0.
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true'));
    // Crazy Pineapple discards exactly 1, so pressing "2" must not also select card 1.
    fireEvent.keyDown(document.body, { key: '2' });
    expect(cardButtons[1]).toHaveAttribute('aria-pressed', 'false');
    // Re-pressing "1" deselects card 0.
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false'));
    // With nothing selected, Enter neither opens the confirm step nor commits.
    fireEvent.keyDown(document.body, { key: 'Enter' });
    expect(screen.queryByTestId('discard-confirm')).not.toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('discard', undefined, expect.anything());
  });

  it('keyboard: Escape from the confirm step returns to selection without committing', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: '1' });
    await waitFor(() => expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.keyDown(document.body, { key: 'Enter' });
    expect(screen.getByTestId('discard-confirm')).toBeInTheDocument();
    // Number keys are inert while confirming — the selection can't be changed.
    fireEvent.keyDown(document.body, { key: '2' });
    expect(cardButtons[1]).toHaveAttribute('aria-pressed', 'false');
    // Escape backs out of the confirm step; the selection is kept.
    fireEvent.keyDown(document.body, { key: 'Escape' });
    await waitFor(() => expect(screen.queryByTestId('discard-confirm')).not.toBeInTheDocument());
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
    expect(mockExec).not.toHaveBeenCalledWith('discard', undefined, expect.anything());
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

  // #5490: 残す2枚と役はディスカード確定前の重要なフィードバックなのに、
  // 同じフッター内の上限超過通知 (pn-discard-limit-announce) が aria-live を
  // 持つのに対し、こちらは無音だった。
  it('announces the kept cards and the hand to a screen reader', async () => {
    const irishDiscardState: PineappleResponse = {
      ...discardState,
      initialDealCount: 4,
      players: [
        {
          ...discardState.players[0],
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 1 },
            { design: 'DIAMOND', value: 5 },
            { design: 'CLOVER', value: 8 },
          ],
        },
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

    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 8').closest('button') as HTMLButtonElement);

    const live = await screen.findByTestId('irishpoker-discard-preview-announce');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveClass('sr-only');
    // **残す2枚と役の両方を読ませる。** どちらが欠けても選択を確認できない。
    expect(live.textContent).toContain('♠ A');
    expect(live.textContent).toContain('♥ A');
    expect(live.textContent).toContain('ワンペア');
  });

  // 1枚しか選んでいない間は出さない。確定した選択だけを読み上げる。
  it('stays silent until the second discard is chosen', async () => {
    const irishDiscardState: PineappleResponse = {
      ...discardState,
      initialDealCount: 4,
      players: [
        {
          ...discardState.players[0],
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 1 },
            { design: 'DIAMOND', value: 5 },
            { design: 'CLOVER', value: 8 },
          ],
        },
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

    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);
    expect(screen.queryByTestId('irishpoker-discard-preview-announce')).not.toBeInTheDocument();
  });

  it('annotates each remaining Irish Poker card once the first discard is chosen', async () => {
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

    // Choose only the FIRST discard (♦5): a staged preview should appear for each
    // of the three still-selectable cards (the candidate second discards).
    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);

    const labels = await screen.findAllByTestId('irishpoker-discard-candidate');
    expect(labels).toHaveLength(3); // one per remaining card, excluding the chosen one
    // Every kept pair (Aces or eights) makes at least one pair with the board.
    for (const label of labels) expect(label).toHaveTextContent('ワンペア');
    // The full two-card kept preview is not shown yet (only one card selected).
    expect(screen.queryByTestId('irishpoker-discard-preview')).not.toBeInTheDocument();
  });

  it('switches from staged candidates to the kept preview at the second Irish Poker discard', async () => {
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

    fireEvent.click(screen.getByAltText('♦ 5').closest('button') as HTMLButtonElement);
    expect(await screen.findAllByTestId('irishpoker-discard-candidate')).toHaveLength(3);

    // Selecting the second discard hands off to the full kept-cards preview.
    fireEvent.click(screen.getByAltText('♣ 8').closest('button') as HTMLButtonElement);
    expect(await screen.findByTestId('irishpoker-discard-preview')).toBeInTheDocument();
    expect(screen.queryAllByTestId('irishpoker-discard-candidate')).toHaveLength(0);

    // Deselecting one card returns to the staged, per-candidate previews.
    fireEvent.click(screen.getByAltText('♣ 8').closest('button') as HTMLButtonElement);
    expect(await screen.findAllByTestId('irishpoker-discard-candidate')).toHaveLength(3);
    expect(screen.queryByTestId('irishpoker-discard-preview')).not.toBeInTheDocument();
  });

  it('does not stage Irish Poker candidate previews before any card is chosen', async () => {
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
    // No discard chosen yet → no staged candidate annotations.
    expect(screen.queryByTestId('irishpoker-discard-candidate')).not.toBeInTheDocument();
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

  it('shows the Irish Poker discard-upcoming banner during the flop bet phase', async () => {
    // phase 2 = FLOP betting round; Irish Poker discards two cards afterwards.
    mockIrishExec.mockResolvedValue({ ...preFlopState, phase: 2, initialDealCount: 4 });
    renderWithProviders(<PineapplePage variant="irishpoker" />);
    await waitFor(() => expect(screen.getByTestId('cp-discard-upcoming-banner')).toBeInTheDocument());
    expect(screen.getByTestId('cp-discard-upcoming-banner')).toHaveTextContent(
      'フロップベット終了後にカードを2枚捨てます',
    );
  });

  it('does not show the Irish Poker discard-upcoming banner outside the flop bet phase', async () => {
    // preFlopState is phase 1 (PRE_FLOP), before the flop betting round.
    mockIrishExec.mockResolvedValue({ ...preFlopState, initialDealCount: 4 });
    renderWithProviders(<PineapplePage variant="irishpoker" />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.queryByTestId('cp-discard-upcoming-banner')).not.toBeInTheDocument();
  });

  it('does not show the discard-upcoming banner for the plain Pineapple variant during the flop', async () => {
    // Plain Pineapple discards before the flop, so no forewarning banner applies.
    mockExec.mockResolvedValue({ ...preFlopState, phase: 2 });
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.queryByTestId('cp-discard-upcoming-banner')).not.toBeInTheDocument();
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

  it('flags the single optimal Crazy Pineapple discard with a recommended badge', async () => {
    // hole 10♠ 10♥ 2♦, board 10♣ 5♥ 8♦.
    // Discard 2♦ → keep 10♠ 10♥ → three tens (best). The two 10-discards only
    // leave a pair of tens, so only the 2♦ discard is recommended.
    const crazyDiscardState: PineappleResponse = {
      ...discardState,
      initialDealCount: 3,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 10 },
            { design: 'HEART', value: 10 },
            { design: 'DIAMOND', value: 2 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
        cpuPlayer(3),
      ],
      communityCards: [
        { design: 'CLOVER', value: 10 },
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 8 },
      ],
      discardDone: [false, true, true, true],
    };
    mockCrazyExec.mockResolvedValue(crazyDiscardState);
    renderWithProviders(<PineapplePage variant="crazypineapple" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const badges = screen.getAllByTestId('cp-discard-recommended');
    expect(badges).toHaveLength(1); // only the strongest keep is recommended
    expect(badges[0]).toHaveTextContent('おすすめ');
    // The recommended card (3rd, discarding 2♦) keeps three of a kind.
    const labels = screen.getAllByTestId('cp-discard-candidate');
    expect(labels[2]).toHaveTextContent('スリーカード');
  });

  it('flags every tied-best Crazy Pineapple discard as recommended', async () => {
    // hole A♠ A♥ 5♦, board 10♠ 5♥ 8♦. Every discard leaves exactly one pair,
    // so all three tie for the best rank and all get the badge.
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
    expect(screen.getAllByTestId('cp-discard-recommended')).toHaveLength(3);
  });

  it('shows no Crazy Pineapple recommended badge when the board is too small', async () => {
    const crazyNoBoardState: PineappleResponse = {
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
      communityCards: [],
      discardDone: [false, true, true, true],
    };
    mockCrazyExec.mockResolvedValue(crazyNoBoardState);
    renderWithProviders(<PineapplePage variant="crazypineapple" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    expect(screen.queryByTestId('cp-discard-recommended')).not.toBeInTheDocument();
  });

  it('does not show a recommended badge for the plain Pineapple variant', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    expect(screen.queryByTestId('cp-discard-recommended')).not.toBeInTheDocument();
  });

  it('omits Crazy Pineapple candidate labels when the board is too small to evaluate', async () => {
    const crazyNoBoardState: PineappleResponse = {
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
      communityCards: [], // keep(2) + board(0) < 5 → no hand can be evaluated
      discardDone: [false, true, true, true],
    };
    mockCrazyExec.mockResolvedValue(crazyNoBoardState);
    renderWithProviders(<PineapplePage variant="crazypineapple" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    expect(screen.queryByTestId('cp-discard-candidate')).not.toBeInTheDocument();
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

  it('annotates each Pineapple hole card with the keep-2 feature during discard', async () => {
    const pineappleDiscardState: PineappleResponse = {
      ...discardState,
      players: [
        humanPlayer({
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'SPADE', value: 9 },
            { design: 'HEART', value: 5 },
          ],
        }),
        cpuPlayer(1),
        cpuPlayer(2),
        cpuPlayer(3),
      ],
    };
    mockExec.mockResolvedValue(pineappleDiscardState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    const notes = screen.getAllByTestId('pn-discard-keep-feature');
    expect(notes).toHaveLength(3); // one per hole card
    // Discard S5 → keep S9,H5: no pair/suited/connector → high card.
    expect(notes[0]).toHaveTextContent('残り2枚: ハイカード');
    // Discard S9 → keep S5,H5: a pair.
    expect(notes[1]).toHaveTextContent('残り2枚: ペア');
    // Discard H5 → keep S5,S9: same suit but not adjacent → suited.
    expect(notes[2]).toHaveTextContent('残り2枚: スーテッド');
  });

  it('does not show Pineapple keep-feature notes outside the discard phase', async () => {
    mockExec.mockResolvedValue(preFlopState);
    renderWithProviders(<PineapplePage />);
    await waitFor(() => expect(screen.getByText('あなたの手札')).toBeInTheDocument());
    expect(screen.queryByTestId('pn-discard-keep-feature')).not.toBeInTheDocument();
  });

  it('does not show Pineapple keep-feature notes for the Crazy Pineapple variant', async () => {
    mockCrazyExec.mockResolvedValue({ ...discardState, initialDealCount: 3 });
    renderWithProviders(<PineapplePage variant="crazypineapple" />);
    await waitFor(() => expect(screen.getByTestId('discard-controls')).toBeInTheDocument());
    expect(screen.queryByTestId('pn-discard-keep-feature')).not.toBeInTheDocument();
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
    await flushPendingDispatch();
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

  // The settings toggles are wrapped in a 44px-tall <label> so the whole row is
  // a tap target (DESIGN.md Interactive Element Minimum Size, issue #4368).
  // Clicking the label's *text* must therefore flip the checkbox -- that is the
  // behaviour the tap target buys, and it was previously untested here.
  describe('settings toggles are driven by their full label row', () => {
    it.each(['ラーニングモード', 'ヒント表示', 'メタAI（CPUがプレイスタイルを学習）'])('toggles %s', async (label) => {
      mockExec.mockResolvedValue(preFlopState);
      renderWithProviders(<PineapplePage />);
      const box = await waitFor(() => screen.getByLabelText(label) as HTMLInputElement);
      const before = box.checked;
      fireEvent.click(screen.getByText(label));
      expect((screen.getByLabelText(label) as HTMLInputElement).checked).toBe(!before);
    });
  });
});
