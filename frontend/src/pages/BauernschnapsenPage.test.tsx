import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bauernschnapsenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BauernschnapsenResponse, Card } from '../types/card';
import { BauernschnapsenPhase } from '../types/phases';
import { BauernschnapsenPage } from './BauernschnapsenPage';

vi.mock('../api/gameApi', () => ({
  bauernschnapsenApi: { exec: vi.fn() },
  actionLogApi: { bauernschnapsen: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(bauernschnapsenApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<BauernschnapsenResponse> = {}): BauernschnapsenResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 13), card('SPADE', 12), card('HEART', 10)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: BauernschnapsenPhase.PLAY,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 1,
    contract: 1,
    declarerIdx: 0,
    validPlayIndices: [0, 1, 2, 3, 4],
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundMarriage: [0, 0],
    marriageIndices: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, targetScore: 24 },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: BauernschnapsenPhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [101, 60],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(initialState);
});

describe('BauernschnapsenPage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 24,
      }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so a prefixed key resolved to the literal
  // "phase.phase.play" on screen. See issue #4374.
  it('renders the translated phase name, not the raw i18n key', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('プレイ'));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('renders the team score table and the settled contract during play', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.getByTestId('bauernschnapsen-contract')).toHaveTextContent('契約: 通常');
    // **山札もめくり札もこのゲームには無い。** クローン元の遺物が残っていないこと。
    expect(screen.queryByText(/山札/)).not.toBeInTheDocument();
    expect(screen.queryByText('めくり札')).not.toBeInTheDocument();
  });

  // 契約フェーズを抜ける操作面が無いと、盤面は最初の手番で固まる。
  it('offers the contract controls on the human bid, and dispatches the declaration', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BauernschnapsenPhase.CONTRACT, contract: 0, declarerIdx: -1 }));
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByTestId('bauernschnapsen-contract-controls')).toBeInTheDocument());
    expect(screen.getByTestId('bauernschnapsen-contract')).toHaveTextContent('宣言中');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ベテル' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec).toHaveBeenCalledWith('contract', 3, undefined, undefined, 1);
  });

  it('hides the contract controls once play has started', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('bauernschnapsen-contract-controls')).not.toBeInTheDocument();
  });

  // **追従はトリック 1 から必須。** 出せない札は押せないこと。
  it('disables hand cards outside validPlayIndices', async () => {
    mockExec.mockResolvedValue(makeState({ validPlayIndices: [0] }));
    renderWithProviders(<BauernschnapsenPage />);
    const allowed = await screen.findByRole('button', { name: '♠ K' });
    expect(allowed).not.toHaveAttribute('aria-disabled', 'true');
    const blocked = screen.getByRole('button', { name: '♠ Q' });
    expect(blocked).toHaveAttribute('aria-disabled', 'true');
  });

  it('dispatches play with the selected card index during play', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('shows the marriage button when the selected card can declare and dispatches marriage', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0] }));
    renderWithProviders(<BauernschnapsenPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    const marriageBtn = await screen.findByRole('button', { name: 'マリッジ' });
    mockExec.mockClear();
    fireEvent.click(marriageBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', undefined, 0));
  });

  it('hides the marriage button when the selected card cannot declare', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ K' });
    fireEvent.click(cardBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'マリッジ' })).not.toBeInTheDocument();
  });

  it('badges both King and Queen when a marriage is available on the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0, 1] }));
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByTestId('card-role-badge-0')).toBeInTheDocument());
    expect(screen.getByTestId('card-role-badge-1')).toBeInTheDocument();
    // The non-marriage card (index 2) gets no badge.
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument();
    expect(screen.getByTestId('card-role-badge-0')).toHaveTextContent('💍');
  });

  it('shows no marriage badge when no marriage is available', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-1')).not.toBeInTheDocument();
  });

  it('shows no marriage badge when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ marriageIndices: [0, 1], currentPlayerIdx: 1 }));
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('card-role-badge-1')).not.toBeInTheDocument();
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('translates the server hint reason instead of leaking the raw code', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ hint: { cardIndex: 2, reason: 'lead_trump', isMarriage: false } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/切り札でリード/)).toBeInTheDocument());
    expect(screen.queryByText(/lead_trump/)).not.toBeInTheDocument();
  });

  it("shows '-' when the hint carries no card index", async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ hint: { reason: 'marriage', isMarriage: true } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/\[-\]/)).toBeInTheDocument());
  });

  it('falls back to the raw code for an unknown hint reason', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<BauernschnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue(makeState({ hint: { cardIndex: 0, reason: 'no_such_reason', isMarriage: false } }));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/no_such_reason/)).toBeInTheDocument());
  });
});
