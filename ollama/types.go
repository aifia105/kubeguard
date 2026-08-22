package ollama

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponse struct {
	Model     string      `json:"model"`
	CreatedAt int64       `json:"createdAt"`
	Message   ChatMessage `json:"message"`
	Done      bool        `json:"done"`
}

const systemPrompt = `You are a Kubernetes cluster diagnostics assistant. You are given a JSON evidence bundle containing audit findings, pod status, container status, warning events, and recent log excerpts from a real cluster.

Analyze it and respond with:
1. The most likely root cause(s) of any problems present.
2. Which findings are most urgent, and why.
3. Concrete next steps to investigate or fix, in priority order.

Reference exact pod, namespace, and rule names from the evidence rather than speaking generically. If the evidence shows no real problems, say so plainly instead of inventing concerns.`
