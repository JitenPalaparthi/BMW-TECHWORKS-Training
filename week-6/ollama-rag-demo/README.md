# Ollama RAG Demo — Gemma3 + EmbeddingGemma

This is a small, dependency-light Retrieval-Augmented Generation (RAG) project.

It uses:

- **Ollama** as the local model server
- **gemma3** to generate the final answer
- **embeddinggemma** to create document/query embeddings
- **NumPy cosine similarity** as the simple vector-search implementation
- Dummy `.txt` documents so the example works immediately

## 1. Architecture

### Indexing

```text
documents/*.txt
      |
      v
Load documents
      |
      v
Split into chunks
      |
      v
embeddinggemma
      |
      v
Vectors + text + metadata
      |
      v
vector_store.json
```

### Query

```text
User question
      |
      v
embeddinggemma
      |
      v
Question vector
      |
      v
Cosine similarity against document vectors
      |
      v
Top 3 chunks
      |
      v
Context + question
      |
      v
gemma3
      |
      v
Final answer
```

## 2. Important idea

`embeddinggemma` does NOT answer the question.

It converts text to vectors.

`gemma3` does NOT perform the vector search in this project.

It receives the retrieved text and generates the final answer.

So RAG has two major parts:

```text
Retrieval  -> embedding + similarity search
Generation -> Gemma3
```

## 3. Prerequisites

You said Ollama is already running in Docker.

Check:

```bash
docker ps
```

The Ollama container should publish port 11434, for example:

```text
0.0.0.0:11434->11434/tcp
```

Check installed models:

```bash
docker exec -it ollama ollama list
```

Expected models:

```text
gemma3:latest
embeddinggemma:latest
```

If your container is not named `ollama`, replace `ollama` with the actual container name.

## 4. Verify Ollama from the host

```bash
curl http://localhost:11434/api/tags
```

Test Gemma3:

```bash
curl http://localhost:11434/api/chat \
  -d '{
    "model":"gemma3",
    "messages":[
      {"role":"user","content":"Say hello in one sentence"}
    ],
    "stream":false
  }'
```

Test EmbeddingGemma:

```bash
curl http://localhost:11434/api/embed \
  -d '{
    "model":"embeddinggemma",
    "input":"Go supports goroutines"
  }'
```

## 5. Recommended: run Python on the host

From this project directory:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python rag.py
```

On Windows PowerShell:

```powershell
python -m venv .venv
.venv\Scripts\Activate.ps1
pip install -r requirements.txt
python rag.py
```

The application defaults to:

```text
OLLAMA_URL=http://localhost:11434
```

That is correct when Python runs on your host and Ollama's Docker port 11434 is published.

## 6. First run

The first run creates the vector index:

```text
=== Ollama RAG Demo ===
Ollama: http://localhost:11434

Processing: company.txt
Chunks: ...
Embedding chunk ...

Processing: golang.txt
...

Indexed ... chunks into vector_store.json
```

The generated file is:

```text
vector_store.json
```

Document embeddings are created once and stored there.

## 7. Ask questions

Try:

```text
How many paid leave days do employees get?
```

Expected relevant document:

```text
company.txt
```

Try:

```text
How much is the learning allowance?
```

Try:

```text
How does Go implement concurrency?
```

Expected relevant document:

```text
golang.txt
```

Try:

```text
Does Rust require a garbage collector for memory safety?
```

Expected relevant document:

```text
rust.txt
```

Try an unknown question:

```text
Who is the CEO of Acme Technologies?
```

The provided documents do not contain that information, so the prompt instructs Gemma3 to answer:

```text
I don't know based on the provided documents.
```

## 8. What happens when you ask a question?

Suppose you ask:

```text
How many paid leave days do employees get?
```

### Step A — Query embedding

EmbeddingGemma converts the question into a numeric vector:

```text
"How many paid leave days..."
        |
        v
embeddinggemma
        |
        v
[0.12, -0.41, 0.73, ...]
```

### Step B — Compare with stored vectors

Each document chunk already has an embedding.

The program calculates cosine similarity:

```text
Question vector <-> company chunk    high similarity
Question vector <-> Go chunk         low similarity
Question vector <-> Rust chunk       low similarity
```

### Step C — Top-K retrieval

The program sorts matches and selects the best 3 chunks.

`TOP_K = 3` is configured near the top of `rag.py`.

### Step D — Build context

The retrieved text is assembled into a context block.

### Step E — Generation

The prompt sent to Gemma3 is conceptually:

```text
CONTEXT:
Employees receive 24 paid leave days every year...

QUESTION:
How many paid leave days do employees get?
```

Gemma3 then generates the answer.

## 9. Why chunk documents?

Large documents should not normally be embedded as one huge piece.

Instead:

```text
large document
    |
    +--> chunk 1
    +--> chunk 2
    +--> chunk 3
```

This improves retrieval precision.

This demo uses:

```python
CHUNK_SIZE = 500
CHUNK_OVERLAP = 100
```

These are character counts for simplicity.

## 10. Why overlap?

Suppose an important sentence is split at a chunk boundary.

Without overlap:

```text
Chunk 1: Employees receive a learning...
Chunk 2: ...allowance of 50,000 rupees...
```

With overlap, part of the previous chunk appears in the next chunk, preserving context.

## 11. Why cosine similarity?

Embeddings represent semantic meaning as vector directions.

Cosine similarity compares those directions.

Conceptually:

```text
question: "employee leave policy"
company chunk: points in a similar semantic direction
Go chunk:      points in a different direction
```

Higher cosine similarity means the embedding vectors are more aligned.

## 12. Reindex after editing documents

If you add or modify `.txt` files in `documents/`, run:

```text
/reindex
```

inside the application.

Or delete the store:

```bash
rm vector_store.json
python rag.py
```

## 13. Commands

Inside the program:

```text
/reindex
```

Rebuild embeddings.

```text
/exit
```

Exit.

## 14. Change the Ollama URL

You can set:

```bash
export OLLAMA_URL=http://localhost:11434
```

Or:

```bash
OLLAMA_URL=http://localhost:11434 python rag.py
```

## 15. Optional: run the RAG app itself in Docker

Build:

```bash
docker compose build
```

Run interactively:

```bash
docker compose run --rm rag
```

The included Compose file uses:

```text
http://host.docker.internal:11434
```

so the RAG container can access the Ollama service published on the Docker host.

## 16. If both containers are on the same Docker network

If you later put Ollama and the RAG service into one Compose file, the cleaner configuration is:

```text
OLLAMA_URL=http://ollama:11434
```

where `ollama` is the Compose service name.

Then the architecture becomes:

```text
Docker network

+----------------------+
| ollama               |
|  gemma3              |
|  embeddinggemma      |
+----------+-----------+
           |
       :11434
           |
+----------v-----------+
| Python RAG           |
+----------------------+
```

## 17. Project files

```text
ollama-rag-demo/
├── README.md
├── rag.py
├── requirements.txt
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── documents/
    ├── company.txt
    ├── golang.txt
    └── rust.txt
```

`vector_store.json` is generated automatically after the first indexing run.

## 18. Next production upgrades

This project intentionally avoids frameworks so the RAG mechanics remain visible.

After understanding this version, useful upgrades are:

1. Replace JSON + linear NumPy search with Qdrant, Chroma, pgvector, Milvus, or another vector store.
2. Add PDF/DOCX/HTML loaders.
3. Use token-aware or semantic chunking.
4. Store page numbers and document metadata.
5. Add similarity thresholds.
6. Add reranking.
7. Return citations/sources with answers.
8. Add FastAPI.
9. Dockerize the complete stack.
10. Add RAG evaluation and observability.
