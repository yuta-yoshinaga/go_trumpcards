import { ApiError } from '../api/gameExec';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import i18n from '../i18n';

/** Maps an API failure to user-facing copy and whether retrying can help. */
export function describeApiFailure(error: unknown): { message: string; retryable: boolean } {
  if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
    return { message: i18n.t('label.requestRejected', { ns: 'common' }), retryable: false };
  }
  if (error instanceof ApiError && error.status >= 500) {
    return { message: i18n.t('label.serverError', { ns: 'common' }), retryable: true };
  }
  return { message: NETWORK_ERROR_MESSAGE(), retryable: true };
}
