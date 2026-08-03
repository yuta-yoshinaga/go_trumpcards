import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { pinochleApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PinochleResponse } from '../types/card';
import { PinochlePage } from './PinochlePage';

vi.mock('../api/gameApi', () => ({
  pinochleApi: { exec: vi.fn() },
  actionLogApi: { pinochle: vi.fn() },
}));

const mockExec = vi.mocked(pinochleApi.exec);

const makePlayers = (overrides?: Partial<PinochleResponse['players'][number]>[]) =>
  [0, 1, 2, 3].map((id) => ({
    id,
    isHuman: id === 0,
    cardCount: 12,
    cards: [],
    team: id % 2,
    trickCount: 0,
    bid: 0,
    hasPassed: false,
    meldScore: 0,
    trickPoints: 0,
    ...(overrides?.[id] ?? {}),
  }));

const bidPhaseState: PinochleResponse = {
  players: makePlayers(),
  phase: 0, // BID
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 0,
  highestBid: 0,
  highestBidder: -1,
  currentTrick: [],
  teamScores: [0, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: -1,
  playerMelds: [[], [], [], []],
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 1500 },
};

const playPhaseState: PinochleResponse = {
  ...bidPhaseState,
  phase: 3, // PLAY
  trumpSuit: 1,
  highestBid: 25,
  highestBidder: 0,
  players: makePlayers([
    {
      id: 0,
      isHuman: true,
      cardCount: 12,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      team: 0,
      trickCount: 0,
      bid: 25,
      hasPassed: false,
      meldScore: 40,
      trickPoints: 0,
    },
  ]),
  validPlayIndices: [0, 1],
};

const trumpPhaseState: PinochleResponse = {
  ...bidPhaseState,
  phase: 1, // TRUMP
  currentPlayerIdx: 0,
};

const meldPhaseState: PinochleResponse = {
  ...bidPhaseState,
  phase: 2, // MELD
  trumpSuit: 2,
};

const trickEndState: PinochleResponse = {
  ...playPhaseState,
  phase: 4, // TRICK_END
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
    { playerIdx: 1, card: { design: 'HEART', value: 10 } },
  ],
};

const roundEndState: PinochleResponse = {
  ...playPhaseState,
  phase: 5, // ROUND_END
};

const gameEndState: PinochleResponse = {
  ...playPhaseState,
  phase: 6, // GAME_END
  gameEndFlag: true,
  winnerTeam: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(bidPhaseState);
});

afterEach(() => {
  localStorage.clear();
});

describe('PinochlePage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PinochlePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 1500 }),
    );
  });

  it('renders bid phase with bid and pass buttons', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
    });
  });

  it('calls bid command when bid button is clicked', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, expect.any(Number)));
  });

  it('labels the bid input and renders 44px steppers', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());
    // Accessible label including the minimum (highestBid 0 → min 20).
    expect(screen.getByLabelText('ビッド額（最低 20）')).toBeInTheDocument();
    const stepper = screen.getByRole('button', { name: 'ビッド額（最低 20） −5' });
    expect(stepper.className).toContain('min-h-[44px]');
    expect(stepper.className).toContain('min-w-[44px]');
  });

  it('blocks a below-minimum bid on the client and shows the reason', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    fireEvent.change(screen.getByLabelText('ビッド額（最低 20）'), { target: { value: '5' } });
    const bidBtn = screen.getByRole('button', { name: 'ビッド' });
    expect(bidBtn).toHaveAttribute('aria-disabled', 'true');
    expect(screen.getByRole('alert')).toHaveTextContent('ビッド額は 20 以上にしてください');

    mockExec.mockClear();
    fireEvent.click(bidBtn);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('bid', undefined, undefined, expect.anything());
  });

  it('blocks an empty (NaN) bid on the client', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText('ビッド額（最低 20）'), { target: { value: '' } });
    expect(screen.getByRole('button', { name: 'ビッド' })).toHaveAttribute('aria-disabled', 'true');
  });

  it('calls pass command when pass button is clicked', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('renders trump phase with suit buttons', async () => {
    mockExec.mockResolvedValue(trumpPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '♠' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♣' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♦' })).toBeInTheDocument();
    });
  });

  it('calls trump command when suit button is clicked', async () => {
    mockExec.mockResolvedValue(trumpPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♠' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, undefined, 1));
  });

  it('renders meld phase with confirm melds button', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'メルド確認' })).toBeInTheDocument();
    });
  });

  it('calls meld command when confirm melds button is clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルド確認' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'メルド確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld'));
  });

  it('clicking a meld badge highlights the matching cards in the human hand', async () => {
    // Construct a meld-phase state where the human holds the cards forming
    // a Common Marriage (K♥ + Q♥), plus an unrelated A♠ that should NOT
    // glow when the marriage badge is selected.
    const meldStateWithHumanMeld: PinochleResponse = {
      ...meldPhaseState,
      players: makePlayers([
        {
          id: 0,
          isHuman: true,
          cardCount: 3,
          cards: [
            { design: 'HEART', value: 13 },
            { design: 'HEART', value: 12 },
            { design: 'SPADE', value: 1 },
          ],
        },
      ]),
      playerMelds: [
        [
          {
            type: 1, // PinochleMeldCommonMarriage
            points: 20,
            cards: [
              { design: 'HEART', value: 13 },
              { design: 'HEART', value: 12 },
            ],
          },
        ],
        [],
        [],
        [],
      ],
    };
    mockExec.mockResolvedValue(meldStateWithHumanMeld);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-meld-badge-0')).toBeInTheDocument());

    // Before click — no card carries the highlight marker.
    expect(screen.queryByLabelText('♥ K')).not.toHaveAttribute('data-meld-highlighted');
    expect(screen.queryByLabelText('♥ Q')).not.toHaveAttribute('data-meld-highlighted');

    fireEvent.click(screen.getByTestId('pn-meld-badge-0'));

    // After click — both Hearts in the meld glow; the unrelated Spade A doesn't.
    expect(screen.getByLabelText('♥ K')).toHaveAttribute('data-meld-highlighted', 'true');
    expect(screen.getByLabelText('♥ Q')).toHaveAttribute('data-meld-highlighted', 'true');
    expect(screen.getByLabelText('♠ A')).not.toHaveAttribute('data-meld-highlighted');

    // Clicking the active badge a second time clears the highlight.
    fireEvent.click(screen.getByTestId('pn-meld-badge-0'));
    expect(screen.getByLabelText('♥ K')).not.toHaveAttribute('data-meld-highlighted');
  });

  it('renders a persistent meld-card badge ("M") on every card that scored in a meld', async () => {
    const meldStateWithHumanMeld: PinochleResponse = {
      ...meldPhaseState,
      players: makePlayers([
        {
          id: 0,
          isHuman: true,
          cardCount: 3,
          cards: [
            { design: 'HEART', value: 13 },
            { design: 'HEART', value: 12 },
            { design: 'SPADE', value: 1 },
          ],
        },
      ]),
      playerMelds: [
        [
          {
            type: 1,
            points: 20,
            cards: [
              { design: 'HEART', value: 13 },
              { design: 'HEART', value: 12 },
            ],
          },
        ],
        [],
        [],
        [],
      ],
    };
    mockExec.mockResolvedValue(meldStateWithHumanMeld);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-meld-card-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('pn-meld-card-badge-1')).toBeInTheDocument();
    expect(screen.queryByTestId('pn-meld-card-badge-2')).not.toBeInTheDocument();
    expect(screen.getByLabelText('♥ K')).toHaveAttribute('data-in-meld', 'true');
    expect(screen.getByLabelText('♠ A')).not.toHaveAttribute('data-in-meld');
  });

  it('keeps the meld-card badge visible in PLAY phase so users do not discard meld components', async () => {
    const playStateWithHumanMeld: PinochleResponse = {
      ...playPhaseState,
      players: makePlayers([
        {
          id: 0,
          isHuman: true,
          cardCount: 3,
          cards: [
            { design: 'HEART', value: 13 },
            { design: 'HEART', value: 12 },
            { design: 'SPADE', value: 1 },
          ],
        },
      ]),
      validPlayIndices: [0, 1, 2],
      playerMelds: [
        [
          {
            type: 1,
            points: 20,
            cards: [
              { design: 'HEART', value: 13 },
              { design: 'HEART', value: 12 },
            ],
          },
        ],
        [],
        [],
        [],
      ],
    };
    mockExec.mockResolvedValue(playStateWithHumanMeld);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-meld-card-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('pn-meld-card-badge-1')).toBeInTheDocument();
    expect(screen.queryByTestId('pn-meld-card-badge-2')).not.toBeInTheDocument();
    // Tooltip carries the localized meld type for the components.
    expect(screen.getByLabelText('♥ K')).toHaveAttribute('title', 'コモンマリッジ');
  });

  it('renders play phase with human cards', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument();
    });
  });

  it('plays card when card button is clicked in play phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♠ A' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('renders trick end phase with next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument();
    });
  });

  it('calls next command when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('renders round end phase with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument();
    });
  });

  it('calls nextround command when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders trick cards with localized player names (not P0/P1)', async () => {
    mockExec.mockResolvedValue(trickEndState);
    const { container } = renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="pn-trick-display"]')).toBeInTheDocument());
    const trick = container.querySelector('[data-tutorial="pn-trick-display"]') as HTMLElement;
    // playerIdx 0 is the human ("あなた"), playerIdx 1 a CPU ("CPU 1").
    expect(within(trick).getByText('あなた')).toBeInTheDocument();
    expect(within(trick).getByText('CPU 1')).toBeInTheDocument();
    expect(within(trick).queryByText('P0')).not.toBeInTheDocument();
  });

  it('renders game end state', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
    });
  });

  it('shows reset dialog and can cancel', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByRole('alertdialog')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument());
  });

  it('shows trump suit info when set', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getByText(/切り札/)).toBeInTheDocument();
    });
  });

  it('shows team scores', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => {
      expect(screen.getAllByText(/チーム 0/).length).toBeGreaterThan(0);
    });
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled and state has backend hint', async () => {
    localStorage.setItem('hint_enabled_pinochle', 'true');
    // provide a state where state.hint is set so getPinochleHint returns non-null
    const hintState: PinochleResponse = { ...bidPhaseState, hint: { reason: 'hint_bid', cardIndex: 0 } };
    mockExec.mockResolvedValue(hintState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows a server-hint button on the human bid turn', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-hint-button')).toBeInTheDocument());
  });

  it('fetches and displays the recommended bid, and applies it to the input', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { bidAmount: 30, reason: 'hint_bid' } } as unknown as PinochleResponse);
    fireEvent.click(screen.getByTestId('pn-hint-button'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    const hintBox = await screen.findByTestId('pn-server-hint');
    expect(hintBox).toHaveTextContent('推奨ビッド: 30');

    // Applying the hint updates the bid input value.
    fireEvent.click(screen.getByTestId('pn-hint-apply-bid'));
    expect(screen.getByLabelText('ビッド額（最低 20）')).toHaveDisplayValue('30');
  });

  it('displays a pass recommendation from the server hint', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { pass: true, reason: 'hint_pass' } } as unknown as PinochleResponse);
    fireEvent.click(screen.getByTestId('pn-hint-button'));

    const hintBox = await screen.findByTestId('pn-server-hint');
    expect(hintBox).toHaveTextContent('推奨: パス');
    // No apply-bid button for a pass recommendation.
    expect(screen.queryByTestId('pn-hint-apply-bid')).not.toBeInTheDocument();
  });

  it('displays a card recommendation on the human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { cardIndex: 1, reason: 'hint_play' } } as unknown as PinochleResponse);
    fireEvent.click(screen.getByTestId('pn-hint-button'));

    const hintBox = await screen.findByTestId('pn-server-hint');
    expect(hintBox).toHaveTextContent('推奨プレイ');
    expect(hintBox).toHaveTextContent('[1]');
  });

  it('shows an error alert when the hint request fails', async () => {
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByTestId('pn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockRejectedValueOnce(new Error('network'));
    fireEvent.click(screen.getByTestId('pn-hint-button'));

    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    expect(screen.queryByTestId('pn-server-hint')).not.toBeInTheDocument();
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PinochlePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 1500 }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });
});
