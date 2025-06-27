import { useEffect, useState } from "react";
import {
  HelpCircle,
  X,
  BookOpen,
  MessageSquare,
  Info,
  ExternalLink,
} from "lucide-react";
import { Link } from "react-router-dom";

export const HelpMenu = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [hasMounted, setHasMounted] = useState(false); // ⬅️ prevent early render

  // Mark component as mounted to avoid hydration mismatch
  useEffect(() => {
    setHasMounted(true);
  }, []);

  // Close on Escape
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === "Escape") setIsOpen(false);
    };
    window.addEventListener("keydown", handleEsc);
    return () => window.removeEventListener("keydown", handleEsc);
  }, []);

  // Don’t render HelpMenu until mounted (avoids SSR hydration issues)
  if (!hasMounted) return null;

  return (
    <>
      {/* Floating Button */}
      <button
        onClick={() => setIsOpen(true)}
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
                onClick={() => setIsOpen(false)}
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
              <a
                href="https://support.aegis.com"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-3 hover:text-blue-400 transition"
              >
                <ExternalLink className="w-4 h-4 text-gray-400" />
                Visit Support Center
              </a>
            </div>
          </div>

          {/* Overlay */}
          <div
            className="fixed inset-0 bg-black/50 z-30"
            onClick={() => setIsOpen(false)}
          />
        </>
      )}
    </>
  );
};
