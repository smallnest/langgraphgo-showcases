package mind_spider

import (
	"context"
	"fmt"
	"strings"

	"github.com/smallnest/langgraphgo/showcases/BettaFish/query_engine"
)

// CrawlSocialMedia simulates fetching social media posts.
func CrawlSocialMedia(ctx context.Context, query string) ([]string, error) {
	// In a real system, this would use specific crawlers for Weibo, TikTok, etc.
	// Here we simulate it by searching for the topic on social media platforms via Tavily.

	platforms := []string{"site:weibo.com", "site:zhihu.com", "site:twitter.com", "site:reddit.com"}
	searchQuery := fmt.Sprintf("%s (%s)", query, strings.Join(platforms, " OR "))

	results, err := query_engine.ExecuteSearch(ctx, searchQuery, "basic_search_news", "", "")
	if err != nil {
		return nil, err
	}

	var posts []string
	for _, r := range results {
		posts = append(posts, fmt.Sprintf("Source: %s\nContent: %s", r.URL, r.Content))
	}

	return posts, nil
}
