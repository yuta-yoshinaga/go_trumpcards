/** Ordinary sentence wraps at the observed 375×667 viewport. */
export const ALLOWED_SENTENCE_WRAPS: Readonly<Record<string, readonly string[]>> = {
  '/bouillotte': ['参加中'],
  '/threecardbrag': ['ブラインド'],
};

/** Video poker paytable wraps required to keep cells within their width at 375×667 (#11724). */
export const ALLOWED_PAYTABLE_WRAPS: Readonly<Record<string, readonly string[]>> = {
  '/videopoker': ['ストレートフラッシュ', 'ロイヤルフラッシュ'],
  '/deuceswild': ['ストレートフラッシュ', 'フォーデューシーズ'],
  '/jokerpoker': ['ストレートフラッシュ'],
};

/** Returns whether a wrapped label is allowed for this route and viewport. */
export function isAllowedLabelWrap(path: string, label: string, viewport: { width: number; height: number }): boolean {
  if (viewport.width !== 375 || viewport.height !== 667) return false;
  return [...(ALLOWED_SENTENCE_WRAPS[path] ?? []), ...(ALLOWED_PAYTABLE_WRAPS[path] ?? [])].includes(label);
}
