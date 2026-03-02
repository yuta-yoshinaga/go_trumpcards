import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
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
});
