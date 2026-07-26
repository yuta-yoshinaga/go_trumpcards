import { fireEvent, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { durakApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DurakConfig, DurakResponse } from '../types/card';
import { DurakPage } from './DurakPage';

vi.mock('../api/gameApi', () => ({
  durakApi: { exec: vi.fn() },
  actionLogApi: { durak: vi.fn() },
}));

// Mock the sound provider so we can assert playback without loading audio. The
// page renders AnimatedCard, which calls useOptionalSound, so both hooks are stubbed.
const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(durakApi.exec);

const defaultConfig: DurakConfig = {
  playerCount: 4,
  cpuDifficulty: 0,
  transferEnabled: false,
};

const baseState: DurakResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      isFinished: false,
      cardCount: 6,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
    },
    { id: 1, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
    { id: 2, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
    { id: 3, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
  ],
  currentTurn: 0,
  phase: 0,
  attackerIdx: 0,
  defenderIdx: 1,
  tablePairs: [],
  trumpSuit: 'HEART',
  trumpCard: { design: 'HEART', value: 6 },
  stockCount: 12,
  loserIdx: -1,
  gameEndFlag: false,
  config: defaultConfig,
  cpuActions: [],
  humanAction: null,
  boutNumber: 1,
  sortMode: 0,
  message: '',
};

const defendPhaseState: DurakResponse = {
  ...baseState,
  phase: 1,
  attackerIdx: 1,
  defenderIdx: 0,
  currentTurn: 0,
  tablePairs: [{ attack: { design: 'CLOVER', value: 7 }, defense: null }],
};

const gameEndState: DurakResponse = {
  ...baseState,
  phase: 3,
  gameEndFlag: true,
  loserIdx: 1,
  message: 'ゲーム終了！',
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
  mockPlaySound.mockClear();
});

afterEach(() => {
  localStorage.clear();
});

describe('DurakPage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<DurakPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, defaultConfig));
  });

  it('announces CPU actions in a polite live region', async () => {
    mockExec.mockResolvedValue({
      ...baseState,
      cpuActions: [{ playerIdx: 1, actionType: 0, card: null, attackIdx: -1 }],
    });
    renderWithProviders(<DurakPage />);
    const log = await screen.findByTestId('durak-cpu-actions');
    expect(log).toHaveAttribute('aria-live', 'polite');
    expect(log).toHaveAttribute('role', 'status');
  });

  it('renders CPU player areas', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => {
      expect(screen.getByText('CPU 1')).toBeInTheDocument();
      expect(screen.getByText('CPU 2')).toBeInTheDocument();
      expect(screen.getByText('CPU 3')).toBeInTheDocument();
    });
  });

  it('renders trump info', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByText(/切り札/)).toBeInTheDocument());
  });

  it('shows human player hand cards', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('labels hand card buttons with cardAlt and tracks selection via aria-pressed', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A' })).toBeInTheDocument());

    const aceBtn = screen.getByRole('button', { name: '♠ A' });
    const jackBtn = screen.getByRole('button', { name: '♥ J' });
    expect(aceBtn).toHaveAttribute('aria-pressed', 'false');
    expect(jackBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(aceBtn);
    expect(aceBtn).toHaveAttribute('aria-pressed', 'true');
    expect(jackBtn).toHaveAttribute('aria-pressed', 'false');

    // Clicking again deselects.
    fireEvent.click(aceBtn);
    expect(aceBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('shows attack button when human is attacker in attack phase', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '攻撃' })).toBeInTheDocument());
  });

  it('shows pass button when table has pairs and human is attacker', async () => {
    const stateWithPairs: DurakResponse = {
      ...baseState,
      tablePairs: [{ attack: { design: 'SPADE', value: 7 }, defense: { design: 'HEART', value: 8 } }],
    };
    mockExec.mockResolvedValue(stateWithPairs);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());
  });

  it('shows defend and take buttons in defend phase', async () => {
    mockExec.mockResolvedValue(defendPhaseState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '防御' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '引き取り' })).toBeInTheDocument();
    });
  });

  it('calls attack on attack button click', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    // Select a card first
    fireEvent.click(screen.getByAltText('♠ A'));
    // Click attack
    fireEvent.click(screen.getByRole('button', { name: '攻撃' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('attack', 0));
  });

  it('calls pass on pass button click', async () => {
    const stateWithPairs: DurakResponse = {
      ...baseState,
      tablePairs: [{ attack: { design: 'SPADE', value: 7 }, defense: { design: 'HEART', value: 8 } }],
    };
    mockExec.mockResolvedValue(stateWithPairs);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('shows game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
  });

  it('hides attack/defend buttons at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '攻撃' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '防御' })).not.toBeInTheDocument();
  });

  it('shows sort buttons when game is not ended', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'スート順' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '値順' })).toBeInTheDocument();
    });
  });

  it('calls sort on sort button click', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '値順' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '値順' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('sort', undefined, undefined, undefined, 1));
  });

  it('does not show win celebration when human is the loser', async () => {
    const humanLosesState: DurakResponse = {
      ...gameEndState,
      loserIdx: 0,
    };
    mockExec.mockResolvedValue(humanLosesState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('does not show win celebration on draw (loserIdx -1)', async () => {
    const drawState: DurakResponse = {
      ...gameEndState,
      loserIdx: -1,
    };
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！')).toBeInTheDocument());
    expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
  });

  it('shows CLI toggle button', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /CLI/i })).toBeInTheDocument());
  });

  it('switches to CLI terminal on toggle click', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => screen.getByRole('button', { name: /CLI/i }));
    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    expect(screen.queryByRole('button', { name: '攻撃' })).not.toBeInTheDocument();
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '攻撃' })).toBeInTheDocument());
    // The SettingsPanel renders the hint checkbox with id="frontendHint"
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled on attack turn', async () => {
    localStorage.setItem('hint_enabled_durak', 'true');
    // baseState: human is attacker (attackerIdx=0), phase=0 (ATTACK), hand has non-trump cards
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('marks an undefended attack and pulses when human is defender', async () => {
    mockExec.mockResolvedValue(defendPhaseState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByTestId('dk-pair-0')).toBeInTheDocument());
    const attack = screen.getByTestId('dk-attack-0');
    expect(attack).toHaveAttribute('data-undefended', 'true');
    expect(attack.className).toContain('animate-pulse');
    expect(screen.queryByTestId('dk-defense-0')).not.toBeInTheDocument();
    // ARIA label includes suit + rank so screen readers convey "Undefended attack: ♣ 7".
    expect(screen.getByTestId('dk-pair-0')).toHaveAttribute('aria-label', '未防御の攻撃: ♣ 7');
  });

  it('does not pulse an undefended attack when the human is not the defender', async () => {
    // baseState: human is attacker (defenderIdx=1), so undefended attacks should NOT pulse.
    mockExec.mockResolvedValue({
      ...baseState,
      tablePairs: [{ attack: { design: 'CLOVER', value: 7 }, defense: null }],
    });
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByTestId('dk-attack-0')).toBeInTheDocument());
    const attack = screen.getByTestId('dk-attack-0');
    expect(attack).toHaveAttribute('data-undefended', 'true');
    expect(attack.className).not.toContain('animate-pulse');
  });

  it('renders a stacked defense card and no pulse when the pair is defended', async () => {
    mockExec.mockResolvedValue({
      ...defendPhaseState,
      defenderIdx: 1,
      tablePairs: [{ attack: { design: 'CLOVER', value: 7 }, defense: { design: 'HEART', value: 9 } }],
    });
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByTestId('dk-pair-0')).toBeInTheDocument());
    expect(screen.getByTestId('dk-defense-0')).toBeInTheDocument();
    const attack = screen.getByTestId('dk-attack-0');
    expect(attack).toHaveAttribute('data-undefended', 'false');
    expect(attack.className).not.toContain('animate-pulse');
    // Defended ARIA label names both attack and defense cards.
    expect(screen.getByTestId('dk-pair-0')).toHaveAttribute('aria-label', '防御済みペア: ♣ 7 を ♥ 9 で防御');
  });

  it('plays a card-place sound when attacking', async () => {
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A'));
    fireEvent.click(screen.getByRole('button', { name: '攻撃' }));
    // The central tap plays after the exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('plays a card-place sound when defending', async () => {
    mockExec.mockResolvedValue(defendPhaseState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '防御' })).toBeInTheDocument());
    // Select a hand card and a table pair so the defend button is enabled.
    fireEvent.click(screen.getByAltText('♠ A'));
    fireEvent.click(screen.getByTestId('dk-pair-0'));
    fireEvent.click(screen.getByRole('button', { name: '防御' }));
    // The central tap plays after the exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('plays an error buzz when taking the table (disadvantageous action)', async () => {
    mockExec.mockResolvedValue(defendPhaseState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引き取り' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '引き取り' }));
    expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz');
  });

  it('plays a win fanfare when the human wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByTestId('win-celebration')).toBeInTheDocument());
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('winFanfare'));
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<DurakPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, defaultConfig));
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });
});
