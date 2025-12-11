# Flight Coding Projects Guide
Singapore to Frankfurt

## Table of Contents
1. [Project 1: Universal LLM CLI](#project-1-universal-llm-cli)
2. [Project 2: Advanced Data Structures](#project-2-advanced-data-structures)
3. [Pre-Flight Setup Checklist](#pre-flight-setup-checklist)

---

## Project 1: Universal LLM CLI

### Overview
Build a Go CLI that connects to multiple local LLM servers (Ollama, MLX-LM) with auto-detection, auto-start capabilities, and basic agentic features.

### Core Features
- Connect to Ollama and MLX-LM servers
- Auto-detect which servers are running
- Auto-start servers if not running
- Discover available models in `~/LLMs`
- Stream chat responses with colored output
- Save/load conversation sessions
- (Optional) Basic agentic tools: file reading, command execution, code search

### Architecture

#### Project Structure
```
llm-cli/
├── main.go
├── cmd/
│   ├── root.go           # Root command setup
│   ├── chat.go           # Chat command
│   ├── list.go           # List servers command
│   └── models.go         # List models command
├── pkg/
│   ├── client/
│   │   ├── client.go     # Common LLMClient interface
│   │   ├── ollama.go     # Ollama implementation
│   │   └── mlx.go        # MLX implementation
│   ├── server/
│   │   ├── manager.go    # ServerManager interface
│   │   ├── ollama.go     # OllamaManager
│   │   └── mlx.go        # MLXManager
│   ├── discovery/
│   │   └── discovery.go  # Detect running servers
│   ├── models/
│   │   └── models.go     # Scan ~/LLMs directory
│   └── session/
│       └── session.go    # Conversation history
└── go.mod
```

#### Key Interfaces

```go
// LLMClient - Common interface for all LLM servers
type LLMClient interface {
    Chat(ctx context.Context, messages []Message, opts ChatOptions) (<-chan string, error)
    ListModels() ([]Model, error)
    IsAvailable() bool
}

// ServerManager - Manages server lifecycle
type ServerManager interface {
    IsAvailable() bool  // Check if server software is installed
    IsRunning() bool    // Check if server is currently running
    Start() error       // Start the server
    Stop() error        // Stop the server
    GetPort() int       // Get the port server runs on
}

// Message - Chat message structure
type Message struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"`
}

// ChatOptions - Configuration for chat requests
type ChatOptions struct {
    Temperature float64
    MaxTokens   int
    Stream      bool
}
```

### Implementation Phases

#### Phase 1: Core Infrastructure (1-2 hours)
**Goal:** Set up Cobra CLI with basic commands and server detection

```go
// Commands to implement:
// - llm-cli chat "your message"
// - llm-cli list-servers
// - llm-cli list-models
// - llm-cli chat --server=mlx "your message"

// Key setup:
cobra.OnInitialize(initConfig)
rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
rootCmd.PersistentFlags().String("server", "", "server type (ollama/mlx, auto-detect if empty)")
```

**Server Detection Functions:**
```go
// In pkg/discovery/discovery.go

func isBinaryAvailable(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

func isPythonModuleAvailable(module string) bool {
    cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s", module))
    return cmd.Run() == nil
}

func detectAvailableServers() []string {
    var available []string
    if isBinaryAvailable("ollama") {
        available = append(available, "ollama")
    }
    if isPythonModuleAvailable("mlx_lm") {
        available = append(available, "mlx")
    }
    return available
}
```

#### Phase 2: Server Detection and Management (1 hour)
**Goal:** Detect available servers and auto-start them

**Step 1: Detect Available Servers**
```go
// isBinaryAvailable checks if a binary exists in PATH
func isBinaryAvailable(name string) bool {
    _, err := exec.LookPath(name)
    return err == nil
}

// isPythonModuleAvailable checks if a Python module can be imported
func isPythonModuleAvailable(module string) bool {
    cmd := exec.Command("python3", "-c", fmt.Sprintf("import %s", module))
    err := cmd.Run()
    return err == nil
}

// detectAvailableServers detects which LLM servers are available on the system
func detectAvailableServers() []string {
    var available []string

    // Check for Ollama
    if isBinaryAvailable("ollama") {
        available = append(available, "ollama")
    }

    // Check for MLX-LM (Python module)
    if isPythonModuleAvailable("mlx_lm") {
        available = append(available, "mlx")
    }

    return available
}
```

**Step 2: Ollama Manager**
```go
type OllamaManager struct {
    port int
    cmd  *exec.Cmd
}

func NewOllamaManager() *OllamaManager {
    return &OllamaManager{port: 11434}
}

func (o *OllamaManager) IsAvailable() bool {
    return isBinaryAvailable("ollama")
}

func (o *OllamaManager) Start() error {
    if !o.IsAvailable() {
        return fmt.Errorf("ollama not installed")
    }
    
    // Start: ollama serve
    cmd := exec.Command("ollama", "serve")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start ollama: %w", err)
    }
    
    o.cmd = cmd
    
    // Wait for health check with timeout (30s)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            o.Stop()
            return fmt.Errorf("timeout waiting for ollama to start")
        case <-ticker.C:
            if o.IsRunning() {
                return nil
            }
        }
    }
}

func (o *OllamaManager) IsRunning() bool {
    resp, err := http.Get("http://localhost:11434/api/tags")
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}

func (o *OllamaManager) Stop() error {
    if o.cmd != nil && o.cmd.Process != nil {
        return o.cmd.Process.Kill()
    }
    return nil
}
```

**Step 3: MLX Manager**
```go
type MLXManager struct {
    port  int
    model string
    cmd   *exec.Cmd
}

func NewMLXManager(model string) *MLXManager {
    return &MLXManager{
        port:  8080,
        model: model,
    }
}

func (m *MLXManager) IsAvailable() bool {
    return isPythonModuleAvailable("mlx_lm")
}

func (m *MLXManager) Start() error {
    if !m.IsAvailable() {
        return fmt.Errorf("mlx_lm not installed")
    }
    
    // Start: python -m mlx_lm.server --port 8080 --model <model>
    cmd := exec.Command(
        "python3", "-m", "mlx_lm.server",
        "--port", fmt.Sprintf("%d", m.port),
        "--model", m.model,
    )
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start mlx server: %w", err)
    }
    
    m.cmd = cmd
    
    // Wait for health check with timeout (60s for model loading)
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            m.Stop()
            return fmt.Errorf("timeout waiting for mlx server to start")
        case <-ticker.C:
            if m.IsRunning() {
                return nil
            }
        }
    }
}

func (m *MLXManager) IsRunning() bool {
    resp, err := http.Get(fmt.Sprintf("http://localhost:%d/v1/models", m.port))
    if err != nil {
        return false
    }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}

func (m *MLXManager) Stop() error {
    if m.cmd != nil && m.cmd.Process != nil {
        return m.cmd.Process.Kill()
    }
    return nil
}
```

**Step 4: Usage in CLI**
```go
func runChat(cmd *cobra.Command, args []string) error {
    // Detect available servers
    available := detectAvailableServers()
    if len(available) == 0 {
        return fmt.Errorf("no LLM servers installed (ollama or mlx_lm required)")
    }
    
    // Get preferred server from flag or use first available
    serverType := viper.GetString("server")
    if serverType == "" {
        serverType = available[0]
        fmt.Printf("Using %s (detected)\n", serverType)
    }
    
    // Create appropriate manager
    var manager ServerManager
    switch serverType {
    case "ollama":
        manager = NewOllamaManager()
        if !manager.IsAvailable() {
            return fmt.Errorf("ollama not installed")
        }
    case "mlx":
        // Get model from config or use default
        model := viper.GetString("mlx-model")
        if model == "" {
            model = "mlx-community/Llama-3.2-3B-Instruct-4bit"
        }
        manager = NewMLXManager(model)
        if !manager.IsAvailable() {
            return fmt.Errorf("mlx_lm not installed")
        }
    default:
        return fmt.Errorf("unknown server type: %s", serverType)
    }
    
    // Check if running, start if not
    if !manager.IsRunning() {
        fmt.Printf("Starting %s...\n", serverType)
        if err := manager.Start(); err != nil {
            return fmt.Errorf("failed to start server: %w", err)
        }
        fmt.Println("✓ Server started successfully")
    } else {
        fmt.Println("✓ Server already running")
    }
    
    // Now proceed with chat...
    return nil
}
```

#### Phase 3: Model Discovery (30-45 min)
**Goal:** Scan for available models

```go
func findMLXModels() ([]string, error) {
    homeDir, _ := os.UserHomeDir()
    locations := []string{
        filepath.Join(homeDir, "LLMs", "mlx-models"),
        filepath.Join(homeDir, ".cache", "huggingface", "hub"),
    }
    
    var models []string
    for _, loc := range locations {
        entries, _ := os.ReadDir(loc)
        for _, entry := range entries {
            if entry.IsDir() && hasMLXFiles(filepath.Join(loc, entry.Name())) {
                models = append(models, entry.Name())
            }
        }
    }
    return models, nil
}

func hasMLXFiles(path string) bool {
    // Check for config.json and weights.npz
    required := []string{"config.json", "weights.npz"}
    for _, file := range required {
        if _, err := os.Stat(filepath.Join(path, file)); os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func findOllamaModels() ([]string, error) {
    // Call Ollama API: GET http://localhost:11434/api/tags
    resp, err := http.Get("http://localhost:11434/api/tags")
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Models []struct {
            Name string `json:"name"`
        } `json:"models"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    var models []string
    for _, m := range result.Models {
        models = append(models, m.Name)
    }
    return models, nil
}
```

#### Phase 4: Streaming Chat (1 hour)
**Goal:** Implement SSE parsing and colored output

```go
func streamChat(url string, messages []Message) error {
    payload := map[string]interface{}{
        "model":    "llama3.2",
        "messages": messages,
        "stream":   true,
    }
    
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadBytes('\n')
        if err != nil {
            break
        }
        
        // Parse SSE: "data: {...}"
        if bytes.HasPrefix(line, []byte("data: ")) {
            data := bytes.TrimPrefix(line, []byte("data: "))
            
            if bytes.Equal(data, []byte("[DONE]\n")) {
                break
            }
            
            var chunk struct {
                Choices []struct {
                    Delta struct {
                        Content string `json:"content"`
                    } `json:"delta"`
                } `json:"choices"`
            }
            
            json.Unmarshal(data, &chunk)
            if len(chunk.Choices) > 0 {
                fmt.Print(chunk.Choices[0].Delta.Content)
            }
        }
    }
    fmt.Println()
    return nil
}
```

**Colored Output with fatih/color:**
```go
import "github.com/fatih/color"

userColor := color.New(color.FgCyan).SprintFunc()
assistantColor := color.New(color.FgGreen).SprintFunc()

fmt.Printf("%s: %s\n", userColor("You"), message)
fmt.Printf("%s: ", assistantColor("Assistant"))
// Stream response here
```

#### Phase 5: Basic Agentic Tools (1-2 hours, optional)
**Goal:** Add function calling support

```go
// Tool definitions
type Tool struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
    Handler     func(args map[string]interface{}) (string, error)
}

var tools = []Tool{
    {
        Name:        "read_file",
        Description: "Read contents of a file",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "path": map[string]string{
                    "type":        "string",
                    "description": "Path to file",
                },
            },
        },
        Handler: func(args map[string]interface{}) (string, error) {
            path := args["path"].(string)
            content, err := os.ReadFile(path)
            return string(content), err
        },
    },
    {
        Name:        "list_files",
        Description: "List files in directory",
        Handler: func(args map[string]interface{}) (string, error) {
            dir := args["path"].(string)
            entries, err := os.ReadDir(dir)
            if err != nil {
                return "", err
            }
            var files []string
            for _, e := range entries {
                files = append(files, e.Name())
            }
            return strings.Join(files, "\n"), nil
        },
    },
    {
        Name:        "execute_command",
        Description: "Execute shell command (use with caution)",
        Handler: func(args map[string]interface{}) (string, error) {
            cmd := args["command"].(string)
            
            // Safety check - whitelist safe commands
            if !isSafeCommand(cmd) {
                return "", fmt.Errorf("command not allowed")
            }
            
            output, err := exec.Command("sh", "-c", cmd).CombinedOutput()
            return string(output), err
        },
    },
}

// Parse tool calls from model response
func parseToolCalls(response string) []ToolCall {
    // Look for patterns like: <tool_call>{"name": "read_file", "args": {...}}</tool_call>
    // Or use JSON mode if model supports it
    return nil
}

// Execute tool and add result to conversation
func executeTools(toolCalls []ToolCall) []Message {
    var results []Message
    for _, tc := range toolCalls {
        for _, tool := range tools {
            if tool.Name == tc.Name {
                result, err := tool.Handler(tc.Arguments)
                if err != nil {
                    result = fmt.Sprintf("Error: %v", err)
                }
                results = append(results, Message{
                    Role:    "tool",
                    Content: result,
                })
            }
        }
    }
    return results
}
```

### API Endpoints Reference

#### Ollama API (Port 11434)
```
GET  /api/tags                    # List models
POST /api/generate                # Generate completion
POST /api/chat                    # Chat completion
GET  /api/show                    # Show model info

Chat Request:
POST /api/chat
{
  "model": "llama3.2",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": true
}

Response (SSE):
data: {"message": {"content": "Hi"}}
data: {"message": {"content": " there"}}
data: {"done": true}
```

#### MLX-LM API (Port 8080)
```
GET  /v1/models                   # List models
POST /v1/chat/completions        # Chat completion

Chat Request:
POST /v1/chat/completions
{
  "model": "mlx-community/Llama-3.2-3B-Instruct-4bit",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "stream": true,
  "max_tokens": 512,
  "temperature": 0.7
}

Response (SSE):
data: {"choices": [{"delta": {"content": "Hi"}}]}
data: {"choices": [{"delta": {"content": " there"}}]}
data: [DONE]
```

### Testing Strategy
1. Start with Ollama (simpler, more stable)
2. Test each phase independently:
   - Server start/stop
   - Model listing
   - Single message chat
   - Streaming chat
   - Conversation history
3. Add MLX support once Ollama works
4. Test graceful shutdown (Ctrl+C handling)

### Time Estimates
- Phase 1: 1-2 hours
- Phase 2: 1 hour
- Phase 3: 30-45 min
- Phase 4: 1 hour
- Phase 5: 1-2 hours (optional)
- **Total: 3.5-6.5 hours**

---

## Project 2: Advanced Data Structures

### Overview
Implement B-Trees and Fibonacci Heaps in Go, Rust, and C++ to practice advanced data structure concepts and compare language implementations.

### B-Tree Implementation

#### Specifications
- **Order (M):** Use M=4 or M=5 for simplicity (2-3-4 tree)
- **Operations:** insert, search, delete, range_query
- **Node Structure:**
  - Array of keys
  - Array of child pointers
  - isLeaf flag
  - Current key count

#### Implementation Phases

**Phase 1: Basic Structure (30 min)**
```go
type BTreeNode struct {
    keys     []int
    children []*BTreeNode
    isLeaf   bool
    n        int  // current number of keys
}

type BTree struct {
    root  *BTreeNode
    order int      // maximum number of children
    t     int      // minimum degree (t = order/2)
}

func NewBTree(order int) *BTree {
    return &BTree{
        root:  &BTreeNode{isLeaf: true},
        order: order,
        t:     order / 2,
    }
}
```

**Phase 2: Search (30 min)**
```go
func (b *BTree) Search(key int) (*BTreeNode, int, bool) {
    return b.root.search(key)
}

func (n *BTreeNode) search(key int) (*BTreeNode, int, bool) {
    i := 0
    // Find first key >= target
    for i < n.n && key > n.keys[i] {
        i++
    }
    
    // Key found
    if i < n.n && key == n.keys[i] {
        return n, i, true
    }
    
    // Key not in tree
    if n.isLeaf {
        return nil, 0, false
    }
    
    // Recurse to appropriate child
    return n.children[i].search(key)
}
```

**Phase 3: Insert (1-1.5 hours)**
```go
func (b *BTree) Insert(key int) {
    root := b.root
    
    // If root is full, split it
    if root.n == b.order-1 {
        newRoot := &BTreeNode{isLeaf: false}
        newRoot.children = append(newRoot.children, root)
        b.splitChild(newRoot, 0)
        b.root = newRoot
    }
    
    b.insertNonFull(b.root, key)
}

func (b *BTree) insertNonFull(node *BTreeNode, key int) {
    i := node.n - 1
    
    if node.isLeaf {
        // Insert key in sorted order
        node.keys = append(node.keys, 0)
        for i >= 0 && key < node.keys[i] {
            node.keys[i+1] = node.keys[i]
            i--
        }
        node.keys[i+1] = key
        node.n++
    } else {
        // Find child to recurse to
        for i >= 0 && key < node.keys[i] {
            i--
        }
        i++
        
        // Split child if full
        if node.children[i].n == b.order-1 {
            b.splitChild(node, i)
            if key > node.keys[i] {
                i++
            }
        }
        b.insertNonFull(node.children[i], key)
    }
}

func (b *BTree) splitChild(parent *BTreeNode, index int) {
    fullChild := parent.children[index]
    newChild := &BTreeNode{isLeaf: fullChild.isLeaf}
    
    t := b.t
    midKey := fullChild.keys[t-1]
    
    // Move upper half of keys to new child
    newChild.keys = append([]int{}, fullChild.keys[t:]...)
    newChild.n = len(newChild.keys)
    fullChild.keys = fullChild.keys[:t-1]
    fullChild.n = len(fullChild.keys)
    
    // Move upper half of children if not leaf
    if !fullChild.isLeaf {
        newChild.children = append([]*BTreeNode{}, fullChild.children[t:]...)
        fullChild.children = fullChild.children[:t]
    }
    
    // Insert midKey into parent
    parent.keys = append(parent.keys[:index], append([]int{midKey}, parent.keys[index:]...)...)
    parent.children = append(parent.children[:index+1], append([]*BTreeNode{newChild}, parent.children[index+1:]...)...)
    parent.n++
}
```

**Phase 4: Delete (2-3 hours - most complex)**
```go
// Three cases:
// 1. Key in leaf node - just remove
// 2. Key in internal node - replace with predecessor/successor
// 3. Child has minimum keys - borrow from sibling or merge

func (b *BTree) Delete(key int) {
    b.root.delete(key, b.t)
    
    // If root is empty after deletion, make its only child the new root
    if b.root.n == 0 && !b.root.isLeaf {
        b.root = b.root.children[0]
    }
}

// Implement: delete, borrowFromPrev, borrowFromNext, merge, getPredecessor, getSuccessor
```

**Phase 5: Range Query (30 min)**
```go
func (b *BTree) RangeQuery(low, high int) []int {
    var result []int
    b.root.rangeQuery(low, high, &result)
    return result
}

func (n *BTreeNode) rangeQuery(low, high int, result *[]int) {
    i := 0
    for i < n.n {
        if !n.isLeaf {
            n.children[i].rangeQuery(low, high, result)
        }
        
        if n.keys[i] >= low && n.keys[i] <= high {
            *result = append(*result, n.keys[i])
        }
        i++
    }
    
    if !n.isLeaf {
        n.children[i].rangeQuery(low, high, result)
    }
}
```

#### Testing
```go
func TestBTree() {
    tree := NewBTree(5)
    
    // Insert test
    keys := []int{10, 20, 5, 6, 12, 30, 7, 17}
    for _, k := range keys {
        tree.Insert(k)
    }
    
    // Search test
    _, _, found := tree.Search(6)
    fmt.Printf("Found 6: %v\n", found)
    
    // Range query test
    result := tree.RangeQuery(5, 15)
    fmt.Printf("Range [5,15]: %v\n", result)
    
    // Verify structure
    tree.Print()
}
```

### Fibonacci Heap Implementation

#### Specifications
- **Operations:**
  - Insert: O(1)
  - Find-min: O(1)
  - Extract-min: O(log n) amortized
  - Decrease-key: O(1) amortized
  - Delete: O(log n) amortized
  - Merge: O(1)

#### Structure
```go
type FibNode struct {
    key      int
    degree   int
    marked   bool
    parent   *FibNode
    child    *FibNode
    left     *FibNode
    right    *FibNode
}

type FibHeap struct {
    min  *FibNode
    size int
}
```

#### Key Concepts
1. **Circular Doubly-Linked List** for root nodes
2. **Lazy Consolidation** - only consolidate on extract-min
3. **Cascading Cuts** - maintain heap structure efficiently
4. **Degree** - number of children
5. **Marked Nodes** - track nodes that lost children

#### Implementation Phases

**Phase 1: Basic Operations (1 hour)**
```go
func NewFibHeap() *FibHeap {
    return &FibHeap{}
}

func (h *FibHeap) Insert(key int) *FibNode {
    node := &FibNode{key: key}
    node.left = node
    node.right = node
    
    if h.min == nil {
        h.min = node
    } else {
        h.insertIntoRootList(node)
        if node.key < h.min.key {
            h.min = node
        }
    }
    h.size++
    return node
}

func (h *FibHeap) FindMin() int {
    if h.min == nil {
        panic("heap is empty")
    }
    return h.min.key
}

func (h *FibHeap) insertIntoRootList(node *FibNode) {
    node.right = h.min.right
    node.left = h.min
    h.min.right.left = node
    h.min.right = node
}
```

**Phase 2: Extract-Min with Consolidation (1.5 hours)**
```go
func (h *FibHeap) ExtractMin() int {
    if h.min == nil {
        panic("heap is empty")
    }
    
    minNode := h.min
    
    // Add all children to root list
    if minNode.child != nil {
        child := minNode.child
        for {
            next := child.right
            h.insertIntoRootList(child)
            child.parent = nil
            child = next
            if child == minNode.child {
                break
            }
        }
    }
    
    // Remove min from root list
    minNode.left.right = minNode.right
    minNode.right.left = minNode.left
    
    if minNode == minNode.right {
        h.min = nil
    } else {
        h.min = minNode.right
        h.consolidate()
    }
    
    h.size--
    return minNode.key
}

func (h *FibHeap) consolidate() {
    maxDegree := int(math.Log2(float64(h.size))) + 1
    degreeTable := make([]*FibNode, maxDegree+1)
    
    // Collect all root nodes
    roots := []*FibNode{}
    curr := h.min
    for {
        roots = append(roots, curr)
        curr = curr.right
        if curr == h.min {
            break
        }
    }
    
    // Consolidate
    for _, node := range roots {
        degree := node.degree
        for degreeTable[degree] != nil {
            other := degreeTable[degree]
            if node.key > other.key {
                node, other = other, node
            }
            h.link(other, node)
            degreeTable[degree] = nil
            degree++
        }
        degreeTable[degree] = node
    }
    
    // Rebuild root list and find new min
    h.min = nil
    for _, node := range degreeTable {
        if node != nil {
            if h.min == nil {
                h.min = node
                node.left = node
                node.right = node
            } else {
                h.insertIntoRootList(node)
                if node.key < h.min.key {
                    h.min = node
                }
            }
        }
    }
}

func (h *FibHeap) link(child, parent *FibNode) {
    // Remove child from root list
    child.left.right = child.right
    child.right.left = child.left
    
    // Make child a child of parent
    child.parent = parent
    if parent.child == nil {
        parent.child = child
        child.left = child
        child.right = child
    } else {
        child.right = parent.child.right
        child.left = parent.child
        parent.child.right.left = child
        parent.child.right = child
    }
    
    parent.degree++
    child.marked = false
}
```

**Phase 3: Decrease-Key with Cascading Cuts (1 hour)**
```go
func (h *FibHeap) DecreaseKey(node *FibNode, newKey int) {
    if newKey > node.key {
        panic("new key is greater than current key")
    }
    
    node.key = newKey
    parent := node.parent
    
    if parent != nil && node.key < parent.key {
        h.cut(node, parent)
        h.cascadingCut(parent)
    }
    
    if node.key < h.min.key {
        h.min = node
    }
}

func (h *FibHeap) cut(node, parent *FibNode) {
    // Remove node from parent's child list
    if node.right == node {
        parent.child = nil
    } else {
        node.left.right = node.right
        node.right.left = node.left
        if parent.child == node {
            parent.child = node.right
        }
    }
    parent.degree--
    
    // Add node to root list
    h.insertIntoRootList(node)
    node.parent = nil
    node.marked = false
}

func (h *FibHeap) cascadingCut(node *FibNode) {
    parent := node.parent
    if parent != nil {
        if !node.marked {
            node.marked = true
        } else {
            h.cut(node, parent)
            h.cascadingCut(parent)
        }
    }
}
```

**Phase 4: Delete (15 min)**
```go
func (h *FibHeap) Delete(node *FibNode) {
    h.DecreaseKey(node, math.MinInt64)
    h.ExtractMin()
}
```

**Phase 5: Merge (30 min)**
```go
func (h *FibHeap) Merge(other *FibHeap) {
    if other.min == nil {
        return
    }
    
    if h.min == nil {
        h.min = other.min
    } else {
        // Concatenate root lists
        h.min.right.left = other.min.left
        other.min.left.right = h.min.right
        h.min.right = other.min
        other.min.left = h.min
        
        if other.min.key < h.min.key {
            h.min = other.min
        }
    }
    
    h.size += other.size
}
```

#### Testing
```go
func TestFibHeap() {
    heap := NewFibHeap()
    
    // Insert test
    nodes := []*FibNode{}
    for _, k := range []int{10, 5, 20, 3, 15} {
        nodes = append(nodes, heap.Insert(k))
    }
    
    // Find-min test
    fmt.Printf("Min: %d\n", heap.FindMin())  // Should be 3
    
    // Extract-min test
    min := heap.ExtractMin()
    fmt.Printf("Extracted: %d\n", min)  // Should be 3
    fmt.Printf("New min: %d\n", heap.FindMin())  // Should be 5
    
    // Decrease-key test
    heap.DecreaseKey(nodes[2], 1)  // Decrease 20 to 1
    fmt.Printf("Min after decrease: %d\n", heap.FindMin())  // Should be 1
}
```

### Language-Specific Notes

#### Go
- Use pointers for nodes
- No need for manual memory management
- Use `math.MinInt64` for negative infinity
- Testing with `testing` package

#### Rust
```rust
use std::rc::Rc;
use std::cell::RefCell;

type NodeRef<T> = Rc<RefCell<Node<T>>>;

// You'll need to handle:
// - Circular references with Rc<RefCell<>>
// - Borrow checker issues
// - Option<> for nullable pointers
// - Manual implementation of comparison traits
```

**Key challenges:**
- Circular references require `Rc` and `Weak`
- Mutable references require `RefCell`
- More verbose but safer

#### C++
```cpp
template<typename T>
class BTreeNode {
    std::vector<T> keys;
    std::vector<BTreeNode*> children;
    bool isLeaf;
    int n;
};

// Remember:
// - Manual memory management (delete in destructor)
// - Use smart pointers (std::unique_ptr) for safety
// - Template for generic types
// - Operator overloading for comparisons
```

### Time Estimates per Language

**B-Tree:**
- Basic structure + search: 1 hour
- Insert with splits: 1.5 hours
- Delete: 2-3 hours
- Range query: 30 min
- Testing: 30 min
- **Total: 5-6.5 hours per language**

**Fibonacci Heap:**
- Basic ops (insert, find-min): 1 hour
- Extract-min + consolidation: 1.5 hours
- Decrease-key + cascading cuts: 1 hour
- Delete + merge: 45 min
- Testing: 30 min
- **Total: 4.5-5 hours per language**

### Testing Strategy
1. Start with simple cases (insert, search)
2. Test edge cases (empty tree, single element)
3. Randomized testing (insert 1000 random elements)
4. Compare with standard library implementations
5. Benchmark performance

---

## Pre-Flight Setup Checklist

### System Setup
```bash
# Verify Go installation
go version  # Should be 1.21+

# Verify Rust installation (optional)
rustc --version
cargo --version

# Verify C++ compiler (optional)
clang++ --version  # or g++ --version

# Verify Python for MLX
python3 --version
```

### Project 1: LLM CLI

#### Install Dependencies
```bash
# Ollama
brew install ollama

# Download models
ollama pull llama3.2
ollama pull llama3.2:1b  # Smaller, faster model

# MLX
pip install mlx-lm

# Download MLX models (choose one or more)
python -m mlx_lm.convert --hf-path mlx-community/Llama-3.2-3B-Instruct-4bit
python -m mlx_lm.convert --hf-path mlx-community/Mistral-7B-Instruct-v0.3-4bit

# Test MLX server
python -m mlx_lm.server --model mlx-community/Llama-3.2-3B-Instruct-4bit --port 8080
# Press Ctrl+C to stop

# Go dependencies
mkdir llm-cli && cd llm-cli
go mod init llm-cli
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
go get github.com/fatih/color@latest
go get github.com/briandowns/spinner@latest
go get github.com/chzyer/readline@latest

# Download dependencies offline
go mod download
go mod vendor  # Optional: vendor dependencies for offline use
```

#### Test Setup
```bash
# Test server detection
cat > test_detection.go << 'EOF'
package main

import "fmt"

func main() {
    servers := detectAvailableServers()
    fmt.Printf("Detected servers: %v\n", servers)
    
    if isBinaryAvailable("ollama") {
        fmt.Println("✓ Ollama is installed")
    } else {
        fmt.Println("✗ Ollama not found")
    }
    
    if isPythonModuleAvailable("mlx_lm") {
        fmt.Println("✓ MLX-LM is installed")
    } else {
        fmt.Println("✗ MLX-LM not found")
    }
}
EOF

go run test_detection.go

# Test Ollama
ollama serve &
curl http://localhost:11434/api/tags

# Test MLX (in another terminal)
python -m mlx_lm.server --model mlx-community/Llama-3.2-3B-Instruct-4bit --port 8080 &
curl http://localhost:8080/v1/models

# Stop servers
pkill ollama
pkill -f mlx_lm.server
```

### Project 2: Data Structures

#### Go Setup
```bash
mkdir data-structures && cd data-structures
go mod init data-structures

# Create basic test file
cat > btree_test.go << 'EOF'
package main

import "testing"

func TestBTree(t *testing.T) {
    tree := NewBTree(5)
    tree.Insert(10)
    _, _, found := tree.Search(10)
    if !found {
        t.Error("Failed to find inserted key")
    }
}
EOF

# Test it works
go test
```

#### Rust Setup (optional)
```bash
cargo new data-structures-rust
cd data-structures-rust

# Add dependencies to Cargo.toml
cat >> Cargo.toml << 'EOF'

[dev-dependencies]
criterion = "0.5"
rand = "0.8"

[[bench]]
name = "btree_bench"
harness = false
EOF

# Download dependencies
cargo build
```

#### C++ Setup (optional)
```bash
mkdir data-structures-cpp && cd data-structures-cpp

# Create CMakeLists.txt
cat > CMakeLists.txt << 'EOF'
cmake_minimum_required(VERSION 3.10)
project(DataStructures)

set(CMAKE_CXX_STANDARD 17)

add_executable(btree btree.cpp)
add_executable(fibheap fibheap.cpp)
EOF

# Test it works
mkdir build && cd build
cmake ..
```

### Offline Documentation

#### Save these locally
```bash
# Go documentation
go doc -all > go_stdlib.txt

# Download Go tour offline
git clone https://github.com/golang/tour

# Save useful pages
curl https://go.dev/ref/spec > go_spec.html

# Algorithm visualizations (save in browser)
# https://www.cs.usfca.edu/~galles/visualization/BTree.html
# https://visualgo.net/en/heap

# B-Tree tutorial
curl https://www.geeksforgeeks.org/introduction-of-b-tree-2/ > btree_tutorial.html
```

### File Organization
```
~/flight-projects/
├── llm-cli/
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── cmd/
│   ├── pkg/
│   └── vendor/  # If using vendored dependencies
├── data-structures/
│   ├── go.mod
│   ├── btree.go
│   ├── btree_test.go
│   ├── fibheap.go
│   └── fibheap_test.go
├── data-structures-rust/  # Optional
│   ├── Cargo.toml
│   └── src/
├── data-structures-cpp/  # Optional
│   ├── CMakeLists.txt
│   └── *.cpp
└── docs/
    ├── flight-projects-guide.md  # This file!
    ├── go_stdlib.txt
    └── btree_tutorial.html
```

### Final Checks

**Before Flight:**
- [ ] All Go dependencies downloaded (`go mod download`)
- [ ] Ollama installed and models pulled
- [ ] MLX installed and at least one model downloaded
- [ ] Test that servers start successfully
- [ ] This guide saved locally
- [ ] Laptop fully charged
- [ ] Power adapter packed

**During Flight:**
- [ ] Put laptop in airplane mode (or turn off WiFi)
- [ ] Start with Project 1 (more immediately rewarding)
- [ ] Take breaks between phases
- [ ] Switch to Project 2 if you need a change of pace

### Troubleshooting

**If Ollama won't start:**
- Check if port 11434 is already in use: `lsof -i :11434`
- Try: `ollama serve` manually in terminal
- Check logs: `tail -f ~/.ollama/logs/server.log`

**If MLX server won't start:**
- Check Python version: `python3 --version` (need 3.8+)
- Verify model exists: `ls ~/LLMs/`
- Try loading model manually in Python:
  ```python
  from mlx_lm import load
  model, tokenizer = load("path/to/model")
  ```

**If Go dependencies missing:**
- Use vendored dependencies: `go mod vendor`
- Check vendor directory: `ls vendor/`

**If data structure implementation stuck:**
- Start with smaller order (M=3) for B-Tree
- Draw diagrams on paper
- Test each operation independently
- Compare with visualizations

---

## Quick Reference Commands

### Project 1: LLM CLI
```bash
# Start servers manually
ollama serve
python -m mlx_lm.server --model mlx-community/Llama-3.2-3B-Instruct-4bit --port 8080

# Build CLI
go build -o llm-cli

# Run CLI
./llm-cli chat "Hello, how are you?"
./llm-cli --server=mlx chat "Explain quantum computing"
./llm-cli list-servers
./llm-cli list-models

# Stop servers
pkill ollama
pkill -f mlx_lm.server
```

### Project 2: Data Structures
```bash
# Go
go test -v                    # Run tests
go test -bench=.              # Run benchmarks
go test -cover                # Test coverage

# Rust
cargo test                    # Run tests
cargo bench                   # Run benchmarks

# C++
mkdir build && cd build
cmake ..
make
./btree                       # Run executable
```

---

## Notes and Tips

### General Advice
- **Start simple, iterate:** Get basic version working before adding features
- **Test frequently:** Don't write 100 lines before testing
- **Use paper:** Draw B-Tree structures, trace algorithms
- **Take breaks:** Long flight = plenty of time, no need to rush
- **Debug with prints:** Add logging to understand what's happening

### Project 1 Tips
- Ollama is more stable than MLX for initial development
- Start without agentic features, add them if time permits
- Colored output makes a huge UX difference
- Save conversation history to JSON files for persistence
- Use context.WithTimeout for all HTTP requests

### Project 2 Tips
- B-Tree delete is hardest operation - save for last or skip
- Draw every operation on paper first
- Fibonacci Heap is easier than it looks
- Start with Go (fastest to prototype)
- Use existing stdlib implementations to verify correctness

### Debugging Strategies
- Binary search debugging: comment out half the code
- Add assert statements liberally
- Print tree/heap structure after each operation
- Use a debugger (dlv for Go, lldb for C++, gdb)
- Simplify input: test with 3 elements before 100

### What to Do If Stuck
1. Take a 10-minute break
2. Explain the problem out loud (rubber duck debugging)
3. Simplify: solve a smaller version of the problem
4. Skip it and come back later
5. Switch to the other project

---

## Success Metrics

### Project 1 (Minimum Viable Product)
- [ ] CLI can start Ollama automatically
- [ ] Can send a message and get streaming response
- [ ] Colored output for user/assistant messages
- [ ] List available models
- [ ] Basic error handling

### Project 1 (Full Features)
- [ ] Support both Ollama and MLX
- [ ] Auto-detect and start servers
- [ ] Save/load conversation history
- [ ] At least 2 working agentic tools
- [ ] Graceful shutdown on Ctrl+C

### Project 2 (B-Tree)
- [ ] Insert works correctly
- [ ] Search works correctly
- [ ] Range query works correctly
- [ ] Passes randomized tests with 100+ elements
- [ ] (Optional) Delete works correctly

### Project 2 (Fibonacci Heap)
- [ ] Insert O(1)
- [ ] Extract-min works with consolidation
- [ ] Decrease-key with cascading cuts
- [ ] Passes tests comparing to std::priority_queue
- [ ] (Optional) Implement in 2+ languages

---

## Post-Flight

### Code Review Checklist
- [ ] Add comments for complex logic
- [ ] Write README with usage examples
- [ ] Add more comprehensive tests
- [ ] Benchmark performance
- [ ] Refactor duplicated code
- [ ] Add CI/CD if publishing

### Potential Extensions

**LLM CLI:**
- WebSocket support for real-time streaming
- Support for more servers (vLLM, LM Studio)
- Plugin system for custom tools
- Voice input/output
- Multi-modal support (images)
- Docker containerization

**Data Structures:**
- Concurrent B-Tree (lock-free algorithms)
- Persistent B-Tree (disk-based)
- Red-Black tree comparison
- Benchmark against stdlib
- Visualization tool
- Generic implementations (templates/traits)

---

## Resources

### Documentation (save offline)
- Go: https://go.dev/doc/
- Rust: https://doc.rust-lang.org/book/
- C++: https://en.cppreference.com/

### Algorithms
- CLRS: Introduction to Algorithms (Chapter 18: B-Trees, Chapter 19: Fibonacci Heaps)
- Visualizations: https://visualgo.net/

### APIs
- Ollama: https://github.com/ollama/ollama/blob/main/docs/api.md
- MLX-LM: https://github.com/ml-explore/mlx-examples/tree/main/llms

---

## Final Thoughts

You have two solid projects with clear goals and phases. Start with what excites you most - the LLM CLI is more immediately rewarding and useful, while the data structures are great for deep algorithmic thinking.

The flight is ~13 hours, so you have plenty of time to make significant progress on both. Don't stress about completing everything - the goal is to learn and have fun coding at 35,000 feet!

Good luck and enjoy your flight! 🚀
