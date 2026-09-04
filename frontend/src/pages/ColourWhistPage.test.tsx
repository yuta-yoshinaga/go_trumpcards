import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { colourwhistApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, ColourWhistResponse } from '../types/card';
import { COLOUR_WHIST_NO_TRUMP } from '../types/games/colourwhist';
import { ColourWhistContract, ColourWhistPhase } from '../types/phases';
import { ColourWhistPage } from './ColourWhistPage';

vi.mock('../api/gameApi', () => ({
  colourwhistApi: { exec: vi.fn() },
  actionLogApi: { colourwhist: vi.fn() },
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

const mockApi = vi.mocked(colourwhistApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

// **文脈型が付かないので明示する。** 単体の配列だと design が string に推論されます。
const hand: Card[] = Array.from({ length: 13 }, (_, i) => ({ design: 'SPADE', value: i + 1 }));

const seat = (id: number, isHuman: boolean, cards: Card[] = []) => ({
  id,
  isHuman,
  cardCount: 13,
  cards,
  trickCount: 0,
  score: 0,
  isDeclarerSide: false,
  hasPassed: false,
});

const bidState: ColourWhistResponse = {
  players: [seat(0, true, hand), seat(1, false), seat(2, false), seat(3, false)],
  phase: ColourWhistPhase.BID,
  validPlays: [],
  dealerIdx: 0,
  contract: ColourWhistContract.NONE,
  declarerIdx: -1,
  partnerIdx: -1,
  calledCard: null,
  trumpSuit: COLOUR_WHIST_NO_TRUMP,
  troelForced: false,
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

const playState: ColourWhistResponse = {
  ...bidState,
  phase: ColourWhistPhase.PLAY,
  validPlays: [0, 3],
  contract: ColourWhistContract.SAMEN,
  declarerIdx: 0,
  trumpSuit: 3,
  currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 13 } }],
  declarerTricks: 4,
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

describe('ColourWhistPage', () => {
  it('resets on mount', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  // **トルールのボタンは無い。** 配りでしか成立しない契約です。
  it('offers only the three biddable contracts, plus pass', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /^サーメン/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /^アレーン/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^ミゼリー/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /トルール/ })).not.toBeInTheDocument();
  });

  it('sends the contract value on a bid', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /^アレーン/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /^アレーン/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', undefined, ColourWhistContract.ALLEEN));
  });

  it('sends pass as contract 0', async () => {
    mockApi.mockResolvedValue(bidState);
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '降りる' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '降りる' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bid', undefined, ColourWhistContract.NONE));
  });

  it('disables contracts at or below the current one', async () => {
    mockApi.mockResolvedValue({ ...bidState, contract: ColourWhistContract.SAMEN });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /^アレーン/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /^サーメン/ })).toBeDisabled();
    expect(screen.getByRole('button', { name: /^アレーン/ })).toBeEnabled();
  });

  // **競りが飛ばされた理由を出す。** 出さないと不具合に見えます。
  it('explains why the auction was skipped on a troel deal', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      troelForced: true,
      contract: ColourWhistContract.TROEL,
      partnerIdx: 2,
    });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-troel-notice')).toBeInTheDocument());
    expect(screen.getByTestId('colourwhist-troel-notice')).toHaveTextContent('エース3枚');
    expect(screen.getByTestId('colourwhist-contract')).toHaveTextContent('トルール');
    // **Troel の相方は最初から分かります。**
    expect(screen.getByTestId('colourwhist-partner')).toHaveTextContent('#2');
  });

  // **競りで決めた契約では出さない。** 負のコントロールです。
  it('shows no troel notice for a bid contract', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-contract')).toBeInTheDocument());
    expect(screen.queryByTestId('colourwhist-troel-notice')).not.toBeInTheDocument();
    // Samen の相方は指名札が出るまで伏せる。
    expect(screen.getByTestId('colourwhist-partner')).toHaveTextContent('未公開');
  });

  // **単独契約には相方がいない。** partnerIdx は「いない」も「いるが未公開」も
  // -1 なので、契約で判定しないと「相方: 非公開」と出て、隠れた味方がいると
  // 誤解させる (#5773)。
  it.each([
    ['alleen', ColourWhistContract.ALLEEN],
    ['miserie', ColourWhistContract.MISERIE],
  ])('shows no partner row for the solo contract %s', async (_name, contract) => {
    mockApi.mockResolvedValue({ ...playState, contract, partnerIdx: -1 });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-contract')).toBeInTheDocument());
    expect(screen.queryByTestId('colourwhist-partner')).not.toBeInTheDocument();
  });

  it('still shows the partner row once Troel reveals it', async () => {
    mockApi.mockResolvedValue({ ...playState, contract: ColourWhistContract.TROEL, partnerIdx: 2 });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-partner')).toHaveTextContent('#2'));
  });

  it('offers the four trumps in the call phase', async () => {
    mockApi.mockResolvedValue({
      ...bidState,
      phase: ColourWhistPhase.CALL,
      contract: ColourWhistContract.SAMEN,
      declarerIdx: 0,
    });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /ハート.*切り札/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /ハート.*切り札/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('call', undefined, undefined, 3));
  });

  it('enables only the legal cards', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-hand')).toBeInTheDocument());
    const buttons = screen.getByTestId('colourwhist-hand').querySelectorAll('button');
    expect(buttons).toHaveLength(13);
    buttons.forEach((btn, i) => {
      expect(btn.getAttribute('aria-disabled')).toBe(playState.validPlays.includes(i) ? 'false' : 'true');
    });
  });

  it('plays a legal card by its hand index', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.getByTestId('colourwhist-hand')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('colourwhist-hand').querySelectorAll('button')[3]);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', 3));
  });

  // **得点は負にもなる。** ゼロサムなので当然そうなります。
  it('renders negative scores', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      players: playState.players.map((p, i) => ({ ...p, score: i === 0 ? 6 : -2 })),
    });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-seats')).toBeInTheDocument());
    expect(screen.getByTestId('colourwhist-seats')).toHaveTextContent('-2');
  });

  it('offers the next round only at a round boundary', async () => {
    mockApi.mockResolvedValue(playState);
    const { unmount } = renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.queryByTestId('cw-next-button')).not.toBeInTheDocument());
    unmount();

    mockApi.mockResolvedValue({ ...playState, phase: ColourWhistPhase.ROUND_END, isHumanTurn: false });
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.getByTestId('cw-next-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('cw-next-button'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
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
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByTestId('colourwhist-hand')).not.toBeInTheDocument();
  });

  // **ギブアップは取り消せない** (#6475)。リセットには確認が挟まるのに、
  // ここは即座に対局を打ち切っていた。
  it('asks before giving up, and only then dispatches', async () => {
    renderWithProviders(<ColourWhistPage />);
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
    renderWithProviders(<ColourWhistPage />);
    const giveUp = await screen.findByTestId('giveup-button');

    mockApi.mockClear();
    fireEvent.click(giveUp);
    await waitFor(() => expect(screen.getByText('投了確認')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));

    await waitFor(() => expect(screen.queryByText('投了確認')).not.toBeInTheDocument());
    expect(mockApi).not.toHaveBeenCalled();
  });

  it('shows the called card when present', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      calledCard: { design: 'SPADE', value: 1 },
    });
    renderWithProviders(<ColourWhistPage />);

    const shown = await screen.findByTestId('colourwhist-called-card');
    expect(shown).toHaveTextContent('指名札:');
    // **どの札が指名されたかが要**。ラベルだけでは、札を描き忘れても通る。
    const img = shown.querySelector('img');
    expect(img).not.toBeNull();
    expect(img).toHaveAttribute('alt', expect.stringContaining('A'));
  });

  it('shows no called card when null', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      calledCard: null,
    });
    renderWithProviders(<ColourWhistPage />);

    await waitFor(() => expect(screen.getByTestId('colourwhist-contract')).toBeInTheDocument());
    expect(screen.queryByTestId('colourwhist-called-card')).not.toBeInTheDocument();
  });

  it('shows colourwhist-declarer when declarerIdx is 0 or greater', async () => {
    // 契約者が決まっている (0〜3) ときに宣言者とその獲得トリック数が表示される
    mockApi.mockResolvedValue({
      ...playState,
      declarerIdx: 0,
    });
    renderWithProviders(<ColourWhistPage />);
    expect(await screen.findByTestId('colourwhist-declarer')).toBeInTheDocument();
  });

  it('does not show colourwhist-declarer when declarerIdx is less than 0', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      declarerIdx: -1,
    });
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.getByTestId('colourwhist-contract')).toBeInTheDocument());
    expect(screen.queryByTestId('colourwhist-declarer')).not.toBeInTheDocument();
  });

  it('shows colourwhist-trick when currentTrick has cards', async () => {
    // 現在のトリックに1枚以上カードが出されているときに場札領域が表示される
    mockApi.mockResolvedValue({
      ...playState,
      currentTrick: [{ playerIdx: 0, card: { design: 'SPADE', value: 1 } }],
    });
    renderWithProviders(<ColourWhistPage />);
    expect(await screen.findByTestId('colourwhist-trick')).toBeInTheDocument();
  });

  it('does not show colourwhist-trick when currentTrick is empty', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      currentTrick: [],
    });
    renderWithProviders(<ColourWhistPage />);
    await waitFor(() => expect(screen.getByTestId('colourwhist-contract')).toBeInTheDocument());
    expect(screen.queryByTestId('colourwhist-trick')).not.toBeInTheDocument();
  });
});
