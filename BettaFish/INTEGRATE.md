# MindSpider 集成文档

本文档介绍如何在 BettaFish Go 项目中集成和使用 MindSpider 的搜索数据。

## 目录

- [概述](#概述)
- [MindSpider 架构](#mindspider-架构)
- [数据库集成](#数据库集成)
- [Go 代码集成](#go-代码集成)
- [API 调用示例](#api-调用示例)
- [实战指南](#实战指南)

---

## 概述

MindSpider 是 BettaFish Python 版本的社交媒体爬虫系统，支持 7 个主流平台的数据采集：

| 平台 | 代码标识 | 数据类型 |
|------|----------|----------|
| 小红书 | `xhs` | 笔记、评论 |
| 抖音 | `dy` | 视频、评论 |
| 快手 | `ks` | 视频、评论 |
| B站 | `bili` | 视频、评论 |
| 微博 | `wb` | 帖子、评论 |
| 贴吧 | `tieba` | 帖子、评论 |
| 知乎 | `zhihu` | 回答、文章、评论 |

### 数据流程

```
MindSpider (Python)
        ↓
    MySQL/PostgreSQL 数据库
        ↓
   Go 数据访问层
        ↓
BettaFish Go 分析引擎
```

---

## MindSpider 架构

### 模块结构

```
MindSpider/
├── BroadTopicExtraction/       # 话题提取模块
│   ├── main.py                 # 热点新闻采集入口
│   ├── topic_extractor.py      # AI 话题提取器
│   ├── database_manager.py     # 数据库管理器
│   └── get_today_news.py       # 各平台新闻采集
│
├── DeepSentimentCrawling/       # 深度舆情爬取模块
│   ├── main.py                 # 深度爬取入口
│   ├── keyword_manager.py      # 关键词管理器
│   └── MediaCrawler/           # 多平台爬虫核心
│
└── schema/
    ├── mindspider_tables.sql    # 数据库表结构
    └── init_database.py         # 数据库初始化脚本
```

### 数据库表结构

#### 核心表

| 表名 | 用途 | 关键字段 |
|------|------|----------|
| `daily_news` | 每日热点新闻 | `news_id`, `source_platform`, `title`, `url` |
| `daily_topics` | AI 提取的每日话题 | `topic_id`, `topic_name`, `keywords`, `processing_status` |
| `crawling_tasks` | 爬取任务记录 | `task_id`, `topic_id`, `platform`, `task_status` |
| `xhs_note` | 小红书笔记 | `note_id`, `title`, `content`, `topic_id` |
| `douyin_aweme` | 抖音视频 | `aweme_id`, `title`, `content`, `topic_id` |
| `kuaishou_video` | 快手视频 | `video_id`, `title`, `content`, `topic_id` |
| `bilibili_video` | B站视频 | `bvid`, `title`, `content`, `topic_id` |
| `weibo_note` | 微博帖子 | `note_id`, `title`, `content`, `topic_id` |
| `tieba_note` | 贴吧帖子 | `note_id`, `title`, `content`, `topic_id` |
| `zhihu_content` | 知乎内容 | `content_id`, `title`, `content`, `topic_id` |

---

## 数据库集成

### 1. 创建数据库连接

在 Go 项目中添加数据库连接支持：

```go
// db/database.go
package db

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/go-sql-driver/mysql"
    // 或使用 PostgreSQL: _ "github.com/lib/pq"
)

type DatabaseConfig struct {
    Dialect   string // "mysql" or "postgresql"
    Host      string
    Port      int
    User      string
    Password  string
    DBName    string
    Charset   string
}

// NewConnection 创建数据库连接
func NewConnection(config DatabaseConfig) (*sql.DB, error) {
    var dsn string

    switch config.Dialect {
    case "mysql":
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true",
            config.User, config.Password, config.Host, config.Port,
            config.DBName, config.Charset)
    case "postgresql":
        dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
            config.Host, config.Port, config.User, config.Password, config.DBName)
    default:
        return nil, fmt.Errorf("unsupported dialect: %s", config.Dialect)
    }

    db, err := sql.Open(config.Dialect, dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // 配置连接池
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    // 测试连接
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Printf("Successfully connected to %s database", config.Dialect)
    return db, nil
}
```

### 2. 数据模型定义

创建与 MindSpider 数据库对应的数据模型：

```go
// models/mindspider.go
package models

import "time"

// DailyTopic 每日话题
type DailyTopic struct {
    ID              int       `json:"id"`
    TopicID         string    `json:"topic_id"`
    TopicName       string    `json:"topic_name"`
    TopicDesc       string    `json:"topic_description"`
    Keywords        string    `json:"keywords"`
    ExtractDate     time.Time `json:"extract_date"`
    RelevanceScore  float64   `json:"relevance_score"`
    NewsCount       int       `json:"news_count"`
    ProcessingStatus string   `json:"processing_status"`
}

// SocialContent 社交媒体内容 (通用结构)
type SocialContent struct {
    ID              int64     `json:"id"`
    ContentID       string    `json:"content_id"`
    Platform        string    `json:"platform"` // xhs, dy, ks, bili, wb, tieba, zhihu
    TopicID         string    `json:"topic_id"`
    Title           string    `json:"title"`
    Content         string    `json:"content"`
    AuthorID        string    `json:"author_id"`
    AuthorName      string    `json:"author_name"`
    LikeCount       int       `json:"like_count"`
    CommentCount    int       `json:"comment_count"`
    ShareCount      int       `json:"share_count"`
    PublishTime     time.Time `json:"publish_time"`
    CrawledAt       time.Time `json:"crawled_at"`
    ExtraData       string    `json:"extra_data"` // JSON 格式存储额外字段
}

// CrawlingTask 爬取任务
type CrawlingTask struct {
    ID              int64     `json:"id"`
    TaskID          string    `json:"task_id"`
    TopicID         string    `json:"topic_id"`
    Platform        string    `json:"platform"`
    SearchKeywords  string    `json:"search_keywords"`
    TaskStatus      string    `json:"task_status"`
    StartTime       *time.Time `json:"start_time"`
    EndTime         *time.Time `json:"end_time"`
    TotalCrawled    int       `json:"total_crawled"`
    SuccessCount    int       `json:"success_count"`
    ErrorCount      int       `json:"error_count"`
    ScheduledDate   time.Time `json:"scheduled_date"`
}
```

### 3. 数据访问层

创建查询 MindSpider 数据的仓储层：

```go
// repository/mindspider.go
package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "strings"
    "time"

    "github.com/yourusername/bettafish/models"
)

type MindSpiderRepository struct {
    db *sql.DB
}

func NewMindSpiderRepository(db *sql.DB) *MindSpiderRepository {
    return &MindSpiderRepository{db: db}
}

// SearchTopicGlobally 全局话题搜索 (模拟 InsightEngine 的 search_topic_globally)
func (r *MindSpiderRepository) SearchTopicGlobally(ctx context.Context, query string, limitPerTable int) ([]models.SocialContent, error) {
    platforms := []string{"xhs", "douyin", "kuaishou", "bilibili", "weibo", "tieba", "zhihu"}

    var allResults []models.SocialContent

    for _, platform := range platforms {
        results, err := r.searchByPlatform(ctx, platform, query, limitPerTable)
        if err != nil {
            // 记录错误但继续查询其他平台
            fmt.Printf("Error searching %s: %v\n", platform, err)
            continue
        }
        allResults = append(allResults, results...)
    }

    return allResults, nil
}

// searchByPlatform 按平台搜索
func (r *MindSpiderRepository) searchByPlatform(ctx context.Context, platform, query string, limit int) ([]models.SocialContent, error) {
    var tableName string
    switch platform {
    case "xhs":
        tableName = "xhs_note"
    case "dy":
        tableName = "douyin_aweme"
    case "ks":
        tableName = "kuaishou_video"
    case "bili":
        tableName = "bilibili_video"
    case "wb":
        tableName = "weibo_note"
    case "tieba":
        tableName = "tieba_note"
    case "zhihu":
        tableName = "zhihu_content"
    default:
        return nil, fmt.Errorf("unsupported platform: %s", platform)
    }

    sqlQuery := fmt.Sprintf(`
        SELECT id, content_id, '%s' as platform, topic_id,
               title, content, author_id, author_name,
               like_count, comment_count, share_count,
               publish_time, crawled_at
        FROM %s
        WHERE title LIKE ? OR content LIKE ?
        ORDER BY like_count DESC
        LIMIT ?
    `, platform, tableName)

    rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", "%"+query+"%", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []models.SocialContent
    for rows.Next() {
        var content models.SocialContent
        err := rows.Scan(
            &content.ID, &content.ContentID, &content.Platform, &content.TopicID,
            &content.Title, &content.Content, &content.AuthorID, &content.AuthorName,
            &content.LikeCount, &content.CommentCount, &content.ShareCount,
            &content.PublishTime, &content.CrawledAt,
        )
        if err != nil {
            continue
        }
        results = append(results, content)
    }

    return results, nil
}

// SearchByDate 按日期范围搜索
func (r *MindSpiderRepository) SearchByDate(ctx context.Context, query string, startDate, endDate time.Time, limitPerTable int) ([]models.SocialContent, error) {
    platforms := []string{"xhs", "douyin", "kuaishou", "bilibili", "weibo", "tieba", "zhihu"}
    var allResults []models.SocialContent

    for _, platform := range platforms {
        results, err := r.searchByPlatformAndDate(ctx, platform, query, startDate, endDate, limitPerTable)
        if err != nil {
            fmt.Printf("Error searching %s: %v\n", platform, err)
            continue
        }
        allResults = append(allResults, results...)
    }

    return allResults, nil
}

// searchByPlatformAndDate 按平台和日期搜索
func (r *MindSpiderRepository) searchByPlatformAndDate(ctx context.Context, platform, query string, startDate, endDate time.Time, limit int) ([]models.SocialContent, error) {
    var tableName string
    switch platform {
    case "xhs":
        tableName = "xhs_note"
    case "dy":
        tableName = "douyin_aweme"
    case "ks":
        tableName = "kuaishou_video"
    case "bili":
        tableName = "bilibili_video"
    case "wb":
        tableName = "weibo_note"
    case "tieba":
        tableName = "tieba_note"
    case "zhihu":
        tableName = "zhihu_content"
    default:
        return nil, fmt.Errorf("unsupported platform: %s", platform)
    }

    sqlQuery := fmt.Sprintf(`
        SELECT id, content_id, '%s' as platform, topic_id,
               title, content, author_id, author_name,
               like_count, comment_count, share_count,
               publish_time, crawled_at
        FROM %s
        WHERE (title LIKE ? OR content LIKE ?)
          AND publish_time >= ? AND publish_time <= ?
        ORDER BY like_count DESC
        LIMIT ?
    `, platform, tableName)

    rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", "%"+query+"%", startDate, endDate, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []models.SocialContent
    for rows.Next() {
        var content models.SocialContent
        err := rows.Scan(
            &content.ID, &content.ContentID, &content.Platform, &content.TopicID,
            &content.Title, &content.Content, &content.AuthorID, &content.AuthorName,
            &content.LikeCount, &content.CommentCount, &content.ShareCount,
            &content.PublishTime, &content.CrawledAt,
        )
        if err != nil {
            continue
        }
        results = append(results, content)
    }

    return results, nil
}

// GetCommentsForTopic 获取话题评论
func (r *MindSpiderRepository) GetCommentsForTopic(ctx context.Context, topicID string, limit int) ([]map[string]any, error) {
    // 这里需要根据实际的评论表结构来实现
    // 各平台的评论表结构可能不同
    return nil, fmt.Errorf("comment retrieval not yet implemented")
}

// GetTopicByID 根据 ID 获取话题
func (r *MindSpiderRepository) GetTopicByID(ctx context.Context, topicID string) (*models.DailyTopic, error) {
    query := `
        SELECT id, topic_id, topic_name, topic_description, keywords,
               extract_date, relevance_score, news_count, processing_status
        FROM daily_topics
        WHERE topic_id = ?
    `

    var topic models.DailyTopic
    var extractDate sql.NullTime

    err := r.db.QueryRowContext(ctx, query, topicID).Scan(
        &topic.ID, &topic.TopicID, &topic.TopicName, &topic.TopicDesc,
        &topic.Keywords, &extractDate, &topic.RelevanceScore,
        &topic.NewsCount, &topic.ProcessingStatus,
    )

    if err != nil {
        return nil, err
    }

    if extractDate.Valid {
        topic.ExtractDate = extractDate.Time
    }

    return &topic, nil
}

// GetActiveTopics 获取活跃话题列表
func (r *MindSpiderRepository) GetActiveTopics(ctx context.Context, days int) ([]models.DailyTopic, error) {
    query := `
        SELECT id, topic_id, topic_name, topic_description, keywords,
               extract_date, relevance_score, news_count, processing_status
        FROM daily_topics
        WHERE extract_date >= DATE_SUB(CURDATE(), INTERVAL ? DAY)
          AND processing_status = 'completed'
        ORDER BY relevance_score DESC
    `

    rows, err := r.db.QueryContext(ctx, query, days)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var topics []models.DailyTopic
    for rows.Next() {
        var topic models.DailyTopic
        var extractDate sql.NullTime

        err := rows.Scan(
            &topic.ID, &topic.TopicID, &topic.TopicName, &topic.TopicDesc,
            &topic.Keywords, &extractDate, &topic.RelevanceScore,
            &topic.NewsCount, &topic.ProcessingStatus,
        )

        if err != nil {
            continue
        }

        if extractDate.Valid {
            topic.ExtractDate = extractDate.Time
        }

        topics = append(topics, topic)
    }

    return topics, nil
}
```

---

## Go 代码集成

### 更新 InsightEngine

更新现有的 InsightEngine 以使用 MindSpider 数据：

```go
// insight_engine/agent.go (更新版)
package insight_engine

import (
    "context"
    "fmt"
    "strings"

    "github.com/yourusername/bettafish/models"
    "github.com/yourusername/bettafish/repository"
    "github.com/yourusername/bettafish/schema"
)

type InsightEngineConfig struct {
    UseMindSpider   bool   // 是否使用 MindSpider 数据库
    MindSpiderDB    *repository.MindSpiderRepository
    FallbackToTavily bool   // 当 MindSpider 无数据时是否回退到 Tavily
}

// InsightEngineNode 增强版洞察引擎节点
func InsightEngineNode(ctx context.Context, state any, config InsightEngineConfig) (any, error) {
    s := state.(*schema.BettaFishState)
    fmt.Printf("InsightEngine: 正在挖掘内部洞察 '%s'...\n", s.Query)

    if config.UseMindSpider && config.MindSpiderDB != nil {
        // 尝试从 MindSpider 数据库获取数据
        results, err := searchMindSpider(ctx, config.MindSpiderDB, s.Query)
        if err == nil && len(results) > 0 {
            fmt.Printf("InsightEngine: 从 MindSpider 找到 %d 条结果\n", len(results))
            // 处理 MindSpider 数据...
            return processMindSpiderResults(ctx, s, results)
        } else if config.FallbackToTavily {
            fmt.Printf("InsightEngine: MindSpider 无数据，回退到 Tavily\n")
        } else {
            return nil, fmt.Errorf("MindSpider 查询失败: %w", err)
        }
    }

    // 原有的 Tavily 搜索逻辑作为后备方案...
    return s, nil
}

func searchMindSpider(ctx context.Context, repo *repository.MindSpiderRepository, query string) ([]models.SocialContent, error) {
    // 全局搜索
    results, err := repo.SearchTopicGlobally(ctx, query, 10)
    if err != nil {
        return nil, err
    }
    return results, nil
}

func processMindSpiderResults(ctx context.Context, s *schema.BettaFishState, results []models.SocialContent) (*schema.BettaFishState, error) {
    // 将 MindSpider 结果转换为 BettaFish 状态格式
    var insights []string

    // 按平台分组统计
    platformStats := make(map[string]int)
    for _, r := range results {
        platformStats[r.Platform]++
    }

    insights = append(insights, fmt.Sprintf("## 数据来源分布"))
    for platform, count := range platformStats {
        platformName := getPlatformName(platform)
        insights = append(insights, fmt.Sprintf("- **%s**: %d 条内容", platformName, count))
    }

    // 提取代表性内容
    insights = append(insights, "\n## 代表性内容")
    for i, r := range results {
        if i >= 5 {
            break
        }
        insights = append(insights, fmt.Sprintf("\n### %s", r.Title))
        insights = append(insights, fmt.Sprintf("- 平台: %s", getPlatformName(r.Platform)))
        insights = append(insights, fmt.Sprintf("- 作者: %s", r.AuthorName))
        insights = append(insights, fmt.Sprintf("- 点赞: %d | 评论: %d", r.LikeCount, r.CommentCount))
        insights = append(insights, fmt.Sprintf("- 内容: %s", truncateString(r.Content, 200)))
    }

    s.InsightResults = insights
    return s, nil
}

func getPlatformName(code string) string {
    names := map[string]string{
        "xhs":    "小红书",
        "dy":     "抖音",
        "ks":     "快手",
        "bili":   "哔哩哔哩",
        "wb":     "微博",
        "tieba":  "贴吧",
        "zhihu":  "知乎",
    }
    if name, ok := names[code]; ok {
        return name
    }
    return code
}

func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}
```

### 更新主程序

```go
// main.go (更新版)
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "os"

    "github.com/yourusername/bettafish/db"
    "github.com/yourusername/bettafish/repository"

    _ "github.com/go-sql-driver/mysql"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("用法: go run main.go <查询>")
        return
    }

    query := os.Args[1]

    // 初始化 MindSpider 数据库连接 (可选)
    var mindspiderRepo *repository.MindSpiderRepository
    useMindSpider := os.Getenv("USE_MINDSPIDER") == "true"

    if useMindSpider {
        dbConfig := db.DatabaseConfig{
            Dialect:   os.Getenv("MINDSPIDER_DB_DIALECT"),
            Host:      os.Getenv("MINDSPIDER_DB_HOST"),
            Port:      parseInt(os.Getenv("MINDSPIDER_DB_PORT")),
            User:      os.Getenv("MINDSPIDER_DB_USER"),
            Password:  os.Getenv("MINDSPIDER_DB_PASSWORD"),
            DBName:    os.Getenv("MINDSPIDER_DB_NAME"),
            Charset:   os.Getenv("MINDSPIDER_DB_CHARSET"),
        }

        spiderDB, err := db.NewConnection(dbConfig)
        if err != nil {
            log.Printf("警告: 无法连接到 MindSpider 数据库: %v", err)
            log.Println("将使用 Tavily API 作为数据源")
        } else {
            mindspiderRepo = repository.NewMindSpiderRepository(spiderDB)
            defer spiderDB.Close()
        }
    }

    // 初始化状态
    initialState := schema.NewBettaFishState(query)

    // 配置 InsightEngine
    insightConfig := insight_engine.InsightEngineConfig{
        UseMindSpider:    useMindSpider,
        MindSpiderDB:     mindspiderRepo,
        FallbackToTavily: true,
    }

    // 创建并运行图...
}
```

---

## API 调用示例

### 直接 SQL 查询

```go
// 查询特定话题的所有相关内容
func queryTopicContents(db *sql.DB, topicID string) ([]SocialContent, error) {
    query := `
        SELECT 'xhs' as platform, note_id, title, content, like_count
        FROM xhs_note WHERE topic_id = ?
        UNION ALL
        SELECT 'dy' as platform, aweme_id, title, content, like_count
        FROM douyin_aweme WHERE topic_id = ?
        UNION ALL
        SELECT 'bili' as platform, bvid, title, content, like_count
        FROM bilibili_video WHERE topic_id = ?
    `

    rows, err := db.Query(query, topicID, topicID, topicID)
    // ... 处理结果
}
```

### 使用仓储层

```go
// 示例: 搜索最近7天关于"人工智能"的内容
func searchRecentContent(repo *MindSpiderRepository) {
    ctx := context.Background()
    query := "人工智能"
    startDate := time.Now().AddDate(0, 0, -7)
    endDate := time.Now()

    results, err := repo.SearchByDate(ctx, query, startDate, endDate, 20)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("找到 %d 条相关内容\n", len(results))
    for _, r := range results {
        fmt.Printf("[%s] %s - %d 赞\n", r.Platform, r.Title, r.LikeCount)
    }
}
```

---

## 实战指南

### 步骤 1: 部署 MindSpider

```bash
# 1. 克隆 MindSpider 目录到 Python 项目
cd /path/to/python/project
cp -r /path/to/bettafish/MindPython/MindSpider .

# 2. 安装依赖
pip install -r MindSpider/requirements.txt

# 3. 配置数据库
cp MindSpider/config.py.example MindSpider/config.py
# 编辑 config.py 设置数据库连接信息

# 4. 初始化数据库
python MindSpider/schema/init_database.py

# 5. 运行话题提取
python MindSpider/BroadTopicExtraction/main.py

# 6. 运行深度爬取
python MindSpider/DeepSentimentCrawling/main.py --platforms xhs,dy,bili --date 2024-01-20
```

### 步骤 2: 配置 Go 项目

创建 `.env` 文件：

```bash
# MindSpider 数据库配置
USE_MINDSPIDER=true
MINDSPIDER_DB_DIALECT=mysql
MINDSPIDER_DB_HOST=localhost
MINDSPIDER_DB_PORT=3306
MINDSPIDER_DB_USER=your_username
MINDSPIDER_DB_PASSWORD=your_password
MINDSPIDER_DB_NAME=mindspider
MINDSPIDER_DB_CHARSET=utf8mb4

# LLM 配置 (用于 BettaFish 分析引擎)
OPENAI_API_KEY=your_api_key
OPENAI_API_BASE=https://api.deepseek.com/v1
OPENAI_MODEL=deepseek-chat
```

### 步骤 3: 运行完整分析

```bash
# 构建 Go 项目
go build -o bettafish

# 运行分析 (将自动使用 MindSpider 数据)
./bettafish "分析人工智能在教育领域的应用"
```

### 步骤 4: 查看分析结果

分析结果将包含：
- 来自 MindSpider 数据库的真实社交媒体内容
- 按平台分布的内容统计
- 基于真实用户评论的舆情分析
- 综合多个平台的观点汇总

---

## 高级配置

### 并发查询优化

```go
// 并发查询所有平台
func (r *MindSpiderRepository) SearchAllPlatformsConcurrent(ctx context.Context, query string, limit int) ([]models.SocialContent, error) {
    platforms := []string{"xhs", "dy", "ks", "bili", "wb", "tieba", "zhihu"}
    resultChan := make(chan []models.SocialContent, len(platforms))
    errChan := make(chan error, len(platforms))

    for _, platform := range platforms {
        go func(p string) {
            results, err := r.searchByPlatform(ctx, p, query, limit)
            if err != nil {
                errChan <- err
                return
            }
            resultChan <- results
        }(platform)
    }

    var allResults []models.SocialContent
    for i := 0; i < len(platforms); i++ {
        select {
        case results := <-resultChan:
            allResults = append(allResults, results...)
        case err := <-errChan:
            log.Printf("Platform query error: %v", err)
        }
    }

    return allResults, nil
}
```

### 数据缓存

```go
import "github.com/patrickmn/go-cache"

type CachedMindSpiderRepository struct {
    repo  *MindSpiderRepository
    cache *cache.Cache
}

func NewCachedMindSpiderRepository(repo *MindSpiderRepository) *CachedMindSpiderRepository {
    return &CachedMindSpiderRepository{
        repo:  repo,
        cache: cache.New(5*time.Minute, 10*time.Minute),
    }
}

func (r *CachedMindSpiderRepository) SearchTopicGlobally(ctx context.Context, query string, limit int) ([]models.SocialContent, error) {
    cacheKey := fmt.Sprintf("search:%s:%d", query, limit)

    if cached, found := r.cache.Get(cacheKey); found {
        return cached.([]models.SocialContent), nil
    }

    results, err := r.repo.SearchTopicGlobally(ctx, query, limit)
    if err != nil {
        return nil, err
    }

    r.cache.Set(cacheKey, results, cache.DefaultExpiration)
    return results, nil
}
```

---

## 故障排除

### 常见问题

| 问题 | 解决方案 |
|------|----------|
| 数据库连接失败 | 检查数据库服务是否运行，验证连接参数 |
| 查询结果为空 | 确认 MindSpider 已完成爬取，检查 `daily_topics` 表 |
| 编码问题 | 确保数据库使用 `utf8mb4` 字符集 |
| 性能问题 | 添加索引，使用查询缓存，限制返回结果数量 |

### 调试建议

```go
// 启用详细日志
db.SetLogger(logger.New(os.Stdout, "db: ", log.Lmicroseconds))

// 查看实际执行的 SQL
db.LogMode(true)

// 检查表结构
DESCRIBE xhs_note;
SHOW INDEX FROM xhs_note;
```

---

## 参考资源

- [MindSpider README](MindPython/MindSpider/README.md)
- [数据库表结构](MindPython/MindSpider/schema/mindspider_tables.sql)
- [配置示例](MindPython/MindSpider/config.py.example)

---

**文档版本**: 1.0
**最后更新**: 2026-01-04
**维护者**: BettaFish Go Team
