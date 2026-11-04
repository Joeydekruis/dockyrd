import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import {
  Archive,
  ArrowRight,
  BarChart3,
  Bell,
  Download,
  Folder,
  Globe,
  Home,
  LayoutGrid,
  Plus,
  Search,
  Settings,
} from "lucide-react";
import type { ReactNode } from "react";
import { DockyrdLogo } from "@/components/dockyard/dockyrd-logo";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import {
  Alert,
  Avatar,
  Button,
  Chip,
  CodeInline,
  CommitChip,
  Divider,
  Field,
  IconButton,
  InlineLink,
  Kbd,
  SplitButton,
  StatusBadge,
  TokenSwatch,
  TypographySample,
} from "@/components/dockyard/ui";
import { CatalogCanvas, CatalogGrid, CatalogRow } from "./catalog-layout";

const meta = {
  title: "Design/1 · Foundations & atoms",
} satisfies Meta;

export default meta;

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) {
  return (
    <section
      className="flex flex-col gap-6 rounded-[20px] border p-6"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <div>
        <h2 className="text-2xl font-semibold tracking-tight">{title}</h2>
        {description ? (
          <p className="mt-1.5 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        ) : null}
      </div>
      {children}
    </section>
  );
}

const colorTokens: { name: string; hex: string }[] = [
  { name: "Page bg", hex: "#FAF9F8" },
  { name: "Surface", hex: "#FFFFFF" },
  { name: "Border", hex: "#ECEAE6" },
  { name: "Input border", hex: "#E1DFDA" },
  { name: "Chip bg", hex: "#F1F0ED" },
  { name: "Ink navy", hex: "#0B1F3A" },
  { name: "Primary hover", hex: "#1A3358" },
  { name: "Muted text", hex: "#5A6577" },
  { name: "Secondary text", hex: "#3A4A63" },
  { name: "Secondary border", hex: "#CBD0D8" },
  { name: "Accent blue", hex: "#3B93F5" },
  { name: "Link blue", hex: "#0B57C9" },
  { name: "Active / tint", hex: "#E3F0FF" },
  { name: "Banner blue", hex: "#8EBEFF" },
  { name: "OK", hex: "#1A7F4B" },
  { name: "Warn", hex: "#D98A00" },
  { name: "Error", hex: "#B3261E" },
  { name: "Idle", hex: "#6B7686" },
];

export const Colors: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section
        title="Colors"
        description="Light mode only. Calendly-inspired navy and light blue on a neutral page background."
      >
        <CatalogGrid>
          {colorTokens.map((t) => (
            <TokenSwatch key={t.name} name={t.name} hex={t.hex} />
          ))}
        </CatalogGrid>
      </Section>
    </CatalogCanvas>
  ),
};

export const Typography: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Typography">
        <div className="flex flex-col gap-4">
          <TypographySample variant="h1">Dashboard</TypographySample>
          <TypographySample variant="h2">Recently updated</TypographySample>
          <TypographySample variant="body">
            Deploy to your own servers without the mess.
          </TypographySample>
          <TypographySample variant="muted">
            Production workloads for the marketing site and API.
          </TypographySample>
          <div>
            <p
              className="font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              Mono value · Geist Mono 13
            </p>
            <TypographySample variant="mono">
              a1b2c3d · api.clientx.nl
            </TypographySample>
          </div>
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const ShapeAndElevation: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section
        title="Shape & elevation"
        description="Radii and the one shadow used (command palette)."
      >
        <CatalogRow>
          {[
            { label: "Chip / pill · 6px", radius: 6 },
            { label: "Button / input · 10px", radius: 10 },
            { label: "Card · 20px", radius: 20 },
          ].map((item) => (
            <div key={item.label} className="flex flex-col gap-2">
              <span
                className="block size-16 border"
                style={{
                  borderRadius: item.radius,
                  borderColor: "var(--border)",
                  background: "var(--chip-bg)",
                }}
              />
              <span
                className="text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                {item.label}
              </span>
            </div>
          ))}
          <div className="flex flex-col gap-2">
            <span
              className="block size-16 rounded-xl border shadow-lg"
              style={{
                borderColor: "var(--border)",
                background: "var(--surface)",
              }}
            />
            <span
              className="text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              Command palette shadow
            </span>
          </div>
        </CatalogRow>
        <div
          className="mt-4 h-20 w-full max-w-md rounded-[20px]"
          style={{ background: "var(--hero-gradient)" }}
          aria-hidden
        />
        <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
          Hero gradient
        </p>
      </Section>
    </CatalogCanvas>
  ),
};

export const Logo: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section
        title="Logo"
        description="Hover the mark for the dobber wave + tilt."
      >
        <CatalogRow>
          <DockyrdLogo />
          <div
            className="inline-flex rounded-xl p-6"
            style={{ background: "var(--hero-gradient)" }}
          >
            <DockyrdLogo variant="dark" />
          </div>
        </CatalogRow>
      </Section>
    </CatalogCanvas>
  ),
};

export const PageLoader: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Page loader">
        <DockyrdPageLoading />
      </Section>
    </CatalogCanvas>
  ),
};

const designIcons = [
  { Icon: Home, label: "home" },
  { Icon: Folder, label: "folder" },
  { Icon: LayoutGrid, label: "app" },
  { Icon: Globe, label: "globe" },
  { Icon: Archive, label: "bucket" },
  { Icon: BarChart3, label: "chart" },
  { Icon: Settings, label: "settings" },
  { Icon: Plus, label: "plus" },
  { Icon: Download, label: "download" },
  { Icon: Search, label: "search" },
  { Icon: Bell, label: "bell" },
  { Icon: ArrowRight, label: "arrow" },
] as const;

export const Icons: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Icons" description="Lucide stroke 1.8 · 18px in nav">
        <CatalogGrid>
          {designIcons.map(({ Icon, label }) => (
            <div key={label} className="flex items-center gap-3">
              <Icon strokeWidth={1.8} className="size-[18px]" aria-hidden />
              <span
                className="font-mono text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                {label}
              </span>
            </div>
          ))}
        </CatalogGrid>
      </Section>
    </CatalogCanvas>
  ),
};

export const Buttons: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Buttons">
        <CatalogRow>
          <Button iconLeft={Plus}>Create application</Button>
          <Button variant="outline">Docs</Button>
          <Button variant="link" iconRight={ArrowRight}>
            View analytics
          </Button>
          <SplitButton iconLeft={Plus} onMenuClick={() => undefined}>
            Deploy
          </SplitButton>
        </CatalogRow>
        <div
          className="mt-4 inline-flex rounded-xl p-6"
          style={{ background: "var(--hero-gradient)" }}
        >
          <Button variant="secondary" iconLeft={Plus}>
            Add service
          </Button>
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const ChipsAndStatusPills: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Chips & status pills">
        <CatalogRow>
          <Chip>Open source</Chip>
          <CommitChip
            sha="a1b2c3d4e5f6"
            message="Rotate database credentials"
          />
          <StatusBadge label="Live" variant="ok" />
          <StatusBadge label="Deploying" variant="info" />
          <StatusBadge label="Needs attention" variant="warn" />
          <StatusBadge label="Failed" variant="err" />
        </CatalogRow>
      </Section>
    </CatalogCanvas>
  ),
};

export const Inputs: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Inputs">
        <div className="max-w-md space-y-4">
          <Field
            label="Public URL"
            defaultValue="https://panel.example.com"
            hint="Shown in install emails and audit logs."
          />
          <Field
            label="Project name"
            defaultValue="Client X"
            error="Name is already in use."
          />
        </div>
      </Section>
    </CatalogCanvas>
  ),
};

export const SmallAtoms: StoryObj = {
  render: () => (
    <CatalogCanvas>
      <Section title="Small atoms">
        <CatalogRow>
          <Kbd>⌘</Kbd>
          <Kbd>K</Kbd>
          <Avatar initials="JD" />
          <IconButton icon={Settings} label="Settings" />
          <CodeInline>DATABASE_URL</CodeInline>
          <InlineLink href="#">View docs</InlineLink>
        </CatalogRow>
        <Divider />
        <Alert
          message="Owners must enable two-factor authentication."
          tone="warning"
        />
      </Section>
    </CatalogCanvas>
  ),
};
