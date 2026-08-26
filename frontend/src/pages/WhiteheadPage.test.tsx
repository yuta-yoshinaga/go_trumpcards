import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, whiteheadApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, WhiteheadResponse, WhiteheadTableauCard } from '../types/card';
import { WhiteheadVegas } from '../types/phases';
import { WhiteheadPage } from './WhiteheadPage';

vi.mock('../api/gameApi', () => ({
  whiteheadApi: { exec: vi.fn() },
  actionLogApi: { whitehead: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(whiteheadApi.exec);

function makeTableau(cols: WhiteheadTableauCard[][]): WhiteheadTableauCard[][] {
  const result: WhiteheadTableauCard[][] = [];
  for (let i = 0; i < 7; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: WhiteheadResponse = {
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

const playingNoWasteState: WhiteheadResponse = {
  ...playingState,
  waste: [],
};

const playingEmptyStockState: WhiteheadResponse = {
  ...playingState,
  stockCount: 0,
};

const gameClearState: WhiteheadResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'whitehead.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: WhiteheadResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'whitehead.gameOver',
};

const withFoundationState: WhiteheadResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []],
};

const withHintState: WhiteheadResponse = {
  ...playingState,
  hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

const withHintTableauState: WhiteheadResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: -1 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('WhiteheadPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<WhiteheadPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(20\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Waste has a card, so there should be an img
    const wasteImages = screen.getAllByRole('img');
    expect(wasteImages.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('空')).toBeInTheDocument());
  });

  it('renders empty stock as button', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Empty stock shows "引く" button
    const drawButtons = screen.getAllByRole('button', { name: '引く' });
    expect(drawButtons.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    // Foundation with cards shows card images
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty foundation placeholder with A', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    // Empty foundation shows "A" placeholder
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('empty foundation buttons have aria-label announcing empty slot with suit', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    for (const suit of ['♠', '♣', '♥', '♦']) {
      expect(screen.getByRole('button', { name: `空の組札 (${suit})` })).toBeInTheDocument();
    }
  });

  it('foundation with cards has aria-label with card count', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    // Pile 0: 1 card (♠ A), pile 2: 2 cards (♥ A, ♥ 2), piles 1 and 3 empty
    expect(screen.getByRole('button', { name: '♠ 組札 1枚' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♥ 組札 2枚' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (♣)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札 (♦)' })).toBeInTheDocument();
  });

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('renders tableau without index headers', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('renders empty tableau column with K placeholder', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // Empty columns show "K" placeholder
    const kElements = screen.getAllByText('K');
    expect(kElements.length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // Footer draw button
    const footerDrawBtn = screen.getAllByRole('button', { name: '引く' });
    fireEvent.click(footerDrawBtn[footerDrawBtn.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('clicking auto complete button dispatches autocomplete when stock empty and tableau face-up', async () => {
    const readyState: WhiteheadResponse = {
      ...playingState,
      stockCount: 0,
      tableau: makeTableau([[{ card: card('SPADE', 13), faceUp: true }], [], [], [], [], [], []]),
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('auto complete button is disabled and pulses hidden when tableau has face-down cards', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const btn = screen.getByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn.className).not.toContain('animate-pulse');
  });

  it('auto complete button is disabled while stock still has cards (even if tableau all face-up)', async () => {
    const tableauFaceUpButStockRemaining: WhiteheadResponse = {
      ...playingState,
      stockCount: 5,
      tableau: makeTableau([[{ card: card('SPADE', 13), faceUp: true }], [], [], [], [], [], []]),
    };
    mockExec.mockResolvedValue(tableauFaceUpButStockRemaining);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const btn = screen.getByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
  });

  it('auto complete button pulses when stock empty and tableau all face-up', async () => {
    const readyState: WhiteheadResponse = {
      ...playingState,
      stockCount: 0,
      tableau: makeTableau([[{ card: card('SPADE', 13), faceUp: true }], [], [], [], [], [], []]),
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const btn = screen.getByTestId('autocomplete-button');
    expect(btn).not.toBeDisabled();
    expect(btn.className).toContain('animate-pulse');
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

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

  it('cancelling the give up dialog does not dispatch giveup', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
  });

  it('clicking reset button dispatches reset', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('clicking waste card selects it as source', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Click waste card - find the button with the waste card image (alt="♣ 3")
    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    // Should show ring highlight after re-render
    await waitFor(() => expect(wasteButton.className).toContain('ring-2'));
  });

  it('waste card button has aria-pressed false initially', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    expect(wasteButton).toHaveAttribute('aria-pressed', 'false');
  });

  it('waste card button has aria-label with card name', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    expect(wasteButton).toHaveAttribute('aria-label', '♣ 3');
  });

  it('waste card button has aria-pressed true when selected', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('clicking waste card when source already selected does nothing', async () => {
    renderWithProviders(<WhiteheadPage />);
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
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Find face-up card in tableau by alt text (K of Spades)
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('clicking tableau card when source selected dispatches move', async () => {
    renderWithProviders(<WhiteheadPage />);
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
    renderWithProviders(<WhiteheadPage />);
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
    renderWithProviders(<WhiteheadPage />);
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
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    // Empty foundation buttons should be disabled when no source selected
    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    for (const btn of kButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('shows hint text from waste after clicking hint', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // When hint button is clicked, whiteheadApi.exec('hint') is called directly
    // and the hook sets the hint state from the response
    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
    // The hint band shows the source card image (not just an abstract string).
    expect(screen.getByTestId('kl-hint-card')).toBeInTheDocument();
  });

  // #5955: ヒントは無言で現れていた。**空のまま先にマウントしてある**領域の中身が
  // 変わることが読み上げの条件なので、hint がある間だけ現れる内側の div ではなく、
  // 常設のラッパーがライブ領域でなければならない。
  it('announces the hint through a region that was already mounted', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    const region = screen.getByTestId('kl-hint-live');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('');

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // **同じ要素**の中身が変わる (別の要素が現れるのではない)。
    await waitFor(() => expect(region).toHaveTextContent(/ヒントがあります/));
  });

  it('shows hint text from tableau after clicking hint', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintTableauState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
    expect(screen.getByText(/場札 0/)).toBeInTheDocument();
    // The source tableau card (SPADE 13) is shown as a card image.
    expect(screen.getByTestId('kl-hint-card')).toBeInTheDocument();
  });

  it('ring-highlights the hint source (info) and destination (success) on the board', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // withHintState: waste (face-up source) → tableau column 3 (destination).
    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => {
      expect(document.querySelector('.ring-ds-info')).not.toBeNull();
      expect(document.querySelector('.ring-ds-success')).not.toBeNull();
    });
  });

  it('omits the hint card image when the source card cannot be resolved', async () => {
    // Mount with an empty waste, then surface a waste-sourced hint → no source card.
    mockExec.mockResolvedValueOnce(playingNoWasteState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue({
      ...playingNoWasteState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
    expect(screen.queryByTestId('kl-hint-card')).not.toBeInTheDocument();
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.whitehead);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→tableau' }],
    });

    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    // Draw, hint, autocomplete, giveup buttons should not be visible
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'プレイ中',
      messageCode: 'whitehead.playing',
    });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getAllByText('プレイ中').length).toBeGreaterThanOrEqual(1));
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('displays error message', async () => {
    renderWithProviders(<WhiteheadPage />);
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
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

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
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    // Footer should not have draw/hint/autocomplete/giveup
    const footerButtons = screen.getAllByRole('button').filter((btn) => btn.closest('.shrink-0.border-t'));
    const buttonNames = footerButtons.map((btn) => btn.textContent);
    expect(buttonNames).not.toContain('ヒント');
    expect(buttonNames).not.toContain('自動完成');
    expect(buttonNames).not.toContain('ギブアップ');
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameClearState);
    vi.mocked(actionLogApi.whitehead).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→tableau' }],
    });

    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Keyboard navigation tests ---

  it('pressing d triggers draw in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('pressing h triggers hint in PLAYING phase', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    fireEvent.keyDown(document, { key: 'h' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // --- Feature 1: 3-card draw mode ---

  it('renders draw mode selector with visible label', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());
    expect(screen.getByText('1枚引き')).toBeInTheDocument();
    expect(screen.getByText('3枚引き')).toBeInTheDocument();
    const select = screen.getByLabelText('ドローモード');
    expect(select).toHaveAttribute('id', 'draw-mode-select');
    const label = document.querySelector('label[for="draw-mode-select"]');
    expect(label).toBeInTheDocument();
  });

  it('changing draw mode mid-game asks for confirmation before resetting', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, drawCount: 3 });
    fireEvent.change(screen.getByLabelText('ドローモード'), { target: { value: '3' } });
    // Mid-game: no reset until the dialog is confirmed (#2179).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 3, scoringMode: 0 }),
    );
  });

  it('cancelling a draw mode change keeps the current setting', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.change(screen.getByLabelText('ドローモード'), { target: { value: '3' } });
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    // The select must revert to the still-active setting.
    expect(screen.getByLabelText('ドローモード')).toHaveValue('1');
  });

  it('changing scoring mode mid-game asks for confirmation before resetting', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('スコアモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 1 });
    fireEvent.change(screen.getByLabelText('スコアモード'), { target: { value: '1' } });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 1, scoringMode: 1 }),
    );
  });

  it('changing draw mode on a fresh deal resets without confirmation', async () => {
    mockExec.mockResolvedValue({ ...playingState, moveCount: 0 });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, drawCount: 3 });
    fireEvent.change(screen.getByLabelText('ドローモード'), { target: { value: '3' } });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 3, scoringMode: 0 }),
    );
  });

  it('cancelling a scoring mode change keeps the current setting', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('スコアモード')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.change(screen.getByLabelText('スコアモード'), { target: { value: '1' } });
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    expect(screen.getByLabelText('スコアモード')).toHaveValue('0');
  });

  it('changing scoring mode after game end resets without confirmation', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('スコアモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 1 });
    fireEvent.change(screen.getByLabelText('スコアモード'), { target: { value: '1' } });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 1, scoringMode: 1 }),
    );
  });

  it('changing draw mode after game end resets without confirmation', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('ドローモード')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, drawCount: 3 });
    fireEvent.change(screen.getByLabelText('ドローモード'), { target: { value: '3' } });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { drawCount: 3, scoringMode: 0 }),
    );
  });

  it('in 3-card mode shows up to 3 waste cards fanned', async () => {
    const threeCardState: WhiteheadResponse = {
      ...playingState,
      drawCount: 3,
      waste: [card('SPADE', 2), card('HEART', 3), card('CLOVER', 4)],
    };
    mockExec.mockResolvedValue(threeCardState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // All 3 cards should be visible as images
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(3);
  });

  // --- Feature 2: Undo ---

  it('renders undo button when playing', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  it('clicking undo dispatches undo command', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('pressing z triggers undo in PLAYING phase', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Feature 3: Vegas scoring & timer ---

  it('renders scoring mode selector', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByLabelText('スコアモード')).toBeInTheDocument());
    expect(screen.getByText('なし')).toBeInTheDocument();
    expect(screen.getByText('ベガス')).toBeInTheDocument();
  });

  it('shows score when Vegas mode is active', async () => {
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 1, score: -42 });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText(/スコア: -42/)).toBeInTheDocument());
  });

  it('does not show score when scoring mode is None', async () => {
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 0 });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText(/手数: 5/)).toBeInTheDocument());
    expect(screen.queryByText(/スコア:/)).not.toBeInTheDocument();
  });

  it('renders timer', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText(/タイム: 00:00/)).toBeInTheDocument());
  });

  it('shows total score on game clear in Vegas mode', async () => {
    mockExec.mockResolvedValue({ ...gameClearState, scoringMode: 1, score: 208 });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByText(/合計スコア:/)).toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with fluid tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="kl-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with fluid tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="kl-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'draw', reason: 'frontendHint.drawFromStock', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3, canUndo: true });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('3');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 4, canUndo: true });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, undefined, 4));
  });

  describe('drag and drop', () => {
    function buildDataTransfer() {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    }

    it('waste card button is draggable when playing', async () => {
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
      const wasteImg = screen.getByAltText('♣ 3');
      const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
      expect(wasteButton).toHaveAttribute('draggable', 'true');
    });

    it('tableau face-up card is draggable when playing', async () => {
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
      const cardImg = screen.getByAltText('♠ K');
      const cardButton = cardImg.closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging waste card to tableau column dispatches move', async () => {
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

      const wasteImg = screen.getByAltText('♣ 3');
      const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();

      fireEvent.dragStart(wasteButton, { dataTransfer });

      // Drop on an empty tableau column (K placeholder wrapped in DropZone)
      const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
      expect(kButtons.length).toBeGreaterThan(0);
      // The DropZone wraps the K button; drop event should fire on the wrapper.
      // Use the button's parent (the DropZone) as the drop target.
      const dropZone = kButtons[0].closest('div');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'waste' }),
          expect.objectContaining({ zone: 'tableau' }),
        ),
      );
    });

    it('dragging waste card to foundation dispatches move', async () => {
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

      const wasteImg = screen.getByAltText('♣ 3');
      const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();

      fireEvent.dragStart(wasteButton, { dataTransfer });

      // Drop on foundation (A placeholder) wrapped in DropZone
      const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
      const dropZone = aButtons[0].closest('div');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'waste' }),
          expect.objectContaining({ zone: 'foundation', col: 0 }),
        ),
      );
    });
  });

  describe('local statistics (#3031)', () => {
    it('renders the stats panel with a win rate readout', async () => {
      localStorage.removeItem('whitehead_stats');
      renderWithProviders(<WhiteheadPage />);
      const panel = await screen.findByTestId('kl-stats-panel');
      expect(panel).toHaveTextContent(/勝率/);
    });

    it('records a cleared game and shows the personal-best badge', async () => {
      localStorage.removeItem('whitehead_stats');
      mockExec.mockResolvedValue(gameClearState);
      renderWithProviders(<WhiteheadPage />);
      // Win recorded once on GAME_CLEAR; first clear beats the (empty) fewest-moves best.
      expect(await screen.findByTestId('kl-best-badge')).toBeInTheDocument();
      await waitFor(() => {
        const raw = localStorage.getItem('whitehead_stats');
        expect(raw).not.toBeNull();
        const stats = JSON.parse(raw ?? '{}');
        expect(stats['1:0']).toMatchObject({ plays: 1, wins: 1, fewestMoves: 5 });
      });
    });

    it('records a lost game as a play without a win', async () => {
      localStorage.removeItem('whitehead_stats');
      mockExec.mockResolvedValue(gameOverState);
      renderWithProviders(<WhiteheadPage />);
      await waitFor(() => {
        const stats = JSON.parse(localStorage.getItem('whitehead_stats') ?? '{}');
        expect(stats['1:0']).toMatchObject({ plays: 1, wins: 0, fewestMoves: null });
      });
      expect(screen.queryByTestId('kl-best-badge')).not.toBeInTheDocument();
    });
  });
});

// #5493: Vegas を選んだプレイヤーは、なぜスコアがマイナスから始まるのか・1枚あたり
// 何点かを知る手段が無かった。ヘッダーは生の数字だけで、チュートリアルにも
// スコアリングの説明が無い。
describe('WhiteheadPage Vegas formula', () => {
  it('spells out the formula while Vegas scoring is on', async () => {
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 1 });
    renderWithProviders(<WhiteheadPage />);
    const note = await screen.findByTestId('kl-vegas-formula');
    // 数値は WhiteheadVegas から補間される。文言に直接書くと定数と乖離する。
    //
    // **`toContain('5')` は無意味だった** -- 買い切りの "-52" が "5" を含むので、
    // 1枚あたりの点が抜けていても通る。1枚あたりは符号付きで、買い切りの数字の
    // 一部として現れない形で確かめる。
    expect(note.textContent).toContain(`${WhiteheadVegas.BUY_IN}`);
    expect(note.textContent).toMatch(new RegExp(`\\+${WhiteheadVegas.PER_CARD}(?![0-9])`));
  });

  // **None モードでは出さない。** ベガス方式の説明が常時出ていると、
  // スコアの付かないモードでも点が入ると読める。
  it('says nothing when scoring is off', async () => {
    mockExec.mockResolvedValue({ ...playingState, scoringMode: 0 });
    renderWithProviders(<WhiteheadPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('kl-vegas-formula')).not.toBeInTheDocument();
  });
});
