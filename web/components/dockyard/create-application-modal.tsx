"use client";

import { CreateApplicationWizard } from "@/components/dockyard/create-application-wizard";
import type { ResourceProvisionContext } from "@/components/dockyard/resource-provisioning-modal";

type CreateApplicationModalProps = {
  open: boolean;
  onClose: () => void;
  onCreated: (context: ResourceProvisionContext) => void;
  defaultProjectId?: string;
};

/** @deprecated Use CreateApplicationWizard — kept for stable imports. */
export function CreateApplicationModal(props: CreateApplicationModalProps) {
  return <CreateApplicationWizard {...props} />;
}
