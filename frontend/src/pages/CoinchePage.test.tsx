import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { coincheApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CoincheResponse } from '../types/card';
import { CoinchePhase } from '../types/phases';
import { CoinchePage } from './CoinchePage';

vi.mock('../api/gameApi', () => ({
  coincheApi: { exec: vi.fn() },
  actionLogApi: { coinche: vi.fn() },
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

const mockExec = vi.mocked(coincheApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CoincheResponse> = {}): CoincheResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 11), card('HEART', 10), card('CLOVER', 9)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: CoinchePhase.BID,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 0,
    contractPoints: 0,
    multiplier: 1,
    double: 0,
    biddablePoints: [80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 250],
    makerTeam: 0,
    makerPlayerIdx: -1,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundBeloteBonus: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: {
      cpuDifficulty: 1,
      targetScore: 1000,
      dixDeDer: 10,
      enableBeloteRebelote: true,
    },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: CoinchePhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [1010, 600],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(initialState);
});

describe('CoinchePage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<CoinchePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 1000,
        dixDeDer: 10,
        enableBeloteRebelote: true,
      }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so a prefixed key resolved to the literal
  // "phase.phase.bid" on screen. See issue #4374.
  it('renders the translated phase name, not the raw i18n key', async () => {
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('競り'));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('offers every trump suit and a pass in the auction', async () => {
    renderWithProviders(<CoinchePage />);
    // Coinche には表向き札の除外が無いので、4 スートすべて宣言できる。
    await waitFor(() => expect(screen.getByRole('button', { name: /♠ スペードで宣言/ })).toBeInTheDocument());
    for (const name of [/♣ クラブで宣言/, /♥ ハートで宣言/, /♦ ダイヤで宣言/]) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument();
    }
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  // **契約は「点 + 切り札」の対。** 点を選ばないうちにスートだけ押せると、
  // 残りに既定値が入って別の契約になる。
  it('cannot bid a suit until a target has been chosen', async () => {
    renderWithProviders(<CoinchePage />);
    const spade = await screen.findByRole('button', { name: /♠ スペードで宣言/ });
    expect(spade).toBeDisabled();

    fireEvent.change(screen.getByLabelText('目標点'), { target: { value: '110' } });
    expect(spade).toBeEnabled();
  });

  it('dispatches the bid with both the target and the suit', async () => {
    renderWithProviders(<CoinchePage />);
    await screen.findByRole('button', { name: /♠ スペードで宣言/ });
    fireEvent.change(screen.getByLabelText('目標点'), { target: { value: '110' } });
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /♠ スペードで宣言/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 110, 1));
  });

  // **上回れない契約はボタンに出さない。** 打てば必ず拒否される値を並べると、
  // 選べるのに何も起きない操作面ができる。
  it('offers only the targets that outrank the standing bid', async () => {
    mockExec.mockResolvedValue(makeState({ biddablePoints: [250] }));
    renderWithProviders(<CoinchePage />);
    const select = (await screen.findByLabelText('目標点')) as HTMLSelectElement;
    const values = Array.from(select.options)
      .map((o) => o.value)
      .filter(Boolean);
    expect(values).toEqual(['250']);
  });

  it('shows the doubling choices to the defending side only', async () => {
    // 人間はチーム0。宣言側がチーム1なら守備側なのでコワンシュできる。
    mockExec.mockResolvedValue(makeState({ phase: CoinchePhase.DOUBLE, makerTeam: 1, contractPoints: 120 }));
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コワンシュ (x2)' })).toBeInTheDocument());
    // シュルコワンシュは倍化された宣言側だけ。
    expect(screen.queryByRole('button', { name: 'シュルコワンシュ (x4)' })).not.toBeInTheDocument();
  });

  it('offers surcoinche to the declaring side once coinched', async () => {
    mockExec.mockResolvedValue(makeState({ phase: CoinchePhase.DOUBLE, makerTeam: 0, double: 1, contractPoints: 120 }));
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'シュルコワンシュ (x4)' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'コワンシュ (x2)' })).not.toBeInTheDocument();
  });

  // **契約と倍率は精算そのもの。** 出さないと、同じカード点でも勝ち負けが
  // 変わる理由が画面から読めない。
  it('names the contract and the multiplier once the auction settles', async () => {
    mockExec.mockResolvedValue(makeState({ contractPoints: 120, makerTeam: 0, multiplier: 2 }));
    renderWithProviders(<CoinchePage />);
    const line = await screen.findByTestId('co-contract');
    expect(line).toHaveTextContent('120');
    expect(line).toHaveTextContent('x2');
  });

  it('shows no contract line before the auction settles', async () => {
    renderWithProviders(<CoinchePage />);
    await screen.findByRole('button', { name: 'パス' });
    expect(screen.queryByTestId('co-contract')).not.toBeInTheDocument();
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('renders bonus trackers (dim by default) during play after trump is set', async () => {
    mockExec.mockResolvedValue(makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 3, makerTeam: 0 }));
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByTestId('coinche-bonus-trackers')).toBeInTheDocument());
    expect(screen.getByTestId('dix-de-der-badge')).not.toHaveAttribute('data-active');
    expect(screen.getByTestId('coinche-rebelote-badge')).not.toHaveAttribute('data-active');
  });

  it('activates the dix-de-der badge on the 8th trick', async () => {
    mockExec.mockResolvedValue(makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 8, makerTeam: 0 }));
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByTestId('dix-de-der-badge')).toHaveAttribute('data-active', 'true'));
  });

  // #5592: バッジの点数は設定から。訳文に +10 と書くと、設定を変えたとき
  // バッジだけが古い数字を出す。
  it('shows the configured dix-de-der points, not a hardcoded 10', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: CoinchePhase.PLAY,
        trumpSuit: 1,
        trickNumber: 8,
        makerTeam: 0,
        config: { ...makeState().config, dixDeDer: 25 },
      }),
    );
    renderWithProviders(<CoinchePage />);
    const badge = await screen.findByTestId('dix-de-der-badge');
    expect(badge).toHaveTextContent('25');
    expect(badge).not.toHaveTextContent('10');
  });

  it('activates the coinche/rebelote badge once the maker team earns the bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [20, 0] }),
    );
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByTestId('coinche-rebelote-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('activates the coinche/rebelote badge when the defender team earns the bonus too', async () => {
    // Backend awards the bonus to whichever team plays K+Q of trump, not the maker.
    mockExec.mockResolvedValue(
      makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [0, 20] }),
    );
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByTestId('coinche-rebelote-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('chimes and shows a confirmation banner when the coinche bonus is freshly earned', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 3, currentPlayerIdx: 0, makerTeam: 0 }),
    );
    renderWithProviders(<CoinchePage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ J' });
    fireEvent.click(cardBtn);
    // The play that lands the K+Q-of-trump bonus.
    mockExec.mockResolvedValueOnce(
      makeState({
        phase: CoinchePhase.PLAY,
        trumpSuit: 1,
        trickNumber: 3,
        currentPlayerIdx: 1,
        makerTeam: 0,
        roundBeloteBonus: [20, 0],
      }),
    );
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(screen.getByTestId('coinche-bonus-confirmed')).toBeInTheDocument());
    expect(mockPlaySound).toHaveBeenCalledWith('winFanfare');
  });

  it('translates a known hint reason', async () => {
    const playState = makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0, makerTeam: 0 });
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValueOnce({ ...playState, hint: { cardIndex: 0, reason: 'trump_cut' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/切り札でカット/)).toBeInTheDocument());
  });

  it('falls back to a generic label for an unknown hint reason', async () => {
    const playState = makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0, makerTeam: 0 });
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // Omit cardIndex to also exercise the `?? '-'` fallback for a missing index.
    mockExec.mockResolvedValueOnce({ ...playState, hint: { reason: 'brand_new_reason' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Unknown reason -> hintReason.fallback, not the raw identifier.
    await waitFor(() => expect(screen.getByText(/最善手/)).toBeInTheDocument());
    expect(screen.queryByText(/brand_new_reason/)).not.toBeInTheDocument();
  });

  it('auto-hides the confirmation banner after the display window closes', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      mockExec.mockResolvedValue(
        makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 3, currentPlayerIdx: 0, makerTeam: 0 }),
      );
      renderWithProviders(<CoinchePage />);
      const cardBtn = await screen.findByRole('button', { name: '♠ J' });
      fireEvent.click(cardBtn);
      mockExec.mockResolvedValueOnce(
        makeState({
          phase: CoinchePhase.PLAY,
          trumpSuit: 1,
          trickNumber: 3,
          currentPlayerIdx: 1,
          makerTeam: 0,
          roundBeloteBonus: [20, 0],
        }),
      );
      fireEvent.click(screen.getByRole('button', { name: '出す' }));
      await waitFor(() => expect(screen.getByTestId('coinche-bonus-confirmed')).toBeInTheDocument());
      await act(async () => {
        vi.advanceTimersByTime(2600);
      });
      await waitFor(() => expect(screen.queryByTestId('coinche-bonus-confirmed')).not.toBeInTheDocument());
    } finally {
      vi.useRealTimers();
    }
  });

  it('does not chime on a plain play that earns no new bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 3, currentPlayerIdx: 0, makerTeam: 0 }),
    );
    renderWithProviders(<CoinchePage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ J' });
    fireEvent.click(cardBtn);
    mockExec.mockResolvedValueOnce(
      makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 4, currentPlayerIdx: 1, makerTeam: 0 }),
    );
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
    await waitFor(() => expect(screen.queryByTestId('coinche-bonus-confirmed')).not.toBeInTheDocument());
    expect(soundCalls('winFanfare')).toBe(0);
  });

  it('does not chime when loaded into a round that already has the bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: CoinchePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [20, 0] }),
    );
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByTestId('coinche-rebelote-badge')).toHaveAttribute('data-active', 'true'));
    expect(soundCalls('winFanfare')).toBe(0);
    expect(screen.queryByTestId('coinche-bonus-confirmed')).not.toBeInTheDocument();
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CoinchePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('shows the hint button during the bid phase and requests a hint', async () => {
    renderWithProviders(<CoinchePage />); // default state is BID_PICK_UP, human's bid turn
    const btn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(initialState);
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('renders a bid hint carrying both the target and the suit', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { bid: 110, suit: 1, reason: 'strategic_bid' } }));
    renderWithProviders(<CoinchePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ヒント' }));
    // 点だけ言って何で取るのか言わない助言にはしない。行そのものを見る:
    // スート名はボタンにも出ているので、素の getByText では区別できない。
    await waitFor(() => expect(screen.getByText(/おすすめ: 110点を♠ スペードで宣言/)).toBeInTheDocument());
  });

  it('renders a pass hint with no target', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { reason: 'pass_recommended' } }));
    renderWithProviders(<CoinchePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.queryByText(/おすすめ: 110/)).not.toBeInTheDocument());
  });

  it('rings only the legal follow-suit card during the human play turn', async () => {
    // Trump ♠(1). Opponent leads ♥; the human holds ♥10 so must follow ♥.
    // Only ♥10 is legal; ♠J and ♣9 are illegal.
    mockExec.mockResolvedValue(
      makeState({
        phase: CoinchePhase.PLAY,
        trumpSuit: 1,
        currentPlayerIdx: 0,
        makerTeam: 0,
        currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }],
      }),
    );
    renderWithProviders(<CoinchePage />);
    const legalCard = await screen.findByRole('button', { name: '♥ 10' });
    const illegalCard = screen.getByRole('button', { name: '♠ J' });
    expect(legalCard).toHaveAttribute('data-legal', 'true');
    expect(illegalCard).not.toHaveAttribute('data-legal');
  });

  it('keeps an illegal card clickable so the backend still validates the play', async () => {
    // Same setup: ♠J is illegal (must follow ♥) but must remain selectable —
    // the highlight is additive only and never blocks clicks (see hearts #3977).
    mockExec.mockResolvedValue(
      makeState({
        phase: CoinchePhase.PLAY,
        trumpSuit: 1,
        currentPlayerIdx: 0,
        makerTeam: 0,
        currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }],
      }),
    );
    renderWithProviders(<CoinchePage />);
    const illegalCard = await screen.findByRole('button', { name: '♠ J' });
    expect(illegalCard).not.toHaveAttribute('aria-disabled');
    // The Play button is disabled until a card is selected.
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
    fireEvent.click(illegalCard);
    // Clicking the illegal card selects it and enables Play — no client-side block.
    expect(illegalCard).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('does not ring any card during a CPU play turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: CoinchePhase.PLAY,
        trumpSuit: 1,
        currentPlayerIdx: 1,
        makerTeam: 0,
        currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }],
      }),
    );
    renderWithProviders(<CoinchePage />);
    const humanCard = await screen.findByRole('button', { name: '♥ 10' });
    expect(humanCard).not.toHaveAttribute('data-legal');
  });
});
