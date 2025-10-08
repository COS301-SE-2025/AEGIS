import React from 'react';
// import { Reply, CheckCircle, ThumbsUp } from 'lucide-react';
import { Reply, CheckCircle} from 'lucide-react';

interface ThreadMessage {
  id: string;
  threadID: string;
  parentMessageID?: string | null;
  userID: string;
  message: string;
  isApproved?: boolean;
  approvedBy?: string | null;
  approvedAt?: string | null;
  createdAt: string;
  updatedAt: string;
  mentions: { messageID: string; mentionedUserID: string; createdAt: string }[];
  reactions: { id: string; messageID: string; userID: string; reaction: string; createdAt: string }[];
  replies?: ThreadMessage[];
}

interface AnnotationThread {
  id: string;
  title: string;
  user: string;
  avatar: string;
  time: string;
  messageCount: number;
  participantCount: number;
  isActive?: boolean;
  status: 'open' | 'resolved' | 'pending_approval';
  priority: 'high' | 'medium' | 'low';
  tags: any[];
  fileId: string;
  createdBy?: string;
}

interface MessageCardProps {
  message: ThreadMessage;
  user: any;
  replyingToMessageId: string | null;
  setReplyingToMessageId: (_id: string | null) => void;
  replyText: string;
  setReplyText: (_text: string) => void;
  showReactionPicker: string | null;
  setShowReactionPicker: (_id: string | null) => void;
  selectedThread: AnnotationThread;
  onSendMessage: (_text: string, _parentId?: string) => Promise<void>;
  onAddReaction: (_messageId: string, _emoji: string) => Promise<void>;
  onApproveMessage: (_messageId: string) => Promise<void>;
  profile: { name: string };
}

const timeAgo = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  const intervals: [number, string][] = [
    [60, "seconds"],
    [3600, "minutes"],
    [86400, "hours"],
    [604800, "days"]
  ];

  for (let [limit, label] of intervals.reverse()) {
    const value = Math.floor(seconds / limit);
    if (value >= 1) return `${value} ${label} ago`;
  }

  return "just now";
};

export const MessageCard: React.FC<MessageCardProps> = ({
  message,
  user,
  replyingToMessageId,
  setReplyingToMessageId,
  replyText,
  setReplyText,
  showReactionPicker,
  setShowReactionPicker,
  selectedThread,
  onSendMessage,
  onAddReaction,
  onApproveMessage,
  profile
}) => {
  const handleReplyClick = () => {
    setReplyingToMessageId(replyingToMessageId === message.id ? null : message.id);
  };

  const handleSendReply = async () => {
    if (!replyText.trim()) return;
    await onSendMessage(replyText, message.id);
    setReplyText('');
    setReplyingToMessageId(null);
  };

  const handleReactionClick = (emoji: string) => {
    onAddReaction(message.id, emoji);
    setShowReactionPicker(null);
  };

  return (
    <div className="space-y-2">
      <div className="flex gap-3">
        <div className="w-8 h-8 bg-muted rounded-full flex items-center justify-center text-xs font-medium flex-shrink-0">
          {profile.name.split(" ").map((n: string) => n[0]).join("").toUpperCase()}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="font-medium text-sm">{profile.name}</span>
            <span className="text-xs text-muted-foreground">{timeAgo(message.createdAt)}</span>
            {message.isApproved === false && (
              <span className="px-2 py-0.5 bg-yellow-600/20 text-yellow-400 text-xs rounded">
                Pending Approval
              </span>
            )}
            {message.isApproved === true && (
              <CheckCircle className="w-3 h-3 text-green-400" />
            )}
          </div>
          <div className="text-sm text-muted-foreground mb-2">{message.message}</div>
          
          {/* Reactions */}
          {message.reactions && message.reactions.length > 0 && (
            <div className="flex items-center gap-2 mb-2">
              {message.reactions.map((reaction, index) => (
                <button key={index} className="flex items-center gap-1 px-2 py-1 bg-muted rounded-full text-xs hover:bg-muted">
                  <span>{reaction.reaction}</span>
                  <span className="text-muted-foreground">1</span>
                </button>
              ))}
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex items-center gap-3 text-xs">
            <button
              className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
              onClick={handleReplyClick}
            >
              <Reply className="w-3 h-3" />
              Reply
            </button>

            <div className="relative">
              <button 
                className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
                onClick={() => setShowReactionPicker(showReactionPicker === message.id ? null : message.id)}
              >
                <span className="text-sm">😊</span>
                React
              </button>
              
              {showReactionPicker === message.id && (
                <div className="absolute bottom-full left-0 mb-1 bg-card border border-border rounded-lg p-2 shadow-lg z-10" onClick={(e) => e.stopPropagation()}>
                  <div className="flex gap-1">
                    {['👍', '❤️', '😂', '😮', '😢', '🎉'].map(emoji => (
                      <button
                        key={emoji}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleReactionClick(emoji);
                        }}
                        className="p-2 hover:bg-muted rounded text-lg transition-colors"
                      >
                        {emoji}
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {message.isApproved === false && (
              <button
                className="text-green-400 hover:text-green-300"
                onClick={() => onApproveMessage(message.id)}
              >
                Approve
              </button>
            )}
          </div>

          {/* Reply Input */}
          {replyingToMessageId === message.id && (
            <div className="mt-2 ml-1">
              <input
                type="text"
                value={replyText}
                onChange={(e) => setReplyText(e.target.value)}
                placeholder="Type your reply..."
                className="w-full bg-muted text-foreground text-sm px-3 py-2 rounded border border-border focus:outline-none focus:ring-2 focus:ring-primary/60"
                onKeyPress={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    handleSendReply();
                  }
                }}
              />
              <button
                className="mt-1 px-3 py-1 bg-primary text-white text-xs rounded hover:bg-primmary/60"
                onClick={handleSendReply}
              >
                Send Reply
              </button>
            </div>
          )}
        </div>
      </div>
      
      {/* Simplified Replies Display */}
      {message.replies && message.replies.length > 0 && (
        <div className="ml-8 pl-4 border-l-2 border-muted mt-2">
          {message.replies.map((reply) => (
            <MessageCard
              key={reply.id}
              message={reply}
              user={user}
              replyingToMessageId={replyingToMessageId}
              setReplyingToMessageId={setReplyingToMessageId}
              replyText={replyText}
              setReplyText={setReplyText}
              showReactionPicker={showReactionPicker}
              setShowReactionPicker={setShowReactionPicker}
              selectedThread={selectedThread}
              onSendMessage={onSendMessage}
              onAddReaction={onAddReaction}
              onApproveMessage={onApproveMessage}
              profile={profile}
            />
          ))}
        </div>
      )}
    </div>
  );
};