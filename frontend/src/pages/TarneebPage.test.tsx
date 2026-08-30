import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tarneebApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TarneebResponse } from '../types/card';
import { TarneebPhase } from '../types/phases';
import { TarneebPage } from './TarneebPage';

vi.mock('../api/gameApi', () => ({
  tarneebApi: { exec: vi.fn() },
  actionLogApi: { tarneeb: vi.fn() },
}));

const mockExec = vi.mocked(tarneebApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function player(id: number, isHuman: boolean, team: number, trickCount: number, roundScore = 0) {
  return {
    id,
    isHuman,
    team,
    cardCount: 5,
    cards: isHuman ? [card('SPADE', 1)] : [],
    bid: -1,
    roundScore,
    cumulativeScore: 0,
    trickCount,
  };
}

function makeState(overrides: Partial<TarneebResponse> = {}): TarneebResponse {
  return {
    players: [
      player(0, true, 0, 3, 4),
      player(1, false, 1, 2, 2),
      player(2, false, 0, 1, 4),
      player(3, false, 1, 0, 2),
    ],
    teamScores: [10, 5],
    phase: 0,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    bidWinnerIdx: -1,
    highestBid: 0,
    trumpSuit: -1,
    redealCount: 0,
    dealerIdx: 0,
    currentTrick: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    validPlayIndices: [],
    config: { cpuDifficulty: 1, pointLimit: 31, minBid: 7 },
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('TarneebPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
  });

  it('labels the trump-declaration buttons with translated suit names', async () => {
    // Trump-declaration phase with the human (player 0) as bid winner.
    mockExec.mockResolvedValue(makeState({ phase: 1, bidWinnerIdx: 0, highestBid: 8 }));
    renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スペード' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'クラブ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ハート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ダイヤ' })).toBeInTheDocument();
    // The old untranslated "trump-N" label is gone.
    expect(screen.queryByRole('button', { name: 'trump-1' })).not.toBeInTheDocument();
  });

  it('shows the redeal count in the round info when a redeal has occurred', async () => {
    mockExec.mockResolvedValue(makeState({ redealCount: 2 }));
    renderWithProviders(<TarneebPage />);
    const redeal = await screen.findByTestId('tarneeb-redeal-count');
    expect(redeal).toHaveTextContent('リディール 2回');
  });

  it('hides the redeal count when no redeal has occurred', async () => {
    mockExec.mockResolvedValue(makeState({ redealCount: 0 }));
    renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByTestId('tarneeb-redeal-count')).not.toBeInTheDocument();
  });

  it('labels the human team and opponents and shows round score + total per team', async () => {
    const { container } = renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="tn-score-table"] table')).not.toBeNull());
    const table = container.querySelector('[data-tutorial="tn-score-table"] table') as HTMLTableElement;
    // 列は チーム / トリック / 今ラウンド / 累計。数字が列をまたいで重複しうるので、
    // テキスト一致ではなく列位置で読む。
    const cellsOf = (label: string) => {
      const row = within(table).getByText(label).closest('tr') as HTMLTableRowElement;
      return Array.from(row.cells).map((c) => c.textContent);
    };
    // Your team (team 0): 4 tricks, round score 4, total 10.
    expect(cellsOf('あなたのチーム')).toEqual(['あなたのチーム', '4', '4', '10']);
    // Opponents (team 1): 2 tricks, round score 2, total 5.
    expect(cellsOf('相手チーム')).toEqual(['相手チーム', '2', '2', '5']);
  });

  it('displays round score (including negative scores for failed bids) instead of trick count in the score table', async () => {
    // Team 0 failed bid: bid 8, took 3+1=4 tricks -> roundScore is -8.
    // Team 1 defenders: took 6+3=9 tricks -> roundScore is +9.
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player(0, true, 0, 3, -8),
          player(1, false, 1, 6, 9),
          player(2, false, 0, 1, -8),
          player(3, false, 1, 3, 9),
        ],
        teamScores: [2, 15],
      }),
    );
    const { container } = renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="tn-score-table"] table')).not.toBeNull());
    const table = container.querySelector('[data-tutorial="tn-score-table"] table') as HTMLTableElement;

    const cellsOf = (label: string) => {
      const row = within(table).getByText(label).closest('tr') as HTMLTableRowElement;
      return Array.from(row.cells).map((c) => c.textContent);
    };
    // チーム / トリック / 今ラウンド / 累計。
    // Team 0: 4 tricks but a *negative* round score — the two columns must not agree here,
    // which is the whole point of #6402.
    expect(cellsOf('あなたのチーム')).toEqual(['あなたのチーム', '4', '-8', '2']);
    // Team 1: 9 tricks, round score 9, total 15.
    expect(cellsOf('相手チーム')).toEqual(['相手チーム', '9', '9', '15']);

    // Player breakdown preserves individual trick counts (acceptance condition 3).
    const breakdown = (await screen.findByTestId('tn-player-breakdown')) as HTMLDetailsElement;
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-0"]')?.textContent).toBe('3トリック');
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-2"]')?.textContent).toBe('1トリック');
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-1"]')?.textContent).toBe('6トリック');
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-3"]')?.textContent).toBe('3トリック');
  });

  it('shows a per-player trick breakdown grouped by team (#3306)', async () => {
    const { container } = renderWithProviders(<TarneebPage />);
    const breakdown = (await screen.findByTestId('tn-player-breakdown')) as HTMLDetailsElement;
    // Both teams appear as breakdown groups.
    const yourTeamGroup = within(breakdown).getByTestId('tn-breakdown-team-0');
    const oppTeamGroup = within(breakdown).getByTestId('tn-breakdown-team-1');
    expect(within(yourTeamGroup).getByText('あなたのチーム')).toBeInTheDocument();
    expect(within(oppTeamGroup).getByText('相手チーム')).toBeInTheDocument();
    // Each player's individual trick count is shown (players 0/2 → team 0, 1/3 → team 1).
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-0"]')?.textContent).toBe('3トリック');
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-2"]')?.textContent).toBe('1トリック');
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-1"]')?.textContent).toBe('2トリック');
    expect(breakdown.querySelector('[data-testid="tn-breakdown-tricks-3"]')?.textContent).toBe('0トリック');
    // The aggregate team table is preserved alongside the breakdown.
    expect(container.querySelector('[data-tutorial="tn-score-table"] table')).not.toBeNull();
  });

  it('renders a bid button group from minBid to 13 and bids the selected value', async () => {
    renderWithProviders(<TarneebPage />);
    // minBid 7 → buttons 7..13, none below 7.
    const bid9 = await screen.findByTestId('bid-option-9');
    expect(screen.queryByTestId('bid-option-6')).not.toBeInTheDocument();
    expect(screen.getByTestId('bid-option-13')).toBeInTheDocument();

    fireEvent.click(bid9);
    expect(bid9).toHaveAttribute('aria-pressed', 'true');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 9));
  });

  it('disables bid buttons that do not beat the current highest bid and pre-selects the lowest legal bid', async () => {
    mockExec.mockResolvedValue(makeState({ highestBid: 9 }));
    renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(screen.getByTestId('bid-option-9')).toBeDisabled());
    expect(screen.getByTestId('bid-option-7')).toBeDisabled();
    expect(screen.getByTestId('bid-option-10')).toBeEnabled();
    // The effect snaps the selection to the lowest legal value (highestBid + 1 = 10).
    expect(screen.getByTestId('bid-option-10')).toHaveAttribute('aria-pressed', 'true');
  });

  it('shows no bid controls outside the human bid turn', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 3 })); // PLAY phase
    renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByTestId('bid-option-7')).not.toBeInTheDocument();
  });

  it('passes by bidding 0', async () => {
    renderWithProviders(<TarneebPage />);
    await screen.findByTestId('bid-option-7');
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });

  // **ドメインは合法手を判定済みなのに、画面が使っていなかった (#4713)。**
  // マストフォローに反する札もクリックでき、サーバーのエラーが返って初めて
  // 出せないと分かる状態だった。
  it('dims the cards that must-follow forbids on the human play turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: TarneebPhase.PLAY,
        currentPlayerIdx: 0,
        players: [
          { ...player(0, true, 0, 3), cards: [card('SPADE', 1), card('HEART', 13)] },
          player(1, false, 1, 2),
          player(2, false, 0, 1),
          player(3, false, 1, 0),
        ],
        // 合法なのは index 1 だけ。
        validPlayIndices: [1],
      }),
    );
    renderWithProviders(<TarneebPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const cards = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
    expect(cards).toHaveLength(2);
    expect(cards[0]).toHaveAttribute('aria-disabled', 'true');
    expect(cards[1]).not.toHaveAttribute('aria-disabled', 'true');
  });

  // **空を「制限なし」と読まない。**CPU の手番では制限そのものを送っていないので
  // 全札が有効のまま。ここを validPlayIndices.length で判定すると、
  // 「1枚も出せない」局面と区別が付かなくなる。
  it('leaves every card enabled when it is not the human turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: TarneebPhase.PLAY,
        currentPlayerIdx: 1,
        players: [
          { ...player(0, true, 0, 3), cards: [card('SPADE', 1), card('HEART', 13)] },
          player(1, false, 1, 2),
          player(2, false, 0, 1),
          player(3, false, 1, 0),
        ],
        validPlayIndices: [],
      }),
    );
    renderWithProviders(<TarneebPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const cards = screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));
    expect(cards).toHaveLength(2);
    for (const c of cards) {
      expect(c).not.toHaveAttribute('aria-disabled', 'true');
    }
  });

  // このページのヒントは `hintAvailable` を使わないので、読み上げガードが
  // 一度も見ておらず aria-live が無いまま出荷されていた (#6663)。領域は
  // **常設**で、分岐ごとにその分岐だけが出す語を見る。
  describe('TarneebPage hint live region', () => {
    it('is mounted and empty before any hint arrives', async () => {
      mockExec.mockResolvedValue(makeState());
      renderWithProviders(<TarneebPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const region = screen.getByTestId('tarneeb-hint-live');
      expect(region).toHaveAttribute('role', 'status');
      expect(region).toHaveAttribute('aria-live', 'polite');
      expect(region).toBeEmptyDOMElement();
    });

    it('names a bid recommendation', async () => {
      const bidTurn = makeState({ phase: TarneebPhase.BID, bidPlayerIdx: 0 });
      mockExec.mockResolvedValue(bidTurn);
      renderWithProviders(<TarneebPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...bidTurn,
        hint: { bid: 8, reason: 'bid_estimate' },
      } as unknown as TarneebResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('tarneeb-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨ビッド'));
      expect(region).toHaveTextContent('8');
      expect(region).not.toHaveTextContent('{{');
    });

    it('names a trump recommendation', async () => {
      const trumpTurn = makeState({ phase: TarneebPhase.TRUMP_DECLARATION, bidWinnerIdx: 0 });
      mockExec.mockResolvedValue(trumpTurn);
      renderWithProviders(<TarneebPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...trumpTurn,
        hint: { trumpSuit: 0, reason: 'trump_longest' },
      } as unknown as TarneebResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('tarneeb-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨トランプ'));
      expect(region).not.toHaveTextContent('推奨ビッド');
      expect(region).not.toHaveTextContent('{{');
    });

    it('names the card to play', async () => {
      const playTurn = makeState({ phase: TarneebPhase.PLAY, currentPlayerIdx: 0 });
      mockExec.mockResolvedValue(playTurn);
      renderWithProviders(<TarneebPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

      mockExec.mockResolvedValue({
        ...playTurn,
        hint: { cardIndex: 2, reason: 'lead_strong' },
      } as unknown as TarneebResponse);
      fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

      const region = await screen.findByTestId('tarneeb-hint-live');
      await waitFor(() => expect(region).toHaveTextContent('推奨カード'));
      expect(region).toHaveTextContent('[2]');
      expect(region).not.toHaveTextContent('推奨トランプ');
      expect(region).not.toHaveTextContent('{{');
    });
  });
});
