import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { UserMenu } from "@/components/dockyrd/user-menu";

const demoUser = {
  id: "1",
  name: "Joey de Kruis",
  email: "joey@derisory.nl",
  role: "owner",
  totp_enabled: true,
};

const meta = {
  title: "Design/Components/User menu",
  component: UserMenu,
  tags: ["autodocs"],
  parameters: { layout: "centered" },
  args: { user: demoUser },
} satisfies Meta<typeof UserMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

export const Open: Story = {
  args: { defaultOpen: true },
};
