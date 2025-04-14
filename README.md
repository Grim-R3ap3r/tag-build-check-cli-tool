# 🔍 GitHub CI Tag Status Checker

A simple CLI tool written in Go to check the GitHub Actions workflow status for a specific tag on a repository.

---

## 📦 Features

- ✅ Check latest or specific tag workflow runs
- 🔁 Watch mode to keep checking until workflows complete
- 💬 Pretty terminal output with color and symbols
- 💨 Lightweight and fast

---

## ⚙️ Installation

### 1. Clone this repository

```bash
git clone https://github.com/yourusername/ci-tag-status.git
cd ci-tag-status
go build -o ts main.go
sudo mv ts /usr/local/bin/
export GITHUB_TOKEN=your_personal_access_token
```

### 2. Basic Commands

```bash
ts -r <repo-name> -t <tag-name>
ts -r <repo-name> -t release-v1 --watch
# Watch mode with custom interval
ts -r <repo-name> -t <tag-name> -w -i 5

