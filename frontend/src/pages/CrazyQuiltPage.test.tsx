import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyquiltApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CrazyQuiltResponse } from '../types/card';
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
