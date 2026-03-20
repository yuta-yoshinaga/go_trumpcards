/** Toggle an item in an array: remove if present, append if absent. */
export function toggleArrayItem<T>(array: readonly T[], item: T): T[] {
  return array.includes(item) ? array.filter((i) => i !== item) : [...array, item];
}
