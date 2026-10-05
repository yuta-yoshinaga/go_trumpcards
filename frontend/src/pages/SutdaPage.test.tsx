import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sutdaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSutdaState } from '../test/stateFactories';
import { SutdaPage } from './SutdaPage';

vi.mock('../api/gameApi', () => ({
  sutdaApi: { exec: vi.fn() },
  actionLogApi: { sutda: vi.fn() },
}));

const mockExec = vi.mocked(sutdaApi.exec);

const betState = makeSutdaState();

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(betState);
});

describe('SutdaPage', () => {
  it('calls reset on mount with the configured table', async () => {
    renderWithProviders(<SutdaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, seats: 3, startChips: 1000 },
      }),
    );
  });

  it('shows the hand counter and the pot', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByText('ハンド 1')).toBeInTheDocument();
    expect(screen.getByTestId('sutda-pot')).toHaveTextContent('ポット 30');
  });

  it('does not show a previous-hand result on the first hand without a saved result', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ handNumber: 1, lastResult: null }));
    renderWithProviders(<SutdaPage />);
    await screen.findByText('ハンド 1');
    expect(screen.queryByTestId('sutda-result')).not.toBeInTheDocument();
  });

  it('pairs each seat’s remaining chips with its current contribution', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-chips-0')).toHaveTextContent('990');
    expect(screen.getByTestId('sutda-bet-0')).toHaveTextContent('10');
    expect(screen.getByTestId('sutda-table')).toHaveTextContent('残りチップ');
    expect(screen.getByTestId('sutda-table')).toHaveTextContent('投入額');
    expect(screen.getByTestId('sutda-pot')).toHaveTextContent('ポット 30');
  });

  it('groups each seat’s chips, contribution, dealer, and folded status under the player name', async () => {
    const base = makeSutdaState();
    const players = base.players.map((player, index) => ({
      ...player,
      isDealer: index === 1,
      folded: index === 1,
    }));
    mockExec.mockResolvedValue(makeSutdaState({ players }));
    renderWithProviders(<SutdaPage />);

    const seat = await screen.findByRole('region', { name: 'CPU 1' });
    expect(within(seat).getByTestId('sutda-chips-1')).toHaveTextContent('990');
    expect(within(seat).getByTestId('sutda-bet-1')).toHaveTextContent('10');
    expect(within(seat).getByText('親')).toBeInTheDocument();
    expect(within(seat).getByTestId('sutda-folded-1')).toHaveTextContent('降り');
  });

  // **自分の役は常に見える。** 伏せているのは相手の札だけ。
  it('always shows your own two cards and what they make', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-hand')).toHaveTextContent('38光ッタン');
  });

  it('offers the complete hand strength ranking', async () => {
    renderWithProviders(<SutdaPage />);
    const ranking = await screen.findByTestId('sutda-hand-ranking');
    expect(ranking).toHaveTextContent('役の強弱順');
    const rankingText = ranking.textContent ?? '';
    const handPositions = ['38光ッタン', 'アリ（1+2）', 'マントン（0ット）'].map((handName) =>
      rankingText.indexOf(handName),
    );
    expect(handPositions.every((position) => position >= 0)).toBe(true);
    expect(handPositions[0]).toBeLessThan(handPositions[1]);
    expect(handPositions[1]).toBeLessThan(handPositions[2]);
  });

  it('keeps opponents face down until they are revealed', async () => {
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-hand');
    expect(screen.queryByTestId('sutda-cards-1')).not.toBeInTheDocument();
    expect(screen.getByTestId('sutda-cards-0')).toBeInTheDocument();
  });

  it('shows an opponent hand once it is revealed', async () => {
    const base = makeSutdaState();
    mockExec.mockResolvedValue(
      makeSutdaState({
        players: base.players.map((p, i) =>
          i === 1 ? { ...p, revealed: true, cards: base.players[0].cards, handName: 'mangtong', handRank: 600 } : p,
        ),
      }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-cards-1')).toHaveTextContent('マントン');
  });

  // 差額 0 のときはチェック、要るときはコール。
  it('labels the call button by what is owed', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-call')).toHaveTextContent('チェック');
    expect(screen.getByTestId('sutda-to-call')).toHaveTextContent('追加は不要');

    mockExec.mockResolvedValue(makeSutdaState({ callAmount: 20 }));
    fireEvent.click(screen.getByTestId('sutda-call'));
    await waitFor(() => expect(screen.getByTestId('sutda-call')).toHaveTextContent('コール（20）'));
    expect(screen.getByTestId('sutda-to-call')).toHaveTextContent('20');
  });

  it.each([
    ['sutda-call', 'call'],
    ['sutda-raise', 'raise'],
    ['sutda-fold', 'fold'],
  ])('%s sends %s', async (testId, command) => {
    renderWithProviders(<SutdaPage />);
    fireEvent.click(await screen.findByTestId(testId));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  // **押せる条件はサーバの canRaise がすべて。** 上限とチップの両方を見た
  // 結果なので、画面側で組み直すと判断が食い違う。
  it('hides raise when the server says it is not available', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ canRaise: false }));
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-call');
    expect(screen.queryByTestId('sutda-raise')).not.toBeInTheDocument();
  });

  it('offers no betting buttons when it is not your turn', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ isHumanTurn: false, currentPlayerIdx: 1 }));
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-hand');
    expect(screen.queryByTestId('sutda-call')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sutda-fold')).not.toBeInTheDocument();
  });

  it('marks a folded seat', async () => {
    const base = makeSutdaState();
    mockExec.mockResolvedValue(
      makeSutdaState({ players: base.players.map((p, i) => (i === 1 ? { ...p, folded: true } : p)) }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-folded-1')).toBeInTheDocument();
    expect(screen.queryByTestId('sutda-folded-0')).not.toBeInTheDocument();
  });

  it('shows the showdown result and advances the hand', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({
        phase: 'showdown',
        isShowdown: true,
        isHumanTurn: false,
        lastResult: {
          winners: [0],
          shares: [70],
          pot: 70,
          handNames: ['gwang38', 'mangtong', 'kkeut5'],
          folded: [false, false, false],
        },
      }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-result')).toHaveTextContent('あなた（70）');
    expect(screen.getByTestId('sutda-result')).toHaveTextContent('合計 70');
    fireEvent.click(screen.getByTestId('sutda-next-hand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nexthand'));
  });

  it('keeps the previous hand number, winners, shares, and winning hands visible during the next hand', async () => {
    const result = {
      winners: [0, 1],
      shares: [26, 25],
      pot: 51,
      handNames: ['ali', 'ali', 'kkeut5'],
      folded: [false, false, false],
    };
    mockExec
      .mockResolvedValueOnce(
        makeSutdaState({
          phase: 'showdown',
          isShowdown: true,
          isHumanTurn: false,
          lastResult: result,
        }),
      )
      .mockResolvedValueOnce(
        makeSutdaState({
          phase: 'bet',
          handNumber: 3,
          lastResult: result,
        }),
      );
    renderWithProviders(<SutdaPage />);
    fireEvent.click(await screen.findByTestId('sutda-next-hand'));
    await waitFor(() => expect(screen.getByText('ハンド 3')).toBeInTheDocument());
    const resultPanel = await screen.findByTestId('sutda-result');
    expect(resultPanel).toHaveTextContent('ハンド 2');
    expect(resultPanel).toHaveTextContent('あなた（26）');
    expect(resultPanel).toHaveTextContent('CPU 1（25）');
    expect(resultPanel).toHaveTextContent('あなた（26）: アリ（1+2）');
    expect(resultPanel).toHaveTextContent('CPU 1（25）: アリ（1+2）');
    expect(resultPanel).toHaveTextContent('合計 51');
  });

  it('shows the winner name without an amount when saved shares are missing', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({
        phase: 'showdown',
        isShowdown: true,
        isHumanTurn: false,
        lastResult: {
          winners: [0],
          shares: [],
          pot: 70,
          handNames: ['gwang38', 'mangtong', 'kkeut5'],
          folded: [false, false, false],
        },
      }),
    );
    renderWithProviders(<SutdaPage />);
    const result = await screen.findByTestId('sutda-result');
    expect(result).toHaveTextContent('あなた');
    expect(result).not.toHaveTextContent('（0）');
  });

  it('keeps the old winner display when the saved hand name is empty', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({
        phase: 'showdown',
        isShowdown: true,
        isHumanTurn: false,
        lastResult: {
          winners: [0],
          shares: [26],
          pot: 26,
          handNames: ['', 'mangtong', 'kkeut5'],
          folded: [false, false, false],
        },
      }),
    );
    renderWithProviders(<SutdaPage />);
    const result = await screen.findByTestId('sutda-result');
    expect(result).toHaveTextContent('あなた（26）');
    expect(result).not.toHaveTextContent('あなた（26）:');
  });

  it('shows only the winner name when both the hand name and saved share are missing', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({
        phase: 'showdown',
        isShowdown: true,
        isHumanTurn: false,
        lastResult: {
          winners: [0],
          shares: [],
          pot: 26,
          handNames: ['', 'mangtong', 'kkeut5'],
          folded: [false, false, false],
        },
      }),
    );
    renderWithProviders(<SutdaPage />);
    const result = await screen.findByTestId('sutda-result');
    expect(result).toHaveTextContent('あなた');
    expect(result).not.toHaveTextContent('あなた:');
    expect(result).not.toHaveTextContent('あなた（');
  });

  it('shows each split-pot share and gives the odd chip to the first winner', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({
        phase: 'showdown',
        isShowdown: true,
        isHumanTurn: false,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [],
            chips: 1026,
            bet: 0,
            folded: false,
            revealed: true,
            handName: 'ali',
            handRank: 10,
            isDealer: false,
          },
          {
            id: 1,
            isHuman: false,
            cardCount: 2,
            cards: [],
            chips: 1025,
            bet: 0,
            folded: false,
            revealed: true,
            handName: 'ali',
            handRank: 10,
            isDealer: false,
          },
          {
            id: 2,
            isHuman: false,
            cardCount: 2,
            cards: [],
            chips: 900,
            bet: 0,
            folded: false,
            revealed: true,
            handName: 'kkeut5',
            handRank: 5,
            isDealer: true,
          },
        ],
        lastResult: {
          winners: [0, 1],
          shares: [26, 25],
          pot: 51,
          handNames: ['ali', 'ali', 'kkeut5'],
          folded: [false, false, false],
        },
      }),
    );
    renderWithProviders(<SutdaPage />);
    const result = await screen.findByTestId('sutda-result');
    expect(result).toHaveTextContent('あなた（26）: アリ（1+2）、CPU 1（25）: アリ（1+2）');
    expect(result).toHaveTextContent('合計 51');
    expect(screen.getByTestId('sutda-chips-0')).toHaveTextContent('1026');
    expect(screen.getByTestId('sutda-chips-1')).toHaveTextContent('1025');
  });

  it('names the winner at the end of the table', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({ phase: 'gameEnd', gameEndFlag: true, isHumanTurn: false, winnerIdx: 2 }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-winner')).toHaveTextContent('CPU 2');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ hintAction: 'raise', hintReason: 'strong_hand', messageCode: '' }));
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-hand');
    expect(screen.queryByText(/強い役なので/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({ hintAction: 'raise', hintReason: 'strong_hand', messageCode: 'sutda.hintRequested' }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByText(/強い役なので/)).toBeInTheDocument();
  });

  // **あと何回レイズできるかは Web だけ出ていなかった (#6627)。** サーバは
  // raiseCount / maxRaises を最初から送っていて、CUI の canRaise 行は
  // `SutdaMaxRaises - GetRaiseCount()` を出しているのに、Web のボタンは
  // どちらのフィールドも読んでいなかった。
  describe('remaining raises', () => {
    it('shows how many raises are left on the button', async () => {
      mockExec.mockResolvedValue(makeSutdaState({ canRaise: true, raiseCount: 0, maxRaises: 3 }));
      renderWithProviders(<SutdaPage />);
      const btn = await screen.findByTestId('sutda-raise');
      expect(btn.textContent).toContain('あと 3 回');
      expect(btn.textContent).not.toContain('{{');
    });

    // **数はレイズごとに減る。** 定数を出しているだけの実装だとここで落ちる。
    it('counts down as raises are used', async () => {
      mockExec.mockResolvedValue(makeSutdaState({ canRaise: true, raiseCount: 2, maxRaises: 3 }));
      renderWithProviders(<SutdaPage />);
      const btn = await screen.findByTestId('sutda-raise');
      expect(btn.textContent).toContain('あと 1 回');
      expect(btn.textContent).not.toContain('あと 3 回');
    });

    // **上限に達したらボタンごと消える** (受け入れ条件「消える直前まで一致する」)。
    // 直前は 1 回、達したら 0 ではなくボタンが無い、という形。
    it('drops the button entirely once the cap is reached', async () => {
      mockExec.mockResolvedValue(makeSutdaState({ canRaise: false, raiseCount: 3, maxRaises: 3 }));
      renderWithProviders(<SutdaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('sutda-raise')).not.toBeInTheDocument();
    });
  });
});
