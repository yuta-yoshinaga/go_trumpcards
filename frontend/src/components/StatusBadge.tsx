export type StatusBadgeVariant = 'success' | 'warning';

const styles: Record<StatusBadgeVariant, React.CSSProperties> = {
  success: {
    background: '#5cb85c',
    color: '#fff',
    borderRadius: 6,
    padding: '1px 8px',
    marginLeft: 6,
    fontSize: '0.8em',
  },
  warning: {
    background: '#f0ad4e',
    color: '#222',
    borderRadius: 6,
    padding: '1px 8px',
    marginLeft: 6,
    fontSize: '0.8em',
    fontWeight: 'bold',
  },
};

export function StatusBadge({ variant, children }: { variant: StatusBadgeVariant; children: React.ReactNode }) {
  return <span style={styles[variant]}>{children}</span>;
}
