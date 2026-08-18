import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gapsApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, GapsResponse } from '../types/card';
import { GapsPage } from './GapsPage';

vi.mock('../api/gameApi', () => ({
  gapsApi: { exec: vi.fn() },
  actionLogApi: { gaps: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockedRun = vi.mocked(gapsApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeGrid(): (Card | null)[][] {
  const suits: CardDesign[] = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];
  return suits.map((s) => {
    const row: (Card | null)[] = [];
    for (let c = 0; c < 12; c++) row.push(card(s, c + 2));
    row.push(null);
    return row;
  });
}

const playingState: GapsResponse = {
  grid: makeGrid(),
  redealsUsed: 0,
  redealsRemaining: 3,
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: GapsResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'gaps.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: GapsResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'gaps.gameOver',
};

const withHintState: GapsResponse = {
  ...playingState,
  hint: { fromRow: 1, fromCol: 0, toRow: 0, toCol: 12 },
  // **頼んだヒントの応答**であることを示す。#4483 以降 Output() も hint を
  // 載せるので、これが無いとページは押していない状態と区別できない (#4605)。
  messageCode: 'gaps.hintAvailable',
};

beforeEach(() => {
  localStorage.clear();
  mockedRun.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('GapsPage', () => {
  it('renders skeleton when state is null', () => {
    mockedRun.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GapsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders move count in playing state', async () => {
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('reflects the gap ghost hint (needed / anySuit / blocked) in each cell aria-label', async () => {
    const grid: (Card | null)[][] = [
      Array(13).fill(null), // row 0: col 0 gap → any suit 2
      [card('SPADE', 3), ...Array(12).fill(null)], // row 1: col 1 gap → needs ♠4
      [card('HEART', 13), ...Array(12).fill(null)], // row 2: col 1 gap → blocked (follows K)
      Array(13).fill(null),
    ];
    mockedRun.mockResolvedValue({ ...playingState, grid });
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.getByTestId('gaps-cell-0-0')).toBeInTheDocument());
    expect(screen.getByTestId('gaps-cell-0-0')).toHaveAttribute('aria-label', '空き（任意のスートの 2 を置けます）');
    expect(screen.getByTestId('gaps-cell-1-1')).toHaveAttribute('aria-label', '空き（♠ 4 を置けます）');
    expect(screen.getByTestId('gaps-cell-2-1')).toHaveAttribute('aria-label', '空き（K の後ろのため使用不可）');
  });

  it('renders redeals remaining', async () => {
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/再配り残り: 3/));
  });

  it('calls run reset on mount', async () => {
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('reset'));
  });

  it('disables redeal button when no redeals remaining', async () => {
    mockedRun.mockResolvedValue({ ...playingState, redealsRemaining: 0, redealsUsed: 3 });
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByRole('button', { name: /再配り/ });
    expect(btn).toBeDisabled();
  });

  it('calls run redeal when redeal clicked', async () => {
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByRole('button', { name: /再配り/ });
    fireEvent.click(btn);
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('redeal'));
  });

  it('shows how many locked cards a redeal will keep', async () => {
    // makeGrid: 4 rows each a full 2..13 run → 4 × 12 = 48 locked cards.
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByTestId('gaps-redeal-button');
    expect(btn).toHaveTextContent('(0/3)');
    expect(btn).toHaveTextContent('48');
  });

  it('calls run undo when undo clicked', async () => {
    mockedRun.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByRole('button', { name: '元に戻す' });
    fireEvent.click(btn);
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('undo'));
  });

  // #5609: ドメインも presenter も手詰まりを返しており、CUI は #4800 で赤い警告を
  // 出すようになったのに、Web 版は何も出さないままだった。動かせる札が見つからず
  // 困るだけで、脱出手段があることに気づけない。
  it('offers the stalemate escape when the server reports one', async () => {
    mockedRun.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3, canUndo: true });
    renderWithProviders(<GapsPage />);

    const btn = await screen.findByTestId('stalemate-escape-button');
    mockedRun.mockClear();
    fireEvent.click(btn);
    // undo_n の 4 番目の引数が戻す手数 (gapsApi.exec(command, from, to, n))。
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('shows no escape button while the game is not stuck', async () => {
    mockedRun.mockResolvedValue({ ...playingState, isStalemate: false, undoToEscape: 3 });
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('calls run hint when hint clicked', async () => {
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByRole('button', { name: 'ヒント' });
    fireEvent.click(btn);
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('hint'));
  });

  it('clicking give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByRole('button', { name: 'ギブアップ' });
    fireEvent.click(btn);
    await flushPendingDispatch();
    expect(mockedRun).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('giveup'));
  });

  it('highlights hint source and target cells with rings instead of coordinate text', async () => {
    mockedRun.mockResolvedValue(withHintState); // from (1,0) card, to (0,12) gap
    renderWithProviders(<GapsPage />);
    const btn = await screen.findByRole('button', { name: 'ヒント' });
    fireEvent.click(btn);

    // Target gap cell gets the hint ring.
    await waitFor(() => expect(screen.getByTestId('gaps-cell-0-12')).toHaveClass('ring-ds-warning'));
    // Source card cell gets the hint ring (locked in this fixture; the hint ring wins over the locked ring).
    expect(screen.getByTestId('gaps-locked-1-0')).toHaveClass('ring-ds-warning');
    // The old coordinate text line is gone.
    expect(screen.queryByText(/\(1,0\)/)).not.toBeInTheDocument();
  });

  it('hides action buttons when game is in game-clear phase', async () => {
    mockedRun.mockResolvedValue(gameClearState);
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument());
  });

  it('hides action buttons when game is over', async () => {
    mockedRun.mockResolvedValue(gameOverState);
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('renders a lock badge for each card that forms the row prefix starting from 2', async () => {
    // makeGrid seeds every row with 2..13 of the same suit ⇒ the first 12 cells of each row are locked.
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByTestId('gaps-locked-0-0')).toBeInTheDocument();
    expect(screen.getByTestId('gaps-locked-0-11')).toBeInTheDocument();
  });

  it('advertises keyboard shortcuts on the action buttons via aria-keyshortcuts and a kbd chip', async () => {
    mockedRun.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<GapsPage />);
    const undoBtn = await screen.findByRole('button', { name: '元に戻す' });
    expect(undoBtn).toHaveAttribute('aria-keyshortcuts', 'z');
    expect(undoBtn.querySelector('kbd')).toHaveTextContent('Z');
    expect(screen.getByTestId('gaps-redeal-button')).toHaveAttribute('aria-keyshortcuts', 'd');
    expect(screen.getByRole('button', { name: 'ヒント' })).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toHaveAttribute('aria-keyshortcuts', 'g');
  });

  it('fires undo when the z key is pressed (only while canUndo)', async () => {
    mockedRun.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<GapsPage />);
    await screen.findByRole('button', { name: '元に戻す' });
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('undo'));
  });

  it('does not fire undo via the z key when canUndo is false', async () => {
    renderWithProviders(<GapsPage />);
    await screen.findByRole('button', { name: '元に戻す' });
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'z' });
    await flushPendingDispatch();
    expect(mockedRun).not.toHaveBeenCalledWith('undo');
  });

  it('fires redeal when the d key is pressed', async () => {
    renderWithProviders(<GapsPage />);
    await screen.findByTestId('gaps-redeal-button');
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('redeal'));
  });

  it('does not fire redeal via the d key when no redeals remain', async () => {
    mockedRun.mockResolvedValue({ ...playingState, redealsRemaining: 0, redealsUsed: 3 });
    renderWithProviders(<GapsPage />);
    await screen.findByTestId('gaps-redeal-button');
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'd' });
    await flushPendingDispatch();
    expect(mockedRun).not.toHaveBeenCalledWith('redeal');
  });

  it('fires hint when the h key is pressed', async () => {
    renderWithProviders(<GapsPage />);
    await screen.findByRole('button', { name: 'ヒント' });
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'h' });
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('hint'));
  });

  it('routes the g key through the give-up confirm dialog', async () => {
    renderWithProviders(<GapsPage />);
    await screen.findByRole('button', { name: 'ギブアップ' });
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'g' });
    await flushPendingDispatch();
    expect(mockedRun).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockedRun).toHaveBeenCalledWith('giveup'));
  });

  it('does not fire shortcuts once the game has ended', async () => {
    mockedRun.mockResolvedValue(gameClearState);
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument());
    mockedRun.mockClear();
    fireEvent.keyDown(document.body, { key: 'h' });
    await flushPendingDispatch();
    expect(mockedRun).not.toHaveBeenCalledWith('hint');
  });

  it('does not lock a row whose leftmost card is not a 2', async () => {
    const grid = makeGrid();
    grid[0] = [card('HEART', 5), card('HEART', 6), ...Array.from({ length: 11 }, () => null)];
    mockedRun.mockResolvedValue({ ...playingState, grid });
    renderWithProviders(<GapsPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('gaps-locked-0-0')).not.toBeInTheDocument();
  });
});
