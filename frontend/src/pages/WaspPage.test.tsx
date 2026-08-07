import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { waspApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, WaspResponse } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { WaspPage } from './WaspPage';

vi.mock('../api/gameApi', () => ({
  waspApi: { exec: vi.fn() },
  actionLogApi: { wasp: vi.fn() },
}));

const mockExec = vi.mocked(waspApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: WaspResponse = {
  tableau: [
    [
      { card: null, faceUp: false },
      { card: card('SPADE', 13), faceUp: true },
    ],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 8), faceUp: true },
    ],
    [{ card: card('CLOVER', 5), faceUp: true }],
    [{ card: card('DIAMOND', 10), faceUp: true }],
    [{ card: card('SPADE', 3), faceUp: true }],
    [{ card: card('HEART', 7), faceUp: true }],
    [{ card: card('CLOVER', 2), faceUp: true }],
  ],
  stockCount: 3,
  completedSuits: 0,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'wasp.playing',
};

const gameClearState: WaspResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  completedSuits: 4,
  messageCode: 'wasp.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: WaspResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'wasp.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  // Reset localStorage so state from previous tests (e.g. CLI mode flag) doesn't leak
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('WaspPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count and stock', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
    expect(screen.getByText(/Stock|ストック/)).toBeInTheDocument();
  });

  it('renders a face-up slot without a card as an empty-labelled button (defensive guard)', async () => {
    // A face-up card with no card object should not crash; the aria-label guard
    // falls back to an empty string.
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: [[{ card: null, faceUp: true }], [], [], [], [], [], []],
    });
    const { container } = renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(container.querySelector('button[aria-label=""]')).toBeInTheDocument();
  });

  it("adds a 'selected' hint to the aria-label of the picked card and removes it on deselect", async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: /選択中/ })).not.toBeInTheDocument();
    // Select the top card of a column.
    fireEvent.click(screen.getByRole('button', { name: '♠ K' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ K 選択中' })).toBeInTheDocument());
    // Re-clicking deselects and restores the plain label.
    fireEvent.click(screen.getByRole('button', { name: '♠ K 選択中' }));
    await waitFor(() => expect(screen.queryByRole('button', { name: /選択中/ })).not.toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♠ K' })).toBeInTheDocument();
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('deal button triggers deal command', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '配る' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('deal button disabled when stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('clicking deal with empty columns triggers shake on empty placeholders and skips API', async () => {
    const stateWithEmptyCol: WaspResponse = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [], // empty
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(stateWithEmptyCol);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getByTestId('sc-empty-col-1')).toBeInTheDocument());
    expect(screen.getByTestId('sc-empty-col-1').className).not.toContain('animate-shake');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));

    await waitFor(() => {
      expect(screen.getByTestId('sc-empty-col-1').className).toContain('animate-shake');
    });
    expect(mockExec).not.toHaveBeenCalledWith('deal');
  });

  it('labels empty columns with the "any card" rule sublabel', async () => {
    const stateWithEmptyCol: WaspResponse = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [], // empty
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(stateWithEmptyCol);
    renderWithProviders(<WaspPage />);
    const emptyCol = await screen.findByTestId('sc-empty-col-1');
    expect(emptyCol).toHaveTextContent('任意カード可');
    expect(emptyCol).toHaveAttribute('aria-label', expect.stringContaining('任意カード可'));
  });

  it('shows a persistent success ring on empty columns while a source is selected (touch-friendly)', async () => {
    const state = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [], // empty column
        [],
        [],
        [],
        [],
        [],
      ],
    };
    mockExec.mockResolvedValue(state);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const emptyCol = screen.getByTestId('sc-empty-col-1');
    // No source selected yet → no persistent ring.
    expect(emptyCol.className).not.toContain('ring-ds-success');
    // Select ♠K as the move source (no hover event fired).
    fireEvent.click(screen.getByRole('button', { name: '♠ K' }));
    await waitFor(() => expect(screen.getByTestId('sc-empty-col-1').className).toContain('ring-ds-success'));
    // The highlight is persistent, not hover-gated.
    expect(screen.getByTestId('sc-empty-col-1').className).not.toContain('hover:ring');
  });

  it('deal button exposes empty-column reason via title when blocked', async () => {
    const stateWithEmptyCol: WaspResponse = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [],
        [{ card: card('CLOVER', 5), faceUp: true }],
        [{ card: card('DIAMOND', 10), faceUp: true }],
        [{ card: card('SPADE', 3), faceUp: true }],
        [{ card: card('HEART', 7), faceUp: true }],
        [{ card: card('CLOVER', 2), faceUp: true }],
      ],
    };
    mockExec.mockResolvedValue(stateWithEmptyCol);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '配る' })).toHaveAttribute('title', '空の列をすべて埋めないと配れません');
  });

  it('autocomplete button triggers autocomplete command', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('renders empty tableau column placeholder', async () => {
    const stateWithEmpty = {
      ...playingState,
      tableau: [[], ...playingState.tableau.slice(1)],
    };
    mockExec.mockResolvedValue(stateWithEmpty);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('re-clicking the selected last card deselects without firing a move', async () => {
    // Regression: a selected last card (isLast) must deselect on re-click,
    // not route into handleSelectTarget and fire a doomed same-column move.
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // SPADE 13 is the last (face-up) card of column 0.
    const lastCardBtn = screen.getByRole('button', { name: cardAlt(card('SPADE', 13)) });
    mockExec.mockClear();
    fireEvent.click(lastCardBtn); // select
    fireEvent.click(lastCardBtn); // re-click → deselect, no API call
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('hint button triggers hint command via apiCall', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('reset button opens confirmation dialog', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
  });

  it('confirming reset issues another reset API call', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getAllByRole('button', { name: /reset|リセット/i })[0]);
    const confirmBtn = await screen.findByRole('button', { name: '確認' });
    fireEvent.click(confirmBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('selecting a tableau card highlights it, clicking again deselects', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // col 2 has a single ♣5; its aria-label renders via cardAlt.
    const [cardBtn] = screen.getAllByRole('button', { name: /5/ });
    // First click selects
    cardBtn.click();
    // Second click deselects (no move issued because both clicks land on the
    // same cardIndex, which is also the last card, so we route through
    // handleSelectSource→deselect)
    cardBtn.click();
    // API should still only have been called with reset (no move dispatched)
    await waitFor(() => expect(mockExec).toHaveBeenCalledTimes(1));
  });

  it('shows StalemateEscapeButton when isStalemate is true', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      canUndo: true,
      undoToEscape: 2,
    });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument();
  });

  it('renders inline hint for tableau move when state.hint is set', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 1, toCol: 3 },
      messageCode: 'wasp.hintAvailable',
    });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Inline hint box shows "場札 3"
    const hintBox = screen.getByRole('status');
    expect(hintBox.textContent).toMatch(/3/);
  });

  it('renders inline hint for deal when hint.fromCol is -1', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: -1, cardIndex: -1, toCol: -1 },
      messageCode: 'wasp.hintAvailable',
    });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintBox = screen.getByRole('status');
    // Deal label is '配る' (from wasp.json)
    expect(hintBox.textContent).toMatch(/配る/);
  });

  it('shows action log button in ended phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /棋譜|action log|アクション/i })).toBeInTheDocument();
  });

  it('keyboard shortcut "g" opens the give up confirm dialog in PLAYING phase', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('keyboard shortcut "d" triggers deal', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('keyboard shortcut "h" triggers hint', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('keyboard shortcut "z" triggers undo', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard shortcut "a" triggers autocomplete', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('keyboard shortcuts are disabled after game ends', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    fireEvent.keyDown(document, { key: 'd' });
    // No additional API calls
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('selecting a source card then clicking another card dispatches a move', async () => {
    // Col 0 has ♠K; col 1 has ♠Q. Click K (last in col 0 → selects as source).
    // Click Q (last in col 1 → handleSelectTarget fires → apiCall('move', ...)).
    const state = {
      ...playingState,
      tableau: [
        [{ card: card('SPADE', 13), faceUp: true }],
        [{ card: card('SPADE', 12), faceUp: true }],
        [],
        [],
        [],
        [],
        [],
      ],
    };
    mockExec.mockResolvedValue(state);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const kBtn = screen.getByRole('button', { name: '♠ K' });
    mockExec.mockClear();
    fireEvent.click(kBtn);
    await waitFor(() => expect(kBtn.className).toMatch(/ring-/));
    const qBtn = screen.getByRole('button', { name: '♠ Q' });
    fireEvent.click(qBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('clicking the same selected card deselects it (same-cell click path)', async () => {
    // Col 2 has a single ♣5, col 1 has ♥8. Click ♥8 (last, selects it),
    // then click ♣5 — since ♥8 was selected and ♣5 is not in col 1,
    // we route through handleSelectSource (switching selection) not handleSelectTarget.
    // Then click the same ♣5 again → deselect (isLast=true, selectedSource===current → noop+deselect).
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const heart8 = screen.getByRole('button', { name: '♥ 8' });
    fireEvent.click(heart8);
    await waitFor(() => expect(heart8.className).toMatch(/ring-/));
    // Clicking again deselects (no API call)
    fireEvent.click(heart8);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('clicking empty column with source selected sends move to that column', async () => {
    const state = {
      ...playingState,
      tableau: [[], playingState.tableau[1], [], [], [], [], []],
    };
    mockExec.mockResolvedValue(state);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Select ♥8 from col 1
    const heart8 = screen.getByRole('button', { name: '♥ 8' });
    mockExec.mockClear();
    fireEvent.click(heart8);
    await waitFor(() => expect(heart8.className).toMatch(/ring-/));
    // Empty-column placeholder button
    const emptyPlaceholders = screen.getAllByRole('button', { name: /空 場札/ });
    fireEvent.click(emptyPlaceholders[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('CLI toggle enables terminal mode', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // CliToggle renders a toggle button; clicking it flips cliEnabled → CliTerminal mounts
    const cliToggle = screen.getByRole('button', { name: /CLI|GUI/i });
    fireEvent.click(cliToggle);
    // CliTerminal mounts and renders the command input
    await waitFor(() => {
      expect(screen.getByLabelText(/コマンドを入力/)).toBeInTheDocument();
    });
  });

  it('renders WinCelebration on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // WinCelebration shows a congratulations banner; assert via text/role
    expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0);
  });

  it('retry button on error state triggers reset via retry callback', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<WaspPage />);
    const retryBtn = await screen.findByRole('button', { name: /retry|再試行|やり直/i });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(retryBtn);
    // retry re-issues the last failed call (reset)
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('toggles frontend hint checkbox', async () => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintToggle = screen.getByRole('checkbox', { name: /ヒント表示/ });
    expect(hintToggle).not.toBeChecked();
    fireEvent.click(hintToggle);
    expect(hintToggle).toBeChecked();
  });

  it('does not light the source card when the hint was not requested', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // The backend attaches its latest hint to every Output(); only an explicit
    // request sets the messageCode, so nothing may glow without one (#4791).
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 1, toCol: 3 },
      messageCode: 'wasp.playing',
    });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(document.querySelectorAll('.ring-ds-info').length).toBe(0);
  });

  it('lights the source card once the hint is requested', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 1, toCol: 3 },
      messageCode: 'wasp.hintAvailable',
    });
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(document.querySelectorAll('.ring-ds-info').length).toBeGreaterThan(0));
  });
});

// **ネイティブ CUI にある legal コマンドが CLI モードでは Unknown command に
// なっていた (#4792)。**手元の状態を読むだけなので API は呼ばない。
describe('WaspPage CLI legal command', () => {
  const runCli = async (command: string) => {
    renderWithProviders(<WaspPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /CLI|GUI/i }));
    const input = await screen.findByLabelText(/コマンドを入力/);
    mockExec.mockClear();
    fireEvent.change(input, { target: { value: command } });
    fireEvent.keyDown(input, { key: 'Enter' });
  };

  it('answers legal without calling the API', async () => {
    await runCli('legal 0');
    await waitFor(() => expect(screen.getAllByText(/column 0/i).length).toBeGreaterThan(0));
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('rejects a column that is out of range', async () => {
    await runCli('legal 99');
    await waitFor(() => expect(screen.getAllByText(/Usage: legal/).length).toBeGreaterThan(0));
  });

  // **空き列も移動先として並べる。**Wasp では空き列はどのカードでも受け入れる。
  // ページのハイライト用 waspLegalTargets は空き列を除くので、それをそのまま
  // 流用すると実際には打てる手を「無い」と答えてしまう。
  it('lists empty columns as legal targets', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: [[{ card: card('SPADE', 5), faceUp: true }], [], [], [], [], [], []],
    });
    await runCli('legal 0');
    await waitFor(() => expect(screen.getAllByText(/empty/).length).toBeGreaterThan(0));
  });

  // **legal を足しても他のコマンドは変わらない。**
  it('still sends other commands to the API', async () => {
    await runCli('d');
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });
});
