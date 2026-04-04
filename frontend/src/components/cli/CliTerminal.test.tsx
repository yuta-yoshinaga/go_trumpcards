import { fireEvent, render, screen } from '@testing-library/react';
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
    expect(screen.getByText('> hit')).toBeInTheDocument();
    expect(screen.getByText('score: 21')).toBeInTheDocument();
    expect(screen.getByText('Unknown command')).toBeInTheDocument();
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
    const el = screen.getByText('fail');
    expect(el.className).toContain('text-ds-error');
  });
});
