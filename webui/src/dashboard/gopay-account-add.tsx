import { AccountAddDialog, AccountPhoneFieldList, ControlledSelectField, accountCallingCodePrefix, accountPhoneSubmitDisabled } from '@byte-v-forge/common-ui';
import type { SelectFieldOption } from '@byte-v-forge/common-ui';
import { createGoPayAccount, type GoPayAccountProjection } from './gopay-api';

type GoPayAddAccountValues = {
  phone: string;
  country_calling_code: string;
  otp_channel: 'wa' | 'sms';
};

export function GoPayAccountAdd({ disabled, onCreated, onError }: {
  disabled?: boolean;
  onCreated: (account: GoPayAccountProjection) => void | Promise<void>;
  onError: (message: string) => void;
}) {
  return (
    <AccountAddDialog<GoPayAddAccountValues>
      formId="gopay-add-account-form"
      title="添加 GoPayAccount"
      description="输入手机号、国家拨号码并选择注册/登录使用的 OTP 渠道。"
      defaultValues={{ phone: '', country_calling_code: '', otp_channel: 'wa' }}
      disabled={disabled}
      submitDisabled={accountPhoneSubmitDisabled}
      onError={onError}
      onDone={(account) => onCreated(account as GoPayAccountProjection)}
      onSubmit={(values) => createGoPayAccount({ phone: values.phone, otp_channel: values.otp_channel, country_code: accountCallingCodePrefix(values.country_calling_code) })}
    >
      {(form) => (
        <>
          <AccountPhoneFieldList control={form.control} idPrefix="gopay-add" countryPlaceholder="+62" phonePlaceholder="812xxxx" />
          <ControlledSelectField control={form.control} name="otp_channel" label="OTP 渠道" options={goPayOTPChannelOptions} />
        </>
      )}
    </AccountAddDialog>
  );
}

const goPayOTPChannelOptions: SelectFieldOption[] = [
  { value: 'wa', label: 'WhatsApp' },
  { value: 'sms', label: 'SMS' },
];
