import Image from "next/image";

type DockyrdLogoProps = {
  variant?: "light" | "dark";
  width?: number;
};

export function DockyrdLogo({ variant = "light", width = 150 }: DockyrdLogoProps) {
  return (
    <Image
      src="/dockyrd-logo.svg"
      alt="Dockyrd"
      width={width}
      height={Math.round(width * 0.21)}
      priority
      className={variant === "dark" ? "brightness-0 invert" : undefined}
    />
  );
}
