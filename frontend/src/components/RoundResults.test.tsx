import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RoundResults } from './RoundResults';

const players = [{ isHuman: true }, { isHuman: false }, { isHuman: false }];

function visible() {
  return within(screen.getByTestId('round-results-visible'));
}

function liveRegion() {
  return screen.getByRole('status');
}

describe('RoundResults', () => {
  it('returns null when results is undefined', () => {
    const { container } = render(<RoundResults results={undefined} players={players} />);
    expect(container.firstChild).toBeNull();
  });

  it('returns null when results is empty', () => {
    const { container } = render(<RoundResults results={[]} players={players} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders human player as あなた', () => {
    const results = [{ playerIdx: 0, handName: 'フルハウス', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().getByText('結果:')).toBeInTheDocument();
    expect(visible().getByText(/あなた/)).toBeInTheDocument();
    expect(visible().getByText(/フルハウス/)).toBeInTheDocument();
    expect(visible().getByText(/\+100チップ/)).toBeInTheDocument();
  });

  it('renders CPU player with index', () => {
    const results = [{ playerIdx: 1, handName: 'ワンペア', wonAmount: 0 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().getByText(/CPU 1/)).toBeInTheDocument();
    expect(visible().getByText(/ワンペア/)).toBeInTheDocument();
  });

  it('does not render handName when empty', () => {
    const results = [{ playerIdx: 0, handName: '', wonAmount: 50 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().getByText(/あなた/)).toBeInTheDocument();
    expect(visible().getByText(/\+50チップ/)).toBeInTheDocument();
  });

  it('does not render wonAmount when zero', () => {
    const results = [{ playerIdx: 1, handName: 'ハイカード', wonAmount: 0 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().queryByText(/チップ/)).not.toBeInTheDocument();
  });

  it('renders kickers when present', () => {
    const results = [{ playerIdx: 0, handName: 'One Pair', kickers: 'A, Q, 10', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().getByText(/キッカー: A, Q, 10/)).toBeInTheDocument();
  });

  it('does not render kickers when absent', () => {
    const results = [{ playerIdx: 0, handName: 'Flush', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().queryByText(/キッカー/)).not.toBeInTheDocument();
  });

  it('does not render kickers when empty string', () => {
    const results = [{ playerIdx: 0, handName: 'Flush', kickers: '', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().queryByText(/キッカー/)).not.toBeInTheDocument();
  });

  it('renders "マック" when mucked is true', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().getByText(/マック/)).toBeInTheDocument();
  });

  it('does not render handName when mucked is true', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().queryByText(/ワンペア/)).not.toBeInTheDocument();
  });

  it('does not render kickers when mucked is true', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().queryByText(/キッカー/)).not.toBeInTheDocument();
  });

  it('renders handName and kickers when mucked is false', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 100, mucked: false }];
    render(<RoundResults results={results} players={players} />);
    expect(visible().getByText(/ワンペア/)).toBeInTheDocument();
    expect(visible().getByText(/キッカー: A, Q/)).toBeInTheDocument();
    expect(visible().queryByText(/マック/)).not.toBeInTheDocument();
  });

  describe('sr-only live region', () => {
    it('renders polite status live region when results are present', () => {
      const results = [{ playerIdx: 0, handName: 'フルハウス', wonAmount: 100 }];
      render(<RoundResults results={results} players={players} />);
      const region = liveRegion();
      expect(region).toHaveAttribute('aria-live', 'polite');
      expect(region).toHaveAttribute('aria-atomic', 'true');
      expect(region.className).toContain('sr-only');
    });

    it('announces hand with chips won for human player', () => {
      const results = [{ playerIdx: 0, handName: 'フルハウス', wonAmount: 100 }];
      render(<RoundResults results={results} players={players} />);
      expect(liveRegion()).toHaveTextContent('ショウダウン結果。あなた: フルハウス、+100チップ');
    });

    it('announces hand with kickers and chips won', () => {
      const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 50 }];
      render(<RoundResults results={results} players={players} />);
      expect(liveRegion()).toHaveTextContent('ショウダウン結果。あなた: ワンペア、キッカー A, Q、+50チップ');
    });

    it('announces hand with kickers but no chips won', () => {
      const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0 }];
      render(<RoundResults results={results} players={players} />);
      expect(liveRegion()).toHaveTextContent('ショウダウン結果。あなた: ワンペア、キッカー A, Q');
    });

    it('announces hand alone when no kickers and no chips won', () => {
      const results = [{ playerIdx: 1, handName: 'ハイカード', wonAmount: 0 }];
      render(<RoundResults results={results} players={players} />);
      expect(liveRegion()).toHaveTextContent('ショウダウン結果。CPU 1: ハイカード');
    });

    it('announces mucked entries without hand name or kickers', () => {
      const results = [{ playerIdx: 1, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
      render(<RoundResults results={results} players={players} />);
      expect(liveRegion()).toHaveTextContent('ショウダウン結果。CPU 1: マック');
    });

    it('joins multiple entries with comma separator', () => {
      const results = [
        { playerIdx: 0, handName: 'フルハウス', wonAmount: 100 },
        { playerIdx: 1, handName: 'ワンペア', wonAmount: 0 },
        { playerIdx: 2, handName: 'フラッシュ', wonAmount: 0, mucked: true },
      ];
      render(<RoundResults results={results} players={players} />);
      expect(liveRegion()).toHaveTextContent(
        'ショウダウン結果。あなた: フルハウス、+100チップ, CPU 1: ワンペア, CPU 2: マック',
      );
    });
  });
});
