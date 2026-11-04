type BackdropProps = {
  visible?: boolean;
  onClick?: () => void;
  label?: string;
};

export function Backdrop({
  visible = true,
  onClick,
  label = "Dismiss overlay",
}: BackdropProps) {
  if (!visible) return null;
  return (
    <button
      type="button"
      aria-label={label}
      className="fixed inset-0 z-40 border-0"
      style={{ background: "rgba(11, 31, 58, 0.45)" }}
      onClick={onClick}
    />
  );
}
