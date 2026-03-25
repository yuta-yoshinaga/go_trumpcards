import { useTranslation } from 'react-i18next';
import { useTutorialContext } from '../../providers/TutorialProvider';
import { btnSecondary } from '../../styles/buttonStyles';

/** Tutorial button that starts the game tutorial. */
export function TutorialButton() {
  const { t } = useTranslation('tutorial');
  const { start } = useTutorialContext();
  return (
    <button
      type="button"
      className={`${btnSecondary} text-xs`}
      onClick={start}
      aria-label={t('tutorialButton')}
      title={t('tutorialButton')}
    >
      ?
    </button>
  );
}
