import { useState, useEffect, useCallback } from "react";
import { Plus, Calendar, Clock, Paperclip, Tag, Edit2, Save, X, FileText, Download, Eye, Shield, AlertTriangle, CheckCircle, Brain, Lightbulb, Zap, Target, Sparkles } from "lucide-react";
import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd";
import { motion, AnimatePresence } from "framer-motion";
import axios from "axios";

const BASE_URL = "https://localhost/api/v1";
const LOCAL_AI_URL = "http://localhost:5000/api/v1"; // Our Flask app


// Enhanced AI Service with all functions
const aiService = {
  async getEventSuggestions(caseId: any, inputText: any, suggestionType = 'completion') {
    const token = sessionStorage.getItem('authToken');
    const response = await fetch(`${LOCAL_AI_URL}/ai/suggestions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({
        case_id: caseId,
        input_text: inputText,
        suggestion_type: suggestionType,
      }),
    });
    return response.json();
  },

  async getSeverityRecommendation(description: any) {
    const token = sessionStorage.getItem('authToken');
    const response = await fetch(`${LOCAL_AI_URL}/ai/severity`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ description }),
    });
    return response.json();
  },

  async getTagSuggestions(description: any) {
    const token = sessionStorage.getItem('authToken');
    const response = await fetch(`${LOCAL_AI_URL}/ai/tags`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({ description }),
    });
    return response.json();
  },

  async analyzeEvent(caseId: any, eventText: any) {
    const token = sessionStorage.getItem('authToken');
    const response = await fetch(`${LOCAL_AI_URL}/ai/analyze-event`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({
        case_id: caseId,
        event_text: eventText,
      }),
    });
    return response.json();
  },

  async getNextSteps(caseId: any) {
    const token = sessionStorage.getItem('authToken');
    const response = await fetch(`${LOCAL_AI_URL}/ai/cases/${caseId}/next-steps`, {
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
    });
    return response.json();
  },
  // Add these new methods to your aiService object
async getWordSuggestions(partialText: string, maxSuggestions = 3) {
  const token = sessionStorage.getItem('authToken');
  const response = await fetch(`${LOCAL_AI_URL}/ai/complete-word`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({
      text: partialText,
      max_suggestions: maxSuggestions
    })
  });
  return response.json();
},

async getSentenceCompletions(partialText: string, maxCompletions = 2) {
  const token = sessionStorage.getItem('authToken');
  const response = await fetch(`${LOCAL_AI_URL}/ai/complete-sentence`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({
      text: partialText,
      max_completions: maxCompletions
    })
  });
  return response.json();
}
};

// Utility function for debouncing
type DebouncedFunction<T extends (...args: any[]) => any> = (...args: Parameters<T>) => void;

// eslint-disable-next-line no-unused-vars
interface Debounce {
    <T extends (...args: any[]) => any>(func: T, wait: number): DebouncedFunction<T>;
}

/* eslint-disable no-unused-vars */

const debounce: Debounce = <T extends (...args: any[]) => any>(func: T, wait: number): DebouncedFunction<T> => {
    let timeout: ReturnType<typeof setTimeout>;
    return function executedFunction(...args: Parameters<T>) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
};

/* eslint-enable no-unused-vars */



function getUserNameFromToken(): string | null {
  const token = sessionStorage.getItem('authToken');
  if (!token) return null;
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')));
    return decoded.fullName || decoded.username || null;
  } catch {
    return null;
  }
}

async function addTimelineEvent(caseId: string, eventData: {
  description: string;
  evidence: string[];
  tags: string[];
  severity: string;
  analystName?: string;
  createatedAt?: string;
}) {
  const token = sessionStorage.getItem('authToken');
  if (!token) throw new Error("No auth token found");

  const res = await fetch(`${BASE_URL}/cases/${caseId}/timeline`, {
    method: "POST",
    headers: { 
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`,
    },
    body: JSON.stringify(eventData),
  });

  if (!res.ok) throw new Error("Failed to add timeline event");
  return await res.json();
}

async function deleteTimelineEvent(eventId: string) {
  const token = sessionStorage.getItem('authToken');
  const res = await fetch(`${BASE_URL}/timeline/${eventId}`, {
    method: "DELETE",
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    }
  });
  if (!res.ok) throw new Error("Failed to delete timeline event");
}

async function updateTimelineEvent(eventId: string, updateData: {
  description?: string;
  evidence?: string[];
  tags?: string[];
  severity?: string;
}) {
  const token = sessionStorage.getItem('authToken');
  const res = await fetch(`${BASE_URL}/timeline/${eventId}`, {
    method: "PATCH",
    headers: { 
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    body: JSON.stringify(updateData),
  });
  if (!res.ok) throw new Error("Failed to update timeline event");
  return await res.json();
}

async function reorderTimelineEvents(caseId: string, orderedIds: string[]) {
  const token = sessionStorage.getItem('authToken');
  const res = await fetch(`${BASE_URL}/cases/${caseId}/timeline/reorder`, {
    method: "POST",
    headers: { 
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    body: JSON.stringify({ ordered_ids: orderedIds }),
  });
  if (!res.ok) throw new Error("Failed to reorder timeline events");
}

export function InvestigationTimeline({
  caseId,
  timelineEvents,
  setTimelineEvents,
  updateCaseTimestamp
}: {
  caseId: string;
  evidenceItems: any[];
  timelineEvents: any[];
  setTimelineEvents: (events: any[]) => void;
  updateCaseTimestamp: (id: string) => void;
}) {
  const [showAddForm, setShowAddForm] = useState(false);
  const [newEventDescription, setNewEventDescription] = useState("");
  const [newEventEvidence, setNewEventEvidence] = useState<string>("");
  const [newEventTags, setNewEventTags] = useState<string[]>([]);
  const [newEventSeverity, setNewEventSeverity] = useState<'low' | 'medium' | 'high' | 'critical'>('medium');
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editDescription, setEditDescription] = useState("");
  const [expandedEvidence, setExpandedEvidence] = useState<{[key: string]: boolean}>({});
  const [evidenceItems, setEvidenceItems] = useState<any[]>([]);
  const [, setEvidenceLoading] = useState(false);
  const [, setEvidenceError] = useState<string | null>(null);

  // AI-related states
  const [, setAiSuggestions] = useState([]);
  const [, setShowSuggestions] = useState(false);
  const [, setIsLoadingSuggestions] = useState(false);
  const [aiRecommendedSeverity, setAiRecommendedSeverity] = useState("");
  const [aiRecommendedTags, setAiRecommendedTags] = useState([]);
  const [showAIPanel, setShowAIPanel] = useState(false);
  const [nextStepSuggestions, setNextStepSuggestions] = useState([]);
  const [wordSuggestions, ] = useState<string[]>([]);
  const [sentenceCompletions, setSentenceCompletions] = useState<string[]>([]);
  const [showWordSuggestions, setShowWordSuggestions] = useState(false);
  const [showSentenceCompletions, setShowSentenceCompletions] = useState(false);

  // Debounced AI suggestions
  const debouncedGetSuggestions = useCallback(
    debounce(async (text: string, caseId: string) => {
      if (text.length > 10) {
        setIsLoadingSuggestions(true);
        try {
          const result = await aiService.getEventSuggestions(caseId, text);
          if (result.success) {
            setAiSuggestions(result.suggestions || []);
            setShowSuggestions(true);
          }
        } catch (error) {
          console.error('Error getting AI suggestions:', error);
        } finally {
          setIsLoadingSuggestions(false);
        }
      } else {
        setAiSuggestions([]);
        setShowSuggestions(false);
      }
    }, 800),
    []
  );



  // Get AI recommendations when description changes
  useEffect(() => {
    if (newEventDescription.trim()) {
      debouncedGetSuggestions(newEventDescription, caseId);
      
      // Get severity and tag recommendations
      if (newEventDescription.length > 20) {
        getSeverityRecommendation(newEventDescription);
        getTagRecommendations(newEventDescription);
      }
    } else {
      setAiSuggestions([]);
      setShowSuggestions(false);
    }
  }, [newEventDescription, caseId, debouncedGetSuggestions]);

  // Load next steps when timeline changes
  useEffect(() => {
    if (timelineEvents.length > 0) {
      loadNextSteps();
    }
  }, [timelineEvents]);

  const getSeverityRecommendation = async (description: string) => {
    try {
      const result = await aiService.getSeverityRecommendation(description);
      if (result.success && result.confidence > 0.7) {
        setAiRecommendedSeverity(result.recommended_severity);
      }
    } catch (error) {
      console.error('Error getting severity recommendation:', error);
    }
  };

  const getTagRecommendations = async (description: string) => {
    try {
      const result = await aiService.getTagSuggestions(description);
      if (result.success) {
        setAiRecommendedTags(result.tags || []);
      }
    } catch (error) {
      console.error('Error getting tag recommendations:', error);
    }
  };

  const loadNextSteps = async () => {
    try {
      const result = await aiService.getNextSteps(caseId);
      if (result.success) {
        setNextStepSuggestions(result.suggestions || []);
      }
    } catch (error) {
      console.error('Error loading next steps:', error);
    }
  };


  const applyAISeverity = () => {
    if (aiRecommendedSeverity) {
      setNewEventSeverity(aiRecommendedSeverity as any);
      setAiRecommendedSeverity("");
    }
  };

  const applyAITags = () => {
    const combinedTags = [...new Set([...newEventTags, ...aiRecommendedTags])];
    setNewEventTags(combinedTags);
    setAiRecommendedTags([]);
  };

  const getSentenceCompletions = async (text: string) => {
    if (text.length > 10) {
      try {
        const result = await aiService.getSentenceCompletions(text);
        if (result.success) {
          setSentenceCompletions(result.completions || []);
          setShowSentenceCompletions(true);
        }
      } catch (error) {
        console.error('Error getting sentence completions:', error);
      }
    }
  };

  const getCurrentTimestamp = () => {
    const now = new Date();
    const date = now.toISOString().split("T")[0];
    const time = now.toTimeString().slice(0, 5);
    return { date, time };
  };

  const addEvent = async () => {
    if (!newEventDescription.trim()) return;
    try {
      const createdEvent = await addTimelineEvent(caseId, {
        description: newEventDescription.trim(),
        evidence: newEventEvidence ? [newEventEvidence] : [],
        tags: newEventTags,
        severity: newEventSeverity,
        analystName: getUserNameFromToken() || undefined,
      });
      setTimelineEvents([...timelineEvents, createdEvent]);
      
      // Reset form and AI states
      setNewEventDescription("");
      setNewEventEvidence("");
      setNewEventTags([]);
      setNewEventSeverity('medium');
      setShowAddForm(false);
      setAiSuggestions([]);
      setShowSuggestions(false);
      setAiRecommendedSeverity("");
      setAiRecommendedTags([]);
      
      updateCaseTimestamp(caseId);
    } catch (err) {
      console.error(err);
    }
  };

  const fetchEvidence = async () => {
    setEvidenceLoading(true);
    setEvidenceError(null);
    try {
      const token = sessionStorage.getItem('authToken');
      if (!token) {
        throw new Error('No authentication token found');
      }
      
      const res = await axios.get(
        `${BASE_URL}/cases/${caseId}/iocs`,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );
      
      console.log('Evidence API response:', res.data);
      
      const data = res.data;
      let files: any[] = [];
      
      if (Array.isArray(data)) {
        files = data;
      } else if (typeof data === 'object' && data !== null) {
        if ('files' in data && Array.isArray((data as any).files)) {
          files = (data as any).files;
        } else if ('evidence' in data && Array.isArray((data as any).evidence)) {
          files = (data as any).evidence;
        } else if ('data' in data && Array.isArray((data as any).data)) {
          files = (data as any).data;
        } else if ('iocs' in data && Array.isArray((data as any).iocs)) {
          files = (data as any).iocs;
        }
      }
      
      console.log('Processed files:', files);
      setEvidenceItems(files);
    } catch (err: any) {
      console.error("Failed to fetch evidence files:", err);
      console.error("Error response:", err.response?.data);
      setEvidenceError(err.response?.data?.message || err.message || 'Failed to fetch evidence');
      setEvidenceItems([]);
    } finally {
      setEvidenceLoading(false);
    }
  };

  useEffect(() => {
    fetchEvidence();
    const interval = setInterval(fetchEvidence, 10 * 60 * 1000);
    return () => clearInterval(interval);
  }, [caseId]);

  const deleteEvent = async (index: number) => {
    const eventId = timelineEvents[index].id;
    try {
      await deleteTimelineEvent(eventId);
      setTimelineEvents(timelineEvents.filter((_, i) => i !== index));
    } catch (err) {
      console.error(err);
    }
  };

  const startEditing = (index: number, currentDesc: string) => {
    setEditingIndex(index);
    setEditDescription(currentDesc);
  };

  const saveEdit = async () => {
    if (editingIndex === null) return;
    const event = timelineEvents[editingIndex];
    try {
      const updatedEvent = await updateTimelineEvent(event.id, { description: editDescription });
      const updatedList = [...timelineEvents];
      updatedList[editingIndex] = updatedEvent;
      setTimelineEvents(updatedList);
      setEditingIndex(null);
    } catch (err) {
      console.error(err);
    }
  };

  const onDragEnd = async (result: any) => {
    if (!result.destination) return;
    const reordered = Array.from(timelineEvents);
    const [moved] = reordered.splice(result.source.index, 1);
    reordered.splice(result.destination.index, 0, moved);

    setTimelineEvents(reordered);

    try {
      const orderedIds = reordered.map(event => event.id);
      await reorderTimelineEvents(caseId, orderedIds);
    } catch (err) {
      console.error(err);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'low': return 'border-green-500 bg-green-500/10';
      case 'medium': return 'border-yellow-500 bg-yellow-500/10';
      case 'high': return 'border-orange-500 bg-orange-500/10';
      case 'critical': return 'border-red-500 bg-red-500/10';
      default: return 'border-blue-500 bg-blue-500/10';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'low': return <CheckCircle size={12} className="text-green-400" />;
      case 'medium': return <Shield size={12} className="text-yellow-400" />;
      case 'high': return <AlertTriangle size={12} className="text-orange-400" />;
      case 'critical': return <AlertTriangle size={12} className="text-red-400" />;
      default: return <Shield size={12} className="text-blue-400" />;
    }
  };

  const handleEvidenceAction = (evidenceName: string, action: 'view' | 'download') => {
    const evidence = evidenceItems.find(item => item.name === evidenceName);
    if (evidence) {
      if (action === 'view') {
        console.log('Viewing evidence:', evidence);
      } else if (action === 'download') {
        console.log('Downloading evidence:', evidence);
      }
    }
  };

  const toggleEvidenceExpansion = (eventId: string) => {
    setExpandedEvidence(prev => ({
      ...prev,
      [eventId]: !prev[eventId]
    }));
  };

  useEffect(() => {
    async function fetchTimeline() {
      try {
        const token = sessionStorage.getItem('authToken');
        const res = await fetch(`${BASE_URL}/cases/${caseId}/timeline`, {
          headers: {
            Authorization: `Bearer ${token}`,
          }
        });
        if (!res.ok) throw new Error("Failed to fetch timeline");
        const events = await res.json();
        setTimelineEvents(Array.isArray(events) ? events : []);
      } catch (error) {
        console.error(error);
      }
    }
    fetchTimeline();
  }, [caseId]);

  return (
    <div className="bg-card border border rounded-lg p-8 text-foreground min-h-screen max-w-7xl mx-auto shadow-lg">
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <h2 className="text-2xl font-semibold text-foreground flex items-center gap-2">
            <Calendar className="text-blue-400" size={24} />
            Investigation Timeline
          </h2>
        </div>
        <div className="flex items-center gap-6">
          <button
            onClick={() => setShowAIPanel(!showAIPanel)}
            className="flex items-center gap-2 px-3 py-1 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-all duration-200"
          >
            <Brain size={16} />
            AI Assistant
          </button>
          <button
            onClick={() => setShowAddForm(!showAddForm)}
            className="flex items-center gap-2 px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary transition-all duration-200 shadow-lg hover:shadow-blue-500/25"
          >
            <Plus size={18} />
            Add Event
          </button>
        </div>
      </div>

      {/* AI Assistant Panel */}
      <AnimatePresence>
        {showAIPanel && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="mb-6 p-4 bg-purple-900/20 border border-purple-500/30 rounded-lg"
          >
            <div className="flex items-center gap-2 mb-4">
              <Brain className="text-purple-400" size={20} />
              <h3 className="text-lg font-semibold text-purple-400">AI Investigation Assistant</h3>
            </div>
            
            {nextStepSuggestions.length > 0 && (
              <div className="mb-4">
                <div className="flex items-center gap-2 mb-2">
                  <Target className="text-green-400" size={16} />
                  <span className="text-sm font-medium text-green-400">Suggested Next Steps:</span>
                </div>
                <div className="grid gap-2">
                  {nextStepSuggestions.map((step, index) => (
                    <div
                      key={index}
                      className="p-2 bg-background border border-border rounded cursor-pointer hover:bg-primary/10 transition-colors"
                      onClick={() => {
                        setNewEventDescription(step);
                        setShowAddForm(true);
                        setShowAIPanel(false);
                      }}
                    >
                      <div className="flex items-center gap-2">
                        <Lightbulb size={14} className="text-yellow-400" />
                        <span className="text-sm text-foreground">{step}</span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>

      {showAddForm && (
        <motion.div 
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8 p-6 bg-background border border-border rounded-lg shadow-xl relative"
        >
          <div className="flex items-center gap-2 mb-4 text-blue-400">
            <Calendar size={16} />
            <span className="text-sm">
              Will be timestamped: {getCurrentTimestamp().date} at {getCurrentTimestamp().time}
            </span>
          </div>
          <div className="relative">
            <textarea
              value={newEventDescription}
              onChange={(e) => setNewEventDescription(e.target.value)}
              onKeyDown={(e) => {
                // Tab key for sentence completions
                if (e.key === 'Tab' && newEventDescription.length > 10) {
                  e.preventDefault();
                  getSentenceCompletions(newEventDescription);
                }
                // Escape to hide suggestions
                if (e.key === 'Escape') {
                  setShowWordSuggestions(false);
                  setShowSentenceCompletions(false);
                }
              }}
              placeholder="Describe the investigation event... (Type 3+ characters for word suggestions, Tab for sentence completions)"
              className="w-full px-4 py-3 bg-background border border-border text-foreground placeholder-gray-400 rounded-lg mb-4 focus:border-primary focus:outline-none resize-none"
              rows={3}
            />
            
            {/* Word Suggestions Dropdown */}
            <AnimatePresence>
              {showWordSuggestions && wordSuggestions.length > 0 && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="absolute z-50 w-full mt-1 bg-background border border-border rounded-lg shadow-lg max-h-40 overflow-y-auto"
                  style={{ top: '100%', marginTop: '-1rem' }}
                >
                  {wordSuggestions.map((suggestion: string, index: number) => (
                    <div
                      key={index}
                      className="p-2 hover:bg-primary/10 cursor-pointer border-b border-border last:border-b-0 flex items-center gap-2"
                      onClick={() => {
                        setNewEventDescription(prev => prev + ' ' + suggestion + ' ');
                        setShowWordSuggestions(false);
                      }}
                    >
                      <Zap size={12} className="text-yellow-400" />
                      <span className="text-sm text-foreground">{suggestion}</span>
                    </div>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
            {/* Add this below your textarea */}
          <div className="text-xs text-foreground/60 mb-4 flex items-center gap-2">
            <Lightbulb size={12} />
            <span>Tip: Type 3+ characters for word suggestions, press Tab for sentence completions</span>
          </div>

            {/* Sentence Completions Dropdown */}
            <AnimatePresence>
              {showSentenceCompletions && sentenceCompletions.length > 0 && (
                <motion.div
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="absolute z-50 w-full mt-1 bg-background border border-purple-500/30 rounded-lg shadow-lg"
                  style={{ top: '100%', marginTop: '-1rem' }}
                >
                  <div className="p-2 bg-purple-900/20 border-b border-purple-500/30">
                    <span className="text-xs text-purple-400 font-medium">AI Sentence Completions (Tab to trigger)</span>
                  </div>
                  {sentenceCompletions.map((completion: string, index: number) => (
                    <div
                      key={index}
                      className="p-3 hover:bg-primary/10 cursor-pointer border-b border-border last:border-b-0"
                      onClick={() => {
                        setNewEventDescription(prev => prev + ' ' + completion);
                        setShowSentenceCompletions(false);
                      }}
                    >
                      <div className="flex items-start gap-2">
                        <Sparkles size={14} className="text-purple-400 mt-0.5 flex-shrink-0" />
                        <span className="text-sm text-foreground">{completion}</span>
                      </div>
                    </div>
                  ))}
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* AI Recommendations */}
          {(aiRecommendedSeverity || aiRecommendedTags.length > 0) && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              className="mb-4 p-3 bg-purple-900/20 border border-purple-500/30 rounded-lg"
            >
              <div className="flex items-center gap-2 mb-2">
                <Zap className="text-purple-400" size={16} />
                <span className="text-sm font-medium text-purple-400">AI Recommendations:</span>
              </div>
              
              {aiRecommendedSeverity && (
                <div className="mb-2 flex items-center gap-2">
                  <span className="text-xs text-foreground">Severity:</span>
                  <button
                    onClick={applyAISeverity}
                    className={`px-2 py-1 rounded text-xs ${getSeverityColor(aiRecommendedSeverity)} border hover:bg-opacity-50 transition-colors flex items-center gap-1`}
                  >
                    {getSeverityIcon(aiRecommendedSeverity)}
                    <span className="ml-1 capitalize">{aiRecommendedSeverity}</span>
                  </button>
                </div>
              )}
              
              {aiRecommendedTags.length > 0 && (
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-xs text-foreground">Tags:</span>
                  <button
                    onClick={applyAITags}
                    className="px-2 py-1 bg-purple-600 text-white rounded text-xs hover:bg-purple-700 transition-colors"
                  >
                    Apply {aiRecommendedTags.length} tags
                  </button>
                  <span className="text-xs text-foreground/60">
                    ({aiRecommendedTags.join(", ")})
                  </span>
                </div>
              )}
            </motion.div>
          )}
          
          {/* Severity Selection */}
          <div className="mb-4">
            <label className="text-sm font-medium text-foreground mb-2 block">Severity Level:</label>
            <div className="flex gap-2">
              {['low', 'medium', 'high', 'critical'].map((level) => (
                <button
                  key={level}
                  onClick={() => setNewEventSeverity(level as any)}
                  className={`px-3 py-2 rounded-lg text-xs font-medium transition-all flex items-center gap-1 ${
                    newEventSeverity === level
                      ? getSeverityColor(level).replace('bg-', 'bg-').replace('/10', '/30') + ' border-2'
                      : 'bg-background border border-border text-foreground hover:bg-primary/10'
                  }`}
                >
                  {getSeverityIcon(level)}
                  <span className="ml-1 capitalize">{level}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Evidence linking */}
          <div className="mb-4">
            <label className="text-sm font-medium text-foreground mb-2 block">Link Evidence Files:</label>
            <div className="max-h-40 overflow-y-auto bg-background rounded-lg p-3">
              {evidenceItems.length === 0 ? (
                <p className="text-foreground text-sm">No evidence files available</p>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  {evidenceItems.map((item) => (
                    <button
                      key={item.id}
                      onClick={() => setNewEventEvidence(item.value)}
                      className={`p-2 border rounded-lg text-xs transition-all text-left ${
                        newEventEvidence === item.value
                          ? "bg-primary border-border text-foreground shadow-lg"
                          : "bg-primary border-border text-foreground hover:bg-primary/10"
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        <FileText size={14} />
                        <span className="truncate">
                          {item.type ? `${item.type}: ${item.value}` : item.value}
                        </span>
                      </div>
                      {item.created_at && (
                        <div className="text-xs opacity-75 mt-1">
                          {new Date(item.created_at).toLocaleString()}
                        </div>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Tags */}
          <div className="mb-6">
            <label className="text-sm font-medium text-foreground mb-2 block">Tags:</label>
            <input
              type="text"
              placeholder="e.g., Analysis, Containment, Malware Detection, Network Forensics"
              value={newEventTags.join(", ")}
              onChange={(e) => setNewEventTags(e.target.value.split(",").map((t) => t.trim()).filter(t => t))}
              className="w-full px-4 py-2 bg-background border border-border text-foreground placeholder-gray-400 rounded-lg focus:border-primary focus:outline-none"
            />
          </div>

          <div className="flex gap-3">
            <button
              onClick={addEvent}
              disabled={!newEventDescription.trim()}
              className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-primary/10 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              <Plus size={16} />
              Add Event
            </button>
            <button
              onClick={() => {
                setShowAddForm(false);
                setAiSuggestions([]);
                setShowSuggestions(false);
                setAiRecommendedSeverity("");
                setAiRecommendedTags([]);
              }}
              className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary/60 transition-colors"
            >
              Cancel
            </button>
          </div>
        </motion.div>
      )}

      <DragDropContext onDragEnd={onDragEnd}>
        <Droppable droppableId="timeline">
          {(provided: any) => (
            <div {...provided.droppableProps} ref={provided.innerRef} className="relative">
              {/* Timeline line */}
              <div className="absolute left-36 top-0 bottom-0 w-px bg-gradient-to-b from-blue-500 via-cyan-400 to-blue-500 opacity-30"></div>
              
              <AnimatePresence>
                {timelineEvents.map((event, index) => (
                <Draggable key={event.id || index} draggableId={`event-${event.id || index}`} index={index}>
                  {(provided: any, snapshot: any) => (
                    <motion.div
                      ref={provided.innerRef}
                      {...provided.draggableProps}
                      {...provided.dragHandleProps}
                      className={`flex items-start mb-8 relative transition-all duration-200 ${
                        snapshot.isDragging 
                          ? 'z-50 scale-105 shadow-2xl border-2 border-blue-500 bg-blue-500/10 rounded-lg transform rotate-1' 
                          : 'hover:scale-[1.02]'
                      }`}
                    >
                        {/* Date/time */}
                        <div className="w-32 text-right pr-4">
                          <div className="text-gray-400 text-sm flex items-center justify-end gap-1 mb-1">
                            <Calendar size={12} />
                            {event.date}
                          </div>
                          <div className="text-gray-400 text-sm flex items-center justify-end gap-1">
                            <Clock size={12} />
                            {event.time}
                          </div>
                        </div>

                        {/* Marker */}
                        <div className={`w-10 h-10 rounded-full border-4 border-gray-900 flex items-center justify-center relative z-10 ${getSeverityColor(event.severity || 'medium')}`}>
                          {getSeverityIcon(event.severity || 'medium')}
                        </div>

                        {/* Content */}
                        <div className="flex-1 ml-6">
                          <div className={`border rounded-lg p-4 ${getSeverityColor(event.severity || 'medium')} bg-background/50 backdrop-blur-sm`}>
                            {editingIndex === index ? (
                              <div>
                                <textarea
                                  value={editDescription}
                                  onChange={(e) => setEditDescription(e.target.value)}
                                  className="w-full mb-3 px-3 py-2 bg-background border border-primary text-foreground rounded-lg resize-none"
                                  rows={3}
                                />
                                <div className="flex gap-2">
                                  <button 
                                    onClick={saveEdit} 
                                    className="px-3 py-1 bg-green-600 text-white rounded-md hover:bg-green-700 flex items-center gap-1"
                                  >
                                    <Save size={14} /> Save
                                  </button>
                                  <button 
                                    onClick={() => setEditingIndex(null)} 
                                    className="px-3 py-1 bg-gray-600 text-white rounded-md hover:bg-gray-500 flex items-center gap-1"
                                  >
                                    <X size={14} /> Cancel
                                  </button>
                                </div>
                              </div>
                            ) : (
                              <div>
                                <div className="flex justify-between items-start mb-3">
                                  <p className="text-foreground leading-relaxed">{event.description}</p>
                                  <div className="flex gap-2 ml-4">
                                    <button 
                                      onClick={() => startEditing(index, event.description)} 
                                      className="text-blue-400 hover:text-blue-300 p-1"
                                      title="Edit event"
                                    >
                                      <Edit2 size={14} />
                                    </button>
                                    <button 
                                      onClick={() => deleteEvent(index)} 
                                      className="text-red-400 hover:text-red-300 p-1"
                                      title="Delete event"
                                    >
                                      <X size={14} />
                                    </button>
                                  </div>
                                </div>

                                {/* Tags */}
                                {event.tags?.length > 0 && (
                                  <div className="mb-3 flex flex-wrap gap-2">
                                    {event.tags.map((tag: string, i: number) => (
                                      <span key={i} className="px-2 py-1 bg-primary/60 border border-primary/60 text-foreground/80 text-xs rounded-full flex items-center gap-1">
                                        <Tag size={10} /> {tag}
                                      </span>
                                    ))}
                                  </div>
                                )}

                                {/* Linked Evidence */}
                                {event.evidence?.length > 0 && (
                                  <div className="mt-3 border-t border-border pt-3">
                                    <div className="flex items-center justify-between mb-2">
                                      <span className="text-sm font-medium text-foreground/80 flex items-center gap-1">
                                        <Paperclip size={14} />
                                        Linked Evidence ({event.evidence.length})
                                      </span>
                                      <button
                                        onClick={() => toggleEvidenceExpansion(event.id || `${index}`)}
                                        className="text-primary hover:text-primary/60 text-xs"
                                      >
                                        {expandedEvidence[event.id || `${index}`] ? 'Collapse' : 'Expand'}
                                      </button>
                                    </div>
                                    <div className={`grid gap-2 transition-all duration-200 ${
                                      expandedEvidence[event.id || `${index}`] 
                                        ? 'grid-cols-1 sm:grid-cols-2' 
                                        : 'grid-cols-1'
                                    }`}>
                                      {event.evidence.map((evidenceName: string, i: number) => {
                                        const evidenceItem = evidenceItems.find(item => item.name === evidenceName);
                                        return (
                                          <div key={i} className="bg-background border border-border rounded-lg p-2">
                                            <div className="flex items-center justify-between">  
                                              <div className="flex items-center gap-2 min-w-0">
                                                <FileText size={14} className="text-foreground/80 flex-shrink-0" />
                                                <span className="text-sm text-foreground/80 truncate">{evidenceName}</span>
                                              </div>
                                              <div className="flex gap-1 ml-2">
                                                <button
                                                  onClick={() => handleEvidenceAction(evidenceName, 'view')}
                                                  className="text-blue-400 hover:text-blue-300 p-1"
                                                  title="View evidence"
                                                >
                                                  <Eye size={12} />
                                                </button>
                                                <button
                                                  onClick={() => handleEvidenceAction(evidenceName, 'download')}
                                                  className="text-green-400 hover:text-green-300 p-1"
                                                  title="Download evidence"
                                                >
                                                  <Download size={12} />
                                                </button>
                                              </div>
                                            </div>
                                            {expandedEvidence[event.id || `${index}`] && evidenceItem && (
                                              <div className="mt-2 pt-2 border-t border-gray-600 text-xs text-gray-400">
                                                {evidenceItem.type && <div>Type: {evidenceItem.type}</div>}
                                                {evidenceItem.created_at && (
                                                  <div>Added: {new Date(evidenceItem.created_at).toLocaleString()}</div>
                                                )}
                                              </div>
                                            )}
                                          </div>
                                        );
                                      })}
                                    </div>
                                  </div>
                                )}

                                {/* Analyst info */}
                                {event.analystName && (
                                  <div className="mt-3 pt-3 border-t border-gray-600 text-xs text-gray-400">
                                    Added by: {event.analystName}
                                  </div>
                                )}
                              </div>
                            )}
                          </div>
                        </div>
                      </motion.div>
                    )}
                  </Draggable>
                ))}
              </AnimatePresence>
              {provided.placeholder}
            </div>
          )}
        </Droppable>
      </DragDropContext>

      {timelineEvents.length === 0 && (
        <div className="text-center py-12 text-foreground/60">
          <Calendar size={48} className="mx-auto mb-4 opacity-50" />
          <h3 className="text-lg font-medium mb-2">No timeline events yet</h3>
          <p>Start building your investigation timeline by adding the first event.</p>
        </div>
      )}
    </div>
  );
}