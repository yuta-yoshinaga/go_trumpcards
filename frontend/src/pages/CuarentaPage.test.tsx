import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cuarentaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CuarentaPlayer, CuarentaResponse } from '../types/card';
import { CuarentaPage } from './CuarentaPage';

vi.mock('../api/gameApi', () => ({
  cuarentaApi: { exec: vi.fn() },
  actionLogApi: { cuarenta: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
/**
 * Counts calls for one sound name. The central taps (useGameApi / GamePageShell)
 * play through this same mocked context, so aggregate assertions on
 * mockPlaySound would also count deal/card sounds this page does not own.
 */
const soundCalls = (name: string) => mockPlaySound.mock.calls.filter((c) => c[0] === name).length;

vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(cuarentaApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makePlayer(overrides: Partial<CuarentaPlayer> = {}): CuarentaPlayer {
  return {
    id: 1,
    team: 1,
    isHuman: false,
    cardCount: 5,
    cards: [],
    capturedCount: 0,
    ...overrides,
  };
}

function makeState(overrides: Partial<CuarentaResponse> = {}): CuarentaResponse {
  return {
    players: [
      makePlayer({
        id: 0,
        team: 0,
        isHuman: true,
        cards: [card('SPADE', 5), card('HEART', 11), card('DIAMOND', 1), card('CLOVER', 7), card('SPADE', 2)],
      }),
      makePlayer({ id: 1, team: 1 }),
      makePlayer({ id: 2, team: 0 }),
      makePlayer({ id: 3, team: 1 }),
    ],
    currentTurn: 0,
    tableCards: [card('CLOVER', 7), card('HEART', 3)],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 0,
    teamScores: [12, 8],
    remainingDeck: 16,
    roundWinners: [],
    cpuActions: [],
    humanAction: null,
    lastRoundDetail: null,
    config: { targetScore: 40, cpuDifficulty: 1 },
    message: '',
    ...overrides,
  };
}

const playState = makeState();
const emptyTableState = makeState({ tableCards: [] });
const roundEndState = makeState({ phase: 1, currentTurn: -1 });
const gameEndState = makeState({
  phase: 2,
  gameEndFlag: true,
  currentTurn: -1,
  roundWinners: [0],
  teamScores: [40, 31],
});
const cpuWinState = makeState({
  phase: 2,
  gameEndFlag: true,
  currentTurn: -1,
  roundWinners: [1],
  teamScores: [29, 40],
});

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockPlaySound.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('CuarentaPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CuarentaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the remaining stock', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText(/山札: 16枚/)).toBeInTheDocument());
  });

  it('renders team scores with the target', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText(/チームA: 12 \/ 40点/)).toBeInTheDocument());
    expect(screen.getByText(/チームB: 8 \/ 40点/)).toBeInTheDocument();
    const announce = screen.getByTestId('cuarenta-team-update-announce');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toBeEmptyDOMElement();
  });

  it('announces increased team scores and captured totals, but not reset decreases', async () => {
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-0');
    mockExec.mockResolvedValueOnce(
      makeState({
        teamScores: [14, 8],
        players: [
          makePlayer({ id: 0, team: 0, isHuman: true, capturedCount: 2 }),
          makePlayer({ id: 1, team: 1 }),
          makePlayer({ id: 2, team: 0, capturedCount: 1 }),
          makePlayer({ id: 3, team: 1 }),
        ],
      }),
    );
    fireEvent.click(cardBtn);
    await waitFor(() =>
      expect(screen.getByTestId('cuarenta-team-update-announce')).toHaveTextContent(
        'チームAの得点は14点、捕獲枚数は3枚',
      ),
    );
    expect(screen.getByTestId('cuarenta-team-update-announce')).toHaveAttribute('role', 'status');
  });

  it('renders the players list with captured counts', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    expect(screen.getAllByText(/捕獲 0枚/).length).toBe(4);
  });

  it('sums each team captured-card total from its two players', async () => {
    // Team A = seats {0,2}: 8 + 5 = 13; Team B = seats {1,3}: 6 + 1 = 7.
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({ id: 0, team: 0, isHuman: true, capturedCount: 8 }),
          makePlayer({ id: 1, team: 1, capturedCount: 6 }),
          makePlayer({ id: 2, team: 0, capturedCount: 5 }),
          makePlayer({ id: 3, team: 1, capturedCount: 1 }),
        ],
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByTestId('cuarenta-team-captured-0')).toHaveTextContent('獲得 13枚'));
    expect(screen.getByTestId('cuarenta-team-captured-1')).toHaveTextContent('獲得 7枚');
  });

  it('highlights a team counter as it approaches the 20-card bonus', async () => {
    // Team A at 19 (approaching) is emphasized; Team B at 7 is not.
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({ id: 0, team: 0, isHuman: true, capturedCount: 10 }),
          makePlayer({ id: 1, team: 1, capturedCount: 4 }),
          makePlayer({ id: 2, team: 0, capturedCount: 9 }),
          makePlayer({ id: 3, team: 1, capturedCount: 3 }),
        ],
      }),
    );
    renderWithProviders(<CuarentaPage />);
    const teamA = await screen.findByTestId('cuarenta-team-captured-0');
    expect(teamA).toHaveTextContent('獲得 19枚');
    expect(teamA.className).toContain('text-ds-accent');
    const teamB = screen.getByTestId('cuarenta-team-captured-1');
    expect(teamB.className).not.toContain('text-ds-accent');
  });

  it('resets team captured totals to zero on a fresh round', async () => {
    // Default fixture has all capturedCount 0 — both team counters read 0.
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByTestId('cuarenta-team-captured-0')).toHaveTextContent('獲得 0枚'));
    expect(screen.getByTestId('cuarenta-team-captured-1')).toHaveTextContent('獲得 0枚');
  });

  it('renders the human hand cards', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByTestId('hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('hand-card-4')).toBeInTheDocument();
  });

  it('names each hand card in its aria-label', async () => {
    renderWithProviders(<CuarentaPage />);
    // hand[0] is ♠5, hand[2] is ♦A.
    expect(await screen.findByRole('button', { name: '♠ 5 を出す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ A を出す' })).toBeInTheDocument();
  });

  it('plays a hand card on the human turn', async () => {
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
  });

  it('rings capturable table cards while a hand card is focused, then clears on blur', async () => {
    // hand[3] is ♣7; table[0] is ♣7 (equal rank → capturable), table[1] is ♥3.
    renderWithProviders(<CuarentaPage />);
    const handCard = await screen.findByTestId('hand-card-3');
    // No preview yet: nothing is ringed.
    expect(screen.getByTestId('cuarenta-table-card-0')).not.toHaveAttribute('data-capturable');

    fireEvent.focus(handCard);
    await waitFor(() => expect(screen.getByTestId('cuarenta-table-card-0')).toHaveAttribute('data-capturable', 'true'));
    expect(screen.getByTestId('cuarenta-capture-preview')).toHaveTextContent('この札で捕獲できます: ♣ 7');
    expect(screen.getByTestId('cuarenta-capture-preview')).toHaveAttribute('aria-live', 'polite');
    expect(screen.getByTestId('cuarenta-table-card-0').className).toContain('ring-ds-success');
    // The non-matching table card stays un-ringed.
    expect(screen.getByTestId('cuarenta-table-card-1')).not.toHaveAttribute('data-capturable');

    fireEvent.blur(handCard);
    await waitFor(() => expect(screen.getByTestId('cuarenta-table-card-0')).not.toHaveAttribute('data-capturable'));
    expect(screen.getByTestId('cuarenta-capture-preview')).toBeEmptyDOMElement();
  });

  it('explains that a hand card with no matches will be laid on the table', async () => {
    renderWithProviders(<CuarentaPage />);
    const handCard = await screen.findByTestId('hand-card-0');
    fireEvent.focus(handCard);
    expect(screen.getByTestId('cuarenta-capture-preview')).toHaveTextContent(
      '捕獲できる札はありません。場に出します。',
    );
  });

  it('does not ring any table card when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2 }));
    renderWithProviders(<CuarentaPage />);
    const handCard = await screen.findByTestId('hand-card-3');
    fireEvent.focus(handCard);
    // Hover preview is suppressed off-turn — no capturable ring appears.
    expect(screen.getByTestId('cuarenta-table-card-0')).not.toHaveAttribute('data-capturable');
  });

  it('shows the empty-table label when the table is empty', async () => {
    mockExec.mockResolvedValue(emptyTableState);
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('場札なし')).toBeInTheDocument());
  });

  it('does not dispatch play when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentTurn: 2 }));
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(cardBtn);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders capture-result badges for the human action', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7), card('HEART', 7), card('SPADE', 7)],
          isCaida: true,
          isLimpia: true,
          rondaBonus: 1,
        },
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('カイーダ! +2')).toBeInTheDocument());
    expect(screen.getByText('ロンダ! +1')).toBeInTheDocument();
    expect(screen.getByText('リンピア! +1')).toBeInTheDocument();
    // The human's bonus badges pop (motion-safe) to draw the eye.
    const popped = screen.getAllByTestId('cuarenta-bonus-pop');
    expect(popped.length).toBeGreaterThan(0);
    expect(popped[0].className).toContain('motion-safe:animate-bounce');
    // The permanent live region announces the server-reported caída and matches its badge.
    const announce = screen.getByTestId('cuarenta-bonus-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce.textContent).toBe('カイーダ! +2ロンダ! +1リンピア! +1');
    const result = screen.getByTestId('cuarenta-action-0');
    expect(within(result).getByRole('img', { name: '♥ 7' })).toBeInTheDocument();
    expect(within(result).getByRole('img', { name: '♠ 7' })).toBeInTheDocument();
  });

  it('shows captured cards for CPU plays and none for a card laid on the table', async () => {
    mockExec.mockResolvedValue(
      makeState({
        cpuActions: [
          {
            playerIdx: 2,
            playedCard: card('HEART', 3),
            capturedCards: [card('HEART', 3)],
            isCaida: false,
            isLimpia: false,
            rondaBonus: 0,
          },
          {
            playerIdx: 1,
            playedCard: card('SPADE', 9),
            capturedCards: [],
            isCaida: false,
            isLimpia: false,
            rondaBonus: 0,
          },
        ],
      }),
    );
    renderWithProviders(<CuarentaPage />);
    const capture = await screen.findByTestId('cuarenta-action-2');
    expect(capture).toHaveTextContent('捕獲 1枚');
    expect(within(capture).getAllByRole('img', { name: '♥ 3' })).toHaveLength(2);
    const laid = screen.getByTestId('cuarenta-action-1');
    expect(laid).toHaveTextContent('場に置いた');
    expect(laid.querySelectorAll('img')).toHaveLength(1);
  });

  it('announces a ronda-only bonus with the complete badge text', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7)],
          isCaida: false,
          isLimpia: false,
          rondaBonus: 1,
        },
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('ロンダ! +1')).toBeInTheDocument());
    expect(screen.getByTestId('cuarenta-bonus-announce').textContent).toBe('ロンダ! +1');
  });

  it('announces a limpia-only bonus with the complete badge text', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7)],
          isCaida: false,
          isLimpia: true,
          rondaBonus: 0,
        },
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('リンピア! +1')).toBeInTheDocument());
    expect(screen.getByTestId('cuarenta-bonus-announce').textContent).toBe('リンピア! +1');
  });

  it('keeps the Caída live region mounted and empty for a plain capture', async () => {
    mockExec.mockResolvedValue(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7)],
          isCaida: false,
          isLimpia: false,
          rondaBonus: 0,
        },
      }),
    );
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('直前のプレイ')).toBeInTheDocument());
    const announce = screen.getByTestId('cuarenta-bonus-announce');
    expect(announce).toHaveAttribute('role', 'status');
    expect(announce).toHaveAttribute('aria-live', 'polite');
    expect(announce).toBeEmptyDOMElement();
  });

  it('chimes once when a fresh human bonus lands, but not on a plain play', async () => {
    renderWithProviders(<CuarentaPage />);
    const cardBtn = await screen.findByTestId('hand-card-1');
    // A plain capture (no bonus) must not chime.
    mockExec.mockResolvedValueOnce(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('CLOVER', 7),
          capturedCards: [card('CLOVER', 7)],
          isCaida: false,
          isLimpia: false,
          rondaBonus: 0,
        },
      }),
    );
    fireEvent.click(cardBtn);
    await waitFor(() => expect(screen.getByText('直前のプレイ')).toBeInTheDocument());
    expect(soundCalls('chipClick')).toBe(0);

    // A subsequent bonus play chimes exactly once.
    const nextCard = await screen.findByTestId('hand-card-0');
    mockExec.mockResolvedValueOnce(
      makeState({
        humanAction: {
          playerIdx: 0,
          playedCard: card('SPADE', 5),
          capturedCards: [card('SPADE', 5)],
          isCaida: true,
          isLimpia: false,
          rondaBonus: 0,
        },
      }),
    );
    fireEvent.click(nextCard);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('chipClick', { pitchVariation: 0.1 }));
    expect(soundCalls('chipClick')).toBe(1);
  });

  it('shows the win message when the human team wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('あなたのチームの勝利です！')).toBeInTheDocument());
  });

  it('shows the lose message naming the winning team', async () => {
    mockExec.mockResolvedValue(cpuWinState);
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('チームB の勝利です。')).toBeInTheDocument());
  });

  it('shows a next-round button at round end and dispatches next', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CuarentaPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows each team’s round points and bonus breakdown separately from cumulative scores', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 1,
        currentTurn: -1,
        lastRoundDetail: {
          capturedCount: { '0': 23, '1': 17 },
          caida: { '0': 2, '1': 0 },
          ronda: { '0': 4, '1': 0 },
          limpia: { '0': 1, '1': 0 },
          mostCards: 0,
          gained: { '0': 7, '1': 0 },
        },
      }),
    );
    renderWithProviders(<CuarentaPage />);

    const summary = await screen.findByRole('region', { name: '今回のラウンド得点' });
    expect(within(summary).getByText('チームA')).toBeInTheDocument();
    expect(within(summary).getByText('チームB')).toBeInTheDocument();
    expect(within(summary).getByText('今回の獲得点: 7点')).toBeInTheDocument();
    expect(within(summary).getByText('今回の獲得点: 0点')).toBeInTheDocument();
    expect(within(summary).getByText('カイーダ: 2点')).toBeInTheDocument();
    expect(within(summary).getByText('ロンダ: 4点')).toBeInTheDocument();
    expect(within(summary).getByText('リンピア: 1点')).toBeInTheDocument();
    expect(summary).toHaveTextContent('最多捕獲ボーナス: 獲得 (+6点)');
    expect(summary).toHaveTextContent('最多捕獲ボーナス: なし');
    expect(screen.getByText('チームA: 12 / 40点')).toBeInTheDocument();
  });

  it('explains when round score details are unavailable', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CuarentaPage />);

    expect(await screen.findByText('このラウンドの得点内訳はありません。')).toBeInTheDocument();
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<CuarentaPage />);
    await waitFor(() => expect(screen.getByText('プレイヤー')).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the settings panel', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CuarentaPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
