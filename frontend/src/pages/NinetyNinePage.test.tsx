import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, ninetyNineApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { NinetyNineResponse } from '../types/card';
import { NinetyNinePage } from './NinetyNinePage';

vi.mock('../api/gameApi', () => ({
  ninetyNineApi: { exec: vi.fn() },
  actionLogApi: { ninetynine: vi.fn() },
}));

const mockExec = vi.mocked(ninetyNineApi.exec);

const playPhaseState: NinetyNineResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 9,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 3,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      buriedCount: 3,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 9,
      cards: [],
      bid: 2,
      roundScore: 0,
      cumulativeScore: 10,
      trickCount: 1,
      buriedCount: 3,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 9,
      cards: [],
      bid: 4,
      roundScore: 0,
      cumulativeScore: 20,
      trickCount: 2,
      buriedCount: 3,
    },
  ],
  phase: 1,
  dealNumber: 2,
  targetScore: 100,
  handSize: 9,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 2,
  trumpSuit: 3,
  currentTrick: [],
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, targetScore: 100 },
};

const bidPhaseState: NinetyNineResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  players: playPhaseState.players.map((p) => ({
    ...p,
    bid: -1,
    cardCount: 12,
    buriedCount: 0,
    cards: p.isHuman
      ? [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 11 },
          { design: 'CLOVER', value: 5 },
          { design: 'DIAMOND', value: 8 },
        ]
      : [],
  })),
};

const bidPhaseCpuState: NinetyNineResponse = { ...bidPhaseState, bidPlayerIdx: 1 };

const trickEndState: NinetyNineResponse = {
  ...playPhaseState,
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: NinetyNineResponse = { ...playPhaseState, phase: 3 };

const gameEndState: NinetyNineResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};

const cpuTurnState: NinetyNineResponse = { ...playPhaseState, currentPlayerIdx: 1 };

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('NinetyNinePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<NinetyNinePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with config', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, targetScore: 100 }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('shows deal/trick/trump info', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => {
      expect(screen.getByText('ディール 2')).toBeInTheDocument();
      expect(screen.getByText('トリック 1')).toBeInTheDocument();
      expect(screen.getByText('切り札: ハート')).toBeInTheDocument();
      expect(screen.getByText('ディーラー: CPU 2')).toBeInTheDocument();
    });
  });

  it('shows bury instruction during human bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() =>
      expect(screen.getByText('ちょうど3枚を埋めてください（スート合計があなたのビッド）')).toBeInTheDocument(),
    );
  });

  it('bury button is focusable + aria-disabled with a reason until 3 cards selected, then calls bid', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    // Not enough cards: the button exposes the reason in its accessible name and
    // is aria-disabled, but NOT HTML-disabled — it stays focusable.
    const disabledBtn = screen.getByRole('button', { name: '3枚埋める（あと 3 枚のカード選択が必要です）' });
    expect(disabledBtn).toHaveAttribute('aria-disabled', 'true');
    expect(disabledBtn).not.toBeDisabled();

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '3枚埋める（あと 1 枚のカード選択が必要です）' })).toHaveAttribute(
      'aria-disabled',
      'true',
    );
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);

    // Exactly 3 selected: plain label, no aria-disabled.
    const readyBtn = screen.getByRole('button', { name: '3枚埋める' });
    expect(readyBtn).not.toHaveAttribute('aria-disabled');

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(readyBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', [0, 1, 2]));
  });

  it('announces the remaining bury count in a polite live region as cards are selected', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    const region = await screen.findByTestId('nn-bury-progress');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveTextContent('あと 3 枚選択してください');

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('nn-bury-progress')).toHaveTextContent('あと 2 枚選択してください');
  });

  it('announces readiness once exactly 3 cards are selected', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('nn-bury-progress')).toHaveTextContent('3枚選択しました。埋めるボタンで確定できます');
  });

  it('shows a live declared-trick preview that matches the domain suit mapping, plus a legend', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    // Legend is visible throughout the bid phase.
    expect(screen.getByTestId('nn-bid-legend')).toHaveTextContent('スート対応: ♦=0 ♠=1 ♥=2 ♣=3');

    // No cards selected yet: total is 0.
    const preview = screen.getByTestId('nn-bid-preview');
    expect(preview).toHaveAttribute('aria-live', 'polite');
    expect(preview).toHaveTextContent('選択中の合計 = 宣言 0 トリック');

    // ♠ A → +1 (SPADE=1).
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('nn-bid-preview')).toHaveTextContent('選択中の合計 = 宣言 1 トリック');

    // ♥ J → +2 (HEART=2) ⇒ 3.
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('nn-bid-preview')).toHaveTextContent('選択中の合計 = 宣言 3 トリック');

    // ♣ 5 → +3 (CLOVER=3) ⇒ 6, the declared bid for these exact 3 cards.
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('nn-bid-preview')).toHaveTextContent('選択中の合計 = 宣言 6 トリック');

    // ♦ 8 → +0 (DIAMOND=0): total stays 6.
    fireEvent.click(screen.getByAltText('♦ 8').closest('button') as HTMLButtonElement);
    expect(screen.getByTestId('nn-bid-preview')).toHaveTextContent('選択中の合計 = 宣言 6 トリック');
  });

  it('announces an over-selection message instead of a negative count when more than 3 are picked', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    // The bid-phase hand has 4 cards; selecting all 4 over-selects by 1.
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♥ J').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♣ 5').closest('button') as HTMLButtonElement);
    fireEvent.click(screen.getByAltText('♦ 8').closest('button') as HTMLButtonElement);

    expect(screen.getByTestId('nn-bury-progress')).toHaveTextContent(
      '1 枚多く選択されています。ちょうど3枚にしてください',
    );
    // Button stays aria-disabled (not ready) with the deselect-reason label.
    const btn = screen.getByRole('button', { name: '3枚埋める（1 枚多いため選択を減らしてください）' });
    expect(btn).toHaveAttribute('aria-disabled', 'true');
  });

  it('does not show bury controls on cpu bid turn', async () => {
    mockExec.mockResolvedValue(bidPhaseCpuState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '3枚埋める' })).not.toBeInTheDocument();
  });

  it('play button disabled until 1 card selected, then calls play', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

    expect(screen.getByRole('button', { name: 'カードを出す' })).toBeDisabled();
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: 'カードを出す' })).not.toBeDisabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'カードを出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'カードを出す' })).not.toBeInTheDocument();
  });

  it('shows next trick button and calls next', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows next round button and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => {
      expect(screen.getByText('現在のトリック')).toBeInTheDocument();
      expect(screen.getByAltText('♦ 3')).toBeInTheDocument();
    });
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('score table shows all players', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => {
      expect(screen.getByText('スコア')).toBeInTheDocument();
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
    });
  });

  it('settings panel changes cpuDifficulty and applies on reset', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByText('スコア')).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 2, targetScore: 100 }),
    );
  });

  it('shows confirm dialog when reset is clicked', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('handles action log fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByText('棋譜を見る')).toBeInTheDocument());

    vi.mocked(actionLogApi.ninetynine).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('棋譜を見る'));
    await waitFor(() => expect(actionLogApi.ninetynine).toHaveBeenCalledTimes(1));
  });

  it('play-phase hint fetches and shows the recommended card with its reason', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByTestId('nn-hint-button')).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...playPhaseState, hint: { cardIndex: 1, reason: 'follow_suit' } });
    fireEvent.click(screen.getByTestId('nn-hint-button'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    expect(screen.getByTestId('nn-server-hint')).toHaveTextContent('推奨カード: [1] (リードスートに追随)');
  });

  // **埋めヒントには「適用」があるのに、プレイヒントは数字を出すだけだった
  // (#4739)。**プレイヤーは [N] を見て手札から目視で探す必要があり、同じヒント
  // なのに体験が非対称だった。
  it('play-phase hint Apply selects the recommended card', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByTestId('nn-hint-button')).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...playPhaseState, hint: { cardIndex: 1, reason: 'follow_suit' } });
    fireEvent.click(screen.getByTestId('nn-hint-button'));
    await waitFor(() => expect(screen.getByTestId('nn-hint-apply-play')).toBeInTheDocument());

    // 適用前は何も選択されていない = 出すボタンは押せない。
    const play = screen.getByRole('button', { name: 'カードを出す' });
    expect(play).toBeDisabled();

    fireEvent.click(screen.getByTestId('nn-hint-apply-play'));

    // 適用すると推奨カードが選択され、そのまま出せる状態になる。
    expect(play).not.toBeDisabled();
  });

  it('bid-phase hint shows the bury recommendation and Apply selects those three cards', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByTestId('nn-hint-button')).toBeInTheDocument());

    mockExec.mockResolvedValueOnce({ ...bidPhaseState, hint: { buryIndices: [0, 1, 2], reason: 'strategic_bury' } });
    fireEvent.click(screen.getByTestId('nn-hint-button'));

    await waitFor(() => expect(screen.getByTestId('nn-server-hint')).toBeInTheDocument());
    expect(screen.getByTestId('nn-server-hint')).toHaveTextContent('到達可能なビッドになるよう埋める');

    // Applying the bury hint replaces the selection with the recommended 3 cards,
    // so the Bury button becomes ready (no aria-disabled).
    fireEvent.click(screen.getByTestId('nn-hint-apply'));
    expect(screen.getByTestId('nn-bury-progress')).toHaveTextContent('3枚選択しました。埋めるボタンで確定できます');
    expect(screen.getByRole('button', { name: '3枚埋める' })).not.toHaveAttribute('aria-disabled');
  });

  it('shows a network error when the hint request fails', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByTestId('nn-hint-button')).toBeInTheDocument());

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(screen.getByTestId('nn-hint-button'));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('does not show the hint button on a cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getByText('CPU 1')).toBeInTheDocument());
    expect(screen.queryByTestId('nn-hint-button')).not.toBeInTheDocument();
  });

  it('renders accessible h1 heading and tutorial button', async () => {
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => {
      expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument();
    });
  });
});

// #5515: サーバは leadPlayerIdx を常に返し、型にも入っていて、TrickDisplay は
// winnerIdx バッジを出す機能を持っているのに、**このページだけ渡していない。**
// 姉妹の Oh Hell は同じ形で出している。
describe('NinetyNinePage trick winner badge', () => {
  it('marks the trick winner once the trick ends', async () => {
    mockExec.mockResolvedValue({ ...trickEndState, leadPlayerIdx: 1 });
    renderWithProviders(<NinetyNinePage />);
    const badge = await screen.findByTestId('trick-winner-badge');
    expect(badge).toBeInTheDocument();
  });

  // **勝者は leadPlayerIdx。** 次のトリックのリードが直前の勝者なので、
  // 固定値を出していれば札の位置が動かない。バッジの付いた札で見分ける。
  it('follows leadPlayerIdx rather than a fixed seat', async () => {
    const badgedCardAlt = async () => {
      const badge = await screen.findByTestId('trick-winner-badge');
      const cell = badge.closest('[data-trick-winner]');
      return cell?.querySelector('img')?.getAttribute('alt') ?? null;
    };

    mockExec.mockResolvedValue({ ...trickEndState, leadPlayerIdx: 0 });
    const { unmount } = renderWithProviders(<NinetyNinePage />);
    const first = await badgedCardAlt();
    expect(first).toBeTruthy();
    unmount();

    mockExec.mockResolvedValue({ ...trickEndState, leadPlayerIdx: 1 });
    renderWithProviders(<NinetyNinePage />);
    expect(await badgedCardAlt()).not.toBe(first);
  });

  // **プレイ中は出さない。** まだ決着していないトリックに勝者を付けると、
  // リードした人が勝ったように読める。
  //
  // 場に札が出ている状態で確かめる -- currentTrick が空のフィクスチャだと、
  // バッジを出す実装でも出ないので何も検証できない。
  it('stays quiet while the trick is still being played', async () => {
    mockExec.mockResolvedValue({ ...trickEndState, phase: 1, leadPlayerIdx: 1 });
    renderWithProviders(<NinetyNinePage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBeGreaterThan(0));
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });
});
