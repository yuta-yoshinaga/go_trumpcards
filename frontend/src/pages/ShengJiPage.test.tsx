import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { shengjiApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, ShengJiPlayer, ShengJiResponse } from '../types/card';
import { ShengJiPhase } from '../types/phases';
import { ShengJiPage } from './ShengJiPage';

vi.mock('../api/gameApi', () => ({
  shengjiApi: { exec: vi.fn() },
  actionLogApi: { shengji: vi.fn() },
}));

const mockExec = vi.mocked(shengjiApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<ShengJiPlayer>): ShengJiPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 25,
    cards: isHuman
      ? [
          card('SPADE', 2),
          card('HEART', 5),
          card('CLOVER', 13),
          card('DIAMOND', 9),
          card('SPADE', 7),
          card('HEART', 3),
          card('CLOVER', 4),
          card('DIAMOND', 6),
          card('SPADE', 10),
        ]
      : [],
    isDeclarer: id % 2 === 0,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ShengJiResponse>): ShengJiResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: ShengJiPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    level: 5,
    teamLevels: [5, 2],
    declarerTeam: 0,
    trumpSuit: 1,
    declaration: null,
    declarableSuits: {},
    kittySize: 8,
    kitty: [],
    trick: [],
    trickLeader: 0,
    leadCombo: null,
    teamPoints: [0, 35],
    trickCount: 4,
    lastTrickWinner: 2,
    lastResult: null,
    minLevel: 2,
    maxLevel: 14,
    kittySizeMax: 8,
    totalPoints: 200,
    defenderTarget: 80,
    advanceStep: 40,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

describe('ShengJiPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **切札は切札スートだけではない。**これが読めないと序列が分からない。
  it('says the trump group is more than the trump suit', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('shengji-trump-note')).toBeInTheDocument());
    const note = screen.getByTestId('shengji-trump-note');
    expect(note).toHaveTextContent('切札は切札スートだけではありません');
    expect(note).toHaveTextContent('全スートの5（レベル札）とジョーカー4枚も切札');
  });

  // **点を集めるのは守備側。**
  it('says the defenders collect the points, and shows their total', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('shengji-points-note')).toBeInTheDocument());
    const note = screen.getByTestId('shengji-points-note');
    expect(note).toHaveTextContent('点を集めるのは守備側（チーム1）');
    expect(note).toHaveTextContent('現在 35 点');
    expect(note).toHaveTextContent('80 点で宣言側が交代');
  });

  it('labels the face levels with letters', async () => {
    mockExec.mockResolvedValue(makeState({ level: 14, teamLevels: [14, 11] }));
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('shengji-info')).toHaveTextContent('この局のレベル: A'));
    expect(screen.getByTestId('shengji-info')).toHaveTextContent('チーム1: J');
  });

  it('shows which side each seat is on', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getAllByTestId('shengji-player')).toHaveLength(4));
    const players = screen.getAllByTestId('shengji-player');
    expect(players[0]).toHaveTextContent('宣言側');
    expect(players[1]).toHaveTextContent('守備側');
    expect(players[2]).toHaveTextContent('宣言側');
  });

  describe('declaring', () => {
    const declareState = (overrides?: Partial<ShengJiResponse>) =>
      makeState({ phase: ShengJiPhase.DECLARE, declarableSuits: { '3': 2, '2': 1 }, ...overrides });

    it('offers only the suits you can declare', async () => {
      mockExec.mockResolvedValue(declareState());
      renderWithProviders(<ShengJiPage />);
      await waitFor(() => expect(screen.getByTestId('shengji-declare-3')).toBeInTheDocument());
      expect(screen.getByTestId('shengji-declare-2')).toBeInTheDocument();
      expect(screen.queryByTestId('shengji-declare-1')).not.toBeInTheDocument();
      expect(screen.getByTestId('shengji-declare-3')).toHaveTextContent('♥');
    });

    it('declares the chosen suit', async () => {
      mockExec.mockResolvedValue(declareState());
      renderWithProviders(<ShengJiPage />);
      await waitFor(() => expect(screen.getByTestId('shengji-declare-3')).toBeInTheDocument());
      fireEvent.click(screen.getByTestId('shengji-declare-3'));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', { suit: 3 }));
    });

    // **0 はパス。**宣言できる札が無くても手番は進められなければならない。
    it('can always pass, even holding nothing to declare', async () => {
      mockExec.mockResolvedValue(declareState({ declarableSuits: {} }));
      renderWithProviders(<ShengJiPage />);
      await waitFor(() => expect(screen.getByTestId('shengji-pass')).toBeInTheDocument());
      fireEvent.click(screen.getByTestId('shengji-pass'));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', { suit: 0 }));
    });

    it('explains that only a stronger showing overrides', async () => {
      mockExec.mockResolvedValue(declareState());
      renderWithProviders(<ShengJiPage />);
      await waitFor(() =>
        expect(screen.getByTestId('shengji-declare-rules')).toHaveTextContent('強い宣言だけが上書き'),
      );
    });
  });

  describe('the kitty', () => {
    const kittyState = () => makeState({ phase: ShengJiPhase.KITTY });

    // **底牌はちょうど 8 枚。**
    it('buries only once exactly eight are picked', async () => {
      mockExec.mockResolvedValue(kittyState());
      renderWithProviders(<ShengJiPage />);
      await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

      const bury = screen.getByRole('button', { name: /底牌に埋める/ });
      expect(bury).toBeDisabled();
      for (let i = 0; i < 7; i++) {
        fireEvent.click(screen.getByTestId(`hand-card-${i}`));
      }
      expect(screen.getByRole('button', { name: /底牌に埋める/ })).toBeDisabled();

      fireEvent.click(screen.getByTestId('hand-card-7'));
      fireEvent.click(screen.getByRole('button', { name: /底牌に埋める/ }));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bury', { cardIndexes: [0, 1, 2, 3, 4, 5, 6, 7] }));
    });

    it('warns against burying points and trumps', async () => {
      mockExec.mockResolvedValue(kittyState());
      renderWithProviders(<ShengJiPage />);
      await waitFor(() =>
        expect(screen.getByTestId('shengji-kitty-rules')).toHaveTextContent('得点札と切札は埋めないでください'),
      );
    });
  });

  it('plays the selected cards', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-2'));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndexes: [0, 2] }));
  });

  // **選び直せる。**一度選んだ札を外せないと手が組めない。
  it('deselects a card that is clicked twice', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndexes: [1] }));
  });

  it('cannot play with nothing selected', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  it('reports the trick, empty or otherwise', async () => {
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('shengji-trick')).toHaveTextContent('まだ誰も出していません'));

    mockExec.mockResolvedValue(
      makeState({
        trick: [{ seat: 1, cards: [card('HEART', 7), card('HEART', 7)] }],
        leadCombo: { kind: 2, rank: 7, size: 2, trump: false, suit: 3 },
      }),
    );
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getAllByTestId('shengji-trick')[1]).toHaveTextContent('対子'));
  });

  describe('hand end', () => {
    it('tells a held hand from a taken one', async () => {
      mockExec.mockResolvedValue(
        makeState({
          phase: ShengJiPhase.HAND_END,
          lastResult: {
            declarerTeam: 0,
            defenderPoints: 35,
            kittyPoints: 0,
            kittyMultiplier: 0,
            declarerHeld: true,
            advance: 2,
            advancingTeam: 0,
          },
        }),
      );
      renderWithProviders(<ShengJiPage />);
      await waitFor(() =>
        expect(screen.getByTestId('shengji-hand-result')).toHaveTextContent('宣言側が守りきりました'),
      );
      // **底牌の倍率は最終トリックを取った側にしか掛からない。**
      expect(screen.queryByTestId('shengji-kitty-line')).not.toBeInTheDocument();
    });

    it('shows the kitty multiplier when it applied', async () => {
      mockExec.mockResolvedValue(
        makeState({
          phase: ShengJiPhase.HAND_END,
          lastResult: {
            declarerTeam: 0,
            defenderPoints: 120,
            kittyPoints: 40,
            kittyMultiplier: 4,
            declarerHeld: false,
            advance: 1,
            advancingTeam: 1,
          },
        }),
      );
      renderWithProviders(<ShengJiPage />);
      await waitFor(() =>
        expect(screen.getByTestId('shengji-hand-result')).toHaveTextContent('守備側が 120 点を集めました'),
      );
      expect(screen.getByTestId('shengji-kitty-line')).toHaveTextContent('倍率 x4');
    });

    it('deals the next hand', async () => {
      mockExec.mockResolvedValue(makeState({ phase: ShengJiPhase.HAND_END }));
      renderWithProviders(<ShengJiPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());
      fireEvent.click(screen.getByRole('button', { name: '次の局へ' }));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    });

    it('offers no play controls once the hand is settled', async () => {
      mockExec.mockResolvedValue(makeState({ phase: ShengJiPhase.HAND_END }));
      renderWithProviders(<ShengJiPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());
      expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    });
  });

  // **勝敗はチームで決まる。**人間は席 0 = チーム 0。
  it('announces the result by team', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: ShengJiPhase.GAME_END, winnerTeam: 0 }));
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByText('あなたのチームの勝利です！')).toBeInTheDocument());

    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: ShengJiPhase.GAME_END, winnerTeam: 1 }));
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByText('相手チームの勝利です。')).toBeInTheDocument());
  });

  it('cannot select cards when it is not your turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentPlayerIdx: 1,
        players: [
          seat(0, true, { isCurrentTurn: false }),
          seat(1, false, { isCurrentTurn: true }),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeDisabled());
  });
});

// **ヒントは前から算出されていたのに、ページが一度も読んでいなかった (#4774)。**
// getShengJiHint も hintFactories への登録もあるのに、画面にトグルもツール
// チップも無かった。check-hint-coverage はファクトリの有無しか見ないので
// CI をすり抜けていた。
describe('ShengJiPage hint', () => {
  it('shows the hint once the toggle is enabled', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ShengJiPhase.DECLARE, declarableSuits: { '3': 2 } }));
    renderWithProviders(<ShengJiPage />);
    await waitFor(() => expect(screen.getByLabelText('ヒント表示')).toBeInTheDocument());

    // トグルを入れるまでは出さない。
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(screen.getByLabelText('ヒント表示'));
    expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
