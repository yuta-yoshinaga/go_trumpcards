import { useTranslation } from 'react-i18next';
import { SettingsPanel } from '../common/SettingsPanel';

interface GoFishSettingsDialogProps {
  cpuDifficulty: number;
  onCpuDifficultyChange: (value: string) => void;
}

/** CPU difficulty options for Go Fish. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
] as const;

/** Renders Go Fish settings panel with CPU difficulty selector. */
export function GoFishSettingsDialog({ cpuDifficulty, onCpuDifficultyChange }: GoFishSettingsDialogProps) {
  const { t } = useTranslation('gofish');

  return (
    <SettingsPanel
      title={t('setup.title')}
      groups={[
        {
          items: [
            {
              type: 'select',
              id: 'cpuDifficulty',
              label: t('setup.cpuDifficulty'),
              value: cpuDifficulty,
              options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                value: o.value,
                label: t(`setup.${o.label}`),
              })),
              onSelect: onCpuDifficultyChange,
            },
          ],
        },
      ]}
    />
  );
}
