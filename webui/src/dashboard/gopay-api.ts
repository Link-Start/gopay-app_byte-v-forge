import { ACCOUNT_PAGE_SIZE, api, fetchAccountList } from '@byte-v-forge/common-ui';
import type { AccountActionCatalog } from '@byte-v-forge/common-ui/proto/byte/v/forge/contracts/account/v1/account';
import type { GetGopayActionCatalogResponse, ListGopayAccountsResponse, GopayAccount } from '../proto/gopay_app';

export type GoPayAccountProjection = GopayAccount;

export type GoPayHealthResponse = {
  success?: boolean;
  ok?: boolean;
  service?: string;
  n8n_webhook_configured?: boolean;
  workflows?: Array<{ key: string; label: string; webhook_path: string }>;
};

export const goPayKeys = {
  health: ['gopay', 'health'] as const,
  accounts: ['gopay', 'accounts'] as const,
  actionCatalog: ['gopay', 'action-catalog'] as const
};

export function getGoPayHealth() {
  return api<GoPayHealthResponse>('/api/gopay/health');
}

export function getGoPayAccounts(cursor = '') {
  return fetchAccountList<GoPayAccountProjection, ListGopayAccountsResponse>({
    path: '/api/gopay/accounts',
    cursor,
    limit: ACCOUNT_PAGE_SIZE
  });
}

export async function getGoPayActionCatalog(): Promise<AccountActionCatalog | undefined> {
  const resp = await api<GetGopayActionCatalogResponse>('/api/gopay/action-catalog');
  if (!resp.success) throw new Error(resp.error_message || 'load GoPay action catalog failed');
  return resp.catalog;
}
