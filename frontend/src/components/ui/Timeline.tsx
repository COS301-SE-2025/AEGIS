import { useState } from "react";
import { Plus, Calendar, Clock, Paperclip, Tag, Edit2, Save, X, FileText, Download, Eye, Shield, AlertTriangle, CheckCircle } from "lucide-react";
import { DragDropContext, Droppable, Draggable } from "@hello-pangea/dnd";
import { motion, AnimatePresence } from "framer-motion";
import { useEffect } from "react";

const BASE_URL = "http://localhost:8080/api/v1";

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
  analystName?: string; // Optional, backend extracts from token
  createatedAt?: string; // Optional, backend sets default
}) {
  const token = sessionStorage.getItem('authToken');
  if (!token) throw new Error("No auth token found");

  // Only send eventData, no analystName — backend extracts analyst from token
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
  evidenceItems,
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
  const [newEventEvidence, setNewEventEvidence] = useState<string[]>([]);
  const [newEventTags, setNewEventTags] = useState<string[]>([]);
  const [newEventSeverity, setNewEventSeverity] = useState<'low' | 'medium' | 'high' | 'critical'>('medium');
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editDescription, setEditDescription] = useState("");
  const [expandedEvidence, setExpandedEvidence] = useState<{[key: string]: boolean}>({});

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
      evidence: newEventEvidence,
      tags: newEventTags,
      severity: newEventSeverity,
      analystName: getUserNameFromToken() || undefined, // Optional, backend extracts from token
    });
    setTimelineEvents([...timelineEvents, createdEvent]); // update local state with backend response
    setNewEventDescription("");
    setNewEventEvidence([]);
    setNewEventTags([]);
    setNewEventSeverity('medium');
    setShowAddForm(false);
    updateCaseTimestamp(caseId);
  } catch (err) {
    console.error(err);
    alert("Failed to add event.");
  }
};


const deleteEvent = async (index: number) => {
  const eventId = timelineEvents[index].id;
  try {
    await deleteTimelineEvent(eventId);
    setTimelineEvents(timelineEvents.filter((_, i) => i !== index));
  } catch (err) {
    console.error(err);
    alert("Failed to delete event.");
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
    alert("Failed to update event.");
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
    alert("Failed to reorder events.");
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
        // In real app, open evidence viewer modal
        console.log('Viewing evidence:', evidence);
      } else if (action === 'download') {
        // In real app, trigger download
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
    <div className="bg-card border border-bg-accent rounded-lg p-8 text-gray-100 min-h screen max-w-6xl mx-auto shadow-lg">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-semibold text-white flex items-center gap-2">
          <Calendar className="text-blue-400" size={24} />
          Investigation Timeline
        </h2>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-all duration-200 shadow-lg hover:shadow-blue-500/25"
        >
          <Plus size={18} />
          Add Event
        </button>
      </div>

      {showAddForm && (
        <motion.div 
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8 p-6 bg-gray-800 border border-gray-600 rounded-lg shadow-xl"
        >
          <div className="flex items-center gap-2 mb-4 text-blue-400">
            <Calendar size={16} />
            <span className="text-sm">
              Will be timestamped: {getCurrentTimestamp().date} at {getCurrentTimestamp().time}
            </span>
          </div>
          
          <textarea
            value={newEventDescription}
            onChange={(e) => setNewEventDescription(e.target.value)}
            placeholder="Describe the investigation event..."
            className="w-full px-4 py-3 bg-gray-700 border border-gray-600 text-gray-100 placeholder-gray-400 rounded-lg mb-4 focus:border-blue-500 focus:outline-none resize-none"
            rows={3}
          />
          
          {/* Severity Selection */}
          <div className="mb-4">
            <label className="text-sm font-medium text-gray-300 mb-2 block">Severity Level:</label>
            <div className="flex gap-2">
              {['low', 'medium', 'high', 'critical'].map((level) => (
                <button
                  key={level}
                  onClick={() => setNewEventSeverity(level as any)}
                  className={`px-3 py-2 rounded-lg text-xs font-medium transition-all ${
                    newEventSeverity === level
                      ? getSeverityColor(level).replace('bg-', 'bg-').replace('/10', '/30') + ' border-2'
                      : 'bg-gray-700 border border-gray-600 text-gray-300 hover:bg-gray-600'
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
            <label className="text-sm font-medium text-gray-300 mb-2 block">Link Evidence Files:</label>
            <div className="max-h-40 overflow-y-auto bg-gray-700 rounded-lg p-3">
              {evidenceItems.length === 0 ? (
                <p className="text-gray-500 text-sm">No evidence files available</p>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  {evidenceItems.map((item) => (
                    <button
                      key={item.name}
                      onClick={() =>
                        setNewEventEvidence((prev) =>
                          prev.includes(item.name)
                            ? prev.filter((e) => e !== item.name)
                            : [...prev, item.name]
                        )
                      }
                      className={`p-2 border rounded-lg text-xs transition-all text-left ${
                        newEventEvidence.includes(item.name)
                          ? "bg-blue-600 border-blue-500 text-white shadow-lg"
                          : "bg-gray-800 border-gray-600 text-gray-300 hover:bg-gray-600"
                      }`}
                    >
                      <div className="flex items-center gap-2">
                        <FileText size={14} />
                        <span className="truncate">{item.name}</span>
                      </div>
                      {item.size && (
                        <div className="text-xs opacity-75 mt-1">{item.size}</div>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Tags */}
          <div className="mb-6">
            <label className="text-sm font-medium text-gray-300 mb-2 block">Tags:</label>
            <input
              type="text"
              placeholder="e.g., Analysis, Containment, Malware Detection, Network Forensics"
              value={newEventTags.join(", ")}
              onChange={(e) => setNewEventTags(e.target.value.split(",").map((t) => t.trim()).filter(t => t))}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 text-gray-100 placeholder-gray-400 rounded-lg focus:border-blue-500 focus:outline-none"
            />
          </div>

          <div className="flex gap-3">
            <button
              onClick={addEvent}
              disabled={!newEventDescription.trim()}
              className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              <Plus size={16} />
              Add Event
            </button>
            <button
              onClick={() => setShowAddForm(false)}
              className="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-500 transition-colors"
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
                {(timelineEvents ?? []).map((event, index) => (
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
                          <div className={`border rounded-lg p-4 ${getSeverityColor(event.severity || 'medium')} bg-gray-800/50 backdrop-blur-sm`}>
                            {editingIndex === index ? (
                              <div>
                                <textarea
                                  value={editDescription}
                                  onChange={(e) => setEditDescription(e.target.value)}
                                  className="w-full mb-3 px-3 py-2 bg-gray-700 border border-gray-600 text-gray-100 rounded-lg resize-none"
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
                                  <p className="text-gray-100 leading-relaxed">{event.description}</p>
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
                                      <span key={i} className="px-2 py-1 bg-blue-500/20 border border-blue-500/30 text-blue-300 text-xs rounded-full flex items-center gap-1">
                                        <Tag size={10} /> {tag}
                                      </span>
                                    ))}
                                  </div>
                                )}

                                {/* Linked Evidence */}
                                {event.evidence?.length > 0 && (
                                  <div className="mt-3 border-t border-gray-700 pt-3">
                                    <div className="flex items-center justify-between mb-2">
                                      <span className="text-sm font-medium text-gray-300 flex items-center gap-1">
                                        <Paperclip size={14} />
                                        Linked Evidence ({event.evidence.length})
                                      </span>
                                      <button
                                        onClick={() => toggleEvidenceExpansion(event.id || `${index}`)}
                                        className="text-blue-400 hover:text-blue-300 text-xs"
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
                                          <div key={i} className="bg-gray-700/50 border border-gray-600 rounded-lg p-2">
                                            <div className="flex items-center justify-between">
                                              <div className="flex items-center gap-2 min-w-0">
                                                <FileText size={14} className="text-gray-400 flex-shrink-0" />
                                                <span className="text-sm text-gray-200 truncate">{evidenceName}</span>
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
                                                {evidenceItem.size && <div>Size: {evidenceItem.size}</div>}
                                                {evidenceItem.type && <div>Type: {evidenceItem.type}</div>}
                                                {evidenceItem.hash && <div className="font-mono">Hash: {evidenceItem.hash.substring(0, 16)}...</div>}
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
                                  <div className="mt-3 text-xs text-gray-500 border-t border-gray-700 pt-2">
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
              {timelineEvents.length === 0 && (
                <div className="text-center py-12 text-gray-500">
                  <Calendar size={48} className="mx-auto mb-4 opacity-50" />
                  <p>No timeline events yet. Add your first investigation event to get started.</p>
                </div>
              )}
              {provided.placeholder}
            </div>
          )}
        </Droppable>
      </DragDropContext>
    </div>
  );
}