import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackholeApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackHoleResponse, Card } from '../types/card';
import { BlackHolePage } from './BlackHolePage';

vi.mock('../api/gameApi', () => ({
  blackholeApi: { exec: vi.fn() },
  actionLogApi: { blackhole: vi.fn() },
}));

const mockExec = vi.mocked(blackholeApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<BlackHoleResponse> = {}): BlackHoleResponse {
  const fans: Card[][] = Array.from({ length: 17 }, () => []);
  fans[0] = [card('HEART', 9), card('CLOVER', 6)];
  fans[1] = [card('DIAMOND', 10)];
  return {
    fans,
    blackHole: [card('SPADE', 1)],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('BlackHolePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BlackHolePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BlackHolePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // #5681: 勝利条件は52枚すべてを吸い込むこと。17個の扇を掘る長いゲームなのに、
  // あと何枚で終わるかがどこにも出ていなかった。
  it('shows how many of the 52 cards have been swallowed', async () => {
    renderWithProviders(<BlackHolePage />);

    const progress = await screen.findByTestId('bh-progress');
    // 既定の盤面は穴に1枚。
    expect(progress).toHaveTextContent('1');
    expect(progress).toHaveTextContent('52');
  });

  it('updates as the hole fills', async () => {
    mockExec.mockResolvedValue(makeState({ blackHole: [card('SPADE', 1), card('SPADE', 2), card('SPADE', 3)] }));
    renderWithProviders(<BlackHolePage />);

    expect(await screen.findByTestId('bh-progress')).toHaveTextContent('3');
  });

  it('renders the 17 fans', async () => {
    renderWithProviders(<BlackHolePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.getByTestId('fan-16')).toBeInTheDocument();
  });

  it('plays a fan top into the black hole', async () => {
    renderWithProviders(<BlackHolePage />);
    const top = await screen.findByTestId('card-0-1');
    mockExec.mockClear();
    fireEvent.click(top);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mb', { fan: 0 }));
  });

  it('does not play a non-top card', async () => {
    renderWithProviders(<BlackHolePage />);
    const buried = await screen.findByTestId('card-0-0');
    mockExec.mockClear();
    fireEvent.click(buried);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('undoes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<BlackHolePage />);
    await screen.findByTestId('hint-button');
    for (const [testid, cmd] of [
      ['undo-button', 'u'],
      ['hint-button', 'hint'],
      ['giveup-button', 'g'],
    ] as const) {
      mockExec.mockClear();
      fireEvent.click(screen.getByTestId(testid));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(cmd));
    }
  });

  it('hides controls at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<BlackHolePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-button')).not.toBeInTheDocument();
  });

  it('rings the ±1 fan tops on hint, leaving non-adjacent fans unmarked', async () => {
    // Hole top is 7 → fan0 top (CLOVER 6) is adjacent; fan1 top (DIAMOND 10) is not.
    const adj = makeState({ blackHole: [card('SPADE', 7)] });
    mockExec.mockResolvedValue(adj);
    renderWithProviders(<BlackHolePage />);
    const legalTop = await screen.findByTestId('card-0-1');
    const otherTop = screen.getByTestId('card-1-0');
    // No highlight until the hint is requested.
    expect(legalTop).not.toHaveAttribute('data-hinted-legal');

    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('card-0-1')).toHaveAttribute('data-hinted-legal', 'true'));
    expect(screen.getByTestId('card-0-1').className).toContain('ring-ds-success');
    // The non-adjacent fan top is never marked.
    expect(otherTop).not.toHaveAttribute('data-hinted-legal');
    // With no backend recommendation, it falls back to the plain legal highlight.
    expect(screen.getByTestId('card-0-1')).not.toHaveAttribute('data-hinted-recommended');
    // Non-visual channels: aria-label gains "置けます" and the live region lists it.
    expect(screen.getByTestId('card-0-1')).toHaveAttribute('aria-label', '♣ 6（ファン1） · 置けます');
    expect(screen.getByTestId('bh-hint-announce')).toHaveTextContent('置けるカード: ♣ 6（ファン1）');
  });

  it('strongly emphasises the backend-recommended fan, distinct from other legal fans', async () => {
    // Hole top 7 → fan0 top ♣6 and fan2 top ♠8 are both legal; the backend
    // recommends fan 2.
    const state = makeState({ blackHole: [card('SPADE', 7)], hint: { fan: 2 } });
    state.fans[2] = [card('SPADE', 8)];
    mockExec.mockResolvedValue(state);
    renderWithProviders(<BlackHolePage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));

    // The recommended fan carries both the legal ring and the distinct gold outline.
    const recommended = await screen.findByTestId('card-2-0');
    await waitFor(() => expect(recommended).toHaveAttribute('data-hinted-recommended', 'true'));
    expect(recommended).toHaveAttribute('data-hinted-legal', 'true');
    expect(recommended.className).toContain('outline-ds-warning');
    expect(recommended).toHaveAttribute('aria-label', '♠ 8（ファン3） · おすすめ · 置けます');

    // The other legal fan is highlighted but NOT recommended (two-tier).
    const otherLegal = screen.getByTestId('card-0-1');
    expect(otherLegal).toHaveAttribute('data-hinted-legal', 'true');
    expect(otherLegal).not.toHaveAttribute('data-hinted-recommended');
    expect(otherLegal.className).not.toContain('outline-ds-warning');

    // The live region leads with the recommendation.
    expect(screen.getByTestId('bh-hint-announce')).toHaveTextContent('おすすめ: ♠ 8（ファン3）');
  });

  it('labels fan cards and the black hole for screen readers', async () => {
    mockExec.mockResolvedValue(makeState({ blackHole: [card('DIAMOND', 6)] }));
    renderWithProviders(<BlackHolePage />);
    // fan0 top ♣6, hole ♦6.
    expect(await screen.findByRole('button', { name: '♣ 6（ファン1）' })).toBeInTheDocument();
    expect(screen.getByTestId('bh-hole-top')).toHaveAttribute('aria-label', 'ブラックホール: ♦ 6');
  });

  it('clears the hint highlight on reset even though moveCount stays 0', async () => {
    const adj = makeState({ blackHole: [card('SPADE', 7)] });
    mockExec.mockResolvedValue(adj);
    renderWithProviders(<BlackHolePage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('card-0-1')).toHaveAttribute('data-hinted-legal', 'true'));
    // Reset (with confirm) — moveCount stays 0, so the highlight must be cleared explicitly.
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByTestId('card-0-1')).not.toHaveAttribute('data-hinted-legal'));
  });

  it('always shows the acceptable ranks (hole top ±1) and the legal-move count', async () => {
    // Hole top 7 → accepts 6 and 8. fan0 top ♣6 is legal, fan1 top ♦10 is not.
    mockExec.mockResolvedValue(makeState({ blackHole: [card('SPADE', 7)] }));
    renderWithProviders(<BlackHolePage />);
    const readout = await screen.findByTestId('bh-acceptable');
    expect(readout).toHaveTextContent('受け入れ可能: 6 / 8');
    const count = screen.getByTestId('bh-legal-count');
    expect(count).toHaveTextContent('合法手: 1');
    expect(count.className).toContain('text-ds-text-muted');
  });

  it('shows only one side of the acceptable ranks at the A and K ends', async () => {
    // Hole top A(1): no wrap, so only rank 2 is acceptable.
    mockExec.mockResolvedValue(makeState({ blackHole: [card('SPADE', 1)] }));
    const { unmount } = renderWithProviders(<BlackHolePage />);
    expect(await screen.findByTestId('bh-acceptable')).toHaveTextContent('受け入れ可能: 2');
    unmount();

    // Hole top K(13): only rank Q(12) is acceptable.
    mockExec.mockResolvedValue(makeState({ blackHole: [card('SPADE', 13)] }));
    renderWithProviders(<BlackHolePage />);
    expect(await screen.findByTestId('bh-acceptable')).toHaveTextContent('受け入れ可能: Q');
  });

  it('marks the legal-move count with a warning colour when no move remains', async () => {
    // Hole top 7, but neither fan top is ±1 (fan0 ♣6 replaced by ♦Q, fan1 ♦10).
    const stuck = makeState({ blackHole: [card('SPADE', 7)] });
    stuck.fans[0] = [card('DIAMOND', 12)];
    mockExec.mockResolvedValue(stuck);
    renderWithProviders(<BlackHolePage />);
    const count = await screen.findByTestId('bh-legal-count');
    expect(count).toHaveTextContent('合法手: 0');
    expect(count.className).toContain('text-ds-warning');
  });

  it('does not ring an A-K wrap (no wrap in Black Hole)', async () => {
    // Hole top A(1): only rank 2 is adjacent — K(13) must NOT highlight.
    const wrap = makeState({ blackHole: [card('SPADE', 1)] });
    wrap.fans[0] = [card('HEART', 13)]; // King
    wrap.fans[1] = [card('CLOVER', 2)]; // Two — the only legal move
    mockExec.mockResolvedValue(wrap);
    renderWithProviders(<BlackHolePage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('card-1-0')).toHaveAttribute('data-hinted-legal', 'true'));
    expect(screen.getByTestId('card-0-0')).not.toHaveAttribute('data-hinted-legal');
  });
});
