import { useEffect } from 'react';
import { ShieldCheck } from 'lucide-react';
import { useForm } from 'react-hook-form';
import {
  ActionButtonGroup,
  ControlledInputField,
  DashboardDialog,
  type ActionButtonDescriptor,
} from '@byte-v-forge/common-ui';

type GoPayPINValues = {
  pin: string;
};

export function GoPayAccountPINDialog({
  open,
  busy,
  onSubmit,
  onOpenChange,
}: {
  open: boolean;
  busy?: boolean;
  onSubmit: (values: GoPayPINValues) => void | Promise<void>;
  onOpenChange: (open: boolean) => void;
}) {
  const form = useForm<GoPayPINValues>({ defaultValues: { pin: '' } });
  const pin = form.watch('pin').trim();

  useEffect(() => {
    if (open) form.reset({ pin: '' });
  }, [form, open]);

  async function submit(values: GoPayPINValues) {
    await onSubmit({ pin: values.pin.trim() });
  }

  return (
    <DashboardDialog
      open={open}
      title="GoPay PIN 设置"
      description="输入要写入 GoPay 的 PIN；PIN 只会提交给 gopay-app 编排，不在前端展示。"
      size="sm"
      footer={<ActionButtonGroup actions={footerActions(Boolean(busy), pin, () => onOpenChange(false))} />}
      onOpenChange={onOpenChange}
    >
      <form id="gopay-pin-action-form" className="grid gap-3" onSubmit={form.handleSubmit(submit)}>
        <ControlledInputField
          control={form.control}
          name="pin"
          label="PIN"
          inputId="gopay-account-action-pin"
          inputMode="numeric"
          autoComplete="off"
          placeholder="6 位 PIN"
          type="password"
        />
      </form>
    </DashboardDialog>
  );
}

function footerActions(busy: boolean, pin: string, onCancel: () => void): ActionButtonDescriptor[] {
  return [
    { id: 'cancel', label: '取消', variant: 'outline', disabled: busy, onClick: onCancel },
    {
      id: 'submit',
      label: busy ? '提交中' : '提交',
      icon: <ShieldCheck className="size-4" />,
      type: 'submit',
      form: 'gopay-pin-action-form',
      disabled: busy || !pin,
    },
  ];
}
