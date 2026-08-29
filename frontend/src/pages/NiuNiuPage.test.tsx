import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { niuniuApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, NiuNiuHand, NiuNiuResponse } from '../types/card';
import { NiuNiuPage } from './NiuNiuPage';

vi.mock('../api/gameApi', () => ({
  niuniuApi: { exec: vi.fn() },
  actionLogApi: { niuniu: vi.fn() },
}));

const mockExec = vi.mocked(niuniuApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function hand(overrides?: Partial<NiuNiuHand>): NiuNiuHand {
  return {
    cards: [card('SPADE', 10), card('HEART', 10), card('CLOVER', 10), card('DIAMOND', 5), card('SPADE', 5)],
    bet: 100,
    comboIdx: [0, 1, 2],
    rank: 10,
    rankKey: 'niuniu',
    multiplier: 3,
    payout: 0,
    hidden: false,
    ...overrides,
  };
}

const hiddenHand = (bet = 20) =>
  hand({ hidden: true, cards: [null, null, null, null, null], rankKey: '', comboIdx: [], multiplier: 0, bet });

function makeState(overrides?: Partial<NiuNiuResponse>): NiuNiuResponse {
  return {
    seats: [
      { name: 'あなた', isCpu: false, hand: hand() },
      { name: 'CPU1', isCpu: true, hand: hiddenHand() },
      { name: 'CPU2', isCpu: true },
      { name: '親', isCpu: true },
    ],
    bankerHand: hiddenHand(0),
    bankerIdx: 3,
    chips: 900,
    maxMultiplier: 3,
    bankerRankKey: '',
    phase: 1,
    message: '',
    ...overrides,
  };
}

describe('NiuNiuPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders chips', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());
  });

  // The server withholds a hidden hand's cards; the page renders backs from
  // `hidden` rather than deciding for itself what may be seen.
  it('renders a hand the server marked hidden as backs', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByLabelText('親の手は伏せられています')).toBeInTheDocument());
    expect(screen.getByLabelText('CPU1 の手は伏せられています')).toBeInTheDocument();
  });

  it('reveals a hand the server did not mark hidden', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByLabelText(/あなた の手 牛牛/)).toBeInTheDocument());
  });

  it('shows the rank and the multiplier', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText('牛牛')).toBeInTheDocument());
    expect(screen.getByText('x3')).toBeInTheDocument();
  });

  it('omits the multiplier at even money', async () => {
    mockExec.mockResolvedValue(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ rank: 3, rankKey: 'n3', multiplier: 1 }) }],
        bankerHand: undefined,
      }),
    );
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText('牛3')).toBeInTheDocument());
    expect(screen.queryByText('x1')).not.toBeInTheDocument();
  });

  it('offers the bet buttons and dispatches the stake', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    const btn = await screen.findByRole('button', { name: '100' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });

  // The cap is the WORST CASE, not the stake: a banker's Niu Niu takes three
  // times what you put down, so 50 chips cannot cover a stake of 50.
  it('disables a stake the stack cannot cover three times over', async () => {
    mockExec.mockResolvedValue(makeState({ chips: 50 }));
    renderWithProviders(<NiuNiuPage />);
    // 10 x 3 = 30 <= 50, so the minimum is playable.
    await waitFor(() => expect(screen.getByRole('button', { name: '10' })).toBeEnabled());
    // 50 x 3 = 150 > 50 -- a stake equal to the whole stack is NOT affordable.
    expect(screen.getByRole('button', { name: '50' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '500' })).toBeDisabled();
  });

  // The round settles at the bet, so the stake buttons go away entirely.
  it('hides the stake buttons once the round has settled', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2, bankerRankKey: 'niuniu' }));
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: '100' })).not.toBeInTheDocument());
  });

  // The page renders the rank from the server's key, not from a display string
  // the server picked -- an English-locale player used to see 牛牛 here (#5567).
  it('names the rank on a revealed hand', async () => {
    mockExec.mockResolvedValue(
      makeState({
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ rank: 3, rankKey: 'n3', multiplier: 1 }) }],
        bankerHand: undefined,
      }),
    );
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText('牛3')).toBeInTheDocument());
  });

  it('shows the payout and the combo hint after the round', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        bankerRankKey: 'none',
        seats: [{ name: 'あなた', isCpu: false, hand: hand({ payout: 300 }) }],
        bankerHand: hand({ rank: 0, rankKey: 'none', multiplier: 1, comboIdx: [] }),
      }),
    );
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText(/\+300/)).toBeInTheDocument());
    expect(screen.getByText(/10の倍数/)).toBeInTheDocument();
  });

  it('skips a seat with no hand', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());
    expect(screen.queryByText('CPU2')).not.toBeInTheDocument();
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 900/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByRole('button', { name: '100' })).not.toBeInTheDocument());
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });

  it('says why a stake is refused rather than only greying it out', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // 100 chips against a 5x ceiling: only the smallest stakes are affordable.
    mockExec.mockResolvedValue(makeState({ chips: 100, maxMultiplier: 5 }));
    renderWithProviders(<NiuNiuPage />);
    await waitFor(() => expect(screen.getAllByRole('button').length).toBeGreaterThan(0));
    const refused = screen.getAllByRole('button').filter((b) => b.hasAttribute('disabled') && b.title);
    expect(refused.length).toBeGreaterThan(0);
    expect(refused[0]?.title).toMatch(/チップ/);
  });
});

// 牛を作る3枚はリング (ring-2 ring-ds-warning) でしか示されておらず、
// このページのコメント自身が「5枚と役名だけでは何も説明にならない」と
// 述べているのに、その情報が読み上げには一切乗っていなかった (#6363)。
describe('NiuNiuPage combo announcement', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('names the three cards that make the niu', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<NiuNiuPage />);

    const human = await screen.findByLabelText(/^あなた の手 /);
    const label = human.getAttribute('aria-label') ?? '';
    // fixture の comboIdx は [0,1,2] = ♠10 ♥10 ♣10。
    expect(label).toContain('牛の3枚: ♠ 10 ♥ 10 ♣ 10');
    // 役名は今までどおり残る。
    expect(label).toContain('牛牛');
    expect(label).not.toContain('{{');
    // 牛に入っていない札を数え上げていないこと。
    expect(label).not.toContain('♦ 5');
  });

  // 否定コントロール: 組が無い手では牛の断りを付けない。
  it('says nothing about a combo when there is none', async () => {
    const base = makeState();
    mockExec.mockResolvedValue({
      ...base,
      seats: base.seats.map((s, i) => (i === 0 ? { ...s, hand: hand({ comboIdx: [] }) } : s)),
    });
    renderWithProviders(<NiuNiuPage />);

    const human = await screen.findByLabelText(/^あなた の手 /);
    expect(human.getAttribute('aria-label') ?? '').not.toContain('牛の3枚');
  });
});
