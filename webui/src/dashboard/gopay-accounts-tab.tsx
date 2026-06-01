import { WalletCards } from 'lucide-react';
import {
  ACCOUNT_PAGE_SIZE,
  AccountManagementDrawerView,
  Badge,
  accountCarrierID,
  accountSubjectRenderConfig,
  deleteAccountCarrier,
  useAccountWorkflowActionRunner,
  useQueryClient,
  type AccountListPagination,
  type AccountManagementControllerOptions,
} from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import type { ListGopayAccountsResponse } from '../proto/gopay_app';
import { deleteGoPayAccount, getGoPayAccounts, goPayKeys, type GoPayAccountProjection } from './gopay-api';
import { GoPayAccountAdd } from './gopay-account-add';
import { GoPayAccountDetails } from './gopay-account-details';
import type { GoPayAccountActionSpec } from './gopay-account-action-specs';

type GoPayAccountController = {
  accounts: GoPayAccountProjection[];
  selectedID: string;
  selected: GoPayAccountProjection | null;
  isLoading: boolean;
  actionBusy: boolean;
  accountsPagination?: AccountListPagination;
  invalidate: () => Promise<void>;
  selectAccount: (account: GoPayAccountProjection) => void;
  clearSelection: () => void;
};

export const goPayAccountControllerOptions = {
  queryKey: goPayKeys.accounts,
  queryFn: getGoPayAccounts,
  refetchInterval: 10000,
  pageSize: ACCOUNT_PAGE_SIZE,
  clearMissingSelection: true,
} satisfies AccountManagementControllerOptions<GoPayAccountProjection, ListGopayAccountsResponse>;

export function GoPayAccountsTab({
  controller,
  actionCatalog,
  loadingCatalog,
  onCreated,
  onToast,
  onError,
}: {
  controller: GoPayAccountController;
  actionCatalog?: AccountActionCatalog;
  loadingCatalog?: boolean;
  onCreated: (account: GoPayAccountProjection) => void | Promise<void>;
  onToast: (kind: 'error' | 'ok', message: string) => void;
  onError: (message: unknown) => void;
}) {
  const queryClient = useQueryClient();
  const runner = useAccountWorkflowActionRunner<GoPayAccountProjection, AccountActionCatalog>({
    catalog: actionCatalog,
    pathPrefix: '/api/gopay',
    actionKeyPrefix: 'gopay',
    toast: { showError: onError, showToast: onToast },
    onSuccess: async ({ account }) => {
      await controller.invalidate();
      await queryClient.invalidateQueries({ queryKey: goPayKeys.profile(accountCarrierID(account)) });
    },
    onError,
  });
  const busy = controller.isLoading || controller.actionBusy || runner.busy || loadingCatalog;

  async function runAction(account: GoPayAccountProjection, spec: GoPayAccountActionSpec, extra: Record<string, unknown> = {}) {
    return runner.runWorkflowAction({
      actionID: spec.actionID,
      account,
      payload: goPayWorkflowPayload(account, spec, extra),
    });
  }

  async function deleteAccount(account: GoPayAccountProjection) {
    const accountID = accountCarrierID(account);
    await runner.tryRunAccountAction('gopay:delete', account, async () => {
      const deleted = await deleteAccountCarrier(account, {
        deleteByID: () => deleteGoPayAccount(account),
        confirmMessage: () => `删除 GoPayAccount ${accountID}？`,
        invalidate: async () => {
          controller.clearSelection();
          await controller.invalidate();
          queryClient.removeQueries({ queryKey: goPayKeys.profile(accountID) });
        },
      });
      if (deleted) onToast('ok', 'GoPayAccount 已删除');
    }, { onError });
  }

  return (
    <AccountManagementDrawerView
      title="GoPayAccount"
      icon={<WalletCards size={16} />}
      actions={<GoPayAccountAdd disabled={busy} onCreated={onCreated} onError={(message) => onError(message)} />}
      carriers={controller.accounts}
      selectedCarrier={controller.selected}
      selectedID={controller.selectedID}
      loading={controller.isLoading}
      loadingText="加载 GoPayAccount..."
      emptyText="暂无已持久化 GoPayAccount"
      onSelectCarrier={controller.selectAccount}
      config={goPayAccountRenderConfig}
      pagination={controller.accountsPagination}
      renderChildren={(carrier) => <GoPayAccountMeta account={carrier} />}
      drawerTitle="GoPay账号详情"
      detail={(account) => (
        <GoPayAccountDetails
          account={account}
          actionCatalog={actionCatalog}
          busy={busy}
          onOTPSubmitted={async () => {
            await controller.invalidate();
            await queryClient.invalidateQueries({ queryKey: goPayKeys.profile(accountCarrierID(account)) });
          }}
          onToast={onToast}
          onError={onError}
          onDelete={deleteAccount}
          onRunAction={(spec, payload) => runAction(account, spec, payload)}
        />
      )}
      onCloseDetails={controller.clearSelection}
    />
  );
}

const goPayAccountRenderConfig = accountSubjectRenderConfig({ icon: () => <WalletCards size={15} /> });

function GoPayAccountMeta({ account }: { account: GoPayAccountProjection }) {
  return (
    <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
      <Badge variant="outline">{account.country_code || '+62'}</Badge>
      {account.phone && <Badge variant="secondary">{account.phone}</Badge>}
      {account.balance_currency && <Badge variant="outline">{account.balance_amount} {account.balance_currency}</Badge>}
    </div>
  );
}

function goPayWorkflowPayload(account: GoPayAccountProjection, spec: GoPayAccountActionSpec, extra: Record<string, unknown>) {
  return {
    gopay_account_id: accountCarrierID(account),
    operation: spec.operation,
    phone: account.phone,
    country_code: account.country_code,
    otp_channel: account.otp_channel || 'wa',
    ...extra,
  };
}
