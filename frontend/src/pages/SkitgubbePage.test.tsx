import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, skitgubbeApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SkitgubbePlayer, SkitgubbeResponse } from '../types/card';
import { SkitgubbePage } from './SkitgubbePage';

vi.mock('../api/gameApi', () => ({
  skitgubbeApi: { exec: vi.fn() },
  actionLogApi: { skitgubbe: vi.fn() },
}));

const mockExec = vi.mocked(skitgubbeApi.exec);
const mockActionLog = vi.mocked(actionLogApi.skitgubbe);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<SkitgubbePlayer>): SkitgubbePlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 1), card('HEART', 9), card('CLOVER', 13)],
    collectedCount: 4,
    finished: false,
    hidden: false,
    ...overrides,
  };
}

function cpu(id: number, overrides?: Partial<SkitgubbePlayer>): SkitgubbePlayer {
  // A hidden seat arrives with a count and NO hand cards.
  return {
    id,
    isHuman: false,
    cardCount: 3,
    cards: [],
    collectedCount: 2,
    finished: false,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<SkitgubbeResponse>): SkitgubbeResponse {
  return {
    players: [human(), cpu(1), cpu(2)],
    phase: 0,
    currentPlayerIdx: 0,
    stockCount: 37,
    trumpSuit: -1,
    duel: [card('SPADE', 9)],
    duelLeader: 1,
    pile: [],
    validIndices: [0, 1, 2],
    canPickUp: false,
    gameEndFlag: false,
    loserIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('SkitgubbePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('lets players open the action log during play', async () => {
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const viewLog = screen.getByRole('button', { name: '棋譜を見る' });
    mockActionLog.mockResolvedValueOnce({ entries: [] });
    fireEvent.click(viewLog);

    await waitFor(() => expect(mockActionLog).toHaveBeenCalled());
    expect(screen.getByRole('button', { name: '閉じる' })).toBeInTheDocument();
  });

  it('shows the collecting objective and rule without the shedding rule', async () => {
    // The two phases are different games, and which one is running decides
    // what clicking a card means.
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(
      screen.getByText(
        /第1フェーズ（集める）.*手札を使って札を集めます.*2人の一騎打ち（スート無関係・強い方が両方取る）/,
      ),
    ).toBeInTheDocument();
    expect(screen.queryByText(/直前の札を上回る（同スートの上位か切札）/)).not.toBeInTheDocument();
  });

  it('shows the shedding objective and rule without the collecting rule', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [2] }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(
      screen.getByText(/第2フェーズ（出し切る）.*手札を出し切ります.*直前の札を上回る（同スートの上位か切札）/),
    ).toBeInTheDocument();
    expect(screen.queryByText(/2人の一騎打ち（スート無関係・強い方が両方取る）/)).not.toBeInTheDocument();
  });

  it('shows the trump as undecided until the stock fixes it', async () => {
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByText(/切札: 未定/)).toBeInTheDocument());
  });

  it('shows every opponent count without their cards', async () => {
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/CPU1 の手札 3 枚/)).toBeInTheDocument();
    expect(screen.getByText(/CPU2 の手札 3 枚/)).toBeInTheDocument();
  });

  it('reveals CPU cards only after the game ends and keeps empty seats visible', async () => {
    const revealedCard = card('DIAMOND', 7);
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        loserIdx: 1,
        players: [
          human(),
          cpu(1, { cards: [revealedCard], hidden: false }),
          cpu(2, { cardCount: 0, cards: [], hidden: false }),
        ],
      }),
    );
    const { unmount } = renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
    expect(screen.getByText(/CPU2 の手札 0 枚/)).toBeInTheDocument();
    expect(screen.getByTestId('sg-loser-seat')).toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ players: [human(), cpu(1, { cards: [revealedCard] }), cpu(2)] }));
    unmount();
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByRole('img', { name: 'CPU1 の手札 3 枚（裏向き）' })).toBeInTheDocument());
    expect(screen.queryByAltText('♦ 7')).not.toBeInTheDocument();
  });

  it('only plays the hand cards the server marked valid', async () => {
    // The beat rule lives on the server; the page must not accept a click on
    // a card it did not offer.
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [2] }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();

    fireEvent.click(handButtons[0]);
    // Without the flush this cannot fail: nothing has had a chance to dispatch.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **「出せない」だけでは理由が分からない** (#5573)。規則はサーバが持つので、
  // 画面は「規則に負けている」のか「手番でない」のかを言う。
  it('tells the screen reader why a card cannot be played', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [2] }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    expect(handButtons[0]).toHaveAttribute('title', '場の札を上回れません（同スートの上位か切札が要ります）');
    expect(handButtons[0].getAttribute('aria-label')).toContain('場の札を上回れません');
    // 出せる札には理由を付けない。
    expect(handButtons[2]).not.toHaveAttribute('title');
    expect(handButtons[2].getAttribute('aria-label')).not.toContain('上回れません');
  });

  // **手番でないだけの札を「規則に負けている」と言わない。**
  it('says it is not your turn instead of blaming the beat rule', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [2], currentPlayerIdx: 1 }),
    );
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    expect(handButtons[2]).not.toHaveAttribute('title');
    expect(handButtons[2].getAttribute('aria-label')).toContain('いまは自分の番ではありません');
    // **サーバが出せないと言っている札でも、手番でなければ理由はそちら。**
    // 相手の番に「規則で負けている」と読み上げると、次の自分の番に出せる札まで
    // 出せないものとして覚えてしまう。
    expect(handButtons[0]).not.toHaveAttribute('title');
    expect(handButtons[0].getAttribute('aria-label')).toContain('いまは自分の番ではありません');
    expect(handButtons[0].getAttribute('aria-label')).not.toContain('上回れません');
  });

  it('enables the pick-up only when the server says nothing beats the pile', async () => {
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const pickUp = screen.getByRole('button', { name: '引き取る' });
    expect(pickUp).not.toHaveAttribute('aria-describedby');
    expect(screen.queryByText('上回れる札がありません')).not.toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(pickUp);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('picks the pile up once that is the only move', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [], canPickUp: true }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const pickUp = screen.getByRole('button', { name: '引き取る' });
    expect(pickUp).toHaveAttribute('aria-describedby', 'sg-pickup-hint');
    expect(document.getElementById('sg-pickup-hint')).toHaveTextContent('上回れる札がありません');
    mockExec.mockClear();
    fireEvent.click(pickUp);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pickup'));
  });

  it('reports each outcome', async () => {
    for (const [loser, text] of [
      [0, 'あなたが Skitgubbe です'],
      [2, '手札を出し切りました'],
    ] as const) {
      const code = loser === 0 ? 'skitgubbe.lose' : 'skitgubbe.win';
      mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, loserIdx: loser, messageCode: code }));
      renderWithProviders(<SkitgubbePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  it('shows the losing CPU seat only after the human wins', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, loserIdx: 2 }));
    const { unmount } = renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('sg-loser-seat')).not.toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, loserIdx: 2, messageCode: 'skitgubbe.win' }));
    renderWithProviders(<SkitgubbePage />);
    expect(await screen.findByTestId('sg-loser-seat')).toHaveTextContent('敗者: CPU2（Skitgubbe）');
  });

  it('does not show a losing seat when the human is the loser', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, loserIdx: 0, messageCode: 'skitgubbe.lose' }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getAllByText('あなたが Skitgubbe です').length).toBeGreaterThan(0));
    expect(screen.queryByTestId('sg-loser-seat')).not.toBeInTheDocument();
  });

  it('shows the loser seat in English', async () => {
    try {
      await i18n.changeLanguage('en');
      mockExec.mockResolvedValue(makeState({ phase: 2, gameEndFlag: true, loserIdx: 2, messageCode: 'skitgubbe.win' }));
      renderWithProviders(<SkitgubbePage />);
      expect(await screen.findByTestId('sg-loser-seat')).toHaveTextContent('Loser: CPU2 (Skitgubbe)');
    } finally {
      await i18n.changeLanguage('ja');
    }
  });

  it('announces that picking up has become forced', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [], canPickUp: true }));
    renderWithProviders(<SkitgubbePage />);
    const notice = await screen.findByTestId('sk-forced-pickup-notice');
    expect(notice).toHaveAttribute('role', 'status');
    expect(notice).toHaveAttribute('aria-live', 'polite');
    expect(notice).toHaveTextContent(/./);
  });

  it('stays silent while a playable card remains', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [2] }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('sk-forced-pickup-notice')).not.toBeInTheDocument();
  });
  // 出し切った席は Skitgubbe になる危険から外れる。残る席にとって「誰がもう安全か」は
  // 読みの前提なのに、finished はレスポンスに載ったまま一度も読まれていなかった。
  it('badges the seats that have already shed their hand', async () => {
    mockExec.mockResolvedValue(makeState({ players: [human(), cpu(1, { finished: true }), cpu(2)] }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('sg-finished-1')).toBeInTheDocument());
    expect(screen.getByTestId('sg-finished-1')).toHaveTextContent('セーフ');
    // まだ手札のある席には付かない。
    expect(screen.queryByTestId('sg-finished-2')).not.toBeInTheDocument();
  });

  it('badges nobody while every seat still holds cards', async () => {
    mockExec.mockResolvedValue(makeState({ players: [human(), cpu(1), cpu(2)] }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('sg-finished-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sg-finished-2')).not.toBeInTheDocument();
  });

  it('badges the current duel leader during the collect phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 0, duelLeader: 1 }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('sg-duel-leader-1')).toBeInTheDocument());
    expect(screen.getByTestId('sg-duel-leader-1')).toHaveTextContent('リード');
    expect(screen.queryByTestId('sg-duel-leader-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sg-duel-leader-2')).not.toBeInTheDocument();
  });

  it('badges another duel leader during the collect phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 0, duelLeader: 2 }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('sg-duel-leader-2')).toBeInTheDocument());
    expect(screen.getByTestId('sg-duel-leader-2')).toHaveTextContent('リード');
    expect(screen.queryByTestId('sg-duel-leader-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sg-duel-leader-1')).not.toBeInTheDocument();
  });

  it('badges the human duel leader during the collect phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 0, duelLeader: 0 }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('sg-duel-leader-0')).toBeInTheDocument());
    expect(screen.getByTestId('sg-duel-leader-0')).toHaveTextContent('リード');
    expect(screen.queryByTestId('sg-duel-leader-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sg-duel-leader-2')).not.toBeInTheDocument();
  });

  it('does not badge the duel leader during the shed phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, duelLeader: 1 }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('sg-duel-leader-1')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sg-duel-leader-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sg-duel-leader-2')).not.toBeInTheDocument();
  });
});
