import { useState } from "react";
import { Plus, Calendar, Clock, Paperclip, Tag, Edit2, Save, X } from "lucide-react";
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import { motion, AnimatePresence } from "framer-motion";

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
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editDescription, setEditDescription] = useState("");

  const getCurrentTimestamp = () => {
    const now = new Date();
    const date = now.toISOString().split("T")[0];
    const time = now.toTimeString().slice(0, 5);
    return { date, time };
  };

  const addEvent = () => {
    if (newEventDescription.trim()) {
      const { date, time } = getCurrentTimestamp();
      const newEvent = {
        date,
        time,
        description: newEventDescription.trim(),
        evidence: newEventEvidence,
        tags: newEventTags
      };
      setTimelineEvents([...timelineEvents, newEvent]);
      setNewEventDescription("");
      setNewEventEvidence([]);
      setNewEventTags([]);
      setShowAddForm(false);
      updateCaseTimestamp(caseId);
    }
  };

  const deleteEvent = (index: number) => {
    setTimelineEvents(timelineEvents.filter((_, i) => i !== index));
  };

  const startEditing = (index: number, currentDesc: string) => {
    setEditingIndex(index);
    setEditDescription(currentDesc);
  };

  const saveEdit = () => {
    if (editingIndex !== null) {
      const updated = [...timelineEvents];
      updated[editingIndex].description = editDescription;
      setTimelineEvents(updated);
      setEditingIndex(null);
    }
  };

  const onDragEnd = (result: any) => {
    if (!result.destination) return;
    const reordered = Array.from(timelineEvents);
    const [moved] = reordered.splice(result.source.index, 1);
    reordered.splice(result.destination.index, 0, moved);
    setTimelineEvents(reordered);
  };

  return (
    <div className="bg-card border border-bg-accent rounded-lg p-6">
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-semibold text-foreground">Investigation Timeline</h2>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-foreground rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus size={18} />
          Add Event
        </button>
      </div>

      {showAddForm && (
        <div className="mb-8 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-center gap-2 mb-3">
            <Calendar size={16} className="text-blue-600" />
            <span className="text-sm text-blue-800">
              Will be timestamped: {getCurrentTimestamp().date} at {getCurrentTimestamp().time}
            </span>
          </div>
          <textarea
            value={newEventDescription}
            onChange={(e) => setNewEventDescription(e.target.value)}
            placeholder="Enter event description..."
            className="w-full px-3 py-2 border border-gray-300 text-gray-700 rounded-md mb-3"
          />
          {/* Evidence linking */}
          <div className="mb-3">
            <label className="text-sm font-medium text-foreground">Link Evidence:</label>
            <div className="flex flex-wrap gap-2 mt-1">
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
                  className={`px-2 py-1 border rounded text-xs ${
                    newEventEvidence.includes(item.name)
                      ? "bg-blue-600 text-white"
                      : "bg-muted text-foreground"
                  }`}
                >
                  <Paperclip size={12} className="inline mr-1" />
                  {item.name}
                </button>
              ))}
            </div>
          </div>
          {/* Tags */}
          <div className="mb-3">
            <label className="text-sm font-medium text-foreground">Tags:</label>
            <input
              type="text"
              placeholder="Comma-separated tags (e.g., Analysis, Containment)"
              value={newEventTags.join(", ")}
              onChange={(e) => setNewEventTags(e.target.value.split(",").map((t) => t.trim()))}
              className="w-full px-3 py-2 border border-gray-300 text-gray-700 rounded-md"
            />
          </div>
          <div className="flex gap-2">
            <button
              onClick={addEvent}
              disabled={!newEventDescription.trim()}
              className="px-4 py-2 bg-green-600 text-foreground rounded-md hover:bg-green-700 disabled:bg-gray-400"
            >
              Add
            </button>
            <button
              onClick={() => setShowAddForm(false)}
              className="px-4 py-2 bg-gray-500 text-foreground rounded-md hover:bg-gray-600"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <DragDropContext onDragEnd={onDragEnd}>
        <Droppable droppableId="timeline">
          {(provided: any) => (
            <div {...provided.droppableProps} ref={provided.innerRef} className="relative">
              <AnimatePresence>
                {timelineEvents.map((event, index) => (
                  <Draggable key={index} draggableId={`event-${index}`} index={index}>
                    {(provided: any) => (
                      <motion.div
                        key={index}
                        ref={provided.innerRef}
                        {...provided.draggableProps}
                        {...provided.dragHandleProps}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -20 }}
                        transition={{ duration: 0.2 }}
                        className="flex items-start mb-8 relative"
                      >
                        {/* Date/time */}
                        <div className="w-32 text-right pr-4">
                          <div className="text-muted-foreground text-sm flex items-center justify-end gap-1">
                            <Calendar size={12} />
                            {event.date}
                          </div>
                          <div className="text-muted-foreground text-sm flex items-center justify-end gap-1">
                            <Clock size={12} />
                            {event.time}
                          </div>
                        </div>

                        {/* Marker */}
                        <div className="w-8 h-8 bg-blue-600 rounded-full border-4 border-background flex items-center justify-center relative z-10">
                          <div className="w-2 h-2 bg-white rounded-full"></div>
                        </div>

                        {/* Description */}
                        <div className="flex-1 ml-4">
                          <div className="bg-muted border rounded-lg p-4">
                            {editingIndex === index ? (
                              <div>
                                <textarea
                                  value={editDescription}
                                  onChange={(e) => setEditDescription(e.target.value)}
                                  className="w-full mb-2 px-2 py-1 border rounded"
                                />
                                <button onClick={saveEdit} className="px-2 py-1 bg-green-600 text-white rounded mr-2">
                                  <Save size={14} className="inline mr-1" /> Save
                                </button>
                                <button onClick={() => setEditingIndex(null)} className="px-2 py-1 bg-gray-500 text-white rounded">
                                  <X size={14} className="inline mr-1" /> Cancel
                                </button>
                              </div>
                            ) : (
                              <div className="flex justify-between items-center">
                                <div>
                                  <p className="text-foreground">{event.description}</p>
                                  {/* Tags */}
                                  {event.tags?.length > 0 && (
                                    <div className="mt-1 flex flex-wrap gap-1">
                                      {event.tags.map((tag: string, i: number) => (
                                        <span key={i} className="px-2 py-0.5 bg-blue-200 text-xs rounded">
                                          <Tag size={10} className="inline mr-1" /> {tag}
                                        </span>
                                      ))}
                                    </div>
                                  )}
                                  {/* Linked Evidence */}
                                  {event.evidence?.length > 0 && (
                                    <div className="mt-1 flex flex-wrap gap-1">
                                      {event.evidence.map((ev: string, i: number) => (
                                        <span key={i} className="px-2 py-0.5 bg-gray-200 text-xs rounded">
                                          <Paperclip size={10} className="inline mr-1" /> {ev}
                                        </span>
                                      ))}
                                    </div>
                                  )}
                                </div>
                                <div className="flex gap-2">
                                  <button onClick={() => startEditing(index, event.description)} className="text-blue-600">
                                    <Edit2 size={14} />
                                  </button>
                                  <button onClick={() => deleteEvent(index)} className="text-red-600">
                                    Delete
                                  </button>
                                </div>
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
    </div>
  );
}
