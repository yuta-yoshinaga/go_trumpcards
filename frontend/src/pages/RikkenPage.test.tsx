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
});
