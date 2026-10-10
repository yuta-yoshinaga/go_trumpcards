import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { nainjauneApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, NainJaunePlayer, NainJauneResponse } from '../types/card';
import { NainJaunePhase } from '../types/phases';
import { NainJaunePage } from './NainJaunePage';

vi.mock('../api/gameApi', () => ({
  nainjauneApi: { exec: vi.fn() },
  actionLogApi: { nainjaune: vi.fn() },
}));

const mockExec = vi.mocked(nainjauneApi.exec);
const mobileState = vi.hoisted(() => ({ isMobile: true }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../hooks/useCardDimensions')>()),
  useIsMobile: () => mobileState.isMobile,
}));

const card = (design: CardDesign, value: number) => ({ design, value });

const BOXES = [
  { name: 'ten', chips: 4, card: card('DIAMOND', 10) },
  { name: 'jack', chips: 8, card: card('CLOVER', 11) },
  { name: 'queen', chips: 12, card: card('SPADE', 12) },
  { name: 'king', chips: 16, card: card('HEART', 13) },
  { name: 'dwarf', chips: 20, card: card('DIAMOND', 7) },
];

function seat(id: number, isHuman: boolean, overrides?: Partial<NainJaunePlayer>): NainJaunePlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 3), card('HEART', 9), card('CLOVER', 13)] : [],
    chips: -15,
    points: 22,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<NainJauneResponse>): NainJauneResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: NainJaunePhase.PLAY,
    validPlays: [],
    currentPlayerIdx: 0,
    boxes: BOXES,
    talonCount: 4,
    awards: [],
    playedPile: [],
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

describe('NainJaunePage', () => {
  beforeEach(() => {
    mobileState.isMobile = true;
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('collapses rule and board explanations and CPU seats on mobile', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const rules = screen.getByTestId('nj-rule-details');
    expect(rules).not.toHaveAttribute('open');
    const ruleText = rules.querySelector('[data-tutorial="nj-rule"]');
    expect(ruleText).not.toBeVisible();
    const ruleSummary = rules.querySelector('summary');
    expect(ruleSummary).toBeInTheDocument();
    if (ruleSummary) fireEvent.click(ruleSummary);
    expect(ruleText).toBeVisible();
    expect(ruleText).toHaveTextContent(/スート無関係.*枚数ではなく【点数】/);

    const board = screen.getAllByTestId('nainjaune-box')[0].closest('[data-tutorial="nj-board"]');
    expect(board).toBeInTheDocument();
    const boardDetails = board?.querySelector('details');
    expect(boardDetails).not.toHaveAttribute('open');
    expect(boardDetails?.querySelector('span')).not.toBeVisible();
    const cpus = screen.getByTestId('cpu-accordion');
    expect(cpus).not.toHaveAttribute('open');
    expect(cpus.querySelector('.flex.flex-wrap')).toBeInTheDocument();
  });

  it('opens the CPU accordion on desktop', async () => {
    mobileState.isMobile = false;
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByTestId('cpu-accordion')).toHaveAttribute('open');
  });

  it('stacks small CPU card backs in one row while keeping the hand count', async () => {
    mobileState.isMobile = false;
    mockExec.mockResolvedValue(
      makeState({
        players: [seat(0, true), seat(1, false, { cardCount: 11, points: 76 }), seat(2, false), seat(3, false)],
      }),
    );
    renderWithProviders(<NainJaunePage />);
    const hand = await screen.findByRole('img', { name: /CPU1 の手札 11 枚/ });
    expect(hand).toHaveClass('flex-nowrap');
    const backs = [...hand.querySelectorAll('img')];
    expect(backs).toHaveLength(11);
    expect(backs.every((back) => back.style.width === '18px')).toBe(true);
    expect(backs.slice(1).every((back) => back.className.includes('-ml-3'))).toBe(true);
    expect(screen.getByText(/CPU1.*手札11枚（76点）/)).toBeInTheDocument();
  });

  it('reveals CPU hands only after a deal ends, while preserving the game-end reveal', async () => {
    const cpuCards = [card('HEART', 4), card('SPADE', 12)];
    mockExec.mockResolvedValue(
      makeState({
        players: [
          seat(0, true),
          seat(1, false, { cards: cpuCards, cardCount: cpuCards.length }),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    const { unmount } = renderWithProviders(<NainJaunePage />);
    const hiddenHand = await screen.findByRole('img', { name: /CPU1 の手札 2 枚/ });
    expect([...hiddenHand.querySelectorAll('img')].every((img) => img.getAttribute('src') === '/images/z01.png')).toBe(
      true,
    );
    unmount();

    mockExec.mockResolvedValue(
      makeState({
        phase: NainJaunePhase.DEAL_END,
        dealWinner: 0,
        players: [
          seat(0, true),
          seat(1, false, { cards: cpuCards, cardCount: cpuCards.length, hidden: false }),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    const { unmount: unmountDeal } = renderWithProviders(<NainJaunePage />);
    const cpuHand = await screen.findByRole('img', { name: /CPU1 の手札 2 枚/ });
    expect([...cpuHand.querySelectorAll('img')].map((img) => img.getAttribute('src'))).toEqual([
      '/images/h04.png',
      '/images/s12.png',
    ]);
    unmountDeal();

    mockExec.mockResolvedValue(
      makeState({
        phase: NainJaunePhase.GAME_END,
        gameEndFlag: true,
        players: [
          seat(0, true),
          seat(1, false, { cards: cpuCards, cardCount: cpuCards.length, hidden: false }),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    renderWithProviders(<NainJaunePage />);
    const gameEndHand = await screen.findByRole('img', { name: /CPU1 の手札 2 枚/ });
    expect([...gameEndHand.querySelectorAll('img')].map((img) => img.getAttribute('src'))).toEqual([
      '/images/h04.png',
      '/images/s12.png',
    ]);
  });

  it('lets the player expand and collapse cards older than the latest ten', async () => {
    const playedPile = Array.from({ length: 12 }, (_, i) => card('SPADE', i + 1));
    mockExec.mockResolvedValue(makeState({ playedPile }));
    renderWithProviders(<NainJaunePage />);

    const showHistory = await screen.findByRole('button', { name: '過去の札を表示' });
    expect(showHistory).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByTestId('nainjaune-played-history')).not.toBeInTheDocument();
    expect(screen.getByTestId('nainjaune-played-recent').querySelectorAll('img')).toHaveLength(10);

    fireEvent.click(showHistory);
    expect(screen.getByRole('button', { name: '過去の札を隠す' })).toHaveAttribute('aria-expanded', 'true');
    expect(screen.getByTestId('nainjaune-played-history').querySelectorAll('img')).toHaveLength(2);
    expect(screen.getByTestId('nainjaune-played-recent').querySelectorAll('img')).toHaveLength(10);

    fireEvent.click(screen.getByRole('button', { name: '過去の札を隠す' }));
    expect(screen.getByRole('button', { name: '過去の札を表示' })).toHaveAttribute('aria-expanded', 'false');
    expect(screen.queryByTestId('nainjaune-played-history')).not.toBeInTheDocument();
    expect(screen.getByTestId('nainjaune-played-recent').querySelectorAll('img')).toHaveLength(10);
  });

  // **区画はスートまで一致した1枚でしか取れない。**札を出さないと判断できない。
  it('shows all five boxes with the exact card that claims each', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getAllByTestId('nainjaune-box')).toHaveLength(5));
    const dwarf = screen.getAllByTestId('nainjaune-box').find((el) => el.textContent?.includes('黄色い小人'));
    expect(dwarf).toHaveTextContent('20');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<NainJaunePage />);
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
    mockExec.mockResolvedValue(makeState({ playedPile: [card('SPADE', 13)] }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByTestId('nainjaune-run-stopped')).toBeInTheDocument());
  });

  it('asks a human to lead any card when the run is stopped', async () => {
    mockExec.mockResolvedValue(makeState({ playedPile: [card('SPADE', 13)], runRank: 0 }));
    renderWithProviders(<NainJaunePage />);
    expect(await screen.findByText(/好きな札から始められます/)).toBeInTheDocument();
  });

  it('shows the stopped-run status while the CPU has the turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, playedPile: [card('SPADE', 13)] }));
    renderWithProviders(<NainJaunePage />);
    expect(await screen.findByTestId('nainjaune-run-stopped')).toHaveTextContent('停止中');
    expect(screen.queryByText(/好きな札から始められます/)).not.toBeInTheDocument();
  });

  // **スートを問わない**のが Pope Joan との決定的な違い。案内にも出す。
  it('asks for the next rank of any suit while a run is live', async () => {
    mockExec.mockResolvedValue(makeState({ runRank: 5, playedPile: [card('SPADE', 5)] }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByText(/6 の札を選んでください（スートは問いません）/)).toBeInTheDocument());
  });

  // 支払いは点数なので、相手の点も見えていないと判断できない。
  it('shows what each hand is worth, not just its size', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getAllByText(/22点/).length).toBeGreaterThan(0));
  });

  it('reports an award', async () => {
    mockExec.mockResolvedValue(makeState({ awards: [{ box: 'dwarf', player: 1, chips: 20 }] }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByTestId('nainjaune-awards')).toHaveTextContent('20'));
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(makeState({ phase: NainJaunePhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByTestId('nainjaune-deal-result')).toHaveTextContent('席2'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'nainjaune.win', '最も多く稼ぎました'],
      [2, 'nainjaune.lose', '及びませんでした'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: NainJaunePhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<NainJaunePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  it('shows final chip standings in descending order and shares rank for ties', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: NainJaunePhase.GAME_END,
        gameEndFlag: true,
        players: [
          seat(0, true, { chips: 30 }),
          seat(1, false, { chips: 50 }),
          seat(2, false, { chips: 50 }),
          seat(3, false, { chips: -5 }),
        ],
      }),
    );
    renderWithProviders(<NainJaunePage />);

    const rows = await screen.findByTestId('nainjaune-final-standings');
    expect(rows.children).toHaveLength(4);
    expect(rows.children[0]).toHaveTextContent('1位');
    expect(rows.children[0]).toHaveTextContent('CPU1');
    expect(rows.children[0]).toHaveTextContent('50');
    expect(rows.children[1]).toHaveTextContent('1位');
    expect(rows.children[1]).toHaveTextContent('CPU2');
    expect(rows.children[2]).toHaveTextContent('3位');
    expect(rows.children[2]).toHaveTextContent('あなた');
    expect(rows.children[2]).toHaveTextContent('30');
    expect(rows.children[3]).toHaveTextContent('4位');
    expect(rows.children[3]).toHaveTextContent('-5');
  });

  it('localizes final chip standings in English', async () => {
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      mockExec.mockResolvedValue(
        makeState({
          phase: NainJaunePhase.GAME_END,
          gameEndFlag: true,
          players: [seat(0, true, { chips: 25 }), seat(1, false, { chips: 40 }), seat(2, false, { chips: 10 })],
          playedPile: Array.from({ length: 11 }, (_, i) => card('SPADE', i + 1)),
        }),
      );
      renderWithProviders(<NainJaunePage />);

      expect(await screen.findByRole('heading', { name: 'Final chip standings' })).toBeInTheDocument();
      const standings = screen.getByTestId('nainjaune-final-standings');
      expect(standings.children[0]).toHaveTextContent('#1: CPU1 — 40 chips');
      expect(standings.children[1]).toHaveTextContent('#2: You — 25 chips');
      expect(standings.children[2]).toHaveTextContent('#3: CPU2 — 10 chips');
      expect(screen.getByRole('button', { name: 'Show earlier cards' })).toHaveAttribute('aria-expanded', 'false');
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  // **並びに従う義務がある。**出せない札を押せてしまうと、サーバに弾かれて
  // 初めて分かる (#4935)。
  describe('playable-card restriction', () => {
    it.each([
      ['CPU turn', { currentPlayerIdx: 1 }],
      ['deal end', { phase: NainJaunePhase.DEAL_END, dealWinner: 1 }],
    ] as const)('disables every hand card and matches its ARIA state during %s', async (_name, overrides) => {
      mockExec.mockResolvedValue(makeState(overrides));
      renderWithProviders(<NainJaunePage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const buttons = screen.getAllByRole('button').filter((b) => b.hasAttribute('data-hint-action'));
      expect(buttons).toHaveLength(3);
      for (const button of buttons) {
        expect(button).toBeDisabled();
        expect(button).toHaveAttribute('aria-disabled', 'true');
      }
    });

    it('disables and dims the cards that cannot legally be played', async () => {
      mockExec.mockResolvedValue(makeState({ validPlays: [0] }));
      renderWithProviders(<NainJaunePage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const buttons = screen.getAllByRole('button').filter((b) => b.hasAttribute('data-hint-action'));
      expect(buttons.length).toBeGreaterThan(1);
      expect(buttons[0]).not.toBeDisabled();
      expect(buttons[0]).toHaveAttribute('aria-disabled', 'false');
      expect(buttons[0]).not.toHaveAttribute('data-unplayable');
      expect(buttons[1]).toBeDisabled();
      expect(buttons[1]).toHaveAttribute('aria-disabled', 'true');
      expect(buttons[1]).toHaveAttribute('data-unplayable', 'true');
    });

    // **空リストは「一枚も出せない」ではなく「情報が無い」。**空で全部塞ぐと
    // 盤面が操作不能になる。
    it('does not restrict anything when the server sent no list', async () => {
      mockExec.mockResolvedValue(makeState({ validPlays: [] }));
      renderWithProviders(<NainJaunePage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      for (const b of screen.getAllByRole('button').filter((x) => x.hasAttribute('data-hint-action'))) {
        expect(b).not.toBeDisabled();
        expect(b).toHaveAttribute('aria-disabled', 'false');
      }
    });
  });

  // #5724: CPU の手番は無言で進むので、手番が自分に回ってきたこと自体が
  // スクリーンリーダー利用者に伝わらなかった (姉妹ゲーム Zheng には既にある仕組み)。
  describe('live region', () => {
    it('is a polite status region and starts empty', async () => {
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
      renderWithProviders(<NainJaunePage />);

      const live = await screen.findByTestId('nainjaune-turn-announce');

      expect(live).toHaveAttribute('role', 'status');
      expect(live).toHaveAttribute('aria-live', 'polite');
      expect(live).toHaveTextContent('');
    });

    it('announces the human turn arriving after CPU turns', async () => {
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
      renderWithProviders(<NainJaunePage />);
      await screen.findByTestId('nainjaune-turn-announce');

      // CPU 手番 → 人間の手番へ遷移させる。
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 0 }));
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));

      await waitFor(() => expect(screen.getByTestId('nainjaune-turn-announce')).toHaveTextContent('あなたの手番です'));
    });

    it('announces the end of the deal', async () => {
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
      renderWithProviders(<NainJaunePage />);
      await screen.findByTestId('nainjaune-turn-announce');

      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, dealWinner: 2 }));
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));

      await waitFor(() =>
        expect(screen.getByTestId('nainjaune-turn-announce')).toHaveTextContent('ディールが終了しました'),
      );
    });

    it('announces the latest award with player, box, and chip count only when awards grow', async () => {
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
      renderWithProviders(<NainJaunePage />);
      await screen.findByTestId('nainjaune-turn-announce');

      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, awards: [{ box: 'dwarf', player: 2, chips: 25 }] }));
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));

      const live = screen.getByTestId('nainjaune-turn-announce');
      await waitFor(() => expect(live).toHaveTextContent('席2 が ♦7 黄色い小人 を獲得（25）'));

      // Receiving the same awards state again must not replace/repeat the award notice.
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, awards: [{ box: 'dwarf', player: 2, chips: 25 }] }));
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));
      await waitFor(() => expect(live).toHaveTextContent('席2 が ♦7 黄色い小人 を獲得（25）'));
    });

    it('announces a newly added human award as you in the live region', async () => {
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
      renderWithProviders(<NainJaunePage />);
      const live = await screen.findByTestId('nainjaune-turn-announce');

      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, awards: [{ box: 'dwarf', player: 0, chips: 20 }] }));
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));

      await waitFor(() => expect(live).toHaveTextContent('あなたが ♦7 黄色い小人 を獲得（20）'));
    });

    it('announces every award added in the same update', async () => {
      mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
      renderWithProviders(<NainJaunePage />);
      await screen.findByTestId('nainjaune-turn-announce');

      mockExec.mockResolvedValue(
        makeState({
          currentPlayerIdx: 1,
          awards: [
            { box: 'ten', player: 1, chips: 10 },
            { box: 'dwarf', player: 2, chips: 25 },
          ],
        }),
      );
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      fireEvent.click(screen.getByRole('button', { name: '確認' }));

      await waitFor(() =>
        expect(screen.getByTestId('nainjaune-turn-announce')).toHaveTextContent(
          '席1 が ♦10 を獲得（10）、席2 が ♦7 黄色い小人 を獲得（25）',
        ),
      );
    });
  });
  // **同じ画面で呼び方を変えない。**手札セクションは「あなたの手札」と呼ぶのに、
  // 獲得通知と出し切り通知だけが「席0が」と出ていた (#6521)。
  it('says "you" for the human in both notices', async () => {
    mockExec.mockResolvedValue(makeState({ awards: [{ box: 'dwarf', player: 0, chips: 20 }] }));
    renderWithProviders(<NainJaunePage />);
    const awards = await screen.findByTestId('nainjaune-awards');
    expect(awards).toHaveTextContent('あなたが');
    expect(awards).not.toHaveTextContent('席0');

    mockExec.mockResolvedValue(makeState({ phase: NainJaunePhase.DEAL_END, dealWinner: 0 }));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(screen.getByTestId('nainjaune-deal-result')).toHaveTextContent('あなたが'));
    expect(screen.getByTestId('nainjaune-deal-result')).not.toHaveTextContent('席0');
  });

  // 負のコントロール: CPU は既存の「席{{n}}」のまま。
  it('keeps the seat wording for a CPU', async () => {
    mockExec.mockResolvedValue(makeState({ awards: [{ box: 'dwarf', player: 2, chips: 20 }] }));
    renderWithProviders(<NainJaunePage />);
    const awards = await screen.findByTestId('nainjaune-awards');
    expect(awards).toHaveTextContent('席2');
    expect(awards).not.toHaveTextContent('あなたが');
  });
});
