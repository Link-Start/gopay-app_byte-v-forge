import type { ReactNode } from 'react';
import { AccountPhoneProbeForm } from '@byte-v-forge/common-ui';
import { resolveGoPayPhone, type GoPayResolvedPhone } from './gopay-phone-utils';

export function GoPayPhoneCheckForm({ disabled, resultSlot, onCheck, onError }: {
  disabled?: boolean;
  resultSlot?: ReactNode;
  onCheck: (target: GoPayResolvedPhone) => void | Promise<void>;
  onError: (message: string) => void;
}) {
  return (
    <AccountPhoneProbeForm
      title="GoPay 号码检测"
      disabled={disabled}
      resultSlot={resultSlot}
      emptyResultText="结果：是否已注册"
      countryPlaceholder="+62"
      phonePlaceholder="81234567890"
      actionLabel="检测 GoPay 号码状态"
      resolve={(values) => ({ target: resolveGoPayPhone(values.phone, values.country_calling_code), error: '请输入手机号和国家拨号码，例如手机号 81234567890、拨号码 62。' })}
      onSubmit={onCheck}
      onError={onError}
    />
  );
}
