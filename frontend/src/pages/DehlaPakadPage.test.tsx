import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { dehlaPakadApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeDehlaPakadState } from '../test/stateFactories';
import { DehlaPakadPage } from './DehlaPakadPage';

vi.mock('../api/gameApi', () => ({
  dehlaPakadApi: { exec: vi.fn() },
  actionLogApi: { dehlapakad: vi.fn() },
}));

const mockExec = vi.mocked(dehlaPakadApi.exec);
const mobileState = vi.hoisted(() => ({ isMobile: true }));

vi.mock('../hooks/useCardDimensions', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/useCardDimensions')>();
  return {
    ...actual,
    useIsMobile: () => mobileState.isMobile,
    useCardDimensions: () => ({ ...actual.useCardDimensions(), isMobile: mobileState.isMobile }),
  };
});

const trumpState = makeDehlaPakadState();

/** Hearts called, the human on turn with all 13 cards. */
const playState = makeDehlaPakadState({
  phase: 'play',
  isTrumpPhase: false,
  trumpSuit: 3,
  trumpSuitName: 'heart',
  currentPlayerIdx: 0,
  playableIndices: [0, 1, 2, 3, 4],
  players: trumpState.players.map((p) => ({ ...p, cardCount: 13, isTrumpChooser: false })),
});

beforeEach(() => {
  mobileState.isMobile = true;
  mockExec.mockReset();
  mockExec.mockResolvedValue(trumpState);
});

describe('DehlaPakadPage', () => {
  it('collapses reference information and CPU seats on mobile', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        handHistory: [{ winnerTeam: 0, teamTens: [3, 1], kot: false, kotReason: '', dealerIdx: 3, trumpSuit: 1 }],
      }),
    );
    renderWithProviders(<DehlaPakadPage />);

    const cpu = await screen.findByTestId('cpu-accordion');
    const history = screen.getByTestId('dehlapakad-hand-history');
    const earnedTens = screen.getByTestId('dehlapakad-earned-tens');
    expect(cpu).not.toHaveAttribute('open');
    expect(cpu.querySelector('div.px-1.pb-1')).not.toBeVisible();
    expect(history.closest('details')).not.toHaveAttribute('open');
    expect(history).not.toBeVisible();
    expect(earnedTens.closest('details')).not.toHaveAttribute('open');
    expect(earnedTens).not.toBeVisible();
    expect(screen.getByTestId('dehlapakad-scores')).toBeVisible();
  });

  it('opens reference information and CPU seats on desktop', async () => {
    mobileState.isMobile = false;
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        handHistory: [{ winnerTeam: 0, teamTens: [3, 1], kot: false, kotReason: '', dealerIdx: 3, trumpSuit: 1 }],
      }),
    );
    renderWithProviders(<DehlaPakadPage />);

    const cpu = await screen.findByTestId('cpu-accordion');
    expect(cpu).toHaveAttribute('open');
    expect(screen.getByTestId('dehlapakad-hand-history').closest('details')).toHaveAttribute('open');
    expect(screen.getByTestId('dehlapakad-earned-tens').closest('details')).toHaveAttribute('open');
  });

  it('shows the five-card suit breakdown, including empty suits', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        players: trumpState.players.map((p) =>
          p.isHuman
            ? {
                ...p,
                cards: [
                  { design: 'SPADE' as const, value: 1, color: 'black' as const },
                  { design: 'SPADE' as const, value: 2, color: 'black' as const },
                  { design: 'CLOVER' as const, value: 3, color: 'black' as const },
                  { design: 'DIAMOND' as const, value: 4, color: 'red' as const },
                  { design: 'DIAMOND' as const, value: 5, color: 'red' as const },
                ],
              }
            : p,
        ),
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    const breakdown = await screen.findByTestId('dehlapakad-trump-breakdown');
    expect(breakdown.querySelector('[aria-label="スペード（♠） 2"]')).toBeInTheDocument();
    expect(breakdown.querySelector('[aria-label="クラブ（♣） 1"]')).toBeInTheDocument();
    expect(breakdown.querySelector('[aria-label="ハート（♥） 0"]')).toBeInTheDocument();
    expect(breakdown.querySelector('[aria-label="ダイヤ（♦） 2"]')).toBeInTheDocument();
  });

  it('calls reset on mount with the configured match length', async () => {
    renderWithProviders(<DehlaPakadPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetKots: 2 } }),
    );
  });

  it('shows the hand counter and the match target', async () => {
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByText('ハンド 1（2 コート先取）')).toBeInTheDocument();
  });

  // **決めるのは親の右隣で、見えているのは最初の 5 枚だけ。**
  it('offers the four suits while the trump is being called', async () => {
    renderWithProviders(<DehlaPakadPage />);
    const box = await screen.findByTestId('dehlapakad-trump-choices');
    expect(box.querySelectorAll('button')).toHaveLength(4);
    fireEvent.click(screen.getByTestId('dehlapakad-trump-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 3 }));
  });

  it('hides the suit buttons once the trump is called', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-trump-choices')).not.toBeInTheDocument();
    expect(screen.getByTestId('dehlapakad-trump')).toHaveTextContent('ハート');
  });

  it('shows the previous trick and its winner when no trick is in progress', async () => {
    const previousTrick = [
      { playerIdx: 0, card: { design: 'HEART' as const, value: 10, color: 'red' as const } },
      { playerIdx: 2, card: { design: 'HEART' as const, value: 9, color: 'red' as const } },
    ];
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ ...playState, currentTrick: [], lastTrick: previousTrick, lastTrickWinner: 0 }),
    );
    renderWithProviders(<DehlaPakadPage />);

    expect(await screen.findByText('直前のトリック')).toBeInTheDocument();
    const cards = screen.getByTestId('trick-display-cards');
    expect(cards).toHaveTextContent('あなた');
    expect(cards).toHaveTextContent('CPU 2');
    expect(cards.querySelector('.ring-ds-warning')).toBeInTheDocument();
  });

  it('shows only the current trick while it has cards', async () => {
    const currentTrick = [{ playerIdx: 0, card: { design: 'HEART' as const, value: 10, color: 'red' as const } }];
    const previousTrick = [{ playerIdx: 1, card: { design: 'SPADE' as const, value: 9, color: 'black' as const } }];
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ ...playState, currentTrick, lastTrick: previousTrick, lastTrickWinner: 1 }),
    );
    renderWithProviders(<DehlaPakadPage />);

    expect(await screen.findByText('現在のトリック')).toBeInTheDocument();
    expect(screen.getByTestId('trick-display-cards')).not.toHaveTextContent('CPU 1');
  });

  it('does not render a trick area before the first trick', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, currentTrick: [], lastTrick: [] }));
    renderWithProviders(<DehlaPakadPage />);

    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('trick-display-cards')).not.toBeInTheDocument();
  });

  it('plays the selected card', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<DehlaPakadPage />);
    const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByTestId('dehlapakad-play'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  // **これがこのゲームの心臓部。** 取っただけでは札は手に入らない。
  it('shows what is sitting in the centre and who would collect it', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ ...playState, centrePileCount: 8, centrePileTens: 2, prevTrickWinner: 1 }),
    );
    renderWithProviders(<DehlaPakadPage />);
    const pile = await screen.findByTestId('dehlapakad-centre-pile');
    expect(pile).toHaveTextContent('中央に 8 枚（うち 10 が 2 枚）');
    expect(screen.getByTestId('dehlapakad-pile-goes-to')).toHaveTextContent('CPU 1');
  });

  it('hides the centre pile while nothing is waiting', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-centre-pile')).not.toBeInTheDocument();
  });

  it('reads the team columns from the human seat', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, humanTeam: 1, teamTens: [3, 1], teamKots: [1, 0] }));
    renderWithProviders(<DehlaPakadPage />);
    const box = await screen.findByTestId('dehlapakad-scores');
    expect(box).toHaveTextContent('味方 1 / 相手 3');
    expect(box).toHaveTextContent('味方 0 / 相手 1');
  });

  it('shows captured ten cards under their teams, separately from the centre pile', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        ...playState,
        centrePileCount: 4,
        centrePileTens: 1,
        players: playState.players.map((player, i) => ({
          ...player,
          gatheredCards:
            i === 0
              ? [
                  { design: 'SPADE', value: 10, color: 'black' },
                  { design: 'HEART', value: 10, color: 'red' },
                ]
              : i === 1
                ? [{ design: 'DIAMOND', value: 10, color: 'red' }]
                : [],
        })),
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    const teamZero = await screen.findByTestId('dehlapakad-earned-tens-team-0');
    const teamOne = screen.getByTestId('dehlapakad-earned-tens-team-1');
    expect(teamZero).toHaveTextContent('組0 (2)');
    expect(teamZero.querySelectorAll('img')).toHaveLength(2);
    expect(teamOne).toHaveTextContent('組1 (1)');
    expect(teamOne.querySelectorAll('img')).toHaveLength(1);
    expect(screen.getByTestId('dehlapakad-centre-pile')).toHaveTextContent('中央に 4 枚');
  });

  // **7 連勝もコートになる。** 出さないと、なぜ同じ組が勝ち続けているのかが
  // 数字にならない。
  it('surfaces a winning streak', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, streakTeam: 1, streakCount: 4 }));
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-streak')).toHaveTextContent('4 連勝');
  });

  it('does not claim a streak on a single win', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, streakTeam: 1, streakCount: 1 }));
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-streak')).not.toBeInTheDocument();
  });

  it('advances the hand and shows the tens breakdown', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        phase: 'handEnd',
        isTrumpPhase: false,
        isHumanTurn: false,
        lastHand: { winnerTeam: 0, teamTens: [3, 1], kot: false, kotReason: '', dealerIdx: 3, trumpSuit: 3 },
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-hand-result')).toHaveTextContent('10 の枚数 3 対 1');
    expect(screen.queryByTestId('dehlapakad-kot')).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId('dehlapakad-next-hand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nexthand'));
  });

  it('shows completed hand history separately from the current hand score', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        ...playState,
        handNumber: 3,
        teamTens: [1, 3],
        teamKots: [1, 0],
        handHistory: [
          { winnerTeam: 0, teamTens: [3, 1], kot: true, kotReason: 'allTens', dealerIdx: 3, trumpSuit: 1 },
          { winnerTeam: 1, teamTens: [1, 3], kot: false, kotReason: '', dealerIdx: 2, trumpSuit: 2 },
          { winnerTeam: 0, teamTens: [2, 2], kot: false, kotReason: '', dealerIdx: 1, trumpSuit: 0 },
        ],
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    const history = await screen.findByTestId('dehlapakad-hand-history');
    expect(history).toHaveTextContent('ハンド 1');
    expect(history).toHaveTextContent('組0');
    expect(history).toHaveTextContent('3 対 1');
    expect(history).toHaveTextContent('コート');
    expect(history).toHaveTextContent('ハンド 2');
    expect(history).toHaveTextContent('組1');
    expect(history).toHaveTextContent('1 対 3');
    expect(history).toHaveTextContent('切り札');
    expect(history).toHaveTextContent('スペード（♠）');
    expect(history).toHaveTextContent('クラブ（♣）');
    expect(history).toHaveTextContent('-');
    expect(history).not.toHaveTextContent('ハート（♥）');
    expect(screen.getByTestId('dehlapakad-scores')).toHaveTextContent('10（デーラ）— 味方 1 / 相手 3');
  });

  it('shows the last trick winner with their team only at hand end', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        phase: 'handEnd',
        isTrumpPhase: false,
        isHumanTurn: false,
        lastTrickWinner: 2,
        lastHand: { winnerTeam: 0, teamTens: [3, 1], kot: false, kotReason: '', dealerIdx: 3, trumpSuit: 3 },
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-last-trick-winner')).toHaveTextContent(
      '最終トリックの勝者: CPU 2（組0）',
    );
  });

  it('does not show last trick winner during play', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, lastTrickWinner: 2 }));
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-last-trick-winner')).not.toBeInTheDocument();
  });

  it.each([
    ['allTens', '10 を 4 枚とも取りました'],
    ['streak', '7 連勝を達成しました'],
  ])('names the reason for a kot (%s)', async (kotReason, text) => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        phase: 'handEnd',
        isTrumpPhase: false,
        isHumanTurn: false,
        lastHand: { winnerTeam: 0, teamTens: [4, 0], kot: true, kotReason, dealerIdx: 3, trumpSuit: 3 },
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-kot')).toHaveTextContent(text);
  });

  it('names the winning team at the end of the match', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        phase: 'gameEnd',
        gameEndFlag: true,
        isTrumpPhase: false,
        winnerTeam: 1,
        lastTrickWinner: 2,
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-winner')).toHaveTextContent('組1 の勝ちです');
    expect(screen.queryByTestId('dehlapakad-last-trick-winner')).not.toBeInTheDocument();
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ ...playState, hint: { cardIndices: [0], reason: 'take_the_ten' }, messageCode: '' }),
    );
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        ...playState,
        hint: { cardIndices: [0], reason: 'take_the_ten' },
        messageCode: 'dehlapakad.hintRequested',
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **相手が選んでいる間の空白 (#6626)。** 切り札を決めるのは親の右隣で、
  // そこが CPU 席だと canPlay/canCallTrump/isHandEnd のどれも立たず、
  // フッターがまるごと空白になっていた。
  describe('trump selection by a CPU', () => {
    it('says who is choosing while a CPU picks the trump', async () => {
      mockExec.mockResolvedValue(makeDehlaPakadState({ isHumanTurn: false, trumpChooserIdx: 2 }));
      renderWithProviders(<DehlaPakadPage />);
      const waiting = await screen.findByTestId('dehlapakad-trump-waiting');
      // 席まで名指しする。「選んでいます」だけだと誰の番か分からない。
      expect(waiting.textContent).toContain('CPU 2');
      expect(waiting.textContent).not.toContain('{{');
      // 選ぶ側のボタンは出さない。
      expect(screen.queryByTestId('dehlapakad-trump-choices')).not.toBeInTheDocument();
    });

    it('moves the name with the seat', async () => {
      mockExec.mockResolvedValue(makeDehlaPakadState({ isHumanTurn: false, trumpChooserIdx: 3 }));
      renderWithProviders(<DehlaPakadPage />);
      const waiting = await screen.findByTestId('dehlapakad-trump-waiting');
      expect(waiting.textContent).toContain('CPU 3');
      expect(waiting.textContent).not.toContain('CPU 2');
    });

    // **人間が選ぶ番の表示は変えない** (受け入れ条件)。
    it('keeps the human choices untouched and shows no waiting text', async () => {
      mockExec.mockResolvedValue(trumpState);
      renderWithProviders(<DehlaPakadPage />);
      expect(await screen.findByTestId('dehlapakad-trump-choices')).toBeInTheDocument();
      expect(screen.queryByTestId('dehlapakad-trump-waiting')).not.toBeInTheDocument();
    });

    // プレイフェーズでは待機テキストを出さない (フェーズガードの負のコントロール)。
    it('shows no waiting text outside the trump phase', async () => {
      mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, isHumanTurn: false, trumpChooserIdx: 2 }));
      renderWithProviders(<DehlaPakadPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('dehlapakad-trump-waiting')).not.toBeInTheDocument();
    });
  });
});
