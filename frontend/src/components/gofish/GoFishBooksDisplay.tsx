import { useTranslation } from 'react-i18next';
import type { GoFishBook } from '../../types/card';
import { valueName } from '../../utils/cardUtils';

/** Props for {@link GoFishBooksDisplay}. */
export interface GoFishBooksDisplayProps {
  books: GoFishBook[];
}

/** Renders a list of completed 4-of-a-kind books. */
export function GoFishBooksDisplay({ books }: GoFishBooksDisplayProps) {
  const { t } = useTranslation('gofish');

  if (books.length === 0) return null;

  return (
    <div className="my-2 p-2 rounded bg-black/30">
      <div className="text-ds-text-muted text-sm mb-1">{t('books', { count: books.length })}</div>
      <div className="flex flex-wrap gap-2">
        {books.map((book) => (
          <span
            key={book.rank}
            className="inline-flex items-center px-2 py-1 rounded bg-ds-success/50 text-white text-xs"
          >
            {valueName(book.rank)}
          </span>
        ))}
      </div>
    </div>
  );
}
