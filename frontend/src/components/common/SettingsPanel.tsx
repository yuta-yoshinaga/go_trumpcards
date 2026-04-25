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
      <div className="glass-panel rounded-lg p-3 mt-1 text-xs text-white">
        {groups.map((group, gi) => {
          const content = (
            <div className="flex flex-wrap gap-x-4 gap-y-1">
              {group.items.map((item) =>
                item.type === 'checkbox' ? (
                  <label
                    key={item.id}
                    htmlFor={item.id}
                    className="flex items-center gap-2 cursor-pointer relative group/tip min-h-[44px] px-1"
                  >
                    <input
                      id={item.id}
                      type="checkbox"
                      checked={item.checked ?? false}
                      disabled={item.disabled}
                      onChange={(e) => item.onToggle?.(e.target.checked)}
                      aria-describedby={item.tooltip ? `${item.id}-tooltip` : undefined}
                    />
                    {item.label}
                    {item.tooltip && (
                      <span
                        id={`${item.id}-tooltip`}
                        role="tooltip"
                        className="hidden group-hover/tip:block group-focus-within/tip:block absolute bottom-full left-0 mb-1 bg-ds-surface-elevated text-ds-text-primary text-xs rounded px-2 py-1 whitespace-nowrap z-10"
                      >
                        {item.tooltip}
                      </span>
                    )}
                  </label>
                ) : (
                  <span key={item.id} className="flex items-center gap-2 relative group/tip min-h-[44px] px-1">
                    <label htmlFor={item.id}>{item.label}</label>
                    <select
                      id={item.id}
                      value={item.value}
                      onChange={(e) => item.onSelect?.(e.target.value)}
                      className="bg-ds-surface-elevated text-ds-text-primary disabled:text-ds-text-muted disabled:opacity-70 rounded px-2 py-2 min-h-[44px]"
                      disabled={item.disabled}
                      aria-describedby={item.tooltip ? `${item.id}-tooltip` : undefined}
                    >
                      {item.options?.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                    {item.tooltip && (
                      <span
                        id={`${item.id}-tooltip`}
                        role="tooltip"
                        className="hidden group-hover/tip:block group-focus-within/tip:block absolute bottom-full left-0 mb-1 bg-ds-surface-elevated text-ds-text-primary text-xs rounded px-2 py-1 whitespace-nowrap z-10"
                      >
                        {item.tooltip}
                      </span>
                    )}
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
