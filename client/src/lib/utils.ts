export function getInitials(name: string) {
  const words = name.split(/\s+/).filter(Boolean);
  // first and last word, so "Allen the Alien" is AA, not AT
  const picked = words.length > 1 ? [words[0], words[words.length - 1]] : words;
  return picked
    .map((w) => w[0])
    .join("")
    .toUpperCase();
}

// "just now", "5m ago", "3h ago", "2d ago", then the date
export function formatTimeAgo(iso: string, now = Date.now()) {
  const seconds = Math.max(0, (now - new Date(iso).getTime()) / 1000);
  if (seconds < 60) return "just now";
  if (seconds < 60 * 60) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 60 * 60 * 24) return `${Math.floor(seconds / 3600)}h ago`;
  if (seconds < 60 * 60 * 24 * 7) return `${Math.floor(seconds / 86400)}d ago`;
  return new Date(iso).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}
