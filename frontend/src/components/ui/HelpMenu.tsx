import { useEffect, useState } from "react";
import {
  HelpCircle,
  X,
  BookOpen,
  MessageSquare,
  Info,
} from "lucide-react";
import { Link } from "react-router-dom";

export const HelpMenu = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [hasMounted, setHasMounted] = useState(false);

  // Load state from localStorage on mount
  useEffect(() => {
    setHasMounted(true);
    const savedState = localStorage.getItem('helpMenuOpen');
    if (savedState === 'true') {
      setIsOpen(true);
    }
  }, []);

  // Save state to localStorage whenever it changes
  useEffect(() => {
    if (hasMounted) {
      localStorage.setItem('helpMenuOpen', isOpen.toString());
    }
  }, [isOpen, hasMounted]);

  const handleOpen = () => setIsOpen(true);
  const handleClose = () => setIsOpen(false);

  // Don't render HelpMenu until mounted (avoids SSR hydration issues)
  if (!hasMounted) return null;

  return (
    <>
      {/* Floating Button */}
      <button
        onClick={handleOpen}
        className="fixed bottom-6 right-6 z-50 bg-blue-600 hover:bg-blue-700 text-white rounded-full p-3 shadow-lg transition"
        aria-label="Help"
      >
        <HelpCircle className="w-6 h-6" />
      </button>

      {/* Conditional Drawer Render */}
      {isOpen && (
        <>
          {/* Drawer */}
          <div className="fixed top-0 right-0 h-full w-80 bg-gray-900 text-white border-l border-gray-700 shadow-lg z-40 animate-slide-in">
            {/* Header */}
            <div className="flex justify-between items-center px-5 py-4 border-b border-gray-700">
              <div className="flex items-center gap-2">
                <HelpCircle className="w-5 h-5 text-blue-400" />
                <h3 className="text-lg font-bold">Help & Support</h3>
              </div>
              <button
                onClick={handleClose}
                className="text-gray-400 hover:text-white"
                aria-label="Close Help"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Content */}
            <div className="p-5 space-y-4 text-sm">
              <Link
                to="/tutorials"
                className="flex items-center gap-3 hover:text-blue-400 transition"
              >
                <BookOpen className="w-4 h-4 text-blue-500" />
                Tutorials & Getting Started
              </Link>
              <Link
                to="/faq"
                className="flex items-center gap-3 hover:text-blue-400 transition"
              >
                <MessageSquare className="w-4 h-4 text-green-400" />
                Frequently Asked Questions
              </Link>
              <Link
                to="/about"
                className="flex items-center gap-3 hover:text-blue-400 transition"
              >
                <Info className="w-4 h-4 text-yellow-400" />
                About AEGIS
              </Link>
            </div>
          </div>

          {/* Overlay */}
          <div className="fixed inset-0 bg-black/50 z-30" />
        </>
      )}
    </>
  );
};
