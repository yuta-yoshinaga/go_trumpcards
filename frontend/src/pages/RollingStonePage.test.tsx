import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rollingstoneApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RollingStoneResponse } from '../types/card';
import { RollingStonePage } from './RollingStonePage';

vi.mock('../api/gameApi', () => ({
  rollingstoneApi: { exec: vi.fn() },
  actionLogApi: { rollingstone: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(rollingstoneApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const hand = [card('SPADE', 9), card('SPADE', 13), card('HEART', 10), card('CLOVER', 7)];

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 8,
  cards: id === 0 ? hand : [],
  pickups: 0,
  finishedAt: 0,
  ...over,
});

function makeState(overrides: Partial<RollingStoneResponse> = {}): RollingStoneResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    mustPickUp: false,
    validPlays: [0, 1],
    currentTrick: [],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 2,
    lastPickupIdx: -1,
    finishedCnt: 0,
    deckSize: 32,
    discarded: 8,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4 },
    message: '',
    ...overrides,
  } as unknown as RollingStoneResponse;
}

/** The human cannot follow, so the only move is to take the trick. */
const forcedPickUp = (over: Partial<RollingStoneResponse> = {}) =>
  makeState({
    mustPickUp: true,
    validPlays: [],
    currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 9) }],
    leadSuit: 4,
    ...over,
  } as Partial<RollingStoneResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('RollingStonePage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **勝利条件が逆さまなのが規則そのもの。**
  it('states that taking tricks is worth nothing', async () => {
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-rule')).toHaveTextContent(/得点にはならず/);
  });

  // **デッキ枚数は人数で変わる。**
  it('shows the deck size and how much is still in play', async () => {
    renderWithProviders(<RollingStonePage />);
    const deck = await screen.findByTestId('rs-deck');
    expect(deck).toHaveTextContent('32');
    expect(deck).toHaveTextContent('24');
  });

  // **手札の枚数がそのまま順位。** 得点表示は無い。
  it('shows every hand size and pickup count', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cardCount: 11, pickups: 2 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<RollingStonePage />);
    const s0 = await screen.findByTestId('rs-seat-0');
    expect(s0).toHaveTextContent('手札11枚');
    expect(s0).toHaveTextContent('引き取り2回');
    expect(screen.getByTestId('rs-seat-3')).toBeInTheDocument();
  });

  // **引き取った席と上がった席は盤面に痕跡が残らない。**
  it('marks the last pickup and finishers', async () => {
    const { unmount } = renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('rs-seat-2')).not.toHaveTextContent(/引き取り\]/);
    unmount();

    mockExec.mockResolvedValue(makeState({ lastPickupIdx: 2 }));
    const second = renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-seat-2')).toHaveTextContent(/直前に引き取り/);
    second.unmount();

    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, { cardCount: 0, finishedAt: 1 }), seat(1), seat(2), seat(3)] }),
    );
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-seat-0')).toHaveTextContent(/1位で上がり/);
  });

  it('plays the clicked card by its hand index', async () => {
    renderWithProviders(<RollingStonePage />);
    const cards = await screen.findAllByRole('button', { name: /を出す$/ });
    mockExec.mockClear();
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  // **出せる札が無い局面は、手札を押させずに引き取らせる。**
  it('offers only the pickup when you cannot follow', async () => {
    mockExec.mockResolvedValue(forcedPickUp());
    renderWithProviders(<RollingStonePage />);

    // **枚数だけでは、なぜ出せないのかが分からない** (#5764)。追従できなかった
    // スートまで書く。
    const banner = await screen.findByTestId('rs-must-pickup');
    expect(banner).toHaveTextContent('1');
    expect(banner).toHaveTextContent('♦');
    const pickup = screen.getByTestId('rs-pickup-btn');
    expect(pickup).toBeEnabled();
    const cards = screen.getAllByRole('button', { name: /を出す$/ });
    expect(cards[0]).toBeDisabled();

    mockExec.mockClear();
    fireEvent.click(pickup);
    // **引き取りは別のコマンド。** cardIndex は送らない。
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pickup'));
  });

  // 場が空のまま引き取りが立つことは規則上ありえないが、そのときでも
  // バナー自体は壊れない。
  it('falls back to a placeholder when no suit has been led', async () => {
    mockExec.mockResolvedValue(forcedPickUp({ leadSuit: 0, currentTrick: [] }));
    renderWithProviders(<RollingStonePage />);

    const banner = await screen.findByTestId('rs-must-pickup');
    expect(banner).toHaveTextContent('?');
    expect(banner).not.toHaveTextContent('♦');
  });

  // **負のコントロール: フォローできるなら引き取りは出さない。**
  it('hides the pickup button while you can still follow', async () => {
    renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('rs-pickup-btn')).not.toBeInTheDocument();
    expect(screen.queryByTestId('rs-must-pickup')).not.toBeInTheDocument();
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<RollingStonePage />);
    const cards = await screen.findAllByRole('button', { name: /を出す$/ });
    expect(cards[0]).toBeDisabled();
    expect(screen.queryByTestId('rs-pickup-btn')).not.toBeInTheDocument();
  });

  // **上限で切った局は「上がった」わけではない。** 言い分ける。
  it('distinguishes running out from the stalemate', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 0,
        players: [seat(0, { cardCount: 0, finishedAt: 1 }), seat(1), seat(2), seat(3)],
      }),
    );
    const { unmount } = renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-result')).toHaveTextContent(/先に手札を出し切りました/);
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 2,
        players: [seat(0), seat(1), seat(2, { cardCount: 3 }), seat(3)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    const banner = await screen.findByTestId('rs-result');
    expect(banner).toHaveTextContent(/決着が付かなかった/);
    expect(banner).toHaveTextContent('3');
  });

  it('reports a CPU running out first', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        phase: 1,
        winnerIdx: 1,
        players: [seat(0), seat(1, { cardCount: 0, finishedAt: 1 }), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('rs-result')).toHaveTextContent(/CPU1 が先に出し切りました/);
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<RollingStonePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'pickup', reason: 'hint.rollingstonePickUp', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<RollingStonePage />);
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/引き取るしかありません/);
  });
});
