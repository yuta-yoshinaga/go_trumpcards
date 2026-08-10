/**
 * Localized Badugi hand name for a hand size (1-4).
 *
 * Sync: `cuiBadugiHandName` (`internal/adapter/presenter/BadugiCuiPresenter.go`).
 *
 * **サーバの `handName` は英語の生値。**バドゥーギの役名はポーカー役表と対応
 * しないので共通の訳表が無く、Web だけ英語のまま残っていた。役は成立枚数
 * そのものなので、`handSize` から引く（文字列一致はしない）。
 */
export function badugiHandName(handSize: number, t: (key: string) => string): string {
  if (handSize < 1 || handSize > 4) return '';
  return t(`handName.${handSize.toString()}`);
}
