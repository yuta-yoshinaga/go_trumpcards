import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

/** A single setting item (checkbox or select) in the settings panel. */
export interface SettingsItem {
  type: 'checkbox' | 'select';
  id: string;
  label: string;
  tooltip?: string;
  // checkbox
  checked?: boolean;
  onToggle?: (checked: boolean) => void;
  // select
  value?: string | number;
  options?: { value: string | number; label: string }[];
  onSelect?: (value: string) => void;
  disabled?: boolean;
  /** Optional `data-testid` applied to the underlying input/select element. */
  testId?: string;
}

/** A group of related settings items with optional title. */
export interface SettingsGroup {
  id?: string;
  title?: string;
  items: SettingsItem[];
}

interface SettingsPanelProps {
  title: string;
  groups: SettingsGroup[];
}

/** Renders a collapsible settings panel with grouped checkboxes and selects. */
export function SettingsPanel({ title, groups }: SettingsPanelProps) {
  const { t } = useTranslation('common');
  const [openTipId, setOpenTipId] = useState<string | null>(null);

  useEffect(() => {
    if (openTipId === null) return;
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpenTipId(null);
    };
    document.addEventListener('keydown', handleEsc);
    return () => document.removeEventListener('keydown', handleEsc);
  }, [openTipId]);

  const renderHelp = (item: SettingsItem) => {
    if (!item.tooltip) return null;
    const isOpen = openTipId === item.id;
    return (
      <>
        <button
          type="button"
          aria-label={isOpen ? t('settings.hideHelp') : t('settings.showHelp')}
          aria-expanded={isOpen}
          aria-controls={`${item.id}-tooltip`}
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            setOpenTipId(isOpen ? null : item.id);
          }}
          className="text-ds-text-muted hover:text-ds-accent w-11 h-11 inline-flex items-center justify-center rounded-full text-xs"
        >
          (?)
        </button>
        {isOpen && (
          <span
            id={`${item.id}-tooltip`}
            role="tooltip"
            className="basis-full text-xs text-ds-text-muted mt-1 max-w-prose"
          >
            {item.tooltip}
          </span>
        )}
      </>
    );
  };

  return (
    <details className="px-4 pt-2">
      <summary className="text-ds-text-primary text-sm cursor-pointer select-none inline-flex items-center gap-1.5 hover:text-ds-accent transition-colors py-1">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
        {title}
      </summary>
      <div className="glass-panel rounded-lg p-3 mt-1 text-xs text-ds-text-primary">
        {groups.map((group, gi) => {
          const content = (
            <div className="flex flex-wrap gap-x-4 gap-y-1">
              {group.items.map((item) =>
                item.type === 'checkbox' ? (
                  <span key={item.id} className="flex flex-wrap items-center gap-2 px-1">
                    <label htmlFor={item.id} className="flex items-center gap-2 cursor-pointer min-h-[44px]">
                      <input
                        id={item.id}
                        type="checkbox"
                        checked={item.checked ?? false}
                        disabled={item.disabled}
                        onChange={(e) => item.onToggle?.(e.target.checked)}
                        aria-describedby={item.tooltip ? `${item.id}-tooltip` : undefined}
                        data-testid={item.testId}
                      />
                      {item.label}
                    </label>
                    {renderHelp(item)}
                  </span>
                ) : (
                  <span key={item.id} className="flex flex-wrap items-center gap-2 min-h-[44px] px-1">
                    <label htmlFor={item.id}>{item.label}</label>
                    <select
                      id={item.id}
                      value={item.value}
                      onChange={(e) => item.onSelect?.(e.target.value)}
                      className="bg-ds-surface-elevated text-ds-text-primary disabled:text-ds-text-muted disabled:opacity-70 rounded px-2 py-2 min-h-[44px]"
                      disabled={item.disabled}
                      aria-describedby={item.tooltip ? `${item.id}-tooltip` : undefined}
                      data-testid={item.testId}
                    >
                      {item.options?.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                    {renderHelp(item)}
                  </span>
                ),
              )}
            </div>
          );

          if (group.title) {
            return (
              <fieldset key={group.id ?? `group-${gi.toString()}`} className={gi > 0 ? 'mt-2' : ''}>
                <legend className="text-ds-warning font-bold text-xs mb-1">{group.title}</legend>
                {content}
              </fieldset>
            );
          }
          return (
            <div key={group.id ?? `group-${gi.toString()}`} className={gi > 0 ? 'mt-2' : ''}>
              {content}
            </div>
          );
        })}
      </div>
    </details>
  );
}
