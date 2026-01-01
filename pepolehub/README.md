# PeopleHub (Go Port)

This is a Go implementation of the [PeopleHub](https://github.com/MeirKaD/pepolehub) research agent using [langgraphgo](https://github.com/smallnest/langgraphgo).

It automates the process of researching a person by:
1.  Fetching their LinkedIn profile (via Web Search).
2.  Searching for them on Google (via Tavily).
3.  Scraping relevant web pages.
4.  Summarizing the content (via OpenAI).
5.  Generating a comprehensive research report.

## Prerequisites

You need the following API keys:
*   `OPENAI_API_KEY`: For generating summaries and reports.
*   `TAVILY_API_KEY`: For searching the web and LinkedIn profiles.

## Usage

Set the environment variables and run:

```bash
export OPENAI_API_KEY="sk-..."
export TAVILY_API_KEY="tvly-..."
go run showcases/pepolehub/*.go -name "John Doe" -linkedin "https://linkedin.com/in/johndoe"
```

## Features

*   **Graph-based Workflow**: Uses `langgraphgo` to orchestrate the research steps.
*   **Real Implementation**: Uses Tavily for search and OpenAI for intelligence (no mocks).
*   **Parallel Execution**: Fetches LinkedIn data and performs web searches in parallel.
*   **Conditional Routing**: Dynamically decides whether to scrape pages based on search results.
*   **State Management**: Robust state handling with `FieldMerger`.

## Architecture

The agent follows this graph workflow:

1.  **Start**: Initializes the research.
2.  **Parallel Step**:
    *   `FetchLinkedIn`: Searches for the profile content.
    *   `ExecuteSearch`: Searches for the person on the web.
3.  **Scrape & Summarize**: If search results are found, it scrapes and summarizes the content.
4.  **Aggregate**: Combines LinkedIn data and web summaries.
5.  **WriteReport**: Generates the final markdown report.

## Original Project

*   [MeirKaD/pepolehub](https://github.com/MeirKaD/pepolehub)