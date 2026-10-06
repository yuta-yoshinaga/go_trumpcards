import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { piedmonteseTarotApi } from '../api/gameApi';
import i18n from '../i18n';
import { renderWithProviders } from '../test/renderWithProviders';
import { makePiedmonteseTarotState } from '../test/stateFactories';
import { PiedmonteseTarotPhase } from '../types/phases';
import { PiedmonteseTarotPage } from './PiedmonteseTarotPage';

vi.mock('../api/gameApi', () => ({
  piedmonteseTarotApi: { exec: vi.fn() },
  actionLogApi: { piedmontesetarot: vi.fn() },
}));

const mockExec = vi.mocked(piedmonteseTarotApi.exec);

const suit = (value: number, design: 'HEART' | 'SPADE' | 'CLOVER' | 'DIAMOND', glyph: string, label: string) => ({
  design,
  value,
  glyph,
  label,
  color: design === 'HEART' || design === 'DIAMOND' ? 'red' : 'black',
  deck: 'tarot',
});

const playState = makePiedmonteseTarotState();

/** The human deals: two cards to bury, a hand mixing pips, a Roi and the Matto. */

// ピップが 1 枚しかない親の手。ドメインは非オヌール切り札 (index 1) を解禁するので、
// サーバはそれを discardableIndices に載せて返す。
const shortOnPipsState = makePiedmonteseTarotState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
  discardableIndices: [0, 1],
  dealerIdx: 0,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 21,
      cards: [
        suit(2, 'HEART', '♥', '2'),
        { design: 'JOKER' as const, value: 10, glyph: 'A', label: 'T10', color: 'purple', deck: 'tarot' },
        { design: 'JOKER' as const, value: 21, glyph: 'A', label: 'T21', color: 'purple', deck: 'tarot' },
        { design: 'JOKER' as const, value: 0, glyph: '★', label: 'Matto', color: 'gold', deck: 'tarot' },
      ],
      trickCount: 0,
      cardThirds: 0,
      scartoThirds: 0,
      cardPoints: '0',
      score: 0,
      isDealer: true,
    },

    ...makePiedmonteseTarotState()
      .players.slice(1)
      .map((p) => ({ ...p, isDealer: false })),
  ],
});

const scartoState = makePiedmonteseTarotState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
  // サーバが返す「捨てられる札」。ピップ 2 枚でタロン 2 枚ぶん足りている。
  discardableIndices: [0, 1],
  dealerIdx: 0,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 21,
      cards: [
        suit(2, 'HEART', '♥', '2'),
        suit(3, 'HEART', '♥', '3'),
        suit(14, 'SPADE', '♠', 'R'),
        { design: 'JOKER' as const, value: 0, glyph: '★', label: 'Matto', color: 'gold', deck: 'tarot' },
      ],
      trickCount: 0,
      cardThirds: 0,
      scartoThirds: 0,
      cardPoints: '0',
      score: 0,
      isDealer: true,
    },
    ...makePiedmonteseTarotState()
      .players.slice(1)
      .map((p) => ({ ...p, isDealer: false })),
  ],
});

/** Three-handed table: the talon is three cards, not two. */
const threeHandedScartoState = makePiedmonteseTarotState({
  ...scartoState,
  talonSize: 3,
  trickCount: 25,
  config: { seats: 3, cpuDifficulty: 1, targetDeals: 4 },
});

const roundEndState = makePiedmonteseTarotState({
  phase: 3,
  isHumanTurn: false,
  playableIndices: [],
  outcome: 1,
  scartoCards: [
    suit(2, 'HEART', '♥', '2'),
    { design: 'JOKER' as const, value: 21, glyph: '✦', label: '21', color: 'purple', deck: 'tarot' },
  ],
  dealScores: [12, -4, -4, -4],
  players: makePiedmonteseTarotState().players.map((p, i) => ({
    ...p,
    cardThirds: i === 0 ? 90 : 48,
    scartoThirds: i === 0 ? 6 : 0,
    cardPoints: i === 0 ? '30' : '16',
    score: i === 0 ? 12 : -4,
  })),
});

const roundEndWithoutScartoState = makePiedmonteseTarotState({ ...roundEndState, scartoCards: [] });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('PiedmonteseTarotPage', () => {
  it('shows empty deal score history at the start of a three-seat match', async () => {
    mockExec.mockResolvedValueOnce(
      makePiedmonteseTarotState({
        config: { seats: 3, cpuDifficulty: 1, targetDeals: 2 },
        players: makePiedmonteseTarotState().players.slice(0, 3),
        playerScores: [0, 0, 0],
        dealScores: [0, 0, 0],
        dealScoreHistory: [],
      }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    const history = await screen.findByTestId('piedmontesetarot-deal-score-history');
    expect(within(history).getByText('完了したディールの得点履歴はありません')).toBeInTheDocument();
  });

  it('renders completed deal changes for all four seats', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({
        gameEndFlag: true,
        phase: PiedmonteseTarotPhase.GAME_END,
        playerScores: [4, -1, -3, 0],
        dealScoreHistory: [{ roundNumber: 1, scores: [4, -1, -3, 0] }],
      }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    const history = await screen.findByTestId('piedmontesetarot-deal-score-history');
    expect(within(history).getByText('ディール 1: あなた +4、CPU 1 -1、CPU 2 -3、CPU 3 ±0')).toBeInTheDocument();
  });

  it('renders completed deal changes for three seats', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({
        config: { seats: 3, cpuDifficulty: 1, targetDeals: 2 },
        players: makePiedmonteseTarotState().players.slice(0, 3),
        playerScores: [2, -1, -1],
        dealScoreHistory: [{ roundNumber: 1, scores: [2, -1, -1] }],
      }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    const history = await screen.findByTestId('piedmontesetarot-deal-score-history');
    expect(within(history).getByText('ディール 1: あなた +2、CPU 1 -1、CPU 2 -1')).toBeInTheDocument();
  });

  it('shows the buried cards in the round result with accessible card names', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const scartoCards = await screen.findByTestId('piedmontesetarot-scarto-cards');
    expect(within(scartoCards).getByRole('img', { name: '2 ♥' })).toBeInTheDocument();
    expect(within(scartoCards).getByRole('img', { name: '21 ✦' })).toBeInTheDocument();
  });

  it('formats zero settlement and earned deltas as ±0', async () => {
    const zeroState = makePiedmonteseTarotState({
      ...roundEndState,
      dealScores: [0, 0, 0, 0],
      players: roundEndState.players.map((player) => ({ ...player, score: 0 })),
    });
    mockExec.mockResolvedValue(zeroState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const result = await screen.findByTestId('piedmontesetarot-result');
    expect(result).toHaveTextContent('あなた: ±0（累計 0）');
    expect(result).toHaveTextContent('変動 ±0');
  });

  it('does not show an empty buried cards list when there is no scarto', async () => {
    mockExec.mockResolvedValue(roundEndWithoutScartoState);
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByTestId('piedmontesetarot-result')).toBeInTheDocument();
    expect(screen.queryByTestId('piedmontesetarot-scarto-cards')).not.toBeInTheDocument();
  });
  it('calls reset on mount with the configured table', async () => {
    renderWithProviders(<PiedmonteseTarotPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { seats: 4, cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('shows the deal and trick counters', async () => {
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByText('ディール 1')).toBeInTheDocument();
    expect(screen.getByText('トリック 1/19')).toBeInTheDocument();
  });

  it('plays the selected card', async () => {
    renderWithProviders(<PiedmonteseTarotPage />);
    const cards = await screen.findAllByRole('button', { name: /♥|♠|♣|✦|★/ });
    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  // **捨てる枚数は卓が決める。** 定数にすると 3 人卓でボタンが押せない。
  it('asks for two cards at a four-handed table', async () => {
    mockExec.mockResolvedValue(scartoState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const prompt = await screen.findByTestId('piedmontesetarot-discard-prompt');
    expect(prompt).toHaveTextContent('0/2');

    const button = screen.getByRole('button', { name: /捨てる/ });
    expect(button).toBeDisabled();
    const cards = await screen.findAllByRole('button', { name: /♥|♠|★/ });
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    expect(button).toBeEnabled();
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 1] }));
  });

  it('asks for three cards at a three-handed table', async () => {
    mockExec.mockResolvedValue(threeHandedScartoState);
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByTestId('piedmontesetarot-discard-prompt')).toHaveTextContent('0/3');
    const cards = await screen.findAllByRole('button', { name: /♥|♠|★/ });
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    // 2 枚では足りない ── 4 人卓の枚数で判定していれば、ここで押せてしまう。
    expect(screen.getByRole('button', { name: /捨てる/ })).toBeDisabled();
  });

  it('waits while a CPU deals', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({ phase: 0, isHumanTurn: false, isHumanScarto: false, playableIndices: [] }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByTestId('piedmontesetarot-waiting')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /捨てる/ })).not.toBeInTheDocument();
  });

  it('advances the trick', async () => {
    mockExec.mockResolvedValue(makePiedmonteseTarotState({ phase: 2, isHumanTurn: false, playableIndices: [] }));
    renderWithProviders(<PiedmonteseTarotPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances the deal', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PiedmonteseTarotPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のディール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **精算の式まで出す。** 取り分と変動は席数倍ちがうので、並べただけでは
  // 計算が合わないように見える。
  it('breaks the settlement down at the end of a deal', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const box = await screen.findByTestId('piedmontesetarot-result');
    expect(box).toHaveTextContent('平均より上');
    const breakdown = screen.getByTestId('piedmontesetarot-breakdown');
    expect(breakdown).toHaveTextContent('78');
    expect(breakdown).toHaveTextContent('+12');
    expect(breakdown).toHaveTextContent('トリック 28点');
    expect(breakdown).toHaveTextContent('スカルト 2点');
    expect(screen.getByTestId('piedmontesetarot-formula')).toHaveTextContent('4');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({ hint: { cardIndices: [0], reason: 'lead_low' }, messageCode: '' }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    await screen.findByText('ディール 1');
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({
        hint: { cardIndices: [0], reason: 'lead_low' },
        messageCode: 'piedmontesetarot.hintRequested',
      }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **#6236 の本体（ピエモンテ側）。** Scarto 側と同じ配線レベルのガード。
  // 捨てられるピップがタロンの枚数に満たない手では、非オヌール切り札を
  // 選べなければならない。
  it('lets the dealer bury a trump when there are too few pips', async () => {
    mockExec.mockResolvedValue(shortOnPipsState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const cards = await screen.findAllByRole('button', { name: /♥|★|A/ });
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    const button = screen.getByRole('button', { name: /捨てる/ });
    expect(button).toBeEnabled();
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 1] }));
  });

  // オヌールとエクスキューズは、ピップが足りなくても選べない。
  it('still refuses the honours when pips run short', async () => {
    mockExec.mockResolvedValue(shortOnPipsState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const cards = await screen.findAllByRole('button', { name: /♥|★|A/ });
    fireEvent.click(cards[2]);
    fireEvent.click(cards[3]);
    expect(screen.getByRole('button', { name: /捨てる/ })).toBeDisabled();
  });

  // **取ったのが誰か画面から読めなかった (#6623)。** サーバは lastTrickWinner を
  // 最初から送っているのにページが読んでおらず、4 枚並んだ場札からタロッコの
  // 切り札優先を毎回自分で解き直すことになっていた。
  describe('trick winner highlight', () => {
    const trickOf = (values: number[]) =>
      values.map((value, i) => ({
        playerIdx: i,
        card: { design: 'HEART' as const, value, glyph: '♥', label: String(value), color: 'red', deck: 'tarot' },
      }));

    it('marks the winning seat once the trick is over', async () => {
      mockExec.mockResolvedValue(
        makePiedmonteseTarotState({
          phase: PiedmonteseTarotPhase.TRICK_END,
          currentTrick: trickOf([2, 3, 9, 4]),
          lastTrickWinner: 2,
        }),
      );
      const { container } = renderWithProviders(<PiedmonteseTarotPage />);
      await waitFor(() => expect(screen.getByTestId('trick-winner-badge')).toBeInTheDocument());
      // 光るのはちょうど1席。全席に付ける実装だとここで落ちる。
      expect(container.querySelectorAll('[data-trick-winner="true"]')).toHaveLength(1);
    });

    // **プレイ中は光らせない。** 最後の1枚で覆る答えを先に見せない。
    it('marks nothing while the trick is still being played', async () => {
      mockExec.mockResolvedValue(
        makePiedmonteseTarotState({
          phase: PiedmonteseTarotPhase.PLAY,
          currentTrick: trickOf([2, 3, 4]),
          lastTrickWinner: 2,
        }),
      );
      const { container } = renderWithProviders(<PiedmonteseTarotPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
      expect(container.querySelectorAll('[data-trick-winner="true"]')).toHaveLength(0);
    });

    // 受け入れ条件の「3人卓・4人卓どちらでも正しい席が光る」。席数と勝者の
    // 両方を変えて、固定の席を光らせているだけでないことを見る。
    it('marks the right seat on a three-handed table', async () => {
      mockExec.mockResolvedValue(
        makePiedmonteseTarotState({
          phase: PiedmonteseTarotPhase.TRICK_END,
          trickCount: 25,
          scartoCount: 3,
          config: { seats: 3, cpuDifficulty: 1, targetDeals: 4 },
          players: makePiedmonteseTarotState()
            .players.slice(0, 3)
            .map((p, i) => ({
              ...p,
              cardCount: 25,
              isDealer: i === 2,
            })),
          playerScores: [0, 0, 0],
          dealScores: [0, 0, 0],
          dealerIdx: 2,
          currentTrick: trickOf([9, 2, 3]),
          lastTrickWinner: 0,
        }),
      );
      const { container } = renderWithProviders(<PiedmonteseTarotPage />);
      await waitFor(() => expect(screen.getByTestId('trick-winner-badge')).toBeInTheDocument());
      const marked = container.querySelectorAll('[data-trick-winner="true"]');
      expect(marked).toHaveLength(1);
      // 席 0 は人間。名前ごと当たっていることを見る (席を取り違えても数は1のため)。
      expect(marked[0].textContent).toContain('あなた');
    });
  });

  // **催促は常設のライブ領域の中にある (#6880)。** フェーズ切り替えで現れる
  // テキストなので、領域が無いとスクリーンリーダには何も届かない。領域を
  // 出現と同時に付けても読み上げられないため、常設にして中身だけ差し替える。
  it('announces the prompt from an always-mounted live region', async () => {
    mockExec.mockResolvedValue(scartoState);
    renderWithProviders(<PiedmonteseTarotPage />);

    const live = await screen.findByTestId('piedmontesetarot-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 催促が**その領域の中**にあること。隣に置いただけの実装は属性の検査を通る。
    expect(live).toContainElement(await screen.findByTestId('piedmontesetarot-discard-prompt'));
  });

  describe('trick history', () => {
    it('renders collapsible trick history with completed tricks', async () => {
      const historyState = makePiedmonteseTarotState({
        completedTricks: [
          {
            trickNumber: 1,
            leadPlayerIdx: 0,
            winnerIdx: 2,
            cards: [
              { playerIdx: 0, card: suit(5, 'HEART', '♥', '5') },
              { playerIdx: 1, card: suit(8, 'HEART', '♥', '8') },
              {
                playerIdx: 2,
                card: { design: 'JOKER', value: 3, glyph: '✦', label: '3', color: 'purple', deck: 'tarot' },
              },
              { playerIdx: 3, card: suit(10, 'HEART', '♥', '10') },
            ],
          },
        ],
      });
      mockExec.mockResolvedValue(historyState);
      renderWithProviders(<PiedmonteseTarotPage />);

      const historySection = await screen.findByTestId('piedmontesetarot-trick-history');
      expect(historySection).toBeInTheDocument();
      expect(within(historySection).getByText('トリック履歴')).toBeInTheDocument();

      // トリック情報
      expect(within(historySection).getByText('トリック 1')).toBeInTheDocument();
      expect(within(historySection).getByText('リード: あなた')).toBeInTheDocument();
      expect(within(historySection).getByText('勝者: CPU 2')).toBeInTheDocument();

      // 各出札
      expect(within(historySection).getByTestId('piedmontesetarot-history-card-1-0')).toBeInTheDocument();
      expect(within(historySection).getByTestId('piedmontesetarot-history-card-1-1')).toBeInTheDocument();
      expect(within(historySection).getByTestId('piedmontesetarot-history-card-1-2')).toBeInTheDocument();
      expect(within(historySection).getByTestId('piedmontesetarot-history-card-1-3')).toBeInTheDocument();
    });

    it('shows empty message when no tricks are completed', async () => {
      mockExec.mockResolvedValue(makePiedmonteseTarotState({ completedTricks: [] }));
      renderWithProviders(<PiedmonteseTarotPage />);

      const historySection = await screen.findByTestId('piedmontesetarot-trick-history');
      expect(within(historySection).getByText('完了したトリックはありません')).toBeInTheDocument();
    });

    it('renders trick history in English', async () => {
      const historyState = makePiedmonteseTarotState({
        completedTricks: [
          {
            trickNumber: 1,
            leadPlayerIdx: 0,
            winnerIdx: 2,
            cards: [
              { playerIdx: 0, card: suit(5, 'HEART', '♥', '5') },
              { playerIdx: 1, card: suit(8, 'HEART', '♥', '8') },
              {
                playerIdx: 2,
                card: { design: 'JOKER', value: 3, glyph: '✦', label: '3', color: 'purple', deck: 'tarot' },
              },
              { playerIdx: 3, card: suit(10, 'HEART', '♥', '10') },
            ],
          },
        ],
      });
      await i18n.changeLanguage('en');
      try {
        mockExec.mockResolvedValue(historyState);
        renderWithProviders(<PiedmonteseTarotPage />);

        const historySection = await screen.findByTestId('piedmontesetarot-trick-history');
        expect(within(historySection).getByText('Trick history')).toBeInTheDocument();
        expect(within(historySection).getByText('Trick 1')).toBeInTheDocument();
        expect(within(historySection).getByText('Lead: You')).toBeInTheDocument();
        expect(within(historySection).getByText('Winner: CPU 2')).toBeInTheDocument();
      } finally {
        await i18n.changeLanguage('ja');
      }
    });
  });
});
