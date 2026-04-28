import React from 'react';

interface SidebarProps {
  activeMenu: string;
  setActiveMenu: (menu: string) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ activeMenu, setActiveMenu }) => {
  const menuItems = [
    { id: 'ap_aging', label: 'AP Aging Report', icon: '📊' },
    { isDivider: true },
    { id: 'items', label: 'Item Master', icon: '📦' },
    { id: 'vendors', label: 'Vendor Master', icon: '🤝' },
    { id: 'pos', label: 'Purchase Orders', icon: '📝' },
    { id: 'logistics', label: 'Logistics (Bookings)', icon: '🚢' },
    { id: 'bls', label: 'BL Management', icon: '📄', isSub: true },
    { id: 'containers', label: 'Container Master', icon: '📥', isSub: true },
    { id: 'invoices', label: 'Commercial Invoices', icon: '🧾' },
    { id: 'aps', label: 'Account Payables', icon: '💰' },
    { id: 'inventory', label: 'Landed Goods', icon: '🏭' },
    { id: 'allocations', label: 'Cost Allocations', icon: '⚖️' },
  ];

  return (
    <div className="sidebar">
      <div className="logo">
        <span>🐺</span> Cerberus Procure
      </div>
      <ul className="menu-list">
        {menuItems.map((item: any, idx) => {
          if (item.isDivider) {
            return <li key={`div-${idx}`} style={{ height: '1px', background: 'rgba(255,255,255,0.1)', margin: '1rem 0.5rem' }}></li>;
          }
          return (
            <li
              key={item.id}
              className={`menu-item ${activeMenu === item.id ? 'active' : ''} ${item.isSub ? 'sub-menu' : ''}`}
              style={item.isSub ? { paddingLeft: '2.5rem', fontSize: '0.9rem', opacity: 0.85 } : {}}
              onClick={() => setActiveMenu(item.id)}
            >
              <span>{item.icon}</span> {item.label}
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default Sidebar;
