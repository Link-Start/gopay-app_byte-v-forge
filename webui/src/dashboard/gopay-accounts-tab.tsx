import { WalletCards } from 'lucide-react';
import {
  ACCOUNT_PAGE_SIZE,
  AccountManagementView,
  Badge,
  DetailDrawer,
  accountCarrierID,
  accountSubject,
  submitAccountWorkflowAction,
  useAsyncActionRunner,
  useQueryClient,
  type AccountListPagination,
  type AccountManagementControllerOptions,
  type AccountRenderConfig,
} from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import type { ListGopayAccountsResponse } from '../proto/gopay_app';
import { getGoPayAccounts, goPayKeys, type GoPayAccountProjection } from './gopay-api';
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
  const runner = useAsyncActionRunner();
  const queryClient = useQueryClient();
  const busy = controller.isLoading || controller.actionBusy || runner.busy || loadingCatalog;

  async function runAction(account: GoPayAccountProjection, spec: GoPayAccountActionSpec, extra: Record<string, unknown> = {}) {
    const accountID = accountCarrierID(account);
    const result = await runner.tryRun(`gopay:${spec.actionID}:${accountID}`, () => submitAccountWorkflowAction({
      catalog: actionCatalog,
      actionID: spec.actionID,
      pathPrefix: '/api/gopay',
      payload: goPayWorkflowPayload(account, spec, extra),
      toast: { showError: onError, showToast: onToast },
      onSuccess: async () => {
        await controller.invalidate();
        await queryClient.invalidateQueries({ queryKey: goPayKeys.profile(accountID) });
      },
    }), { onError });
    return result.ok && !result.value.error_message;
  }

  return (
    <>
      <AccountManagementView
        title="GoPayAccount"
        icon={<WalletCards size={16} />}
        actions={<GoPayAccountAdd disabled={busy} onCreated={onCreated} onError={(message) => onError(message)} />}
        carriers={controller.accounts}
        selectedID={controller.selectedID}
        loading={controller.isLoading}
        loadingText="加载 GoPayAccount..."
        emptyText="暂无已持久化 GoPayAccount"
        onSelectCarrier={controller.selectAccount}
        config={goPayAccountRenderConfig}
        pagination={controller.accountsPagination}
        renderChildren={(carrier) => <GoPayAccountMeta account={carrier} />}
      />
      <DetailDrawer open={!!controller.selected} title="GoPay账号详情" size="wide" onClose={controller.clearSelection}>
        {controller.selected && (
          <GoPayAccountDetails
            account={controller.selected}
            actionCatalog={actionCatalog}
            busy={busy}
            onRunAction={(spec, payload) => runAction(controller.selected!, spec, payload)}
          />
        )}
      </DetailDrawer>
    </>
  );
}

const goPayAccountRenderConfig: AccountRenderConfig = {
  icon: () => <WalletCards size={15} />,
  title: (record) => <span className="font-mono">{accountSubject(record) || record.key?.account_id}</span>,
  subtitle: (record) => record.key?.account_id || '',
  meta: (record) => <span className="text-xs text-muted-foreground">{record.status?.label || record.status?.value || '-'}</span>,
};

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
