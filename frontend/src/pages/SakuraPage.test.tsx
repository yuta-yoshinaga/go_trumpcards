import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sakuraApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSakuraState } from '../test/stateFactories';
import { SakuraPage } from './SakuraPage';

vi.mock('../api/gameApi', () => ({
  sakuraApi: { exec: vi.fn() },
  actionLogApi: { sakura: vi.fn() },
}));

const mockExec = vi.mocked(sakuraApi.exec);

const playState = makeSakuraState();
const roundEndState = makeSakuraState({
  phase: 1,
  isHumanTurn: false,
  lastResult: {
    round: 1,
    winner: 0,
    seats: [
      { cardPoints: 40, bonuses: [{ key: 'sakuraSake', points: 30 }], bonusPoints: 30, total: 70 },
      { cardPoints: 25, bonuses: [], bonusPoints: 0, total: 25 },
      { cardPoints: 12, bonuses: [], bonusPoints: 0, total: 12 },
    ],
  },
});
const gameEndState = makeSakuraState({
  phase: 2,
  gameEndFlag: true,
  isHumanTurn: false,
  winner: 0,
  message: 'ゲーム終了！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('SakuraPage', () => {
  // **点数を札そのものに書く** (#5785)。さくらは役ではなく点数の合計で競うので、
  // どの札が何点かが読めないと打ち手を決められない（CUI は最初からそう出している）。
  it('shows each card point value on the card', async () => {
    mockExec.mockResolvedValue(
      makeSakuraState({
        players: [
          {
            ...playState.players[0],
            cards: [{ ...playState.players[0].cards[0], points: 20 }],
            taken: [{ ...playState.players[0].cards[1], points: 1 }],
            takenCount: 1,
          },
          playState.players[1],
          playState.players[2],
        ],
        fieldCards: [{ ...playState.fieldCards[0], points: 10 }],
      }),
    );
    renderWithProviders(<SakuraPage />);

    const hand = await screen.findByTestId('hand-card-0');
    expect(within(hand).getByTestId('sakura-card-points')).toHaveTextContent('20');
    expect(within(screen.getByTestId('field-card-0')).getByTestId('sakura-card-points')).toHaveTextContent('10');
    // 取り札にも出る。
    expect(within(screen.getByTestId('sakura-human')).getAllByTestId('sakura-card-points')[0]).toHaveTextContent('1');
  });

  // **点数を持たない札には出さない。** 他のゲームの札に空のバッジを付けない。
  it('leaves the badge out when the card carries no points', async () => {
    renderWithProviders(<SakuraPage />);
    const hand = await screen.findByTestId('hand-card-0');
    expect(within(hand).queryByTestId('sakura-card-points')).not.toBeInTheDocument();
  });
  it('renders the loading fallback when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SakuraPage />);
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SakuraPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the round, stock and dealer', async () => {
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByText('ラウンド 1/3')).toBeInTheDocument();
    expect(screen.getByText('山札: 21')).toBeInTheDocument();
    expect(screen.getByText('親: あなた')).toBeInTheDocument();
  });

  it('plays a hand card with a single match immediately', async () => {
    renderWithProviders(<SakuraPage />);
    const card = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(card);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  // **2 枚一致のときだけ場札を選ばせる。** 選ばせずに送ると、どちらを取るかを
  // サーバの既定 (点数の高いほう) に任せることになり、選択そのものが消える。
  it('requires a field pick for a two-way match, then plays with fieldIndex', async () => {
    mockExec.mockResolvedValue(makeSakuraState({ captureOptions: { 0: [0, 1] }, choiceOptions: { 0: [0, 1] } }));
    renderWithProviders(<SakuraPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));

    expect(await screen.findByTestId('sakura-field-pick')).toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('field-card-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 1 }));
  });

  // **3 枚一致では選ばせない。** まとめて取るのでどれを押しても結果が変わらず、
  // 「選べ」と出すと効かない選択を求めることになる (#5338 のレビュー指摘)。
  it('does not ask for a pick when every matching field card is captured anyway', async () => {
    mockExec.mockResolvedValue(makeSakuraState({ captureOptions: { 0: [0, 1, 2] }, choiceOptions: {} }));
    renderWithProviders(<SakuraPage />);
    const card = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(card);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
    expect(screen.queryByTestId('sakura-field-pick')).not.toBeInTheDocument();
  });

  // 一致していない場札は押せない (押しても送らない)。
  it('ignores a field card that does not match the selected hand card', async () => {
    mockExec.mockResolvedValue(
      makeSakuraState({
        captureOptions: { 0: [0, 1] },
        choiceOptions: { 0: [0, 1] },
        fieldCards: playState.fieldCards,
      }),
    );
    renderWithProviders(<SakuraPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    await screen.findByTestId('sakura-field-pick');

    const candidates = await screen.findAllByTestId(/^field-card-/);
    // 候補は 2 枚とも押せる。候補でない札は disabled になる。
    for (const el of candidates) {
      expect(el).toHaveAttribute('data-capture-candidate', 'true');
    }
  });

  it('does not send a play while it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeSakuraState({ isHumanTurn: false, currentTurn: 1 }));
    renderWithProviders(<SakuraPage />);
    const card = await screen.findByTestId('hand-card-0');
    expect(card).toBeDisabled();

    mockExec.mockClear();
    fireEvent.click(card);
    await waitFor(() => expect(mockExec).not.toHaveBeenCalled());
  });

  it('names the seat that is thinking', async () => {
    mockExec.mockResolvedValue(makeSakuraState({ isHumanTurn: false, currentTurn: 2 }));
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByTestId('sakura-prompt')).toHaveTextContent('CPU2');
  });

  it('shows every opponent seat without their hands', async () => {
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByTestId('sakura-seat-1')).toBeInTheDocument();
    expect(screen.getByTestId('sakura-seat-2')).toBeInTheDocument();
    // 相手の手札はワイヤに乗らないので、手札のボタンは人間の 3 枚だけ。
    expect(screen.getAllByTestId(/^hand-card-/)).toHaveLength(3);
  });

  // 追加役と点数は、サーバが出した値をそのまま出す (画面で計算し直さない)。
  it('shows the human bonuses and points from the server', async () => {
    mockExec.mockResolvedValue(
      makeSakuraState({
        players: [
          {
            ...playState.players[0],
            takenCount: 2,
            taken: playState.fieldCards,
            cardPoints: 40,
            bonuses: [{ key: 'sakuraSake', points: 30 }],
            bonusPoints: 30,
            totalPoints: 70,
          },
          playState.players[1],
          playState.players[2],
        ],
      }),
    );
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByTestId('sakura-human-bonus')).toHaveTextContent('桜酒 (30)');
    expect(screen.getByTestId('sakura-human-points')).toHaveTextContent('70');
  });

  it('shows "no bonus" when none are held', async () => {
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByTestId('sakura-human-bonus')).toHaveTextContent('追加役なし');
  });

  it('shows the round result with each seat breakdown and advances', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SakuraPage />);

    const result = await screen.findByTestId('sakura-round-result');
    expect(result).toHaveTextContent('勝者: あなた');
    expect(screen.getByTestId('sakura-round-seat-0')).toHaveTextContent('札40点 + 追加役30点 = 70点');
    expect(screen.getByTestId('sakura-round-seat-2')).toHaveTextContent('12');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('sakura-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports a tied round', async () => {
    mockExec.mockResolvedValue(
      makeSakuraState({
        phase: 1,
        lastResult: { round: 1, winner: -1, seats: roundEndState.lastResult?.seats ?? [] },
      }),
    );
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByTestId('sakura-round-result')).toHaveTextContent('同点');
  });

  it('shows the final scores and restarts with the chosen settings', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SakuraPage />);

    expect(await screen.findByTestId('sakura-result')).toHaveTextContent('勝者: あなた');
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { seats: 3, rounds: 3 } }));
  });

  it('reports a tied game', async () => {
    mockExec.mockResolvedValue(makeSakuraState({ phase: 2, gameEndFlag: true, winner: -1 }));
    renderWithProviders(<SakuraPage />);
    expect(await screen.findByTestId('sakura-result')).toHaveTextContent('引き分け');
  });

  // 盤面が出たあとの失敗を出す。初回が落ちるとスケルトンのままなので、
  // そこだけ見ると「エラー表示がある」ことを確かめたことにならない。
  it('surfaces an API error raised after the board is up', async () => {
    renderWithProviders(<SakuraPage />);
    const card = await screen.findByTestId('hand-card-0');

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(card);
    // 表示されるのは共通の通信エラー文言なので、投げた文字列ではなく警告帯を見る。
    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });
});
