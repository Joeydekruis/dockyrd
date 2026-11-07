"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/dockyrd/ui/button";
import { Field } from "@/components/dockyrd/ui/field";
import { Modal } from "@/components/dockyrd/ui/modal";

type AddCustomDomainModalProps = {
  open: boolean;
  busy?: boolean;
  onClose: () => void;
  onSubmit: (hostname: string) => Promise<void>;
};

export function AddCustomDomainModal({
  open,
  busy,
  onClose,
  onSubmit,
}: AddCustomDomainModalProps) {
  const [hostname, setHostname] = useState("");

  useEffect(() => {
    if (open) setHostname("");
  }, [open]);

  return (
    <Modal
      open={open}
      title="Add custom domain"
      description="Serve this service on a domain you own. Next, you'll add DNS records to verify it."
      onClose={onClose}
      footer={
        <>
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={busy}
          >
            Cancel
          </Button>
          <Button
            type="button"
            disabled={busy || !hostname.trim()}
            onClick={() => void onSubmit(hostname.trim())}
          >
            Add custom domain
          </Button>
        </>
      }
    >
      <Field
        label="Domain name"
        value={hostname}
        onChange={(e) => setHostname(e.target.value)}
        placeholder="example.com"
        hint="A domain or subdomain you own, like shop.example.com."
        autoFocus
      />
    </Modal>
  );
}
