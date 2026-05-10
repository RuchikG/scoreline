<div align="center">
  <img src="assets/scoreline-logo.png" alt="Scoreline demo" width="150">
  <h1>Scoreline</h1>
</div>

<div align="center">

[![GitHub Stars](https://img.shields.io/github/stars/RuchikG/scoreline?style=social)](https://github.com/RuchikG/scoreline)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/RuchikG/scoreline)](https://goreportcard.com/report/github.com/RuchikG/scoreline)
[![GitHub Release](https://img.shields.io/github/v/release/RuchikG/scoreline)](https://github.com/RuchikG/scoreline/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/RuchikG/scoreline/build.yml)](https://github.com/RuchikG/scoreline/actions/workflows/build.yml)

[![GitHub Downloads](https://img.shields.io/github/downloads/RuchikG/scoreline/total)](https://github.com/RuchikG/scoreline/releases)
![macOS](https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white)
![Linux](https://img.shields.io/badge/Linux-FCC624?logo=linux&logoColor=black)
![Windows](https://img.shields.io/badge/Windows-0078D6?logo=windows&logoColor=white)

A minimalist terminal user interface (TUI) for following soccer and cricket scores. Get live updates, match details, and sport-specific settings directly in your terminal.

Scoreline was created for those moments when you can't stream or watch matches live. It gives you a handy, non-intrusive, and minimalist way to keep up with your favourite teams and competitions.

*Perfect for developers and terminal enthusiasts who want match updates without leaving their workflow.*
</div>

> [!NOTE]
> If you enjoy Scoreline, give it a star and share it with your friends. That helps others find it and keeps the project going!

<div align="center">
  <img src="assets/golazo-demo-v0.18.0.gif" alt="Scoreline demo" width="800">
</div>

<div align="center">

**Quick Install:** [Install script](#install-script) · Homebrew packaging coming soon

</div>

## Features

- **Sport Selector**: Start in your last selected sport, with Soccer and Cricket routes separated.
- **Soccer Live Match Tracking**: Timeline and real-time updates for goals, cards, and substitutions with automatic polling.
- **Soccer Finished Matches**: View results from today, last 3 days, or last 5 days.
- **Soccer Match Statistics & Details**: Possession, shots, passes, standings, formations with player ratings, and more in focused dialogs.
- **Cricket Live Dashboard**: Cricket-native match list and scorecard-oriented detail panel, with mock data and CricketData current-matches foundation.
- **Cricket Archives**: Browse locally cached Cricsheet completed matches without spending CricketData API credits.
- **Cricket Settings**: Configure formats, refresh intervals, archive window, and API key environment status inside the Cricket menu.
- **65+ Soccer Leagues**: Organized by region (Europe, Americas, Global) with tab navigation in Soccer settings.

## Installation & Update

**Self-update:** Run `scoreline --update` anytime to get the latest version.

### Homebrew

Homebrew packaging is planned for the Scoreline fork but is not configured yet.
Use the install script until the new tap/formula is published.

### Install script

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/RuchikG/scoreline/main/scripts/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/RuchikG/scoreline/main/scripts/install.ps1 | iex
```

### Build from source

```bash
git clone https://github.com/RuchikG/scoreline.git
cd scoreline
go build 
./scoreline
```

## Usage

Run the application:
```bash
scoreline
```

**Navigation:** `↑`/`↓` or `j`/`k` to move, `Enter` to select, `/` to filter, `Tab` to focus view, `Esc` to go back, `q` to quit.

### Cricket Data

For live CricketData.org results, set the environment variable used by Cricket settings:

```bash
export CRICKETDATA_API_KEY="your-key"
```

Cricket archives use Cricsheet JSON data cached under the Scoreline cache directory. Open Cricket → Archives and press `r` to refresh the local archive cache. The current Cricsheet JSON archive is about 100 MB, so refresh is manual.

## Docs

- [Supported Leagues](docs/SUPPORTED_LEAGUES.md): Full list of available leagues and competitions, customize your preferences in the **Settings** menu.
- [Notifications](docs/NOTIFICATIONS.md): Desktop notification setup and configuration

---

<div align="center">

**Built with** [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss) & [Bubbles](https://github.com/charmbracelet/bubbles) by [Charm](https://charm.sh)

Author: [@RuchikG](https://github.com/RuchikG/)

</div>
