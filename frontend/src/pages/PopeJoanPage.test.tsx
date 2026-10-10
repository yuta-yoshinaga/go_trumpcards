import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { popejoanApi } from '../api/gameApi';
import { ApiError } from '../api/gameExec';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, PopeJoanPlayer, PopeJoanResponse } from '../types/card';
import { PopeJoanPhase } from '../types/phases';
import { PopeJoanPage } from './PopeJoanPage';

const mobileState = vi.hoisted(() => ({ isMobile: true }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../hooks/useCardDimensions')>()),
  useIsMobile: () => mobileState.isMobile,
}));

vi.mock('../api/gameApi', () => ({
  popejoanApi: { exec: vi.fn() },
  actionLogApi: { popejoan: vi.fn() },
}));

const mockExec = vi.mocked(popejoanApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

const COMPS = ['ace', 'king', 'queen', 'jack', 'game', 'pope', 'matrimony', 'intrigue'];
const DRESS = [1, 1, 1, 1, 1, 6, 2, 2];

function seat(id: number, isHuman: boolean, overrides?: Partial<PopeJoanPlayer>): PopeJoanPlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 3), card('HEART', 9), card('CLOVER', 13)] : [],
    chips: -15,
    holdsPope: false,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PopeJoanResponse>): PopeJoanResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: PopeJoanPhase.PLAY,
    validPlays: [],
    currentPlayerIdx: 0,
    dealerIdx: 0,
    compartments: COMPS.map((name, i) => ({ name, chips: DRESS[i] })),
    trumpSuit: 1,
    turnUp: card('SPADE', 5),
    awards: [],
    playedPile: [],
    runSuit: -1,
    runRank: 0,
    dealNo: 0,
    targetDeals: 5,
    dealWinner: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

/** Click the hand card at the given index. */
function pickHand(i: number) {
  const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
  fireEvent.click(hand[i]);
}

describe('PopeJoanPage', () => {
  beforeEach(() => {
    mobileState.isMobile = true;
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('keeps rules and CPU details collapsed while their content stays in the DOM', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const ruleDetails = screen.getByTestId('pj-rule-details');
    expect(ruleDetails).not.toHaveAttribute('open');
    const ruleText = ruleDetails.querySelector('[data-tutorial="pj-rule"]');
    expect(ruleText).toBeInTheDocument();
    expect(ruleText?.parentElement?.tagName).not.toBe('SUMMARY');
    expect(ruleText).not.toBeVisible();
    fireEvent.click(ruleDetails.querySelector('summary')!);
    expect(ruleDetails).toHaveAttribute('open');
    expect(ruleText).toBeVisible();
    const cpuDetails = screen.getByTestId('cpu-accordion');
    expect(cpuDetails).not.toHaveAttribute('open');
    expect(cpuDetails.querySelector('[role="img"]')).toBeInTheDocument();
    const seatRow = cpuDetails.querySelector('.flex.flex-wrap');
    expect(seatRow).toBeInTheDocument();
    expect(seatRow?.classList.contains('sm:flex-nowrap')).toBe(true);
  });

  it('keeps CPU seats open on desktop', async () => {
    mobileState.isMobile = false;
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByTestId('cpu-accordion')).toHaveAttribute('open');
  });

  it('resets on mount', async () => {
    renderWithProviders(<PopeJoanPage />);
    expect(screen.getByTestId('popejoan-live')).toHaveAttribute('aria-live', 'polite');
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows an initial request error and retries into the game', async () => {
    mockExec.mockRejectedValueOnce(new Error('network error')).mockResolvedValueOnce(makeState());
    renderWithProviders(<PopeJoanPage />);

    expect(await screen.findByRole('alert')).toHaveTextContent('通信エラーが発生しました。もう一度お試しください。');
    fireEvent.click(screen.getByRole('button', { name: '再試行' }));
    expect(await screen.findAllByTestId('popejoan-compartment')).toHaveLength(8);
    expect(mockExec).toHaveBeenLastCalledWith('reset');
  });

  it('does not offer retry for a non-retryable initial request error', async () => {
    mockExec.mockRejectedValueOnce(new ApiError(400));
    renderWithProviders(<PopeJoanPage />);

    expect(await screen.findByRole('alert')).toHaveTextContent('この操作は受け付けられませんでした。');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('shows opponent hand counts only while the deal is in play', async () => {
    renderWithProviders(<PopeJoanPage />);
    expect(await screen.findByRole('img', { name: 'CPU1 の手札 3 枚（裏向き）' })).toBeInTheDocument();
  });

  it('hides opponent hand counts after a deal ends and shows the end phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PopeJoanPhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<PopeJoanPage />);
    expect(await screen.findByTestId('popejoan-deal-result')).toHaveTextContent('席2');
    expect(screen.getByText('ディール終了')).toBeInTheDocument();
    expect(screen.getByText('CPU1: チップ-15')).toBeInTheDocument();
    expect(screen.queryByRole('img', { name: /CPU1 の手札/ })).not.toBeInTheDocument();
    expect(screen.queryByTestId('popejoan-final-standings')).not.toBeInTheDocument();
  });

  it('shows final chip standings with shared ranks for ties only after the game ends', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: PopeJoanPhase.GAME_END,
        players: [
          seat(0, true, { chips: 12 }),
          seat(1, false, { chips: 20 }),
          seat(2, false, { chips: 20 }),
          seat(3, false, { chips: -3 }),
        ],
      }),
    );
    renderWithProviders(<PopeJoanPage />);

    const standings = await screen.findByTestId('popejoan-final-standings');
    expect(standings).toHaveTextContent('1位 CPU1: チップ20');
    expect(standings).toHaveTextContent('1位 CPU2: チップ20');
    expect(standings).toHaveTextContent('3位 あなた: チップ12');
    expect(standings).toHaveTextContent('4位 CPU3: チップ-3');
  });

  it('shows both rules permanently', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    fireEvent.click(screen.getByTestId('pj-rule-details').querySelector('summary')!);
    expect(screen.getByText(/トランプの札でしか取れません/)).toBeInTheDocument();
    expect(screen.getByText(/♦8 が抜いてあるので/)).toBeInTheDocument();
    expect(screen.getByText(/めくり札がポープ.*ディーラーが対応する区画を即時に獲得/)).toBeInTheDocument();
  });

  it('explains the turn-up dealer award in English', async () => {
    const previousLanguage = i18n.language;
    try {
      await i18n.changeLanguage('en');
      renderWithProviders(<PopeJoanPage />);
      expect(
        await screen.findByText(
          /when the turn-up is the Pope, ace, king, queen, or jack, the dealer immediately takes the matching compartment/i,
        ),
      ).toBeInTheDocument();
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('shows which trump combinations can be targeted from the hand', async () => {
    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, true, { cards: [card('SPADE', 13), card('SPADE', 12), card('HEART', 11)] })] }),
    );
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByTestId('popejoan-targets')).toHaveTextContent('マトリモニー（K-Q）'));
    expect(screen.getByTestId('popejoan-targets')).not.toHaveTextContent('イントリーグ（Q-J）');
  });

  it('shows both targets with a separator, and none without a target', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, true, { cards: [card('SPADE', 13), card('SPADE', 12), card('SPADE', 11)] })],
      }),
    );
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() =>
      expect(screen.getByTestId('popejoan-targets')).toHaveTextContent('マトリモニー（K-Q） / イントリーグ（Q-J）'),
    );

    mockExec.mockResolvedValue(
      makeState({ players: [seat(0, true, { cards: [card('HEART', 13), card('SPADE', 12)] })] }),
    );
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getAllByTestId('popejoan-targets')[1]).toHaveTextContent('狙える役: なし'));
  });

  // **8 区画すべてが出ていないと、持ち越しがどこに乗っているか読めない。**
  it("shows all eight compartments with the dealer's fixed dress", async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getAllByTestId('popejoan-compartment')).toHaveLength(8));
    const pope = screen.getAllByTestId('popejoan-compartment').find((el) => el.textContent?.includes('ポープ'));
    expect(pope).toHaveTextContent('6');
  });

  // めくり札での即取りは通常の獲得と区別して読めなければならない。
  it('marks a turn-up award apart from an ordinary one', async () => {
    mockExec.mockResolvedValue(makeState({ awards: [{ compartment: 'pope', player: 0, chips: 6, byTurnUp: true }] }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByTestId('popejoan-awards')).toHaveTextContent('めくり札'));
    await waitFor(() =>
      expect(screen.getByTestId('popejoan-live')).toHaveTextContent('席0 がめくり札で ポープ を獲得（6）'),
    );
  });

  it('announces only awards added by each response', async () => {
    const first = { compartment: 'pope' as const, player: 0, chips: 6, byTurnUp: false };
    const second = { compartment: 'ace' as const, player: 1, chips: 3, byTurnUp: false };
    const third = { compartment: 'king' as const, player: 2, chips: 4, byTurnUp: false };
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockResolvedValueOnce(makeState({ awards: [first] }));
    pickHand(0);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(screen.getByTestId('popejoan-live')).toHaveTextContent('席0 が ポープ を獲得（6）'));

    mockExec.mockResolvedValueOnce(makeState({ awards: [first, second] }));
    pickHand(0);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(screen.getByTestId('popejoan-live')).toHaveTextContent('席1 が A を獲得（3）'));
    expect(screen.getByTestId('popejoan-live')).not.toHaveTextContent('ポープ');

    mockExec.mockResolvedValueOnce(makeState({ awards: [first, second, third, { ...third, player: 3 }] }));
    pickHand(0);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() =>
      expect(screen.getByTestId('popejoan-live')).toHaveTextContent('席2 が K を獲得（4）、席3 が K を獲得（4）'),
    );
    expect(screen.getByTestId('popejoan-live')).not.toHaveTextContent('ポープ');
  });

  it('announces the deal result through the live region', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PopeJoanPhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByTestId('popejoan-live')).toHaveTextContent('席2 が出し切りました'));
  });

  it('announces new awards and the deal result from the same response', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    mockExec.mockResolvedValueOnce(
      makeState({
        phase: PopeJoanPhase.DEAL_END,
        awards: [{ compartment: 'pope', player: 0, chips: 6, byTurnUp: false }],
        dealWinner: 2,
      }),
    );
    pickHand(0);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));

    await waitFor(() => {
      expect(screen.getByTestId('popejoan-live')).toHaveTextContent('席0 が ポープ を獲得（6）、席2 が出し切りました');
    });
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    pickHand(1);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  // 止まっているかどうかで出せる札がまるで違うので、案内も変わる。
  it('tells a stopped run apart from one in progress', async () => {
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByText(/最も低い札を選んでください/)).toBeInTheDocument());
  });

  it('asks for the next higher card while a run is live', async () => {
    mockExec.mockResolvedValue(makeState({ runSuit: 1, runRank: 5, playedPile: [card('SPADE', 5)] }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByText(/同じスートの次に高い札/)).toBeInTheDocument());
  });

  // **Pope 保持者は支払いを免除される。**伏せ手でも見えていないと精算が読めない。
  it('marks a hidden opponent as holding the Pope', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, true), seat(1, false, { holdsPope: true }), seat(2, false), seat(3, false)],
      }),
    );
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getAllByText(/ポープ保持/).length).toBeGreaterThan(0));
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PopeJoanPhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<PopeJoanPage />);
    await waitFor(() => expect(screen.getByTestId('popejoan-deal-result')).toHaveTextContent('席2'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'popejoan.win', '最も多く稼ぎました'],
      [2, 'popejoan.lose', '及びませんでした'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: PopeJoanPhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<PopeJoanPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **並びに従う義務がある。**出せない札を押せてしまうと、サーバに弾かれて
  // 初めて分かる (#4934)。
  describe('playable-card restriction', () => {
    it('disables and dims the cards that cannot legally be played', async () => {
      mockExec.mockResolvedValue(makeState({ validPlays: [0] }));
      renderWithProviders(<PopeJoanPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const buttons = screen.getAllByRole('button').filter((b) => b.hasAttribute('data-hint-action'));
      expect(buttons.length).toBeGreaterThan(1);
      expect(buttons[0]).not.toBeDisabled();
      expect(buttons[0]).not.toHaveAttribute('data-unplayable');
      expect(buttons[1]).toBeDisabled();
      expect(buttons[1]).toHaveAttribute('data-unplayable', 'true');
    });

    // **空リストは「一枚も出せない」ではなく「情報が無い」。**空で全部塞ぐと
    // 盤面が操作不能になる。
    it('does not restrict anything when the server sent no list', async () => {
      mockExec.mockResolvedValue(makeState({ validPlays: [] }));
      renderWithProviders(<PopeJoanPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      for (const b of screen.getAllByRole('button').filter((x) => x.hasAttribute('data-hint-action'))) {
        expect(b).not.toBeDisabled();
      }
    });
  });
  // **ディーラーは区画の種銭を負担し、めくり札が Pope/A/K/Q/J ならその区画を総取りする。**
  // 毎ディール回るのに、盤面に「dealer」という語自体が無かった (#6520)。
  it('marks who is dealing this hand', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<PopeJoanPage />);
    expect(await screen.findByTestId('pj-dealer-human')).toHaveTextContent('ディーラー');
    // 印は 1 席にしか付かない。
    expect(screen.queryByTestId('pj-dealer-2')).not.toBeInTheDocument();
  });

  // ディールごとに回るので、席が変われば印も移る。
  it('follows the dealer as it rotates', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 2 }));
    renderWithProviders(<PopeJoanPage />);
    expect(await screen.findByTestId('pj-dealer-2')).toHaveTextContent('ディーラー');
    expect(screen.queryByTestId('pj-dealer-human')).not.toBeInTheDocument();
  });
});
