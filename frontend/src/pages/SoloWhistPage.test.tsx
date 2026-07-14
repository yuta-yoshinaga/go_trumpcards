import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { soloWhistApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSoloWhistState } from '../test/stateFactories';
import { SoloWhistPage } from './SoloWhistPage';

vi.mock('../api/gameApi', () => ({
  soloWhistApi: { exec: vi.fn() },
  actionLogApi: { solowhist: vi.fn() },
}));

const mockExec = vi.mocked(soloWhistApi.exec);

// Default fixture: a human bid turn (bid phase).
const bidPhaseState = makeSoloWhistState();
// A human play turn with a started trick (so the play control is shown).
const playPhaseState = makeSoloWhistState({
  phase: 1,
  declarerIdx: 0,
  contract: 1,
  trumpSuit: 3,
  isHumanBidTurn: false,
  isHumanTurn: true,
  playableIndices: [0, 1, 2],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      trickCount: 0,
      score: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 13, cards: [], trickCount: 0, score: 0, isDeclarer: false },
  ],
});
const cpuTurnState = makeSoloWhistState({
  phase: 1,
  declarerIdx: 1,
  isHumanBidTurn: false,
  isHumanTurn: false,
  currentPlayerIdx: 1,
});
const trickEndState = makeSoloWhistState({ phase: 2, isHumanBidTurn: false });
const roundEndState = makeSoloWhistState({ phase: 3, isHumanBidTurn: false, roundTricks: [8, 2, 2, 1] });
const gameEndState = makeSoloWhistState({
  phase: 4,
  isHumanBidTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidPhaseState);
});

describe('SoloWhistPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SoloWhistPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 21 },
      }),
    );
  });

  it('shows bid buttons on a human bid turn', async () => {
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByTestId('bid-0')).toBeInTheDocument());
    expect(screen.getByTestId('bid-1')).toBeInTheDocument();
    expect(screen.getByTestId('bid-2')).toBeInTheDocument();
    expect(screen.getByTestId('bid-3')).toBeInTheDocument();
  });

  it('dispatches a bid when a bid button is clicked', async () => {
    renderWithProviders(<SoloWhistPage />);
    const bidSolo = await screen.findByTestId('bid-1');
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(bidSolo);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1 }));
  });

  it('disables bids that do not beat the current highest bid', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ bids: [2, 0, 0, 0] }));
    renderWithProviders(<SoloWhistPage />);
    // Highest bid is 2 (Misère): Solo (1) and Misère (2) are disabled, Abundance (3) and Pass (0) enabled.
    await waitFor(() => expect(screen.getByTestId('bid-1')).toBeDisabled());
    expect(screen.getByTestId('bid-2')).toBeDisabled();
    expect(screen.getByTestId('bid-3')).toBeEnabled();
    expect(screen.getByTestId('bid-0')).toBeEnabled();
  });

  it('shows the current highest bid and a reason tooltip on too-low bids', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ bids: [2, 0, 0, 0] }));
    renderWithProviders(<SoloWhistPage />);
    const info = await screen.findByTestId('sw-highest-bid');
    expect(info).not.toHaveTextContent('まだ入札なし');
    // Too-low bids carry a tooltip on the wrapping span; valid bids do not.
    expect(screen.getByTestId('bid-wrap-1')).toHaveAttribute('title');
    expect(screen.getByTestId('bid-1')).toHaveAttribute('aria-label');
    expect(screen.getByTestId('bid-wrap-3')).not.toHaveAttribute('title');
  });

  it('shows "no bids yet" before anyone has bid', async () => {
    mockExec.mockResolvedValue(makeSoloWhistState({ bids: [0, 0, 0, 0] }));
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByTestId('sw-highest-bid')).toHaveTextContent('まだ入札なし'));
  });

  it('exposes the declarer line as a polite live region', async () => {
    renderWithProviders(<SoloWhistPage />);
    const line = await screen.findByTestId('solowhist-declarer');
    expect(line).toHaveAttribute('role', 'status');
    expect(line).toHaveAttribute('aria-live', 'polite');
    expect(line.className).not.toContain('animate-pulse');
  });

  it('pulses the declarer line when the contract is decided', async () => {
    // Mount with an undecided contract, then a bid resolves it (declarerIdx -1 → 0).
    mockExec.mockResolvedValue(makeSoloWhistState({ declarerIdx: -1 }));
    renderWithProviders(<SoloWhistPage />);
    const bidSolo = await screen.findByTestId('bid-1');
    mockExec.mockResolvedValue(makeSoloWhistState({ declarerIdx: 0, contract: 1, isHumanBidTurn: false }));
    fireEvent.click(bidSolo);
    await waitFor(() => expect(screen.getByTestId('solowhist-declarer').className).toContain('animate-pulse'));
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('宣言者')).toBeInTheDocument();
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('bid-0')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SoloWhistPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });
});
