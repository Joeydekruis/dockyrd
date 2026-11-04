/** Open a third-party URL without giving it access to this window. */
export function openExternalUrl(url: string) {
  window.open(url, "_blank", "noopener,noreferrer");
}

/** GitHub App manifest registration requires POST, not GET with ?manifest=. */
export function postGitHubManifestLaunch(launch: {
  action: string;
  manifest: string;
}) {
  const form = document.createElement("form");
  form.method = "POST";
  form.action = launch.action;
  form.target = "_blank";
  form.rel = "noopener noreferrer";
  form.style.display = "none";

  const manifestField = document.createElement("input");
  manifestField.type = "hidden";
  manifestField.name = "manifest";
  manifestField.value = launch.manifest;
  form.appendChild(manifestField);

  document.body.appendChild(form);
  form.submit();
  form.remove();
}
