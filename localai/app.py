from flask import Flask, request, jsonify
from transformers import pipeline
import sys
import json
from flask_cors import CORS

app = Flask(__name__)
CORS(app)

try:
    print("[LocalAI] Flan-t5 model loading...", flush=True)
    generator = pipeline("text2text-generation", model="google/flan-t5-base")
    print("[LocalAI] Model loaded successfully.", flush=True)
except Exception as e:
    print(f"[LocalAI] Error loading model: {e}", file=sys.stderr, flush=True)
    generator = None

# Add a text completion pipeline specifically for word suggestions
try:
    print("[LocalAI] Loading text completion model for suggestions...", flush=True)
    # Use a smaller model or same model for completions
    completer = pipeline('text-generation', model='gpt2', max_length=50)
    print("[LocalAI] Completion model loaded successfully.", flush=True)
except Exception as e:
    print(f"[LocalAI] Error loading completion model: {e}", file=sys.stderr, flush=True)
    completer = generator  # Fallback to main generator


@app.route('/generate', methods=['POST'])
def generate():
    if generator is None:
        return jsonify({"error": "Model not loaded"}), 500
    try:
        data = request.json
        prompt = data.get("prompt", "")
        max_length = int(data.get("max_length", 200))  # default 200

        # Stronger DFIR context
        dfir_context = (
            "You are a DFIR (Digital Forensics and Incident Response) analyst. "
            "Write a clear, concise, and professional section of a forensic report.\n\n"
        )
        full_prompt = dfir_context + prompt

        print(f"[LocalAI] Prompt received:\n{full_prompt}", flush=True)

        # Add decoding settings to reduce repetition and improve fluency
        result = generator(
            full_prompt,
            max_length=max_length,
            num_return_sequences=1,
            temperature=0.7,          # controls creativity
            top_p=0.9,                # nucleus sampling
            repetition_penalty=2.0,   # discourages loops
            no_repeat_ngram_size=3,   # prevents repeated 3-grams
            do_sample=True            # enable sampling for variety
        )

        return jsonify({"text": result[0]["generated_text"]})
    except Exception as e:
        print(f"[LocalAI] Error in /generate: {e}", file=sys.stderr, flush=True)
        return jsonify({"error": str(e)}), 500


# HEALTH CHECK ENDPOINT
@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({
        'status': 'online' if generator is not None else 'offline',
        'model': 'gpt2' if generator is not None else None
    })

# TIMELINE ENDPOINTS

# Helper functions
def parse_suggestions(text):
    # Extract suggestions from generated text
    lines = text.split('\n')
    suggestions = [line.strip() for line in lines if len(line.strip()) > 10]
    return suggestions[:5] if suggestions else ["No suggestions generated"]

def extract_severity(text):
    text_lower = text.lower()
    if 'critical' in text_lower: return 'critical'
    if 'high' in text_lower: return 'high' 
    if 'medium' in text_lower: return 'medium'
    if 'low' in text_lower: return 'low'
    return 'medium'

def extract_tags(text):
    # Extract tags from text
    tags = []
    common_tags = ['malware', 'network', 'forensics', 'analysis', 'incident', 'response', 
                  'phishing', 'ioc', 'investigation', 'security', 'breach', 'compromise']
    text_lower = text.lower()
    for tag in common_tags:
        if tag in text_lower:
            tags.append(tag)
    return tags[:3] if tags else ['investigation']

def parse_next_steps(text):
    lines = text.split('\n')
    steps = [line.strip() for line in lines if len(line.strip()) > 10]
    return steps[:5] if steps else ["Review the current evidence", "Document findings"]

def simple_parse_iocs(text):
    # Simple IOC parser fallback
    iocs = []
    # Basic pattern matching for common IOCs
    import re
    ip_pattern = r'\b(?:\d{1,3}\.){3}\d{1,3}\b'
    domain_pattern = r'\b[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.([a-zA-Z]{2,})\b'
    hash_pattern = r'\b[a-fA-F0-9]{32,64}\b'
    
    for ip in re.findall(ip_pattern, text):
        iocs.append({"type": "ip", "value": ip, "confidence": 0.7})
    
    for domain in re.findall(domain_pattern, text):
        iocs.append({"type": "domain", "value": domain, "confidence": 0.6})
    
    for hash_val in re.findall(hash_pattern, text):
        if len(hash_val) in [32, 40, 64]:  # MD5, SHA1, SHA256
            iocs.append({"type": "hash", "value": hash_val, "confidence": 0.8})
    
    return iocs[:10]  # Limit to 10 IOCs

# ---- Fallbacks for AI endpoints ----

def get_fallback_suggestions():
    return [
        "Review evidence integrity",
        "Assign collaborators to unresolved tasks",
        "Check timeline for missing events",
        "Verify chain of custody records",
        "Document initial findings",
        "Identify key events in the timeline",
    ]

def get_fallback_severity():
    return {
        "severity": "medium",
        "confidence": 0.65
    }

def get_fallback_tags():
    return ["investigation", "analysis", "forensics"]

def get_fallback_next_steps(case_id: str):
    return [
        f"Case {case_id}: Collect additional logs from endpoints",
        f"Case {case_id}: Escalate to incident response team",
        f"Case {case_id}: Draft preliminary incident report",
        f"Case {case_id}: Conduct malware analysis",
        f"Case {case_id}: Review user activity around incident time"
    ]

def get_fallback_analysis(event_text, case_id):
    return f"Analysis for case {case_id}: This event requires further investigation. Key areas to examine include timeline correlation, evidence validation, and impact assessment."

def get_fallback_recommendations():
    return [
        "Review all collected evidence",
        "Correlate events across timeline",
        "Validate IOC relationships",
        "Document investigation methodology",
        "Prepare summary report"
    ]

# ---- AI TIMELINE ENDPOINTS ----
@app.route('/api/v1/ai/suggestions', methods=['POST'])
def ai_suggestions():
    try:
        data = request.json or {}
        input_text = data.get('input_text', '')
        case_id = data.get('case_id', '')
        
        if generator is None:
            return jsonify({
                'success': True,  # Still success with fallback
                'suggestions': get_fallback_suggestions(),
                'fallback': True,
                'message': 'AI model not loaded, using fallback suggestions'
            })
        
        prompt = f"Case {case_id}: Suggest investigation timeline events for: {input_text}. Provide 3-5 specific suggestions."
        result = generator(prompt, max_length=120, num_return_sequences=1)
        
        # Parse the result into multiple suggestions
        generated_text = result[0]['generated_text']
        suggestions = parse_suggestions(generated_text)
        
        # If no suggestions were parsed, use fallback
        if not suggestions or len(suggestions) < 2:
            suggestions = get_fallback_suggestions()
            fallback = True
        else:
            fallback = False
        
        return jsonify({
            'success': True,
            'suggestions': suggestions,
            'fallback': fallback
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/suggestions: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,  # Still return success with fallback
            'suggestions': get_fallback_suggestions(),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/severity', methods=['POST'])
def ai_severity():
    try:
        data = request.json or {}
        description = data.get('description', '')
        
        if generator is None:
            fallback_result = get_fallback_severity()
            return jsonify({
                'success': True,
                'recommended_severity': fallback_result['severity'],
                'confidence': fallback_result['confidence'],
                'fallback': True
            })
        
        prompt = f"Analyze severity for cybersecurity event: {description}. Respond with only one word: low, medium, high, or critical"
        result = generator(prompt, max_length=30, num_return_sequences=1)
        
        generated_text = result[0]['generated_text'].lower()
        severity = extract_severity(generated_text)
        
        return jsonify({
            'success': True,
            'recommended_severity': severity,
            'confidence': 0.8,
            'fallback': False
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/severity: {e}", file=sys.stderr, flush=True)
        fallback_result = get_fallback_severity()
        return jsonify({
            'success': True,
            'recommended_severity': fallback_result['severity'],
            'confidence': fallback_result['confidence'],
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/tags', methods=['POST'])
def ai_tags():
    try:
        data = request.json or {}
        description = data.get('description', '')
        
        if generator is None:
            return jsonify({
                'success': True,
                'tags': get_fallback_tags(),
                'fallback': True
            })
        
        prompt = f"Suggest tags for cybersecurity event: {description}. Respond with comma-separated tags."
        result = generator(prompt, max_length=60, num_return_sequences=1)
        
        generated_text = result[0]['generated_text']
        tags = extract_tags(generated_text)
        
        return jsonify({
            'success': True,
            'tags': tags,
            'fallback': False
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/tags: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'tags': get_fallback_tags(),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/analyze-event', methods=['POST'])
def ai_analyze_event():
    try:
        data = request.json or {}
        event_text = data.get('event_text', '')
        case_id = data.get('case_id', '')
        
        if generator is None:
            return jsonify({
                'success': True,
                'analysis': get_fallback_analysis(event_text, case_id),
                'fallback': True
            })
        
        prompt = f"Analyze DFIR event for case {case_id}: {event_text}. Provide concise analysis."
        result = generator(prompt, max_length=150, num_return_sequences=1)
        
        return jsonify({
            'success': True,
            'analysis': result[0]['generated_text'],
            'fallback': False
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/analyze-event: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'analysis': get_fallback_analysis(event_text, case_id),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/cases/<case_id>/next-steps', methods=['GET'])
def ai_next_steps(case_id):
    try:
        if generator is None:
            return jsonify({
                'success': True,
                'suggestions': get_fallback_next_steps(case_id),
                'fallback': True
            })
        
        prompt = f"Suggest next DFIR investigation steps for case {case_id}. Provide 3-5 specific steps."
        result = generator(prompt, max_length=120, num_return_sequences=1)
        
        generated_text = result[0]['generated_text']
        steps = parse_next_steps(generated_text)
        
        # If no steps parsed, use fallback
        if not steps or len(steps) < 2:
            steps = get_fallback_next_steps(case_id)
            fallback = True
        else:
            fallback = False
        
        return jsonify({
            'success': True,
            'suggestions': steps,
            'fallback': fallback
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/next-steps: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'suggestions': get_fallback_next_steps(case_id),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/recommendations', methods=['POST'])
def ai_recommendations():
    try:
        data = request.json or {}
        history = data.get('history', [])
        events = [h.get('input_text', '') for h in history[:5]]  # Reduced from 10 to 5

        if generator is None:
            return jsonify({
                'success': True,
                'recommendations': get_fallback_recommendations(),
                'fallback': True
            })

        prompt = "Based on this DFIR investigation timeline, recommend 3-5 next actions.\n"
        prompt += "Focus on gaps in investigation, missing evidence, or next logical steps.\n\n"
        prompt += "Investigation events:\n" + "\n".join(events) + "\n\nRecommendations:"

        result = generator(prompt, max_length=150, num_return_sequences=1)
        generated_text = result[0]['generated_text']
        recs = parse_suggestions(generated_text)

        # Fallback if no recommendations
        if not recs or len(recs) < 2:
            recs = get_fallback_recommendations()
            fallback = True
        else:
            fallback = False

        return jsonify({
            'success': True,
            'recommendations': recs,
            'fallback': fallback
        })

    except Exception as e:
        print(f"[LocalAI] Error in /ai/recommendations: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'recommendations': get_fallback_recommendations(),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/iocs', methods=['POST'])
def ai_iocs():
    try:
        data = request.json or {}
        text = data.get('text', '')

        if generator is None:
            return jsonify({
                'success': True,
                'iocs': [],
                'fallback': True,
                'message': 'AI model not available for IOC extraction'
            })

        prompt = f"""Extract indicators of compromise (IOCs) from this text.
                    Return them in JSON format with type, value, and confidence.
                    Types: ip, domain, hash, url, email, filename

                    Text: {text}

                    IOCs:"""

        result = generator(prompt, max_length=200, num_return_sequences=1)
        generated_text = result[0]['generated_text']

        try:
            # Try to extract JSON from the response
            start_idx = generated_text.find('[')
            end_idx = generated_text.rfind(']') + 1
            if start_idx != -1 and end_idx != 0:
                json_text = generated_text[start_idx:end_idx]
                iocs = json.loads(json_text)
            else:
                iocs = simple_parse_iocs(generated_text)
        except Exception as json_error:
            print(f"[LocalAI] JSON parse error: {json_error}", file=sys.stderr, flush=True)
            iocs = simple_parse_iocs(generated_text)

        return jsonify({
            'success': True,
            'iocs': iocs,
            'fallback': False
        })

    except Exception as e:
        print(f"[LocalAI] Error in /ai/iocs: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'iocs': [],
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/correlate-evidence', methods=['POST'])
def correlate_evidence():
    try:
        data = request.get_json() or {}
        case_id = data.get("case_id", "")
        event_description = data.get("event_description", "")

        if not case_id or not event_description:
            return jsonify({
                "success": False,
                "error": "case_id and event_description are required"
            }), 400

        # Simple correlation logic (replace with actual AI service if available)
        correlated_evidence = {
            "case_id": case_id,
            "related_events": [],
            "suggested_connections": f"Analyzing event: {event_description}",
            "confidence": 0.7
        }

        return jsonify({
            "success": True,
            "correlated_evidence": correlated_evidence,
            "fallback": True if generator is None else False
        }), 200

    except Exception as e:
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

#---- WORD COMPLETION ENDPOINTS ----
@app.route('/api/v1/ai/complete-word', methods=['POST'])
def complete_word():
    """Suggest the next word or few words based on current typing"""
    try:
        data = request.json or {}
        partial_text = data.get('text', '').strip()
        max_suggestions = data.get('max_suggestions', 2)
        
        if not partial_text:
            return jsonify({
                'success': True,
                'suggestions': [],
                'message': 'No text provided for completion'
            })
        
        if completer is None:
            return jsonify({
                'success': True,
                'suggestions': get_fallback_word_suggestions(partial_text),
                'fallback': True
            })
        
        # Clean the input text
        cleaned_text = re.sub(r'\s+', ' ', partial_text).strip()
        
        # Generate completion suggestions
        prompt = cleaned_text
        result = completer(
            prompt, 
            max_length=len(prompt.split()) + 5,  # Add a few more tokens
            num_return_sequences=1,
            do_sample=True,
            temperature=0.7,  # More creative suggestions
            pad_token_id=50256  # GPT2 pad token
        )
        
        generated_text = result[0]['generated_text']
        
        # Extract the completion part (text after the original prompt)
        if generated_text.startswith(cleaned_text):
            completion = generated_text[len(cleaned_text):].strip()
        else:
            completion = generated_text.strip()
        
        # Split into word suggestions
        suggestions = generate_word_suggestions(cleaned_text, completion, max_suggestions)
        
        return jsonify({
            'success': True,
            'suggestions': suggestions,
            'original_text': partial_text,
            'fallback': False
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/complete-word: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'suggestions': get_fallback_word_suggestions(data.get('text', '') if 'data' in locals() else ''),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/complete-sentence', methods=['POST'])
def complete_sentence():
    """Suggest sentence completions"""
    try:
        data = request.json or {}
        partial_text = data.get('text', '').strip()
        max_completions = data.get('max_completions', 2)
        
        if not partial_text:
            return jsonify({
                'success': True,
                'completions': [],
                'message': 'No text provided for completion'
            })
        
        if completer is None:
            return jsonify({
                'success': True,
                'completions': get_fallback_sentence_completions(partial_text),
                'fallback': True
            })
        
        # Generate multiple completions
        prompt = partial_text
        results = completer(
            prompt,
            max_length=len(prompt.split()) + 15,  # Longer completions
            num_return_sequences=max_completions,
            do_sample=True,
            temperature=0.8,
            pad_token_id=50256
        )
        
        completions = []
        for result in results:
            full_text = result['generated_text']
            if full_text.startswith(prompt):
                completion = full_text[len(prompt):].strip()
                # Clean up the completion
                completion = re.split(r'[.!?]', completion)[0]  # Take until first sentence end
                if completion and len(completion) > 3:
                    completions.append(completion)
        
        # Remove duplicates and limit
        unique_completions = list(dict.fromkeys(completions))[:max_completions]
        
        return jsonify({
            'success': True,
            'completions': unique_completions,
            'original_text': partial_text,
            'fallback': False
        })
        
    except Exception as e:
        print(f"[LocalAI] Error in /ai/complete-sentence: {e}", file=sys.stderr, flush=True)
        return jsonify({
            'success': True,
            'completions': get_fallback_sentence_completions(partial_text),
            'fallback': True,
            'error': str(e)
        })

@app.route('/api/v1/ai/stream-completion', methods=['POST'])
def stream_completion():
    """Stream completion word by word (for real-time typing)"""
    def generate():
        try:
            data = request.json or {}
            partial_text = data.get('text', '').strip()
            
            if not partial_text or completer is None:
                yield f"data: {json.dumps({'success': False, 'error': 'Invalid request or model not available'})}\n\n"
                return
            
            prompt = partial_text
            # Generate tokens one by one for streaming
            result = completer(
                prompt,
                max_length=len(prompt.split()) + 10,
                num_return_sequences=1,
                do_sample=True,
                temperature=0.7,
                return_full_text=False  # Only return the completion part
            )
            
            completion = result[0]['generated_text'].strip()
            words = completion.split()
            
            for i, word in enumerate(words):
                if i >= 8:  # Limit stream to 8 words
                    break
                yield f"data: {json.dumps({'word': word, 'completed': False})}\n\n"
            
            yield f"data: {json.dumps({'completed': True, 'total_words': min(len(words), 8)})}\n\n"
            
        except Exception as e:
            yield f"data: {json.dumps({'success': False, 'error': str(e)})}\n\n"
    
    return Response(stream_with_context(generate()), mimetype='text/plain')

# Helper functions for word completion
def generate_word_suggestions(original_text, completion, max_suggestions=3):
    """Generate word suggestions from completion text"""
    if not completion:
        return []
    
    # Extract the next few words
    words = completion.split()
    suggestions = []
    
    # Create suggestions of 1-3 words
    for i in range(min(3, len(words))):
        suggestion = ' '.join(words[:i+1])
        if suggestion and len(suggestion) > 1:
            suggestions.append(suggestion)
    
    # Also suggest single next word
    if words and words[0] not in suggestions:
        suggestions.insert(0, words[0])
    
    return suggestions[:max_suggestions]

def get_fallback_word_suggestions(partial_text):
    """Fallback word suggestions based on common DFIR terminology"""
    dfir_terms = [
        "investigation", "analysis", "evidence", "timeline", "incident",
        "malware", "network", "forensic", "digital", "response",
        "detection", "prevention", "recovery", "assessment", "evaluation",
        "examination", "documentation", "verification", "validation", "reviewed", "review",

    ]
    
    last_word = partial_text.split()[-1].lower() if partial_text.split() else ""
    
    # Filter terms that might logically follow the last word
    suggestions = []
    for term in dfir_terms:
        if term not in partial_text.lower():
            suggestions.append(term)
    
    return suggestions[:3]

# def get_fallback_sentence_completions(partial_text):
#     """Fallback sentence completions"""
#     common_completions = [
#         "requires further investigation and analysis.",
#         "should be documented in the incident report.",
#         "needs to be correlated with other timeline events.",
#         "must be verified through additional evidence.",
#         "appears to be a key event in the investigation.",
#         "is critical for understanding the incident scope.",
#         "should be escalated to the incident response team.",
#         "indicates potential compromise of the system.",
#         "warrants a deeper forensic examination.",
#         "is consistent with known attack patterns.",
#         "must be further investigated for root cause.",
#         "is essential for building the case timeline.",
#         "should be cross-referenced with IOC databases.",
#         "is a significant finding in the overall analysis.",
#         "may provide insights into attacker behavior.",
#         "is a common tactic used in phishing attacks.",
#         "should be prioritized for immediate action.",
#         "is often overlooked but crucial for evidence.",
#         "needs to be reviewed by senior analysts.",
#         "could be linked to other suspicious activities.",
#         "should be cross-referenced with log files.",
#         "needs validation through endpoint analysis.",
#         "may indicate lateral movement in the network.",
#         "should be documented for chain of custody purposes.",
#         "This event is pivotal in understanding the breach.",
#         "Further analysis is required to determine its impact.",
#         "Collaboration with other teams may yield additional insights.",
#         "Cross-referencing with threat intelligence could be beneficial.",
#         "Documenting this thoroughly will aid in future investigations.",
#         "Verifying the integrity of this evidence is crucial.",
#         "Additional context from logs may clarify this event.", 
#         "This evidence file requires our urgent attention.",
#         "Correlating this with other findings could be insightful.",
#         "Documenting this thoroughly will aid in future investigations.",
#         "Verifying the integrity of this evidence is crucial.",
#         "Additional context from logs may clarify this event.",
#         "and the team must coordinate closely to ensure all evidence is reviewed.",
#         "and the team must verify that all findings are properly documented.",
#         "and the team must communicate any updates immediately to all members.",
#         "and the team must follow standard operating procedures for case handling.",
#         "and the team must assign tasks based on expertise to improve efficiency.",
#         "and the team must review each other's work to avoid errors.",
#         "and the team must escalate critical findings to the lead investigator without delay.",
#         "and the team must ensure all communications are logged in the case management system.",
#         "and the team must prepare a summary of findings for review during the next meeting.",
#         "and the team must maintain a clear chain of custody for all evidence collected.",
#         "the team must reconcile conflicting information before moving forward.",
#         "the team must cross-check evidence with external sources.",
#         "the team must maintain confidentiality throughout the investigation.",
#         "the team must update the timeline as new events are discovered.",
#         "the team must ensure all digital evidence is securely stored.",
#         "the team must document any assumptions made during the analysis.",
#         "the team must validate findings with multiple sources.",
#         "awaiting admin approval in order to proceed with the next steps.",
#         "awaiting admin approval to finalize documentation and close the case.",
#         "awaiting admin approval for any additional evidence collection.",
#         "awaiting admin approval to assign new tasks to team members.",
#         "awaiting admin approval before publishing the report to stakeholders.",
#         "awaiting admin approval to update case priorities or deadlines.",
#         "awaiting admin approval to escalate the case to higher authorities.",
#         "awaiting admin approval to share findings with external partners.",
#         "awaiting admin approval to implement recommended security measures.",
#         "awaiting admin approval to archive the case after resolution.",
#         "awaiting admin approval to proceed with forensic imaging of affected systems."
#     ]
    
#     # Simple logic to choose relevant completions
#     text_lower = partial_text.lower()
#     relevant_completions = []
    
#     for completion in common_completions:
#         if any(keyword in text_lower for keyword in ['investigat', 'analys', 'review']):
#             if 'document' in completion or 'correlat' in completion:
#                 relevant_completions.append(completion)
#         elif any(keyword in text_lower for keyword in ['evidence', 'find']):
#             if 'verif' in completion or 'additional' in completion:
#                 relevant_completions.append(completion)
#         else:
#             relevant_completions.append(completion)
    
#     return relevant_completions[:5]

def get_fallback_sentence_completions(partial_text: str, max_suggestions: int = 5):
    """
    Return context-aware sentence completions for investigation events.
    Suggestions are based on keywords in the partial text.
    """
    # Categorized sentence templates
    completions_by_context = {
        "teamwork": [
            "The team must coordinate to ensure all steps are documented.",
            "The team must review findings and escalate issues as needed.",
            "The team must communicate updates to all stakeholders.",
            "The team must assign responsibilities for pending tasks.",
            "The team must verify that all evidence is properly logged.",
            "The team must reconcile conflicting information before moving forward.",
            "The team must cross-check evidence with external sources.",
            "The team must maintain confidentiality throughout the investigation.",
            "The team must update the timeline as new events are discovered.",
            "The team must ensure all digital evidence is securely stored.",
        ],
        "admin": [
            "Awaiting admin approval in order to proceed with the next steps.",
            "Awaiting admin approval to finalize documentation and close the case.",
            "Awaiting admin approval for any additional evidence collection.",
            "Awaiting admin approval to assign new tasks to team members.",
            "Awaiting admin approval before publishing the report to stakeholders.",
            "Awaiting admin approval to update case priorities or deadlines.",
            "Awaiting admin approval to escalate the case to higher authorities.",
            "Awaiting admin approval to share findings with external partners.",
            "Awaiting admin approval to implement recommended security measures.",
            "Awaiting admin approval to archive the case after resolution.",
        ],
        "evidence": [
            "Requires further investigation and analysis.",
            "Should be documented in the incident report.",
            "Needs to be correlated with other timeline events.",
            "Must be verified through additional evidence.",
            "Appears to be a key event in the investigation.",
            "Is critical for understanding the incident scope.",
            "Should be escalated to the incident response team.",
            "Indicates potential compromise of the system.",
            "Warrants a deeper forensic examination.",
            "Is consistent with known attack patterns.",
            "Must be further investigated for root cause.",
            "Is essential for building the case timeline.",
            "Should be cross-referenced with IOC databases.",
            "Is a significant finding in the overall analysis.",
            "May provide insights into attacker behavior.",
            "Should be prioritized for immediate action.",
            "Needs to be reviewed by senior analysts.",
            "Could be linked to other suspicious activities.",
            "Should be cross-referenced with log files.",
            "Needs validation through endpoint analysis.",
            "May indicate lateral movement in the network.",
            "Should be documented for chain of custody purposes.",
            "This event is pivotal in understanding the breach.",
            "Further analysis is required to determine its impact.",
            "Verifying the integrity of this evidence is crucial.",
            "Additional context from logs may clarify this event.",
        ],
        "general": [
            "Collaboration with other teams may yield additional insights.",
            "Cross-referencing with threat intelligence could be beneficial.",
            "Documenting this thoroughly will aid in future investigations.",
        ]
    }

    # Keyword mapping to contexts
    keyword_mapping = {
        "teamwork": ["team", "members", "collaborate", "coordinate", "assign", "review", "update"],
        "admin": ["admin", "approval", "authorize", "lead", "confirm", "escalate"],
        "evidence": ["evidence", "investigat", "find", "log", "verify", "document", "analysis", "incident"],
    }

    text_lower = partial_text.lower()
    matched_contexts = set()

    # Determine which contexts match the input
    for context, keywords in keyword_mapping.items():
        if any(kw in text_lower for kw in keywords):
            matched_contexts.add(context)

    # If no context matches, default to general + evidence
    if not matched_contexts:
        matched_contexts = {"general", "evidence"}

    # Collect relevant completions
    relevant_completions = []
    for context in matched_contexts:
        relevant_completions.extend(completions_by_context.get(context, []))

    # Remove duplicates while preserving order
    seen = set()
    filtered_completions = []
    for comp in relevant_completions:
        if comp not in seen:
            filtered_completions.append(comp)
            seen.add(comp)

    return filtered_completions[:max_suggestions]


if __name__ == '__main__':
    print("[LocalAI] Flask app starting on 0.0.0.0:5000", flush=True)
    app.run(host='0.0.0.0', port=5000, debug=True)