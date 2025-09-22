
from flask import Flask, request, jsonify
from transformers import pipeline
import sys

app = Flask(__name__)

try:
    print("[LocalAI] Starting model loading...", flush=True)
    generator = pipeline('text-generation', model='gpt2')
    print("[LocalAI] Model loaded successfully.", flush=True)
except Exception as e:
    print(f"[LocalAI] Error loading model: {e}", file=sys.stderr, flush=True)
    generator = None

@app.route('/generate', methods=['POST'])
def generate():
    if generator is None:
        return jsonify({'error': 'Model not loaded'}), 500
    try:
        prompt = request.json.get('prompt')
        dfir_context = (
            "You are a professional Digital Forensics and Incident Response (DFIR) report writer. "
            "Generate clear, concise, and relevant forensic report content for the following section:\n"
        )
        full_prompt = dfir_context + prompt
        result = generator(full_prompt, max_length=100)
        return jsonify({'text': result[0]['generated_text']})
    except Exception as e:
        print(f"[LocalAI] Error in /generate: {e}", file=sys.stderr, flush=True)
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    print("[LocalAI] Flask app starting on 0.0.0.0:5000", flush=True)
    app.run(host='0.0.0.0', port=5000)
