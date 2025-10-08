import React, { useState, useRef, useEffect, useCallback } from 'react';
import ReactQuill from 'react-quill';
import 'react-quill/dist/quill.snow.css';
//import { At } from 'lucide-react';
import { useMentions, MentionItem } from './useMentions';
import { MentionsDropdown } from './MentionsDropdown';
//import "./quill_dark.css";

interface EnhancedEditorProps {
  value: string;
  onChange: (_value: string) => void;
  placeholder?: string;
  caseId: string;
  apiBaseUrl?: string;
}

export const EnhancedEditor: React.FC<EnhancedEditorProps> = ({
  value,
  onChange,
  placeholder = "Start writing your report content here...",
  caseId,
  apiBaseUrl = 'https://localhost/api/v1'
}) => {
  const quillRef = useRef<ReactQuill>(null);
  const [showMentions, setShowMentions] = useState(false);
  const [mentionQuery, setMentionQuery] = useState('');
  const [mentionPosition, setMentionPosition] = useState({ x: 0, y: 0 });
  const [selectedMentionIndex, setSelectedMentionIndex] = useState(0);
  const [mentionStartIndex, setMentionStartIndex] = useState(-1);

  const { getMentionItems, loading } = useMentions({ caseId, apiBaseUrl });

  // Custom toolbar with mentions button
  const modules = {
    toolbar: {
      container: [
        [{ 'header': [1, 2, 3, false] }],
        ['bold', 'italic', 'underline', 'strike'],
        [{ 'color': [] }, { 'background': [] }],
        [{ 'list': 'ordered'}, { 'list': 'bullet' }],
        [{ 'indent': '-1'}, { 'indent': '+1' }],
        ['link', 'image', 'code-block'],
        [{ 'align': [] }],
        ['mentions'], // Custom mentions button
        ['clean']
      ],
      handlers: {
        mentions: () => {
          insertMentionTrigger();
        }
      }
    },
  };

  const formats = [
    'header', 'bold', 'italic', 'underline', 'strike', 
    'color', 'background', 'list', 'bullet', 'indent',
    'link', 'image', 'code-block', 'align'
  ];

  // Insert @ symbol to trigger mentions
  const insertMentionTrigger = useCallback(() => {
    const quill = quillRef.current?.getEditor();
    if (!quill) return;

    const range = quill.getSelection(true);
        if (range) {
        quill.insertText(range.index, '@', 'user');
        quill.setSelection({
            index: range.index + 1,
            length: 0
        });
        }
  }, []);

  // Handle text changes and detect mentions
  const handleChange = useCallback((content: string) => {
    onChange(content);
    
    const quill = quillRef.current?.getEditor();
    if (!quill) return;

    const selection = quill.getSelection();
    if (!selection) return;

    const text = quill.getText();
    const cursorPosition = selection.index;
    
    // Look for @ symbol before cursor
    let mentionStart = -1;
    let currentQuery = '';
    
    for (let i = cursorPosition - 1; i >= 0; i--) {
      const char = text[i];
      if (char === '@') {
        mentionStart = i;
        currentQuery = text.substring(i + 1, cursorPosition);
        break;
      }
      if (char === ' ' || char === '\n') {
        break;
      }
    }

    if (mentionStart !== -1 && currentQuery.length >= 0) {
      // Get cursor position for dropdown placement
      const bounds = quill.getBounds(mentionStart);
     // const editorRect = quill.container.getBoundingClientRect();
      const editorRect = quill.root.getBoundingClientRect();
      
      if (bounds) {
        setMentionPosition({
          x: editorRect.left + bounds.left,
          y: editorRect.top + bounds.top + bounds.height
        });
      }
      
      setMentionStartIndex(mentionStart);
      setMentionQuery(currentQuery);
      setShowMentions(true);
      setSelectedMentionIndex(0);
    } else {
      setShowMentions(false);
    }
  }, [onChange]);

  // Handle mention selection
  const handleMentionSelect = useCallback((item: MentionItem) => {
    const quill = quillRef.current?.getEditor();
    if (!quill || mentionStartIndex === -1) return;

    const selection = quill.getSelection();
    if (!selection) return;

    // Calculate the length to replace (@ + query)
    const replaceLength = mentionQuery.length + 1; // +1 for @
    
    // Create mention format
    let mentionText = '';
    let mentionHtml = '';
    
    if (item.type === 'ioc') {
      const ioc = item.data as any;
      mentionText = `@${ioc.type}:${ioc.value}`;
      mentionHtml = `<span class="mention mention-ioc" data-id="${ioc.id}" data-type="ioc" data-value="${ioc.value}" style="background-color: #3B82F6; color: white; padding: 2px 6px; border-radius: 4px; font-weight: 500;">${mentionText}</span>`;
    } else {
      const evidence = item.data as any;
      mentionText = `@evidence:${evidence.filename}`;
      mentionHtml = `<span class="mention mention-evidence" data-id="${evidence.id}" data-type="evidence" data-filename="${evidence.filename}" style="background-color: #059669; color: white; padding: 2px 6px; border-radius: 4px; font-weight: 500;">${mentionText}</span>`;
    }

    // Replace the @ and query with the mention
    quill.deleteText(mentionStartIndex, replaceLength);
    quill.clipboard.dangerouslyPasteHTML(mentionStartIndex, mentionHtml + '&nbsp;');
    
    // Move cursor after the mention
    // const newPosition = mentionStartIndex + mentionText.length + 1;
    // setTimeout(() => {
    //   quill.setSelection(newPosition);
    // }, 0);

    const newPosition = mentionStartIndex + mentionText.length + 1;
    setTimeout(() => {
    quill.setSelection({ index: newPosition, length: 0 });
    }, 0);

    setShowMentions(false);
    setMentionQuery('');
    setMentionStartIndex(-1);
  }, [mentionStartIndex, mentionQuery]);

  // Handle keyboard navigation in mentions dropdown
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!showMentions) return;

      const items = getMentionItems(mentionQuery);
      
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedMentionIndex(prev => 
          prev < items.length - 1 ? prev + 1 : 0
        );
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedMentionIndex(prev => 
          prev > 0 ? prev - 1 : items.length - 1
        );
      } else if (e.key === 'Enter') {
        e.preventDefault();
        if (items[selectedMentionIndex]) {
          handleMentionSelect(items[selectedMentionIndex]);
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        setShowMentions(false);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [showMentions, mentionQuery, selectedMentionIndex, getMentionItems, handleMentionSelect]);

  // Close mentions when clicking outside
  useEffect(() => {
    const handleClickOutside = () => {
      setShowMentions(false);
    };

    if (showMentions) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [showMentions]);

  // Add custom mentions button to toolbar and inject styles
  useEffect(() => {
    const toolbar = document.querySelector('.ql-toolbar');
    if (toolbar) {
      // Remove existing mentions button
      const existingButton = toolbar.querySelector('.ql-mentions');
      if (existingButton) {
        existingButton.remove();
      }

      // Add mentions button
      const mentionsButton = document.createElement('button');
      mentionsButton.className = 'ql-mentions';
      mentionsButton.innerHTML = '<svg viewBox="0 0 18 18"><circle cx="9" cy="9" r="8" fill="none" stroke="currentColor" stroke-width="1.5"/><text x="9" y="13" text-anchor="middle" font-size="10" fill="currentColor">@</text></svg>';
      mentionsButton.title = 'Insert Mention (@)';
      
      // Insert before clean button
      const cleanButton = toolbar.querySelector('.ql-clean');
      if (cleanButton) {
        toolbar.insertBefore(mentionsButton, cleanButton);
      } else {
        toolbar.appendChild(mentionsButton);
      }
    }

    // Inject custom styles
    const styleId = 'enhanced-editor-styles';
    if (!document.getElementById(styleId)) {
      const style = document.createElement('style');
      style.id = styleId;
      style.textContent = `
        .ql-snow {
          border: 1px solid #374151 !important;
          background-color: #1f2937 !important;
        }
        
        .ql-snow .ql-toolbar {
          border-bottom: 1px solid #374151 !important;
          background-color: #1f2937 !important;
        }
        
        .ql-snow .ql-container {
          border-top: none !important;
          background-color: #1f2937 !important;
        }
        
        .ql-editor {
          color: #e5e7eb !important;
          background-color: #1f2937 !important;
          min-height: 300px !important;
          font-size: 16px !important;
          line-height: 1.6 !important;
        }
        
        .ql-editor.ql-blank::before {
          color: #6b7280 !important;
          font-style: italic;
        }
        
        .ql-snow .ql-tooltip {
          background-color: #374151 !important;
          border: 1px solid #4b5563 !important;
          color: #e5e7eb !important;
        }
        
        .ql-snow .ql-tooltip input {
          background-color: #1f2937 !important;
          color: #e5e7eb !important;
          border: 1px solid #4b5563 !important;
        }
        
        .ql-snow .ql-picker-options {
          background-color: #374151 !important;
          border: 1px solid #4b5563 !important;
        }
        
        .ql-snow .ql-picker-item:hover {
          background-color: #4b5563 !important;
          color: #e5e7eb !important;
        }
        
        .ql-snow .ql-stroke {
          stroke: #9ca3af !important;
        }
        
        .ql-snow .ql-fill {
          fill: #9ca3af !important;
        }
        
        .ql-snow .ql-picker-label:hover .ql-stroke,
        .ql-snow .ql-picker-label.ql-active .ql-stroke {
          stroke: #e5e7eb !important;
        }
        
        .ql-snow .ql-picker-label:hover .ql-fill,
        .ql-snow .ql-picker-label.ql-active .ql-fill {
          fill: #e5e7eb !important;
        }

        .ql-mentions {
          width: 28px !important;
          height: 28px !important;
        }

        .ql-mentions svg {
          width: 18px;
          height: 18px;
        }

        .mention {
          display: inline-block;
          margin: 0 2px;
          cursor: pointer;
          user-select: none;
        }

        .mention-ioc {
          background-color: #3B82F6 !important;
        }

        .mention-evidence {
          background-color: #059669 !important;
        }

        .mention:hover {
          opacity: 0.8;
        }
      `;
      document.head.appendChild(style);
    }

    // Cleanup function
    return () => {
      const existingStyle = document.getElementById(styleId);
      if (existingStyle) {
        existingStyle.remove();
      }
    };
  }, []);

  return (
    <div className="relative">
      <ReactQuill
        ref={quillRef}
        theme="snow"
        value={value}
        onChange={handleChange}
        modules={modules}
        formats={formats}
        placeholder={placeholder}
      />
      
      {showMentions && (
        <MentionsDropdown
          items={getMentionItems(mentionQuery)}
          selectedIndex={selectedMentionIndex}
          onSelect={handleMentionSelect}
          onClose={() => setShowMentions(false)}
          loading={loading}
          position={mentionPosition}
          query={mentionQuery}
        />
      )}
    </div>
  );
};