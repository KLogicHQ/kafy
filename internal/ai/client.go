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
                model = "gpt-4"
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
        var metricsData strings.Builder
        
        // Group metrics by category for better analysis
        kafkaMetrics := []Metric{}
        jvmMetrics := []Metric{}
        processMetrics := []Metric{}
        
        for _, metric := range metrics {
                if strings.HasPrefix(metric.Name, "kafka_") {
                        kafkaMetrics = append(kafkaMetrics, metric)
                } else if strings.HasPrefix(metric.Name, "jvm_") {
                        jvmMetrics = append(jvmMetrics, metric)
                } else if strings.HasPrefix(metric.Name, "process_") {
                        processMetrics = append(processMetrics, metric)
                }
        }
        
        metricsData.WriteString("Kafka Broker Metrics Analysis Request\n\n")
        metricsData.WriteString("KAFKA METRICS:\n")
        for _, m := range kafkaMetrics {
                metricsData.WriteString(fmt.Sprintf("%s{%s} = %s\n", m.Name, m.Labels, m.Value))
        }
        
        metricsData.WriteString("\nJVM METRICS:\n")
        for _, m := range jvmMetrics {
                metricsData.WriteString(fmt.Sprintf("%s{%s} = %s\n", m.Name, m.Labels, m.Value))
        }
        
        metricsData.WriteString("\nPROCESS METRICS:\n")
        for _, m := range processMetrics {
                metricsData.WriteString(fmt.Sprintf("%s{%s} = %s\n", m.Name, m.Labels, m.Value))
        }
        
        return fmt.Sprintf(`You are a Kafka expert analyzing broker metrics. Based on the following Prometheus metrics data, provide a detailed analysis.

%s

Please analyze these metrics and provide:

1. **ISSUES IDENTIFIED**: List any performance issues, bottlenecks, or concerning metric values
2. **ROOT CAUSE ANALYSIS**: Explain what might be causing these issues
3. **RECOMMENDATIONS**: Provide specific, actionable recommendations to address the issues
4. **SUMMARY**: Brief overall health assessment

Focus on:
- Memory usage and garbage collection patterns
- Disk I/O and log segment management
- Network throughput and request latencies
- Consumer lag and replication issues
- Resource utilization trends

Respond in JSON format:
{
  "issues": ["issue1", "issue2"],
  "root_causes": ["cause1", "cause2"],
  "recommendations": ["rec1", "rec2"],
  "summary": "overall assessment"
}`, metricsData.String())
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
                                "role":    "user",
                                "content": prompt,
                        },
                },
                "max_tokens":   2000,
                "temperature":  0.3,
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
        payload := map[string]interface{}{
                "model":      c.Model,
                "max_tokens": 2000,
                "messages": []map[string]string{
                        {
                                "role":    "user",
                                "content": prompt,
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
        payload := map[string]interface{}{
                "contents": []map[string]interface{}{
                        {
                                "parts": []map[string]string{
                                        {"text": prompt},
                                },
                        },
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