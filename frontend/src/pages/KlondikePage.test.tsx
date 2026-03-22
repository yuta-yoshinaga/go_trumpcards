import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, klondikeApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, KlondikeResponse, KlondikeTableauCard } from '../types/card';
import { KlondikePage } from './KlondikePage';

vi.mock('../api/gameApi', () => ({
  klondikeApi: { exec: vi.fn() },
  actionLogApi: { klondike: vi.fn() },
}));

const mockExec = vi.mocked(klondikeApi.exec);

function makeTableau(cols: KlondikeTableauCard[][]): KlondikeTableauCard[][] {
  const result: KlondikeTableauCard[][] = [];
  for (let i = 0; i < 7; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: KlondikeResponse = {
  tableau: makeTableau([
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 5), faceUp: true },
    ],
    [],
    [],
    [],
    [],
    [],
  ]),
  stockCount: 20,
  waste: [card('CLOVER', 3)],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 5,
  drawCount: 1,
  canUndo: false,
  isStalemate: false,
  score: -52,
  scoringMode: 0,
  message: '',
};

const playingNoWasteState: KlondikeResponse = {
  ...playingState,
  waste: [],
};

const playingEmptyStockState: KlondikeResponse = {
  ...playingState,
  stockCount: 0,
};

const gameClearState: KlondikeResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'klondike.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: KlondikeResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'klondike.gameOver',
};

const withFoundationState: KlondikeResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []],
};

const withHintState: KlondikeResponse = {
  ...playingState,
  hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

const withHintTableauState: KlondikeResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: -1 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
});

describe('KlondikePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KlondikePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());
    expect(screen.getByText(/\(20\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Waste has a card, so there should be an img
    const wasteImages = screen.getAllByRole('img');
    expect(wasteImages.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('空')).toBeInTheDocument());
  });

  it('renders empty stock as button', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Empty stock shows "引く" button
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    expect(drawButtons.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    // Foundation with cards shows card images
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty foundation placeholder with A', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    // Empty foundation shows "A" placeholder
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('empty foundation buttons have aria-label announcing empty slot with suit', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    for (const suit of ['♠', '♣', '♥', '♦']) {
      expect(screen.getByRole('button', { name: `空の組札 (${suit})` })).toBeInTheDocument();
    }
  });

  it('foundation with cards has aria-label with card count', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    // Pile 0: 1 card (♠ A), pile 2: 2 cards (♥ A, ♥ 2), piles 1 and 3 empty
    expect(screen.getByRole('button', { name: '♠ 組札 1枚' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♥ 組札 2枚' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (♣)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (♦)' })).toBeInTheDocument();
  });

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('renders tableau columns', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());
    // Column labels 0-6
    for (let i = 0; i < 7; i++) {
      expect(screen.getByText(i.toString())).toBeInTheDocument();
    }
  });

  it('renders empty tableau column with K placeholder', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());
    // Empty columns show "K" placeholder
    const kElements = screen.getAllByText('K');
    expect(kElements.length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // Footer draw button
    const footerDrawBtn = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(footerDrawBtn[footerDrawBtn.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('clicking auto complete button dispatches autocomplete', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('clicking give up button dispatches giveup', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking reset button dispatches reset', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('clicking waste card selects it as source', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Click waste card - find the button with the waste card image (alt="♣ 3")
    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    // Should show ring highlight after re-render
    await waitFor(() => expect(wasteButton.className).toContain('ring-2'));
  });

  it('waste card button has aria-pressed false initially', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    expect(wasteButton).toHaveAttribute('aria-pressed', 'false');
  });

  it('waste card button has aria-label with card name', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    expect(wasteButton).toHaveAttribute('aria-label', '♣ 3');
  });

  it('waste card button has aria-pressed true when selected', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('clicking waste card when source already selected does nothing', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Select a tableau card (King of Spades) as source first
    const tableauImg = screen.getByAltText('♠ K');
    const tableauButton = tableauImg.closest('button') as HTMLButtonElement;
    fireEvent.click(tableauButton);
    await waitFor(() => expect(tableauButton.className).toContain('ring-2'));

    // Click waste card while tableau source is selected - should return early (no source change)
    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);

    // Tableau card should still be the selected source (ring-2 still present)
    expect(tableauButton.className).toContain('ring-2');
  });

  it('clicking tableau face-up card selects as source', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    // Find face-up card in tableau by alt text (K of Spades)
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('clicking tableau card when source selected dispatches move', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Select waste as source
    const wasteArea = screen.getByText('ウェイスト').closest('.text-center');
    const wasteButton = wasteArea?.querySelector('button');
    if (wasteButton) {
      fireEvent.click(wasteButton);
    }

    // Click a tableau target
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);

    // Click an empty column (K placeholder)
    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    if (kButtons.length > 0) {
      fireEvent.click(kButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('clicking foundation when source selected dispatches move', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Select waste as source
    const wasteArea = screen.getByText('ウェイスト').closest('.text-center');
    const wasteButton = wasteArea?.querySelector('button');
    if (wasteButton) {
      fireEvent.click(wasteButton);
    }

    // Click empty foundation (A placeholder)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    if (aButtons.length > 0) {
      fireEvent.click(aButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('foundation with cards is clickable when source selected', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    // Select waste as source first
    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton.className).toContain('ring-2'));

    // Click foundation card (Ace of Spades in pile 0)
    mockExec.mockClear();
    mockExec.mockResolvedValue(withFoundationState);
    const foundationImg = screen.getByAltText('♠ A');
    const foundationButton = foundationImg.closest('button') as HTMLButtonElement;
    fireEvent.click(foundationButton);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
  });

  it('foundation disabled when no source selected', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    // Empty foundation buttons should be disabled when no source selected
    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());

    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    for (const btn of kButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('shows hint text from waste after clicking hint', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // When hint button is clicked, klondikeApi.exec('hint') is called directly
    // and the hook sets the hint state from the response
    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
    // Hint text contains both "ヒントがあります" and "ウェイスト" in the same element
    expect(screen.getByText(/ヒントがあります/).textContent).toContain('ウェイスト');
  });

  it('shows hint text from tableau after clicking hint', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintTableauState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
    expect(screen.getByText(/場札 0/)).toBeInTheDocument();
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.klondike);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→tableau' }],
    });

    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    // Draw, hint, autocomplete, giveup buttons should not be visible
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'プレイ中',
      messageCode: 'klondike.playing',
    });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getAllByText('プレイ中').length).toBeGreaterThanOrEqual(1));
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('displays error message', async () => {
    renderWithProviders(<KlondikePage />);
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
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText(/山札/)).toBeInTheDocument());

    // Stock has cards, so CardBack should be rendered and clickable
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // The stock card back has an aria-label of "引く"
    const drawLabel = screen.getByLabelText('引く');
    if (drawLabel) {
      fireEvent.click(drawLabel);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    }
  });

  it('game clear hides playing-only footer buttons', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    // Footer should not have draw/hint/autocomplete/giveup
    const footerButtons = screen.getAllByRole('button').filter((btn) => btn.closest('.shrink-0.border-t'));
    const buttonNames = footerButtons.map((btn) => btn.textContent);
    expect(buttonNames).not.toContain('ヒント');
    expect(buttonNames).not.toContain('自動完成');
    expect(buttonNames).not.toContain('ギブアップ');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameClearState);
    vi.mocked(actionLogApi.klondike).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→tableau' }],
    });

    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Keyboard navigation tests ---

  it('pressing d triggers draw in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('pressing h triggers hint in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 'h' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  // --- Feature 1: 3-card draw mode ---

  it('renders draw mode selector', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());
    expect(screen.getByText('1枚引き')).toBeInTheDocument();
    expect(screen.getByText('3枚引き')).toBeInTheDocument();
  });

  it('changing draw mode resets game with config', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, drawCount: 3 });
    fireEvent.change(screen.getByLabelText('ドローモード'), { target: { value: '3' } });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 3, scoringMode: 0 }),
    );
  });

  it('in 3-card mode shows up to 3 waste cards fanned', async () => {
    const threeCardState: KlondikeResponse = {
      ...playingState,
      drawCount: 3,
      waste: [card('SPADE', 2), card('HEART', 3), card('CLOVER', 4)],
    };
    mockExec.mockResolvedValue(threeCardState);
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // All 3 cards should be visible as images
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(3);
  });

  // --- Feature 2: Undo ---

  it('renders undo button when playing', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  it('clicking undo dispatches undo command', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('pressing z triggers undo in PLAYING phase', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Feature 3: Vegas scoring & timer ---

  it('renders scoring mode selector', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByLabelText('スコアモード')).toBeInTheDocument());
    expect(screen.getByText('なし')).toBeInTheDocument();
    expect(screen.getByText('ベガス')).toBeInTheDocument();
  });

  it('shows score when Vegas mode is active', async () => {
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 1, score: -42 });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText(/スコア: -42/)).toBeInTheDocument());
  });

  it('does not show score when scoring mode is None', async () => {
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 0 });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText(/手数: 5/)).toBeInTheDocument());
    expect(screen.queryByText(/スコア:/)).not.toBeInTheDocument();
  });

  it('renders timer', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText(/タイム: 00:00/)).toBeInTheDocument());
  });

  it('changing scoring mode resets game with config', async () => {
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByLabelText('スコアモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 1 });
    fireEvent.change(screen.getByLabelText('スコアモード'), { target: { value: '1' } });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 1, scoringMode: 1 }),
    );
  });

  it('shows total score on game clear in Vegas mode', async () => {
    mockExec.mockResolvedValue({ ...gameClearState, scoringMode: 1, score: 208 });
    renderWithProviders(<KlondikePage />);
    await waitFor(() => expect(screen.getByText(/合計スコア:/)).toBeInTheDocument());
  });
});
