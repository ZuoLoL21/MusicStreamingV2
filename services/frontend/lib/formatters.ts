/**
 * Format duration in seconds to MM:SS format
 * @param seconds - Duration in seconds (can be float)
 * @returns Formatted string like "4:07"
 */
export function formatDuration(seconds: number): string {
  if (!seconds || seconds < 0) return '0:00';

  const totalSeconds = Math.floor(seconds);
  const minutes = Math.floor(totalSeconds / 60);
  const remainingSeconds = totalSeconds % 60;

  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
}

/**
 * Format date to locale string
 * @param date - Date string or Date object
 * @returns Formatted date string
 */
export function formatDate(date: string | Date | null | undefined): string {
  if (!date) return 'Unknown';

  try {
    const d = typeof date === 'string' ? new Date(date) : date;
    if (isNaN(d.getTime())) return 'Invalid date';
    return d.toLocaleDateString();
  } catch {
    return 'Invalid date';
  }
}

/**
 * Format number with commas
 * @param num - Number to format
 * @returns Formatted string like "1,234"
 */
export function formatNumber(num: number | null | undefined): string {
  if (num === null || num === undefined) return '0';
  return num.toLocaleString();
}

/**
 * Format timestamp as relative time
 * @param timestamp - ISO 8601 timestamp string
 * @returns Relative time string like "2 hours ago", "3 days ago"
 */
export function formatRelativeTime(timestamp: string | null | undefined): string {
  if (!timestamp) return 'Unknown';

  try {
    const date = new Date(timestamp);
    if (isNaN(date.getTime())) return 'Invalid date';

    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSeconds = Math.floor(diffMs / 1000);
    const diffMinutes = Math.floor(diffSeconds / 60);
    const diffHours = Math.floor(diffMinutes / 60);
    const diffDays = Math.floor(diffHours / 24);
    const diffWeeks = Math.floor(diffDays / 7);
    const diffMonths = Math.floor(diffDays / 30);
    const diffYears = Math.floor(diffDays / 365);

    if (diffSeconds < 60) {
      return 'Just now';
    } else if (diffMinutes < 60) {
      return `${diffMinutes} ${diffMinutes === 1 ? 'minute' : 'minutes'} ago`;
    } else if (diffHours < 24) {
      return `${diffHours} ${diffHours === 1 ? 'hour' : 'hours'} ago`;
    } else if (diffDays < 7) {
      return `${diffDays} ${diffDays === 1 ? 'day' : 'days'} ago`;
    } else if (diffWeeks < 4) {
      return `${diffWeeks} ${diffWeeks === 1 ? 'week' : 'weeks'} ago`;
    } else if (diffMonths < 12) {
      return `${diffMonths} ${diffMonths === 1 ? 'month' : 'months'} ago`;
    } else {
      return `${diffYears} ${diffYears === 1 ? 'year' : 'years'} ago`;
    }
  } catch {
    return 'Invalid date';
  }
}
