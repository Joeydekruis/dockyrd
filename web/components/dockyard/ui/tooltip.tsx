"use client";

import {
  cloneElement,
  isValidElement,
  type ReactElement,
  type ReactNode,
  useId,
  useState,
} from "react";

export type TooltipPlacement = "top" | "bottom" | "left" | "right";

type TooltipProps = {
  trigger: ReactNode;
  content: ReactNode;
  placement?: TooltipPlacement;
  /** Storybook / controlled demos */
  defaultOpen?: boolean;
};

const panelPosition: Record<TooltipPlacement, string> = {
  top: "bottom-full left-1/2 mb-2 -translate-x-1/2",
  bottom: "top-full left-1/2 mt-2 -translate-x-1/2",
  left: "right-full top-1/2 mr-2 -translate-y-1/2",
  right: "left-full top-1/2 ml-2 -translate-y-1/2",
};

const arrowClass: Record<TooltipPlacement, string> = {
  top: "dockyrd-tooltip-arrow dockyrd-tooltip-arrow-bottom",
  bottom: "dockyrd-tooltip-arrow dockyrd-tooltip-arrow-top",
  left: "dockyrd-tooltip-arrow dockyrd-tooltip-arrow-right",
  right: "dockyrd-tooltip-arrow dockyrd-tooltip-arrow-left",
};

function mergeHandler(
  theirs: ((e: never) => void) | undefined,
  ours: () => void,
) {
  return (e: never) => {
    theirs?.(e);
    ours();
  };
}

export function Tooltip({
  trigger,
  content,
  placement = "top",
  defaultOpen = false,
}: TooltipProps) {
  const [hoverOpen, setHoverOpen] = useState(false);
  const tipId = useId();
  const show = defaultOpen || hoverOpen;

  const interactiveTrigger = isValidElement(trigger)
    ? cloneElement(trigger as ReactElement<Record<string, unknown>>, {
        "aria-describedby": show ? tipId : undefined,
        onMouseEnter: mergeHandler(
          (trigger as ReactElement<{ onMouseEnter?: unknown }>).props
            .onMouseEnter as ((e: React.MouseEvent) => void) | undefined,
          () => setHoverOpen(true),
        ),
        onMouseLeave: mergeHandler(
          (trigger as ReactElement<{ onMouseLeave?: unknown }>).props
            .onMouseLeave as ((e: React.MouseEvent) => void) | undefined,
          () => setHoverOpen(false),
        ),
        onFocus: mergeHandler(
          (trigger as ReactElement<{ onFocus?: unknown }>).props.onFocus as
            | ((e: React.FocusEvent) => void)
            | undefined,
          () => setHoverOpen(true),
        ),
        onBlur: mergeHandler(
          (trigger as ReactElement<{ onBlur?: unknown }>).props.onBlur as
            | ((e: React.FocusEvent) => void)
            | undefined,
          () => setHoverOpen(false),
        ),
      })
    : trigger;

  return (
    <span className="relative inline-flex">
      {interactiveTrigger}
      {show ? (
        <span
          id={tipId}
          role="tooltip"
          className={`dockyrd-tooltip absolute z-50 ${panelPosition[placement]}`}
        >
          {content}
          <span className={arrowClass[placement]} aria-hidden />
        </span>
      ) : null}
    </span>
  );
}

/** Inline icon chip for tooltip content (e.g. external link). */
export function TooltipInlineIcon({ children }: { children: ReactNode }) {
  return (
    <span className="dockyrd-tooltip-inline-icon" aria-hidden>
      {children}
    </span>
  );
}
