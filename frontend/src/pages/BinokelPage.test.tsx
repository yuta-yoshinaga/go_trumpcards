import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { binokelApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BinokelResponse } from '../types/card';
import { BinokelPhase } from '../types/phases';
import { BinokelPage } from './BinokelPage';

vi.mock('../api/gameApi', () => ({
  binokelApi: { exec: vi.fn() },
  actionLogApi: { binokel: vi.fn() },
}));

const mockExec = vi.mocked(binokelApi.exec);

const makePlayers = (overrides?: Partial<BinokelResponse['players'][number]>[]) =>
  [0, 1, 2].map((id) => ({
    id,
    isHuman: id === 0,
    cardCount: 15,
    cards: [],
    trickCount: 0,
    bid: 0,
    hasPassed: false,
    meldScore: 0,
    trickPoints: 0,
    score: 0,
    ...(overrides?.[id] ?? {}),
  }));

const bidPhaseState: BinokelResponse = {
  players: makePlayers(),
  phase: BinokelPhase.BID, // 0
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 2,
  trumpSuit: 0,
  highestBid: 0,
  highestBidder: -1,
  currentTrick: [],
  scores: [0, 0, 0],
  gameEndFlag: false,
  winnerPlayer: -1,
  leadPlayerIdx: -1,
  playerMelds: [[], [], []],
  meldTable: [
    { type: 0, points: 10 },
    { type: 1, points: 20 },
    { type: 2, points: 40 },
    { type: 9, points: 150 },
    { type: 14, points: 1000 },
  ],
  dabb: [],
  dabbDiscarded: [],
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 1500 },
};

const dabbPhaseStateHuman: BinokelResponse = {
  ...bidPhaseState,
  phase: BinokelPhase.DABB, // 1
  highestBid: 150,
  highestBidder: 0,
  currentPlayerIdx: 0,
  dabb: [
    { design: 'SPADE', value: 1 },
    { design: 'HEART', value: 10 },
    { design: 'DIAMOND', value: 12 },
  ],
  players: makePlayers([
    {
      id: 0,
      isHuman: true,
      cardCount: 18,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'SPADE', value: 10 },
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 12 },
        { design: 'HEART', value: 1 },
        { design: 'HEART', value: 10 },
        { design: 'HEART', value: 13 },
        { design: 'HEART', value: 12 },
        { design: 'CLOVER', value: 1 },
        { design: 'CLOVER', value: 10 },
        { design: 'CLOVER', value: 13 },
        { design: 'CLOVER', value: 12 },
        { design: 'DIAMOND', value: 1 },
        { design: 'DIAMOND', value: 10 },
        { design: 'DIAMOND', value: 13 },
        { design: 'DIAMOND', value: 12 },
        { design: 'SPADE', value: 11 },
        { design: 'HEART', value: 11 },
      ],
      bid: 150,
    },
  ]),
};

const dabbPhaseStateCpu: BinokelResponse = {
  ...bidPhaseState,
  phase: BinokelPhase.DABB, // 1
  highestBid: 150,
  highestBidder: 1,
  currentPlayerIdx: 1,
  dabb: [
    { design: 'SPADE', value: 1 },
    { design: 'HEART', value: 10 },
    { design: 'DIAMOND', value: 12 },
  ],
};

const trumpPhaseState: BinokelResponse = {
  ...bidPhaseState,
  phase: BinokelPhase.TRUMP, // 2
  highestBid: 150,
  highestBidder: 0,
  currentPlayerIdx: 0,
};

const meldPhaseState: BinokelResponse = {
  ...bidPhaseState,
  phase: BinokelPhase.MELD, // 3
  highestBid: 150,
  highestBidder: 0,
  trumpSuit: 2,
};

const playPhaseState: BinokelResponse = {
  ...bidPhaseState,
  phase: BinokelPhase.PLAY, // 4
  trumpSuit: 1,
  highestBid: 150,
  highestBidder: 0,
  currentPlayerIdx: 0,
  players: makePlayers([
    {
      id: 0,
      isHuman: true,
      cardCount: 15,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      bid: 150,
      hasPassed: false,
      meldScore: 40,
      trickPoints: 0,
    },
  ]),
  validPlayIndices: [0, 1],
};

const trickEndState: BinokelResponse = {
  ...playPhaseState,
  phase: BinokelPhase.TRICK_END, // 5
  currentTrick: [
    { playerIdx: 0, card: { design: 'SPADE', value: 1 } },
    { playerIdx: 1, card: { design: 'HEART', value: 10 } },
    { playerIdx: 2, card: { design: 'DIAMOND', value: 12 } },
  ],
};

const roundEndState: BinokelResponse = {
  ...playPhaseState,
  phase: BinokelPhase.ROUND_END, // 6
};

const gameEndState: BinokelResponse = {
  ...playPhaseState,
  phase: BinokelPhase.GAME_END, // 7
  gameEndFlag: true,
  winnerPlayer: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(bidPhaseState);
});

afterEach(() => {
  localStorage.clear();
});

describe('BinokelPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BinokelPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
    expect(screen.queryByText('考え中...')).not.toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 1500 }),
    );
  });

  it('renders bid phase with bid and pass buttons', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
    });
  });

  it('calls bid command with minimum bid by default', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, 150));
  });

  it('allows selecting a discrete bid amount and bidding', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bid-option-170')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('bid-option-170'));
    expect(screen.getByTestId('bid-option-170')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('bn-bid-selected')).toHaveTextContent('170');

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ビッド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', undefined, undefined, 170));
  });

  it('does not render bid options below the minimum legal bid (150 initially, or highestBid + 10)', async () => {
    const { unmount } = renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    // When highestBid is 0 (initial), minimum is 150. Options below 150 must NOT exist.
    expect(screen.queryByTestId('bid-option-20')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-140')).not.toBeInTheDocument();
    expect(screen.getByTestId('bid-option-150')).toBeInTheDocument();

    unmount();

    // When highestBid is 160, minimum is 170. Options below 170 must NOT exist.
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 160 });
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());

    expect(screen.queryByTestId('bid-option-20')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-150')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-option-160')).not.toBeInTheDocument();
    expect(screen.getByTestId('bid-option-170')).toBeInTheDocument();
  });

  it('calls pass command when pass button is clicked', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bid-pass')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByTestId('bid-pass'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('renders Dabb phase for human declarer and enforces exactly 3 cards for discard', async () => {
    mockExec.mockResolvedValue(dabbPhaseStateHuman);
    const { container } = renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-dabb-display')).toBeInTheDocument());

    const discardBtn = screen.getByTestId('discard-dabb-button');
    expect(discardBtn).toBeDisabled();

    const handContainer = container.querySelector('[data-tutorial="bn-player-hand"]') as HTMLElement;
    const cardButtons = within(handContainer).getAllByRole('button');
    expect(cardButtons).toHaveLength(18);

    // 0 selected -> disabled
    expect(discardBtn).toBeDisabled();

    // 1 selected -> disabled
    fireEvent.click(cardButtons[0]);
    expect(discardBtn).toBeDisabled();

    // 2 selected -> disabled
    fireEvent.click(cardButtons[1]);
    expect(discardBtn).toBeDisabled();

    // 3 selected -> enabled!
    fireEvent.click(cardButtons[2]);
    expect(discardBtn).toBeEnabled();

    // 4 selected -> disabled again
    fireEvent.click(cardButtons[3]);
    expect(discardBtn).toBeDisabled();

    // Deselect 4th card -> enabled (3 cards)
    fireEvent.click(cardButtons[3]);
    expect(discardBtn).toBeEnabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(trumpPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('discard', undefined, undefined, undefined, undefined, [0, 1, 2]),
    );
  });

  it('renders waiting message when CPU is declarer in Dabb phase', async () => {
    mockExec.mockResolvedValue(dabbPhaseStateCpu);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('dabb-waiting')).toBeInTheDocument());

    expect(screen.getByTestId('dabb-waiting')).toHaveTextContent('CPU 1 が Dabb を整理中...');
    expect(screen.queryByTestId('discard-dabb-button')).not.toBeInTheDocument();
  });

  it('renders trump phase with suit buttons', async () => {
    mockExec.mockResolvedValue(trumpPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '♠' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♣' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♦' })).toBeInTheDocument();
    });
  });

  it('calls trump command when suit button is clicked', async () => {
    mockExec.mockResolvedValue(trumpPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♠' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', undefined, undefined, undefined, 1));
  });

  it('renders meld phase with confirm melds button', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'メルド確認' })).toBeInTheDocument();
    });
  });

  it('calls meld command when confirm melds button is clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BinokelPage />);
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
    const meldStateWithHumanMeld: BinokelResponse = {
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
            type: 1, // BinokelMeldCommonMarriage
            points: 20,
            cards: [
              { design: 'HEART', value: 13 },
              { design: 'HEART', value: 12 },
            ],
          },
        ],
        [],
        [],
      ],
    };
    mockExec.mockResolvedValue(meldStateWithHumanMeld);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-meld-badge-0')).toBeInTheDocument());

    // Before click — no card carries the highlight marker.
    expect(screen.queryByLabelText('♥ K')).not.toHaveAttribute('data-meld-highlighted');
    expect(screen.queryByLabelText('♥ Q')).not.toHaveAttribute('data-meld-highlighted');

    fireEvent.click(screen.getByTestId('bn-meld-badge-0'));

    // After click — both Hearts in the meld glow; the unrelated Spade A doesn't.
    expect(screen.getByLabelText('♥ K')).toHaveAttribute('data-meld-highlighted', 'true');
    expect(screen.getByLabelText('♥ Q')).toHaveAttribute('data-meld-highlighted', 'true');
    expect(screen.getByLabelText('♠ A')).not.toHaveAttribute('data-meld-highlighted');

    // Clicking the active badge a second time clears the highlight.
    fireEvent.click(screen.getByTestId('bn-meld-badge-0'));
    expect(screen.getByLabelText('♥ K')).not.toHaveAttribute('data-meld-highlighted');
  });

  it('renders a persistent meld-card badge ("M") on every card that scored in a meld', async () => {
    const meldStateWithHumanMeld: BinokelResponse = {
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
      ],
    };
    mockExec.mockResolvedValue(meldStateWithHumanMeld);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-meld-card-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('bn-meld-card-badge-1')).toBeInTheDocument();
    expect(screen.queryByTestId('bn-meld-card-badge-2')).not.toBeInTheDocument();
    expect(screen.getByLabelText('♥ K')).toHaveAttribute('data-in-meld', 'true');
    expect(screen.getByLabelText('♠ A')).not.toHaveAttribute('data-in-meld');
  });

  it('keeps the meld-card badge visible in PLAY phase so users do not discard meld components', async () => {
    const playStateWithHumanMeld: BinokelResponse = {
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
      ],
    };
    mockExec.mockResolvedValue(playStateWithHumanMeld);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-meld-card-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('bn-meld-card-badge-1')).toBeInTheDocument();
    expect(screen.queryByTestId('bn-meld-card-badge-2')).not.toBeInTheDocument();
    // Tooltip carries the localized meld type for the components.
    expect(screen.getByLabelText('♥ K')).toHaveAttribute('title', 'コモンマリッジ');
  });

  it('renders play phase with human cards', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument();
    });
  });

  it('plays card when card button is clicked in play phase', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '♠ A' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('renders trick end phase with next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument();
    });
  });

  it('calls next command when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('renders round end phase with next round button', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument();
    });
  });

  it('calls nextround command when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders trick cards with localized player names (not P0/P1)', async () => {
    mockExec.mockResolvedValue(trickEndState);
    const { container } = renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="bn-trick-display"]')).toBeInTheDocument());
    const trick = container.querySelector('[data-tutorial="bn-trick-display"]') as HTMLElement;
    // playerIdx 0 is the human ("あなた"), playerIdx 1 a CPU ("CPU 1").
    expect(within(trick).getByText('あなた')).toBeInTheDocument();
    expect(within(trick).getByText('CPU 1')).toBeInTheDocument();
    expect(within(trick).getByText('CPU 2')).toBeInTheDocument();
    expect(within(trick).queryByText('P0')).not.toBeInTheDocument();
  });

  it('renders game end state', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
    });
  });

  it('shows reset dialog and can cancel', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    await waitFor(() => expect(screen.getByRole('alertdialog')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument());
  });

  it('shows trump suit info when set', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByText(/切り札/)).toBeInTheDocument();
    });
  });

  it('shows individual player scores and declarer info', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => {
      expect(screen.getByText('あなた')).toBeInTheDocument();
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getAllByText(/スコア: 0/).length).toBe(3);
      expect(screen.getAllByText('落札者').length).toBeGreaterThan(0);
    });
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ビッド' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled and state has backend hint', async () => {
    localStorage.setItem('hint_enabled_binokel', 'true');
    // provide a state where state.hint is set so getBinokelHint returns non-null
    const hintState: BinokelResponse = { ...bidPhaseState, hint: { reason: 'hint_bid', cardIndex: 0 } };
    mockExec.mockResolvedValue(hintState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows a server-hint button on the human bid turn', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());
  });

  it('fetches and displays the recommended bid, and applies it to selection', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { bidAmount: 160, reason: 'hint_bid' } } as unknown as BinokelResponse);
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    const hintBox = await screen.findByTestId('bn-server-hint');
    expect(hintBox).toHaveTextContent('推奨ビッド: 160');

    // Applying the hint updates the selected bid option
    fireEvent.click(screen.getByTestId('bn-hint-apply-bid'));
    expect(screen.getByTestId('bid-option-160')).toHaveAttribute('aria-pressed', 'true');
  });

  it('displays a pass recommendation from the server hint', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { pass: true, reason: 'hint_pass' } } as unknown as BinokelResponse);
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    const hintBox = await screen.findByTestId('bn-server-hint');
    expect(hintBox).toHaveTextContent('推奨: パス');
    // No apply-bid button for a pass recommendation.
    expect(screen.queryByTestId('bn-hint-apply-bid')).not.toBeInTheDocument();
  });

  it('displays a card recommendation on the human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { cardIndex: 1, reason: 'hint_play' } } as unknown as BinokelResponse);
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    const hintBox = await screen.findByTestId('bn-server-hint');
    expect(hintBox).toHaveTextContent('推奨プレイ');
    expect(hintBox).toHaveTextContent('[1]');
  });

  it('shows an error alert when the hint request fails', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockRejectedValueOnce(new Error('network'));
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    expect(screen.queryByTestId('bn-server-hint')).not.toBeInTheDocument();
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1, pointLimit: 1500 }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('drops the turn ring once the round is over', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // currentPlayerIdx still holds the last trick winner here; that is not a turn.
    mockExec.mockResolvedValue({ ...playPhaseState, phase: BinokelPhase.ROUND_END, currentPlayerIdx: 2 });
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(document.querySelectorAll('[data-on-turn]')).toHaveLength(0);
  });

  it('marks whose turn it is in the players grid', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // Bidding jumps between seats, which is where the grid was least helpful.
    mockExec.mockResolvedValue({ ...bidPhaseState, bidPlayerIdx: 2 });
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(document.querySelectorAll('[data-on-turn]')).toHaveLength(1));
  });
});

// メルド早見表テスト
describe('BinokelPage meld reference', () => {
  it('lists the melds and the points the server sent, while bidding', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<BinokelPage />);
    const panel = await screen.findByTestId('bn-meld-table');
    // 点数はサーバが送った値をそのまま出す。
    expect(panel).toHaveTextContent('ディクス');
    expect(panel).toHaveTextContent('10');
    expect(panel).toHaveTextContent('切り札ファミリー');
    expect(panel).toHaveTextContent('150');
    expect(panel).toHaveTextContent('ダブルキングアラウンド');
    expect(panel).toHaveTextContent('1000');
  });

  // プレイ中は出さない。盤面を見る場面で参照表は邪魔になる。
  it('is gone once the cards are being played', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド: \d+/)).toBeInTheDocument());
    expect(screen.queryByTestId('bn-meld-table')).not.toBeInTheDocument();
  });

  // 古いサーバや未対応のレスポンスでも落ちないこと。
  it('renders nothing when the server sent no table', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, meldTable: undefined });
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド: \d+/)).toBeInTheDocument());
    expect(screen.queryByTestId('bn-meld-table')).not.toBeInTheDocument();
  });

  it('labels the per-player trick count instead of a bare "T:"', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド: \d+/)).toBeInTheDocument());
    expect(screen.getAllByText(/獲得トリック: 0/).length).toBeGreaterThan(0);
    expect(screen.queryByText(/\| T: /)).not.toBeInTheDocument();
  });

  it('displays a trump recommendation from the server hint', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { suit: 0, reason: 'hint_trump' } } as unknown as BinokelResponse);
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    const hintBox = await screen.findByTestId('bn-server-hint');
    expect(hintBox).toHaveTextContent('推奨トランプ');
    expect(hintBox).not.toHaveTextContent('{{');
  });

  it('displays a play recommendation from the server hint', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { cardIndex: 0, reason: 'hint_play' } } as unknown as BinokelResponse);
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    const hintBox = await screen.findByTestId('bn-server-hint');
    expect(hintBox).toHaveTextContent('推奨プレイ');
    expect(hintBox).toHaveTextContent('[0]');
    expect(hintBox).not.toHaveTextContent('{{');
  });

  // 理由だけのヒント。どの推奨にもならないので本文は空だが、理由は出る。
  it('still names the reason when the hint recommends nothing concrete', async () => {
    renderWithProviders(<BinokelPage />);
    await waitFor(() => expect(screen.getByTestId('bn-hint-button')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ hint: { reason: 'hint_play' } } as unknown as BinokelResponse);
    fireEvent.click(screen.getByTestId('bn-hint-button'));

    const hintBox = await screen.findByTestId('bn-server-hint');
    expect(hintBox).toHaveTextContent('(');
    expect(hintBox).not.toHaveTextContent('推奨プレイ');
    expect(hintBox).not.toHaveTextContent('推奨ビッド');
    expect(hintBox).not.toHaveTextContent('{{');
  });
});
