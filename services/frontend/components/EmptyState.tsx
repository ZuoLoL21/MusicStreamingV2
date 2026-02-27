import Link from 'next/link';

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  actionLabel?: string;
  actionHref?: string;
}

export default function EmptyState({
  icon,
  title,
  description,
  actionLabel,
  actionHref,
}: EmptyStateProps) {
  return (
    <div className="text-center py-12">
      {icon && <div className="mb-4 flex justify-center opacity-50">{icon}</div>}
      <h3 className="text-xl font-semibold mb-2">{title}</h3>
      {description && <p className="text-gray-400 mb-6">{description}</p>}
      {actionLabel && actionHref && (
        <Link
          href={actionHref}
          className="inline-block bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded-lg font-semibold"
        >
          {actionLabel}
        </Link>
      )}
    </div>
  );
}
