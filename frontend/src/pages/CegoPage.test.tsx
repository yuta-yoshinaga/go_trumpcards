import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cegoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCegoState } from '../test/stateFactories';
import { CegoPage } from './CegoPage';

vi.mock('../api/gameApi', () => ({
  cegoApi: { exec: vi.fn() },
  actionLogApi: { cego: vi.fn() },
}));

const mockExec = vi.mocked(cegoApi.exec);

const suit = (value: number, design: 'HEART' | 'SPADE' | 'CLOVER' | 'DIAMOND', glyph: string, label: string) => ({
  design,
  value,
  glyph,
  label,
  color: design === 'HEART' || design === 'DIAMOND' ? 'red' : 'black',
  deck: 'tarot',
});

const cpuPlayers = [
  { id: 1, isHuman: false, cardCount: 11, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
  { id: 2, isHuman: false, cardCount: 11, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
  { id: 3, isHuman: false, cardCount: 11, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
];

const playPhaseState = makeCegoState();

const bidPhaseState = makeCegoState({
  phase: 0,
  isHumanTurn: true,
  isHumanBidTurn: true,
  contract: 0,
  contractType: 0,
  highestBid: 0,
  highestBidder: -1,
  declarerIdx: -1,
  playableIndices: [],
});

const contractPhaseState = makeCegoState({
  phase: 1,
  isHumanTurn: true,
  isHumanBidTurn: false,
  isHumanContract: true,
  contract: 1,
  contractType: 0,
  declarerIdx: 0,
  playableIndices: [],
});

const exchangePhaseState = makeCegoState({
  phase: 2,
  isHumanTurn: true,
  isHumanBidTurn: false,
  isHumanExchange: true,
  contract: 1,
  contractType: 1,
  declarerIdx: 0,
  playableIndices: [],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 11,
      cards: [suit(2, 'HEART', '♥', '2'), suit(3, 'HEART', '♥', '3'), suit(4, 'HEART', '♥', '4')],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: true,
    },
    ...cpuPlayers,
  ],
});

const trickEndState = makeCegoState({
  phase: 4,
  isHumanTurn: false,
  currentTrick: [
    { playerIdx: 0, card: suit(7, 'HEART', '♥', 'Q') },
    { playerIdx: 1, card: suit(8, 'CLOVER', '♣', 'K') },
  ],
});

const roundEndState = makeCegoState({
  phase: 5,
  isHumanTurn: false,
  outcome: 1,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 0,
      cards: [],
      trickCount: 6,
      cardPoints: 60,
      score: 3,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 0, cards: [], trickCount: 2, cardPoints: 16, score: -1, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 0, cards: [], trickCount: 2, cardPoints: 16, score: -1, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 0, cards: [], trickCount: 1, cardPoints: 14, score: -1, isDeclarer: false },
  ],
});

const gameEndState = makeCegoState({
  phase: 6,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});

const gameEndDrawState = makeCegoState({
  phase: 6,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: -1,
  message: 'ゲーム終了！ 引き分け！',
});

const cpuTurnState = makeCegoState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CegoPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CegoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<CegoPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetDeals: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    renderWithProviders(<CegoPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '21 ✦' })).toBeInTheDocument();
    });
    expect(screen.getAllByText('デクレアラー').length).toBeGreaterThan(0);
  });

  it('renders the bid phase with Pass and the Declare button', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プレイ宣言' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('declaring dispatches bid with the play string', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CegoPage />);
    const bidBtn = await screen.findByRole('button', { name: 'プレイ宣言' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bidBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'play' }));
  });

  it('passing dispatches the pass command', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CegoPage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('does not show bid controls on a CPU/non-bid turn', async () => {
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'プレイ宣言' })).not.toBeInTheDocument();
  });

  it('renders the contract phase with both contract buttons', async () => {
    mockExec.mockResolvedValue(contractPhaseState);
    renderWithProviders(<CegoPage />);
    expect(await screen.findByRole('button', { name: 'チェゴ' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'ハントシュピール' })).toBeEnabled();
  });

  it('choosing Cego dispatches the contract command', async () => {
    mockExec.mockResolvedValue(contractPhaseState);
    renderWithProviders(<CegoPage />);
    const cegoBtn = await screen.findByRole('button', { name: 'チェゴ' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(contractPhaseState);
    fireEvent.click(cegoBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 'cego' }));
  });

  it('choosing Handspiel dispatches the contract command', async () => {
    mockExec.mockResolvedValue(contractPhaseState);
    renderWithProviders(<CegoPage />);
    const handspielBtn = await screen.findByRole('button', { name: 'ハントシュピール' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(contractPhaseState);
    fireEvent.click(handspielBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 'handspiel' }));
  });

  it('shows the contract explainer describing Cego vs Handspiel during contract selection', async () => {
    mockExec.mockResolvedValue(contractPhaseState);
    renderWithProviders(<CegoPage />);
    const panel = await screen.findByTestId('cego-contract-explainer');
    // Cego side: mentions the blind (盲札) and its count (10 by default).
    expect(panel).toHaveTextContent(/盲札/);
    expect(panel).toHaveTextContent(/10枚の場札/);
    // Handspiel side: plays the dealt hand without the blind.
    expect(panel).toHaveTextContent(/配られた手札のまま/);
  });

  it('does not show the contract explainer outside contract selection', async () => {
    mockExec.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<CegoPage />);
    await screen.findByRole('button', { name: '2 ♥' });
    expect(screen.queryByTestId('cego-contract-explainer')).not.toBeInTheDocument();
  });

  it('renders the exchange phase and keeps the keep button disabled until 1 card is chosen', async () => {
    mockExec.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<CegoPage />);
    await screen.findByRole('button', { name: '2 ♥' });
    expect(screen.getByRole('button', { name: /残す/ })).toBeDisabled();
  });

  it('keeping exactly 1 card dispatches discard with that single index', async () => {
    mockExec.mockResolvedValue(exchangePhaseState);
    renderWithProviders(<CegoPage />);
    const card = await screen.findByRole('button', { name: '3 ♥' });
    fireEvent.click(card);
    const keepBtn = screen.getByRole('button', { name: /残す/ });
    expect(keepBtn).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(exchangePhaseState);
    fireEvent.click(keepBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [1] }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<CegoPage />);
    const card = await screen.findByRole('button', { name: 'K ♥' });
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByTestId('cego-result')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('renders a draw at game end without a false win celebration', async () => {
    mockExec.mockResolvedValue(gameEndDrawState);
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ 引き分け！')).toBeInTheDocument());
  });

  it('the next-game button at game end resets immediately', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CegoPage />);
    const nextGame = await screen.findByRole('button', { name: '次のゲーム' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameEndState);
    fireEvent.click(nextGame);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', expect.anything()));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('changing the CPU difficulty and target-deals selects updates the config', async () => {
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    const difficulty = screen.getByLabelText('CPU難易度') as HTMLSelectElement;
    fireEvent.change(difficulty, { target: { value: '2' } });
    expect(difficulty.value).toBe('2');
    const deals = screen.getByLabelText('マッチのディール数') as HTMLSelectElement;
    fireEvent.change(deals, { target: { value: '3' } });
    expect(deals.value).toBe('3');
  });

  it('renders the backend hint banner with its card indices', async () => {
    mockExec.mockResolvedValue(makeCegoState({ hint: { cardIndices: [0, 2], reason: 'lead_high' } }));
    renderWithProviders(<CegoPage />);
    await waitFor(() => expect(screen.getByText(/\[0\], \[2\]/)).toBeInTheDocument());
  });
});
