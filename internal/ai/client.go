package ai

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "os"
        "strings"
        "time"
)

// Provider represents different AI/LLM providers
type Provider string

const (
        OpenAI    Provider = "openai"
        Claude    Provider = "claude"
        Grok      Provider = "grok"
        Gemini    Provider = "gemini"
)

// Client represents an AI client for metrics analysis
type Client struct {
        Provider Provider
        APIKey   string
        BaseURL  string
        Model    string
}

// MetricsAnalysis represents the AI analysis result
type MetricsAnalysis struct {
        Issues          []string `json:"issues"`
        Recommendations []string `json:"recommendations"`
        RootCauses      []string `json:"root_causes"`
        Summary         string   `json:"summary"`
}

// NewClient creates a new AI client based on the provider
func NewClient(provider Provider) (*Client, error) {
        var apiKey, baseURL, model string

        switch provider {
        case OpenAI:
                apiKey = os.Getenv("OPENAI_API_KEY")
                baseURL = "https://api.openai.com/v1/chat/completions"
                model = "gpt-4o"
        case Claude:
                apiKey = os.Getenv("ANTHROPIC_API_KEY")
                baseURL = "https://api.anthropic.com/v1/messages"
                model = "claude-3-sonnet-20240229"
        case Grok:
                apiKey = os.Getenv("XAI_API_KEY")
                baseURL = "https://api.x.ai/v1/chat/completions"
                model = "grok-beta"
        case Gemini:
                apiKey = os.Getenv("GOOGLE_API_KEY")
                baseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
                model = "gemini-pro"
        default:
                return nil, fmt.Errorf("unsupported provider: %s", provider)
        }

        if apiKey == "" {
                var envVar string
                switch provider {
                case OpenAI:
                        envVar = "OPENAI_API_KEY"
                case Claude:
                        envVar = "ANTHROPIC_API_KEY"
                case Grok:
                        envVar = "XAI_API_KEY"
                case Gemini:
                        envVar = "GOOGLE_API_KEY"
                }
                return nil, fmt.Errorf("API key not found for %s. Please set environment variable: %s", provider, envVar)
        }

        return &Client{
                Provider: provider,
                APIKey:   apiKey,
                BaseURL:  baseURL,
                Model:    model,
        }, nil
}

// AnalyzeMetrics sends metrics data to AI for analysis
func (c *Client) AnalyzeMetrics(metrics []Metric) (*MetricsAnalysis, error) {
        prompt := c.buildAnalysisPrompt(metrics)

        switch c.Provider {
        case OpenAI, Grok:
                return c.callOpenAIAPI(prompt)
        case Claude:
                return c.callClaudeAPI(prompt)
        case Gemini:
                return c.callGeminiAPI(prompt)
        default:
                return nil, fmt.Errorf("unsupported provider: %s", c.Provider)
        }
}

// buildAnalysisPrompt creates a structured prompt for metrics analysis
func (c *Client) buildAnalysisPrompt(metrics []Metric) string {
        // Exclude verbose metric prefixes to reduce token usage
        EXCLUDE_METRICS_PREFIX := []string{
                // "kafka_server_fetcherlagmetrics",       // Consumer lag metrics - very verbose
                // "kafka_server_replicafetchermanager",   // Replica fetch manager metrics
                // "kafka_network_requestmetrics",         // Network request metrics - high volume
                // "kafka_controller_controllerstats",     // Controller statistics
                // "kafka_server_delayedoperationpurgatory", // Delayed operation metrics
        }

        // Convert metrics to compact JSON format to reduce token usage
        type MetricData struct {
                Kafka   []map[string]string `json:"kafka_metrics,omitempty"`
                JVM     []map[string]string `json:"jvm_metrics,omitempty"`
                Process []map[string]string `json:"process_metrics,omitempty"`
        }

        data := MetricData{}

        // Group and format metrics compactly
        for _, m := range metrics {
                // Skip metrics with excluded prefixes
                shouldExclude := false
                for _, prefix := range EXCLUDE_METRICS_PREFIX {
                        if strings.HasPrefix(m.Name, prefix) {
                                shouldExclude = true
                                break
                        }
                }
                if shouldExclude {
                        continue
                }

                // Remove common prefixes to shorten metric names
                shortName := m.Name
                // if strings.HasPrefix(shortName, "kafka_server_") {
                //         shortName = strings.TrimPrefix(shortName, "kafka_server_")
                // } else if strings.HasPrefix(shortName, "kafka_") {
                //         shortName = strings.TrimPrefix(shortName, "kafka_")
                // } else if strings.HasPrefix(shortName, "jvm_") {
                //         shortName = strings.TrimPrefix(shortName, "jvm_")
                // } else if strings.HasPrefix(shortName, "process_") {
                //         shortName = strings.TrimPrefix(shortName, "process_")
                // }

                metric := map[string]string{
                        "name": shortName,
                        "value": m.Value,
                }
                if m.Labels != "" {
                        metric["labels"] = m.Labels
                }

                if strings.HasPrefix(m.Name, "kafka_") {
                        data.Kafka = append(data.Kafka, metric)
                } else if strings.HasPrefix(m.Name, "jvm_") {
                        data.JVM = append(data.JVM, metric)
                } else if strings.HasPrefix(m.Name, "process_") {
                        data.Process = append(data.Process, metric)
                }
        }

        // Convert to compact JSON
        jsonData, _ := json.Marshal(data)

        return fmt.Sprintf(`Role: Kafka monitoring expert. Analyze broker metrics JSON (n=name,v=value,l=labels).

%s

IMPORTANT: Return ONLY valid JSON in this EXACT format with NO additional text:
{
  "issues": ["string1", "string2"],
  "root_causes": ["string1", "string2"], 
  "recommendations": ["string1", "string2"],
  "summary": "string"
}

Rules:
- All fields must be present
- issues/root_causes/recommendations must be arrays of strings (not objects)
- summary must be a single string
- No markdown, no explanations, no extra fields`, string(jsonData))
}

// Metric represents a parsed metric for AI analysis
type Metric struct {
        Name   string
        Labels string
        Value  string
}

// OpenAI/Grok API call
func (c *Client) callOpenAIAPI(prompt string) (*MetricsAnalysis, error) {
        payload := map[string]interface{}{
                "model": c.Model,
                "messages": []map[string]string{
                        {
                                "role":    "system",
                                "content": "You are a Kafka monitoring expert. Analyze broker metrics and provide concise recommendations.",
                        },
                        {
                                "role":    "user",
                                "content": prompt,
                        },
                },
                "response_format": map[string]string{
                        "type": "json_object",
                },
        }
        
        // Only add max_tokens and temperature for models below GPT-5
        // GPT-5, GPT-5-mini, GPT-5-nano don't support these parameters
        if !strings.HasPrefix(strings.ToLower(c.Model), "gpt-5") && !strings.Contains(strings.ToLower(c.Model), "o1") {
                payload["max_tokens"] = 1500
                payload["temperature"] = 0.3
        }

        return c.makeAPICall(payload, func(body []byte) (*MetricsAnalysis, error) {
                var response struct {
                        Choices []struct {
                                Message struct {
                                        Content string `json:"content"`
                                } `json:"message"`
                        } `json:"choices"`
                }

                if err := json.Unmarshal(body, &response); err != nil {
                        return nil, err
                }

                if len(response.Choices) == 0 {
                        return nil, fmt.Errorf("no response from API")
                }

                return c.parseAnalysisResponse(response.Choices[0].Message.Content)
        })
}

// Claude API call
func (c *Client) callClaudeAPI(prompt string) (*MetricsAnalysis, error) {
        // Update prompt to ensure JSON response
        jsonPrompt := prompt + "\n\nRespond only with valid JSON in the exact format specified."
        
        payload := map[string]interface{}{
                "model":      c.Model,
                "max_tokens": 1500,
                "system":     "You are a Kafka monitoring expert. Analyze broker metrics and provide concise recommendations. Always respond with valid JSON only.",
                "messages": []map[string]string{
                        {
                                "role":    "user",
                                "content": jsonPrompt,
                        },
                },
        }

        return c.makeAPICall(payload, func(body []byte) (*MetricsAnalysis, error) {
                var response struct {
                        Content []struct {
                                Text string `json:"text"`
                        } `json:"content"`
                }

                if err := json.Unmarshal(body, &response); err != nil {
                        return nil, err
                }

                if len(response.Content) == 0 {
                        return nil, fmt.Errorf("no response from API")
                }

                return c.parseAnalysisResponse(response.Content[0].Text)
        })
}

// Gemini API call
func (c *Client) callGeminiAPI(prompt string) (*MetricsAnalysis, error) {
        systemPrompt := "You are a Kafka monitoring expert. Analyze broker metrics and provide concise recommendations. Always respond with valid JSON only."
        jsonPrompt := prompt + "\n\nRespond only with valid JSON in the exact format specified."
        fullPrompt := systemPrompt + "\n\n" + jsonPrompt

        payload := map[string]interface{}{
                "contents": []map[string]interface{}{
                        {
                                "parts": []map[string]string{
                                        {"text": fullPrompt},
                                },
                        },
                },
                "generationConfig": map[string]interface{}{
                        "response_mime_type": "application/json",
                },
        }

        return c.makeAPICall(payload, func(body []byte) (*MetricsAnalysis, error) {
                var response struct {
                        Candidates []struct {
                                Content struct {
                                        Parts []struct {
                                                Text string `json:"text"`
                                        } `json:"parts"`
                                } `json:"content"`
                        } `json:"candidates"`
                }

                if err := json.Unmarshal(body, &response); err != nil {
                        return nil, err
                }

                if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
                        return nil, fmt.Errorf("no response from API")
                }

                return c.parseAnalysisResponse(response.Candidates[0].Content.Parts[0].Text)
        })
}

// makeAPICall handles the HTTP request logic
func (c *Client) makeAPICall(payload interface{}, responseParser func([]byte) (*MetricsAnalysis, error)) (*MetricsAnalysis, error) {
        jsonData, err := json.Marshal(payload)
        if err != nil {
                return nil, err
        }

        req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
        if err != nil {
                return nil, err
        }

        // Set headers based on provider
        switch c.Provider {
        case OpenAI, Grok:
                req.Header.Set("Authorization", "Bearer "+c.APIKey)
                req.Header.Set("Content-Type", "application/json")
        case Claude:
                req.Header.Set("x-api-key", c.APIKey)
                req.Header.Set("Content-Type", "application/json")
                req.Header.Set("anthropic-version", "2023-06-01")
        case Gemini:
                req.Header.Set("Content-Type", "application/json")
                // API key is added to URL for Gemini
                c.BaseURL += "?key=" + c.APIKey
        }

        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
                return nil, err
        }

        if resp.StatusCode != http.StatusOK {
                return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
        }

        return responseParser(body)
}

// parseAnalysisResponse extracts JSON from the AI response
func (c *Client) parseAnalysisResponse(content string) (*MetricsAnalysis, error) {
        // Try to find JSON in the response
        start := strings.Index(content, "{")
        end := strings.LastIndex(content, "}") + 1

        if start == -1 || end <= start {
                return nil, fmt.Errorf("no valid JSON found in response")
        }

        jsonStr := content[start:end]

        var analysis MetricsAnalysis
        if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
                return nil, fmt.Errorf("failed to parse analysis response: %w", err)
        }

        return &analysis, nil
}