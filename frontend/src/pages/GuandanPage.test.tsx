import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { guandanApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, GuandanPlayer, GuandanResponse } from '../types/card';
import { GuandanPhase } from '../types/phases';
import { GuandanPage } from './GuandanPage';

vi.mock('../api/gameApi', () => ({
  guandanApi: { exec: vi.fn() },
  actionLogApi: { guandan: vi.fn() },
}));

const mockExec = vi.mocked(guandanApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<GuandanPlayer>): GuandanPlayer {
  return {
    id,
    isHuman,
    team: id % 2,
    cardCount: 27,
    cards: isHuman ? [card('SPADE', 2), card('HEART', 5), card('CLOVER', 13)] : [],
    finishedRank: 0,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<GuandanResponse>): GuandanResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: GuandanPhase.PLAY,
    handNumber: 1,
    currentPlayerIdx: 0,
    level: 2,
    teamLevels: [2, 2],
    declarerTeam: 0,
    lastCombo: null,
    lastPlayerIdx: -1,
    finished: [],
    tributes: [],
    tributeCancelled: false,
    lastResult: null,
    minLevel: 2,
    maxLevel: 14,
    advanceFirstSecond: 4,
    advanceFirstThird: 2,
    advanceFirstFourth: 1,
    gameEndFlag: false,
    winnerTeam: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

describe('GuandanPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **レベル札が A より強いことがこのゲームの肝。**画面に書いていないと読めない。
  it('says the level cards beat aces and which of them are wild', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('guandan-level-note')).toBeInTheDocument());
    const note = screen.getByTestId('guandan-level-note');
    expect(note).toHaveTextContent('Aより強く');
    expect(note).toHaveTextContent('♥の2枚はワイルド');
  });

  // **昇級量は 1 / 2 / 4。**3 段階は存在しない。
  it('states the advance table, including that three levels is impossible', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('guandan-advance-note')).toBeInTheDocument());
    const note = screen.getByTestId('guandan-advance-note');
    expect(note).toHaveTextContent('1着2着の独占で4段階');
    expect(note).toHaveTextContent('1着3着で2段階');
    expect(note).toHaveTextContent('1着4着で1段階');
    expect(note).toHaveTextContent('3段階上がることはありません');
  });

  // **レベルは 2〜A。**数字のままでは読めない。
  it('labels the face levels with letters', async () => {
    mockExec.mockResolvedValue(makeState({ level: 14, teamLevels: [14, 11] }));
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('guandan-info')).toBeInTheDocument());
    const info = screen.getByTestId('guandan-info');
    expect(info).toHaveTextContent('この局のレベル: A');
    expect(info).toHaveTextContent('チーム1: J');
  });

  it('shows every seat with its team and card count', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getAllByTestId('guandan-player')).toHaveLength(4));
    // **パートナーは向かい合わせ。**席 0/2 が同じチーム。
    const players = screen.getAllByTestId('guandan-player');
    expect(players[0]).toHaveTextContent('チーム0');
    expect(players[1]).toHaveTextContent('チーム1');
    expect(players[2]).toHaveTextContent('チーム0');
  });

  it('shows the finishing position once a seat is out', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, true, { finishedRank: 2 }), seat(1, false), seat(2, false), seat(3, false)] }),
    );
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getAllByTestId('guandan-player')[0]).toHaveTextContent('着順2'));
  });

  it('reports the table, empty or otherwise', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('guandan-table')).toHaveTextContent('流れています'));

    mockExec.mockResolvedValue(makeState({ lastCombo: { kind: 8, rank: 9, size: 4 }, lastPlayerIdx: 2 }));
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getAllByTestId('guandan-table')[1]).toHaveTextContent('ボム'));
  });

  it('plays the selected cards as one combination', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-2'));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndexes: [0, 2] }));
  });

  // **選び直せる。**一度選んだ札を外せないと役が組めない。
  it('deselects a card that is clicked twice', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByTestId('hand-card-1'));
    fireEvent.click(screen.getByTestId('hand-card-0'));
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndexes: [1] }));
  });

  it('cannot play with nothing selected', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeDisabled());
  });

  // **場が流れているときはパスできない。**リードは必ず何か出す。
  it('disables pass while the table is clear and enables it once a combination is down', async () => {
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeDisabled());

    mockExec.mockResolvedValue(makeState({ lastCombo: { kind: 2, rank: 5, size: 2 }, lastPlayerIdx: 3 }));
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getAllByRole('button', { name: 'パス' })[1]).toBeEnabled());

    fireEvent.click(screen.getAllByRole('button', { name: 'パス' })[1] as HTMLElement);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  describe('tribute', () => {
    const tributeState = (overrides?: Partial<GuandanResponse>) =>
      makeState({
        phase: GuandanPhase.TRIBUTE,
        handNumber: 2,
        tributes: [{ from: 3, to: 0, card: card('SPADE', 1), returned: null }],
        ...overrides,
      });

    it('returns the single selected card', async () => {
      mockExec.mockResolvedValue(tributeState());
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByTestId('guandan-tributes')).toBeInTheDocument());

      fireEvent.click(screen.getByTestId('hand-card-1'));
      fireEvent.click(screen.getByRole('button', { name: '還貢する' }));

      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('tribute', { cardIndex: 1 }));
    });

    // **返すのは 1 枚だけ。**複数選んだまま送れてはいけない。
    it('will not return two cards at once', async () => {
      mockExec.mockResolvedValue(tributeState());
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());

      fireEvent.click(screen.getByTestId('hand-card-0'));
      fireEvent.click(screen.getByTestId('hand-card-1'));
      expect(screen.getByRole('button', { name: '還貢する' })).toBeDisabled();
    });

    // **赤ジョーカー 2 枚で貢は流れる。**理由が出ないと不可解に見える。
    it('explains a cancelled tribute', async () => {
      mockExec.mockResolvedValue(tributeState({ tributes: [], tributeCancelled: true }));
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByTestId('guandan-tribute-cancelled')).toHaveTextContent('抗貢'));
      expect(screen.getByTestId('guandan-tribute-cancelled')).toHaveTextContent('赤ジョーカー2枚');
    });

    // **還貢を負っていない席には操作が出ない。**
    it('offers no return when this seat owes none', async () => {
      mockExec.mockResolvedValue(
        tributeState({ tributes: [{ from: 2, to: 1, card: card('SPADE', 1), returned: null }] }),
      );
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByTestId('guandan-tributes')).toBeInTheDocument());
      expect(screen.queryByRole('button', { name: '還貢する' })).not.toBeInTheDocument();
    });
  });

  describe('hand end', () => {
    it('calls out first-and-second separately from an ordinary advance', async () => {
      mockExec.mockResolvedValue(
        makeState({
          phase: GuandanPhase.HAND_END,
          teamLevels: [6, 2],
          lastResult: { order: [0, 2, 1, 3], winnerTeam: 0, advance: 4, firstSecond: true },
        }),
      );
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByTestId('guandan-hand-result')).toHaveTextContent('1着2着の独占'));
      expect(screen.getByTestId('guandan-hand-result')).toHaveTextContent('4段階昇級');
    });

    it('shows a plain advance without the first-and-second banner', async () => {
      mockExec.mockResolvedValue(
        makeState({
          phase: GuandanPhase.HAND_END,
          teamLevels: [2, 3],
          lastResult: { order: [1, 0, 2, 3], winnerTeam: 1, advance: 1, firstSecond: false },
        }),
      );
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByTestId('guandan-hand-result')).toHaveTextContent('1段階昇級'));
      expect(screen.getByTestId('guandan-hand-result')).not.toHaveTextContent('独占');
    });

    it('deals the next hand', async () => {
      mockExec.mockResolvedValue(
        makeState({
          phase: GuandanPhase.HAND_END,
          lastResult: { order: [0, 2, 1, 3], winnerTeam: 0, advance: 4, firstSecond: true },
        }),
      );
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());
      fireEvent.click(screen.getByRole('button', { name: '次の局へ' }));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    });

    // **局の終わりに札は出せない。**
    it('offers no play controls once the hand is settled', async () => {
      mockExec.mockResolvedValue(makeState({ phase: GuandanPhase.HAND_END }));
      renderWithProviders(<GuandanPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: '次の局へ' })).toBeInTheDocument());
      expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    });
  });

  // **勝敗はチームで決まる。**人間は席 0 = チーム 0。
  it('announces the result by team', async () => {
    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: GuandanPhase.GAME_END, winnerTeam: 0 }));
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByText('あなたのチームの勝利です！')).toBeInTheDocument());

    mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: GuandanPhase.GAME_END, winnerTeam: 1 }));
    renderWithProviders(<GuandanPage />);
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
    renderWithProviders(<GuandanPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeDisabled());
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<GuandanPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
