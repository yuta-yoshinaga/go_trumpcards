import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyquiltApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CrazyQuiltResponse } from '../types/card';
import { CrazyQuiltPhase } from '../types/phases';
import { CrazyQuiltPage } from './CrazyQuiltPage';

vi.mock('../api/gameApi', () => ({
  crazyquiltApi: { exec: vi.fn() },
  actionLogApi: { crazyquilt: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(crazyquiltApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

// A full quilt with cell 5 already taken, and only the first eight cells
// takeable -- enough to exercise both the available and the boxed-in paths.
function makeQuilt(): (Card | null)[] {
  const quilt = Array.from({ length: 64 }, (_, i) => card('SPADE', (i % 13) + 1));
  const cells: (Card | null)[] = [...quilt];
  cells[5] = null;
  return cells;
}

const playingState: CrazyQuiltResponse = {
  quilt: makeQuilt(),
  available: Array.from({ length: 64 }, (_, i) => i < 8 && i !== 5),
  foundationAscending: [true, true, true, true, false, false, false, false],
  redealsLeft: 1,
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 32,
  waste: [],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

describe('CrazyQuiltPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    mockExec.mockResolvedValue(playingState);
  });

  it('resets on mount', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders all sixty-four quilt cells', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toBeInTheDocument());
    expect(screen.getByTestId('cq-cell-63')).toBeInTheDocument();
    expect(screen.queryByTestId('cq-cell-64')).not.toBeInTheDocument();
  });

  it('shows foundation progress and updates it after a card moves there', async () => {
    const initialFoundation = Array.from({ length: 8 }, () => [] as Card[]);
    initialFoundation[0] = [card('SPADE', 1)];
    const movedFoundation = initialFoundation.map((pile) => [...pile]);
    movedFoundation[0].push(card('SPADE', 2));
    mockExec.mockResolvedValueOnce({ ...playingState, foundation: initialFoundation });
    mockExec.mockResolvedValueOnce({ ...playingState, foundation: movedFoundation });

    renderWithProviders(<CrazyQuiltPage />);
    const progress = await screen.findByRole('progressbar', { name: '組札の完成度' });
    expect(progress).toHaveAttribute('aria-valuemin', '0');
    expect(progress).toHaveAttribute('aria-valuemax', '104');
    expect(progress).toHaveAttribute('aria-valuenow', '1');
    expect(progress).toHaveTextContent('組札 1/104枚');
    expect(progress).toHaveAttribute('aria-valuetext', '組札 1/104枚');

    fireEvent.click(screen.getByTestId('cq-cell-0'));
    fireEvent.click(screen.getAllByRole('button', { name: /組札0/ })[0]);
    await waitFor(() => expect(progress).toHaveAttribute('aria-valuenow', '2'));
    expect(progress).toHaveTextContent('組札 2/104枚');
  });

  it('localizes foundation progress in English', async () => {
    await i18n.changeLanguage('en');
    const { unmount } = renderWithProviders(<CrazyQuiltPage />);
    const progress = await screen.findByRole('progressbar', { name: 'Foundation progress' });
    expect(progress).toHaveTextContent('0/104 cards on the foundations');
    expect(progress).toHaveAttribute('aria-valuetext', '0/104 cards on the foundations');
    unmount();
    await i18n.changeLanguage('ja');
  });

  it('includes the zero-based cell and card in quilt card names in Japanese and English', async () => {
    await i18n.changeLanguage('ja');
    const { unmount } = renderWithProviders(<CrazyQuiltPage />);
    const jaCell = await screen.findByTestId('cq-cell-0');
    expect(jaCell).toHaveAccessibleName('マス 0、♠ A');
    expect(screen.getByTestId('cq-cell-5')).toHaveTextContent('空のマス 5');

    unmount();
    await i18n.changeLanguage('en');
    renderWithProviders(<CrazyQuiltPage />);
    const enCell = await screen.findByTestId('cq-cell-0');
    expect(enCell).toHaveAccessibleName('Cell 0, ♠ A');
    expect(screen.getByTestId('cq-cell-5')).toHaveTextContent('Empty cell 5');
    await i18n.changeLanguage('ja');
  });

  it('keeps foundation targets focusable and explains that a source must be selected', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await screen.findByTestId('cq-cell-0');
    const target = screen.getAllByRole('button').find((button) => button.getAttribute('aria-disabled') === 'true');
    expect(target).toBeDefined();
    expect(target).not.toBeDisabled();
    expect(target).toHaveAttribute('aria-disabled', 'true');
    const describedBy = target?.getAttribute('aria-describedby');
    expect(describedBy).toBeTruthy();
    expect(document.getElementById(describedBy ?? '')).toHaveTextContent('先に移動する札を選んでください');
    mockExec.mockClear();
    fireEvent.click(target as HTMLElement);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('keeps every cell at the card dimensions while exposing orientation', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toBeInTheDocument());

    const vertical = screen.getByTestId('cq-cell-0');
    const horizontal = screen.getByTestId('cq-cell-1');
    expect(vertical).toHaveAttribute('data-orientation', 'vertical');
    expect(horizontal).toHaveAttribute('data-orientation', 'horizontal');
    expect(horizontal).toHaveStyle({ width: vertical.style.width, height: vertical.style.height });
    expect(horizontal.querySelector('.rotate-90')).not.toBeInTheDocument();
  });

  it('uses compact mobile quilt cards and keeps all foundations in one row', async () => {
    const originalWidth = window.innerWidth;
    window.innerWidth = 375;
    const { unmount } = renderWithProviders(<CrazyQuiltPage />);

    const cell = await screen.findByTestId('cq-cell-0');
    const foundations = document.querySelector('[data-tutorial="cg-foundation"]');
    expect(cell).toHaveStyle({ width: '24px', height: '36px' });
    expect(foundations).toHaveClass('flex-nowrap');
    unmount();

    window.innerWidth = 800;
    renderWithProviders(<CrazyQuiltPage />);
    expect(await screen.findByTestId('cq-cell-0')).toHaveStyle({ width: '60px', height: '84px' });

    window.innerWidth = originalWidth;
    window.dispatchEvent(new Event('resize'));
  });

  it('shows effective orientation corner shapes on both cards and empty cells', async () => {
    const quilt = makeQuilt();
    quilt[0] = null;
    mockExec.mockResolvedValue({ ...playingState, quilt });
    renderWithProviders(<CrazyQuiltPage />);

    const emptyVertical = await screen.findByTestId('cq-cell-0');
    const horizontalCard = screen.getByTestId('cq-cell-1');
    const emptyHorizontal = screen.getByTestId('cq-cell-5');

    expect(emptyVertical).toHaveAttribute('data-orientation', 'vertical');
    expect(horizontalCard).toHaveAttribute('data-orientation', 'horizontal');
    expect(emptyHorizontal).toHaveAttribute('data-orientation', 'horizontal');
    expect(emptyVertical.className).toContain('rounded-tl-lg rounded-br-lg');
    expect(horizontalCard.className).toContain('rounded-tr-lg rounded-bl-lg');
    expect(emptyHorizontal.className).toContain('rounded-tr-lg rounded-bl-lg');
    expect(emptyVertical.className).not.toContain('rounded-tr-lg rounded-bl-lg');
    expect(horizontalCard.className).not.toContain('rounded-tl-lg rounded-br-lg');
  });

  // **The rule the issue got wrong.** Availability depends on the card's
  // orientation, so the server decides it and the page only obeys.
  it('only lets an available card be picked up', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toBeEnabled());
    expect(screen.getByTestId('cq-cell-0')).toHaveAttribute('data-available', 'true');

    // Cell 20 is boxed in, so it renders but cannot be pressed.
    expect(screen.getByTestId('cq-cell-20')).toBeDisabled();
    expect(screen.getByTestId('cq-cell-20')).not.toHaveAttribute('data-available');
  });

  it('announces and highlights only legal destinations for the selected quilt card', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('CLOVER', 2)] });
    renderWithProviders(<CrazyQuiltPage />);
    const source = await screen.findByTestId('cq-cell-0');
    fireEvent.click(source);

    const waste = await screen.findByTestId('cq-waste');
    expect(waste).toHaveAttribute('data-valid-destination', 'true');
    expect(waste.className).toContain('ring-ds-success');
    expect(screen.getByTestId('cq-destinations')).toHaveTextContent('捨て札');
    expect(screen.getAllByRole('button', { name: /組札/ })[0]).not.toHaveAttribute('data-valid-destination');

    fireEvent.click(source);
    await waitFor(() => expect(screen.queryByTestId('cq-destinations')).not.toBeInTheDocument());
    expect(waste).not.toHaveAttribute('data-valid-destination');
  });

  it('announces foundation destinations for a selected waste card', async () => {
    const foundation = Array.from({ length: 8 }, () => [] as Card[]);
    foundation[2] = [card('HEART', 7)];
    mockExec.mockResolvedValue({ ...playingState, foundation, waste: [card('HEART', 8)] });
    renderWithProviders(<CrazyQuiltPage />);

    const waste = await screen.findByTestId('cq-waste');
    fireEvent.click(waste);

    expect(await screen.findByTestId('cq-destinations')).toHaveTextContent('有効な移動先: 組札2');
    expect(waste).not.toHaveAttribute('data-valid-destination');
  });

  it('announces when the selected card has no legal destination', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('HEART', 8)] });
    renderWithProviders(<CrazyQuiltPage />);

    fireEvent.click(await screen.findByTestId('cq-cell-0'));

    expect(await screen.findByTestId('cq-destinations')).toHaveTextContent('有効な移動先はありません');
  });

  it('hides the destination announcement when no source is selected', async () => {
    renderWithProviders(<CrazyQuiltPage />);

    await screen.findByTestId('cq-cell-0');
    expect(screen.queryByTestId('cq-destinations')).not.toBeInTheDocument();
  });

  // A boxed-in card must not reach the server even if something clicks it.
  it('never dispatches a move for a boxed-in card', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-20')).toBeDisabled());
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('cq-cell-20'));
    // Without this await the assertion passes whether or not a call fired (#4439).
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());

    // Negative control: an available cell DOES select.
    fireEvent.click(screen.getByTestId('cq-cell-0'));
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toHaveAttribute('aria-pressed', 'true'));
  });

  it('sends an available quilt card to a foundation', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toBeEnabled());
    fireEvent.click(screen.getByTestId('cq-cell-0'));
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getAllByRole('button', { name: /組札/ })[0]);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'quilt', col: 0 }, { zone: 'foundation', col: 0 }),
    );
  });

  // **キルト→捨て札の連番置き。**キルトを崩す主要な手なので、UI から出せなければ
  // ゲームが成立しない（レビュー指摘）。
  it('plays a quilt card onto the waste', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('HEART', 8)] });
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toBeEnabled());

    fireEvent.click(screen.getByTestId('cq-cell-0'));
    await waitFor(() => expect(screen.getByTestId('cq-cell-0')).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('cq-waste'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'quilt', col: 0 }, { zone: 'waste' }));
  });

  // 負のコントロール: 何も選んでいなければ、捨て札は移動元として振る舞う。
  it('treats the waste as a source when nothing is selected', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('HEART', 8)] });
    renderWithProviders(<CrazyQuiltPage />);
    const waste = await screen.findByTestId('cq-waste');
    mockExec.mockClear();

    fireEvent.click(waste);
    await waitFor(() => expect(waste).toHaveAttribute('aria-pressed', 'true'));
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  it('draws from the stock', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り32枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('redeals from the operation control when the stock is empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 1 });
    renderWithProviders(<CrazyQuiltPage />);
    const redeal = await screen.findByRole('button', { name: '組み直し' });
    expect(redeal).toBeEnabled();
    mockExec.mockClear();
    fireEvent.click(redeal);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('localizes the operation control redeal name in English', async () => {
    await i18n.changeLanguage('en');
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 1 });
    const { unmount } = renderWithProviders(<CrazyQuiltPage />);
    expect(await screen.findByRole('button', { name: 'Redeal' })).toBeEnabled();
    unmount();
    await i18n.changeLanguage('ja');
  });

  it('disables the operation control when the stock and redeals are empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 0 });
    renderWithProviders(<CrazyQuiltPage />);
    expect(await screen.findByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  // The stock button doubles as the redeal, so an empty stock with a redeal
  // left must stay pressable.
  it('keeps the stock pressable while a redeal remains', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 1 });
    renderWithProviders(<CrazyQuiltPage />);
    const stock = await screen.findByRole('button', { name: /山札/ });
    expect(stock).toBeEnabled();
  });

  it('disables the stock once the redeal is spent too', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 0 });
    renderWithProviders(<CrazyQuiltPage />);
    const stock = await screen.findByRole('button', { name: /山札/ });
    await waitFor(() => expect(stock).toBeDisabled());
  });

  // **残り組み直し回数は CUI にしか出ていなかった。**山札が 0 になった瞬間に
  // 打ち切りなのか、まだ一度組み直せるのかが Web では読めなかった。
  it('shows how many redeals are left while one remains', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 1 });
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('crazyquilt-redeals')).toHaveTextContent('組み直し: 残り1回'));
    // 空の山札でも「もう終わり」ではないことが読み上げでも分かる。
    expect(screen.getByRole('button', { name: '山札は空です（組み直し 残り1回）' })).toBeEnabled();
  });

  // 使い切った状態でも可視表示と読み上げが矛盾しない（受け入れ条件の3つ目）。
  it('still reads consistently once the redeal is spent', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, redealsLeft: 0 });
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('crazyquilt-redeals')).toHaveTextContent('組み直し: 残り0回'));
    expect(screen.getByRole('button', { name: '山札は空です（もう組み直せません）' })).toBeDisabled();
  });

  it('shows an emptied cell without a button', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cq-cell-5')).toBeInTheDocument());
    expect(screen.getByTestId('cq-cell-5').tagName).not.toBe('BUTTON');
  });

  const gameClearState: CrazyQuiltResponse = {
    ...playingState,
    phase: 1,
    message: 'ゲームクリア！',
    messageCode: 'crazyquilt.gameClear',
    messageParams: { moveCount: '42' },
  };

  it('hides the playing controls once the game clears', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('gives up through the confirm dialog', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await waitFor(() => expect(screen.getByRole('button', { name: '確認' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('surfaces a failed request instead of hanging on the skeleton', async () => {
    mockExec.mockRejectedValue(new Error('boom'));
    renderWithProviders(<CrazyQuiltPage />);
    // The page never reaches a board, so the skeleton stays and no cell renders.
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('cq-cell-0')).not.toBeInTheDocument();
  });

  // ゲームオーバー時は組札の達成状況サマリーを表示する。
  it('shows game over summary when game is over', async () => {
    mockExec.mockResolvedValue({ ...playingState, phase: CrazyQuiltPhase.GAME_OVER });
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('cg-gameover-summary')).toBeInTheDocument());
  });

  // プレイ中などゲームオーバー以外のときは達成状況サマリーを表示しない。
  it('hides game over summary while playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CrazyQuiltPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('cg-gameover-summary')).not.toBeInTheDocument();
  });
});

// **組札は A 始まりと K 始まりが混在する** (#5743)。向きが出ていないと
// 途中から見て次に何が要るのか読めない。CUI は ↑/↓ を出していたのに、
// Web は foundationAscending を一度も参照していなかった。
describe('CrazyQuiltPage foundation direction', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    mockExec.mockResolvedValue(playingState);
  });

  it('marks the first four foundations up and the last four down', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    // 0-3 が昇順、4-7 が降順という fixture の並びをそのまま突き合わせる。
    for (const idx of [0, 1, 2, 3]) {
      expect(await screen.findByTestId(`cq-foundation-head-${idx}`)).toHaveTextContent('↑');
    }
    for (const idx of [4, 5, 6, 7]) {
      expect(screen.getByTestId(`cq-foundation-head-${idx}`)).toHaveTextContent('↓');
    }
  });

  it('says the direction in the accessible name of an empty foundation', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    expect(await screen.findByRole('button', { name: '空の組札0（♠、A から昇順）' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '空の組札4（♠、K から降順）' })).toBeInTheDocument();
  });

  it('says the direction in the accessible name of a filled foundation', async () => {
    const foundation = Array.from({ length: 8 }, () => [] as Card[]);
    foundation[0] = [card('SPADE', 1)];
    foundation[4] = [card('SPADE', 13)];
    mockExec.mockResolvedValue({ ...playingState, foundation });
    renderWithProviders(<CrazyQuiltPage />);

    expect(await screen.findByRole('button', { name: '♠ 組札0（A から昇順）1枚' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♠ 組札4（K から降順）1枚' })).toBeInTheDocument();
  });

  it('shows K rather than A in an empty descending foundation', async () => {
    renderWithProviders(<CrazyQuiltPage />);
    const ascending = await screen.findByRole('button', { name: /空の組札0/ });
    expect(ascending).toHaveTextContent('A');
    expect(screen.getByRole('button', { name: /空の組札4/ })).toHaveTextContent('K');
  });
});
