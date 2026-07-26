import type { ErrorInfo, ReactNode } from 'react';
import { Component } from 'react';
import type { WithTranslation } from 'react-i18next';
import { withTranslation } from 'react-i18next';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';

/** Props for {@link ErrorBoundary}, injected via `withTranslation()`. */
export interface ErrorBoundaryProps extends WithTranslation {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
  retryCount: number;
}

/** GitHub repo issue endpoint used to pre-fill bug report links from the fallback UI. */
const ISSUE_URL = 'https://github.com/yuta-yoshinaga/go_trumpcards/issues/new';

/** Threshold (inclusive) at which the fallback warns that retrying keeps producing the same crash. */
const REPEATED_RETRY_THRESHOLD = 2;

class ErrorBoundaryInner extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null, retryCount: 0 };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('ErrorBoundary caught:', error, info);
    this.setState({ errorInfo: info });
  }

  handleRetry = () => {
    this.setState((prev) => ({
      hasError: false,
      error: null,
      errorInfo: null,
      retryCount: prev.retryCount + 1,
    }));
  };

  handleReload = () => {
    window.location.reload();
  };

  buildIssueUrl(): string {
    const { error } = this.state;
    const title = encodeURIComponent(`[Bug] ${error?.name ?? 'Unknown error'}: ${error?.message ?? ''}`);
    const stack = (error?.stack ?? '(none)').slice(0, 1000);
    const body = encodeURIComponent(
      `**URL:** ${typeof window !== 'undefined' ? window.location.href : '(unknown)'}\n` +
        `**User Agent:** ${typeof navigator !== 'undefined' ? navigator.userAgent : '(unknown)'}\n` +
        `**Error:** ${error?.name ?? ''}: ${error?.message ?? ''}\n` +
        `**Stack:**\n\`\`\`\n${stack}\n\`\`\`\n`,
    );
    return `${ISSUE_URL}?title=${title}&body=${body}`;
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children;
    }
    const { t } = this.props;
    const { error, retryCount } = this.state;
    const showStack = import.meta.env.DEV && error?.stack;
    return (
      <div
        role="alert"
        className="flex flex-col items-center justify-center h-full bg-ds-surface text-ds-text-primary gap-4 p-6 max-w-prose mx-auto"
      >
        <h1 className="text-2xl font-bold">{t('label.errorBoundaryTitle')}</h1>
        <p className="text-ds-text-muted text-sm text-center">{t('label.errorBoundaryHelp')}</p>

        <div className="flex gap-2 flex-wrap justify-center">
          <button type="button" onClick={this.handleRetry} className={btnPrimary}>
            {t('label.errorBoundaryRetry')}
          </button>
          <button type="button" onClick={this.handleReload} className={btnSecondary}>
            {t('label.errorBoundaryReload')}
          </button>
          <a href={this.buildIssueUrl()} target="_blank" rel="noopener noreferrer" className={btnSecondary}>
            {t('label.errorBoundaryReport')}
          </a>
        </div>

        {retryCount >= REPEATED_RETRY_THRESHOLD && (
          <p role="status" className="text-ds-warning text-sm">
            {t('label.errorBoundaryRepeated')}
          </p>
        )}

        {error && (
          <details className="w-full text-xs text-ds-text-muted">
            <summary className="cursor-pointer">{t('label.errorBoundaryDetails')}</summary>
            <pre className="mt-2 whitespace-pre-wrap break-all bg-black/30 p-3 rounded">
              {error.name}: {error.message}
              {showStack ? `\n\n${error.stack}` : ''}
            </pre>
          </details>
        )}
      </div>
    );
  }
}

/** Error boundary component that catches render errors and shows recovery actions, error details, and a pre-filled issue report link. */
export const ErrorBoundary = withTranslation()(ErrorBoundaryInner);
