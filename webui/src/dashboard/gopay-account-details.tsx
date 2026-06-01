import { useState } from 'react';
import {
  ActionSection,
  AccountCarrierManualOTPSubmit,
  AccountDangerZone,
  AccountDetailActionSection,
  AccountDetailTabs,
  Badge,
  KVList,
  accountActionButtons,
  accountCarrierID,
  accountSubject,
  useQuery,
  type ActionButtonDescriptor,
} from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import { getGoPayAccountProfile, goPayKeys, submitGoPayManualOTP, type GoPayAccountProjection } from './gopay-api';
import {
  GOPAY_ACCOUNT_LIFECYCLE_ACTIONS,
  GOPAY_ACCOUNT_PRIMARY_ACTIONS,
  GOPAY_ACCOUNT_TOOL_ACTIONS,
  type GoPayAccountActionSpec,
} from './gopay-account-action-specs';
import { GoPayAccountPINDialog } from './gopay-account-pin-dialog';

export function GoPayAccountDetails({
  account,
  actionCatalog,
  busy,
  onRunAction,
  onOTPSubmitted,
  onDelete,
  onToast,
  onError,
}: {
  account: GoPayAccountProjection;
  actionCatalog?: AccountActionCatalog;
  busy?: boolean;
  onRunAction: (spec: GoPayAccountActionSpec, payload?: Record<string, unknown>) => Promise<boolean>;
  onOTPSubmitted?: () => void | Promise<void>;
  onDelete: (account: GoPayAccountProjection) => void | Promise<void>;
  onToast: (kind: 'error' | 'ok', message: string) => void;
  onError: (message: unknown) => void;
}) {
  const [pinAction, setPinAction] = useState<GoPayAccountActionSpec | null>(null);
  const accountID = accountCarrierID(account);
  const profile = useQuery({
    queryKey: goPayKeys.profile(accountID),
    queryFn: () => getGoPayAccountProfile(accountID),
    enabled: !!accountID,
  });

  async function run(spec: GoPayAccountActionSpec, payload: Record<string, unknown> = {}) {
    if (spec.requiresPinInput && !payload.pin) {
      setPinAction(spec);
      return;
    }
    await onRunAction(spec, payload);
  }

  async function submitPIN(payload: Record<string, unknown>) {
    if (!pinAction) return;
    if (await onRunAction(pinAction, payload)) setPinAction(null);
  }

  return (
    <div className="details gopayAccountDetails">
      <AccountDetailTabs tabs={[
        { value: 'details', label: '账户详情', content: <GoPayAccountOverview account={account} accountID={accountID} busy={busy} pin={profile.data?.pin || ''} pinConfigured={profile.data?.pin_configured} pinLoading={profile.isLoading} onOTPSubmitted={onOTPSubmitted} onDelete={onDelete} onToast={onToast} onError={onError} /> },
        { value: 'actions', label: '账户动作', content: <GoPayActionSection title="账户动作" description="先选中账号，再在账号详情里执行账户级操作。" specs={GOPAY_ACCOUNT_PRIMARY_ACTIONS} catalog={actionCatalog} account={account} busy={Boolean(busy)} run={run} /> },
        { value: 'checks', label: '状态检查', content: <GoPayActionSection title="状态检查" specs={GOPAY_ACCOUNT_TOOL_ACTIONS} catalog={actionCatalog} account={account} busy={Boolean(busy)} run={run} /> },
        { value: 'lifecycle', label: '生命周期', content: <GoPayActionSection title="账号生命周期" specs={GOPAY_ACCOUNT_LIFECYCLE_ACTIONS} catalog={actionCatalog} account={account} busy={Boolean(busy)} run={run} /> },
      ]} />
      <GoPayAccountPINDialog open={!!pinAction} busy={busy} onSubmit={submitPIN} onOpenChange={(open) => { if (!open) setPinAction(null); }} />
    </div>
  );
}

function GoPayAccountOverview({
  account,
  accountID,
  busy,
  pin,
  pinConfigured,
  pinLoading,
  onOTPSubmitted,
  onDelete,
  onToast,
  onError,
}: {
  account: GoPayAccountProjection;
  accountID: string;
  busy?: boolean;
  pin: string;
  pinConfigured?: boolean;
  pinLoading?: boolean;
  onOTPSubmitted?: () => void | Promise<void>;
  onDelete: (account: GoPayAccountProjection) => void | Promise<void>;
  onToast: (kind: 'error' | 'ok', message: string) => void;
  onError: (message: unknown) => void;
}) {
  return (
    <>
      <section className="grid gap-2 rounded-xl border bg-card p-3">
        <div className="sectionTitle">
          <h3 className="font-mono text-sm">{account.account ? accountSubject(account.account) || accountID : accountID}</h3>
        </div>
        <GoPayAccountBadges account={account} pinConfigured={pinConfigured} />
      </section>
      <AccountCarrierManualOTPSubmit
        account={account}
        keyPrefix="gopay-manual-otp"
        subtitle="GoPay 流程等待 OTP 时可从这里一次性提交；只恢复当前等待流程，不写入 OTP 缓存或历史。"
        disabled={busy}
        inputLabel="GoPay OTP"
        submit={submitGoPayManualOTP}
        onSuccess={async (resp) => {
          onToast('ok', resp.resume_count ? `已恢复 ${resp.resume_count} 个 GoPay OTP 流程` : 'GoPay OTP 已提交');
          await onOTPSubmitted?.();
        }}
        onError={onError}
      />
      <ActionSection title="账户详情">
        <KVList items={[
          { label: 'GoPayAccount ID', value: accountID, mono: true },
          { label: '手机号', value: formatPhone(account) },
          { label: '国家码', value: account.country_code || '-' },
          { label: 'OTP 渠道', value: account.otp_channel || '-' },
          { label: '状态', value: account.account?.status?.label || account.account?.status?.value || '-' },
          { label: 'PIN', value: pin || pinState(pinConfigured, pinLoading), mono: Boolean(pin) },
          { label: '余额', value: account.balance_currency ? `${account.balance_amount} ${account.balance_currency}` : '-' },
        ]} />
      </ActionSection>
      <AccountDangerZone account={account} busy={busy} onDelete={onDelete} />
    </>
  );
}

function GoPayAccountBadges({ account, pinConfigured }: { account: GoPayAccountProjection; pinConfigured?: boolean }) {
  return (
    <div className="flex flex-wrap gap-2">
      <Badge variant="outline">{account.account?.status?.label || account.account?.status?.value || 'created'}</Badge>
      <Badge variant={pinConfigured ? 'secondary' : 'outline'}>{pinConfigured ? 'PIN 已配置' : 'PIN 未配置'}</Badge>
      {account.otp_channel && <Badge variant="outline">{account.otp_channel}</Badge>}
    </div>
  );
}

function buttons(
  specs: GoPayAccountActionSpec[],
  catalog: AccountActionCatalog | undefined,
  account: GoPayAccountProjection,
  busy: boolean,
  run: (spec: GoPayAccountActionSpec) => void | Promise<void>,
): ActionButtonDescriptor[] {
  return accountActionButtons(
    { catalog, account, busy, placement: 'gopay' },
    specs.map((spec) => ({ ...spec, onClick: () => run(spec) })),
  );
}

function GoPayActionSection({ title, description, specs, catalog, account, busy, run }: {
  title: string;
  description?: string;
  specs: GoPayAccountActionSpec[];
  catalog?: AccountActionCatalog;
  account: GoPayAccountProjection;
  busy: boolean;
  run: (spec: GoPayAccountActionSpec) => void | Promise<void>;
}) {
  return <AccountDetailActionSection title={title} description={description} actions={buttons(specs, catalog, account, busy, run)} />;
}

function formatPhone(account: GoPayAccountProjection) {
  if (!account.phone) return '-';
  return `${account.country_code || '+62'} ${account.phone}`;
}

function pinState(configured?: boolean, loading?: boolean) {
  if (loading) return '读取中';
  return configured ? '已配置' : '未配置';
}
