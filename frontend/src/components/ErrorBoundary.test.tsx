import { fireEvent, render, screen } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';
import jaCommon from '../i18n/locales/ja/common.json';
import { ErrorBoundary } from './ErrorBoundary';

const errorTitle = jaCommon.label.errorBoundaryTitle;
const retryLabel = jaCommon.label.errorBoundaryRetry;
const reloadLabel = jaCommon.label.errorBoundaryReload;
const reportLabel = jaCommon.label.errorBoundaryReport;
const detailsLabel = jaCommon.label.errorBoundaryDetails;
const repeatedLabel = jaCommon.label.errorBoundaryRepeated;

function ThrowingChild(): ReactNode {
  throw new Error('test error');
}

describe('ErrorBoundary', () => {
  it('renders children when no error occurs', () => {
    render(
      <ErrorBoundary>
        <p>child content</p>
      </ErrorBoundary>,
    );
    expect(screen.getByText('child content')).toBeInTheDocument();
  });

  it('shows fallback UI when a child throws during render', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>,
    );
    expect(screen.getByText(errorTitle)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: retryLabel })).toBeInTheDocument();
    vi.restoreAllMocks();
  });

  it('renders error container with role="alert"', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>,
    );
    expect(screen.getByRole('alert')).toBeInTheDocument();
    vi.restoreAllMocks();
  });

  it('applies btnPrimary style to retry button', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>,
    );
    const btn = screen.getByRole('button', { name: retryLabel });
    expect(btn.className).toContain('min-h-[44px]');
    expect(btn.className).toContain('focus-visible:ring-2');
    vi.restoreAllMocks();
  });

  it('resets and re-renders children when retry button is clicked', () => {
    let shouldThrow = true;

    function MaybeThrow(): ReactNode {
      if (shouldThrow) {
        throw new Error('test error');
      }
      return <p>recovered</p>;
    }

    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <MaybeThrow />
      </ErrorBoundary>,
    );
    expect(screen.getByText(errorTitle)).toBeInTheDocument();

    shouldThrow = false;
    fireEvent.click(screen.getByRole('button', { name: retryLabel }));

    expect(screen.getByText('recovered')).toBeInTheDocument();
    expect(screen.queryByText(errorTitle)).not.toBeInTheDocument();
    vi.restoreAllMocks();
  });

  it('renders reload button and report link alongside retry', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>,
    );
    expect(screen.getByRole('button', { name: reloadLabel })).toBeInTheDocument();
    const reportLink = screen.getByRole('link', { name: reportLabel });
    expect(reportLink).toHaveAttribute('href');
    expect(reportLink.getAttribute('href')).toContain('github.com/yuta-yoshinaga/go_trumpcards/issues/new');
    expect(reportLink).toHaveAttribute('target', '_blank');
    expect(reportLink).toHaveAttribute('rel', expect.stringContaining('noopener'));
    vi.restoreAllMocks();
  });

  it('shows error details inside a collapsed <details>', () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>,
    );
    const summary = screen.getByText(detailsLabel);
    expect(summary).toBeInTheDocument();
    expect(summary.closest('details')).toBeInTheDocument();
    expect(screen.getByText(/Error: test error/)).toBeInTheDocument();
    vi.restoreAllMocks();
  });

  it('shows the repeated-retry warning after the threshold', () => {
    let shouldThrow = true;

    function MaybeThrow(): ReactNode {
      if (shouldThrow) {
        throw new Error('boom');
      }
      return <p>ok</p>;
    }

    vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <MaybeThrow />
      </ErrorBoundary>,
    );
    expect(screen.queryByText(repeatedLabel)).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: retryLabel }));
    fireEvent.click(screen.getByRole('button', { name: retryLabel }));
    expect(screen.getByText(repeatedLabel)).toBeInTheDocument();

    shouldThrow = false;
    fireEvent.click(screen.getByRole('button', { name: retryLabel }));
    expect(screen.getByText('ok')).toBeInTheDocument();
    vi.restoreAllMocks();
  });
});
