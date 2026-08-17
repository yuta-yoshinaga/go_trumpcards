import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, pyramidApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { PYRAMID_STATS_KEY } from '../hooks/usePyramidStats';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PyramidCard, PyramidResponse } from '../types/card';
import { PyramidPage } from './PyramidPage';

vi.mock('../api/gameApi', () => ({
  pyramidApi: { exec: vi.fn() },
  actionLogApi: { pyramid: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(pyramidApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makePyramidCard(c: Card | null, removed: boolean, exposed: boolean): PyramidCard {
  return { card: c, removed, exposed };
}

// Build a minimal 3-row pyramid for testing (rows 0-2)
function makeTestPyramid(): PyramidCard[][] {
  return [
    [makePyramidCard(card('SPADE', 13), false, false)],
    [makePyramidCard(card('HEART', 5), false, false), makePyramidCard(card('CLOVER', 8), false, false)],
    [
      makePyramidCard(card('DIAMOND', 3), false, true),
      makePyramidCard(card('SPADE', 10), false, true),
      makePyramidCard(card('HEART', 13), false, true),
    ],
  ];
}

const playingState: PyramidResponse = {
  pyramid: makeTestPyramid(),
  stockCount: 20,
  waste: [card('CLOVER', 3)],
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const playingNoWasteState: PyramidResponse = {
  ...playingState,
  waste: [],
};

const playingEmptyStockState: PyramidResponse = {
  ...playingState,
  stockCount: 0,
};

const gameClearState: PyramidResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'pyramid.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: PyramidResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'pyramid.gameOver',
};

const withHintState: PyramidResponse = {
  ...playingState,
  hint: { type: 'pair', row1: 2, col1: 0, row2: 2, col2: 1 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('PyramidPage', () => {
  it('conveys blocked / selected / pair-candidate state in the card aria-labels', async () => {
    renderWithProviders(<PyramidPage />);
    await screen.findByLabelText('♠ 10');
    // Top-row cards are covered by the row below → blocked.
    expect(screen.getByRole('button', { name: '♠ K （ブロック中）' })).toBeInTheDocument();
    // Selecting ♠10 marks it selected and makes ♦3 (sum 13) a pair candidate.
    fireEvent.click(screen.getByLabelText('♠ 10'));
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 10 （選択中）' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♦ 3 （合計13の相手）' })).toBeInTheDocument();
  });

  it('rings both cards of a pair hint on the board', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByLabelText('♦ 3')).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...playingState, hint: { type: 'pair', row1: 2, col1: 0, row2: 2, col2: 1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByLabelText('♦ 3')).toHaveClass('ring-ds-warning'));
    expect(screen.getByLabelText('♠ 10')).toHaveClass('ring-ds-warning');
    expect(screen.getByLabelText(/♥ K/)).not.toHaveClass('ring-ds-warning');
  });

  it('shows only the hint ring when a card is both hinted and a pair candidate', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByLabelText('♠ 10')).toBeInTheDocument());

    // Select ♠10 (partner value 3) so ♦3 becomes a pair candidate…
    // (its aria-label now gains the "（合計13の相手）" suffix, so match by regex).
    fireEvent.click(screen.getByLabelText('♠ 10'));
    await waitFor(() => expect(screen.getByLabelText(/♦ 3/)).toHaveClass('ring-ds-success'));

    // …then request a hint that also targets ♦3: the hint ring must win alone.
    mockExec.mockResolvedValueOnce({ ...playingState, hint: { type: 'pair', row1: 2, col1: 0, row2: 2, col2: 1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByLabelText(/♦ 3/)).toHaveClass('ring-ds-warning'));
    expect(screen.getByLabelText(/♦ 3/)).not.toHaveClass('ring-ds-success');
  });

  it('rings the king cell for a king hint and clears it on the next card click', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByLabelText(/♥ K/)).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...playingState, hint: { type: 'king', row1: 2, col1: 2, row2: -1, col2: -1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByLabelText(/♥ K/)).toHaveClass('ring-ds-warning'));

    // Any card interaction clears the hint highlight.
    fireEvent.click(screen.getByLabelText('♦ 3'));
    await waitFor(() => expect(screen.getByLabelText(/♥ K/)).not.toHaveClass('ring-ds-warning'));
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PyramidPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(20\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const wasteImages = screen.getAllByRole('img');
    expect(wasteImages.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('renders pyramid cards', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Should render card images for pyramid
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(3);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(drawButtons[drawButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking reset button dispatches reset', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('clicking exposed pyramid card selects it', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // ♦ 3 is exposed in row 2
    const cardImg = screen.getByAltText('♦ 3');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('clicking exposed King auto-removes it', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // ♥ K is exposed in row 2
    const kingImg = screen.getByAltText('♥ K');
    const kingButton = kingImg.closest('button') as HTMLButtonElement;
    fireEvent.click(kingButton);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('remove', expect.objectContaining({ zone: 'pyramid', row: 2, col: 2 })),
    );
  });

  it('clicking two exposed cards sends remove pair command', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Select ♦ 3 (value 3, row 2, col 0)
    const card3Img = screen.getByAltText('♦ 3');
    const card3Button = card3Img.closest('button') as HTMLButtonElement;
    fireEvent.click(card3Button);
    await waitFor(() => expect(card3Button.className).toContain('ring-2'));

    // Select ♠ 10 (value 10, row 2, col 1) - should trigger remove pair (3+10=13)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const card10Img = screen.getByAltText('♠ 10');
    const card10Button = card10Img.closest('button') as HTMLButtonElement;
    fireEvent.click(card10Button);

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'remove',
        expect.objectContaining({ zone: 'pyramid', row: 2, col: 0 }),
        expect.objectContaining({ zone: 'pyramid', row: 2, col: 1 }),
      ),
    );
  });

  it('flags exposed pair candidates after the first selection', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const card3 = screen.getByAltText('♦ 3').closest('button') as HTMLButtonElement;
    fireEvent.click(card3);
    const card10 = screen.getByAltText('♠ 10').closest('button') as HTMLButtonElement;
    await waitFor(() => expect(card10).toHaveAttribute('data-pair-candidate', 'true'));
    expect(card3).not.toHaveAttribute('data-pair-candidate');
  });

  it('selecting the waste card highlights matching pyramid partners', async () => {
    // playingState.waste = [♣ 3]; partner value 10 should match the exposed ♠ 10 in the pyramid.
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    fireEvent.click(wasteButton);
    const card10 = screen.getByAltText('♠ 10').closest('button') as HTMLButtonElement;
    await waitFor(() => expect(card10).toHaveAttribute('data-pair-candidate', 'true'));
  });

  it('selecting a pyramid card whose partner is the waste top flags the waste card', async () => {
    // waste top is ♣ 3 (value 3); the exposed ♠ 10 (value 10) is its partner.
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const card10 = screen.getByAltText('♠ 10').closest('button') as HTMLButtonElement;
    fireEvent.click(card10);
    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    await waitFor(() => expect(wasteButton).toHaveAttribute('data-pair-candidate', 'true'));
  });

  it('selecting a King (value 13) does not flag any pair candidates', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const kingButton = screen.getByAltText('♥ K').closest('button') as HTMLButtonElement;
    fireEvent.click(kingButton);
    // King auto-removes via handleSelectCard, so the API is called immediately.
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('remove', expect.objectContaining({ zone: 'pyramid', row: 2, col: 2 })),
    );
    // No pair-candidate attribute should appear anywhere because partnerValue stays null for Kings.
    expect(document.querySelectorAll('[data-pair-candidate="true"]')).toHaveLength(0);
  });

  it('marks an exposed King as removable-alone with a success ring and aria hint', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // ♥ K is exposed in row 2 → flagged as removable alone.
    const kingButton = screen.getByAltText('♥ K').closest('button') as HTMLButtonElement;
    expect(kingButton).toHaveAttribute('data-king-removable', 'true');
    expect(kingButton.className).toContain('ring-ds-success');
    // The always-on King ring must not reuse the pulsing pair-candidate style.
    expect(kingButton.className).not.toContain('animate-pulse');
    expect(kingButton).toHaveAttribute('aria-label', '♥ K （単独除去可能なK）');
  });

  it('does not mark a covered King or a non-King exposed card as removable-alone', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // ♠ K in row 0 is covered (exposed:false) → no marker.
    const coveredKing = screen.getByAltText('♠ K').closest('button') as HTMLButtonElement;
    expect(coveredKing).not.toHaveAttribute('data-king-removable');
    // ♦ 3 is exposed but not a King → no marker.
    const nonKing = screen.getByAltText('♦ 3').closest('button') as HTMLButtonElement;
    expect(nonKing).not.toHaveAttribute('data-king-removable');
  });

  it('marks a King on top of the waste as removable-alone', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('SPADE', 13)] });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const wasteButton = screen.getByRole('button', { name: '♠ K （単独除去可能なK）' });
    expect(wasteButton).toHaveAttribute('data-king-removable', 'true');
    expect(wasteButton.className).toContain('ring-ds-success');
  });

  it('clicking same card twice deselects it', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const card3Img = screen.getByAltText('♦ 3');
    const card3Button = card3Img.closest('button') as HTMLButtonElement;
    fireEvent.click(card3Button);
    await waitFor(() => expect(card3Button.className).toContain('ring-ds-warning'));

    // Click again to deselect
    fireEvent.click(card3Button);
    await waitFor(() => expect(card3Button.className).not.toContain('ring-ds-warning'));
  });

  it('waste card button has aria-pressed', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    expect(wasteButton).toHaveAttribute('aria-pressed', 'false');
  });

  it('clicking waste card selects it', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('shows hint text after clicking hint', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.pyramid);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'remove', detail: 'pair removed' }],
    });

    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'プレイ中',
      messageCode: 'pyramid.playing',
    });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getAllByText('プレイ中').length).toBeGreaterThanOrEqual(1));
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('displays error message', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('stock card back is clickable during playing phase', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawLabel = screen.getByLabelText('引く');
    if (drawLabel) {
      fireEvent.click(drawLabel);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    }
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  it('clicking undo dispatches undo command', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Keyboard navigation tests ---

  it('pressing d triggers draw in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('pressing h triggers hint in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 'h' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders correctly on mobile viewport (isMobile branch)', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      renderWithProviders(<PyramidPage />);
      await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders on desktop viewport without mobile min-width', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      renderWithProviders(<PyramidPage />);
      await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3, canUndo: true });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('3');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 4, canUndo: true });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 4));
  });
});

// --- Best-record persistence (#3083) ---

describe('PyramidPage best-record', () => {
  // Clear only the stats key so we don't disturb tutorial-suggest state used by other tests.
  beforeEach(() => localStorage.removeItem(PYRAMID_STATS_KEY));
  afterEach(() => localStorage.removeItem(PYRAMID_STATS_KEY));

  it('records the fewest-moves clear and shows the badge, panel, and header chip', async () => {
    // gameClearState.moveCount is 5.
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<PyramidPage />);

    await waitFor(() => expect(screen.getByTestId('py-best-badge')).toBeInTheDocument());
    expect(screen.getByTestId('py-best-badge')).toHaveTextContent('新記録');
    expect(screen.getByTestId('py-best-badge')).toHaveTextContent('5');
    expect(screen.getByTestId('py-stats-panel')).toHaveTextContent('5 手');
    expect(screen.getByTestId('py-stats-panel')).toHaveTextContent('クリア 1/1');
    expect(screen.getByTestId('py-best-moves')).toHaveTextContent('5 手');
    // Persisted for next time.
    expect(JSON.parse(localStorage.getItem(PYRAMID_STATS_KEY) ?? '{}')).toEqual({
      plays: 1,
      wins: 1,
      fewestMoves: 5,
    });
  });

  it('keeps a better existing record and shows no new-best badge for a slower clear', async () => {
    localStorage.setItem(PYRAMID_STATS_KEY, JSON.stringify({ plays: 1, wins: 1, fewestMoves: 3 }));
    mockExec.mockResolvedValue(gameClearState); // 5 moves > existing best 3
    renderWithProviders(<PyramidPage />);

    await waitFor(() => expect(screen.getByTestId('py-stats-panel')).toHaveTextContent('3 手'));
    expect(screen.queryByTestId('py-best-badge')).not.toBeInTheDocument();
    expect(screen.getByTestId('py-stats-panel')).toHaveTextContent('クリア 2/2');
  });

  it('counts a loss without recording a fewest-moves value', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PyramidPage />);

    await waitFor(() => expect(screen.getByTestId('py-stats-panel')).toHaveTextContent('クリア 0/1'));
    expect(screen.getByTestId('py-stats-panel')).toHaveTextContent('—');
    expect(screen.queryByTestId('py-best-moves')).not.toBeInTheDocument();
    expect(screen.queryByTestId('py-best-badge')).not.toBeInTheDocument();
  });
});

// #5510: Draw は山札を引き切ると二度と引けない設計なのに、空表示は「なし」としか
// 言わない。**標準のピラミッド (3回配り直し可) を知っているプレイヤーほど**、
// 手詰まりの原因を誤解する。
describe('PyramidPage no-redeal notice', () => {
  it('says there is no redeal once the stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<PyramidPage />);
    const note = await screen.findByTestId('py-no-redeal');
    expect(note.textContent).toMatch(/配り直し/);
  });

  // **山札が残っているうちは出さない。** まだ引けるのに「配り直し無し」と出ると、
  // もう引けないと読める。
  it('stays quiet while the stock still has cards', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 5 });
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('py-no-redeal')).not.toBeInTheDocument();
  });
});
