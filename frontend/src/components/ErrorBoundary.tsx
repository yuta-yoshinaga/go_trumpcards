import type { ErrorInfo, ReactNode } from 'react';
import { Component } from 'react';
import type { WithTranslation } from 'react-i18next';
import { withTranslation } from 'react-i18next';

interface ErrorBoundaryProps extends WithTranslation {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

class ErrorBoundaryInner extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('ErrorBoundary caught:', error, info);
  }

  handleRetry = () => {
    this.setState({ hasError: false });
  };

  render() {
    if (this.state.hasError) {
      const { t } = this.props;
      return (
        <div className="flex flex-col items-center justify-center h-full bg-gray-900 text-white gap-4">
          <h1 className="text-2xl font-bold">{t('label.errorBoundaryTitle')}</h1>
          <button
            type="button"
            onClick={this.handleRetry}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold"
          >
            {t('label.errorBoundaryRetry')}
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}

export const ErrorBoundary = withTranslation()(ErrorBoundaryInner);
