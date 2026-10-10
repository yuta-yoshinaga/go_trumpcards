import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trogguApi } from '../api/gameApi';
import i18n from '../i18n';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTrogguState } from '../test/stateFactories';
import { TrogguPage } from './TrogguPage';

vi.mock('../api/gameApi', () => ({
  trogguApi: { exec: vi.fn() },
  actionLogApi: { troggu: vi.fn() },
}));

const mockExec = vi.mocked(trogguApi.exec);

/** The hand renders as buttons carrying `aria-pressed` (no test id). */
const handButtons = () => screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));

const bidState = makeTrogguState();
const playState = makeTrogguState({
  phase: 1,
  trickNumber: 1,
  declarerIdx: 0,
  contract: 2,
  contractName: 'solo',
  players: bidState.players.map((p, i) => (i === 0 ? { ...p, isDeclarer: true } : p)),
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidState);
});

describe('TrogguPage', () => {
  it('keeps the tutorial action target present outside the human bidding turn', async () => {
    mockExec.mockResolvedValue(playState);
    const { container } = renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-info')).toBeInTheDocument();
    expect(container.querySelector('[data-tutorial="tg-actions"]')).toBeInTheDocument();
  });

  it('marks ordinary trumps visually and in accessible names, but not the Excuse', async () => {
    mockExec.mockResolvedValue(
      makeTrogguState({
        players: makeTrogguState().players.map((player) =>
          player.isHuman
            ? {
                ...player,
                cards: [
                  { design: 'JOKER', value: 21, glyph: '✦', label: '21', color: 'purple', deck: 'tarot' },
                  { design: 'JOKER', value: 0, glyph: '★', label: 'Excuse', color: 'gold', deck: 'tarot' },
                ],
              }
            : player,
        ),
      }),
    );
    renderWithProviders(<TrogguPage />);
    await screen.findAllByRole('button');
    const hand = handButtons();
    const trump = hand.find((button) => button.getAttribute('data-trump') === 'true');
    const excuse = hand.find((button) => button.getAttribute('aria-label')?.includes('Excuse'));
    expect(trump).toHaveAttribute('data-trump', 'true');
    expect(trump?.getAttribute('aria-label')).toContain('切り札');
    expect(excuse).not.toHaveAttribute('data-trump');
    expect(excuse?.getAttribute('aria-label')).not.toContain('切り札');
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<TrogguPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the deal, trick and contract', async () => {
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-info')).toHaveTextContent('ディール 1/4');
  });

  it('shows the current auction contract and bidder, including the no-bids state', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ highestBid: 3, highestBidder: 2 }));
    const { unmount } = renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-auction-highest')).toHaveTextContent(
      '最高入札: ピッコロ（ちょうど1トリック）',
    );
    expect(screen.getByTestId('tg-auction-highest')).toHaveTextContent('入札者: CPU2');
    try {
      await i18n.changeLanguage('en');
      expect(screen.getByTestId('tg-auction-highest')).toHaveTextContent(
        'Highest bid: Piccolo (exactly 1 trick) Bidder: CPU2',
      );
    } finally {
      await i18n.changeLanguage('ja');
    }
    unmount();

    mockExec.mockResolvedValue(makeTrogguState({ highestBid: 0, highestBidder: -1 }));
    const { unmount: unmountNoBid } = renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-auction-highest')).toHaveTextContent('未入札');
    unmountNoBid();
  });

  it('shows the settled contract after the auction phase', async () => {
    mockExec.mockResolvedValue(
      makeTrogguState({
        phase: 1,
        highestBid: 3,
        highestBidder: 2,
        declarerIdx: 2,
        contractName: 'piccolo',
      }),
    );
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-info')).toHaveTextContent('契約: ピッコロ（ちょうど1トリック）');
    expect(screen.queryByTestId('tg-auction-highest')).not.toBeInTheDocument();
  });

  it('keeps the previous trick and winner visible while the next trick is in progress', async () => {
    const previousCard = { design: 'HEART', value: 3, glyph: '♥', label: '3', color: 'red', deck: 'tarot' } as const;
    const currentCard = { design: 'SPADE', value: 14, glyph: '♠', label: 'K', color: 'black', deck: 'tarot' } as const;
    mockExec.mockResolvedValue(
      makeTrogguState({
        phase: 1,
        trickNumber: 2,
        declarerIdx: 0,
        contract: 2,
        contractName: 'solo',
        currentTrick: [{ playerIdx: 0, card: currentCard }],
        lastTrickCards: [previousCard],
        lastTrickWinner: 2,
      }),
    );

    renderWithProviders(<TrogguPage />);

    const previous = await screen.findByTestId('tg-previous-trick');
    expect(previous).toHaveTextContent('直前のトリック');
    expect(previous).toHaveTextContent('勝者: CPU2');
    expect(previous.querySelector('[role="img"]')).toHaveAttribute('aria-label', '3 ♥');
    expect(screen.getByText('現在のトリック')).toBeInTheDocument();
  });

  // **4 契約すべてが打てる。** どれか一つ欠けると、その契約だけが遊べなくなる。
  it('offers all four contracts and pass', async () => {
    renderWithProviders(<TrogguPage />);
    const actions = await screen.findByTestId('game-footer-actions');
    for (const c of ['trois', 'solo', 'piccolo', 'misere']) {
      expect(await screen.findByTestId(`tg-bid-${c}`)).toBeInTheDocument();
      expect(within(actions).getByTestId(`tg-bid-${c}`)).toBeInTheDocument();
    }
    expect(screen.getByTestId('tg-pass')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('tg-bid-misere'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'misere' }));
  });

  // **今の最高入札を超えられない契約は押せない。**押せてしまうとサーバーに
  // 却下されるだけで、画面は入札のまま何も変わらない (#5808)。トロワ(1) <
  // ソロ(2) < ピッコロ(3) < ミゼール(4) で、`bid <= highestBid` は却下される。
  it('keeps losing bids focusable and explains why they cannot be selected', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ highestBid: 2 }));
    renderWithProviders(<TrogguPage />);

    // ソロ(2) 以下はフォーカス可能なまま、aria-disabled で却下対象を示す。
    const trois = await screen.findByTestId('tg-bid-trois');
    const solo = screen.getByTestId('tg-bid-solo');
    expect(trois).not.toBeDisabled();
    expect(trois).toHaveAttribute('aria-disabled', 'true');
    expect(solo).not.toBeDisabled();
    expect(solo).toHaveAttribute('aria-disabled', 'true');

    const describedBy = solo.getAttribute('aria-describedby');
    expect(describedBy).toBeTruthy();
    const reason = document.getElementById(describedBy ?? '');
    expect(reason).toBeInTheDocument();
    expect(reason).toHaveTextContent('この契約では今の最高入札（ソロ（92点以上））を超えられません');

    // 負の対照: aria-disabled はクリックを自動抑止しないので、明示的に門番する。
    mockExec.mockClear();
    fireEvent.click(solo);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    // 上回る契約は押せる。
    const piccolo = screen.getByTestId('tg-bid-piccolo');
    const misere = screen.getByTestId('tg-bid-misere');
    expect(piccolo).toBeEnabled();
    expect(piccolo).not.toHaveAttribute('aria-disabled');
    expect(piccolo).not.toHaveAttribute('aria-describedby');
    expect(misere).toBeEnabled();
    expect(misere).not.toHaveAttribute('aria-disabled');
    expect(misere).not.toHaveAttribute('aria-describedby');
    fireEvent.click(piccolo);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'piccolo' }));

    // パスはいつでも押せる。
    expect(screen.getByTestId('tg-pass')).toBeEnabled();
    // 理由が読める (押せない理由が画面に無いと、ただ壊れて見える)。
    expect(solo.getAttribute('title') ?? '').not.toBe('');
  });

  // 負のコントロール: 誰も入札していなければ 4 契約すべて押せる。
  it('leaves every contract available while no one has bid', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ highestBid: 0 }));
    renderWithProviders(<TrogguPage />);
    for (const c of ['trois', 'solo', 'piccolo', 'misere']) {
      const button = await screen.findByTestId(`tg-bid-${c}`);
      expect(button).toBeEnabled();
      expect(button).not.toHaveAttribute('aria-disabled');
      expect(button).not.toHaveAttribute('aria-describedby');
    }
  });

  it('sends pass', async () => {
    renderWithProviders(<TrogguPage />);
    fireEvent.click(await screen.findByTestId('tg-pass'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **キーボードから届かない操作を残さない** (#5787)。
  it('パスと「次へ」をキーからも打てる', async () => {
    const { unmount } = renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-pass');

    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'p' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
    unmount();

    // トリック終了では next、ディール終了では nextround。
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 2, lastTrickWinner: 1 }));
    const trick = renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-next-trick');
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    trick.unmount();

    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 3 }));
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-next-round');
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **押せない場面ではキーも効かない。** ボタンの表示条件と同じ値で門番する。
  it('プレイ中はパスのキーが効かない', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-info');

    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'p' });
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(screen.getByTestId('tg-info')).toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });

  // キー一覧がフッターに出る（受け入れ条件4）。
  it('キーの一覧をフッターに出す', async () => {
    renderWithProviders(<TrogguPage />);
    const panel = await screen.findByTestId('tg-kbd-shortcuts');
    // 既定は閉じたまま（畳んでいる間は行そのものが mount されない）。
    expect(panel).not.toHaveAttribute('open');
    fireEvent.click(screen.getByText('キーボードショートカット'));
    expect(screen.getByText('パスする')).toBeInTheDocument();
  });

  it('plays a card immediately during the play phase', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-info');
    mockExec.mockClear();
    fireEvent.click(handButtons()[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('advances the trick', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 2, lastTrickWinner: 1 }));
    renderWithProviders(<TrogguPage />);
    fireEvent.click(await screen.findByTestId('tg-next-trick'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **ソロは点数で語る。** 単位を取り違えると読めない結果になる。
  it('reports a Solo deal in card points', async () => {
    mockExec.mockResolvedValue(
      makeTrogguState({
        ...playState,
        phase: 3,
        breakdown: {
          contract: 2,
          contractName: 'solo',
          declarerPoints: 60,
          declarerTricks: 9,
          target: 46,
          targetIsTricks: false,
          won: true,
          base: 20,
          seats: [60, -20, -20, -20],
        },
      }),
    );
    renderWithProviders(<TrogguPage />);
    const result = await screen.findByTestId('tg-round-result');
    expect(result).toHaveTextContent('宣言者: あなた');
    expect(result).toHaveTextContent('60点');
    expect(result).not.toHaveTextContent('トリック');
    expect(screen.getByTestId('tg-round-seat-0')).toHaveTextContent('60');
  });

  // **他の契約はトリック数で語る。**
  it('reports a Misere deal in tricks', async () => {
    mockExec.mockResolvedValue(
      makeTrogguState({
        ...playState,
        phase: 3,
        contractName: 'misere',
        breakdown: {
          contract: 4,
          contractName: 'misere',
          declarerPoints: 12,
          declarerTricks: 1,
          target: 0,
          targetIsTricks: true,
          won: false,
          base: 40,
          seats: [-120, 40, 40, 40],
        },
      }),
    );
    renderWithProviders(<TrogguPage />);
    const result = await screen.findByTestId('tg-round-result');
    expect(result).toHaveTextContent('トリック');
    expect(result).toHaveTextContent('失敗');
    expect(result).not.toHaveTextContent('12点');
  });

  // 流局は精算そのものが無い。
  it('reports a thrown-in deal', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ phase: 3, breakdown: null }));
    renderWithProviders(<TrogguPage />);
    const result = await screen.findByTestId('tg-round-result');
    expect(result).toHaveTextContent('流局');
    expect(result).not.toHaveTextContent('宣言者');
  });

  it('shows the final scores and restarts with the chosen settings', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 4, gameEndFlag: true, winnerPlayer: 0 }));
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-result')).toHaveTextContent('勝者: あなた');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('reports a tied match', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 4, gameEndFlag: true, winnerPlayer: -1 }));
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-result')).toHaveTextContent('引き分け');
  });

  it('surfaces an API error raised after the board is up', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-info');

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(handButtons()[0]);
    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });

  it('displays the solo target during play without unresolved placeholders', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    const info = await screen.findByTestId('tg-info');
    expect(info).toHaveTextContent('92');
    expect(info.textContent).not.toContain('{{');
    expect(info.textContent).not.toContain('}}');
  });

  it('displays the solo target in the bid buttons without unresolved placeholders', async () => {
    mockExec.mockResolvedValue(bidState);
    renderWithProviders(<TrogguPage />);
    const soloBtn = await screen.findByTestId('tg-bid-solo');
    expect(soloBtn).toHaveTextContent('92');
    expect(soloBtn.textContent).not.toContain('{{');
    expect(soloBtn.textContent).not.toContain('}}');
  });
});
