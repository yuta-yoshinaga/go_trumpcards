import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { rikkenApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RikkenResponse } from '../types/card';
import { RIKKEN_NO_TRUMP } from '../types/games/rikken';
import { RikkenContract, RikkenPhase } from '../types/phases';
import { RikkenPage } from './RikkenPage';

vi.mock('../api/gameApi', () => ({
  rikkenApi: { exec: vi.fn() },
  actionLogApi: { rikken: vi.fn() },
}));

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockApi = vi.mocked(rikkenApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

// **文脈型が付かないので明示する。** 単体の配列だと design が string に推論されます。
const hand: Card[] = Array.from({ length: 13 }, (_, i) => ({ design: 'SPADE', value: i + 1 }));

const bidState: RikkenResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: hand,
      trickCount: 0,
      score: 0,
      isDeclarerSide: false,
      hasPassed: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      trickCount: 0,
      score: 0,
      isDeclarerSide: false,
      hasPassed: false,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 13,
      cards: [],
      trickCount: 0,
      score: 0,
      isDeclarerSide: false,
      hasPassed: false,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 13,
      cards: [],
      trickCount: 0,
      score: 0,
      isDeclarerSide: false,
      hasPassed: false,
    },
  ],
  phase: RikkenPhase.BID,
  validPlays: [],
  dealerIdx: 0,
  contract: RikkenContract.NONE,
  declarerIdx: -1,
  partnerIdx: -1,
  trumpSuit: RIKKEN_NO_TRUMP,
  currentTurn: 0,
  isHumanTurn: true,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  trickCount: 0,
  declarerTricks: 0,
  roundNumber: 1,
  gameEndFlag: false,
  winnerIdx: -1,
  config: { rounds: 8 },
  message: '',
};

const playState: RikkenResponse = {
  ...bidState,
  phase: RikkenPhase.PLAY,
  validPlays: [1, 4],
  contract: RikkenContract.RIK,
  declarerIdx: 0,
  trumpSuit: 3,
  currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 13 } }],
  declarerTricks: 3,
};

beforeEach(() => {
  vi.clearAllMocks();
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('RikkenPage', () => {
  it('resets on mount', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('offers all four contracts plus pass', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /^リク/ })).toBeInTheDocument());
    for (const name of [/^リク/, /^ミゼール/, /^ソロ/, /^オープンミゼール/]) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument();
    }
    expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument();
  });

  it('sends the contract value on a bid', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /^ソロ/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /^ソロ/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', undefined, RikkenContract.SOLO));
  });

  // **パスは契約 0。** 送らないのとは違います。
  it('sends pass as contract 0', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '降りる' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', undefined, RikkenContract.NONE));
  });

  // **競りは上へしか積めない。** 弱い契約のボタンは押せません。
  it('disables contracts at or below the current one', async () => {
    mockApi.mockResolvedValue({ ...bidState, contract: RikkenContract.MISERE });
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /^ソロ/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /^リク/ })).toBeDisabled();
    expect(screen.getByRole('button', { name: /^ミゼール/ })).toBeDisabled();
    expect(screen.getByRole('button', { name: /^ソロ/ })).toBeEnabled();
    expect(screen.getByRole('button', { name: /^オープンミゼール/ })).toBeEnabled();
  });

  it('offers the four trumps in the call phase', async () => {
    mockApi.mockResolvedValue({ ...bidState, phase: RikkenPhase.CALL, contract: RikkenContract.RIK, declarerIdx: 0 });
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /ハート.*を切り札/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /ハート.*を切り札/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('call', undefined, undefined, 3));
  });

  it('enables only the legal cards', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByTestId('rikken-hand')).toBeInTheDocument());
    const buttons = screen.getByTestId('rikken-hand').querySelectorAll('button');
    expect(buttons).toHaveLength(13);
    buttons.forEach((btn, i) => {
      const legal = playState.validPlays.includes(i);
      expect(btn.getAttribute('aria-disabled')).toBe(legal ? 'false' : 'true');
    });
  });

  it('plays a legal card by its hand index', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.getByTestId('rikken-hand')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('rikken-hand').querySelectorAll('button')[4]);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', 4));
  });

  it('shows the contract, trump and declarer', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByTestId('rikken-contract')).toBeInTheDocument());
    expect(screen.getByTestId('rikken-contract')).toHaveTextContent('リク');
    expect(screen.getByTestId('rikken-contract')).toHaveTextContent('ハート');
    expect(screen.getByTestId('rikken-declarer')).toHaveTextContent('#0');
  });

  // **相方は公開されるまで伏せる。**
  it('shows the partner as hidden until revealed', async () => {
    mockApi.mockResolvedValue(playState);
    const { unmount } = renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.getByTestId('rikken-partner')).toHaveTextContent('未公開'));
    unmount();

    mockApi.mockResolvedValue({ ...playState, partnerIdx: 2 });
    renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.getByTestId('rikken-partner')).toHaveTextContent('#2'));
  });

  // **得点は負にもなる。** ゼロサムなので当然そうなります。
  it('renders negative scores', async () => {
    const scored = {
      ...playState,
      players: playState.players.map((p, i) => ({ ...p, score: i === 0 ? 9 : -3 })),
    };
    mockApi.mockResolvedValue(scored);
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByTestId('rikken-seats')).toBeInTheDocument());
    expect(screen.getByTestId('rikken-seats')).toHaveTextContent('-3');
  });

  it('offers the next round only at a round boundary', async () => {
    mockApi.mockResolvedValue(playState);
    const { unmount } = renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.queryByTestId('rk-next-button')).not.toBeInTheDocument());
    unmount();

    mockApi.mockResolvedValue({ ...playState, phase: RikkenPhase.ROUND_END, isHumanTurn: false });
    renderWithProviders(<RikkenPage />);
    await waitFor(() => expect(screen.getByTestId('rk-next-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('rk-next-button'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  // **オープンミゼールは宣言者の手札を全員に公開する契約。**サーバは CPU 宣言者の
  // cards も送っているのに、Web だけ描画していなかった。
  it('shows the open-misere hand of a CPU declarer', async () => {
    const openHand: Card[] = [
      { design: 'HEART', value: 2 },
      { design: 'CLOVER', value: 7 },
      { design: 'DIAMOND', value: 11 },
    ];
    mockApi.mockResolvedValue({
      ...playState,
      contract: RikkenContract.OPEN_MISERE,
      declarerIdx: 1,
      players: playState.players.map((p) => (p.id === 1 ? { ...p, cards: openHand } : p)),
    });
    renderWithProviders(<RikkenPage />);

    const shown = await screen.findByTestId('rikken-open-misere-hand');
    // CUI の writeOpenMisereHand と同じ情報量＝全カード。
    expect(shown.querySelectorAll('img')).toHaveLength(openHand.length);
    expect(screen.getByText(/公開手札（席1）/)).toBeInTheDocument();
  });

  // 契約が違えば公開しない。**CPU 側に札を持たせた盤で見る** — 既定の
  // フィクスチャは CPU の cards が空なので、契約の判定を外しても通ってしまう。
  it('shows no open hand under an ordinary contract, even with the cards present', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      contract: RikkenContract.RIK,
      declarerIdx: 1,
      players: playState.players.map((p) =>
        p.id === 1 ? { ...p, cards: [{ design: 'HEART', value: 2 } as Card] } : p,
      ),
    });
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByTestId('rikken-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('rikken-open-misere-hand')).not.toBeInTheDocument();
  });

  // 宣言者が人間なら自分の手札欄がその役目を果たす。二重に出さない。
  it('shows no separate open hand when the human declared it', async () => {
    mockApi.mockResolvedValue({ ...playState, contract: RikkenContract.OPEN_MISERE, declarerIdx: 0 });
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByTestId('rikken-hand')).toBeInTheDocument());
    expect(screen.queryByTestId('rikken-open-misere-hand')).not.toBeInTheDocument();
  });

  it('renders the CLI terminal when CLI mode is on', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<RikkenPage />);

    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByTestId('rikken-hand')).not.toBeInTheDocument();
  });

  // **ギブアップは取り消せない** (#6475)。リセットには確認が挟まるのに、
  // ここは即座に対局を打ち切っていた。
  it('asks before giving up, and only then dispatches', async () => {
    renderWithProviders(<RikkenPage />);
    const giveUp = await screen.findByTestId('giveup-button');

    mockApi.mockClear();
    fireEvent.click(giveUp);
    await waitFor(() => expect(screen.getByText('投了確認')).toBeInTheDocument());
    expect(mockApi).not.toHaveBeenCalledWith('giveup');

    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('giveup'));
  });

  // キャンセルしたら何も起きない ── ダイアログを出すだけで通す実装を落とす。
  it('leaves the game untouched when the give-up dialog is cancelled', async () => {
    renderWithProviders(<RikkenPage />);
    const giveUp = await screen.findByTestId('giveup-button');

    mockApi.mockClear();
    fireEvent.click(giveUp);
    await waitFor(() => expect(screen.getByText('投了確認')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));

    await waitFor(() => expect(screen.queryByText('投了確認')).not.toBeInTheDocument());
    expect(mockApi).not.toHaveBeenCalled();
  });
});
