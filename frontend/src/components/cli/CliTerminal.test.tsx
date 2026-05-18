import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CliLogEntry } from '../../utils/cli/types';
import { CliTerminal } from './CliTerminal';

let nextId = 0;
function makeEntry(type: CliLogEntry['type'], text: string): CliLogEntry {
  return { type, text, id: ++nextId };
}

describe('CliTerminal', () => {
  it('renders log entries', () => {
    const entries: CliLogEntry[] = [
      makeEntry('input', 'hit'),
      makeEntry('output', 'score: 21'),
      makeEntry('error', 'Unknown command'),
    ];
    render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
    // Scope to the visual log so we don't also match the sr-only live
    // region that echoes the latest announceable entry (issue #1843).
    const log = within(screen.getByRole('log'));
    expect(log.getByText('> hit')).toBeInTheDocument();
    expect(log.getByText('score: 21')).toBeInTheDocument();
    expect(log.getByText('Unknown command')).toBeInTheDocument();
  });

  it('submits command on Enter', () => {
    const onCommand = vi.fn();
    render(<CliTerminal logEntries={[]} onCommand={onCommand} disabled={false} />);
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'hit' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onCommand).toHaveBeenCalledWith('hit');
  });

  it('clears input after submit', () => {
    render(<CliTerminal logEntries={[]} onCommand={vi.fn()} disabled={false} />);
    const input = screen.getByRole('textbox') as HTMLInputElement;
    fireEvent.change(input, { target: { value: 'hit' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(input.value).toBe('');
  });

  it('does not submit empty input', () => {
    const onCommand = vi.fn();
    render(<CliTerminal logEntries={[]} onCommand={onCommand} disabled={false} />);
    const input = screen.getByRole('textbox');
    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onCommand).not.toHaveBeenCalled();
  });

  it('disables input when disabled prop is true', () => {
    render(<CliTerminal logEntries={[]} onCommand={vi.fn()} disabled={true} />);
    expect(screen.getByRole('textbox')).toBeDisabled();
  });

  it('navigates command history with ArrowUp', () => {
    const onCommand = vi.fn();
    render(<CliTerminal logEntries={[]} onCommand={onCommand} disabled={false} />);
    const input = screen.getByRole('textbox') as HTMLInputElement;

    // Submit two commands
    fireEvent.change(input, { target: { value: 'hit' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    fireEvent.change(input, { target: { value: 'stand' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    // ArrowUp should recall last command
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    expect(input.value).toBe('stand');

    // ArrowUp again should recall first command
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    expect(input.value).toBe('hit');
  });

  it('navigates command history with ArrowDown', () => {
    const onCommand = vi.fn();
    render(<CliTerminal logEntries={[]} onCommand={onCommand} disabled={false} />);
    const input = screen.getByRole('textbox') as HTMLInputElement;

    fireEvent.change(input, { target: { value: 'hit' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    fireEvent.change(input, { target: { value: 'stand' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    // Go up twice
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    fireEvent.keyDown(input, { key: 'ArrowUp' });
    expect(input.value).toBe('hit');

    // Go down
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    expect(input.value).toBe('stand');

    // Go down again to clear
    fireEvent.keyDown(input, { key: 'ArrowDown' });
    expect(input.value).toBe('');
  });

  it('applies input style for input entries', () => {
    const entries: CliLogEntry[] = [makeEntry('input', 'hit')];
    render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
    const el = screen.getByText('> hit');
    expect(el.className).toContain('text-ds-success');
  });

  it('applies error style for error entries', () => {
    const entries: CliLogEntry[] = [makeEntry('error', 'fail')];
    render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
    const log = within(screen.getByRole('log'));
    const el = log.getByText('fail');
    expect(el.className).toContain('text-ds-error');
  });

  describe('a11y announcement policy (issue #1843)', () => {
    it('sets aria-live="off" on the log container to override role="log" implicit polite default', () => {
      render(<CliTerminal logEntries={[]} onCommand={vi.fn()} disabled={false} />);
      // role="log" has an implicit aria-live="polite" per WAI-ARIA; we must
      // explicitly set "off" so SR users hop in via focus instead of being
      // flooded with every entry during normal play (issue #1843).
      const log = screen.getByRole('log');
      expect(log).toHaveAttribute('aria-live', 'off');
      expect(log).toHaveAttribute('tabindex', '0');
    });

    it('announces only the latest error in the live region', () => {
      const entries: CliLogEntry[] = [
        makeEntry('output', 'score: 17'),
        makeEntry('input', 'hit'),
        makeEntry('output', 'score: 24'),
        makeEntry('error', 'BUST'),
      ];
      const { container } = render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
      const liveRegion = container.querySelector('[aria-live="polite"]');
      expect(liveRegion?.textContent).toBe('BUST');
    });

    it('announces only the last line of the latest output (state summary)', () => {
      const entries: CliLogEntry[] = [makeEntry('output', 'Player: 7H KC (17)\nDealer: 5D ?\nYour move: hit / stand')];
      const { container } = render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
      const liveRegion = container.querySelector('[aria-live="polite"]');
      expect(liveRegion?.textContent).toBe('Your move: hit / stand');
    });

    it('announces the last non-empty line when output has a trailing newline', () => {
      const entries: CliLogEntry[] = [makeEntry('output', 'Player: 7H KC (17)\nYour move: hit / stand\n')];
      const { container } = render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
      const liveRegion = container.querySelector('[aria-live="polite"]');
      expect(liveRegion?.textContent).toBe('Your move: hit / stand');
    });

    it('does not announce input entries (user-typed echoes are noise)', () => {
      const entries: CliLogEntry[] = [makeEntry('output', 'score: 17'), makeEntry('input', 'hit')];
      const { container } = render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
      const liveRegion = container.querySelector('[aria-live="polite"]');
      // The output behind the input is the most recent announceable entry.
      expect(liveRegion?.textContent).toBe('score: 17');
    });

    it('leaves the live region empty when no announceable entries exist', () => {
      const entries: CliLogEntry[] = [makeEntry('input', 'hit')];
      const { container } = render(<CliTerminal logEntries={entries} onCommand={vi.fn()} disabled={false} />);
      const liveRegion = container.querySelector('[aria-live="polite"]');
      expect(liveRegion?.textContent).toBe('');
    });
  });
});
