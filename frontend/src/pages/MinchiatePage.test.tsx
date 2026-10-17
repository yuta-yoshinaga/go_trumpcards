import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { minchiateApi } from '../api/gameApi';
import i18n from '../i18n';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMinchiateState } from '../test/stateFactories';
import { MINCHIATE_SURPLUS } from '../types/card';
import { MinchiatePage } from './MinchiatePage';

vi.mock('../api/gameApi', () => ({
  minchiateApi: { exec: vi.fn() },
  actionLogApi: { minchiate: vi.fn() },
}));

const mobileState = vi.hoisted(() => ({ isMobile: undefined as boolean | undefined }));
vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useCardDimensions: () => {
      const dimensions = actual.useCardDimensions();
      return mobileState.isMobile === undefined ? dimensions : { ...dimensions, isMobile: mobileState.isMobile };
    },
  };
});

const mockExec = vi.mocked(minchiateApi.exec);

const playState = makeMinchiateState();
const scartoState = makeMinchiateState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
});
const cpuTurnState = makeMinchiateState({ isHumanTurn: false, currentPlayerIdx: 1, playableIndices: [] });
const trickEndState = makeMinchiateState({ phase: 2, isHumanTurn: false, playableIndices: [] });
const roundEndState = makeMinchiateState({
  phase: 3,
  isHumanTurn: false,
  playableIndices: [],
  roundTricks: [5, 4, 3, 3],
  roundBreakdown: [
    { tricks: 8, lastTrickBonus: 0, scartoBonus: 13, total: 21 },
    { tricks: 5, lastTrickBonus: 3, scartoBonus: 0, total: 8 },
  ],
});

beforeEach(() => {
  mobileState.isMobile = undefined;
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('MinchiatePage', () => {
  it('renders the game message in the footer message slot', async () => {
    mockExec.mockResolvedValue(makeMinchiateState({ message: 'Footer message location' }));
    renderWithProviders(<MinchiatePage />);

    const messageSlot = await screen.findByTestId('game-footer-message');
    expect(messageSlot).toHaveTextContent('Footer message location');
  });

  it('renders the game message in the play area on mobile', async () => {
    mockExec.mockResolvedValue(makeMinchiateState({ message: 'Footer message location' }));
    mobileState.isMobile = true;
    renderWithProviders(<MinchiatePage />);

    const messageText = await screen.findByText('Footer message location');
    const message = messageText.closest('[role="status"]');
    expect(message).toHaveTextContent('Footer message location');
    expect(screen.queryByTestId('game-footer-message')).not.toBeInTheDocument();
  });

  it('uses the taller mobile footer cap', async () => {
    renderWithProviders(<MinchiatePage />);
    const content = await screen.findByTestId('game-footer-content');
    expect(content.closest('footer')?.className).toContain('max-h-[60vh]');
  });
  it('stacks the mobile hand rows when the hand spans at least two rows', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      const cards = Array.from({ length: 13 }, (_, i) => ({
        design: 'SPADE' as const,
        value: (i % 14) + 1,
        glyph: '♠',
        label: `Card ${i}`,
        color: 'black' as const,
        deck: 'tarot' as const,
      }));
      mockExec.mockResolvedValue(
        makeMinchiateState({
          players: [{ ...playState.players[0], cards, cardCount: cards.length }, ...playState.players.slice(1)],
        }),
      );
      renderWithProviders(<MinchiatePage />);
      await screen.findByTestId('phase-indicator');
      const hand = document.querySelector<HTMLElement>('[data-tutorial="minchiate-player-hand"]');
      expect(hand).not.toBeNull();
      if (!hand) throw new Error('Expected the Minchiate player hand');
      await waitFor(() => expect(hand.querySelectorAll('[data-testid="hand-row"]').length).toBeGreaterThanOrEqual(2));
      const rows = hand.querySelectorAll<HTMLElement>('[data-testid="hand-row"]');
      for (const row of Array.from(rows).slice(1)) expect(row.style.marginTop).toMatch(/^-/);
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('keeps the hand and scarto progress in the footer', async () => {
    mockExec.mockResolvedValue(scartoState);
    renderWithProviders(<MinchiatePage />);

    const footer = await screen.findByTestId('game-footer-content');
    const playArea = screen.getByTestId('minchiate-trump-count-note').closest('.flex-1.overflow-y-auto');
    expect(playArea).not.toBeNull();
    expect(footer.querySelector('[data-tutorial="minchiate-player-hand"]')).not.toBeNull();
    expect(within(footer).getByTestId('minchiate-scarto-progress')).toBeInTheDocument();
    expect(within(playArea as HTMLElement).queryByTestId('minchiate-scarto-progress')).not.toBeInTheDocument();
    expect(within(footer).queryByTestId('minchiate-hint-live')).not.toBeInTheDocument();
    expect(within(playArea as HTMLElement).getByTestId('minchiate-hint-live')).toBeInTheDocument();
    expect(footer.querySelector('[data-tutorial="minchiate-reset-button"]')).toBeNull();
    expect(playArea?.querySelector('[data-tutorial="minchiate-reset-button"]')).not.toBeNull();
  });

  it('keeps the tutorial action target present during a CPU turn with no actions', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    const { container } = renderWithProviders(<MinchiatePage />);
    expect(await screen.findByTestId('phase-indicator')).toBeInTheDocument();
    expect(container.querySelector('[data-tutorial="minchiate-action-buttons"]')).toBeInTheDocument();
  });

  it('renders minchiate-scarto-done when scartoCount is greater than 0', async () => {
    // スカルト（不要牌捨て）で既に捨てた枚数があるならその旨のメッセージを出す
    const state = makeMinchiateState({ scartoCount: 1 });
    mockExec.mockResolvedValue(state);
    const { getByTestId } = renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(getByTestId('minchiate-scarto-done')).toBeInTheDocument());
  });

  it('does not render minchiate-scarto-done when scartoCount is 0', async () => {
    // スカルトでまだ1枚も捨てていないなら出さない
    const state = makeMinchiateState({ scartoCount: 0 });
    mockExec.mockResolvedValue(state);
    const { queryByTestId, getByTestId } = renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(getByTestId('phase-indicator')).toBeInTheDocument());
    expect(queryByTestId('minchiate-scarto-done')).not.toBeInTheDocument();
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MinchiatePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MinchiatePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetRounds: 4 } }),
    );
  });

  // 切札が40枚あることは手札からは読めない唯一の情報なので、常時表示が必要。
  it('always states the 40-trump count', async () => {
    renderWithProviders(<MinchiatePage />);
    const note = await screen.findByTestId('minchiate-trump-count-note');
    expect(note).toHaveTextContent('切札は40枚');
  });

  it('lists played trumps by rank and shows the remaining count', async () => {
    const state = makeMinchiateState({
      playedTrumps: [
        { design: 'JOKER', value: 4, glyph: '✦', label: '4 Foco', color: 'purple', deck: 'tarot' },
        { design: 'JOKER', value: 39, glyph: '✦', label: '39 Angelo', color: 'purple', deck: 'tarot' },
      ],
      remainingTrumps: 38,
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<MinchiatePage />);
    const tracker = await screen.findByTestId('minchiate-trump-tracker');
    expect(tracker).toHaveTextContent('4 Foco');
    expect(tracker).toHaveTextContent('39 Angelo');
    expect(tracker).toHaveTextContent('38');
  });

  // #5715: マットの特殊ルール (取らない・フォロー免除・リードを定めない) は
  // ヒント文言の中にしか無く、ヒントを切っていると知る手段が無かった。
  it('always explains the Matto', async () => {
    renderWithProviders(<MinchiatePage />);

    const note = await screen.findByTestId('minchiate-matto-note');

    expect(note).toHaveTextContent('マット');
    expect(note).toHaveTextContent('トリックを取らず');
  });

  // 個人戦ではなくチーム戦。対面がパートナーであることが画面に出ていないと、
  // 味方のトリックを奪う手が「正しく」見えてしまう。
  it('shows team scores rather than per-player scores', async () => {
    mockExec.mockResolvedValue(makeMinchiateState({ teamScores: [7, 4] }));
    renderWithProviders(<MinchiatePage />);
    const scores = await screen.findByTestId('minchiate-team-scores');
    expect(scores).toHaveTextContent('7');
    expect(scores).toHaveTextContent('4');
  });

  it('shows the server-provided team round score breakdown', async () => {
    mockExec.mockResolvedValue(
      makeMinchiateState({
        phase: 3,
        roundBreakdown: [
          { tricks: 8, lastTrickBonus: 0, scartoBonus: 13, total: 21 },
          { tricks: 5, lastTrickBonus: 3, scartoBonus: 0, total: 8 },
        ],
      }),
    );
    renderWithProviders(<MinchiatePage />);
    expect(await screen.findByTestId('mc-round-breakdown-0')).toHaveTextContent(
      'チーム0: トリック 8 点 / 最終トリック +0 / スカルト +13 / ラウンド計 21 点',
    );
    expect(screen.getByTestId('mc-round-breakdown-1')).toHaveTextContent(
      'チーム1: トリック 5 点 / 最終トリック +3 / スカルト +0 / ラウンド計 8 点',
    );
  });

  it('plays the selected card', async () => {
    renderWithProviders(<MinchiatePage />);
    const playButton = await screen.findByRole('button', { name: '出す' });
    expect(within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: '出す' })).toBe(playButton);
    expect(playButton).toBeDisabled();

    fireEvent.click(screen.getByRole('button', { name: 'Re ♠' }));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('hides the play control when it is not the human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  describe('scarto phase', () => {
    // 余剰ぶんの札を持つ手を組む。**枚数は定数から。**Tarocchini から写すと
    // 2 枚のままになるが、Minchiate は 13 枚選ばないと捨てられない。
    const scartoHandState = () =>
      makeMinchiateState({
        ...scartoState,
        players: [
          {
            ...scartoState.players[0],
            cardCount: MINCHIATE_SURPLUS,
            cards: Array.from({ length: MINCHIATE_SURPLUS }, (_, i) => ({
              design: 'SPADE' as const,
              value: i + 1,
              glyph: '\u2660',
              label: String(i + 1),
              color: 'black',
              deck: 'tarot',
            })),
          },
          ...scartoState.players.slice(1),
        ],
      });

    it('shows the scarto selection progress in English', async () => {
      await i18n.changeLanguage('en');
      mockExec.mockResolvedValue(scartoHandState());
      renderWithProviders(<MinchiatePage />);

      expect(await screen.findByTestId('minchiate-scarto-progress')).toHaveTextContent('Selected 0/13 cards');
      await i18n.changeLanguage('ja');
    });

    it('prompts for the full surplus and dispatches every index', async () => {
      mockExec.mockResolvedValue(scartoHandState());
      renderWithProviders(<MinchiatePage />);
      expect(await screen.findByTestId('minchiate-scarto-prompt')).toHaveTextContent(String(MINCHIATE_SURPLUS));
      expect(
        within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: '捨てる' }),
      ).toBeInTheDocument();
      const progress = screen.getByTestId('minchiate-scarto-progress');
      expect(progress).toHaveTextContent('選択済み 0/13枚');

      expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();

      // 1 枚足りないうちは押せない。
      for (let i = 1; i < MINCHIATE_SURPLUS; i++) {
        fireEvent.click(screen.getByRole('button', { name: `${i} \u2660` }));
      }
      expect(progress).toHaveTextContent('選択済み 12/13枚');
      expect(screen.getByRole('button', { name: '捨てる' })).toBeDisabled();

      fireEvent.click(screen.getByRole('button', { name: `${MINCHIATE_SURPLUS} \u2660` }));
      expect(progress).toHaveTextContent('選択済み 13/13枚');
      mockExec.mockClear();
      mockExec.mockResolvedValue(scartoHandState());
      fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('scarto', {
          cardIndices: Array.from({ length: MINCHIATE_SURPLUS }, (_, i) => i),
        }),
      );
    });

    it('shows no scarto control when the dealer is a CPU', async () => {
      mockExec.mockResolvedValue(makeMinchiateState({ phase: 0, isHumanScarto: false, isHumanTurn: false }));
      renderWithProviders(<MinchiatePage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByRole('button', { name: '捨てる' })).not.toBeInTheDocument();
    });
  });

  it('advances to the next trick', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<MinchiatePage />);
    const button = await screen.findByRole('button', { name: '次のトリック' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(trickEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances to the next round and shows the trick tally', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MinchiatePage />);
    const button = await screen.findByRole('button', { name: '次のラウンド' });
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
    mockExec.mockClear();
    mockExec.mockResolvedValue(roundEndState);
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('has no bid controls — Minchiate has no bidding phase', async () => {
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: 'パス' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [0], reason: 'lead_trump' } });
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playState,
      hint: { cardIndices: [0], reason: 'lead_trump' },
      messageCode: 'minchiate.hintRequested',
    });
    renderWithProviders(<MinchiatePage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **CUI と CLI からは呼べるのに、盤面には要求する手段が無かった (#4819)。**
  // isRequestedHint は hintRequested の messageCode を待つので、コマンドを
  // 送らないと表示側は永遠に出ない (実質デッドコードだった)。
  it('asks the server for a hint and shows the answer', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<MinchiatePage />);
    const hintBtn = await screen.findByTestId('minchiate-hint-button');

    mockExec.mockClear();
    mockExec.mockResolvedValue({
      ...playState,
      messageCode: 'minchiate.hintRequested',
      hint: { cardIndices: [2], reason: 'lead_trump' },
    });
    fireEvent.click(hintBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    await waitFor(() => expect(screen.getByText(/\[2\]/)).toBeInTheDocument());
  });

  it('shows no hint line until one is requested', async () => {
    mockExec.mockResolvedValue({ ...playState, hint: { cardIndices: [2], reason: 'lead_trump' } });
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(screen.getByTestId('minchiate-hint-button')).toBeInTheDocument());
    expect(screen.queryByText(/\[2\]/)).not.toBeInTheDocument();
  });
  // **トリック数を足しても teamScores の増分と合わない。**精算には最終トリック
  // ボーナスとスカルト枚数分が乗っているのに、どちらの画面にも出ていなかった (#6512)。
  it('breaks the settlement bonuses out of the trick tally', async () => {
    // 席 1 はチーム 1、ディーラー席 0 はチーム 0 (対面同士 0-2 / 1-3)。
    mockExec.mockResolvedValue({ ...roundEndState, lastTrickWinner: 1, dealerIdx: 0, scartoCount: 4 });
    renderWithProviders(<MinchiatePage />);

    const lastTrick = await screen.findByTestId('mc-last-trick-bonus');
    expect(lastTrick).toHaveTextContent('チーム1');
    expect(lastTrick).toHaveTextContent('3');
    const scarto = screen.getByTestId('mc-scarto-bonus');
    expect(scarto).toHaveTextContent('チーム0');
    expect(scarto).toHaveTextContent('4');
    expect(lastTrick.textContent).not.toContain('{{');
  });

  // 負のコントロール: 起きていない加点は行ごと出さない。
  it('shows no bonus rows when neither bonus applies', async () => {
    mockExec.mockResolvedValue({ ...roundEndState, lastTrickWinner: -1, scartoCount: 0 });
    renderWithProviders(<MinchiatePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('mc-last-trick-bonus')).not.toBeInTheDocument();
    expect(screen.queryByTestId('mc-scarto-bonus')).not.toBeInTheDocument();
  });

  // 催促はフェーズや手番が変わったときに現れるテキスト。領域が無いと、あるいは
  // 領域が催促と**同時に**生えると、スクリーンリーダには何も届かない (#6880)。
  it('announces the minchiate-scarto-prompt from the always-mounted live region', async () => {
    mockExec.mockResolvedValue(scartoState);
    renderWithProviders(<MinchiatePage />);

    const live = await screen.findByTestId('minchiate-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 隣に置いただけの実装は属性の検査を通る。**中にあること**を見る。
    expect(live).toContainElement(await screen.findByTestId('minchiate-scarto-prompt'));

    const progress = await screen.findByTestId('minchiate-scarto-progress');
    expect(progress.closest('[aria-live]')).toBeNull();
    const promptLive = screen.getByTestId('minchiate-prompt-live');
    expect(progress.parentElement).not.toBeNull();
    expect(promptLive.parentElement).not.toBeNull();
    expect(progress.parentElement).toBe(promptLive.parentElement);
    // On one line, not two footer rows: the shared parent is the prompt line, not the footer content.
    expect(progress.parentElement).not.toBe(screen.getByTestId('game-footer-content'));
  });
});
