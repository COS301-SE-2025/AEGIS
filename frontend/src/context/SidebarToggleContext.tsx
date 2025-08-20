import { createContext, useContext, useState, useEffect } from 'react';

interface SidebarContextType {
  sidebarVisible: boolean;
  toggleSidebar: () => void;
  setSidebarVisible: (_: boolean) => void;
}

const SidebarContext = createContext<SidebarContextType | undefined>(undefined);

export const useSidebar = () => {
  const context = useContext(SidebarContext);
  if (!context) {
    throw new Error('useSidebar must be used within a SidebarProvider');
  }
  return context;
};

interface SidebarProviderProps {
  children: React.ReactNode;
}

export const SidebarProvider: React.FC<SidebarProviderProps> = ({ children }) => {
  const [sidebarVisible, setSidebarVisible] = useState(true);

  const toggleSidebar = () => {
    setSidebarVisible(prev => !prev);
  };

  // Apply CSS class to any sidebar found on the page
  useEffect(() => {
    // Look for common sidebar class patterns
    const sidebarSelectors = [
      '.sidebar-toggle-target',
      '[class*="fixed"][class*="left-0"]',
      '[class*="w-80"][class*="fixed"]'
    ];
    
    let sidebar: HTMLElement | null = null;
    
    for (const selector of sidebarSelectors) {
      sidebar = document.querySelector(selector) as HTMLElement;
      if (sidebar) break;
    }

    if (sidebar) {
      if (sidebarVisible) {
        sidebar.style.transform = 'translateX(0)';
        sidebar.style.transition = 'transform 0.3s ease-in-out';
        // Adjust main content margin
        const mainContent = sidebar.nextElementSibling as HTMLElement;
        if (mainContent && mainContent.classList.contains('ml-80')) {
          mainContent.style.marginLeft = '20rem'; // 80 * 0.25rem = 20rem
        }
      } else {
        sidebar.style.transform = 'translateX(-100%)';
        sidebar.style.transition = 'transform 0.3s ease-in-out';
        // Adjust main content margin
        const mainContent = sidebar.nextElementSibling as HTMLElement;
        if (mainContent && mainContent.classList.contains('ml-80')) {
          mainContent.style.marginLeft = '0';
        }
      }
    }
  }, [sidebarVisible]);

  return (
    <SidebarContext.Provider value={{ sidebarVisible, toggleSidebar, setSidebarVisible }}>
      {children}
    </SidebarContext.Provider>
  );
};

// SidebarToggleButton.tsx
import React from 'react';
import { Menu, X } from 'lucide-react';
//import { useSidebar } from './SidebarToggleContext';

interface SidebarToggleButtonProps {
  className?: string;
  style?: React.CSSProperties;
}

// export const SidebarToggleButton: React.FC<SidebarToggleButtonProps> = ({ 
//   className = '', 
//   style = {} 
// }) => {
//   const { sidebarVisible, toggleSidebar } = useSidebar();

//   const defaultStyle: React.CSSProperties = {
//     position: 'fixed',
//     top: '20px',
//     left: sidebarVisible ? '340px' : '20px', // 320px (sidebar width) + 20px padding
//     zIndex: 1000,
//     padding: '8px',
//     backgroundColor: '#636ae8',
//     color: 'white',
//     border: 'none',
//     borderRadius: '6px',
//     cursor: 'pointer',
//     transition: 'left 0.3s ease-in-out, background-color 0.2s ease',
//     boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
//     ...style
//   };

//   return (
//     <button
//       onClick={toggleSidebar}
//       className={`hover:bg-blue-700 ${className}`}
//       style={defaultStyle}
//       aria-label={sidebarVisible ? 'Hide Sidebar' : 'Show Sidebar'}
//     >
//       {sidebarVisible ? <X size={20} /> : <Menu size={20} />}
//     </button>
//   );
// };

export const SidebarToggleButton: React.FC<SidebarToggleButtonProps> = ({
  className = '',
  style = {},
}) => {
  const { sidebarVisible, toggleSidebar } = useSidebar();

  return (
    <button
      onClick={toggleSidebar}
      className={`ml-2 p-1 rounded bg-blue-600 text-white hover:bg-indigo-700 transition ${className}`}
      style={style}
      aria-label={sidebarVisible ? 'Hide Sidebar' : 'Show Sidebar'}
    >
      {sidebarVisible ? <X size={20} /> : <Menu size={20} />}
    </button>
  );
};
