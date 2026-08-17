docker run -d \
  --name ollama \
  -p 11434:11434 \
  -v ollama:/root/.ollama \
  ollama/ollama



 docker exec -it ollama ollama pull llama3.2

 docker exec -it ollama ollama run llama3.2

 curl http://localhost:11434/api/generate \
  -d '{
    "model": "llama3.2",
    "prompt": "Explain vectors using simple mathematics",
    "stream": true 
  }'
