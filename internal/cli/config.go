package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/objectisnotdefined/consensus-agent/ca/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	guiFlag bool
	port    int
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage consensus-agent configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if guiFlag {
			startGUIPortal()
			return
		}
		// Default: just print where the config is
		fmt.Printf("Config file: %s\n", getConfigPath())
		fmt.Println("Use 'ca config --gui' to open the visual configuration portal.")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().BoolVarP(&guiFlag, "gui", "g", false, "Open the visual configuration portal")
	configCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for the configuration portal")
}

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "ca", "ca.yaml")
	return path
}

func startGUIPortal() {
	addr := fmt.Sprintf("localhost:%d", port)
	fmt.Printf("Starting Visual Config Portal at http://%s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")
	
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/api/config", handleConfigAPI)

	go openBrowser("http://" + addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlTemplate))
}

func handleConfigAPI(w http.ResponseWriter, r *http.Request) {
	path := getConfigPath()

	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(path)
		var cfg config.Config
		if err == nil {
			_ = yaml.Unmarshal(data, &cfg)
		} else {
			cfg = *config.Default()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)

	case http.MethodPost:
		var cfg config.Config
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &cfg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Ensure directory exists
		os.MkdirAll(filepath.Dir(path), 0755)

		data, _ := yaml.Marshal(cfg)
		if err := os.WriteFile(path, data, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Consensus-Agent | Configuration</title>
    <script src="https://unpkg.com/vue@3/dist/vue.global.js"></script>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        body { background-color: #0f172a; color: #f8fafc; }
        .card { background-color: #1e293b; border: 1px solid #334155; }
        input { background-color: #0f172a; border: 1px solid #334155; color: #f8fafc; }
        input:focus { border-color: #38bdf8; outline: none; }
    </style>
</head>
<body class="p-8">
    <div id="app" class="max-w-4xl mx-auto">
        <header class="mb-8 flex justify-between items-center">
            <div>
                <h1 class="text-3xl font-bold text-sky-400">⚡ Consensus-Agent</h1>
                <p class="text-slate-400">Visual Configuration Portal</p>
            </div>
            <button @click="saveConfig" class="bg-sky-500 hover:bg-sky-600 px-6 py-2 rounded-lg font-semibold transition">
                Save Changes
            </button>
        </header>

        <div v-if="loading" class="text-center py-20">Loading configuration...</div>
        
        <div v-else class="space-y-6">
            <!-- Consensus Section -->
            <section class="card p-6 rounded-xl">
                <h2 class="text-xl font-semibold mb-4 border-b border-slate-700 pb-2">Protocol Settings</h2>
                <div class="grid grid-cols-2 gap-6">
                    <div>
                        <label class="block text-sm text-slate-400 mb-1">Consensus Threshold</label>
                        <input type="number" step="0.01" v-model.number="cfg.consensus.threshold" class="w-full px-3 py-2 rounded">
                    </div>
                    <div>
                        <label class="block text-sm text-slate-400 mb-1">Max Debate Rounds</label>
                        <input type="number" v-model.number="cfg.consensus.max_rounds" class="w-full px-3 py-2 rounded">
                    </div>
                </div>
            </section>

            <!-- Cost Section -->
            <section class="card p-6 rounded-xl">
                <h2 class="text-xl font-semibold mb-4 border-b border-slate-700 pb-2">Budget Control</h2>
                <div>
                    <label class="block text-sm text-slate-400 mb-1">Token Budget per Session</label>
                    <input type="number" v-model.number="cfg.cost.token_budget" class="w-full px-3 py-2 rounded">
                </div>
            </section>

            <!-- Models Section -->
            <section class="card p-6 rounded-xl">
                <div class="flex justify-between items-center mb-4 border-b border-slate-700 pb-2">
                    <h2 class="text-xl font-semibold">LLM Models</h2>
                    <button @click="addModel" class="text-sky-400 hover:text-sky-300 text-sm font-medium">+ Add Model</button>
                </div>
                <div class="space-y-4">
                    <div v-for="(model, index) in cfg.models" :key="index" class="p-4 bg-slate-900/50 rounded-lg relative border border-slate-800">
                        <button @click="removeModel(index)" class="absolute top-2 right-2 text-slate-500 hover:text-red-400">✕</button>
                        <div class="grid grid-cols-4 gap-4 mb-4">
                            <div>
                                <label class="block text-xs text-slate-500 mb-1">Name</label>
                                <input v-model="model.name" placeholder="e.g. gpt-4o" class="w-full px-3 py-2 rounded text-sm">
                            </div>
                            <div>
                                <label class="block text-xs text-slate-500 mb-1">Provider</label>
                                <select v-model="model.provider" @change="updateDefaultEndpoint(index)" class="w-full px-3 py-2 rounded text-sm bg-slate-900 border border-slate-700">
                                    <option value="openai">OpenAI</option>
                                    <option value="anthropic">Anthropic</option>
                                    <option value="google">Google Gemini</option>
                                </select>
                            </div>
                            <div>
                                <label class="block text-xs text-slate-500 mb-1">Endpoint URL</label>
                                <input v-model="model.endpoint_url" placeholder="https://..." class="w-full px-3 py-2 rounded text-sm">
                            </div>
                            <div>
                                <label class="block text-xs text-slate-500 mb-1">API Key</label>
                                <input type="password" v-model="model.api_key" class="w-full px-3 py-2 rounded text-sm">
                            </div>
                        </div>

                        <!-- Role Selection -->
                        <div class="flex items-center gap-4 border-t border-slate-800 pt-3">
                            <span class="text-xs font-semibold text-slate-400">Enable for Roles:</span>
                            <div v-for="role in ['Navigator', 'Architect', 'Executor', 'Validator']" :key="role" class="flex items-center gap-1">
                                <input type="checkbox" :id="'role-'+index+'-'+role" :value="role" v-model="model.roles" class="w-3 h-3 rounded border-slate-700 bg-slate-900 text-sky-500 focus:ring-sky-500">
                                <label :for="'role-'+index+'-'+role" class="text-xs text-slate-300">{{ role }}</label>
                            </div>
                        </div>
                    </div>
                    <div v-if="cfg.models.length === 0" class="text-center py-4 text-slate-500 italic">
                        No models configured. Add one to start using agents.
                    </div>
                </div>
            </section>
        </div>

        <!-- Success Toast -->
        <div v-if="showToast" class="fixed bottom-8 right-8 bg-emerald-500 text-white px-6 py-3 rounded-lg shadow-lg transition-opacity">
            Configuration saved successfully!
        </div>
    </div>

    <script>
        const { createApp } = Vue
        const DEFAULT_ENDPOINTS = {
            openai: 'https://api.openai.com/v1',
            anthropic: 'https://api.anthropic.com',
            google: 'https://generativelanguage.googleapis.com'
        }

        createApp({
            data() {
                return {
                    cfg: { consensus: {}, cost: {}, models: [] },
                    loading: true,
                    showToast: false
                }
            },
            mounted() {
                this.fetchConfig()
            },
            methods: {
                async fetchConfig() {
                    try {
                        const res = await fetch('/api/config')
                        if (!res.ok) throw new Error('Failed to fetch config')
                        this.cfg = await res.json()
                        if (!this.cfg.models) this.cfg.models = []
                    } catch (err) {
                        console.error(err)
                        alert('Error loading configuration. Please check backend logs.')
                    } finally {
                        this.loading = false
                    }
                },
                async saveConfig() {
                    const res = await fetch('/api/config', {
                        method: 'POST',
                        body: JSON.stringify(this.cfg)
                    })
                    if (res.ok) {
                        this.showToast = true
                        setTimeout(() => this.showToast = false, 3000)
                    }
                },
                addModel() {
                    this.cfg.models.push({ 
                        name: '', 
                        provider: 'openai', 
                        endpoint_url: DEFAULT_ENDPOINTS.openai, 
                        api_key: '',
                        roles: ['Navigator', 'Architect', 'Executor', 'Validator']
                    })
                },
                removeModel(index) {
                    this.cfg.models.splice(index, 1)
                },
                updateDefaultEndpoint(index) {
                    const model = this.cfg.models[index]
                    if (DEFAULT_ENDPOINTS[model.provider]) {
                        model.endpoint_url = DEFAULT_ENDPOINTS[model.provider]
                    }
                }
            }
        }).mount('#app')
    </script>
</body>
</html>
`
