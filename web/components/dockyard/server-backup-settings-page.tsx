"use client";

import { useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { Field } from "@/components/dockyard/ui/field";
import {
  SettingsDetailBlock,
  SettingsEnabledStatus,
} from "@/components/dockyard/ui/settings-detail-block";
import {
  configureGarageBackupTarget,
  getServerBackupTarget,
  patchServerBackupTarget,
  type ServerBackupTarget,
  testServerBackupTarget,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

export function ServerBackupSettingsPage() {
  const [target, setTarget] = useState<ServerBackupTarget | null>(null);
  const [endpoint, setEndpoint] = useState("");
  const [bucket, setBucket] = useState("");
  const [prefix, setPrefix] = useState("");
  const [region, setRegion] = useState("");
  const [accessKey, setAccessKey] = useState("");
  const [secretKey, setSecretKey] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [saved, setSaved] = useState(false);
  const [garageBusy, setGarageBusy] = useState(false);
  const [testBusy, setTestBusy] = useState(false);

  useEffect(() => {
    void getServerBackupTarget()
      .then((t) => {
        setTarget(t);
        setEndpoint(t.endpoint);
        setBucket(t.bucket);
        setPrefix(t.prefix);
        setRegion(t.region);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : "Failed to load");
      });
  }, []);

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setSaved(false);
    try {
      const t = await patchServerBackupTarget({
        endpoint,
        bucket,
        prefix,
        region,
        access_key: accessKey || undefined,
        secret_key: secretKey || undefined,
      });
      setTarget(t);
      setAccessKey("");
      setSecretKey("");
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    } finally {
      setBusy(false);
    }
  }

  async function onUseGarage() {
    setGarageBusy(true);
    setError(null);
    try {
      const t = await configureGarageBackupTarget();
      setTarget(t);
      setEndpoint(t.endpoint);
      setBucket(t.bucket);
      setPrefix(t.prefix);
      setRegion(t.region);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Garage setup failed");
    } finally {
      setGarageBusy(false);
    }
  }

  async function onTestTarget() {
    setTestBusy(true);
    setError(null);
    try {
      const t = await testServerBackupTarget();
      setTarget(t);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Test failed");
    } finally {
      setTestBusy(false);
    }
  }

  return (
    <DashboardShell title="Backups" breadcrumb="Server settings">
      <section
        className={RESOURCE_DETAIL_CARD}
        style={resourceDetailCardStyle()}
      >
        <h1 className="text-2xl font-semibold tracking-tight">
          External backup target
        </h1>
        <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
          S3-compatible storage for scheduled database dumps from every project.
          Credentials are encrypted at rest.
        </p>
        {target ? (
          <div className="mt-6">
            <SettingsDetailBlock
              title="Current target"
              rows={[
                {
                  label: "Status",
                  value: <SettingsEnabledStatus enabled={target.configured} />,
                },
                { label: "Bucket", value: target.bucket || "—" },
                { label: "Prefix", value: target.prefix || "—" },
                {
                  label: "Keys stored",
                  value: <SettingsEnabledStatus enabled={target.has_keys} />,
                },
                ...(target.managed_garage
                  ? [
                      {
                        label: "Location",
                        value: "Same server (Garage)",
                      },
                    ]
                  : []),
                ...(target.last_test_at
                  ? [
                      {
                        label: "Last test",
                        value: new Date(
                          target.last_test_at * 1000,
                        ).toLocaleString(),
                      },
                    ]
                  : []),
                ...(target.last_test_error
                  ? [
                      {
                        label: "Last test error",
                        value: target.last_test_error,
                      },
                    ]
                  : []),
              ]}
              actionDisabled
            />
          </div>
        ) : null}
        <div className="mt-6 flex flex-wrap gap-3">
          <Button
            type="button"
            variant="outline"
            disabled={garageBusy}
            onClick={() => void onUseGarage()}
          >
            {garageBusy ? "Configuring…" : "Use built-in Garage"}
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={testBusy || !target?.configured}
            onClick={() => void onTestTarget()}
          >
            {testBusy ? "Testing…" : "Test connection"}
          </Button>
        </div>
        <form
          className="mt-8 max-w-lg space-y-4"
          onSubmit={(e) => void onSave(e)}
        >
          {error ? (
            <p className="text-sm text-red-600" role="alert">
              {error}
            </p>
          ) : null}
          {saved ? (
            <p className="text-sm" style={{ color: "#2d8a4e" }}>
              Saved.
            </p>
          ) : null}
          <Field
            label="Endpoint (leave empty for AWS)"
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            placeholder="https://s3.amazonaws.com"
          />
          <Field
            label="Bucket"
            value={bucket}
            onChange={(e) => setBucket(e.target.value)}
            required
          />
          <Field
            label="Prefix"
            value={prefix}
            onChange={(e) => setPrefix(e.target.value)}
            placeholder="dockyrd/backups"
          />
          <Field
            label="Region"
            value={region}
            onChange={(e) => setRegion(e.target.value)}
            placeholder="eu-west-1"
          />
          <Field
            label="Access key"
            value={accessKey}
            onChange={(e) => setAccessKey(e.target.value)}
            placeholder={target?.has_keys ? "Leave blank to keep existing" : ""}
            autoComplete="off"
          />
          <Field
            label="Secret key"
            type="password"
            value={secretKey}
            onChange={(e) => setSecretKey(e.target.value)}
            placeholder={target?.has_keys ? "Leave blank to keep existing" : ""}
            autoComplete="new-password"
          />
          <Button type="submit" disabled={busy || !bucket.trim()}>
            Save target
          </Button>
        </form>
      </section>
    </DashboardShell>
  );
}
