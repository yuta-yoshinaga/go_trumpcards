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
    fireEvent.click(await screen.findByRole('button', { name: '山札から引く' }));
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

    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
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
});
