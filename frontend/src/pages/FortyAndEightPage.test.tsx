import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, fortyAndEightApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FortyAndEightResponse, FortyAndEightTableauCard } from '../types/card';
import { FortyAndEightPage } from './FortyAndEightPage';

vi.mock('../api/gameApi', () => ({
  fortyAndEightApi: { exec: vi.fn() },
  actionLogApi: { fortyandeight: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(fortyAndEightApi.exec);

function makeTableau(cols: FortyAndEightTableauCard[][]): FortyAndEightTableauCard[][] {
  const result: FortyAndEightTableauCard[][] = [];
  for (let i = 0; i < 8; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: FortyAndEightResponse = {
  tableau: makeTableau([
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: card('HEART', 5), faceUp: true },
      { card: card('CLOVER', 7), faceUp: true },
    ],
    [],
    [],
    [],
    [],
    [],
    [],
  ]),
  stockCount: 60,
  waste: [card('CLOVER', 3)],
  foundation: [[], [], [], [], [], [], [], []],
  redealUsed: false,
  canRedeal: false,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const playingNoWasteState: FortyAndEightResponse = {
  ...playingState,
  waste: [],
};

const playingEmptyStockState: FortyAndEightResponse = {
  ...playingState,
  stockCount: 0,
};

const canRedealState: FortyAndEightResponse = {
  ...playingState,
  stockCount: 0,
  canRedeal: true,
};

const gameClearState: FortyAndEightResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'fortyandeight.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: FortyAndEightResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'fortyandeight.gameOver',
};

const withFoundationState: FortyAndEightResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], [], [], [], [], []],
};

// Waste holds an Ace, so selecting it makes every empty foundation eligible (#3288).
const aceWasteState: FortyAndEightResponse = {
  ...playingState,
  waste: [card('SPADE', 1)],
  foundation: [[], [], [], [], [], [], [], []],
};

// Waste holds ♠2 and foundation pile 0 tops out at ♠A, so only pile 0 is a legal target (#3288).
const singleTargetState: FortyAndEightResponse = {
  ...playingState,
  waste: [card('SPADE', 2)],
  foundation: [[card('SPADE', 1)], [], [], [], [], [], [], []],
};

const withHintState: FortyAndEightResponse = {
  ...playingState,
  hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('FortyAndEightPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FortyAndEightPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('distinguishes the two same-suit foundation piles in their aria-labels', async () => {
    mockExec.mockResolvedValue(playingState); // all 8 foundations empty
    renderWithProviders(<FortyAndEightPage />);
    // The two spade foundations (idx 0, 1) now read as pile 1 and pile 2.
    await waitFor(() => expect(screen.getByRole('button', { name: '空の組札 ♠ 1' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '空の組札 ♠ 2' })).toBeInTheDocument();
  });

  it('numbers a filled same-suit foundation pile distinctly from its empty twin', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<FortyAndEightPage />);
    // idx 0 = ♠ foundation pile 1 with 1 card; idx 1 = the still-empty ♠ pile 2.
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ 組札 1 1枚' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '空の組札 ♠ 2' })).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(60\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const wasteImages = screen.getAllByRole('img');
    expect(wasteImages.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getAllByText('♠').length).toBeGreaterThanOrEqual(1));
    expect(screen.getAllByText('♣').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('♥').length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText('♦').length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getAllByText('♠').length).toBeGreaterThanOrEqual(1));
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty foundation placeholder with A', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getAllByText('♠').length).toBeGreaterThanOrEqual(1));
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('renders empty tableau column with empty placeholder', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyElements = screen.getAllByText('空');
    expect(emptyElements.length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawBtns = screen.getAllByRole('button', { name: '引く' });
    const drawBtn = drawBtns[drawBtns.length - 1];
    fireEvent.click(drawBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('redeal button shown and dispatches redeal when canRedeal', async () => {
    mockExec.mockResolvedValue(canRedealState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リディール' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(canRedealState);
    fireEvent.click(screen.getByRole('button', { name: 'リディール' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('redeal button hidden when canRedeal is false', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'リディール' })).not.toBeInTheDocument();
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('clicking auto complete button dispatches autocomplete when ready', async () => {
    const readyState: FortyAndEightResponse = {
      ...playingState,
      stockCount: 0,
      waste: [],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('auto complete button is disabled while stock or waste has cards', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('clicking give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking reset button dispatches reset', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('clicking waste card selects it as source', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton.className).toContain('ring-2'));
  });

  it('waste card button has aria-pressed false initially', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    expect(wasteButton).toHaveAttribute('aria-pressed', 'false');
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('clicking waste card when source already selected does nothing', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const tableauImg = screen.getByAltText('♠ K');
    const tableauButton = tableauImg.closest('button') as HTMLButtonElement;
    fireEvent.click(tableauButton);
    await waitFor(() => expect(tableauButton.className).toContain('ring-2'));

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);

    expect(tableauButton.className).toContain('ring-2');
  });

  it('highlights every empty foundation when an Ace source is selected', async () => {
    mockExec.mockResolvedValue(aceWasteState);
    const { container } = renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // No source selected yet -> no highlight.
    expect(container.querySelectorAll('[data-eligible-foundation="true"]')).toHaveLength(0);

    const wasteButton = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);

    // Ace is placeable on all 8 empty foundations.
    await waitFor(() => expect(container.querySelectorAll('[data-eligible-foundation="true"]')).toHaveLength(8));
    for (const el of container.querySelectorAll('[data-eligible-foundation="true"]')) {
      expect(el.className).toContain('ring-ds-info');
    }
  });

  it('highlights only the single legal foundation and leaves invalid piles unhighlighted', async () => {
    mockExec.mockResolvedValue(singleTargetState);
    const { container } = renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteButton = screen.getByAltText('♠ 2').closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);

    // ♠2 only fits on pile 0 (top ♠A); the 7 empty piles reject a non-Ace.
    await waitFor(() => expect(container.querySelectorAll('[data-eligible-foundation="true"]')).toHaveLength(1));
    const highlighted = container.querySelector('[data-eligible-foundation="true"]') as HTMLElement;
    expect(highlighted).toHaveAttribute('aria-label', expect.stringContaining('組札'));
    expect(highlighted.className).toContain('ring-ds-info');
  });

  it('clicking tableau card when source selected dispatches move', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteArea = screen.getByText('ウェイスト').closest('.text-center');
    const wasteButton = wasteArea?.querySelector('button');
    if (wasteButton) {
      fireEvent.click(wasteButton);
    }

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);

    const emptyButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '空');
    if (emptyButtons.length > 0) {
      fireEvent.click(emptyButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('clicking foundation when source selected dispatches move', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteArea = screen.getByText('ウェイスト').closest('.text-center');
    const wasteButton = wasteArea?.querySelector('button');
    if (wasteButton) {
      fireEvent.click(wasteButton);
    }

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    if (aButtons.length > 0) {
      fireEvent.click(aButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('foundation disabled when no source selected', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getAllByText('♠').length).toBeGreaterThanOrEqual(1));

    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('shows hint text after clicking hint', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.fortyandeight);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→tableau' }],
    });

    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'カードを移動してください',
      messageCode: 'fortyandeight.playing',
    });
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getAllByText('カードを移動してください').length).toBeGreaterThanOrEqual(1));
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('draw button disabled when stock is empty', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<FortyAndEightPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const drawBtns = screen.getAllByRole('button', { name: '引く' });
    const drawBtn = drawBtns[drawBtns.length - 1];
    expect(drawBtn).toBeDisabled();
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

    it('waste card is draggable when playing', async () => {
      renderWithProviders(<FortyAndEightPage />);
      await waitFor(() => expect(screen.getByAltText('♣ 3')).toBeInTheDocument());
      const wasteButton = screen.getByAltText('♣ 3').closest('button') as HTMLButtonElement;
      expect(wasteButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging waste card to foundation dispatches move', async () => {
      renderWithProviders(<FortyAndEightPage />);
      await waitFor(() => expect(screen.getByAltText('♣ 3')).toBeInTheDocument());

      const wasteButton = screen.getByAltText('♣ 3').closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();
      fireEvent.dragStart(wasteButton, { dataTransfer });

      const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
      expect(aButtons.length).toBeGreaterThan(0);
      const dropZone = aButtons[0].closest('div');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'waste' }),
          expect.objectContaining({ zone: 'foundation' }),
        ),
      );
    });
  });
});
