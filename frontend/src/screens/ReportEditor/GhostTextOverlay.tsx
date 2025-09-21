
import React, { useEffect } from "react";

interface GhostTextOverlayProps {
  suggestion: string;
  editorId: string;
  onAccept: () => void;
}

export const GhostTextOverlay: React.FC<GhostTextOverlayProps> = ({ suggestion, editorId, onAccept }) => {
  useEffect(() => {
    if (!suggestion) return;
    const quillEditor = document.querySelector(`#${editorId} .ql-editor`);
    if (quillEditor) {
      // Remove any previous ghost text
      const prevGhost = quillEditor.querySelector(".ghost-text-inline");
      if (prevGhost) prevGhost.remove();

      // Find last line (or fallback to editor)
      const lastLine = quillEditor.lastChild as HTMLElement | null;
      const target = lastLine || quillEditor;

      // Create ghost text span
      const ghostSpan = document.createElement("span");
      ghostSpan.className = "ghost-text-inline";
      ghostSpan.textContent = suggestion;
      ghostSpan.style.color = "#6b7280";
      ghostSpan.style.opacity = "0.7";
      ghostSpan.style.fontStyle = "italic";
      ghostSpan.style.fontSize = "16px";
      ghostSpan.style.marginLeft = "8px";
      ghostSpan.style.pointerEvents = "none";
      ghostSpan.style.background = "none";

      target.appendChild(ghostSpan);

      // Animate fade-in
      ghostSpan.style.transition = "opacity 0.4s";
      ghostSpan.style.opacity = "0";
      setTimeout(() => {
        ghostSpan.style.opacity = "0.7";
      }, 10);

      // Cleanup on unmount or suggestion change
      return () => {
        if (ghostSpan && ghostSpan.parentNode) ghostSpan.parentNode.removeChild(ghostSpan);
      };
    }
  }, [suggestion, editorId]);

  // Keyboard shortcut: Tab to accept suggestion
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Tab" && suggestion) {
        e.preventDefault();
        onAccept();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [suggestion, onAccept]);

  // Accept button (inline, only if suggestion exists)
  if (!suggestion) return null;
  return (
    <button
      type="button"
      onClick={onAccept}
      style={{
        position: "absolute",
        right: 24,
        bottom: 24,
        zIndex: 3,
        padding: "6px 12px",
        background: "#16a34a",
        color: "white",
        borderRadius: "6px",
        boxShadow: "0 2px 8px rgba(0,0,0,0.12)",
        fontSize: "15px",
        border: "none",
        cursor: "pointer",
        transition: "background 0.2s",
      }}
    >
      Accept Suggestion
    </button>
  );
};