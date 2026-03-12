import { useTranslation } from 'react-i18next';

interface LoadingSpinnerProps {
  loading: boolean;
}

export function LoadingSpinner({ loading }: LoadingSpinnerProps) {
  const { t: tc } = useTranslation('common');
  if (!loading) return null;
  return (
    <output aria-label={tc('status.loading')} className="flex justify-center py-2">
      <div aria-hidden="true" className="w-6 h-6 rounded-full border-4 border-white/30 border-t-white animate-spin" />
      <span className="sr-only">{tc('status.loading')}</span>
    </output>
  );
}
