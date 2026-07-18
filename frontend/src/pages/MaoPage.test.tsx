import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { maoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { MaoResponse } from '../types/card';
import { MaoPage } from './MaoPage';

vi.mock('../api/gameApi', () => ({
  maoApi: { exec: vi.fn() },
  actionLogApi: { mao: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(maoApi.exec);

const playPhaseState: MaoResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      hasDeclared: false,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 3, cumulativeScore: 10, hasDeclared: false },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 5, cumulativeScore: 20, hasDeclared: false },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 5, hasDeclared: false },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  chosenSuit: 0,
  penaltyDrawCount: 0,
  direction: 1,
  gameEndFlag: false,
  winnerIdx: -1,
  awaitingWord: false,
  correctCount: 0,
  hintUnlocked: false,
  ruleHint: '',
  rulePenalty: false,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

const chooseSuitState: MaoResponse = { ...playPhaseState, phase: 1 };
const mustDeclareState: MaoResponse = { ...playPhaseState, phase: 2 };
const roundEndState: MaoResponse = { ...playPhaseState, phase: 3 };
const gameEndState: MaoResponse = {
  ...playPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
};
const penaltyState: MaoResponse = { ...playPhaseState, penaltyDrawCount: 4 };
const rulePenaltyState: MaoResponse = { ...playPhaseState, rulePenalty: true };

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('MaoPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MaoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('renders human cards', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
      expect(screen.getByAltText('♥ J')).toBeInTheDocument();
    });
  });

  it('calls play when play button clicked', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    fireEvent.click(screen.getByAltText('♠ A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls draw when draw button clicked', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('buzzes and flashes the rule panel when a rule penalty lands', async () => {
    mockExec.mockResolvedValue(rulePenaltyState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('rule-penalty')).toBeInTheDocument());
    expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz');
    const panel = screen.getByTestId('mao-rule-panel');
    expect(panel.className).toContain('motion-safe:animate-pulse');
    expect(panel.className).toContain('ring-ds-error');
  });

  it('does not buzz when there is no rule penalty', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(mockPlaySound).not.toHaveBeenCalledWith('errorBuzz');
  });

  it('shows penalty banner and take-penalty draw label when penalty active', async () => {
    mockExec.mockResolvedValue(penaltyState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByText(/ドローペナルティ/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /引き受ける/ })).toBeInTheDocument();
    expect(await screen.findByTestId('penalty-badge')).toHaveTextContent('4');
  });

  it('renders choose suit phase and calls suit', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スペード' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 3));
  });

  it('shows a color-coded suit symbol alongside each suit-selection button', async () => {
    mockExec.mockResolvedValue(chooseSuitState);
    renderWithProviders(<MaoPage />);
    // CrazyEightsSuit: SPADE=1, CLOVER=2, HEART=3, DIAMOND=4
    await waitFor(() => expect(screen.getByTestId('suit-symbol-1')).toBeInTheDocument());

    // Every suit button shows both its glyph and its text label.
    expect(screen.getByTestId('suit-symbol-1')).toHaveTextContent('♠');
    expect(screen.getByTestId('suit-symbol-2')).toHaveTextContent('♣');
    expect(screen.getByTestId('suit-symbol-3')).toHaveTextContent('♥');
    expect(screen.getByTestId('suit-symbol-4')).toHaveTextContent('♦');
    expect(screen.getByRole('button', { name: 'ダイヤ' })).toBeInTheDocument();

    // Red suits (hearts, diamonds) carry the red token; black suits use the ivory primary token.
    expect(screen.getByTestId('suit-symbol-3').className).toContain('text-ds-error');
    expect(screen.getByTestId('suit-symbol-4').className).toContain('text-ds-error');
    expect(screen.getByTestId('suit-symbol-1').className).toContain('text-ds-text-primary');
    expect(screen.getByTestId('suit-symbol-2').className).toContain('text-ds-text-primary');
  });

  it('calls declare when declare button clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'マオ！' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'マオ！' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare'));
  });

  it('calls skipdeclare when skip button clicked', async () => {
    mockExec.mockResolvedValue(mustDeclareState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '宣言しない' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '宣言しない' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipdeclare'));
  });

  it('shows next round button and calls nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the hidden-rule panel with compliance progress', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, correctCount: 2 });
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(screen.getByText(/ルール遵守: 2\/3/)).toBeInTheDocument();
  });

  it('says a word via declareword and clears the input', async () => {
    renderWithProviders(<MaoPage />);
    const input = await screen.findByLabelText('唱える言葉を入力…');
    fireEvent.change(input, { target: { value: 'mao' } });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '発言する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declareword', undefined, undefined, undefined, 'mao'));
    await waitFor(() => expect((input as HTMLInputElement).value).toBe(''));
  });

  it('disables say-word button when the input is empty', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '発言する' })).toBeDisabled());
  });

  it('shows a penalty notice when rulePenalty is true', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, rulePenalty: true });
    renderWithProviders(<MaoPage />);
    expect(await screen.findByTestId('rule-penalty')).toBeInTheDocument();
  });

  it('shows awaiting-word signal when awaitingWord is true', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, awaitingWord: true });
    renderWithProviders(<MaoPage />);
    expect(await screen.findByTestId('awaiting-word')).toBeInTheDocument();
  });

  it('shows the rule hint only when hintUnlocked', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hintUnlocked: true, ruleHint: 'be polite' });
    renderWithProviders(<MaoPage />);
    expect(await screen.findByTestId('rule-hint')).toHaveTextContent('be polite');
  });

  it('hides the rule hint when not unlocked', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByTestId('mao-rule-panel')).toBeInTheDocument());
    expect(screen.queryByTestId('rule-hint')).not.toBeInTheDocument();
  });

  it('shows discard top card and round info', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => {
      expect(screen.getByText('捨て札')).toBeInTheDocument();
      expect(screen.getByAltText('♥ 7')).toBeInTheDocument();
      expect(screen.getByText('ラウンド 1')).toBeInTheDocument();
    });
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MaoPage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('棋譜を見る')).toBeInTheDocument();
    });
  });

  it('reset button calls exec with confirm', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1, pointLimit: 200 }),
    );
  });

  it('renders accessible h1 heading', async () => {
    renderWithProviders(<MaoPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });
});
