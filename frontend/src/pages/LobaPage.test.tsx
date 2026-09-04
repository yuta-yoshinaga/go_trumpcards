import { fireEvent, screen, waitFor } from '@testing-library/react';
import i18n from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { lobaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, LobaPlayer, LobaResponse } from '../types/card';
import { LobaPhase } from '../types/phases';
import { LobaPage } from './LobaPage';

vi.mock('../api/gameApi', () => ({
  lobaApi: { exec: vi.fn() },
  actionLogApi: { loba: vi.fn() },
}));

const mockExec = vi.mocked(lobaApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function seat(id: number, isHuman: boolean, overrides?: Partial<LobaPlayer>): LobaPlayer {
  return {
    id,
    isHuman,
    cardCount: 9,
    cards: isHuman ? [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 2)] : [],
    score: 12,
    eliminated: false,
    hasMelded: false,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<LobaResponse>): LobaResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: LobaPhase.ACT,
    currentPlayerIdx: 0,
    stockCount: 70,
    discardTop: card('HEART', 9),
    melds: [{ owner: 1, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] }],
    roundNo: 0,
    knockOut: 101,
    roundWinner: -1,
    roundClean: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

/** Select the given hand indices by clicking them. */
function selectCards(indices: number[]) {
  const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'discard');
  for (const i of indices) {
    fireEvent.click(handButtons[i]);
  }
}

describe('LobaPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // ヒントのトグルは localStorage に残る。消さないと次のテストが
    // チェック済みで始まり、クリックで off になる。
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently and the knock-out threshold', async () => {
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/異なる3スート/)).toBeInTheDocument();
    expect(screen.getByText(/ジョーカーは1枚まで、ピエルナ不可/)).toBeInTheDocument();
    expect(screen.getByText(/101点で脱落/)).toBeInTheDocument();
  });

  it('offers the two draws only in the draw step', async () => {
    mockExec.mockResolvedValue(makeState({ phase: LobaPhase.DRAW }));
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '捨て札を取る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('needs three selected cards before it will meld', async () => {
    // 2 枚では押せない。押せてしまうとサーバー往復が無駄になる。
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    selectCards([0, 1]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    selectCards([2]);
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, [0, 1, 2]));
  });

  it('lays off exactly one card onto a chosen meld', async () => {
    // 手札の添字と場の添字は別物なので、両方を選ばないと押せない。
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    selectCards([3]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '付ける' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(screen.getAllByTestId('loba-meld')[0]);
    fireEvent.click(screen.getByRole('button', { name: '付ける' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', 3, 0));
  });

  it('discards exactly one card', async () => {
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    selectCards([0, 1]);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    selectCards([1]); // deselect, leaving one
    fireEvent.click(screen.getByRole('button', { name: '捨てる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 0));
  });

  // ピエルナとエスカレラで付けられる札が違うので、種別が見えている必要がある。
  // 1 つの it で 2 度 render すると前の DOM が残り、[0] が古い方を指すので分ける。
  it('names a pierna on the table', async () => {
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getAllByTestId('loba-meld')[0]).toHaveTextContent('ピエルナ'));
  });

  it('names an escalera on the table', async () => {
    mockExec.mockResolvedValue(
      makeState({ melds: [{ owner: 0, kind: 1, cards: [card('SPADE', 5), card('SPADE', 6), card('SPADE', 7)] }] }),
    );
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getAllByTestId('loba-meld')[0]).toHaveTextContent('エスカレラ'));
  });

  it('names the owner of a meld via playerName for human player', async () => {
    mockExec.mockResolvedValue(
      makeState({ melds: [{ owner: 0, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] }] }),
    );
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getAllByTestId('loba-meld')[0]).toHaveTextContent('ピエルナ · あなた'));
    expect(screen.getAllByTestId('loba-meld')[0]).not.toHaveTextContent('席0');
  });

  it('names the owner of a meld via playerName for CPU player', async () => {
    mockExec.mockResolvedValue(
      makeState({ melds: [{ owner: 1, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] }] }),
    );
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getAllByTestId('loba-meld')[0]).toHaveTextContent('ピエルナ · CPU 1'));
    expect(screen.getAllByTestId('loba-meld')[0]).not.toHaveTextContent('席1');
  });

  it('renders meld owner labels through the shared playerName helper, not a literal', async () => {
    i18n.addResourceBundle('ja', 'common', { player: { cpu: 'コンピュータ{{id}}' } }, true, true);
    try {
      mockExec.mockResolvedValue(
        makeState({ melds: [{ owner: 2, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] }] }),
      );
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(screen.getAllByTestId('loba-meld')[0]).toHaveTextContent('ピエルナ · コンピュータ2'));
    } finally {
      i18n.addResourceBundle('ja', 'common', { player: { cpu: 'CPU {{id}}' } }, true, true);
    }
  });

  it('tells a clean go-out apart at the end of a round', async () => {
    mockExec.mockResolvedValue(makeState({ phase: LobaPhase.ROUND_END, roundWinner: 2, roundClean: true }));
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getByTestId('loba-round-result')).toHaveTextContent('-10'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のラウンドへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'loba.win', '最後まで残りました'],
      [2, 'loba.lose', '脱落しました'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: LobaPhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **ロバの核心の 2 アクション。**CUI は hintDraw / hintMeld で明示しているのに、
  // Web は捨て札 1 枚のリングしか使っていなかった (#4882)。
  describe('meld and draw hints', () => {
    const enableHints = async () => {
      const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
      fireEvent.click(toggle);
    };

    it('rings every card of the suggested meld', async () => {
      mockExec.mockResolvedValue(makeState({ hint: { cardIndices: [0, 2], drawStock: false, reason: 'x' } }));
      renderWithProviders(<LobaPage />);
      await enableHints();

      await waitFor(() => {
        const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'discard');
        expect(hand.filter((b) => b.className.includes('ring-ds-warning'))).toHaveLength(2);
      });
    });

    it('does not ring meld cards outside the act phase', async () => {
      mockExec.mockResolvedValue(
        makeState({ phase: LobaPhase.DRAW, hint: { cardIndices: [0, 2], drawStock: true, reason: 'x' } }),
      );
      renderWithProviders(<LobaPage />);
      await enableHints();

      await waitFor(() => expect(document.querySelectorAll('[data-hint-draw="true"]')).toHaveLength(1));
      const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'discard');
      expect(hand.filter((b) => b.className.includes('ring-ds-warning'))).toHaveLength(0);
    });

    it('rings the stock button when the hint says draw from stock', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LobaPhase.DRAW, hint: { drawStock: true, reason: 'x' } }));
      renderWithProviders(<LobaPage />);
      await enableHints();

      await waitFor(() => expect(document.querySelectorAll('[data-hint-draw="true"]')).toHaveLength(1));
      expect(screen.getByRole('button', { name: '山札から引く' })).toHaveAttribute('data-hint-draw', 'true');
    });

    it('rings the discard button when the hint says draw from the discard', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LobaPhase.DRAW, hint: { drawStock: false, reason: 'x' } }));
      renderWithProviders(<LobaPage />);
      await enableHints();

      await waitFor(() =>
        expect(screen.getByRole('button', { name: '捨て札を取る' })).toHaveAttribute('data-hint-draw', 'true'),
      );
      expect(screen.getByRole('button', { name: '山札から引く' })).not.toHaveAttribute('data-hint-draw');
    });

    it('rings nothing while hints are off', async () => {
      mockExec.mockResolvedValue(makeState({ phase: LobaPhase.DRAW, hint: { drawStock: true, reason: 'x' } }));
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
      expect(document.querySelectorAll('[data-hint-draw="true"]')).toHaveLength(0);
    });
  });

  // #5574: 「付ける」だけが二段階（手札1枚 + 付け先のメルド）なのに、その条件は
  // ボタンの disabled にしか無く、選んだのに押せない理由が画面のどこにも無かった。
  describe('LobaPage lay-off hint', () => {
    const hint = () => screen.queryByTestId('loba-layoff-hint');

    it('appears once a card is selected and no meld is', async () => {
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(hint()).not.toBeInTheDocument();

      selectCards([3]);
      await waitFor(() => expect(hint()).toBeInTheDocument());
    });

    // **足りているときは黙る。**出したままだと、押せる状態でも押せないように読める。
    it('goes away once a meld is chosen', async () => {
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      selectCards([3]);
      await waitFor(() => expect(hint()).toBeInTheDocument());
      fireEvent.click(screen.getAllByTestId('loba-meld')[0]);
      await waitFor(() => expect(hint()).not.toBeInTheDocument());
    });

    // 2 枚以上を選んでいる人はメルドを作ろうとしている。付ける話をしない。
    it('stays away while more than one card is selected', async () => {
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      selectCards([0, 1]);
      await flushPendingDispatch();
      expect(hint()).not.toBeInTheDocument();
    });

    // 付ける先が 1 つも無い局面で「メルドをクリックしてください」は嘘になる。
    it('stays away when there is no meld to lay off onto', async () => {
      mockExec.mockResolvedValue(makeState({ melds: [] }));
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      selectCards([3]);
      await flushPendingDispatch();
      expect(hint()).not.toBeInTheDocument();
    });

    // 常時ヒントは置き換えない。捨てるつもりの人を急かさないため (受け入れ条件2)。
    it('does not replace the standing hint', async () => {
      renderWithProviders(<LobaPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      selectCards([3]);
      await waitFor(() => expect(hint()).toBeInTheDocument());
      expect(screen.getByText(/手札をクリックして選び/)).toBeInTheDocument();
    });
  });
});

describe('LobaPage names the same opponent one way', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  // メルド欄と相手の手札欄が別々に名前を組んでいると、同じ相手が
  // 「CPU 1」と「CPU1」の二通りで同じ画面に出る (#6370)。
  it('uses the shared helper everywhere, so no CPU label is spelled two ways', async () => {
    mockExec.mockResolvedValue(
      makeState({ melds: [{ owner: 1, kind: 0, cards: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 5)] }] }),
    );
    renderWithProviders(<LobaPage />);
    await waitFor(() => expect(screen.getAllByTestId('loba-meld')[0]).toBeInTheDocument());

    const body = document.body.textContent ?? '';
    // 共有ヘルパの綴りは "CPU 1" (player.cpu = "CPU {{id}}")。
    expect(body).toContain('CPU 1');
    // 空白なしの綴りはどこにも残っていない。
    expect(body).not.toMatch(/CPU\d/);
  });
});
