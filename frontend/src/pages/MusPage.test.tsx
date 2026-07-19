import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { musApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMusState } from '../test/stateFactories';
import { MusPage } from './MusPage';

vi.mock('../api/gameApi', () => ({
  musApi: { exec: vi.fn() },
  actionLogApi: { mus: vi.fn() },
}));

const mockExec = vi.mocked(musApi.exec);

const grandeState = makeMusState();
const musPhaseState = makeMusState({ phase: 0, musTurn: 0, isHumanTurn: true });
const discardState = makeMusState({ phase: 1, discardTurn: 0, isHumanTurn: true });
const respondState = makeMusState({
  phase: 2,
  pendingStake: 3,
  canPaso: false,
  canEnvido: false,
  canOrdago: false,
  canQuiero: true,
  canNoQuiero: true,
});
const roundEndState = makeMusState({
  phase: 7,
  isHumanTurn: false,
  results: [
    { kind: 0, stake: 2, team: 0 },
    { kind: 1, stake: 1, team: 1 },
    { kind: 2, stake: 0, team: -1 },
    { kind: 3, stake: 0, team: -1 },
  ],
});
const gameEndState = makeMusState({
  phase: 8,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ちです！',
});
const cpuTurnState = makeMusState({ isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(grandeState);
});

describe('MusPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MusPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MusPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetAmarrakos: 40 },
      }),
    );
  });

  it('renders the human cards and the amarrakos panel', async () => {
    renderWithProviders(<MusPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('アマラコ')).toBeInTheDocument();
  });

  it('renders mus / corte buttons in the mus phase and dispatches mus', async () => {
    mockExec.mockResolvedValue(musPhaseState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ムス（交換）/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /コルテ（勝負）/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mus', { mus: false }));
  });

  it('disables the discard button and shows the guide when nothing is selected', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /交換する/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /交換する/ })).toBeDisabled();
    expect(screen.getByTestId('mus-discard-guide')).toHaveTextContent('交換するカードを1〜4枚選んでください');
  });

  it('enables the discard button once a card is selected and dispatches the discard', async () => {
    mockExec.mockResolvedValue(discardState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /交換する/ })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '♠ A' }));
    expect(screen.getByRole('button', { name: /交換する/ })).toBeEnabled();
    expect(screen.getByTestId('mus-discard-guide')).toHaveTextContent('1枚選択中');

    mockExec.mockClear();
    mockExec.mockResolvedValue(discardState);
    fireEvent.click(screen.getByRole('button', { name: /交換する/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [0] }));
  });

  it('renders bet buttons gated by can flags and dispatches envido', async () => {
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パソ（パス）' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(grandeState);
    fireEvent.click(screen.getByRole('button', { name: 'エンビード（賭け）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { betAction: 1, betAmount: 2 }));
  });

  it('exposes the envido stepper to assistive tech with a group and labelled buttons', async () => {
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('group', { name: 'エンビードの賭け額' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'エンビードを1増やす' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'エンビードを1減らす' })).toBeInTheDocument();
  });

  it('shows quiero / noquiero when a bet is pending', async () => {
    mockExec.mockResolvedValue(respondState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'キエロ（受ける）' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'パソ（パス）' })).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'ノキエロ（降りる）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { betAction: 4, betAmount: 0 }));
  });

  it('renders round end with the next round button and the results', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('賭けラウンド結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ちです！')).toBeInTheDocument());
  });

  it('does not show bet buttons on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByText('アマラコ')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'パソ（パス）' })).not.toBeInTheDocument();
  });

  it('shows the hand-summary panel with Pares and Juego (default hand: par of Q, Punto 28)', async () => {
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByTestId('mus-hand-summary')).toBeInTheDocument());
    expect(screen.getByText('手札評価')).toBeInTheDocument();
    // A(1) + Q(12) + Q(12) + 7 -> pair of queens, points 28 (< 31 -> Punto).
    expect(screen.getByTestId('mus-summary-pares')).toHaveTextContent('パレス: ペア');
    expect(screen.getByTestId('mus-summary-juego')).toHaveTextContent('プント 28点');
  });

  it('highlights the Pares row during the Pares betting round', async () => {
    mockExec.mockResolvedValue(makeMusState({ phase: 4 }));
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByTestId('mus-summary-pares')).toBeInTheDocument());
    expect(screen.getByTestId('mus-summary-pares').className).toContain('font-semibold');
    expect(screen.getByTestId('mus-summary-juego').className).not.toContain('font-semibold');
  });

  it('shows the best-hand marker for a 31-point Juego', async () => {
    const juego31 = makeMusState({
      phase: 5,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 4,
          cards: [
            { design: 'SPADE', value: 13 },
            { design: 'CLOVER', value: 13 },
            { design: 'HEART', value: 13 },
            { design: 'DIAMOND', value: 1 },
          ],
          teamScore: 5,
        },
        { id: 1, isHuman: false, cardCount: 4, cards: [], teamScore: 3 },
        { id: 2, isHuman: false, cardCount: 4, cards: [], teamScore: 5 },
        { id: 3, isHuman: false, cardCount: 4, cards: [], teamScore: 3 },
      ],
    });
    mockExec.mockResolvedValue(juego31);
    renderWithProviders(<MusPage />);
    await waitFor(() => expect(screen.getByTestId('mus-summary-juego')).toBeInTheDocument());
    // Three Kings = medias; 10+10+10+1 = 31 -> best.
    expect(screen.getByTestId('mus-summary-pares')).toHaveTextContent('パレス: メディアス');
    expect(screen.getByTestId('mus-summary-juego')).toHaveTextContent('31点 ★');
    expect(screen.getByTestId('mus-summary-juego').className).toContain('font-semibold');
  });
});
