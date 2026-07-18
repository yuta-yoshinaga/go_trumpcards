import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kempsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { KempsPlayer, KempsResponse } from '../types/card';
import { KempsPage } from './KempsPage';

vi.mock('../api/gameApi', () => ({
  kempsApi: { exec: vi.fn() },
  actionLogApi: { kemps: vi.fn() },
}));

const mockExec = vi.mocked(kempsApi.exec);

function makePlayer(overrides: Partial<KempsPlayer> = {}): KempsPlayer {
  return {
    name: 'CPU',
    isHuman: false,
    team: 1,
    handSize: 4,
    hand: [],
    hasFourOfAKind: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<KempsResponse> = {}): KempsResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerTeam: -1,
    currentPlayerIdx: 0,
    isHumanTurn: true,
    teamScores: [0, 0],
    field: [
      { design: 'SPADE', value: 5 },
      { design: 'HEART', value: 6 },
      { design: 'CLOVER', value: 7 },
      { design: 'DIAMOND', value: 8 },
    ],
    signalType: 0,
    partnerSignaling: false,
    opponentSignaling: false,
    fourHolderIdx: -1,
    roundResult: 0,
    roundWinnerTeam: -1,
    roundNumber: 1,
    cpuDifficulty: 1,
    targetScore: 5,
    message: '',
    players: [
      makePlayer({
        name: 'You',
        isHuman: true,
        team: 0,
        hand: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 2 },
          { design: 'CLOVER', value: 3 },
          { design: 'DIAMOND', value: 4 },
        ],
      }),
      makePlayer({ team: 1 }),
      makePlayer({ team: 0 }),
      makePlayer({ team: 1 }),
    ],
    ...overrides,
  };
}

const exchangeState = makeState();
// Field shares rank 2 with the human's ♥2 hand card at indices 0 and 2 (#3554).
const rankMatchState = makeState({
  field: [
    { design: 'SPADE', value: 2 },
    { design: 'HEART', value: 6 },
    { design: 'CLOVER', value: 2 },
    { design: 'DIAMOND', value: 8 },
  ],
});
const declareState = makeState({ phase: 1, isHumanTurn: false, fourHolderIdx: 2 });
const roundEndState = makeState({ phase: 2, isHumanTurn: false, roundResult: 1, roundWinnerTeam: 0 });
const gameEndState = makeState({ phase: 3, gameEndFlag: true, winnerTeam: 0, isHumanTurn: false, teamScores: [5, 2] });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(exchangeState);
});

describe('KempsPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KempsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<KempsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows round, team scores and target', async () => {
    renderWithProviders(<KempsPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド 1/)).toBeInTheDocument());
    expect(screen.getByText(/目標: 5点/)).toBeInTheDocument();
  });

  it('selecting a hand card then a field card dispatches swap', async () => {
    renderWithProviders(<KempsPage />);
    // Hand[1] is ♥2, field[2] is ♣7 — buttons are now named by their card.
    const heart2 = await screen.findByRole('button', { name: '♥ 2 を選択' });
    expect(heart2).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(heart2);
    expect(screen.getByRole('button', { name: '♥ 2 を選択' })).toHaveAttribute('aria-pressed', 'true');
    const clover7 = await screen.findByRole('button', { name: '♣ 7 と交換' });
    // The swap-pending state is described for SR.
    expect(clover7).toHaveAttribute('aria-describedby', 'kemps-swap-pending');
    mockExec.mockClear();
    fireEvent.click(clover7);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('swap', { handIndex: 1, fieldIndex: 2 }));
  });

  it('highlights same-rank field cards when a hand card is selected and clears on deselect', async () => {
    mockExec.mockResolvedValue(rankMatchState);
    renderWithProviders(<KempsPage />);
    // Select ♥2 — field indices 0 and 2 (both rank 2) should be marked as matches.
    const heart2 = await screen.findByRole('button', { name: '♥ 2 を選択' });
    fireEvent.click(heart2);
    expect(screen.getByTestId('kemps-field-0')).toHaveAttribute('data-rank-match', 'true');
    expect(screen.getByTestId('kemps-field-2')).toHaveAttribute('data-rank-match', 'true');
    // The non-matching field cards are not highlighted.
    expect(screen.getByTestId('kemps-field-1')).toHaveAttribute('data-rank-match', 'false');
    expect(screen.getByTestId('kemps-field-3')).toHaveAttribute('data-rank-match', 'false');
    // Deselecting the hand card removes the field swap buttons (and their highlight).
    fireEvent.click(screen.getByRole('button', { name: '♥ 2 を選択' }));
    expect(screen.queryByTestId('kemps-field-0')).not.toBeInTheDocument();
  });

  it('does not highlight field cards when no hand rank matches the selection', async () => {
    // Default field [♠5 ♥6 ♣7 ♦8] shares no rank with the ♠1 hand card.
    renderWithProviders(<KempsPage />);
    const spadeAce = await screen.findByRole('button', { name: '♠ A を選択' });
    fireEvent.click(spadeAce);
    for (const i of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`kemps-field-${i}`)).toHaveAttribute('data-rank-match', 'false');
    }
  });

  it('does not show field swap buttons until a hand card is selected', async () => {
    renderWithProviders(<KempsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /と交換$/ })).not.toBeInTheDocument();
  });

  it('dispatches pass on the exchange turn', async () => {
    renderWithProviders(<KempsPage />);
    const btn = await screen.findByRole('button', { name: 'パス（交換しない）' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('dispatches signal when a signal option is clicked', async () => {
    renderWithProviders(<KempsPage />);
    const btn = await screen.findByRole('button', { name: '瞬き' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('signal', { signalType: 1 }));
  });

  it('shows the Kemps button in the declare window and dispatches kemps', async () => {
    mockExec.mockResolvedValue(declareState);
    renderWithProviders(<KempsPage />);
    const btn = await screen.findByRole('button', { name: 'ケムプス！' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('kemps'));
  });

  it('dispatches counter against an opponent seat', async () => {
    mockExec.mockResolvedValue(declareState);
    renderWithProviders(<KempsPage />);
    const btns = await screen.findAllByRole('button', { name: /カウンター・ケムプス/ });
    mockExec.mockClear();
    fireEvent.click(btns[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('counter', { targetSeat: 1 }));
  });

  it('shows the partner signaling cue in the declare window as a live region', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, isHumanTurn: false, partnerSignaling: true, fourHolderIdx: 2 }));
    renderWithProviders(<KempsPage />);
    const cue = await screen.findByTestId('kemps-partner-signaling');
    expect(cue).toHaveAttribute('role', 'status');
    expect(cue).toHaveAttribute('aria-live', 'polite');
    expect(cue).toHaveTextContent(/パートナーが合図/);
  });

  it('shows the opponent signaling cue in the declare window as a live region', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, isHumanTurn: false, opponentSignaling: true, fourHolderIdx: 1 }));
    renderWithProviders(<KempsPage />);
    const cue = await screen.findByTestId('kemps-opponent-signaling');
    expect(cue).toHaveAttribute('role', 'status');
    expect(cue).toHaveTextContent(/合図しているかもしれません/);
  });

  it('announces the chosen secret signal in a live region', async () => {
    mockExec.mockResolvedValue(makeState({ signalType: 1 })); // blink
    renderWithProviders(<KempsPage />);
    const sent = await screen.findByTestId('kemps-signal-sent');
    expect(sent).toHaveAttribute('role', 'status');
    expect(sent).toHaveTextContent('シグナル設定: 瞬き');
  });

  it('shows the next round button at round end and dispatches next', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KempsPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows the win message when the human team wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KempsPage />);
    await waitFor(() => expect(screen.getByText('あなたのチームの勝利です！')).toBeInTheDocument());
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<KempsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<KempsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });
});
