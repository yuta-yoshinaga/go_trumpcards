import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tehonbikiApi } from '../api/games/tehonbiki';
import { useGameApi } from '../hooks/useGameApi';
import i18n from '../i18n';
import { renderWithProviders } from '../test/renderWithProviders';
import type { TehonbikiResponse } from '../types/games/tehonbiki';
import { TehonbikiPage } from './TehonbikiPage';

vi.mock('../api/games/tehonbiki', () => ({ tehonbikiApi: { exec: vi.fn() } }));
vi.mock('../hooks/useGameApi', () => ({ useGameApi: vi.fn() }));
vi.mock('../components/skeleton/GameSkeleton', () => ({ GameSkeleton: () => null }));

const mockUseGameApi = vi.mocked(useGameApi);
const mockExec = vi.mocked(tehonbikiApi.exec);
const base: TehonbikiResponse = {
  phase: 0,
  numbers: [],
  betType: '',
  bet: 0,
  result: 0,
  payout: 0,
  chips: 1000,
  roundNumber: 1,
  remainingCards: 6,
  gameEndFlag: false,
  payoutNum: 9,
  payoutDen: 2,
  message: '',
};

const state = (over: Partial<TehonbikiResponse> = {}): TehonbikiResponse => ({ ...base, ...over });

beforeEach(() => {
  vi.clearAllMocks();
  mockUseGameApi.mockReturnValue({
    state: state(),
    loading: false,
    error: null,
    exec: mockExec,
    retry: vi.fn(),
  } as never);
});

describe('TehonbikiPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the six selectable numbers', () => {
    renderWithProviders(<TehonbikiPage />);
    for (let n = 1; n <= 6; n++) expect(screen.getByRole('button', { name: String(n) })).toBeInTheDocument();
  });

  it('renders the translated phase name', () => {
    renderWithProviders(<TehonbikiPage />);
    expect(screen.getByText('予想')).toBeInTheDocument();
    expect(screen.queryByText('0')).not.toBeInTheDocument();
  });

  it('renders translated wager options', () => {
    renderWithProviders(<TehonbikiPage />);
    for (const name of ['単張り', '二丁掛け', '三丁掛け', '片山']) {
      expect(screen.getByRole('option', { name })).toBeInTheDocument();
    }
    expect(screen.queryByText('single')).not.toBeInTheDocument();
  });

  it('associates translated labels with the wager type and amount inputs', async () => {
    renderWithProviders(<TehonbikiPage />);
    expect(screen.getByRole('combobox', { name: '賭け方' })).toBeInTheDocument();
    expect(screen.getByRole('spinbutton', { name: '張り金' })).toBeInTheDocument();

    await i18n.changeLanguage('en');
    expect(screen.getByRole('combobox', { name: 'Wager type' })).toBeInTheDocument();
    expect(screen.getByRole('spinbutton', { name: 'Stake' })).toBeInTheDocument();
    await i18n.changeLanguage('ja');
  });

  it('summarizes the wager and clearly shows an empty number selection', async () => {
    renderWithProviders(<TehonbikiPage />);
    const summary = await screen.findByRole('status');
    expect(summary).toHaveTextContent('数字 未選択');
    expect(summary).toHaveTextContent('賭け方 単張り');
    expect(summary).toHaveTextContent('賭け額 50');

    fireEvent.click(screen.getByRole('button', { name: '2' }));
    fireEvent.click(screen.getByRole('button', { name: '4' }));
    expect(summary).toHaveTextContent('数字 2、4');
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'double' } });
    fireEvent.change(screen.getByRole('spinbutton'), { target: { value: '' } });
    expect(summary).toHaveTextContent('数字 未選択');
    expect(summary).toHaveTextContent('賭け方 二丁掛け');
    expect(summary).toHaveTextContent('賭け額 未入力');
  });

  it('sends the selected numbers, wager kind, and amount', async () => {
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '張る' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '1' }));
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'double' } });
    fireEvent.click(screen.getByRole('button', { name: '2' }));
    fireEvent.click(screen.getByRole('button', { name: '3' }));
    fireEvent.click(screen.getByRole('button', { name: '張る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { numbers: [2, 3], betType: 'double', bet: 50 }));
  });

  it('shows the revealed parent card and next-round action after a result', async () => {
    mockUseGameApi.mockReturnValue({
      state: state({ phase: 1, parentCard: 6, numbers: [6], betType: 'single', bet: 50, result: 1, payout: 225 }),
      loading: false,
      error: null,
      exec: mockExec,
      retry: vi.fn(),
    } as never);
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(screen.getByText('親の札: 6')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '次の勝負' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '張る' })).not.toBeInTheDocument();
  });

  it('clears selected numbers when the wager kind changes', async () => {
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '張る' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '1' }));
    expect(screen.getByRole('button', { name: '1' })).toHaveAttribute('aria-pressed', 'true');
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'half' } });
    expect(screen.getByRole('button', { name: '1' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('sends an incomplete selection when the user submits it', async () => {
    renderWithProviders(<TehonbikiPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '張る' })).toBeInTheDocument());
    fireEvent.change(screen.getByRole('combobox'), { target: { value: 'double' } });
    fireEvent.click(screen.getByRole('button', { name: '1' }));
    fireEvent.click(screen.getByRole('button', { name: '張る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { numbers: [1], betType: 'double', bet: 50 }));
  });

  it('starts the next round from the result phase', async () => {
    mockUseGameApi.mockReturnValue({
      state: state({ phase: 1, parentCard: 2, numbers: [2], betType: 'single', result: 1, payout: 225 }),
      loading: false,
      error: null,
      exec: mockExec,
      retry: vi.fn(),
    } as never);
    renderWithProviders(<TehonbikiPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次の勝負' }));
    expect(mockExec).toHaveBeenCalledWith('next');
  });

  it('confirms a reset and resets at game end', async () => {
    renderWithProviders(<TehonbikiPage />);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    expect(mockExec).toHaveBeenCalledWith('reset');
  });

  it('resets immediately from the game-end phase', () => {
    cleanup();
    mockUseGameApi.mockReturnValue({
      state: state({ phase: 2, gameEndFlag: true }),
      loading: false,
      error: null,
      exec: mockExec,
      retry: vi.fn(),
    } as never);
    renderWithProviders(<TehonbikiPage />);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(mockExec).toHaveBeenCalledWith('reset');
  });
});
