import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { EnvVarRow } from "@/components/dockyrd/ui/env-var-row";

/** **Source:** `web/components/dockyrd/ui/env-var-row.tsx` */
const meta = {
  title: "Design/Components/Env var row",
  component: EnvVarRow,
  tags: ["autodocs"],
  argTypes: {
    name: { control: "text" },
    value: { control: "text" },
    secret: { control: "boolean" },
  },
  args: {
    name: "DATABASE_URL",
    value: "postgres://localhost:5432/app",
    secret: true,
  },
} satisfies Meta<typeof EnvVarRow>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {};
