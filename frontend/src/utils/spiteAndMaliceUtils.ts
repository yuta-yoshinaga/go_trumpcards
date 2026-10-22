/** Spite & Malice: K acts as a wild that fills any foundation slot. */
export const SpiteAndMaliceWildValue = 13;
/**
 * True when the card is the wild King.
 *
 * **同じ 13 を画面側にも書かない (#5560)。**ワイルドという規則はここが持ち主で、
 * 表示がその規則を自前で持つと、片方だけ直したときに黙って食い違う。
 */
export function isSpiteAndMaliceWild(card: { value: number } | null | undefined): boolean {
  return card?.value === SpiteAndMaliceWildValue;
}
