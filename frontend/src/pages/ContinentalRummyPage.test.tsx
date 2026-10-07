import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { continentalrummyApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeContinentalRummyState } from '../test/stateFactories';
import { ContinentalRummyPage } from './ContinentalRummyPage';

vi.mock('../api/gameApi', () => ({
  continentalrummyApi: { exec: vi.fn() },
  actionLogApi: { continentalrummy: vi.fn() },
}));

const mockExec = vi.mocked(continentalrummyApi.exec);

const discardState = makeContinentalRummyState();
const drawState = makeContinentalRummyState({
  phase: 'draw',
  hintReason: 'draw_stock',
  goOutIdx: -1,
  hintDiscardIdx: -1,
  messageCode: 'continentalrummy.drawPhase',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(drawState);
});

describe('ContinentalRummyPage', () => {
  it('shows every final score in descending order with shared ranks for ties', async () => {
    mockExec.mockResolvedValue(
      makeContinentalRummyState({
        phase: 'gameEnd',
        gameEndFlag: true,
        players: makeContinentalRummyState().players.map((player, index) => ({
          ...player,
          score: [20, 45, 45, 10][index],
        })),
      }),
    );
    renderWithProviders(<ContinentalRummyPage />);
    const ranking = await screen.findByTestId('cont-final-ranking');
    expect(ranking).toHaveTextContent('最終順位');
    expect(
      Array.from(ranking.querySelectorAll('tbody tr')).map((row) =>
        Array.from(row.children).map((cell) => cell.textContent),
      ),
    ).toEqual([
      ['1', 'CPU 1', '45'],
      ['1', 'CPU 2', '45'],
      ['3', 'あなた', '20'],
      ['4', 'CPU 3', '10'],
    ]);
  });

  it('shows the complete final ranking when the game ends in a tie', async () => {
    mockExec.mockResolvedValue(
      makeContinentalRummyState({
        phase: 'gameEnd',
        gameEndFlag: true,
        winnerIdx: -1,
        players: makeContinentalRummyState().players.map((player) => ({ ...player, score: 12 })),
      }),
    );
    renderWithProviders(<ContinentalRummyPage />);
    const ranking = await screen.findByTestId('cont-final-ranking');
    expect(ranking.querySelectorAll('tbody tr')).toHaveLength(4);
    expect(Array.from(ranking.querySelectorAll('tbody tr')).map((row) => row.textContent?.match(/^\d+/)?.[0])).toEqual([
      '1',
      '1',
      '1',
      '1',
    ]);
    expect(ranking).toHaveTextContent('12');
  });

  it('shows completed round score gains in a round by player table', async () => {
    mockExec.mockResolvedValue(
      makeContinentalRummyState({
        roundScoreHistory: [
          { roundNumber: 1, scores: [0, 54, 0, 0] },
          { roundNumber: 2, scores: [42, 0, 0, 0] },
        ],
      }),
    );
    renderWithProviders(<ContinentalRummyPage />);
    const history = await screen.findByTestId('cont-score-history');
    expect(history).toHaveTextContent('得点履歴');
    expect(history.querySelectorAll('thead th')).toHaveLength(14);
    expect(history.querySelectorAll('tbody tr')).toHaveLength(2);
    expect(history).toHaveTextContent('累計');
    expect(Array.from(history.querySelectorAll('tbody tr')[0].children).map((cell) => cell.textContent)).toEqual([
      '1',
      '0',
      '0',
      '54',
      '54',
      '0',
      '0',
      '0',
      '0',
    ]);
    expect(Array.from(history.querySelectorAll('tbody tr')[1].children).map((cell) => cell.textContent)).toEqual([
      '2',
      '42',
      '42',
      '0',
      '54',
      '0',
      '0',
      '0',
      '0',
    ]);
  });

  it('sends selected difficulty and rounds when resetting', async () => {
    mockExec.mockResolvedValue(makeContinentalRummyState({ gameEndFlag: true, phase: 'gameEnd' }));
    renderWithProviders(<ContinentalRummyPage />);
    fireEvent.change(await screen.findByTestId('continentalrummy-cpuDifficulty'), { target: { value: '2' } });
    fireEvent.change(screen.getByTestId('continentalrummy-totalRounds'), { target: { value: '10' } });
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2, totalRounds: 10 } }),
    );
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<ContinentalRummyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **上がれる形は常に見えていること。** 5+5+5 がそこに無いのが肝。
  it('shows the legal layouts and never advertises 5+5+5', async () => {
    renderWithProviders(<ContinentalRummyPage />);
    const layouts = await screen.findByTestId('cont-layouts');
    expect(layouts).toHaveTextContent('3+3+3+3+3');
    expect(layouts).toHaveTextContent('4+4+4+3');
    expect(layouts).toHaveTextContent('5+4+3+3');
    expect(layouts).not.toHaveTextContent('5+5+5');
    expect(screen.getByTestId('cont-nosets')).toHaveTextContent('セット');
  });

  it('highlights the server supplied run cards only when a go-out exists', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<ContinentalRummyPage />);
    await screen.findByTestId('cont-goout');
    const hand = document.querySelector('[data-tutorial="continentalrummy-player-hand"]');
    expect(hand).toBeInTheDocument();
    expect(hand?.querySelectorAll('.ring-ds-success').length).toBeGreaterThanOrEqual(15);
    for (let runIndex = 1; runIndex <= 5; runIndex++) {
      expect(screen.getAllByRole('button', { name: new RegExp(`連番${runIndex}`) })).toHaveLength(3);
    }
    expect(screen.getByRole('button', { name: /捨てる札/ })).toBeInTheDocument();
  });

  it('shows no run markers when the hand cannot go out', async () => {
    mockExec.mockResolvedValue(makeContinentalRummyState({ goOutIdx: -1, goOutGroups: [] }));
    renderWithProviders(<ContinentalRummyPage />);
    await screen.findByTestId('cont-discard-notice');
    expect(screen.queryByRole('button', { name: /連番[1-5]/ })).not.toBeInTheDocument();
  });

  it('does not show a zero run marker for an uncovered hand index', async () => {
    mockExec.mockResolvedValue(makeContinentalRummyState({ goOutIdx: 0, goOutGroups: [] }));
    renderWithProviders(<ContinentalRummyPage />);
    await screen.findByTestId('cont-goout');
    expect(screen.queryByRole('button', { name: /連番0/ })).not.toBeInTheDocument();
  });

  it('shows the stock, the discard top and every seat', async () => {
    renderWithProviders(<ContinentalRummyPage />);
    expect(await screen.findByTestId('cont-stock')).toHaveTextContent('30');
    expect(screen.getByTestId('cont-discard-top').children.length).toBeGreaterThan(0);
    for (const id of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`cont-seat-${id}`)).toBeInTheDocument();
    }
    expect(screen.getByTestId('cont-seat-1')).toHaveTextContent('15');
  });

  // **山と捨て札は別のボタンで、別のコマンドとして届く。**
  it('offers both draws and sends each as its own command', async () => {
    renderWithProviders(<ContinentalRummyPage />);
    const stockButton = await screen.findByRole('button', { name: '山札から引く' });
    const discardButton = screen.getByRole('button', { name: '捨て札を取る' });
    expect(stockButton.querySelector('kbd')).toHaveTextContent('S');
    expect(discardButton.querySelector('kbd')).toHaveTextContent('T');
    fireEvent.click(stockButton);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stock'));
    expect(mockExec).not.toHaveBeenCalledWith('take');

    mockExec.mockClear();
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<ContinentalRummyPage />);
    const takes = await screen.findAllByRole('button', { name: '捨て札を取る' });
    fireEvent.click(takes[takes.length - 1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('take'));
    expect(mockExec).not.toHaveBeenCalledWith('stock');
  });

  it('does not offer the discard buttons while it is the draw step', async () => {
    renderWithProviders(<ContinentalRummyPage />);
    await screen.findByTestId('cont-layouts');
    expect(screen.queryByTestId('cont-goout')).not.toBeInTheDocument();
    expect(screen.queryByTestId('cont-discard-notice')).not.toBeInTheDocument();
  });

  it('does not show the discard shortcut when the discard pile is empty', async () => {
    mockExec.mockResolvedValue(makeContinentalRummyState({ phase: 'draw', discardTop: undefined }));
    renderWithProviders(<ContinentalRummyPage />);
    const takeButton = await screen.findByRole('button', { name: '捨て札を取る' });
    expect(takeButton).toBeDisabled();
    expect(takeButton.querySelector('kbd')).not.toBeInTheDocument();
  });

  // **上がれるときは黙っていない。** 15 枚の分割は目で追いきれない。
  it('offers going out with the index the server solved', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<ContinentalRummyPage />);
    fireEvent.click(await screen.findByTestId('cont-goout'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('goout', { handIndex: 15 }));
  });

  // 負のコントロール: 上がれないときはボタンを出さない。
  it('replaces the go-out button with a prompt when the hand cannot go out', async () => {
    mockExec.mockResolvedValue(makeContinentalRummyState({ goOutIdx: -1 }));
    renderWithProviders(<ContinentalRummyPage />);
    expect(await screen.findByTestId('cont-discard-notice')).toBeInTheDocument();
    expect(screen.queryByTestId('cont-goout')).not.toBeInTheDocument();
  });

  // **引かずに上がるほうが重い (10 点 vs 7 点)。** 別のボタン・別の命令。
  it('offers going out on the deal, ahead of drawing', async () => {
    mockExec.mockResolvedValue(makeContinentalRummyState({ phase: 'draw', canGoOutOnDeal: true, goOutIdx: -1 }));
    renderWithProviders(<ContinentalRummyPage />);
    fireEvent.click(await screen.findByTestId('cont-goout-deal'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('gooutdeal'));
    expect(mockExec).not.toHaveBeenCalledWith('goout', expect.anything());
  });

  // 負のコントロール: 配られたままで上がれないなら出さない。
  it('hides the deal go-out when the dealt hand does not go out', async () => {
    renderWithProviders(<ContinentalRummyPage />);
    await screen.findByTestId('cont-layouts');
    expect(screen.queryByTestId('cont-goout-deal')).not.toBeInTheDocument();
  });

  it('discards the card that is clicked', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<ContinentalRummyPage />);
    const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { handIndex: 1 }));
  });

  // **加点は内訳で見せる。** 合計だけだと、どう上がると得なのかが伝わらない。
  it('breaks the settlement down and offers the next round', async () => {
    mockExec.mockResolvedValue(
      makeContinentalRummyState({
        phase: 'roundEnd',
        isHumanTurn: false,
        goOutIdx: -1,
        lastResult: {
          winnerIdx: 0,
          bonuses: [
            { key: 'win', points: 1 },
            { key: 'dealt', points: 10 },
          ],
          perOpponent: 11,
          total: 33,
        },
      }),
    );
    renderWithProviders(<ContinentalRummyPage />);
    expect(await screen.findByTestId('cont-bonus-win')).toHaveTextContent('1');
    expect(screen.getByTestId('cont-bonus-dealt')).toHaveTextContent('10');
    expect(screen.getByTestId('cont-collected')).toHaveTextContent('11');
    expect(screen.getByTestId('cont-collected')).toHaveTextContent('33');

    const nextButton = screen.getByRole('button', { name: '次のラウンド' });
    expect(nextButton.querySelector('kbd')).toHaveTextContent('N');
    fireEvent.click(nextButton);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports a washout without a bonus breakdown', async () => {
    mockExec.mockResolvedValue(
      makeContinentalRummyState({
        phase: 'roundEnd',
        isHumanTurn: false,
        goOutIdx: -1,
        lastResult: { winnerIdx: -1, bonuses: [], perOpponent: 0, total: 0 },
      }),
    );
    renderWithProviders(<ContinentalRummyPage />);
    expect(await screen.findByTestId('cont-result')).toHaveTextContent('山札が尽きました');
    expect(screen.queryByTestId('cont-collected')).not.toBeInTheDocument();
  });

  it('stops offering actions once the game is over', async () => {
    mockExec.mockResolvedValue(
      makeContinentalRummyState({
        phase: 'gameEnd',
        gameEndFlag: true,
        winnerIdx: 0,
        isHumanTurn: false,
        goOutIdx: -1,
      }),
    );
    renderWithProviders(<ContinentalRummyPage />);
    await screen.findByTestId('cont-layouts');
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '次のラウンド' })).not.toBeInTheDocument();
  });

  it('renders the wait notice when it is not the human turn', async () => {
    // 他のプレイヤーのターン中に待機メッセージを表示する
    mockExec.mockResolvedValue(makeContinentalRummyState({ isHumanTurn: false, phase: 'draw', gameEndFlag: false }));
    renderWithProviders(<ContinentalRummyPage />);
    await waitFor(() => expect(screen.getByTestId('cont-wait-notice')).toBeInTheDocument());
  });

  it('hides the wait notice when it is the human turn', async () => {
    // 自分のターン中は待機メッセージを表示しない
    mockExec.mockResolvedValue(makeContinentalRummyState({ isHumanTurn: true, phase: 'draw', gameEndFlag: false }));
    renderWithProviders(<ContinentalRummyPage />);
    await waitFor(() => expect(screen.queryByTestId('cont-wait-notice')).not.toBeInTheDocument());
  });
});
