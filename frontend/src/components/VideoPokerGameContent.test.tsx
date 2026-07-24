import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { TutorialProvider } from '../providers/TutorialProvider';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, VideoPokerResponse } from '../types/card';
import type { TutorialConfig } from '../types/tutorial';
import { VideoPokerGameContent } from './VideoPokerGameContent';

vi.mock('../api/gameApi', () => ({
  videopokerApi: { exec: vi.fn() },
  actionLogApi: {
    videopoker: vi.fn().mockResolvedValue([]),
    deuceswild: vi.fn().mockResolvedValue([]),
    jokerpoker: vi.fn().mockResolvedValue([]),
  },
}));

const mockExec = vi.fn();

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: VideoPokerResponse = {
  hand: [],
  phase: 1,
  chips: 1000,
  betAmount: 0,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: '',
};

const drawPhaseState: VideoPokerResponse = {
  hand: [card('SPADE', 1), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 13)],
  phase: 2,
  chips: 997,
  betAmount: 3,
  result: 0,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: '',
};

const resultPhaseWin: VideoPokerResponse = {
  hand: [card('SPADE', 11), card('CLOVER', 11), card('HEART', 3), card('DIAMOND', 5), card('SPADE', 9)],
  phase: 3,
  chips: 1001,
  betAmount: 1,
  result: 1,
  payout: 5,
  handRank: 1,
  handName: 'Jacks or Better',
  heldIndices: [true, true, false, false, false],
  variantName: 'jacksorbetter',
  message: 'Jacks or Better! You win!',
  messageCode: 'videopoker.result.win',
  messageParams: { handName: 'Jacks or Better', payout: '5' },
};

const resultPhaseLose: VideoPokerResponse = {
  hand: [card('SPADE', 2), card('CLOVER', 5), card('HEART', 7), card('DIAMOND', 9), card('SPADE', 11)],
  phase: 3,
  chips: 999,
  betAmount: 1,
  result: -1,
  payout: 0,
  handRank: 0,
  handName: '',
  heldIndices: [false, false, false, false, false],
  variantName: 'jacksorbetter',
  message: 'No winning hand.',
  messageCode: 'videopoker.result.lose',
};

const tutorialConfig: TutorialConfig = {
  gameName: 'videopoker',
  steps: [],
};

function renderContent(gameName: 'videopoker' | 'deuceswild' | 'jokerpoker' = 'videopoker') {
  return renderWithProviders(
    <TutorialProvider config={tutorialConfig} translateMessage={(k: string) => k}>
      <VideoPokerGameContent
        gameName={gameName}
        i18nNamespace={gameName}
        apiExec={mockExec}
        gamePath={`/${gameName}`}
        cliGameConfig={{
          parseCommand: () => ({ args: ['reset'] }),
          formatResponse: () => '',
          helpText: [],
        }}
      />
    </TutorialProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  // Clear localStorage so the per-variant autoHold/hint toggles start at
  // their hardcoded defaults rather than carrying state across test cases.
  localStorage.clear();
});

describe('VideoPokerGameContent', () => {
  it('calls reset on mount and renders bet phase', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/チップ.*1000/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument();
  });

  it('bets the maximum (5) and deals when Bet Max is clicked', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '最大ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 5));
  });

  it('renders draw phase with hold toggles', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    // Toggle hold on first card
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBe(5);
    fireEvent.click(cardButtons[0]);
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
  });

  it('sends hold indices on draw', async () => {
    mockExec
      .mockResolvedValueOnce(betPhaseState)
      .mockResolvedValueOnce(drawPhaseState)
      .mockResolvedValueOnce(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Disable auto-hold before entering draw so the test can drive holds
    // purely from explicit clicks (auto-hold seeds the state on phase entry).
    fireEvent.click(screen.getByTestId('vp-auto-hold-toggle'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    // Hold first two cards
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    fireEvent.click(cardButtons[0]);
    fireEvent.click(cardButtons[1]);

    fireEvent.click(screen.getByRole('button', { name: /ドロー/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hold', undefined, [0, 1]));
  });

  it('renders result phase with payout on win', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByText(/配当.*5/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument();
  });

  it('renders result phase on lose', async () => {
    mockExec.mockResolvedValue(resultPhaseLose);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
  });

  it('auto-hold pre-selects the hint-recommended cards on entering draw phase', async () => {
    // drawPhaseState hand = [A♠, J♥, 5♣, 8♦, K♠]. With no pairs/draws, the
    // base hint engine recommends holding the high cards (J + K → idx 1, 4).
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'false'); // A — not a high card under threshold=11
    expect(cardButtons[1]).toHaveAttribute('aria-pressed', 'true'); // J
    expect(cardButtons[2]).toHaveAttribute('aria-pressed', 'false'); // 5
    expect(cardButtons[3]).toHaveAttribute('aria-pressed', 'false'); // 8
    expect(cardButtons[4]).toHaveAttribute('aria-pressed', 'true'); // K
  });

  it('toggling auto-hold off then dealing leaves all cards unheld', async () => {
    mockExec.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByTestId('vp-auto-hold-toggle'));
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());

    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    for (const btn of cardButtons) {
      expect(btn).toHaveAttribute('aria-pressed', 'false');
    }
  });

  it('next game button fires reset directly without confirm dialog', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /次のゲーム/ }));
    expect(screen.queryByText(/リセットしますか/)).not.toBeInTheDocument();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('changes bet amount with selector', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    const select = screen.getByLabelText(/コイン数/);
    fireEvent.change(select, { target: { value: '5' } });
    fireEvent.click(screen.getByRole('button', { name: /ディール/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 5));
  });

  it('renders payout table collapsed in draw phase once dismissed', async () => {
    localStorage.setItem('paytable_open_videopoker', 'false');
    mockExec.mockResolvedValue(drawPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByText(/配当表/)).toBeInTheDocument());
    // Once the player has collapsed the table, it stays collapsed across phases.
    const details = screen.getByText(/配当表/).closest('details');
    expect(details).not.toHaveAttribute('open');
  });

  it('shows payout table during bet phase for reference', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    expect(screen.getByText(/配当表/)).toBeInTheDocument();
  });

  it('renders the payout grid with bet-scaled cell values (royal jackpot at max bet)', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByTestId('vp-payout-row-royalFlush')).toBeInTheDocument());
    const royalRow = within(screen.getByTestId('vp-payout-row-royalFlush'));
    expect(royalRow.getByText('250')).toBeInTheDocument(); // 1-coin
    expect(royalRow.getByText('1000')).toBeInTheDocument(); // 4-coin (250x4)
    expect(royalRow.getByText('4000')).toBeInTheDocument(); // 5-coin jackpot
  });

  it('highlights the winning hand row in result phase', async () => {
    mockExec.mockResolvedValue(resultPhaseWin); // handName: 'Jacks or Better'
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.getByTestId('vp-payout-row-jacksOrBetter')).toHaveAttribute('aria-current', 'true');
    expect(screen.getByTestId('vp-payout-row-royalFlush')).not.toHaveAttribute('aria-current');
  });

  it('highlights the winning row by handKey, not the English handName', async () => {
    // handKey points at wildRoyalFlush while handName is a mismatching English
    // string; the row highlight must follow the stable key.
    mockExec.mockResolvedValue({
      ...resultPhaseWin,
      variantName: 'deuceswild',
      handKey: 'wildRoyalFlush',
      handName: 'SOME UNTRANSLATED STRING',
    });
    renderContent('deuceswild');
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.getByTestId('vp-payout-row-wildRoyalFlush')).toHaveAttribute('aria-current', 'true');
    expect(screen.getByTestId('vp-payout-row-fourDeuces')).not.toHaveAttribute('aria-current');
  });

  it('clicking card in result phase does not toggle hold', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.length).toBe(5);
    // heldIndices[0] is true in resultPhaseWin, so aria-pressed starts as "true"
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(cardButtons[0]);
    // Should remain held because toggleHold returns early when not in draw phase
    expect(cardButtons[0]).toHaveAttribute('aria-pressed', 'true');
  });

  it('shows action log panel when view log button is clicked', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /棋譜を見る/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /棋譜を見る/ }));
    // After clicking, the action log panel should appear (even if empty)
    await waitFor(() => expect(screen.getByText(/棋譜/)).toBeInTheDocument());
  });

  it('renders result phase with undefined heldIndices (falls back to empty)', async () => {
    const resultNoHeld = {
      ...resultPhaseLose,
      heldIndices: undefined,
    } as unknown as VideoPokerResponse;
    mockExec.mockResolvedValue(resultNoHeld);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    const cardButtons = screen.getAllByRole('button').filter((b) => b.getAttribute('aria-pressed') !== null);
    expect(cardButtons.every((b) => b.getAttribute('aria-pressed') === 'false')).toBe(true);
  });

  it('renders the HOLD badge overlay only on held cards', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    // resultPhaseWin holds card indices 0 and 1; 2-4 are not held.
    expect(screen.getByTestId('vp-hold-badge-0')).toBeInTheDocument();
    expect(screen.getByTestId('vp-hold-badge-1')).toBeInTheDocument();
    expect(screen.queryByTestId('vp-hold-badge-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-hold-badge-3')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-hold-badge-4')).not.toBeInTheDocument();
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('expands the payout table on first visit', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const details = screen.getByText(/配当表/).closest('details') as HTMLDetailsElement;
    expect(details.open).toBe(true);
  });

  it('keeps the payout table collapsed once dismissed', async () => {
    localStorage.setItem('paytable_open_videopoker', 'false');
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect((screen.getByText(/配当表/).closest('details') as HTMLDetailsElement).open).toBe(false);
  });

  it('persists the collapsed state when the payout table is toggled shut', async () => {
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const details = screen.getByText(/配当表/).closest('details') as HTMLDetailsElement;
    details.open = false;
    fireEvent(details, new Event('toggle'));
    expect(localStorage.getItem('paytable_open_videopoker')).toBe('false');
  });

  it('marks twos with a WILD badge in Deuces Wild', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      variantName: 'deuceswild',
      hand: [card('SPADE', 2), card('HEART', 11), card('CLOVER', 2), card('DIAMOND', 8), card('SPADE', 13)],
    });
    renderContent('deuceswild');
    await waitFor(() => expect(screen.getByTestId('vp-wild-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('vp-wild-badge-2')).toBeInTheDocument();
    // Wild status is surfaced to screen readers via the button aria-label.
    expect(screen.getAllByRole('button', { name: /ワイルド/ })).toHaveLength(2);
    expect(screen.queryByTestId('vp-wild-badge-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-wild-badge-3')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-wild-badge-4')).not.toBeInTheDocument();
  });

  it('marks jokers with a WILD badge in Joker Poker', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      variantName: 'jokerpoker',
      hand: [card('JOKER', 0), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 13)],
    });
    renderContent('jokerpoker');
    await waitFor(() => expect(screen.getByTestId('vp-wild-badge-0')).toBeInTheDocument());
    expect(screen.queryByTestId('vp-wild-badge-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-wild-badge-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-wild-badge-3')).not.toBeInTheDocument();
    expect(screen.queryByTestId('vp-wild-badge-4')).not.toBeInTheDocument();
  });

  it('shows no WILD badge in Jacks or Better even on twos', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      hand: [card('SPADE', 2), card('HEART', 11), card('CLOVER', 5), card('DIAMOND', 8), card('SPADE', 13)],
    });
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /ドロー/ })).toBeInTheDocument());
    expect(screen.queryByTestId('vp-wild-badge-0')).not.toBeInTheDocument();
  });

  it('records a win into the session stats (win rate + net) and persists it', async () => {
    // resultPhaseWin: bet 1, payout 5, Jacks or Better -> net +4, win rate 100%.
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByTestId('vp-stats-summary')).toBeInTheDocument());
    const summary = screen.getByTestId('vp-stats-summary');
    expect(summary).toHaveTextContent('1 ハンド');
    expect(summary).toHaveTextContent('勝率 100%');
    expect(summary).toHaveTextContent('収支 +4');
    // The winning hand is tallied in the payout table's count column.
    expect(screen.getByTestId('vp-payout-count-jacksOrBetter')).toHaveTextContent('1');
    // And it is written to localStorage under the per-variant key.
    const stored = JSON.parse(localStorage.getItem('vp_stats_videopoker') ?? '{}');
    expect(stored).toMatchObject({ hands: 1, wins: 1, totalBet: 1, totalPayout: 5, handCounts: { jacksOrBetter: 1 } });
  });

  it('reads persisted stats back on mount', async () => {
    localStorage.setItem(
      'vp_stats_videopoker',
      JSON.stringify({ hands: 4, wins: 1, totalBet: 4, totalPayout: 10, handCounts: { flush: 1 } }),
    );
    mockExec.mockResolvedValue(betPhaseState);
    renderContent();
    await waitFor(() => expect(screen.getByRole('button', { name: /ディール/ })).toBeInTheDocument());
    const summary = screen.getByTestId('vp-stats-summary');
    expect(summary).toHaveTextContent('4 ハンド');
    expect(summary).toHaveTextContent('勝率 25%');
    expect(summary).toHaveTextContent('収支 +6');
    expect(screen.getByTestId('vp-payout-count-flush')).toHaveTextContent('1');
  });

  it('clears the session stats when the clear button is clicked', async () => {
    mockExec.mockResolvedValue(resultPhaseWin);
    renderContent();
    await waitFor(() => expect(screen.getByTestId('vp-stats-clear')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('vp-stats-clear'));
    await waitFor(() => expect(screen.getByTestId('vp-stats-summary')).toHaveTextContent(/まだプレイ記録がありません/));
    const stored = JSON.parse(localStorage.getItem('vp_stats_videopoker') ?? '{}');
    expect(stored).toMatchObject({ hands: 0, wins: 0 });
  });

  it('does not show the session stats panel for other variants (deuceswild)', async () => {
    mockExec.mockResolvedValue({ ...resultPhaseWin, variantName: 'deuceswild' });
    renderContent('deuceswild');
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.queryByTestId('vp-session-stats')).not.toBeInTheDocument();
    // And nothing is written to any stats bucket for the disabled variant.
    expect(localStorage.getItem('vp_stats_deuceswild')).toBeNull();
  });
});
