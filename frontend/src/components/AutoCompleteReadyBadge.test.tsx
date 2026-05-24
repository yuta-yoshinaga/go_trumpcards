import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AutoCompleteReadyBadge } from './AutoCompleteReadyBadge';

describe('AutoCompleteReadyBadge', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders nothing when ready is false', () => {
    render(<AutoCompleteReadyBadge ready={false} />);
    expect(screen.queryByTestId('autocomplete-ready-badge')).not.toBeInTheDocument();
  });

  it('appears on initial mount when ready=true (e.g., resuming a ready game)', () => {
    render(<AutoCompleteReadyBadge ready={true} />);
    expect(screen.getByTestId('autocomplete-ready-badge')).toBeInTheDocument();
  });

  it('appears when ready transitions from false to true', () => {
    const { rerender } = render(<AutoCompleteReadyBadge ready={false} />);
    expect(screen.queryByTestId('autocomplete-ready-badge')).not.toBeInTheDocument();

    rerender(<AutoCompleteReadyBadge ready={true} />);
    expect(screen.getByTestId('autocomplete-ready-badge')).toBeInTheDocument();
    expect(screen.getByTestId('autocomplete-ready-badge')).toHaveAttribute('aria-live', 'polite');
  });

  it('auto-dismisses after the show timeout', () => {
    const { rerender } = render(<AutoCompleteReadyBadge ready={false} />);
    rerender(<AutoCompleteReadyBadge ready={true} />);
    expect(screen.getByTestId('autocomplete-ready-badge')).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(5000);
    });
    expect(screen.queryByTestId('autocomplete-ready-badge')).not.toBeInTheDocument();
  });

  it('hides immediately when ready becomes false again', () => {
    const { rerender } = render(<AutoCompleteReadyBadge ready={false} />);
    rerender(<AutoCompleteReadyBadge ready={true} />);
    expect(screen.getByTestId('autocomplete-ready-badge')).toBeInTheDocument();

    rerender(<AutoCompleteReadyBadge ready={false} />);
    expect(screen.queryByTestId('autocomplete-ready-badge')).not.toBeInTheDocument();
  });

  it('honors a custom testId', () => {
    const { rerender } = render(<AutoCompleteReadyBadge ready={false} testId="custom-id" />);
    rerender(<AutoCompleteReadyBadge ready={true} testId="custom-id" />);
    expect(screen.getByTestId('custom-id')).toBeInTheDocument();
  });
});
