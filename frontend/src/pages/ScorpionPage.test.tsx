import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { scorpionApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ScorpionResponse } from '../types/card';
import { ScorpionPage } from './ScorpionPage';

vi.mock('../api/gameApi', () => ({
  scorpionApi: { exec: vi.fn() },
  actionLogApi: { scorpion: vi.fn() },
}));

const mockExec = vi.mocked(scorpionApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: ScorpionResponse = {
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
  messageCode: 'scorpion.playing',
};

const gameClearState: ScorpionResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  completedSuits: 4,
  messageCode: 'scorpion.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: ScorpionResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'scorpion.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  // Reset localStorage so state from previous tests (e.g. CLI mode flag) doesn't leak
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('ScorpionPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it("adds a 'selected' hint to the aria-label of the picked source card", async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Base labels carry no selection hint.
    expect(screen.getByRole('button', { name: '♠ K' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /選択中/ })).not.toBeInTheDocument();
    // Selecting the top card of a column marks it selected in its label.
    fireEvent.click(screen.getByRole('button', { name: '♠ K' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ K 選択中' })).toBeInTheDocument());
    // Other cards keep their plain, hint-free label.
    expect(screen.getByRole('button', { name: '♥ 8' })).toBeInTheDocument();
  });

  it('shows move count and stock', async () => {
    renderWithProviders(<ScorpionPage />);
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
    const { container } = renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(container.querySelector('button[aria-label=""]')).toBeInTheDocument();
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('deal button triggers deal command', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '配る' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('deal button disabled when stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('clicking deal with empty columns triggers shake on empty placeholders and skips API', async () => {
    const stateWithEmptyCol: ScorpionResponse = {
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
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getByTestId('sc-empty-col-1')).toBeInTheDocument());
    expect(screen.getByTestId('sc-empty-col-1').className).not.toContain('animate-shake');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配る' }));

    await waitFor(() => {
      expect(screen.getByTestId('sc-empty-col-1').className).toContain('animate-shake');
    });
    expect(mockExec).not.toHaveBeenCalledWith('deal');
  });

  it('deal button exposes empty-column reason via title when blocked', async () => {
    const stateWithEmptyCol: ScorpionResponse = {
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
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '配る' })).toHaveAttribute('title', '空の列をすべて埋めないと配れません');
  });

  it('shows the deal-blocked reason as visible text when an empty column exists', async () => {
    const stateWithEmptyCol: ScorpionResponse = {
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
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配る' })).toBeInTheDocument());
    const reason = screen.getByTestId('sc-deal-blocked-reason');
    expect(reason).toHaveTextContent('空の列をすべて埋めないと配れません');
    expect(reason).toHaveAttribute('role', 'status');
  });

  it('does not show the deal-blocked reason when every column is filled', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('sc-deal-blocked-reason')).not.toBeInTheDocument();
  });

  it('autocomplete button triggers autocomplete command', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: '自動完成' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /next game|次のゲーム/i }).length).toBeGreaterThan(0);
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('renders empty tableau column placeholder', async () => {
    const stateWithEmpty = {
      ...playingState,
      tableau: [[], ...playingState.tableau.slice(1)],
    };
    mockExec.mockResolvedValue(stateWithEmpty);
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('hint button triggers hint command via apiCall', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('reset button opens confirmation dialog', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
  });

  it('confirming reset issues another reset API call', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getAllByRole('button', { name: /reset|リセット/i })[0]);
    const confirmBtn = await screen.findByRole('button', { name: '確認' });
    fireEvent.click(confirmBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('selecting a tableau card highlights it, clicking again deselects', async () => {
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument();
  });

  it('renders inline hint for tableau move when state.hint is set', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 1, toCol: 3 },
      messageCode: 'scorpion.hintAvailable',
    });
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Inline hint box shows "場札 3"
    const hintBox = screen.getByRole('status');
    expect(hintBox.textContent).toMatch(/3/);
  });

  it('renders inline hint for deal when hint.fromCol is -1', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: -1, cardIndex: -1, toCol: -1 },
      messageCode: 'scorpion.hintAvailable',
    });
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintBox = screen.getByRole('status');
    // Deal label is '配る' (from scorpion.json)
    expect(hintBox.textContent).toMatch(/配る/);
  });

  it('shows action log button in ended phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /棋譜|action log|アクション/i })).toBeInTheDocument();
  });

  it('keyboard shortcut "g" opens the give up confirm dialog in PLAYING phase', async () => {
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('keyboard shortcut "h" triggers hint', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('keyboard shortcut "z" triggers undo', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard shortcut "a" triggers autocomplete', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('keyboard shortcuts are disabled after game ends', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
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
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // WinCelebration shows a congratulations banner; assert via text/role
    expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0);
  });

  it('retry button on error state triggers reset via retry callback', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<ScorpionPage />);
    const retryBtn = await screen.findByRole('button', { name: /retry|再試行|やり直/i });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(retryBtn);
    // retry re-issues the last failed call (reset)
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('toggles frontend hint checkbox', async () => {
    renderWithProviders(<ScorpionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintToggle = screen.getByRole('checkbox', { name: /ヒント表示/ });
    expect(hintToggle).not.toBeChecked();
    fireEvent.click(hintToggle);
    expect(hintToggle).toBeChecked();
  });
});
