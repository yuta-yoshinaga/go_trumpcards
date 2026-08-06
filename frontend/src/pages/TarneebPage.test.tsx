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

function player(id: number, isHuman: boolean, team: number, trickCount: number) {
  return {
    id,
    isHuman,
    team,
    cardCount: 5,
    cards: isHuman ? [card('SPADE', 1)] : [],
    bid: -1,
    roundScore: 0,
    cumulativeScore: 0,
    trickCount,
  };
}

function makeState(overrides: Partial<TarneebResponse> = {}): TarneebResponse {
  return {
    players: [player(0, true, 0, 3), player(1, false, 1, 2), player(2, false, 0, 1), player(3, false, 1, 0)],
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

  it('labels the human team and opponents and shows round tricks + total per team', async () => {
    const { container } = renderWithProviders(<TarneebPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="tn-score-table"] table')).not.toBeNull());
    const table = container.querySelector('[data-tutorial="tn-score-table"] table') as HTMLTableElement;
    // Your team (team 0): round tricks 3 + 1 = 4, total 10.
    const yourRow = within(table).getByText('あなたのチーム').closest('tr') as HTMLTableRowElement;
    expect(within(yourRow).getByText('4')).toBeInTheDocument();
    expect(within(yourRow).getByText('10')).toBeInTheDocument();
    // Opponents (team 1): round tricks 2 + 0 = 2, total 5.
    const oppRow = within(table).getByText('相手チーム').closest('tr') as HTMLTableRowElement;
    expect(within(oppRow).getByText('2')).toBeInTheDocument();
    expect(within(oppRow).getByText('5')).toBeInTheDocument();
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
});
