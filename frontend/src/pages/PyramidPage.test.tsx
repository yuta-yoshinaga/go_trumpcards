import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, pyramidApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PyramidCard, PyramidResponse } from '../types/card';
import { PyramidPage } from './PyramidPage';

vi.mock('../api/gameApi', () => ({
  pyramidApi: { exec: vi.fn() },
  actionLogApi: { pyramid: vi.fn() },
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
});

describe('PyramidPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PyramidPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
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

  it('clicking give up button dispatches giveup', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));

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

  it('clicking same card twice deselects it', async () => {
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const card3Img = screen.getByAltText('♦ 3');
    const card3Button = card3Img.closest('button') as HTMLButtonElement;
    fireEvent.click(card3Button);
    await waitFor(() => expect(card3Button.className).toContain('ring-yellow-400'));

    // Click again to deselect
    fireEvent.click(card3Button);
    await waitFor(() => expect(card3Button.className).not.toContain('ring-yellow-400'));
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
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
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
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());

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
    renderWithProviders(<PyramidPage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
  });
});
