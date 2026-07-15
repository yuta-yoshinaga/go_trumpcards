import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackholeApi } from '../api/gameApi';
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
    // Non-visual channels: aria-label gains "置けます" and the live region lists it.
    expect(screen.getByTestId('card-0-1')).toHaveAttribute('aria-label', '♣ 6（ファン1） · 置けます');
    expect(screen.getByTestId('bh-hint-announce')).toHaveTextContent('置けるカード: ♣ 6（ファン1）');
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
