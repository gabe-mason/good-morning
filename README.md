# Good Morning

A personal daily summary generator that aggregates information from your calendar, GitHub, and Linear to provide a comprehensive morning briefing.

## Features

- Calendar integration (via ICS URL)
- GitHub activity tracking
- Linear task management
- AI-powered summary generation using Anthropic's Claude
- Daily markdown summaries

## Prerequisites

- Go 1.24 or later
- Anthropic API key
- GitHub personal access token
- Linear API token
- Calendar ICS URL
- A directory for storing summaries

## Installation

1. Clone the repository:
```bash
git clone https://github.com/gabe-mason/good-morning.git
cd good-morning
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export GOOD_MORNING_ANTHROPIC_API_KEY="your-anthropic-api-key"
export GOOD_MORNING_ROOT="path/to/store/summaries"
export GOOD_MORNING_ICS_URL="your-calendar-ics-url"
export GOOD_MORNING_GITHUB_TOKEN="your-github-token"
export GOOD_MORNING_LINEAR_TOKEN="your-linear-token"
export GOOD_MORNING_LINEAR_TEAMS="your-linear-teams"
export GOOD_MORNING_MY_NAME="your-name"
```

## Usage

Run the program:
```bash
go run main.go
```

The program will:
1. Connect to your configured services
2. Gather information about your calendar, GitHub activity, and Linear tasks
3. Generate a daily summary using AI
4. Save the summary as a markdown file in your configured directory

## Configuration

The following environment variables are required:

- `GOOD_MORNING_ANTHROPIC_API_KEY`: Your Anthropic API key
- `GOOD_MORNING_ROOT`: Directory where summaries will be stored
- `GOOD_MORNING_ICS_URL`: URL to your calendar's ICS feed
- `GOOD_MORNING_GITHUB_TOKEN`: GitHub personal access token
- `GOOD_MORNING_LINEAR_TOKEN`: Linear API token
- `GOOD_MORNING_LINEAR_TEAMS`: Comma-separated list of Linear team IDs
- `GOOD_MORNING_MY_NAME`: Your name for personalization

## Output

The program generates a daily markdown file with the following format:
```
YYYY-MM-DD.md
```

The summary includes:
- Calendar events for the day
- Recent GitHub activity
- Linear tasks and updates
- AI-generated insights and recommendations

## Next Steps
- [ ] Create good morning daemon
- [ ] Implement webhooks to update the document from Linear and Github (or this could be just syncing)
- [ ] Periodically sync calendar
- [ ] Add file watcher to analyse meeting notes and create actions off the back of notes
- [ ] Implement context summarisation to reduce token count with sliding window
- [ ] Implement auto git commit via Anthropic computer use