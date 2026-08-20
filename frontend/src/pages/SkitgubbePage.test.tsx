import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { skitgubbeApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SkitgubbePlayer, SkitgubbeResponse } from '../types/card';
import { SkitgubbePage } from './SkitgubbePage';

vi.mock('../api/gameApi', () => ({
  skitgubbeApi: { exec: vi.fn() },
  actionLogApi: { skitgubbe: vi.fn() },
}));

const mockExec = vi.mocked(skitgubbeApi.exec);

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

  it('names the running phase and shows both phases rules permanently', async () => {
    // The two phases are different games, and which one is running decides
    // what clicking a card means.
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/第1フェーズ（集める）/)).toBeInTheDocument();
    expect(screen.getByText(/2人の一騎打ち/)).toBeInTheDocument();
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
    mockExec.mockClear();
    fireEvent.click(pickUp);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('picks the pile up once that is the only move', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pile: [card('SPADE', 10)], validIndices: [], canPickUp: true }));
    renderWithProviders(<SkitgubbePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '引き取る' }));
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
});
