import React from 'react';
import { Hash, FileText, Loader2, AlertTriangle, Globe, Mail, Link as LinkIcon } from 'lucide-react';
import { MentionItem } from './useMentions';

interface MentionsDropdownProps {
  items: MentionItem[];
  selectedIndex: number;
  onSelect: (_item: MentionItem) => void;
  //onClose: () => void;
  loading?: boolean;
  position: { x: number; y: number };
  query: string;
}

const getIOCIcon = (type: string) => {
  switch (type.toLowerCase()) {
    case 'ip':
      return <Globe className="w-4 h-4 text-blue-400" />;
    case 'domain':
      return <Globe className="w-4 h-4 text-green-400" />;
    case 'hash':
    case 'md5':
    case 'sha1':
    case 'sha256':
      return <Hash className="w-4 h-4 text-purple-400" />;
    case 'url':
      return <LinkIcon className="w-4 h-4 text-orange-400" />;
    case 'email':
      return <Mail className="w-4 h-4 text-yellow-400" />;
    case 'file':
    case 'filename':
      return <FileText className="w-4 h-4 text-red-400" />;
    default:
      return <AlertTriangle className="w-4 h-4 text-gray-400" />;
  }
};

// Remove the getThreatLevelColor function since it's no longer needed
// const getThreatLevelColor = (level?: string) => { ... }

export const MentionsDropdown: React.FC<MentionsDropdownProps> = ({
  items,
  selectedIndex,
  onSelect,
  //onClose,
  loading = false,
  position,
  query
}) => {
  if (loading) {
    return (
      <div 
        className="absolute z-50 bg-gray-800 border border-gray-600 rounded-lg shadow-xl p-4 min-w-80"
        style={{ 
          left: position.x, 
          top: position.y + 20,
          maxHeight: '300px'
        }}
      >
        <div className="flex items-center justify-center gap-2 text-gray-300">
          <Loader2 className="w-4 h-4 animate-spin" />
          <span>Loading mentions...</span>
        </div>
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div 
        className="absolute z-50 bg-gray-800 border border-gray-600 rounded-lg shadow-xl p-4 min-w-80"
        style={{ 
          left: position.x, 
          top: position.y + 20,
          maxHeight: '300px'
        }}
      >
        <div className="text-gray-400 text-center">
          {query ? `No matches found for "${query}"` : 'No IOCs or evidence available'}
        </div>
      </div>
    );
  }

  return (
    <div 
      className="absolute z-50 bg-gray-800 border border-gray-600 rounded-lg shadow-xl overflow-hidden min-w-80"
      style={{ 
        left: position.x, 
        top: position.y + 20,
        maxHeight: '300px'
      }}
    >
      <div className="max-h-72 overflow-y-auto">
        {items.map((item, index) => (
          <div
            key={item.id}
            className={`px-4 py-3 cursor-pointer border-b border-gray-700 last:border-b-0 transition-colors ${
              index === selectedIndex 
                ? 'bg-blue-600 text-white' 
                : 'hover:bg-gray-700 text-gray-200'
            }`}
            onClick={() => onSelect(item)}
          >
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 mt-1">
                {item.type === 'ioc' ? (
                  getIOCIcon((item.data as any).type)
                ) : (
                  <FileText className="w-4 h-4 text-gray-400" />
                )}
              </div>
              
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium truncate">
                    {item.display}
                  </span>
                </div>
                
                <div className={`text-sm truncate mt-1 ${
                  index === selectedIndex ? 'text-gray-200' : 'text-gray-400'
                }`}>
                  {item.subtitle}
                </div>
                
                {item.type === 'evidence' && (
                  <div className={`text-xs mt-1 ${
                    index === selectedIndex ? 'text-gray-300' : 'text-gray-500'
                  }`}>
                    Checksum: {(item.data as any).checksum?.substring(0, 16)}...
                  </div>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
      
      <div className="px-4 py-2 bg-gray-900 border-t border-gray-600">
        <div className="flex items-center justify-between text-xs text-gray-400">
          <span>Use ↑↓ to navigate, Enter to select</span>
          <span>{items.length} results</span>
        </div>
      </div>
    </div>
  );
};