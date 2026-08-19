import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { botifarraApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BotifarraResponse, Card } from '../types/card';
import { BOTIFARRA_NO_TRUMP } from '../types/games/botifarra';
import { BotifarraPhase } from '../types/phases';
import { BotifarraPage } from './BotifarraPage';

vi.mock('../api/gameApi', () => ({
  botifarraApi: { exec: vi.fn() },
  actionLogApi: { botifarra: vi.fn() },
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

const mockApi = vi.mocked(botifarraApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

// **文脈型が付かないので明示する。** 単体の配列だと design が string に推論され、
// vitest は素通しするのに tsc だけが落ちます。
const hand: Card[] = Array.from({ length: 12 }, (_, i) => ({ design: 'SPADE', value: i + 1 }));

const declareState: BotifarraResponse = {
  players: [
    { id: 0, isHuman: true, team: 0, cardCount: 12, cards: hand, trickCount: 0 },
    { id: 1, isHuman: false, team: 1, cardCount: 12, cards: [], trickCount: 0 },
    { id: 2, isHuman: false, team: 0, cardCount: 12, cards: [], trickCount: 0 },
    { id: 3, isHuman: false, team: 1, cardCount: 12, cards: [], trickCount: 0 },
  ],
  phase: BotifarraPhase.DECLARE,
  validPlays: [],
  dealerIdx: 0,
  declarerIdx: -1,
  trumpSuit: BOTIFARRA_NO_TRUMP,
  multiplier: 1,
  currentTurn: 0,
  isHumanTurn: true,
  currentTrick: [],
  lastTrick: [],
  lastTrickWinner: -1,
  trickCount: 0,
  roundPoints: [0, 0],
  scores: [0, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  config: { targetScore: 101, allowDoubling: true },
  message: '',
};

const playState: BotifarraResponse = {
  ...declareState,
  phase: BotifarraPhase.PLAY,
  // **出せる札は手札の一部だけ。** 勝てるなら勝つ義務があるためです。
  validPlays: [2, 5],
  declarerIdx: 0,
  trumpSuit: 3,
  currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 12 } }],
  currentTurn: 0,
  roundPoints: [8, 5],
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

describe('BotifarraPage', () => {
  it('resets on mount', async () => {
    mockApi.mockResolvedValue(declareState);
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('offers every trump plus no trump, and delegation', async () => {
    mockApi.mockResolvedValue(declareState);
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /スペード を宣言/ })).toBeInTheDocument());
    for (const name of [/スペード を宣言/, /クラブ を宣言/, /ハート を宣言/, /ダイヤ を宣言/, /切り札なし を宣言/]) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument();
    }
    expect(screen.getByRole('button', { name: '相方に委ねる' })).toBeInTheDocument();
  });

  // **切り札なしは -1 という有効な値。** 送らないのとは違います。
  it('sends the suit value on declare, including -1 for no trump', async () => {
    mockApi.mockResolvedValue(declareState);
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ハート を宣言/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /ハート を宣言/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('declare', undefined, 3));

    fireEvent.click(screen.getByRole('button', { name: /切り札なし を宣言/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('declare', undefined, BOTIFARRA_NO_TRUMP));
  });

  it('delegates to the partner', async () => {
    mockApi.mockResolvedValue(declareState);
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '相方に委ねる' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '相方に委ねる' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('delegate'));
  });

  it('hides delegation once the partner is the one declaring', async () => {
    mockApi.mockResolvedValue({ ...declareState, phase: BotifarraPhase.DELEGATED });
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: /ハート を宣言/ })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '相方に委ねる' })).not.toBeInTheDocument();
  });

  it('offers doubling and passing', async () => {
    mockApi.mockResolvedValue({ ...declareState, phase: BotifarraPhase.DOUBLE, trumpSuit: 3, declarerIdx: 1 });
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByRole('button', { name: '倍付けする' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '倍付けする' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('double'));

    fireEvent.click(screen.getByRole('button', { name: 'そのまま進む' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('passdouble'));
  });

  // **出せない札は押せない。** 「勝てるなら勝つ義務」を UI 側でも守ります。
  it('enables only the legal cards', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByTestId('botifarra-hand')).toBeInTheDocument());
    const buttons = screen.getByTestId('botifarra-hand').querySelectorAll('button');
    expect(buttons).toHaveLength(12);
    buttons.forEach((btn, i) => {
      const legal = playState.validPlays.includes(i);
      expect(btn.getAttribute('aria-disabled')).toBe(legal ? 'false' : 'true');
      expect((btn as HTMLButtonElement).disabled).toBe(!legal);
    });
  });

  it('plays a legal card by its hand index', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.getByTestId('botifarra-hand')).toBeInTheDocument());

    const buttons = screen.getByTestId('botifarra-hand').querySelectorAll('button');
    fireEvent.click(buttons[5]);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('play', 5));
  });

  it('shows the trump, the stake and the running round points', async () => {
    mockApi.mockResolvedValue({ ...playState, multiplier: 2 });
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByTestId('botifarra-trump')).toBeInTheDocument());
    expect(screen.getByTestId('botifarra-trump')).toHaveTextContent('ハート');
    expect(screen.getByTestId('botifarra-trump')).toHaveTextContent('x2');
    // 合計 72 のうち何点取ったかが出る。
    expect(screen.getByTestId('botifarra-round-points')).toHaveTextContent('72');
  });

  // **ゼロサムで動く 2 つの数字** (#5771)。どちらが自分の組かをラベルで言う。
  it('labels which score is yours', async () => {
    mockApi.mockResolvedValue({ ...playState, scores: [3, 8], roundPoints: [40, 32] });
    renderWithProviders(<BotifarraPage />);

    const score = await screen.findByTestId('botifarra-score');
    expect(score).toHaveTextContent('あなた 3');
    expect(score).toHaveTextContent('相手 8');
    expect(screen.getByTestId('botifarra-round-points')).toHaveTextContent('あなた 40');
    expect(screen.getByTestId('botifarra-round-points')).toHaveTextContent('相手 32');
  });

  // **勝利演出も自分の組で決まる。** 添字 0 決め打ちだと、組 1 の席に座った
  // ときに勝ったのに祝われない (レビュー指摘)。
  it('celebrates only when the human team wins', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      gameEndFlag: true,
      winnerTeam: 1,
      players: [
        { id: 0, isHuman: false, team: 0, cardCount: 0, cards: [], trickCount: 0 },
        { id: 1, isHuman: true, team: 1, cardCount: 0, cards: [], trickCount: 0 },
        { id: 2, isHuman: false, team: 0, cardCount: 0, cards: [], trickCount: 0 },
        { id: 3, isHuman: false, team: 1, cardCount: 0, cards: [], trickCount: 0 },
      ],
    });
    const { unmount } = renderWithProviders(<BotifarraPage />);
    expect(await screen.findByTestId('win-celebration')).toBeInTheDocument();
    unmount();

    // 同じ勝ちチームでも、人間が組 0 なら負け。
    mockApi.mockResolvedValue({ ...playState, gameEndFlag: true, winnerTeam: 1 });
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.getByTestId('botifarra-score')).toBeInTheDocument());
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  // 席が読めない応答でも 0 で埋めて描画は続ける（数値の欠けで画面を落とさない）。
  it('falls back to zeroes when no seat is the human', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      scores: [],
      roundPoints: [],
      players: [{ id: 0, isHuman: false, team: 0, cardCount: 12, cards: [], trickCount: 0 }],
    });
    renderWithProviders(<BotifarraPage />);

    const score = await screen.findByTestId('botifarra-score');
    expect(score).toHaveTextContent('あなた 0');
    expect(score).toHaveTextContent('相手 0');
  });

  // **人間が組 1 の席のこともある。** 添字 0 決め打ちだと逆さまに出る。
  it('follows the human seat team, not index 0', async () => {
    mockApi.mockResolvedValue({
      ...playState,
      scores: [3, 8],
      players: [
        { id: 0, isHuman: false, team: 0, cardCount: 12, cards: [], trickCount: 0 },
        { id: 1, isHuman: true, team: 1, cardCount: 12, cards: hand, trickCount: 0 },
        { id: 2, isHuman: false, team: 0, cardCount: 12, cards: [], trickCount: 0 },
        { id: 3, isHuman: false, team: 1, cardCount: 12, cards: [], trickCount: 0 },
      ],
    });
    renderWithProviders(<BotifarraPage />);

    const score = await screen.findByTestId('botifarra-score');
    expect(score).toHaveTextContent('あなた 8');
    expect(score).toHaveTextContent('相手 3');
  });

  it('renders the current trick', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.getByTestId('botifarra-trick')).toBeInTheDocument());
    expect(screen.getByTestId('botifarra-trick')).toHaveTextContent('#1');
  });

  it('offers the next round only at a round boundary', async () => {
    mockApi.mockResolvedValue(playState);
    const { unmount } = renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.queryByTestId('bf-next-button')).not.toBeInTheDocument());
    unmount();

    mockApi.mockResolvedValue({ ...playState, phase: BotifarraPhase.ROUND_END, isHumanTurn: false });
    renderWithProviders(<BotifarraPage />);
    await waitFor(() => expect(screen.getByTestId('bf-next-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('bf-next-button'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('next'));
  });

  it('shows every seat with its team and counts', async () => {
    mockApi.mockResolvedValue(playState);
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByTestId('botifarra-seats')).toBeInTheDocument());
    const seats = screen.getByTestId('botifarra-seats');
    expect(seats).toHaveTextContent('あなた');
    expect(seats).toHaveTextContent('相方');
    expect(seats).toHaveTextContent('相手');
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
    mockApi.mockResolvedValue(declareState);
    renderWithProviders(<BotifarraPage />);

    await waitFor(() => expect(screen.getByRole('log')).toBeInTheDocument());
    expect(screen.queryByTestId('botifarra-hand')).not.toBeInTheDocument();
  });
});
