import { AccountAddDialog, ControlledInputFieldList, ControlledSelectField } from '@byte-v-forge/common-ui';
import type { ControlledInputFieldDescriptor, SelectFieldOption } from '@byte-v-forge/common-ui';
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
      submitDisabled={(values) => !values.phone.trim() || !values.country_calling_code.replace(/\D+/g, '')}
      onError={onError}
      onDone={(account) => onCreated(account as GoPayAccountProjection)}
      onSubmit={(values) => createGoPayAccount({ phone: values.phone, otp_channel: values.otp_channel, country_code: `+${values.country_calling_code.replace(/\D+/g, '')}` })}
    >
      {(form) => (
        <>
          <ControlledInputFieldList control={form.control} fields={goPayAddFields} />
          <ControlledSelectField control={form.control} name="otp_channel" label="OTP 渠道" options={goPayOTPChannelOptions} />
        </>
      )}
    </AccountAddDialog>
  );
}

const goPayAddFields: ControlledInputFieldDescriptor<GoPayAddAccountValues>[] = [{
  id: 'country_calling_code',
  name: 'country_calling_code',
  label: '拨号码',
  placeholder: '+62',
  inputId: 'gopay-add-country-calling-code',
}, {
  id: 'phone',
  name: 'phone',
  label: '手机号',
  placeholder: '812xxxx',
  inputId: 'gopay-add-phone',
}];

const goPayOTPChannelOptions: SelectFieldOption[] = [
  { value: 'wa', label: 'WhatsApp' },
  { value: 'sms', label: 'SMS' },
];
