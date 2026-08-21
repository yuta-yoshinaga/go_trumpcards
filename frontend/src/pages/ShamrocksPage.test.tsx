import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { shamrocksApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ShamrocksResponse } from '../types/card';
import { ShamrocksPage } from './ShamrocksPage';

vi.mock('../api/gameApi', () => ({
  shamrocksApi: { exec: vi.fn() },
  actionLogApi: { shamrocks: vi.fn() },
}));

const mockExec = vi.mocked(shamrocksApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<ShamrocksResponse> = {}): ShamrocksResponse {
  return {
    fans: [[card('SPADE', 9)], [card('SPADE', 8)], [card('DIAMOND', 1)]],
    foundation: [[], [], [], []],
    redealsLeft: 3,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('ShamrocksPage', () => {
  // **リングは 1 つだけ。** ring-* は同じ box-shadow を共有するので重ねられず、
  // 連結すると選択中かつ移動可能な扇で選択リングが黙って消える (レビュー指摘)。
  it('keeps the selection ring on a fan that can also move', async () => {
    renderWithProviders(<ShamrocksPage />);

    // 既定の盤面では扇 1 (♠8) が動かせる。
    const fan = await screen.findByTestId('fan-1');
    expect(fan).toHaveAttribute('data-movable', 'true');
    fireEvent.click(fan);

    const selected = screen.getByTestId('fan-1');
    expect(selected.className).toContain('ring-2');
    expect(selected.className).toContain('ring-ds-warning');
    // 移動可能の細いリングは出さない (2 つは重ならない)。
    expect(selected.className).not.toContain('ring-1');
  });
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<ShamrocksPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the stuck redeal banner when no legal move remains', async () => {
    // No Aces, no foundation builds, no same-suit stacks -> stuck.
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 5)], [card('HEART', 9)], [card('CLOVER', 2)]],
        foundation: [[], [], [], []],
        redealsLeft: 3,
      }),
    );
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('ll-stuck-banner')).toBeInTheDocument());
  });

  it('hides the stuck banner when a legal move exists', async () => {
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('ll-stuck-banner')).not.toBeInTheDocument();
  });

  it('shows the deadlock banner and pulses give up when redeals are exhausted with no move', async () => {
    // No legal move AND no redeals left -> true deadlock.
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 5)], [card('HEART', 9)], [card('CLOVER', 2)]],
        foundation: [[], [], [], []],
        redealsLeft: 0,
      }),
    );
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('ll-deadlock-banner')).toBeInTheDocument());
    // The redeal-recommendation banner must not show when redeals are gone.
    expect(screen.queryByTestId('ll-stuck-banner')).not.toBeInTheDocument();
    expect(screen.getByTestId('giveup-button').className).toContain('animate-pulse');
  });

  it('hides the deadlock banner when a legal move exists even with no redeals', async () => {
    // Redeals exhausted but a move (SPADE 8 stacks under SPADE 9) still exists.
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 9)], [card('SPADE', 8)], [card('DIAMOND', 1)]],
        redealsLeft: 0,
      }),
    );
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('ll-deadlock-banner')).not.toBeInTheDocument();
    expect(screen.getByTestId('giveup-button').className).not.toContain('animate-pulse');
  });

  it('hides the deadlock banner while redeals remain', async () => {
    // No move, but redeals remain -> stuck banner, not the deadlock banner.
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 5)], [card('HEART', 9)], [card('CLOVER', 2)]],
        redealsLeft: 2,
      }),
    );
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('ll-stuck-banner')).toBeInTheDocument());
    expect(screen.queryByTestId('ll-deadlock-banner')).not.toBeInTheDocument();
  });

  it('renders fans and foundations', async () => {
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.getByTestId('foundation-0')).toBeInTheDocument();
  });

  it('selects a source fan then moves to another fan', async () => {
    renderWithProviders(<ShamrocksPage />);
    const src = await screen.findByTestId('fan-1');
    fireEvent.click(src);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('fan-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mf', 1, 0));
  });

  it('selects a source fan then sends it to a foundation', async () => {
    renderWithProviders(<ShamrocksPage />);
    const src = await screen.findByTestId('fan-2');
    fireEvent.click(src);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('foundation-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ff', 2));
  });

  it('redeals, auto-completes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<ShamrocksPage />);
    await screen.findByTestId('redeal-button');
    for (const [testid, cmd] of [
      ['redeal-button', 'rd'],
      ['autocomplete-button', 'ac'],
      ['undo-button', 'u'],
      ['hint-button', 'hint'],
      ['giveup-button', 'giveup'],
    ] as const) {
      mockExec.mockClear();
      fireEvent.click(screen.getByTestId(testid));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(cmd));
    }
  });

  it('disables redeal when none are left', async () => {
    mockExec.mockResolvedValue(makeState({ redealsLeft: 0 }));
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('redeal-button')).toBeDisabled());
  });

  it('hides action buttons at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('redeal-button')).not.toBeInTheDocument();
  });

  // #5678: どの扇が動かせるかは、ヒント (4秒で消える) を押さないと分からなかった。
  // 既定の盤面: ♠9 / ♠8 / ♦A — ♠8 は ♠9 の上へ、♦A は空のファウンデーションへ動ける。
  it('marks the fans that can move without asking for a hint', async () => {
    renderWithProviders(<ShamrocksPage />);

    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.getByTestId('fan-1')).toHaveAttribute('data-movable', 'true');
    expect(screen.getByTestId('fan-2')).toHaveAttribute('data-movable', 'true');
    // ♠9 の行き先 (♠10) は無い。
    expect(screen.getByTestId('fan-0')).not.toHaveAttribute('data-movable');
  });

  it('marks nothing when the board is stuck', async () => {
    mockExec.mockResolvedValue(
      makeState({ fans: [[card('SPADE', 5)], [card('HEART', 9)]], foundation: [[card('CLOVER', 1)], [], [], []] }),
    );
    renderWithProviders(<ShamrocksPage />);

    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.getByTestId('fan-0')).not.toHaveAttribute('data-movable');
    expect(screen.getByTestId('fan-1')).not.toHaveAttribute('data-movable');
  });

  // **ヒントの強調とは別物として読めること。**同じ印だと、4秒で消える推奨手と
  // 常時出ている「動かせる」が区別できない。
  it('keeps the movable marker distinct from the hint markers', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromFan: 1, toFan: 0, toFoundation: false } }));
    renderWithProviders(<ShamrocksPage />);
    await screen.findByTestId('hint-button');

    // ヒント前: 控えめな 1px のリングだけ。パルスはしない。
    const before = screen.getByTestId('fan-1');
    expect(before).toHaveAttribute('data-movable', 'true');
    expect(before).not.toHaveAttribute('data-hint-source');
    expect(before.className).toContain('ring-1');
    expect(before.className).not.toContain('animate-pulse');

    fireEvent.click(screen.getByTestId('hint-button'));

    // ヒント後: 2px + パルスに変わる。**同じ見た目なら区別が付かない。**
    await waitFor(() => expect(screen.getByTestId('fan-1')).toHaveAttribute('data-hint-source', 'true'));
    const after = screen.getByTestId('fan-1');
    expect(after).toHaveAttribute('data-movable', 'true');
    expect(after.className).toContain('animate-pulse');
  });

  it('highlights the hint source and destination fans after a hint', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromFan: 1, toFan: 0, toFoundation: false } }));
    renderWithProviders(<ShamrocksPage />);
    await screen.findByTestId('hint-button');
    // No highlight until the player asks for a hint.
    expect(screen.getByTestId('fan-1')).not.toHaveAttribute('data-hint-source');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('fan-1')).toHaveAttribute('data-hint-source', 'true'));
    expect(screen.getByTestId('fan-0')).toHaveAttribute('data-hint-dest', 'true');
    expect(screen.getByTestId('ll-foundation-row')).not.toHaveAttribute('data-hint-foundation');
  });

  it('highlights the foundation row for a to-foundation hint', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { fromFan: 2, toFan: -1, toFoundation: true } }));
    renderWithProviders(<ShamrocksPage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() =>
      expect(screen.getByTestId('ll-foundation-row')).toHaveAttribute('data-hint-foundation', 'true'),
    );
    expect(screen.getByTestId('fan-2')).toHaveAttribute('data-hint-source', 'true');
    // No fan is marked as the destination when the move targets a foundation.
    expect(screen.getByTestId('fan-0')).not.toHaveAttribute('data-hint-dest');
  });

  it('clears the hint highlight when the board changes', async () => {
    mockExec
      .mockResolvedValueOnce(makeState()) // mount reset
      .mockResolvedValueOnce(makeState({ hint: { fromFan: 1, toFan: 0, toFoundation: false } })) // hint
      .mockResolvedValue(makeState({ moveCount: 1 })); // subsequent move advances the board
    renderWithProviders(<ShamrocksPage />);
    await screen.findByTestId('hint-button');
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('fan-1')).toHaveAttribute('data-hint-source', 'true'));
    // Perform a move: select fan-1 then drop on fan-0.
    fireEvent.click(screen.getByTestId('fan-1'));
    fireEvent.click(screen.getByTestId('fan-0'));
    await waitFor(() => expect(screen.getByTestId('fan-1')).not.toHaveAttribute('data-hint-source'));
    expect(screen.getByTestId('fan-0')).not.toHaveAttribute('data-hint-dest');
  });
});
