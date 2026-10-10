import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { hasenpfefferApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, HasenpfefferResponse } from '../types/card';
import { HasenpfefferPage } from './HasenpfefferPage';

vi.mock('../api/gameApi', () => ({
  hasenpfefferApi: { exec: vi.fn() },
  actionLogApi: { hasenpfeffer: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(hasenpfefferApi.exec);
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
  bid: -1,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<HasenpfefferResponse> = {}): HasenpfefferResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    handNumber: 1,
    trickNumber: 0,
    trumpSuit: 0,
    declarerIdx: -1,
    contract: 0,
    minBid: 3,
    mustBid: false,
    blindSize: 1,
    scores: [0, 0],
    teamTricks: [0, 0],
    lastHandEuchred: false,
    lastHandTricks: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    validPlays: [0, 1, 2],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 10 },
    message: '',
    ...overrides,
  } as unknown as HasenpfefferResponse;
}

/** A state where trump is settled and it is the human's turn to play. */
const playing = (over: Partial<HasenpfefferResponse> = {}) =>
  makeState({
    phase: 2,
    trumpSuit: 3,
    declarerIdx: 1,
    contract: 4,
    blindSize: 0,
    ...over,
  } as Partial<HasenpfefferResponse>);

beforeEach(() => {
  mobileState.isMobile = true;
  vi.clearAllMocks();
  mockExec.mockResolvedValue(makeState());
});

describe('HasenpfefferPage', () => {
  it('includes the card and player name in the trick card accessible name', async () => {
    mockExec.mockResolvedValue(
      playing({ currentTrick: [{ playerIdx: 1, card: card('HEART', 11) }] } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);

    const trickCards = await screen.findByTestId('trick-display-cards');
    expect(trickCards.querySelector('img')).toHaveAttribute('alt', expect.stringMatching(/CPU 1が出した♥ J/));
  });

  it('announces playable cards during play and keeps discard labels unchanged', async () => {
    mockExec.mockResolvedValue(playing({ validPlays: [1] }));
    renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findAllByRole('button', { name: /プレイ可能/ })).toHaveLength(1);
    expect(screen.getAllByRole('button', { name: /を出す/ })).toHaveLength(3);
  });

  it('blocks invalid cards during play while keeping valid cards selectable', async () => {
    mockExec.mockResolvedValue(playing({ validPlays: [1] }));
    renderWithProviders(<HasenpfefferPage />);

    const playable = await screen.findByRole('button', { name: /プレイ可能/ });
    const invalid = screen.getAllByRole('button', { name: /を出す/ });
    expect(playable).not.toHaveAttribute('aria-disabled', 'true');
    expect(playable).toHaveClass('ring-ds-success');
    expect(invalid[0]).toHaveAttribute('aria-disabled', 'true');

    fireEvent.click(invalid[0]);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());

    fireEvent.click(playable);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', expect.anything()));
  });

  it('resets on mount', async () => {
    renderWithProviders(<HasenpfefferPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('announces when the human must bid', async () => {
    mockExec.mockResolvedValue(makeState({ mustBid: true }));
    renderWithProviders(<HasenpfefferPage />);

    const visible = await screen.findByTestId('hpf-must-bid');
    const live = await waitFor(() => {
      const element = [...document.querySelectorAll('[role="status"][aria-live="polite"][aria-atomic="true"]')].find(
        (candidate) => candidate.textContent === visible.textContent,
      );
      expect(element).not.toBeUndefined();
      return element;
    });
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveAttribute('aria-atomic', 'true');
  });

  it('marks only the dealer seat', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 2 }));
    renderWithProviders(<HasenpfefferPage />);

    const dealerSeat = await screen.findByTestId('hpf-seat-2');
    expect(dealerSeat).toHaveClass('border-0', 'm-0', 'min-w-0');
    expect(dealerSeat).toHaveTextContent('/ 親');
    expect(screen.getByTestId('hpf-seat-0')).not.toHaveTextContent('/ 親');
  });

  it('moves the dealer mark when dealerIdx changes', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 1 }));
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-seat-1')).toHaveTextContent('/ 親');
    expect(screen.getByTestId('hpf-seat-3')).not.toHaveTextContent('/ 親');
  });

  // **ジョーカーが最強という序列は知らないと打ち方が変わる。**
  it('keeps the rule details closed until requested', async () => {
    renderWithProviders(<HasenpfefferPage />);
    const details = await screen.findByTestId('hpf-rule');
    expect(details).not.toHaveAttribute('open');
    expect(details.querySelector('div')).not.toBeVisible();
    fireEvent.click(details.querySelector('summary') as HTMLElement);
    expect(details.querySelector('div')).toBeVisible();
    expect(details).toHaveTextContent(/Best Bower/);
  });

  it('closes the seat accordion on mobile and opens it on desktop', async () => {
    const { unmount } = renderWithProviders(<HasenpfefferPage />);
    const mobileSeats = await screen.findByTestId('cpu-accordion');
    expect(mobileSeats).not.toHaveAttribute('open');
    expect(screen.getByTestId('hpf-seat-1')).not.toBeVisible();

    unmount();
    mobileState.isMobile = false;
    renderWithProviders(<HasenpfefferPage />);
    const desktopSeats = await screen.findByTestId('cpu-accordion');
    expect(desktopSeats).toHaveAttribute('open');
    expect(screen.getByTestId('hpf-seat-1')).toBeVisible();
  });

  // **サーバが必ず拒否する額は出さない (#5304)。**
  it('offers only bids at or above the minimum', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 5 }));
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-bid-5-btn')).toBeInTheDocument();
    expect(screen.getByTestId('hpf-bid-6-btn')).toBeInTheDocument();
    for (const n of [3, 4]) {
      expect(screen.queryByTestId(`hpf-bid-${n.toString()}-btn`)).not.toBeInTheDocument();
    }
  });

  // **上限が立っていたら宣言できない。** 降りるボタンだけが残る。
  it('offers no bid at all once the maximum is standing', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 0 }));
    renderWithProviders(<HasenpfefferPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    for (const n of [3, 4, 5, 6]) {
      expect(screen.queryByTestId(`hpf-bid-${n.toString()}-btn`)).not.toBeInTheDocument();
    }
    expect(screen.getByTestId('hpf-pass-btn')).toBeInTheDocument();
  });

  // **親が降りられない場面では降りるボタンを出さない。**
  it('hides the pass button when the dealer cannot pass', async () => {
    mockExec.mockResolvedValue(makeState({ mustBid: true }));
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-must-bid')).toBeInTheDocument();
    expect(screen.queryByTestId('hpf-pass-btn')).not.toBeInTheDocument();
    // 負のコントロール: 宣言ボタンは出ている
    expect(screen.getByTestId('hpf-bid-3-btn')).toBeInTheDocument();
  });

  // **宣言は5番目の引数で送る。** 位置がずれると別の値として届く。
  it.each([3, 4, 5, 6])('sends bid %s', async (n) => {
    renderWithProviders(<HasenpfefferPage />);
    const btn = await screen.findByTestId(`hpf-bid-${n.toString()}-btn`);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, n));
  });

  it('sends a pass as bid 0', async () => {
    renderWithProviders(<HasenpfefferPage />);
    const btn = await screen.findByTestId('hpf-pass-btn');
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, undefined, 0));
  });

  // **捨て札は札を選んでからスートを選ぶ。** 選ぶ前は確定できない。
  it('requires a card before the trump buttons work', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 1, declarerIdx: 0, contract: 4, blindSize: 0 } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);

    const suitBtn = await screen.findByTestId('hpf-discard-3-btn');
    expect(suitBtn).toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(suitBtn);
    await waitFor(() => expect(mockExec).not.toHaveBeenCalled());

    const cards = screen.getAllByRole('button', { name: /捨て札に選ぶ/ });
    expect(screen.queryByRole('button', { name: /プレイ可能/ })).not.toBeInTheDocument();
    fireEvent.click(cards[1]);
    expect(cards[1]).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(screen.getByTestId('hpf-discard-3-btn'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 1, undefined, 3));
  });

  it('plays the clicked card by its hand index', async () => {
    mockExec.mockResolvedValue(playing());
    renderWithProviders(<HasenpfefferPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    mockExec.mockClear();
    fireEvent.click(cards[2]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  // **宣言の状態は 3 通り。**
  it('renders every bid state on the seats', async () => {
    mockExec.mockResolvedValue(
      playing({
        players: [seat(0, { bid: 4 }), seat(1, { bid: 0 }), seat(2), seat(3)],
      } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-seat-0')).toHaveTextContent('4');
    expect(screen.getByTestId('hpf-seat-1')).toHaveTextContent(/降り/);
    expect(screen.getByTestId('hpf-seat-2')).toHaveTextContent(/未宣言/);
    expect(screen.getByTestId('hpf-seat-1')).toHaveTextContent('[落札]');
  });

  it('groups each player name with their bid state and tricks taken', async () => {
    mockExec.mockResolvedValue(
      playing({
        players: [seat(0, { bid: 4, trickCount: 2 }), seat(1, { bid: 0, trickCount: 1 })],
      } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);
    fireEvent.click((await screen.findByTestId('cpu-accordion')).querySelector('summary') as HTMLElement);

    const humanSeat = await screen.findByTestId('hpf-seat-0');
    const cpuSeat = screen.getByTestId('hpf-seat-1');
    expect(humanSeat).toHaveAccessibleName('あなた');
    expect(humanSeat).toHaveTextContent('宣言4');
    expect(humanSeat).toHaveTextContent('獲得2');
    expect(cpuSeat).toHaveAccessibleName('CPU1');
    expect(cpuSeat).toHaveTextContent('降り');
    expect(cpuSeat).toHaveTextContent('獲得1');
  });

  it('shows tricks taken by each team during the hand', async () => {
    mockExec.mockResolvedValue(playing({ teamTricks: [3, 1] }));
    renderWithProviders(<HasenpfefferPage />);

    expect(await screen.findByTestId('hpf-team-tricks')).toHaveTextContent('T0 獲得3トリック');
    expect(screen.getByTestId('hpf-team-tricks')).toHaveTextContent('T1 獲得1トリック');
  });

  // 伏せ札・未宣言・確定の 3 状態を踏む。
  it('shows the blind, then trump', async () => {
    const { unmount } = renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-trump')).toHaveTextContent(/伏せ札/);
    unmount();

    mockExec.mockResolvedValue(playing({ trumpSuit: 4 } as Partial<HasenpfefferResponse>));
    renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-trump')).toHaveTextContent('♦');
  });

  // **落としたのか達成したのかは盤面から読めない。** 両側を踏む。
  it.each([
    [false, /達成/],
    [true, /落とし/],
  ])('explains how the hand ended (euchred=%s)', async (lastHandEuchred, expected) => {
    mockExec.mockResolvedValue(
      playing({ phase: 3, lastHandEuchred, lastHandTricks: 3 } as Partial<HasenpfefferResponse>),
    );
    const { unmount } = renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-hand-result')).toHaveTextContent(expected);
    unmount();
  });

  // **達成した側に入るのは契約数ではなく取ったトリック数。**2 つの数字を
  // 並べるだけでは、どちらが加点されるのかが分からない。
  it('names the points actually scored, which are the tricks taken and not the contract', async () => {
    mockExec.mockResolvedValue(
      playing({ phase: 3, lastHandEuchred: false, contract: 4, lastHandTricks: 6 } as Partial<HasenpfefferResponse>),
    );
    renderWithProviders(<HasenpfefferPage />);

    // 契約 4 に対して 6 トリック。入るのは 6 点であって 4 点ではない。
    expect(await screen.findByTestId('hpf-hand-result')).toHaveTextContent('達成しました（+6 点）');
  });

  it('states the asymmetric scoring rule in the standing box', async () => {
    renderWithProviders(<HasenpfefferPage />);
    const rule = await screen.findByTestId('hpf-rule');
    expect(rule).toHaveTextContent('契約数ではなく実際に取ったトリック数');
    expect(rule).toHaveTextContent('落とすと相手に契約数ぶん');
    // ジョーカーの序列の説明は落ちていない。
    expect(rule).toHaveTextContent('Best Bower');
  });

  it('advances to the next hand when the button is pressed', async () => {
    mockExec.mockResolvedValue(playing({ phase: 3 } as Partial<HasenpfefferResponse>));
    renderWithProviders(<HasenpfefferPage />);

    const btn = await screen.findByRole('button', { name: '次のハンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('gives up when the give-up button is pressed', async () => {
    renderWithProviders(<HasenpfefferPage />);
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
      mockExec.mockResolvedValue(playing({ gameEndFlag: true, phase: 4, winnerTeam }));
      const { unmount } = renderWithProviders(<HasenpfefferPage />);
      expect(await screen.findByTestId('hpf-result')).toHaveTextContent(expected);
      unmount();
    }
  });

  it('disables the hand while it is a CPU turn', async () => {
    mockExec.mockResolvedValue(playing({ currentPlayerIdx: 1 } as Partial<HasenpfefferResponse>));
    renderWithProviders(<HasenpfefferPage />);
    const cards = await screen.findAllByRole('button', { name: /を出す/ });
    expect(cards[0]).toBeDisabled();
  });

  it('shows the hint when one is enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'bid-3', reason: 'hint.hasenpfefferMustBid', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<HasenpfefferPage />);
    // **ルール帯にも同じ言い回しが出る。** ツールチップに絞って見る。
    expect(await screen.findByTestId('hint-tooltip')).toHaveTextContent(/最低額で受けましょう/);
  });
});

// **上限に達すると宣言ボタンが 1 つも出ない** (#5758)。理由が書かれていないと、
// ボタンが急に消えたようにしか見えない。
describe('HasenpfefferPage capped bidding', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('explains why no bid button is offered', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 0, mustBid: false }));
    renderWithProviders(<HasenpfefferPage />);

    const banner = await screen.findByTestId('hpf-bid-capped');
    expect(banner).toHaveTextContent('上限');
    expect(banner).toHaveTextContent('6');
    // 強制ビッドのバナーとは別物 (受け入れ条件2)。
    expect(screen.queryByTestId('hpf-must-bid')).not.toBeInTheDocument();
    // 宣言ボタンは 1 つも無い。
    expect(screen.queryByTestId('hpf-bid-6-btn')).not.toBeInTheDocument();
  });

  it('shows the forced-bid banner instead when the dealer cannot pass', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 2, mustBid: true }));
    renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-must-bid')).toBeInTheDocument();
    expect(screen.queryByTestId('hpf-bid-capped')).not.toBeInTheDocument();
  });

  // **強制ビッドが優先される。**CUI の switch も MustBid を先に見る。矛盾する
  // 2 枚のバナーを同時に出さないことを、両方立った応答で押さえる。
  it('never shows both banners at once', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 0, mustBid: true }));
    renderWithProviders(<HasenpfefferPage />);
    expect(await screen.findByTestId('hpf-must-bid')).toBeInTheDocument();
    expect(screen.queryByTestId('hpf-bid-capped')).not.toBeInTheDocument();
  });

  it('stays quiet while bids can still be made', async () => {
    mockExec.mockResolvedValue(makeState({ minBid: 3, mustBid: false }));
    renderWithProviders(<HasenpfefferPage />);
    await waitFor(() => expect(screen.getByTestId('hpf-bid-3-btn')).toBeInTheDocument());
    expect(screen.queryByTestId('hpf-bid-capped')).not.toBeInTheDocument();
  });
});
