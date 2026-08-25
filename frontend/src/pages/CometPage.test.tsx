import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cometApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCometState } from '../test/stateFactories';
import { CometPage } from './CometPage';

vi.mock('../api/gameApi', () => ({
  cometApi: { exec: vi.fn() },
  actionLogApi: { comet: vi.fn() },
}));

const mockExec = vi.mocked(cometApi.exec);

const playState = makeCometState();

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('CometPage', () => {
  it('calls reset on mount with the configured table', async () => {
    renderWithProviders(<CometPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, players: 4, targetScore: 100 },
      }),
    );
  });

  it('shows the round and the dead hand', async () => {
    renderWithProviders(<CometPage />);
    expect(await screen.findByText('局 1（100 点勝負）')).toBeInTheDocument();
    // **死に手の枚数は見せる。** ここに眠った札で連なりが止まる。
    expect(screen.getByTestId('comet-dead')).toHaveTextContent('3');
  });

  // **先頭は何でも出せると書く。** 「次に要るランク」と別の文言でないと、
  // 何を出してよいのか分からない。
  it('says the lead is free, then names the rank needed', async () => {
    renderWithProviders(<CometPage />);
    expect(await screen.findByTestId('comet-need')).toHaveTextContent('好きな札');

    mockExec.mockResolvedValue(makeCometState({ need: 8, playableIdxs: [0] }));
    renderWithProviders(<CometPage />);
    await waitFor(() => {
      const needs = screen.getAllByTestId('comet-need');
      expect(needs[needs.length - 1]).toHaveTextContent('8');
    });
  });

  it('shows the sequence, empty at first', async () => {
    renderWithProviders(<CometPage />);
    expect(await screen.findByTestId('comet-pile')).toBeInTheDocument();

    mockExec.mockResolvedValue(
      makeCometState({
        pile: [
          { design: 'SPADE', value: 5, color: 'black' },
          { design: 'HEART', value: 6, color: 'red' },
        ],
        need: 7,
      }),
    );
    renderWithProviders(<CometPage />);
    await waitFor(() => {
      const piles = screen.getAllByTestId('comet-pile');
      expect(piles[piles.length - 1].children).toHaveLength(2);
    });
  });

  it('plays the card that is clicked', async () => {
    renderWithProviders(<CometPage />);
    const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
  });

  // **パスは出せる札が無いときだけ。** 常に出しておくと、出せるのに押して
  // 弾かれるだけのボタンになる。
  it('offers pass only when nothing is playable', async () => {
    renderWithProviders(<CometPage />);
    await screen.findByTestId('comet-pile');
    expect(screen.queryByTestId('comet-pass')).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeCometState({ need: 8, playableIdxs: [] }));
    renderWithProviders(<CometPage />);
    const pass = await screen.findByTestId('comet-pass');
    fireEvent.click(pass);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('shows every seat with its card count and score', async () => {
    renderWithProviders(<CometPage />);
    const scores = await screen.findByTestId('comet-scores');
    expect(scores).toHaveTextContent('手札 3 枚');
    expect(scores).toHaveTextContent('得点 0');
  });

  it('shows the round result and advances', async () => {
    mockExec.mockResolvedValue(
      makeCometState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastResult: {
          winnerIdx: 0,
          cardsLeft: [0, 2, 3, 1],
          gained: [13, 0, 0, 0],
          unplayedKings: 2,
          heldWildIdx: 2,
        },
      }),
    );
    renderWithProviders(<CometPage />);
    const result = await screen.findByTestId('comet-round-result');
    expect(result).toHaveTextContent('13');
    // **出なかった K は取り分の一部なので、内訳に出す。**
    expect(result).toHaveTextContent('2');
    expect(screen.getByTestId('comet-held-wild')).toBeInTheDocument();

    fireEvent.click(screen.getByTestId('comet-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('omits the comet penalty line when nobody held it', async () => {
    mockExec.mockResolvedValue(
      makeCometState({
        phase: 'roundEnd',
        isHumanTurn: false,
        lastResult: { winnerIdx: 0, cardsLeft: [0, 1, 1, 1], gained: [4, 0, 0, 0], unplayedKings: 0, heldWildIdx: -1 },
      }),
    );
    renderWithProviders(<CometPage />);
    await screen.findByTestId('comet-round-result');
    expect(screen.queryByTestId('comet-held-wild')).not.toBeInTheDocument();
  });

  it('shows the winner at game end', async () => {
    mockExec.mockResolvedValue(makeCometState({ phase: 'gameEnd', gameEndFlag: true, winnerIdx: 0 }));
    renderWithProviders(<CometPage />);
    expect(await screen.findByTestId('comet-winner')).toBeInTheDocument();
  });

  // **ヒントのライブ領域は常設。** 出る側と出ない側の両方を見る。
  it('announces a requested hint, and stays quiet otherwise', async () => {
    renderWithProviders(<CometPage />);
    expect(await screen.findByTestId('comet-hint-live')).toBeEmptyDOMElement();

    mockExec.mockResolvedValue(
      makeCometState({ messageCode: 'comet.hintRequested', hintHandIdx: 2, hintReason: 'comet' }),
    );
    renderWithProviders(<CometPage />);
    await waitFor(() => {
      const lives = screen.getAllByTestId('comet-hint-live');
      expect(lives[lives.length - 1]).not.toBeEmptyDOMElement();
    });
  });
});
