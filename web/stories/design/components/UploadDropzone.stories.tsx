import type { Meta, StoryObj } from "@storybook/nextjs-vite";
import { useState } from "react";
import {
  UploadDropzone,
  type UploadDropzoneState,
} from "@/components/dockyard/ui/upload-dropzone";

const states: UploadDropzoneState[] = ["idle", "uploading", "done", "error"];

const meta = {
  title: "Design/Components/Upload dropzone",
  component: UploadDropzone,
  tags: ["autodocs"],
  argTypes: {
    state: { control: "select", options: states },
    title: { control: "text" },
    hint: { control: "text" },
    errorMessage: { control: "text" },
  },
  args: {
    state: "idle",
    title: "Drop files here or browse",
    hint: "PNG, JPG up to 10 MB",
    errorMessage: "File too large",
    onBrowse: () => {},
  },
} satisfies Meta<typeof UploadDropzone>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Playground: Story = {
  render: (args) => {
    const [state, setState] = useState<UploadDropzoneState>(
      args.state ?? "idle",
    );
    return (
      <div className="max-w-md p-6">
        <UploadDropzone
          {...args}
          state={state}
          onBrowse={() => setState("uploading")}
        />
      </div>
    );
  },
};

export const AllVariants: Story = {
  parameters: { controls: { disable: true } },
  render: () => (
    <div className="grid max-w-2xl grid-cols-2 gap-4 p-6">
      {states.map((state) => (
        <UploadDropzone
          key={state}
          state={state}
          title={state}
          errorMessage={state === "error" ? "Upload failed" : undefined}
        />
      ))}
    </div>
  ),
};
