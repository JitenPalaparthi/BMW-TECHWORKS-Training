import json
import os
from pathlib import Path

import numpy as np
import requests

OLLAMA_URL = os.getenv("OLLAMA_URL", "http://localhost:11434")
EMBEDDING_MODEL = os.getenv("EMBEDDING_MODEL", "embeddinggemma")
LLM_MODEL = os.getenv("LLM_MODEL", "gemma3")

DOCUMENT_DIRECTORY = "documents"
VECTOR_STORE_FILE = "vector_store.json"

CHUNK_SIZE = 500
CHUNK_OVERLAP = 100
TOP_K = 3


def get_embedding(text: str) -> list[float]:
    response = requests.post(
        f"{OLLAMA_URL}/api/embed",
        json={"model": EMBEDDING_MODEL, "input": text},
        timeout=120,
    )
    response.raise_for_status()
    return response.json()["embeddings"][0]


def generate_answer(question: str, context: str) -> str:
    # gaurd rails
    system_prompt = """You are a Retrieval Augmented Generation assistant.
Answer the user's question using ONLY the supplied context.

Rules:
1. Do not use outside knowledge.
2. If the answer is not present in the context, say:
   "I don't know based on the provided documents."
3. Do not invent information.
4. Give a concise but complete answer.
5. Mention the source document when useful.
"""

    user_prompt = f"""CONTEXT
=======
{context}

QUESTION
========
{question}
"""

    response = requests.post(
        f"{OLLAMA_URL}/api/chat",
        json={
            "model": LLM_MODEL,
            "messages": [
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": user_prompt},
            ],
            "stream": False,
            "options": {"temperature": 0.1},
        },
        timeout=300,
    )
    response.raise_for_status()
    return response.json()["message"]["content"]


def load_documents(directory: str) -> list[dict]:
    documents = []
    for path in Path(directory).glob("*.txt"):
        documents.append(
            {
                "source": path.name,
                "text": path.read_text(encoding="utf-8"),
            }
        )
    return documents


def chunk_text(text: str, chunk_size: int = CHUNK_SIZE, overlap: int = CHUNK_OVERLAP) -> list[str]:
    if overlap >= chunk_size:
        raise ValueError("overlap must be smaller than chunk_size")

    chunks = []
    start = 0
    while start < len(text):
        end = min(start + chunk_size, len(text))
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        if end >= len(text):
            break
        start += chunk_size - overlap
    return chunks


def create_index():
    print("\n=== Creating RAG index ===")
    documents = load_documents(DOCUMENT_DIRECTORY)

    if not documents:
        raise RuntimeError(f"No .txt documents found in {DOCUMENT_DIRECTORY}/")

    vector_store = []

    for document in documents:
        print(f"\nProcessing: {document['source']}")
        chunks = chunk_text(document["text"])
        print(f"Chunks: {len(chunks)}")

        for chunk_id, chunk in enumerate(chunks):
            print(f"  Embedding chunk {chunk_id + 1}/{len(chunks)}")
            vector_store.append(
                {
                    "source": document["source"],
                    "chunk_id": chunk_id,
                    "text": chunk,
                    "embedding": get_embedding(chunk),
                }
            )

    Path(VECTOR_STORE_FILE).write_text(
        json.dumps(vector_store),
        encoding="utf-8",
    )
    print(f"\nIndexed {len(vector_store)} chunks into {VECTOR_STORE_FILE}")


def load_vector_store() -> list[dict]:
    return json.loads(Path(VECTOR_STORE_FILE).read_text(encoding="utf-8"))


def cosine_similarity(vector_a: list[float], vector_b: list[float]) -> float:
    a = np.asarray(vector_a, dtype=np.float32)
    b = np.asarray(vector_b, dtype=np.float32)
    denominator = np.linalg.norm(a) * np.linalg.norm(b)
    if denominator == 0:
        return 0.0
    return float(np.dot(a, b) / denominator)


def retrieve(question: str, vector_store: list[dict], top_k: int = TOP_K) -> list[dict]:
    question_embedding = get_embedding(question)
    results = []

    for item in vector_store:
        score = cosine_similarity(question_embedding, item["embedding"])
        results.append(
            {
                "source": item["source"],
                "chunk_id": item["chunk_id"],
                "text": item["text"],
                "score": score,
            }
        )

    results.sort(key=lambda item: item["score"], reverse=True)
    return results[:top_k]


def build_context(results: list[dict]) -> str:
    sections = []
    for i, result in enumerate(results, start=1):
        sections.append(
            f"""Retrieved Document {i}
Source: {result["source"]}
Chunk: {result["chunk_id"]}
Similarity: {result["score"]:.4f}

{result["text"]}
"""
        )
    return "\n------------------------\n".join(sections)


def rag(question: str, vector_store: list[dict]):
    results = retrieve(question, vector_store)

    print("\n=== Retrieved chunks ===")
    for i, result in enumerate(results, start=1):
        print(
            f"\n#{i} source={result['source']} "
            f"chunk={result['chunk_id']} "
            f"similarity={result['score']:.4f}"
        )
        print(result["text"])

    context = build_context(results)
    answer = generate_answer(question, context)

    print("\n=== Final answer ===")
    print(answer)


def check_ollama() -> bool:
    try:
        response = requests.get(f"{OLLAMA_URL}/api/tags", timeout=10)
        response.raise_for_status()
        model_names = [m["name"] for m in response.json().get("models", [])]

        print(f"Ollama: {OLLAMA_URL}")
        print("Models:")
        for name in model_names:
            print(f"  - {name}")

        required = [LLM_MODEL, EMBEDDING_MODEL]
        missing = [
            model for model in required
            if not any(name == model or name.startswith(model + ":") for name in model_names)
        ]
        if missing:
            print(f"\nWarning: missing expected model(s): {', '.join(missing)}")

        return True
    except requests.RequestException as exc:
        print(f"Cannot connect to Ollama at {OLLAMA_URL}: {exc}")
        return False


def main():
    print("=== Ollama RAG Demo ===")

    if not check_ollama():
        return

    if not Path(VECTOR_STORE_FILE).exists():
        create_index()

    vector_store = load_vector_store()
    print(f"\nLoaded {len(vector_store)} indexed chunks.")
    print("Commands: /reindex, /exit")

    while True:
        question = input("\nQuestion > ").strip()
        if not question:
            continue
        if question.lower() == "/exit":
            break
        if question.lower() == "/reindex":
            create_index()
            vector_store = load_vector_store()
            continue

        try:
            rag(question, vector_store)
        except Exception as exc:
            print(f"\nError: {exc}")


if __name__ == "__main__":
    main()
