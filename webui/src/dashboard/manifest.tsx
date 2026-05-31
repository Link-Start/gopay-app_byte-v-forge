import { WalletCards } from 'lucide-react';
import { DashboardNavSection, type DashboardModuleRegistration } from '@byte-v-forge/common-ui';
import { GoPayPage } from './gopay-page';

const registration: DashboardModuleRegistration = {
  manifest: {
    id: 'gopay-app',
    nav: [
      {
        key: 'gopay',
        label: 'GoPay',
        icon: 'gopay-app',
        section: DashboardNavSection.DASHBOARD_NAV_SECTION_MAIN,
        required_services: ['gopay-app'],
        order: 16
      }
    ]
  },
  icons: { 'gopay-app': <WalletCards size={17} /> },
  views: { gopay: () => <GoPayPage /> }
};

export default registration;
