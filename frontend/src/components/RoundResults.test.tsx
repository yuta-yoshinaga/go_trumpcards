import { describe, expect, it } from 'bun:test';
import { render, screen } from '@testing-library/react';
import { RoundResults } from './RoundResults';

const players = [{ isHuman: true }, { isHuman: false }, { isHuman: false }];

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
    expect(screen.getByText('結果:')).toBeInTheDocument();
    expect(screen.getByText(/あなた/)).toBeInTheDocument();
    expect(screen.getByText(/フルハウス/)).toBeInTheDocument();
    expect(screen.getByText(/\+100チップ/)).toBeInTheDocument();
  });

  it('renders CPU player with index', () => {
    const results = [{ playerIdx: 1, handName: 'ワンペア', wonAmount: 0 }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    expect(screen.getByText(/ワンペア/)).toBeInTheDocument();
  });

  it('does not render handName when empty', () => {
    const results = [{ playerIdx: 0, handName: '', wonAmount: 50 }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.getByText(/あなた/)).toBeInTheDocument();
    expect(screen.getByText(/\+50チップ/)).toBeInTheDocument();
  });

  it('does not render wonAmount when zero', () => {
    const results = [{ playerIdx: 1, handName: 'ハイカード', wonAmount: 0 }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.queryByText(/チップ/)).not.toBeInTheDocument();
  });

  it('renders kickers when present', () => {
    const results = [{ playerIdx: 0, handName: 'One Pair', kickers: 'A, Q, 10', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.getByText(/キッカー: A, Q, 10/)).toBeInTheDocument();
  });

  it('does not render kickers when absent', () => {
    const results = [{ playerIdx: 0, handName: 'Flush', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.queryByText(/キッカー/)).not.toBeInTheDocument();
  });

  it('does not render kickers when empty string', () => {
    const results = [{ playerIdx: 0, handName: 'Flush', kickers: '', wonAmount: 100 }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.queryByText(/キッカー/)).not.toBeInTheDocument();
  });

  it('renders "マック" when mucked is true', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.getByText(/マック/)).toBeInTheDocument();
  });

  it('does not render handName when mucked is true', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.queryByText(/ワンペア/)).not.toBeInTheDocument();
  });

  it('does not render kickers when mucked is true', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 0, mucked: true }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.queryByText(/キッカー/)).not.toBeInTheDocument();
  });

  it('renders handName and kickers when mucked is false', () => {
    const results = [{ playerIdx: 0, handName: 'ワンペア', kickers: 'A, Q', wonAmount: 100, mucked: false }];
    render(<RoundResults results={results} players={players} />);
    expect(screen.getByText(/ワンペア/)).toBeInTheDocument();
    expect(screen.getByText(/キッカー: A, Q/)).toBeInTheDocument();
    expect(screen.queryByText(/マック/)).not.toBeInTheDocument();
  });
});
