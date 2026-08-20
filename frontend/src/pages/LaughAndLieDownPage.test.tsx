import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { laughandliedownApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, LaughAndLieDownPlayer, LaughAndLieDownResponse } from '../types/card';
import { LaughAndLieDownPhase } from '../types/phases';
import { LaughAndLieDownPage } from './LaughAndLieDownPage';

vi.mock('../api/gameApi', () => ({
  laughandliedownApi: { exec: vi.fn() },
  actionLogApi: { laughandliedown: vi.fn() },
}));

const mockExec = vi.mocked(laughandliedownApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<LaughAndLieDownPlayer>): LaughAndLieDownPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 7), card('HEART', 9), card('CLOVER', 13)],
    wonCount: 4,
    laidDown: false,
    score: 0,
    hidden: false,
    ...overrides,
  };
}

function cpu(id: number, overrides?: Partial<LaughAndLieDownPlayer>): LaughAndLieDownPlayer {
  return {
    id,
    isHuman: false,
    cardCount: 3,
    cards: [],
    wonCount: 2,
    laidDown: false,
    score: 0,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<LaughAndLieDownResponse>): LaughAndLieDownResponse {
  return {
    players: [human(), cpu(1), cpu(2), cpu(3), cpu(4)],
    layout: [card('CLOVER', 7), card('DIAMOND', 7), card('SPADE', 3)],
    phase: 0,
    currentPlayerIdx: 0,
    validIndices: [0],
    threeTakeIndices: [],
    dealerIdx: 0,
    lastInIdx: -1,
    lastInBonus: 5,
    pot: 11,
    gameEndFlag: false,
    message: '',
    ...overrides,
  };
}

describe('LaughAndLieDownPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the pot, the dealer and both rules permanently', async () => {
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/ポット: 11/)).toBeInTheDocument();
    expect(screen.getByText(/1枚か3枚/)).toBeInTheDocument();
    expect(screen.getByText(/手札を全部場に置いて降りる/)).toBeInTheDocument();
  });

  it('shows every seat won count and marks who has laid down', async () => {
    // 取り札の枚数は 8 との差がそのまま収支なので、常に見えている必要がある。
    mockExec.mockResolvedValue(makeState({ players: [human(), cpu(1, { laidDown: true, wonCount: 9 })] }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/取り札9枚 · 降りた/)).toBeInTheDocument();
  });

  it('only plays the hand cards the server marked valid', async () => {
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();

    fireEvent.click(handButtons[1]);
    // Without the flush this cannot fail: nothing has had a chance to dispatch.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 1));
  });

  it('offers the three-card take only where the server said three are on the table', async () => {
    mockExec.mockResolvedValue(makeState({ validIndices: [0, 1], threeTakeIndices: [0] }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    // 2 枚の合法手のうち、3 枚取りが提示されるのは 1 枚だけ。
    expect(screen.getAllByRole('button', { name: '3枚取る' })).toHaveLength(1);
  });

  // **CUI は「1枚 or 3枚」を書いている。**Web はカードを光らせるだけで
  // takeCount を捨てており、さらにボタンを押す必要があることが伝わらなかった (#4884)。
  describe('three-take hint', () => {
    const armed = (takeCount: number) =>
      makeState({
        validIndices: [0],
        threeTakeIndices: [0],
        hint: { cardIndex: 0, takeCount, reason: 'x' },
      });

    it('marks the three-take button when the hint asks for three', async () => {
      mockExec.mockResolvedValue(armed(3));
      renderWithProviders(<LaughAndLieDownPage />);
      const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
      fireEvent.click(toggle);

      await waitFor(() => expect(document.querySelectorAll('[data-hint-take-three="true"]')).toHaveLength(1));
    });

    it('leaves it unmarked when the hint asks for one', async () => {
      mockExec.mockResolvedValue(armed(1));
      renderWithProviders(<LaughAndLieDownPage />);
      const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
      fireEvent.click(toggle);

      await waitFor(() => expect(screen.getByRole('button', { name: '3枚取る' })).toBeInTheDocument());
      expect(document.querySelectorAll('[data-hint-take-three="true"]')).toHaveLength(0);
    });

    it('leaves it unmarked while hints are off', async () => {
      mockExec.mockResolvedValue(armed(3));
      renderWithProviders(<LaughAndLieDownPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: '3枚取る' })).toBeInTheDocument());
      expect(document.querySelectorAll('[data-hint-take-three="true"]')).toHaveLength(0);
    });
  });

  it('sends a take count of three once the option is armed', async () => {
    mockExec.mockResolvedValue(makeState({ validIndices: [0], threeTakeIndices: [0] }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: '3枚取る' }));
    mockExec.mockClear();

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    fireEvent.click(handButtons[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0, 3));
  });

  it('reports each outcome by net result, not by finishing order', async () => {
    for (const [score, text] of [
      [3, '勝ち越しました'],
      [0, '収支ゼロでした'],
      [-2, '負け越しました'],
    ] as const) {
      const code = score > 0 ? 'laughandliedown.win' : score === 0 ? 'laughandliedown.even' : 'laughandliedown.lose';
      mockExec.mockResolvedValue(
        makeState({
          phase: 1,
          gameEndFlag: true,
          players: [human({ cards: [], cardCount: 0, score }), cpu(1)],
          messageCode: code,
        }),
      );
      renderWithProviders(<LaughAndLieDownPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // #5576: 訳文 (`lastIn`) もサーバのデータ (`lastInIdx`) も既にあったのに、
  // 画面が一度も読んでいなかった。最終点差の理由の一つが出ないまま終わっていた。
  describe('last-in bonus', () => {
    it('names the opponent who was last in, with the amount', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LaughAndLieDownPhase.GAME_END, lastInIdx: 1, lastInBonus: 5 }));
      renderWithProviders(<LaughAndLieDownPage />);
      const row = await screen.findByTestId('lld-lastin-1');
      expect(row).toHaveTextContent('5');
      // 他の席には出ない。全員に出すと誰が受け取ったか分からない。
      expect(screen.queryByTestId('lld-lastin-0')).not.toBeInTheDocument();
      expect(screen.queryByTestId('lld-lastin-2')).not.toBeInTheDocument();
    });

    it('names the human seat when it was last in', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LaughAndLieDownPhase.GAME_END, lastInIdx: 0, lastInBonus: 5 }));
      renderWithProviders(<LaughAndLieDownPage />);
      expect(await screen.findByTestId('lld-lastin-0')).toBeInTheDocument();
    });

    // -1 は「該当なし」。出すと、誰も受け取っていないボーナスを説明することになる。
    it('says nothing when nobody was last in', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LaughAndLieDownPhase.GAME_END, lastInIdx: -1 }));
      renderWithProviders(<LaughAndLieDownPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(document.querySelector('[data-testid^="lld-lastin-"]')).toBeNull();
    });

    // 決着前は出さない。まだ確定していない。
    it('says nothing before the game ends', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LaughAndLieDownPhase.PLAY, lastInIdx: 1, lastInBonus: 5 }));
      renderWithProviders(<LaughAndLieDownPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(document.querySelector('[data-testid^="lld-lastin-"]')).toBeNull();
    });

    // **額はサーバから。**訳文に数字を書くと、額を変えたとき片方だけ嘘になる。
    it('renders whatever amount the server sends', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LaughAndLieDownPhase.GAME_END, lastInIdx: 1, lastInBonus: 9 }));
      renderWithProviders(<LaughAndLieDownPage />);
      const row = await screen.findByTestId('lld-lastin-1');
      expect(row).toHaveTextContent('9');
      expect(row).not.toHaveTextContent('5');
    });
  });
});

// #5936: チュートリアルの精算ステップは「ポットから5を受け取る」と訳文に
// 数字を書いていた。ドメイン定数を変えると盤面だけが追随し、チュートリアルは
// 古い数字を教え続ける。**盤面と同じ値**を補間で渡す。
describe('LaughAndLieDownPage tutorial amount', () => {
  it('teaches the amount the server sent, not a number baked into the translation', async () => {
    mockExec.mockResolvedValue(makeState({ lastInBonus: 9 }));
    renderWithProviders(<LaughAndLieDownPage />);
    await waitFor(() => expect(document.querySelector('[data-tutorial="lld-table"]')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    // 精算は最後のステップ。手前の 3 つを送る。
    for (let i = 0; i < 3; i += 1) {
      fireEvent.click(await screen.findByRole('button', { name: '次へ' }));
    }

    const tooltip = await screen.findByText(/ポットから/);
    expect(tooltip).toHaveTextContent('9');
    expect(tooltip).not.toHaveTextContent('5を受け取ります');
  });
});
