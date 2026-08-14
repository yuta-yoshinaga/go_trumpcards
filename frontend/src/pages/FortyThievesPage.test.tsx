import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, fortyThievesApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FortyThievesResponse, FortyThievesTableauCard } from '../types/card';
import { FortyThievesPage } from './FortyThievesPage';

vi.mock('../api/gameApi', () => ({
  fortyThievesApi: { exec: vi.fn() },
  actionLogApi: { fortythieves: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(fortyThievesApi.exec);

function makeTableau(cols: FortyThievesTableauCard[][]): FortyThievesTableauCard[][] {
  const result: FortyThievesTableauCard[][] = [];
  for (let i = 0; i < 10; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: FortyThievesResponse = {
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
    [],
    [],
  ]),
  stockCount: 60,
  waste: [card('CLOVER', 3)],
  foundation: [[], [], [], [], [], [], [], []],
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const playingNoWasteState: FortyThievesResponse = {
  ...playingState,
  waste: [],
};

const playingEmptyStockState: FortyThievesResponse = {
  ...playingState,
  stockCount: 0,
};

const gameClearState: FortyThievesResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'fortythieves.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: FortyThievesResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'fortythieves.gameOver',
};

const withFoundationState: FortyThievesResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], [], [], [], [], []],
};

const withHintState: FortyThievesResponse = {
  ...playingState,
  hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('FortyThievesPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FortyThievesPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders stock count', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(60\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders waste card', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    const wasteImages = screen.getAllByRole('img');
    expect(wasteImages.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty waste', async () => {
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());
    // Empty stock shows "空" placeholder
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  // **空の組札にスートを断定しない (#4786)。**宛先インデックスはサーバに届かず
  // findFoundation が置ける最初の山を選ぶので、固定ラベルは何も保証しなかった。
  it('labels empty foundations neutrally instead of asserting a suit', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getAllByText('—').length).toBe(8));
    expect(screen.queryByText('♠')).not.toBeInTheDocument();
    expect(screen.queryByText('♦')).not.toBeInTheDocument();
  });

  it('labels a populated foundation with the suit actually sitting on it', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<FortyThievesPage />);
    // 山0 は ♠A、山2 は ♥A→♥2。残り6つは空のまま。
    await waitFor(() => expect(screen.getAllByText('—').length).toBe(6));
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.queryByText('♣')).not.toBeInTheDocument();
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders empty foundation placeholder with A', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getAllByText('—').length).toBe(8));
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('renders empty tableau column with empty placeholder', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // Empty columns show "空" placeholder
    const emptyElements = screen.getAllByText('空');
    expect(emptyElements.length).toBeGreaterThanOrEqual(1);
  });

  it('clicking draw button dispatches draw', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const drawBtns = screen.getAllByRole('button', { name: '引く' });
    const drawBtn = drawBtns[drawBtns.length - 1];
    fireEvent.click(drawBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('clicking auto complete button dispatches autocomplete when ready', async () => {
    const readyState: FortyThievesResponse = {
      ...playingState,
      stockCount: 0,
      waste: [],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(readyState);
    fireEvent.click(screen.getByRole('button', { name: '自動完成' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('auto complete button is disabled while stock or waste has cards', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('clicking give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<FortyThievesPage />);
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
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('clicking waste card selects it as source', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton.className).toContain('ring-2'));
  });

  it('waste card button has aria-pressed false initially', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    expect(wasteButton).toHaveAttribute('aria-pressed', 'false');
  });

  it('waste card button has aria-label with card name', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteButton = screen.getByRole('button', { name: '♣ 3' });
    expect(wasteButton).toHaveAttribute('aria-label', '♣ 3');
  });

  it('waste card button has aria-pressed true when selected', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);
    await waitFor(() => expect(wasteButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  it('clicking waste card when source already selected does nothing', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByText('ウェイスト')).toBeInTheDocument());

    // Select a tableau card as source first
    const tableauImg = screen.getByAltText('♠ K');
    const tableauButton = tableauImg.closest('button') as HTMLButtonElement;
    fireEvent.click(tableauButton);
    await waitFor(() => expect(tableauButton.className).toContain('ring-2'));

    // Click waste card while tableau source is selected - should return early
    const wasteImg = screen.getByAltText('♣ 3');
    const wasteButton = wasteImg.closest('button') as HTMLButtonElement;
    fireEvent.click(wasteButton);

    expect(tableauButton.className).toContain('ring-2');
  });

  it('clicking tableau face-up card selects as source', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('clicking tableau card when source selected dispatches move', async () => {
    renderWithProviders(<FortyThievesPage />);
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

    // Click an empty column placeholder
    const emptyButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '空');
    if (emptyButtons.length > 0) {
      fireEvent.click(emptyButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('clicking foundation when source selected dispatches move', async () => {
    renderWithProviders(<FortyThievesPage />);
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
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

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
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getAllByText('—').length).toBe(8));

    // Empty foundation buttons should be disabled when no source selected
    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('double-clicking a foundation-playable tableau top card sends it to a foundation', async () => {
    const aceState: FortyThievesResponse = {
      ...playingState,
      tableau: makeTableau([[{ card: card('SPADE', 1), faceUp: true }]]),
      waste: [],
      foundation: [[], [], [], [], [], [], [], []],
    };
    mockExec.mockResolvedValue(aceState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('ft-tableau-top-0')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(aceState);
    fireEvent.dblClick(screen.getByTestId('ft-tableau-top-0'));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'move',
        { zone: 'tableau', col: 0, cardIndex: 0 },
        { zone: 'foundation', col: 0 },
      ),
    );
  });

  it('double-clicking a foundation-playable waste card sends it to a foundation', async () => {
    const aceWasteState: FortyThievesResponse = {
      ...playingState,
      tableau: makeTableau([]),
      waste: [card('SPADE', 1)],
      foundation: [[], [], [], [], [], [], [], []],
    };
    mockExec.mockResolvedValue(aceWasteState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('ft-waste-top')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(aceWasteState);
    fireEvent.dblClick(screen.getByTestId('ft-waste-top'));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'foundation', col: 0 }),
    );
  });

  it('double-clicking a card with no legal foundation target does nothing', async () => {
    // playingState: tableau col 0 top is ♠K and waste top is ♣3 with every
    // foundation empty, so neither card has a legal foundation move.
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('ft-waste-top')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.dblClick(screen.getByTestId('ft-tableau-top-0'));
    fireEvent.dblClick(screen.getByTestId('ft-waste-top'));

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('shows hint text after clicking hint', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('announces the hinted card and destination in a polite live region', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // withHintState: waste top ♣3 → tableau column 3.
    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('ft-hint-announcement');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('ヒント: ♣ 3をタブロー列3へ移動');
  });

  it('renders an empty hint announcement region when there is no hint', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // The region stays mounted (only its text is conditional) so AT announces the first hint.
    expect(screen.getByTestId('ft-hint-announcement')).toHaveTextContent('');
  });

  it('announces a tableau-origin hint by resolving the card at [fromCol][cardIndex]', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // playingState.tableau[1][0] is ♥5; move it to tableau column 3.
    const tableauHintState: FortyThievesResponse = {
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 3 },
    };
    mockExec.mockResolvedValue(tableauHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('ft-hint-announcement');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: ♥ 5をタブロー列3へ移動'));
  });

  it('announces with an empty card name when a waste hint has no waste card', async () => {
    // Mount with an empty waste pile so state.waste stays empty (handleHint updates
    // only the hint, not state) — a waste-origin hint then resolves to a null card.
    mockExec.mockResolvedValue(playingNoWasteState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingNoWasteState,
      hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 0 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('ft-hint-announcement');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: を組札へ移動'));
  });

  it('announces with an empty card name when a tableau hint points at a missing card', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // Column 2 is empty in playingState, so [2][0] is undefined → card resolves to null.
    const missingCardHintState: FortyThievesResponse = {
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'tableau', toCol: 3 },
    };
    mockExec.mockResolvedValue(missingCardHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('ft-hint-announcement');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: をタブロー列3へ移動'));
  });

  it('game clear shows action log button', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over shows action log button', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.fortythieves);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'waste→tableau' }],
    });

    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '自動完成' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  it('reset button always visible', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
  });

  it('displays message with messageCode', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      message: 'カードを移動してください',
      messageCode: 'fortythieves.playing',
    });
    renderWithProviders(<FortyThievesPage />);
    // **訳が引ければ messageCode が勝つ。** サーバの message はフォールバックで、
    // 引けなかったときにだけ出る (#5291)。
    await waitFor(() => expect(screen.getAllByText('プレイ中').length).toBeGreaterThanOrEqual(1));
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('draw button disabled when stock is empty', async () => {
    mockExec.mockResolvedValue(playingEmptyStockState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const drawBtns = screen.getAllByRole('button', { name: '引く' });
    const drawBtn = drawBtns[drawBtns.length - 1];
    expect(drawBtn).toBeDisabled();
  });

  it('ring-highlights the hinted source card after requesting a hint', async () => {
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // withHintState: waste top ♣3 → tableau column 3.
    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => {
      const wasteBtn = screen.getByRole('button', { name: '♣ 3' });
      expect(wasteBtn.className).toContain('ring-ds-info');
    });
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
      renderWithProviders(<FortyThievesPage />);
      await waitFor(() => expect(screen.getByAltText('♣ 3')).toBeInTheDocument());
      const wasteButton = screen.getByAltText('♣ 3').closest('button') as HTMLButtonElement;
      expect(wasteButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging waste card to foundation dispatches move', async () => {
      renderWithProviders(<FortyThievesPage />);
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

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('FortyThievesPage keyboard shortcuts', () => {
  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FortyThievesPage />);
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
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FortyThievesPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
