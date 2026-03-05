export function toggleArrayItem<T>(array: readonly T[], item: T): T[] {
  return array.includes(item) ? array.filter((i) => i !== item) : [...array, item];
}
