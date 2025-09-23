from flask import Flask, request, jsonify
from transformers import pipeline
import sys

app = Flask(__name__)

try:
    print("[LocalAI] Flan-t5 model loading...", flush=True)
    generator = pipeline("text2text-generation", model="google/flan-t5-base")
    print("[LocalAI] Model loaded successfully.", flush=True)
except Exception as e:
    print(f"[LocalAI] Error loading model: {e}", file=sys.stderr, flush=True)
    generator = None


@app.route("/generate", methods=["POST"])
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


if __name__ == "__main__":
    print("[LocalAI] Flask app starting on 0.0.0.0:5000", flush=True)
    app.run(host="0.0.0.0", port=5000)
