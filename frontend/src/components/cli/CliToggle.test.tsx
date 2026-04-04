import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CliToggle } from './CliToggle';

describe('CliToggle', () => {
  it('renders GUI label when CLI is disabled', () => {
    render(<CliToggle cliEnabled={false} onToggle={vi.fn()} />);
    expect(screen.getByText('GUI')).toBeInTheDocument();
  });

  it('renders CLI label when CLI is enabled', () => {
    render(<CliToggle cliEnabled={true} onToggle={vi.fn()} />);
    expect(screen.getByText('CLI')).toBeInTheDocument();
  });

  it('calls onToggle when clicked', () => {
    const onToggle = vi.fn();
    render(<CliToggle cliEnabled={false} onToggle={onToggle} />);
    fireEvent.click(screen.getByRole('button'));
    expect(onToggle).toHaveBeenCalledOnce();
  });

  it('has accessible aria-label when CLI disabled', () => {
    render(<CliToggle cliEnabled={false} onToggle={vi.fn()} />);
    expect(screen.getByRole('button')).toHaveAttribute('aria-label', 'CLIモードに切り替え');
  });

  it('has accessible aria-label when CLI enabled', () => {
    render(<CliToggle cliEnabled={true} onToggle={vi.fn()} />);
    expect(screen.getByRole('button')).toHaveAttribute('aria-label', 'GUIモードに切り替え');
  });
});
