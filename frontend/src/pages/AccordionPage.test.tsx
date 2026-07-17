import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { accordionApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AccordionResponse, Card, CardDesign } from '../types/card';
import { AccordionPage } from './AccordionPage';

vi.mock('../api/gameApi', () => ({
  accordionApi: { exec: vi.fn() },
  actionLogApi: { accordion: vi.fn() },
}));

const mockExec = vi.mocked(accordionApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: AccordionResponse = {
  piles: [
    { cards: [card('SPADE', 1)], size: 1 },
    { cards: [card('HEART', 2)], size: 1 },
    { cards: [card('CLOVER', 3)], size: 1 },
    { cards: [card('DIAMOND', 4)], size: 1 },
  ],
  pileCount: 4,
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'accordion.playing',
};

const gameClearState: AccordionResponse = {
  ...playingState,
  piles: [{ cards: [card('SPADE', 1)], size: 52 }],
  pileCount: 1,
  phase: 1,
  moveCount: 51,
  messageCode: 'accordion.gameClear',
  messageParams: { moveCount: '51' },
};

const gameOverState: AccordionResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'accordion.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('AccordionPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count and pile count', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
    expect(screen.getByText(/パイル数/)).toBeInTheDocument();
  });

  it('lists the keyboard shortcuts in a collapsible panel and tags action buttons', async () => {
    renderWithProviders(<AccordionPage />);
    const panel = await screen.findByTestId('ac-kbd-shortcuts');
    // Closed by default so it stays discreet.
    expect(panel).not.toHaveAttribute('open');
    expect(screen.getByText('キーボードショートカット')).toBeInTheDocument();
    expect(screen.getByText('選択を解除')).toBeInTheDocument();
    // Action buttons advertise their single-key shortcuts to assistive tech.
    expect(screen.getByRole('button', { name: 'ヒント' })).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toHaveAttribute('aria-keyshortcuts', 'g');
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('shows a clear summary banner after game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByTestId('result-banner')).toHaveTextContent('クリア！ 51手'));
  });

  it('shows the remaining-piles summary banner after game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByTestId('result-banner')).toHaveTextContent('残り4パイル / 0手'));
  });

  it('does not show the result banner while playing', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
    expect(screen.queryByTestId('result-banner')).not.toBeInTheDocument();
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button disabled when canUndo is false', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('hovering a pile lights up legal -1/-3 merge targets', async () => {
    const hoverState: AccordionResponse = {
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
      pileCount: 4,
    };
    mockExec.mockResolvedValue(hoverState);
    renderWithProviders(<AccordionPage />);
    const pile3 = await screen.findByRole('button', { name: /3:/ });
    fireEvent.mouseEnter(pile3);
    const pile0 = screen.getByRole('button', { name: /0:/ });
    expect(pile0.dataset.hoverTarget).toBe('true');
    expect(pile0.className).toContain('ring-ds-success');
    // Adjacent pile (index 2) shares neither suit nor rank with SPADE 9.
    expect(screen.getByRole('button', { name: /2:/ }).dataset.hoverTarget).toBe('false');
  });

  it('hover ring is suppressed on hintFrom pile even when it is a legal target', async () => {
    // pile 0 (SPADE 7) is both hintFrom AND a legal hover-target for pile 3 (SPADE 9, same suit)
    const hintFromHoverState: AccordionResponse = {
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
      pileCount: 4,
      hint: { fromIdx: 0, toIdx: 3 },
    };
    mockExec.mockResolvedValue(hintFromHoverState);
    renderWithProviders(<AccordionPage />);
    const pile3 = await screen.findByRole('button', { name: /3:/ });
    fireEvent.mouseEnter(pile3);
    const pile0 = screen.getByRole('button', { name: /0:/ });
    // pile 0 is a legal target (SPADE match) but hintFrom wins; no success ring
    expect(pile0.className).not.toContain('ring-ds-success');
    // hintFrom ring is still present
    expect(pile0.className).toContain('ring-ds-info');
  });

  it('hover does not highlight targets when game is over', async () => {
    const twoSpadePiles = {
      ...gameOverState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
      pileCount: 2,
    };
    mockExec.mockResolvedValue(twoSpadePiles);
    renderWithProviders(<AccordionPage />);
    const pile1 = await screen.findByRole('button', { name: /1:/ });
    fireEvent.mouseEnter(pile1);
    // pile 0 shares suit but isPlaying=false → no hover targets
    expect(screen.getByRole('button', { name: /0:/ }).dataset.hoverTarget).toBe('false');
  });

  it('mouseleave clears the hover highlight', async () => {
    const hoverState: AccordionResponse = {
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
      pileCount: 4,
    };
    mockExec.mockResolvedValue(hoverState);
    renderWithProviders(<AccordionPage />);
    const pile3 = await screen.findByRole('button', { name: /3:/ });
    fireEvent.mouseEnter(pile3);
    fireEvent.mouseLeave(pile3);
    expect(screen.getByRole('button', { name: /0:/ }).dataset.hoverTarget).toBe('false');
  });

  it('selecting a pile persistently highlights its legal merge targets without hover (touch parity)', async () => {
    // pile 3 (SPADE 9) can merge onto pile 0 (SPADE 7) at offset 3 (suit match).
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
    });
    renderWithProviders(<AccordionPage />);
    const pile3 = await screen.findByRole('button', { name: /^3:/ });
    // Select via click only — no mouseEnter, mimicking a touch device.
    fireEvent.click(pile3);
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    await waitFor(() => expect(pile0.dataset.legalTarget).toBe('true'));
    expect(pile0.className).toContain('ring-ds-success');
    // Adjacent pile (index 2, CLOVER 3) shares neither suit nor rank with SPADE 9.
    expect(screen.getByRole('button', { name: /^2:/ }).dataset.legalTarget).toBe('false');
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('reset button opens confirmation dialog', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
  });

  it('clicking pile selects it, clicking again deselects', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    // Pile 0 ♠1
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    await waitFor(() => expect(pile0.className).toMatch(/ring-/));
    fireEvent.click(pile0);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('selecting pile 1 then clicking pile 0 (offset=1) dispatches a move', async () => {
    // Same rank 2 on pile 0 and 1 for a valid move
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [
        { cards: [card('SPADE', 2)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('DIAMOND', 4)], size: 1 },
      ],
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const pile1 = screen.getByRole('button', { name: /^1:/ });
    fireEvent.click(pile1);
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('selecting pile 3 then clicking pile 0 (offset=3) dispatches a move', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const pile3 = screen.getByRole('button', { name: /^3:/ });
    fireEvent.click(pile3);
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('selecting then clicking a pile 2 away re-selects that pile', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const pile2 = screen.getByRole('button', { name: /^2:/ });
    fireEvent.click(pile2);
    await waitFor(() => expect(pile2.className).toMatch(/ring-/));
    const pile0 = screen.getByRole('button', { name: /^0:/ });
    fireEvent.click(pile0);
    // offset=2 is invalid; pile0 becomes the new selection instead
    expect(mockExec).not.toHaveBeenCalled();
    await waitFor(() => expect(pile0.className).toMatch(/ring-/));
  });

  it('renders inline hint when state.hint is set', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromIdx: 3, toIdx: 0 },
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintBox = await screen.findByText(/パイル3 → パイル0/);
    expect(hintBox.textContent).toMatch(/3/);
    expect(hintBox.textContent).toMatch(/0/);
  });

  it('CLI toggle enables terminal mode', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const cliToggle = screen.getByRole('button', { name: /CLI|GUI/i });
    fireEvent.click(cliToggle);
    await waitFor(() => {
      expect(screen.getByLabelText(/コマンドを入力/)).toBeInTheDocument();
    });
  });

  it('shows action log button in ended phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /棋譜|action log|アクション/i })).toBeInTheDocument();
  });

  it('shows StalemateEscapeButton when isStalemate is true', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      canUndo: true,
      undoToEscape: 2,
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument();
  });

  it('StalemateEscapeButton dispatches undo_n with the escape count', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      canUndo: true,
      undoToEscape: 2,
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 2));
  });

  it('renders empty placeholder when pile has no cards', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [{ cards: [], size: 0 }, ...playingState.piles.slice(1)],
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Empty pile still renders a button with 0: empty aria-label
    expect(screen.getByRole('button', { name: /^0: empty$/ })).toBeInTheDocument();
  });

  it('reflects a selected pile’s legal merge offsets in its aria-label and live region', async () => {
    // pile 3 (SPADE 9) can merge onto pile 0 (SPADE 7) at offset 3 (suit match).
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
    });
    renderWithProviders(<AccordionPage />);
    const pile3 = await screen.findByRole('button', { name: /^3: ♠ 9/ });
    fireEvent.click(pile3);
    await waitFor(() => expect(screen.getByRole('button', { name: /3: ♠ 9 — 左3へマージ可/ })).toBeInTheDocument());
    const status = screen.getByTestId('ac-selection-status');
    expect(status).toHaveAttribute('role', 'status');
    expect(status).toHaveTextContent('パイル3を選択中。マージ可能な手が1通り');
  });

  it('lists both merge offsets joined by the localized separator when both are legal', async () => {
    // pile 3 (SPADE 9): offset 1 → pile 2 (SPADE 3, suit) AND offset 3 → pile 0 (SPADE 7, suit).
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('SPADE', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
    });
    renderWithProviders(<AccordionPage />);
    const pile3 = await screen.findByRole('button', { name: /^3: ♠ 9/ });
    fireEvent.click(pile3);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /3: ♠ 9 — 左1へマージ可、左3へマージ可/ })).toBeInTheDocument(),
    );
    expect(screen.getByTestId('ac-selection-status')).toHaveTextContent('パイル3を選択中。マージ可能な手が2通り');
  });

  it('announces when a selected pile has no legal merge', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [
        { cards: [card('SPADE', 7)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 },
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('SPADE', 9)], size: 1 },
      ],
    });
    renderWithProviders(<AccordionPage />);
    const pile1 = await screen.findByRole('button', { name: /^1: ♥ 2/ });
    fireEvent.click(pile1);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /1: ♥ 2 — マージ可能な手なし/ })).toBeInTheDocument(),
    );
    expect(screen.getByTestId('ac-selection-status')).toHaveTextContent('パイル1を選択中。マージ可能な手なし');
  });

  it('renders multi-card pile size badge', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      piles: [{ cards: [card('SPADE', 1)], size: 5 }, ...playingState.piles.slice(1)],
    });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('+4')).toBeInTheDocument();
  });

  it('CLI mode dispatches an m command via parseAccordionCommand', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /CLI|GUI/i }));
    const input = await screen.findByLabelText(/コマンドを入力/);
    mockExec.mockClear();
    fireEvent.change(input, { target: { value: 'm 3 0' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'pile', index: 3 }, { zone: 'pile', index: 0 }),
    );
  });

  it('CLI mode rejects malformed move command', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    fireEvent.click(screen.getByRole('button', { name: /CLI|GUI/i }));
    const input = await screen.findByLabelText(/コマンドを入力/);
    mockExec.mockClear();
    fireEvent.change(input, { target: { value: 'm 3' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    // No API call — parse error was rendered in the terminal log instead
    await waitFor(() => expect(mockExec).not.toHaveBeenCalled());
  });

  it('retry button re-dispatches reset after error', async () => {
    mockExec.mockRejectedValueOnce(new Error('boom'));
    renderWithProviders(<AccordionPage />);
    const retryBtn = await screen.findByRole('button', { name: /retry|再試行|やり直/i });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(retryBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('toggles frontend hint checkbox', async () => {
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const hintToggle = screen.getByRole('checkbox', { name: /ヒント表示/ });
    expect(hintToggle).not.toBeChecked();
    fireEvent.click(hintToggle);
    expect(hintToggle).toBeChecked();
  });

  it('pressing "1" merges the selected pile onto its left neighbour', async () => {
    const fourPileState: AccordionResponse = {
      ...playingState,
      piles: [
        { cards: [card('SPADE', 2)], size: 1 },
        { cards: [card('HEART', 2)], size: 1 }, // pile 1: same rank as pile 0 → 1-left merge legal
        { cards: [card('CLOVER', 3)], size: 1 },
        { cards: [card('DIAMOND', 4)], size: 1 },
      ],
    };
    mockExec.mockResolvedValue(fourPileState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // ArrowRight from no-selection lands on index 0; press once more to reach 1.
    fireEvent.keyDown(document, { key: 'ArrowRight' });
    fireEvent.keyDown(document, { key: 'ArrowRight' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(fourPileState);
    fireEvent.keyDown(document, { key: '1' });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'pile', index: 1 }, { zone: 'pile', index: 0 }),
    );
  });

  it('pressing "u" issues an undo when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'u' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AccordionPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });
});
