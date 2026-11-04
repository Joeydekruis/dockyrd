"use client";

import {
  type CSSProperties,
  type RefObject,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

const MENU_GAP_PX = 4;
const MENU_MAX_HEIGHT_PX = 280;
const VIEWPORT_PADDING_PX = 8;
/** Above wizard/dialog overlay (`z-50`). */
const MENU_Z_INDEX = 200;

export function useSelectMenuPosition(
  open: boolean,
  anchorRef: RefObject<HTMLElement | null>,
) {
  const menuRef = useRef<HTMLDivElement>(null);
  const [menuStyle, setMenuStyle] = useState<CSSProperties>({});

  useLayoutEffect(() => {
    if (!open || !anchorRef.current) {
      return;
    }

    const update = () => {
      const anchor = anchorRef.current;
      if (!anchor) return;

      const rect = anchor.getBoundingClientRect();
      const spaceBelow =
        window.innerHeight - rect.bottom - MENU_GAP_PX - VIEWPORT_PADDING_PX;
      const spaceAbove = rect.top - MENU_GAP_PX - VIEWPORT_PADDING_PX;

      const openUp =
        spaceBelow < Math.min(MENU_MAX_HEIGHT_PX, 160) &&
        spaceAbove > spaceBelow;

      const maxHeight = Math.min(
        MENU_MAX_HEIGHT_PX,
        Math.max(0, openUp ? spaceAbove : spaceBelow),
      );

      const top = openUp
        ? Math.max(VIEWPORT_PADDING_PX, rect.top - MENU_GAP_PX - maxHeight)
        : rect.bottom + MENU_GAP_PX;

      setMenuStyle({
        position: "fixed",
        top,
        left: rect.left,
        width: rect.width,
        maxHeight,
        zIndex: MENU_Z_INDEX,
      });
    };

    update();
    window.addEventListener("resize", update);
    window.addEventListener("scroll", update, true);
    return () => {
      window.removeEventListener("resize", update);
      window.removeEventListener("scroll", update, true);
    };
  }, [open, anchorRef]);

  return { menuRef, menuStyle };
}

export function isSelectMenuOutsideClick(
  target: Node,
  anchorRef: RefObject<HTMLElement | null>,
  menuRef: RefObject<HTMLElement | null>,
): boolean {
  if (anchorRef.current?.contains(target)) return false;
  if (menuRef.current?.contains(target)) return false;
  return true;
}
