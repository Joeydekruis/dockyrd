/** Open a third-party URL without giving it access to this window. */
export function openExternalUrl(url: string) {
  window.open(url, "_blank", "noopener,noreferrer");
}
