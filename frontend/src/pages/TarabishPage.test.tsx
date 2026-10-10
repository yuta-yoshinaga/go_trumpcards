import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tarabishApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TarabishResponse } from '../types/card';
import { TarabishPage } from './TarabishPage';

vi.mock('../api/gameApi', () => ({
  tarabishApi: { exec: vi.fn() },
  actionLogApi: { tarabish: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(tarabishApi.exec);
const mobileState = vi.hoisted(() => ({ isMobile: true }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../hooks/useCardDimensions')>()),
  useIsMobile: () => mobileState.isMobile,
}));

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 3,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9), card('CLOVER', 1)] : [],
  meldPoints: 0,
  runLen: 0,
  hasBella: false,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<TarabishResponse> = {}): TarabishResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    trumpSuit: 3,
    upCard: card('HEART', 9),
    trumpTakerIdx: -1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    // 既定では親を 2 番にして、人間が見送れる状況にする。
    dealerIdx: 2,
    scores: [0, 0],
    roundPoints: [0, 0],
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 500 },
    message: '',
    ...overrides,
  } as unknown as TarabishResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<TarabishResponse> = {}) =>
  makeState({ phase: 1, trumpTakerIdx: 0, ...over } as Partial<TarabishResponse>);

beforeEach(() => {
  vi.clearAllMocks();
  mobileState.isMobile = true;
  mockExec.mockResolvedValue(makeState());
});

describe('TarabishPage', () => {
  it('keeps the original points label off turn', async () => {
    renderWithProviders(<TarabishPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toHaveAccessibleName(/♥ J（20点）を出す/);
    expect(cards[0]).not.toHaveAccessibleName(/出せる|出せない/);
    expect(cards[0]).not.toHaveAttribute('aria-disabled');
  });

  it('announces playable and unplayable cards from validPlays', async () => {
    mockExec.mockResolvedValue(playing({ validPlays: [1] }));
    renderWithProviders(<TarabishPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards).toHaveLength(3);
    expect(cards[0]).toHaveAccessibleName(/出せない/);
    expect(cards[1]).toHaveAccessibleName(/出せる/);
    expect(cards[2]).toHaveAccessibleName(/出せない/);
  });

  it('keeps illegal cards focusable but aria-disabled and does not play them', async () => {
    mockExec.mockResolvedValue(playing({ validPlays: [1] }));
    renderWithProviders(<TarabishPage />);
    const cards = await screen.findAllByRole('button', { name: /出す/ });
    expect(cards[1]).not.toHaveAttribute('aria-disabled');
    expect(cards[0]).toHaveAttribute('aria-disabled', 'true');
    expect(cards[0]).not.toBeDisabled();
    expect(document.getElementById(cards[0].getAttribute('aria-describedby') ?? '')).toHaveTextContent(
      'この札は現在出せません',
    );

    mockExec.mockClear();
    fireEvent.click(cards[0]);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
  });
  it('resets on mount', async () => {
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('keeps the trump order collapsed on mobile and exposes it on demand', async () => {
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-order-details')).not.toHaveAttribute('open');
    expect(screen.getByTestId('tb-order')).not.toBeVisible();
    fireEvent.click(screen.getByText('ルール'));
    expect(screen.getByTestId('tb-order')).toHaveTextContent(/J\(Jass\)=20/);
    expect(screen.getByTestId('tb-order')).toHaveTextContent(/9\(Menel\)=14/);
  });

  it('keeps the rules disclosure available on desktop', async () => {
    mobileState.isMobile = false;
    renderWithProviders(<TarabishPage />);
    const details = await screen.findByTestId('tb-order-details');
    expect(details).not.toHaveAttribute('open');
    expect(details.querySelector('summary')).toHaveTextContent('ルール');
    mobileState.isMobile = true;
  });

  it('collapses CPU seats on mobile and opens them on desktop', async () => {
    const firstRender = renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('cpu-accordion')).not.toHaveAttribute('open');
    expect(screen.getByTestId('tb-seat-0')).not.toBeVisible();
    fireEvent.click(screen.getByTestId('cpu-accordion').querySelector('summary') as HTMLElement);
    expect(screen.getByTestId('tb-seat-0')).toBeVisible();
    firstRender.unmount();
    mobileState.isMobile = false;
    const { unmount } = renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('cpu-accordion')).toHaveAttribute('open');
    expect(screen.getByTestId('tb-seat-0')).toBeVisible();
    unmount();
    mobileState.isMobile = true;
  });

  it('offers both choices while bidding', async () => {
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-take-btn')).toBeInTheDocument();
    expect(screen.getByTestId('tb-pass-btn')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: /出す/ })[0]).not.toHaveAttribute('aria-disabled');
  });

  // **親は見送れないので、見送りボタンを出さない。** 負のコントロール付き。
  it('hides the pass button when the human is the dealer', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<TarabishPage />);

    expect(await screen.findByTestId('tb-take-btn')).toBeInTheDocument();
    expect(screen.queryByTestId('tb-pass-btn')).not.toBeInTheDocument();
    expect(screen.getByTestId('tb-dealer-stuck')).toBeInTheDocument();
  });

  // 引き受けと見送りは別のコマンドを送る。
  it.each([
    ['tb-take-btn', 'take'],
    ['tb-pass-btn', 'pass'],
  ])('sends %s as the %s command', async (testId, command) => {
    renderWithProviders(<TarabishPage />);
    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('hides the bid buttons once trump is settled', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('tb-take-btn')).not.toBeInTheDocument();
  });

  // 入札前は候補、決まったあとは切り札。両側を踏む。
  it('shows the turned card before trump is settled', async () => {
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-upcard')).toBeInTheDocument();
    expect(screen.queryByTestId('tb-trump')).not.toBeInTheDocument();
  });

  it('shows who took trump once it is settled', async () => {
    mockExec.mockResolvedValue(playing({ trumpTakerIdx: 2 } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-trump')).toHaveTextContent('CPU2');
    expect(screen.queryByTestId('tb-upcard')).not.toBeInTheDocument();
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **チーム番号とメルドは盤面から読めない。** 席ごとに出す。
  it('labels each seat with its team and meld', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { meldPoints: 70, runLen: 4, hasBella: true }), seat(1), seat(2), seat(3)],
      } as Partial<TarabishResponse>),
    );
    renderWithProviders(<TarabishPage />);
    fireEvent.click((await screen.findByTestId('cpu-accordion')).querySelector('summary') as HTMLElement);

    expect(await screen.findByTestId('tb-seat-0')).toHaveTextContent('T0');
    expect(screen.getByTestId('tb-seat-0')).toHaveTextContent('ラン4枚');
    expect(screen.getByTestId('tb-seat-0')).toHaveTextContent('ベラ');
    expect(screen.getByTestId('tb-seat-1')).toHaveTextContent('T1');
    expect(screen.getByTestId('tb-seat-1')).toHaveTextContent('メルドなし');
    // 0 と 2 が味方であることが席表示から読める。
    expect(screen.getByTestId('tb-seat-2')).toHaveTextContent('T0');
  });

  // **トリック配分がそのまま点数の趨勢。**CUI の playerLine は毎回出している。
  it('shows how many tricks each seat has taken', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, { trickCount: 3 }), seat(1, { trickCount: 1 }), seat(2), seat(3, { trickCount: 4 })],
      } as Partial<TarabishResponse>),
    );
    renderWithProviders(<TarabishPage />);

    // 席ごとに別の値が、その席の要素に出る（全席同じ値を出す実装では落ちる）。
    expect(await screen.findByTestId('tb-seat-tricks-0')).toHaveTextContent('獲得3');
    expect(screen.getByTestId('tb-seat-tricks-1')).toHaveTextContent('獲得1');
    expect(screen.getByTestId('tb-seat-tricks-2')).toHaveTextContent('獲得0');
    expect(screen.getByTestId('tb-seat-tricks-3')).toHaveTextContent('獲得4');
  });

  it('shows the remaining hand count for all four seats', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          seat(0, { cardCount: 5 }),
          seat(1, { cardCount: 4 }),
          seat(2, { cardCount: 2 }),
          seat(3, { cardCount: 1 }),
        ],
      } as Partial<TarabishResponse>),
    );
    renderWithProviders(<TarabishPage />);

    expect(await screen.findByTestId('tb-seat-cards-0')).toHaveTextContent('残り手札5枚');
    expect(screen.getByTestId('tb-seat-cards-1')).toHaveTextContent('残り手札4枚');
    expect(screen.getByTestId('tb-seat-cards-2')).toHaveTextContent('残り手札2枚');
    expect(screen.getByTestId('tb-seat-cards-3')).toHaveTextContent('残り手札1枚');
  });

  it('shows the running team scores', async () => {
    mockExec.mockResolvedValue(makeState({ scores: [220, 140] } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-score')).toHaveTextContent('220');
    expect(screen.getByTestId('tb-score')).toHaveTextContent('140');
  });

  it("shows this round's team points separately from running scores", async () => {
    mockExec.mockResolvedValue(makeState({ scores: [220, 140], roundPoints: [35, 12] } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByTestId('tb-score')).toHaveTextContent('220');
    expect(screen.getByTestId('tb-round-points')).toHaveTextContent('ラウンド得点: あなたのチーム 35 － 相手 12');
    expect(screen.getByTestId('tb-round-points')).toHaveTextContent('ラウンド得点');
  });

  it('advances the round when the next-round button is pressed', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<TarabishPage />);

    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '投了' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('renders the result banner for each outcome', async () => {
    for (const [winnerTeam, expected] of [
      [0, /あなたのチームの勝ち/],
      [1, /相手チームの勝ち/],
      [-1, /同点/],
    ] as const) {
      mockExec.mockResolvedValue(makeState({ gameEndFlag: true, phase: 3, winnerTeam }));
      const { unmount } = renderWithProviders(<TarabishPage />);
      expect(await screen.findByText(expected)).toBeInTheDocument();
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<TarabishResponse>));
    renderWithProviders(<TarabishPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'take-trump', reason: 'hint.tarabishTakeTrump', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByText(/引き受けてよいでしょう/)).toBeInTheDocument();
  });

  it('resolves the game-end message with last-trick bonus to Japanese', async () => {
    mockExec.mockResolvedValue(
      makeState({
        gameEndFlag: true,
        winnerTeam: 0,
        messageCode: 'tarabish.result.team0.bonusYou',
        messageParams: { t0: '520', t1: '300', bonus: '10' },
      }),
    );
    renderWithProviders(<TarabishPage />);

    expect(
      await screen.findByText(
        'あなたのチームの勝ちです（520 － 300）。あなたのチームが最終トリックボーナス +10 点を獲得。',
      ),
    ).toBeInTheDocument();
    expect(screen.queryByText(/tarabish\.result/)).not.toBeInTheDocument();
  });
});

// **切り札だけ点数表が入れ替わるのがこの系統の肝** (#5749)。同じ J でも
// 切り札なら 20 点 (Jass)、そうでなければ 2 点。暗算させるとパートナーに
// 寄せる札を間違える。
describe('TarabishPage card points', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('scores the trump jack and nine by the trump table', async () => {
    // 切り札 = ♥。手札は ♥J(Jass) / ♥9(Menel) / ♠J / ♠9。
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 3,
        players: [
          seat(0, {
            cards: [card('HEART', 11), card('HEART', 9), card('SPADE', 11), card('SPADE', 9)],
            cardCount: 4,
          }),
          seat(1),
          seat(2),
          seat(3),
        ],
      }),
    );
    renderWithProviders(<TarabishPage />);

    expect(await screen.findByTestId('tb-points-0')).toHaveTextContent('20');
    expect(screen.getByTestId('tb-points-1')).toHaveTextContent('14');
    // 同じランクでも切り札でなければ別の表。
    expect(screen.getByTestId('tb-points-2')).toHaveTextContent('2');
    expect(screen.getByTestId('tb-points-3')).toHaveTextContent('0');
  });

  it('says the points in the accessible name too', async () => {
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 3,
        players: [seat(0, { cards: [card('HEART', 11)], cardCount: 1 }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<TarabishPage />);
    expect(await screen.findByRole('button', { name: '♥ J（20点）を出す' })).toBeInTheDocument();
  });

  // **切り札が決まるまで点は定まらない。**入札中に出すと嘘になる。
  it('shows no points until a trump suit is called', async () => {
    mockExec.mockResolvedValue(
      makeState({
        trumpSuit: 0,
        players: [seat(0, { cards: [card('HEART', 11)], cardCount: 1 }), seat(1), seat(2), seat(3)],
      }),
    );
    renderWithProviders(<TarabishPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('tb-points-0')).not.toBeInTheDocument();
  });

  it('shows the human hand suit counts only while bidding', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 0 }));
    renderWithProviders(<TarabishPage />);
    const counts = await screen.findByTestId('tb-bid-suit-counts');
    expect(counts).toHaveTextContent('スペード 1');
    expect(counts).toHaveTextContent('クラブ 1');
    expect(counts).toHaveTextContent('ハート 1');
    expect(counts).toHaveTextContent('ダイヤ 0');
    expect(counts).toHaveTextContent('枚数');
  });

  it('does not show bid suit counts after trump is settled', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<TarabishPage />);
    await screen.findByTestId('tb-trump');
    expect(screen.queryByTestId('tb-bid-suit-counts')).not.toBeInTheDocument();
  });
});
