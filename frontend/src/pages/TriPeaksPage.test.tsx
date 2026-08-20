import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, tripeaksApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { TRIPEAKS_STATS_KEY } from '../hooks/useTriPeaksStats';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, TriPeaksCard, TriPeaksResponse } from '../types/card';
import { computePeakRemaining, TriPeaksPage } from './TriPeaksPage';

vi.mock('../api/gameApi', () => ({
  tripeaksApi: { exec: vi.fn() },
  actionLogApi: { tripeaks: vi.fn() },
}));

const mockPlaySound = vi.fn();
vi.mock('../providers/SoundProvider', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../providers/SoundProvider')>();
  return { ...actual, useSound: () => ({ playSound: mockPlaySound, muted: false, toggleMute: vi.fn() }) };
});

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

vi.mock('../hooks/useChainCombo', () => ({
  useChainCombo: vi.fn().mockReturnValue(0),
}));

import { useChainCombo } from '../hooks/useChainCombo';

const mockExec = vi.mocked(tripeaksApi.exec);
const mockCombo = vi.mocked(useChainCombo);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTriPeaksCard(c: Card | null, removed: boolean, exposed: boolean): TriPeaksCard {
  return { card: c, removed, exposed };
}

/** Build a minimal TriPeaks layout for testing (4 rows × 10 cols). */
function makeTestLayout(): TriPeaksCard[][] {
  const layout: TriPeaksCard[][] = [];
  for (let r = 0; r < 4; r++) {
    const row: TriPeaksCard[] = [];
    for (let c = 0; c < 10; c++) {
      row.push(makeTriPeaksCard(null, true, false));
    }
    layout.push(row);
  }
  layout[3][0] = makeTriPeaksCard(card('SPADE', 5), false, true);
  layout[3][1] = makeTriPeaksCard(card('HEART', 6), false, true);
  layout[3][2] = makeTriPeaksCard(card('CLOVER', 7), false, true);
  layout[0][0] = makeTriPeaksCard(card('DIAMOND', 10), false, false);
  layout[0][3] = makeTriPeaksCard(card('SPADE', 11), false, false);
  layout[0][6] = makeTriPeaksCard(card('HEART', 12), false, false);
  return layout;
}

const playingState: TriPeaksResponse = {
  layout: makeTestLayout(),
  stockCount: 20,
  waste: [card('CLOVER', 4)],
  phase: 0,
  moveCount: 3,
  score: 0,
  combo: 0,
  canUndo: true,
  isStalemate: false,
  message: '',
};

const gameClearState: TriPeaksResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'tripeaks.gameClear',
  messageParams: { moveCount: '28' },
};

const gameOverState: TriPeaksResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'tripeaks.gameOver',
};

beforeEach(() => {
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  mockCombo.mockReturnValue(0);
  mockPlaySound.mockClear();
});

describe('TriPeaksPage playable summary', () => {
  // CUI は playableSummary / drawRecommended を毎回出していたのに、Web は
  // 個々のリングだけで合計が無かった (#4783)。
  it('shows how many cards are playable, and rings exactly that many', async () => {
    renderWithProviders(<TriPeaksPage />);
    // 捨て札トップは ♣4。露出している ♠5 だけが出せる。
    await waitFor(() => expect(screen.getByTestId('tp-playable')).toHaveTextContent('今出せるカード: 1枚'));
    expect(document.querySelectorAll('.ring-ds-success\\/70')).toHaveLength(1);
    expect(screen.getByTestId('tp-playable')).not.toHaveTextContent('ドロー推奨');
  });

  it('recommends a draw when nothing is playable and the stock still has cards', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('CLOVER', 9)] });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('tp-playable')).toHaveTextContent('今出せるカード: 0枚'));
    expect(screen.getByTestId('tp-playable')).toHaveTextContent('ドロー推奨');
  });

  // **山札が尽きていれば勧めない。**この枝を踏まないと、常に推奨する実装でも通る。
  it('does not recommend a draw once the stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('CLOVER', 9)], stockCount: 0 });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('tp-playable')).toHaveTextContent('今出せるカード: 0枚'));
    expect(screen.getByTestId('tp-playable')).not.toHaveTextContent('ドロー推奨');
  });

  it('hides the summary once the game has ended', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.queryByTestId('tp-playable')).not.toBeInTheDocument());
  });
});

describe('TriPeaksPage', () => {
  it('rings exposed cards adjacent to the waste top during play', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toBeInTheDocument());
    // Waste top is ♣4: the exposed ♠5 is playable, ♥6/♣7 are not, ♦10 is face-down.
    expect(screen.getByLabelText('♠ 5')).toHaveClass('ring-ds-success/70');
    expect(screen.getByLabelText('♥ 6')).not.toHaveClass('ring-ds-success/70');
    expect(screen.getByLabelText('♦ 10')).not.toHaveClass('ring-ds-success/70');
  });

  it('prefers the hint ring over the playable ring on the same card', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toBeInTheDocument());

    // Server hint targets ♠5 at row 3, col 0 — the same card the playable ring covers.
    mockExec.mockResolvedValueOnce({ ...playingState, hint: { type: 'remove', row: 3, col: 0 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toHaveClass('ring-ds-warning'));
    expect(screen.getByLabelText('♠ 5')).not.toHaveClass('ring-ds-success/70');
  });

  it('shows no playable ring when the waste is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [] });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toBeInTheDocument());
    expect(screen.getByLabelText('♠ 5')).not.toHaveClass('ring-ds-success/70');
  });

  it('shows no playable ring after the game ends', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toBeInTheDocument());
    expect(screen.getByLabelText('♠ 5')).not.toHaveClass('ring-ds-success/70');
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TriPeaksPage />);
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('renders stock count', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(20\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 3/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [] });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(drawButtons[drawButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking undo button dispatches undo', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('renders game clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('ゲームクリア')).toBeInTheDocument());
  });

  it('renders game over state', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
  });

  it('hides action buttons when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
  });

  it('disables undo button when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('suppresses unused import warning', () => {
    expect(actionLogApi).toBeDefined();
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('renders correctly on mobile viewport (isMobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 393 });
    try {
      renderWithProviders(<TriPeaksPage />);
      await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
      // 10-column tableau should render with effectiveCardWidth derived from viewport
      const tableauRows = document.querySelectorAll('[data-tutorial="tp-peaks"] > div');
      expect(tableauRows.length).toBe(4);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders correctly on desktop viewport (non-mobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      renderWithProviders(<TriPeaksPage />);
      await waitFor(() => expect(screen.getByText('捨て札')).toBeInTheDocument());
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3, canUndo: true });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('3');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 4, canUndo: true });
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 4));
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('does not render the combo badge when combo < 2', async () => {
    mockCombo.mockReturnValue(1);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('combo-badge')).not.toBeInTheDocument();
  });

  // **コンボは色とテキストだけで示されていた。** Golf と同じ欠落 (#5821 / #5520)。
  it('announces a running combo to screen readers', async () => {
    mockCombo.mockReturnValue(3);
    renderWithProviders(<TriPeaksPage />);
    const live = await screen.findByTestId('tripeaks-combo-announce');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveTextContent('3');
  });

  it('stays silent while there is no combo', async () => {
    mockCombo.mockReturnValue(1);
    renderWithProviders(<TriPeaksPage />);
    const live = await screen.findByTestId('tripeaks-combo-announce');
    // 領域は常設し、中身だけ空にする（出し入れすると読み上げが飛ぶ）。
    expect(live).toBeEmptyDOMElement();
  });

  it('renders the combo badge with blue styling when combo is 2', async () => {
    mockCombo.mockReturnValue(2);
    renderWithProviders(<TriPeaksPage />);
    const badge = await screen.findByTestId('combo-badge');
    expect(badge.className).toContain('bg-ds-info');
  });

  it('renders the combo badge with warning styling when combo is between 3 and 4', async () => {
    mockCombo.mockReturnValue(3);
    renderWithProviders(<TriPeaksPage />);
    const badge = await screen.findByTestId('combo-badge');
    expect(badge.className).toContain('bg-ds-warning');
  });

  it('renders the combo badge with error styling when combo >= 5', async () => {
    mockCombo.mockReturnValue(5);
    renderWithProviders(<TriPeaksPage />);
    const badge = await screen.findByTestId('combo-badge');
    expect(badge.className).toContain('bg-ds-error');
  });

  // --- Per-peak remaining indicator (issue #3085) ---

  it('renders the per-peak remaining-card indicator during play', async () => {
    renderWithProviders(<TriPeaksPage />);
    const indicator = await screen.findByTestId('peak-remaining');
    // makeTestLayout: left peak 4 cards, middle 1, right 1.
    expect(indicator.textContent).toMatch(/4\/1\/1/);
  });

  it('shows a check mark for a peak whose remaining count is zero', async () => {
    const layout = makeTestLayout();
    layout[0][6] = makeTriPeaksCard(null, true, false); // clear the right peak
    mockExec.mockResolvedValue({ ...playingState, layout });
    renderWithProviders(<TriPeaksPage />);
    const indicator = await screen.findByTestId('peak-remaining');
    expect(indicator).toHaveTextContent('✓');
  });

  it('hides the peak indicator after the game ends', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByText('ゲームクリア')).toBeInTheDocument());
    expect(screen.queryByTestId('peak-remaining')).not.toBeInTheDocument();
  });

  it('plays a sound once when a peak is cleared', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('peak-remaining')).toBeInTheDocument());
    mockPlaySound.mockClear();

    // Next server response clears the right peak (its only card is removed).
    const cleared = makeTestLayout();
    cleared[0][6] = makeTriPeaksCard(null, true, false);
    mockExec.mockResolvedValue({ ...playingState, layout: cleared });
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(drawButtons[drawButtons.length - 1]);

    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
    expect(mockPlaySound.mock.calls.filter((c) => c[0] === 'cardPlace')).toHaveLength(1);
  });
});

describe('TriPeaksPage chain-bonus score & best record (#3087)', () => {
  it('displays the score, starting at 0', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('tp-score')).toBeInTheDocument());
    expect(screen.getByTestId('tp-score')).toHaveTextContent('0');
  });

  it('adds chain-bonus points after removing a card', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toBeInTheDocument());
    // **得点はサーバが数える。** ページは state.score を映すだけ (#5511)。
    // 連鎖倍率とピークボーナスの式そのものは domain のテストが押さえている。
    mockExec.mockResolvedValueOnce({ ...playingState, moveCount: 4, score: 100, combo: 1 });
    fireEvent.click(screen.getByLabelText('♠ 5'));
    await waitFor(() => expect(screen.getByTestId('tp-score')).toHaveTextContent('100'));
  });

  it('shows a previously stored best score in the record panel', async () => {
    localStorage.setItem(TRIPEAKS_STATS_KEY, JSON.stringify({ plays: 5, wins: 2, bestScore: 1200 }));
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('tp-stats-panel')).toBeInTheDocument());
    expect(screen.getByTestId('tp-stats-panel')).toHaveTextContent('1200');
    expect(screen.getByTestId('tp-stats-panel')).toHaveTextContent('2/5');
  });

  it('records a new best score on game clear and persists it', async () => {
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 5')).toBeInTheDocument());
    // The clearing move both scores 100 and ends the game.
    mockExec.mockResolvedValueOnce({
      ...playingState,
      moveCount: 4,
      score: 100,
      combo: 1,
      phase: 1,
      messageCode: 'tripeaks.gameClear',
      messageParams: { moveCount: '4' },
    });
    fireEvent.click(screen.getByLabelText('♠ 5'));
    await waitFor(() => expect(screen.getByTestId('tp-best-badge')).toBeInTheDocument());
    expect(JSON.parse(localStorage.getItem(TRIPEAKS_STATS_KEY) ?? '{}')).toEqual({
      plays: 1,
      wins: 1,
      bestScore: 100,
    });
    expect(screen.getByTestId('tp-stats-panel')).toHaveTextContent('100');
  });
});

describe('computePeakRemaining', () => {
  it('returns [0, 0, 0] for an undefined layout', () => {
    expect(computePeakRemaining(undefined)).toEqual([0, 0, 0]);
  });

  it('counts present, non-removed cards per peak by column range', () => {
    expect(computePeakRemaining(makeTestLayout())).toEqual([4, 1, 1]);
  });

  it('ignores removed cells and empty positions', () => {
    const layout = makeTestLayout();
    layout[3][0] = makeTriPeaksCard(card('SPADE', 5), true, true); // removed (left peak)
    layout[0][3] = makeTriPeaksCard(null, false, false); // empty (middle peak)
    expect(computePeakRemaining(layout)).toEqual([3, 0, 1]);
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('TriPeaksPage keyboard shortcuts', () => {
  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    // give-up is irreversible, so the key must route through the dialog (#2099)
    // instead of dispatching straight away.
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<TriPeaksPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
