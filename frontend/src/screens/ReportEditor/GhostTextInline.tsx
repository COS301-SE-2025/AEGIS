import { useEffect } from "react";

// GhostTextInline: inserts faded suggestion span directly into editor content at caret
export function GhostTextInline({ suggestion, quillRef }: { suggestion: string, quillRef: any }) {
  useEffect(() => {
    const quill = quillRef?.current;
    if (!quill) return;
    const editor = quill.getEditor?.() || quill.editor;
    if (!editor) return;
    const qlEditor = editor.root;
    if (!qlEditor) return;

    // Remove ALL ghost text spans
    Array.from(qlEditor.querySelectorAll('.ghost-text-inline')).forEach(node => (node as HTMLElement).remove());

    if (!suggestion) return;

    // Get caret position
    const range = editor.getSelection(true);
    let caretNode = null;
    let caretOffset = 0;
    if (range) {
      // Find the node at the caret
      const [leaf, offset] = editor.getLeaf(range.index);
      caretNode = leaf && leaf.domNode;
      caretOffset = offset;
    }

    // Create ghost text span
    const ghostSpan = document.createElement("span");
    ghostSpan.className = "ghost-text-inline";
    ghostSpan.textContent = suggestion;
    ghostSpan.style.pointerEvents = "none";
    ghostSpan.style.color = "#9ca3af";
    ghostSpan.style.fontStyle = "italic";
    ghostSpan.style.opacity = "0.7";
    ghostSpan.style.fontSize = "16px";
    ghostSpan.style.paddingLeft = "2px";
    ghostSpan.setAttribute("aria-hidden", "true");

    // Insert ghost text inline at caret
    if (caretNode && caretNode.nodeType === Node.TEXT_NODE) {
      // Split text node at caret
      const text = caretNode.textContent || "";
      const before = text.slice(0, caretOffset);
      const after = text.slice(caretOffset);
      const beforeNode = document.createTextNode(before);
      const afterNode = document.createTextNode(after);
      const parent = caretNode.parentNode;
      if (parent) {
        parent.insertBefore(beforeNode, caretNode);
        parent.insertBefore(ghostSpan, caretNode);
        parent.insertBefore(afterNode, caretNode);
        parent.removeChild(caretNode);
      }
    } else if (caretNode) {
      // If not a text node, just append after
      caretNode.parentNode?.insertBefore(ghostSpan, caretNode.nextSibling);
    } else {
      // Fallback: append to end of editor
      qlEditor.appendChild(ghostSpan);
    }

    // Cleanup on unmount or suggestion change
    return () => {
      if (ghostSpan && ghostSpan.parentNode) ghostSpan.parentNode.removeChild(ghostSpan);
    };
  }, [suggestion, quillRef]);

  // No React rendering; DOM is managed directly for robust inline ghost text
  return null;
}