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

  // **Shamrocks はリディールを持たない。**合法手が尽きたらそこで終わりなので、
  // ラ・ベル・リュシー版の「リディールを勧める」中間バナーは存在しない。
  it('shows the deadlock banner and pulses give up when no legal move remains', async () => {
    // 5 / 9 / 2 -- no rank is within one of another, no Ace, no empty fan.
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 5)], [card('HEART', 9)], [card('CLOVER', 2)]],
        foundation: [[], [], [], []],
      }),
    );
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('ll-deadlock-banner')).toBeInTheDocument());
    // There is no redeal to recommend, so no intermediate banner may appear.
    expect(screen.queryByTestId('ll-stuck-banner')).not.toBeInTheDocument();
    expect(screen.getByTestId('giveup-button').className).toContain('animate-pulse');
  });

  it('hides the deadlock banner when a legal move exists', async () => {
    // Default board: SPADE 8 stacks onto SPADE 9 (adjacent rank), DIAMOND A to a foundation.
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('ll-deadlock-banner')).not.toBeInTheDocument();
    expect(screen.getByTestId('giveup-button').className).not.toContain('animate-pulse');
  });

  // **スート不問**なので、ラ・ベル・リュシーなら詰みの盤でも Shamrocks は動ける。
  it('sees a cross-suit adjacent-rank move as legal', async () => {
    mockExec.mockResolvedValue(
      makeState({
        fans: [[card('SPADE', 5)], [card('HEART', 6)], [card('CLOVER', 9)]],
        foundation: [[], [], [], []],
      }),
    );
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
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

  it('auto-completes, undoes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<ShamrocksPage />);
    await screen.findByTestId('autocomplete-button');
    for (const [testid, cmd] of [
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

  // **リディールのボタンごと消えていること。**盤を配り直さないのに 200 を返す
  // 無言の no-op を押させないため、UI 側にも残さない。
  it('offers no redeal button at all', async () => {
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('redeal-button')).not.toBeInTheDocument();
  });

  it('hides action buttons at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<ShamrocksPage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('autocomplete-button')).not.toBeInTheDocument();
  });

  // #5678: どの扇が動かせるかは、ヒント (4秒で消える) を押さないと分からなかった。
  // 既定の盤面 ♠9 / ♠8 / ♦A は Shamrocks では**3 つとも動く**。ランクが 1 つ違えば
  // 上でも下でもよいので ♠9→♠8 も ♠8→♠9 も合法で、♦A は空のファウンデーションへ行ける。
  // (ラ・ベル・リュシーは同スート降順のみなので ♠9 は動けない。そちらの期待値のまま
  //  クローンすると、通ってしまうのに間違っているテストになる。)
  it('marks the fans that can move without asking for a hint', async () => {
    renderWithProviders(<ShamrocksPage />);

    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    for (const id of ['fan-0', 'fan-1', 'fan-2']) {
      expect(screen.getByTestId(id)).toHaveAttribute('data-movable', 'true');
    }
  });

  // 負のコントロール: 動かせない扇には印が付かないこと。5 / 9 / 2 はどれも
  // 互いにランクが 2 以上離れていて、A も無い。
  it('marks no fan when nothing can move', async () => {
    mockExec.mockResolvedValue(makeState({ fans: [[card('SPADE', 5)], [card('HEART', 9)], [card('CLOVER', 2)]] }));
    renderWithProviders(<ShamrocksPage />);

    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    for (const id of ['fan-0', 'fan-1', 'fan-2']) {
      expect(screen.getByTestId(id)).not.toHaveAttribute('data-movable');
    }
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
