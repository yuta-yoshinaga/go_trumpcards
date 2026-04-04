import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CliToggle } from './CliToggle';

describe('CliToggle', () => {
  it('shows CLI label (destination) when in GUI mode', () => {
    render(<CliToggle cliEnabled={false} onToggle={vi.fn()} />);
    expect(screen.getByText('CLI')).toBeInTheDocument();
  });

  it('shows GUI label (destination) when in CLI mode', () => {
    render(<CliToggle cliEnabled={true} onToggle={vi.fn()} />);
    expect(screen.getByText('GUI')).toBeInTheDocument();
  });

  it('calls onToggle when clicked', () => {
    const onToggle = vi.fn();
    render(<CliToggle cliEnabled={false} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onToggle).toHaveBeenCalledOnce();
  });

  it('has accessible aria-label when CLI disabled', () => {
    render(<CliToggle cliEnabled={false} onToggle={vi.fn()} />);
    expect(screen.getByRole('button')).toHaveAttribute(
      'aria-label',
      'CLI\u30e2\u30fc\u30c9\u306b\u5207\u308a\u66ff\u3048',
    );
  });

  it('has accessible aria-label when CLI enabled', () => {
    render(<CliToggle cliEnabled={true} onToggle={vi.fn()} />);
    expect(screen.getByRole('button')).toHaveAttribute(
      'aria-label',
      'GUI\u30e2\u30fc\u30c9\u306b\u5207\u308a\u66ff\u3048',
    );
  });
});
