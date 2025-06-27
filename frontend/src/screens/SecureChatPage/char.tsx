import {
  Send,
  FileText,
  Folder,
  Home,
  MessageSquare,
} from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { useEffect, useState } from 'react';
import { useUserKeys } from "../../store/userKeys";
import { fetchBundle } from '../../services/x3dhService';
import { deriveSharedSecretInitiator } from "../../lib/crypto/x3dh";
import { encryptMessage, decryptMessage } from '../../lib/crypto/aes_gcm';

export const SecureChatPage = (): JSX.Element => {
  const { userId } = useParams();
  const { ik, loadPersistedKeys } = useUserKeys();
  const [recipientBundle, setRecipientBundle] = useState<any>(null);
  const [sharedKey, setSharedKey] = useState<Uint8Array | null>(null);
  const [messageInput, setMessageInput] = useState('');
  const [messages, setMessages] = useState<
    { user: string; content: string; encrypted?: boolean; self?: boolean }[]
  >([]);

  useEffect(() => {
    loadPersistedKeys();
  }, []);

  useEffect(() => {
    async function loadBundle() {
      if (userId) {
        const bundle = await fetchBundle(userId);
        setRecipientBundle(bundle);
      }
    }
    loadBundle();
  }, [userId]);

  useEffect(() => {
    async function deriveKey() {
      if (ik && recipientBundle) {
        const sharedSecret = await deriveSharedSecretInitiator(ik, recipientBundle);
        setSharedKey(sharedSecret);
      }
    }
    deriveKey();
  }, [ik, recipientBundle]);

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!sharedKey || !messageInput.trim()) return;

    const encrypted = await encryptMessage(messageInput, sharedKey);
    setMessages((prev) => [...prev, {
      user: "You",
      content: JSON.stringify(encrypted),
      encrypted: true,
      self: true,
    }]);
    setMessageInput('');
  }

  return (
    <div className="bg-black flex flex-row justify-center w-full min-h-screen text-white">
      {/* Sidebar */}
      <div className="w-72 bg-gray-900 p-6 flex flex-col justify-between">
        <div>
          <div className="flex items-center gap-3 mb-12">
            <div className="w-12 h-12 rounded-lg overflow-hidden">
              <img
                src="https://c.animaapp.com/mawlyxkuHikSGI/img/image-5.png"
                alt="AEGIS Logo"
                className="w-full h-full object-cover"
              />
            </div>
            <span className="font-bold text-white text-2xl">AEGIS</span>
          </div>

          {/* Navigation */}
          <nav className="space-y-4">
            <Link to="/dashboard">
              <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
                <Home className="w-5 h-5" /> Dashboard
              </button>
            </Link>
            <Link to="/case-management">
              <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
                <Folder className="w-5 h-5" /> Case Management
              </button>
            </Link>
            <Link to="/evidence-viewer">
              <button className="w-full flex items-center gap-3 text-left px-4 py-2 hover:bg-gray-800 rounded-lg">
                <FileText className="w-5 h-5" /> Evidence Viewer
              </button>
            </Link>
            <button className="w-full flex items-center gap-3 text-left px-4 py-2 bg-gray-800 hover:bg-gray-700 rounded-lg">
              <MessageSquare className="w-5 h-5" /> Secure Chat
            </button>
          </nav>
        </div>
      </div>

      {/* Chat Area */}
      <div className="flex-1 flex flex-col bg-gray-950 p-8">
        <h1 className="text-3xl font-bold mb-6">Secure Chat</h1>

  <div className="flex-1 overflow-y-auto space-y-4 mb-4 pr-2">
  {messages.map((msg, index) => {
    const [decryptedContent, setDecryptedContent] = useState<string | null>(null);

    useEffect(() => {
      async function decrypt() {
        if (msg.encrypted && sharedKey) {
          try {
            const { ciphertext, nonce } = JSON.parse(msg.content);
            const decrypted = await decryptMessage(new Uint8Array(ciphertext), new Uint8Array(nonce), sharedKey);
            setDecryptedContent(decrypted);
          } catch (e) {
            console.error("Decryption error:", e);
            setDecryptedContent("[Decryption failed]");
          }
        } else {
          setDecryptedContent(msg.content);
        }
      }

      decrypt();
    }, [msg, sharedKey]);

    return (
      <div
        key={index}
        className={`${msg.self ? "self-end ml-auto" : ""} max-w-md bg-gray-800 p-4 rounded-lg`}
      >
        <p>
          <span className="font-bold text-green-400">{msg.user}:</span>{" "}
          {decryptedContent ?? "[Decrypting...]"}
        </p>
      </div>
    );
  })}
</div>



        <form onSubmit={handleSend} className="flex items-center gap-2 mt-auto">
          <input
            value={messageInput}
            onChange={(e) => setMessageInput(e.target.value)}
            type="text"
            placeholder="Type a secure message..."
            className="flex-1 p-3 rounded-lg bg-gray-800 text-white border border-gray-700 placeholder-gray-400"
          />
          <button
            type="submit"
            className="px-4 py-3 bg-blue-600 hover:bg-blue-500 rounded-lg flex items-center justify-center"
          >
            <Send className="w-5 h-5" />
          </button>
        </form>
      </div>
    </div>
  );
};
