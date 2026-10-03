import { describe, expect, it } from 'vitest';
import { ApiError } from '../api/gameExec';
import { describeApiFailure } from './describeApiFailure';

describe('describeApiFailure', () => {
  it('describes network failures with retry', () => {
    expect(describeApiFailure(new TypeError('failed to fetch'))).toEqual({
      message: '通信エラーが発生しました。もう一度お試しください。',
      retryable: true,
    });
  });

  it('describes rejected requests without retry', () => {
    expect(describeApiFailure(new ApiError(400, 'param error'))).toEqual({
      message: 'この操作は受け付けられませんでした。',
      retryable: false,
    });
  });

  it('describes server failures with retry', () => {
    expect(describeApiFailure(new ApiError(503))).toEqual({
      message: 'サーバでエラーが発生しました。解決しない場合は新しいゲームを始めてください。',
      retryable: true,
    });
  });

  it('treats other errors as network failures', () => {
    expect(describeApiFailure(new Error('unexpected'))).toEqual({
      message: '通信エラーが発生しました。もう一度お試しください。',
      retryable: true,
    });
  });
});
