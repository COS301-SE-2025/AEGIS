import React from 'react';
import {
  Bold,
  Italic,
  Underline,
  List,
  ListOrdered,
  Link,
  Image,
  Table,
  Code,
} from 'lucide-react';

interface ToolbarAction {
  type: 'bold' | 'italic' | 'underline' | 'bulletList' | 'orderedList' | 'link' | 'image' | 'table' | 'code';
  active?: boolean;
}

interface EditorToolbarProps {
  onAction: (action: ToolbarAction['type']) => void;
  activeFormats?: Set<string>;
}

const ToolbarButton = ({ 
  icon, 
  active = false, 
  onClick,
  tooltip 
}: { 
  icon: React.ReactNode; 
  active?: boolean; 
  onClick?: () => void;
  tooltip?: string;
}) => (
  <button 
    onClick={onClick}
    title={tooltip}
    className={`p-2 rounded hover:bg-gray-700 transition-colors ${
      active ? 'bg-gray-600 text-white' : 'text-gray-300'
    }`}
  >
    {icon}
  </button>
);

export const EditorToolbar: React.FC<EditorToolbarProps> = ({ 
  onAction, 
  activeFormats = new Set() 
}) => {
  const handleAction = (actionType: ToolbarAction['type']) => {
    onAction(actionType);
  };

  return (
    <div className="bg-gray-800 border-b border-gray-700 p-2">
      <div className="flex items-center gap-1">
        {/* Text formatting */}
        <ToolbarButton 
          icon={<Bold className="w-4 h-4" />} 
          active={activeFormats.has('bold')}
          onClick={() => handleAction('bold')}
          tooltip="Bold (Ctrl+B)"
        />
        <ToolbarButton 
          icon={<Italic className="w-4 h-4" />} 
          active={activeFormats.has('italic')}
          onClick={() => handleAction('italic')}
          tooltip="Italic (Ctrl+I)"
        />
        <ToolbarButton 
          icon={<Underline className="w-4 h-4" />} 
          active={activeFormats.has('underline')}
          onClick={() => handleAction('underline')}
          tooltip="Underline (Ctrl+U)"
        />
        
        {/* Separator */}
        <div className="w-px h-6 bg-gray-600 mx-2"></div>
        
        {/* Lists */}
        <ToolbarButton 
          icon={<List className="w-4 h-4" />} 
          active={activeFormats.has('bulletList')}
          onClick={() => handleAction('bulletList')}
          tooltip="Bullet List"
        />
        <ToolbarButton 
          icon={<ListOrdered className="w-4 h-4" />} 
          active={activeFormats.has('orderedList')}
          onClick={() => handleAction('orderedList')}
          tooltip="Numbered List"
        />
        
        {/* Separator */}
        <div className="w-px h-6 bg-gray-600 mx-2"></div>
        
        {/* Media and special content */}
        <ToolbarButton 
          icon={<Link className="w-4 h-4" />} 
          onClick={() => handleAction('link')}
          tooltip="Insert Link"
        />
        <ToolbarButton 
          icon={<Image className="w-4 h-4" />} 
          onClick={() => handleAction('image')}
          tooltip="Insert Image"
        />
        <ToolbarButton 
          icon={<Table className="w-4 h-4" />} 
          onClick={() => handleAction('table')}
          tooltip="Insert Table"
        />
        <ToolbarButton 
          icon={<Code className="w-4 h-4" />} 
          active={activeFormats.has('code')}
          onClick={() => handleAction('code')}
          tooltip="Code Block"
        />
      </div>
    </div>
  );
};